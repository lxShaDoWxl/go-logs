package logs

import (
	"context"
	"fmt"
	"github.com/cockroachdb/errors"
	"log"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"google.golang.org/grpc"
)

type Plugin interface {
	Initialize()
}
type ConfigSentry struct {
	DSN         string
	Environment string
	// The sample rate for performance data submission (0.0 - 1.0)
	// (defaults to 0.0 (meaning performance monitoring disabled))
	TracesSampleRate float64
	// In debug mode, the debug information is printed to stdout to help you
	// understand what sentry is doing.
	Debug   bool
	Plugins []Plugin
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
		L.Fatal(context.Background(), fmt.Sprintf("%+v", err))

		hub := sentry.CurrentHub().Clone()
		if v, ok := err.(Exception); ok {
			hub.Scope().SetContext("exception_metadata", recursiveUnwrap(v.GetMeta(), 1))
			err = v.Err
			sendLogSentry(context.Background(), v)
		} else {
			sendLogSentry(context.Background(), err.(error))
		}
	}
	sentry.Flush(2 * time.Second)
}
func initSentry() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              configSentry.DSN,
		Environment:      configSentry.Environment,
		TracesSampleRate: configSentry.TracesSampleRate,
		Debug:            configSentry.Debug,
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
		// hub.Scope().SetFingerprint([]string{v.ErrorStack()})

		err = v.Err
	}

	// eventID := hub.CaptureException(err)
	event, extraDetails := errors.BuildSentryReport(err)

	for extraKey, extraValue := range extraDetails {
		event.Extra[extraKey] = extraValue
	}

	// Avoid leaking the machine's hostname by injecting the literal "<redacted>".
	// Otherwise, sentry.Client.Capture will see an empty ServerName field and
	// automatically fill in the machine's hostname.
	event.ServerName = "<redacted>"

	tags := map[string]string{
		"report_type": "error",
	}
	for key, value := range tags {
		event.Tags[key] = value
	}
	eventID := hub.CaptureEvent(event)
	// eventID := errors.ReportError(err)
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
