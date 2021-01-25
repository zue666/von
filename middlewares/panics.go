package middlewares

import (
	"context"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/pkg/errors"
	"github.com/zue666/von"
	"go.opentelemetry.io/otel/trace"
)

// Panics recovers from panics and converts the panic to an error so it is reported
// in metrics and handled in Errors.
func Panics(log *log.Logger) von.Middleware {
	m := func(after von.Handler) von.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
			ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "von.middlewares.Panics")
			defer span.End()

			v, ok := ctx.Value(von.KeyValues).(*von.Values)
			if !ok {
				return von.NewShutdownError("value missing from context")
			}

			defer func() {
				if r := recover(); r != nil {
					err = errors.Errorf("panic : %v", r)

					log.Printf("%s :\n%s", v.TraceID, debug.Stack())
				}
			}()

			return after(ctx, w, r)
		}
		return h
	}
	return m
}
