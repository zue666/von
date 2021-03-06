package middlewares

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/zue666/von"
	"go.opentelemetry.io/otel/trace"
)

// Logger logs the request information:
// TraceID : (200) : GET /example -> IP ADDR (latency)
func Logger(log *log.Logger) von.Middleware {
	f := func(before von.Handler) von.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "von.middlewares.Logger")
			defer span.End()
			v, ok := ctx.Value(von.KeyValues).(*von.Values)
			if !ok {
				return von.NewShutdownError("value missing from context")
			}

			err := before(ctx, w, r)
			log.Printf("%s : (%d) : %s %s -> %s (%s)",
				v.TraceID, v.Status,
				r.Method, r.URL.Path,
				r.RemoteAddr, time.Since(v.Now))

			return err
		}

		return h
	}

	return f
}
