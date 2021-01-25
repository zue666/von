package von

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
)

// Respond converts a Go value to JSON and sends it to the client.
func Respond(ctx context.Context, w http.ResponseWriter, data interface{}, statusCode int) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "von.respond")
	defer span.End()

	v, ok := ctx.Value(KeyValues).(*Values)
	if !ok {
		return NewShutdownError("web value missing from context")
	}

	v.Status = statusCode

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}

// RespondError sends an error response back to the client
func RespondError(ctx context.Context, w http.ResponseWriter, err error) error {
	if webErr, ok := errors.Cause(err).(*Error); ok {
		er := ErrorResponse{
			Error:  webErr.Err.Error(),
			Fields: webErr.Fields,
		}

		if err := Respond(ctx, w, er, webErr.Status); err != nil {
			return err
		}

		return nil
	}

	er := ErrorResponse{
		Error: http.StatusText(http.StatusInternalServerError),
	}

	if err := Respond(ctx, w, er, http.StatusInternalServerError); err != nil {
		return err
	}

	return nil
}
