package logs

import (
	"context"
	"github.com/go-errors/errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"google.golang.org/grpc"
)

//TODO в плагинах добавить функцию которая будет модифицировать получение hub в функции sendLogSentry это надо что бы вытащить из gin контекта hub и реквест
type Plugin interface {
	Initialize()
}
type ConfigSentry struct {
	DSN         string
	Environment string
	Plugins     []Plugin
}

var configSentry = ConfigSentry{}

func initializeSentry(c ConfigSentry) {
	configSentry = c
	if configSentry.DSN == "" {
		return
	}
	initSentry()
	for _, plugin := range c.Plugins {
		plugin.Initialize()
	}
}
func AddedSentryPlugin(plugins ...Plugin) {
	for _, plugin := range plugins {
		plugin.Initialize()
	}
}
func AddSentryHubCtx(ctx context.Context, hub *sentry.Hub) context.Context {

	checkHub := sentry.GetHubFromContext(ctx)
	if checkHub != nil {
		return ctx
	}
	return sentry.SetHubOnContext(ctx, hub)
}

func ModifyGrpc(
	streamMiddlewares []grpc.StreamServerInterceptor,
	unaryMiddlewares []grpc.UnaryServerInterceptor,
) (
	[]grpc.StreamServerInterceptor,
	[]grpc.UnaryServerInterceptor,
) {
	if configSentry.DSN != "" {
		return append(streamMiddlewares, StreamServerInterceptor()),
			append(unaryMiddlewares, UnaryServerInterceptor())

	}
	return streamMiddlewares, unaryMiddlewares
}
func DefferSentry() {

	if err := recover(); err != nil {
		hub := sentry.CurrentHub().Clone()
		if v, ok := err.(Exception); ok {
			hub.Scope().SetContext("exception_metadata", recursiveUnwrap(v.GetMeta(), 1))
			err = v.Err
		}
		vError, ok := err.(*errors.Error)
		if !ok {
			vError = errors.Wrap(err, 2)
		}
		// filterFrames removes frames from outgoing events that reference the
		filterFrames := func(event *sentry.Event) {
			for _, e := range event.Exception {
				if e.Stacktrace == nil {
					continue
				}
				frames := e.Stacktrace.Frames[:0]
				for index := range e.Stacktrace.Frames {
					frame := e.Stacktrace.Frames[index]
					if strings.HasSuffix(frame.Module, "grpc_sentry") && strings.HasPrefix(frame.Function, "Recover") {
						continue
					}
					frames = append(frames, frame)
				}
				e.Stacktrace.Frames = frames
			}
		}

		// Add an EventProcessor to the scope. The event processor is a function
		// that can change events before they are sent to Sentry.
		// Alternatively, see also ClientOptions.BeforeSend, which is a special
		// event processor applied to error events.
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.AddEventProcessor(func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
				filterFrames(event)
				return event
			})
		})
		// Create an event and enqueue it for reporting.
		hub.Recover(vError)
		// Because the goroutine running this code is going to crash the
		// program, call Flush to send the event to Sentry before it is too
		// late. Set the timeout to an appropriate value depending on your
		// program. The value is the maximum time to wait before giving up
		// and dropping the event.
		hub.Flush(5 * time.Second)
		// Note that if multiple goroutines panic, possibly only the first
		// one to call Flush will succeed in sending the event. If you want
		// to capture multiple panics and still crash the program
		// afterwards, you need to coordinate error reporting and
		// termination differently.
		L.Fatal(context.Background(), vError.ErrorStack())
	}
	sentry.Flush(2 * time.Second)
}
func initSentry() {
	sentryDNS := configSentry.DSN

	sentryEnvironment := configSentry.Environment

	err := sentry.Init(sentry.ClientOptions{
		Dsn:         sentryDNS,
		Environment: sentryEnvironment,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
}
func sendLogSentry(ctx context.Context, err error) {
	if configSentry.DSN == "" {
		return
	}
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	if v, ok := err.(Exception); ok {
		hub.Scope().SetContext("exception_metadata", recursiveUnwrap(v.GetMeta(), 1))
		err = v.Err
	}
	eventID := hub.CaptureException(err)
	if eventID != nil {
		hub.Flush(time.Second * time.Duration(5))
	}
}
func recursiveUnwrap(maps map[string]interface{}, level int) map[string]interface{} {
	var result = make(map[string]interface{})
	if value, ok := maps["level_1"]; ok {
		result["level_"+strconv.Itoa(level)] = value
		if value2, ok2 := maps["level_2"]; ok2 {
			values := recursiveUnwrap(value2.(map[string]interface{}), level+1)
			for _, v := range values {
				level++
				result["level_"+strconv.Itoa(level)] = v
			}
		}
	}
	return result
}
