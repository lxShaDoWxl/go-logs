package gin

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const valuesKey = "sentry"

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
	hub.Scope().SetRequest(ctx.Request)
	ctx.Set(valuesKey, hub)
	defer h.recoverWithSentry(hub, ctx.Request)
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
	if hub, ok := ctx.Get(valuesKey); ok {
		if hub, ok := hub.(*sentry.Hub); ok {
			return hub
		}
	}
	return nil
}
