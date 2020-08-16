package middlewares

import (
	"context"
	"expvar"
	"net/http"
	"runtime"

	"github.com/zue666/von"
	"go.opentelemetry.io/otel/api/global"
)

// m contains the global program counters for the application.
var m = struct {
	gr  *expvar.Int
	req *expvar.Int
	err *expvar.Int
}{
	gr:  expvar.NewInt("goroutines"),
	req: expvar.NewInt("requests"),
	err: expvar.NewInt("errors"),
}

// Metrics updates program counters.
func Metrics() von.Middleware {
	m := func(before von.Handler) von.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx, span := global.Tracer("von").Start(ctx, "von.middlewares.Metrics")
			defer span.End()

			err := before(ctx, w, r)
			m.req.Add(1)
			if m.req.Value()%100 == 0 {
				m.gr.Set(int64(runtime.NumGoroutine()))
			}
			if err != nil {
				m.err.Add(1)
			}
			return err
		}
		return h
	}

	return m
}
