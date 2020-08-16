package middlewares

import (
	"context"
	"log"
	"net/http"

	"github.com/zue666/von"
	"go.opentelemetry.io/otel/api/global"
)

// Errors handles errors coming out of the handler call chanin.
// It detects normal application errors which are used to respond to the client in a uniform way.
//  Unexpected errors (status >= 500) are logged.
func Errors(log *log.Logger) von.Middleware {
	m := func(before von.Handler) von.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx, span := global.Tracer("von").Start(ctx, "von.middlewares.Errors")
			defer span.End()

			v, ok := ctx.Value(von.KeyValues).(*von.Values)
			if !ok {
				return von.NewShutdownError("value missing from context")
			}

			if err := before(ctx, w, r); err != nil {
				log.Printf("%s : ERROR : %v", v.TraceID, err)
				if err := von.RespondError(ctx, w, err); err != nil {
					return err
				}
				if ok := von.IsShutdown(err); ok {
					return err
				}
			}
			return nil
		}
		return h
	}

	return m
}
