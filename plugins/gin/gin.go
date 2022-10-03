package gin

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
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
	hub := sentry.GetHubFromContext(ctx.Request.Context())
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	span := sentry.StartSpan(ctx.Request.Context(), "http.server",
		sentry.TransactionName(fmt.Sprintf("%s %s", ctx.Request.Method, ctx.Request.URL.Path)),
		sentry.ContinueFromRequest(ctx.Request),
	)
	defer span.Finish()
	r := ctx.Request.WithContext(span.Context())
	if cip := ctx.ClientIP(); cip != "" {
		r.RemoteAddr = cip
	}
	hub.Scope().SetRequest(r)
	ctx.Set(valuesKeyHub, hub)
	ctx.Set(valuesKeySpan, span)
	defer h.recoverWithSentry(hub, r)
	ctx.Next()
}

func (h *handler) recoverWithSentry(hub *sentry.Hub, r *http.Request) {
	if err := recover(); err != nil {
		if !isBrokenPipeError(err) {
			v, ok := err.(*errors.Error)
			if !ok {
				v = errors.Wrap(err, 2)
			}
			eventID := hub.RecoverWithContext(
				context.WithValue(r.Context(), sentry.RequestContextKey, r),
				v,
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
