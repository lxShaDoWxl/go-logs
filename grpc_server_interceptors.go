package logs

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/getsentry/sentry-go"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_tags "github.com/grpc-ecosystem/go-grpc-middleware/tags"

	"google.golang.org/grpc"
)

func recoverWithSentry(ctx context.Context) {
	if err := recover(); err != nil {
		FatalError(ctx, errors.Wrap(err, 2))
	}
}

func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	o := newConfig(opts)
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.CurrentHub().Clone()
			ctx = sentry.SetHubOnContext(ctx, hub)
		}

		span := sentry.StartSpan(ctx, "grpc.server")
		defer span.Finish()
		span.Description = info.FullMethod
		// TODO: Perhaps makes sense to use SetRequestBody instead?
		hub.Scope().SetExtra("requestBody", req)
		defer recoverWithSentry(ctx)

		resp, err := handler(ctx, req)
		o.ReportOn(err)

		return resp, err
	}
}

func StreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor {
	o := newConfig(opts)
	return func(srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler) error {
		ctx := ss.Context()
		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.CurrentHub().Clone()
			ctx = sentry.SetHubOnContext(ctx, hub)
		}

		span := sentry.StartSpan(ctx, "grpc.server")
		defer span.Finish()

		stream := grpc_middleware.WrapServerStream(ss)
		stream.WrappedContext = ctx
		span.Description = info.FullMethod
		defer recoverWithSentry(ctx)

		err := handler(srv, stream)
		if err != nil && o.ReportOn(err) {
			tags := grpc_tags.Extract(ctx)
			for k, v := range tags.Values() {
				hub.Scope().SetTag(k, v.(string))
			}

			hub.CaptureException(err)
		}

		return err
	}
}
