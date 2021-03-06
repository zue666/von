package von

import (
	"context"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

// ctxKey is the type of value for the context key.
type ctxKey int

// KeyValues is how request values are stored/retrieved.
const KeyValues ctxKey = 1

// ParamsKey is how params values are stored/retrieved.
const ParamsKey ctxKey = 2

// Values represent state for each request.
type Values struct {
	TraceID string
	Now     time.Time
	Status  int
}

// Handler is a type that handles an http request.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// App is the entrypoint to our http service and what configures our context object
// for each of our http handlers.
type App struct {
	*httprouter.Router
	oth      http.Handler
	shutdown chan os.Signal
	mw       []Middleware
}

// NewApp creates an App value that handle a set of routes for the application.
func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {
	router := httprouter.New()
	app := App{
		Router:   router,
		oth:      otelhttp.NewHandler(router, "request"),
		shutdown: shutdown,
		mw:       mw,
	}

	// app.Router.NotFound =
	// http.HandlerFunc(w http.ResponseWriter, r *http.Request) {

	// }

	// app.oth = othttp.NewHandler(app.Router, "von")
	return &app
}

// Handle is a mechanism for mounting Handlers  for a given HTTP verb and path pair,
// this makes for really easy convenient routing.
func (a *App) Handle(method string, path string, handler Handler, mw ...Middleware) {
	// wrap the provided middlewares
	handler = wrapMiddleware(mw, handler)
	// wrap the application middlewares
	handler = wrapMiddleware(a.mw, handler)

	h := func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// start or expand a distributed trace.
		ctx := r.Context()
		ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, r.URL.Path)
		defer span.End()

		ctx = context.WithValue(ctx, ParamsKey, &Params{params})

		// set the context with the required values to process the request.
		v := Values{
			TraceID: span.SpanContext().SpanID.String(),
			Now:     time.Now(),
		}

		// setting request's Values
		ctx = context.WithValue(ctx, KeyValues, &v)

		// call the wrapped handler functions.
		if err := handler(ctx, w, r); err != nil {
			a.SignalShutdown()
			return
		}
	}

	// a.Router.Handler(method,path,h)

	a.Router.Handle(method, path, h)
}

// SignalShutdown implements the http.Handler interface. It overrides the ServeHTTP of the embedded Router by
// using opentelemetry ServeHTTP instead.
// That handler wraps the HTTPRouter handler so the routes are served.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// a.Router.ServeHTTP(w, r)
	a.oth.ServeHTTP(w, r)
}

// SignalShutdown is used to gracefully shutdown the app when an integrity issue is identified.
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}
