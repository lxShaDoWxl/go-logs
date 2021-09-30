package logs

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"google.golang.org/grpc"
)

type ConfigSentry struct {
	DSN         string
	Environment string
}

var configSentry = ConfigSentry{}

func initializeSentry(c ConfigSentry) {
	configSentry = c
	if configSentry.DSN != "" {
		initSentry()
	}
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
			hub.Scope().SetContext("exception_metadata", v.Meta)
			err = v.Err
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
		hub.Recover(err)
		// Because the goroutine running this code is going to crash the
		// program, call Flush to send the event to Sentry before it is too
		// late. Set the timeout to an appropriate value depending on your
		// program. The value is the maximum time to wait before giving up
		// and dropping the event.
		hub.Flush(2 * time.Second)
		// Note that if multiple goroutines panic, possibly only the first
		// one to call Flush will succeed in sending the event. If you want
		// to capture multiple panics and still crash the program
		// afterwards, you need to coordinate error reporting and
		// termination differently.
		log.Fatalf("%v", err)
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
		hub.Scope().SetContext("exception_metadata", v.Meta)
		err = v.Err
	}
	eventID := hub.CaptureException(err)
	if eventID != nil {
		hub.Flush(time.Second * time.Duration(5))
	}
}
