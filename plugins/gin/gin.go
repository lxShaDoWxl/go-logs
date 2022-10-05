package gin

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

const valuesKeyHub = "sentry-hub"
const valuesKeySpan = "sentry-span"

type handler struct {
	repanic         bool
	waitForDelivery bool
	timeout         time.Duration
}

type ConfigPluginGin struct {
	Gin *gin.Engine
	// Repanic configures whether Sentry should repanic after recovery, in most cases it should be set to true,
	// as gin.Default includes it's own Recovery middleware what handles http responses.
	Repanic bool
	// WaitForDelivery configures whether you want to block the request before moving forward with the response.
	// Because Gin's default Recovery handler doesn't restart the application,
	// it's safe to either skip this option or set it to false.
	WaitForDelivery bool
	// Timeout for the event delivery requests.
	Timeout time.Duration
}
type PluginGin struct {
	config *ConfigPluginGin
}

func NewPluginGin(c *ConfigPluginGin) *PluginGin {
	return &PluginGin{config: c}
}

func (p PluginGin) Initialize() {
	timeout := p.config.Timeout
	if timeout == 0 {
		timeout = 2 * time.Second
	}
	p.config.Gin.Use((&handler{
		repanic:         p.config.Repanic,
		timeout:         timeout,
		waitForDelivery: p.config.WaitForDelivery,
	}).handle)
}

func (h *handler) handle(ctx *gin.Context) {
	ctxStd := ctx.Request.Context()
	hub := sentry.GetHubFromContext(ctxStd)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
		ctxStd = sentry.SetHubOnContext(ctxStd, hub)
	}
	span := sentry.StartSpan(ctxStd, "http.server",
		sentry.TransactionName(fmt.Sprintf("%s %s", ctx.Request.Method, ctx.Request.URL.Path)),
		sentry.ContinueFromRequest(ctx.Request),
	)
	defer span.Finish()
	ctx.Request = ctx.Request.WithContext(span.Context())
	if cip := ctx.ClientIP(); cip != "" {
		hub.Scope().SetUser(sentry.User{
			IPAddress: cip,
		})
	}

	hub.Scope().SetRequest(ctx.Request)

	ctx.Set(valuesKeyHub, hub)
	ctx.Set(valuesKeySpan, span)
	defer h.recoverWithSentry(hub, ctx.Request)
	ctx.Next()
}

func (h *handler) recoverWithSentry(hub *sentry.Hub, r *http.Request) {
	if err := recover(); err != nil {
		if !isBrokenPipeError(err) {
			eventID := hub.RecoverWithContext(
				context.WithValue(r.Context(), sentry.RequestContextKey, r),
				errors.WrapWithDepth(2, err.(error), ""),
			)
			if eventID != nil && h.waitForDelivery {
				hub.Flush(h.timeout)
			}
		}
		if h.repanic {
			panic(err)
		}
	}
}

// Check for a broken connection, as this is what Gin does already.
func isBrokenPipeError(err interface{}) bool {
	if netErr, ok := err.(*net.OpError); ok {
		if sysErr, ok := netErr.Err.(*os.SyscallError); ok {
			if strings.Contains(strings.ToLower(sysErr.Error()), "broken pipe") ||
				strings.Contains(strings.ToLower(sysErr.Error()), "connection reset by peer") {
				return true
			}
		}
	}
	return false
}

// GetHubFromContext retrieves attached *sentry.Hub instance from gin.Context.
func GetHubFromContext(ctx *gin.Context) *sentry.Hub {
	if hub, ok := ctx.Get(valuesKeyHub); ok {
		if hubReal, correct := hub.(*sentry.Hub); correct {
			return hubReal
		}
	}
	return nil
}

// GetSpanFromContext retrieves attached *sentry.Span instance from gin.Context.
func GetSpanFromContext(ctx *gin.Context) *sentry.Span {
	if span, ok := ctx.Get(valuesKeySpan); ok {
		if spanReal, correct := span.(*sentry.Span); correct {
			return spanReal
		}
	}
	return nil
}
