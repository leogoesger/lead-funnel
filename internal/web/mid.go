package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ardanlabs/kronk/sdk/kronk/observ/otel"
	"github.com/leogoesger/lead-funnel/internal/logger"

	"go.opentelemetry.io/otel/trace"
)

func MidTracer() MidFunc {
	m := func(next HandlerFunc) HandlerFunc {
		h := func(ctx context.Context, r *http.Request) Encoder {
			span := trace.SpanFromContext(ctx)
			tracer := span.TracerProvider().Tracer("")

			ctx = otel.SetTracer(ctx, tracer)

			return next(ctx, r)
		}

		return h
	}

	return m
}

func MidLogger(log *logger.Logger) MidFunc {
	m := func(next HandlerFunc) HandlerFunc {
		h := func(ctx context.Context, r *http.Request) Encoder {
			now := time.Now()

			path := r.URL.Path
			if r.URL.RawQuery != "" {
				path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
			}

			log.Info(ctx, "request started", "method", r.Method, "path", path, "remoteaddr", r.RemoteAddr)

			resp := next(ctx, r)

			log.Info(ctx, "request completed", "method", r.Method, "path", path, "remoteaddr", r.RemoteAddr,
				"since", time.Since(now).String())

			return resp
		}

		return h
	}

	return m
}

func MidErrors(log *logger.Logger) MidFunc {
	m := func(next HandlerFunc) HandlerFunc {
		h := func(ctx context.Context, r *http.Request) Encoder {
			resp := next(ctx, r)

			err := checkIsError(resp)
			if err == nil {
				return resp
			}

			_, span := otel.AddSpan(ctx, "app.mid.error")
			span.RecordError(err)
			defer span.End()

			// Return error as an ErrorResponse that implements Encoder
			return ErrorResponse{
				Error:  err.Error(),
				Status: http.StatusInternalServerError,
			}
		}

		return h
	}

	return m
}

func checkIsError(e Encoder) error {
	err, hasError := e.(error)
	if hasError {
		return err
	}

	return nil
}
