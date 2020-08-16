package von

// Middleware is a function designed to run code before and/or after another Handler.
// It is designed to remove boilerplate or other concerns not direct to any given handler
type Middleware func(Handler) Handler

// wrapMiddleware creates a new Handler by wrapping the given middleware around a final handler.
// The middlewares' Handlers will be executed by the request in the order the were proviede.
func wrapMiddleware(mw []Middleware, handler Handler) Handler {
	// Loop backwards ensures that the first provided middleware is the first to be exeucted
	for i := len(mw) - 1; i >= 0; i-- {
		h := mw[i]
		if h != nil {
			handler = h(handler)
		}
	}
	return handler
}
