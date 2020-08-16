# Von
Von is a web platform built to simplify deploying web services. Von uses [httprouter](https://github.com/julienschmidt/httprouter) to route HTTP requests.

## Using Von
First import `Von`. Von uses Go Modules: ```GO111MODULE=on go get -d github.com/zue666/von```

```
package main

func main() {
    shutdown := make(chan os.Signal,1)
    signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
    app := von.New(shutdown)
    app.Handle(http.MethodGet, "/", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
        data := map[string]string{"message": "Hello, World!"}

        return von.Respond(ctx,w,data,http.StatusOK)
    })

    server := http.Server{
        Addr: os.Getenv("HTTP_ADDRESS"),
        Handler: app,
    }

    log.Fatal(server.ListenAndServe())
}
```

## Middlewares
A middleware is wrapper around von.Handler, you can hookup a middleware at the app level or at the handler level

A sample middleware:
```
func GreeterMiddleware(log *log.Logger) von.Handler {
    return func(before von.Handler) von.Handler {
        handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
            err := before(ctx, w, r)
            log.Println("hello from GreeterMiddleware")
            return err
        }
    }
}
```
### App level hookup:
```
log := log.New(os.Stdout, "app ", log.Lstdflags)
app := von.New(shutdown, GreeterMiddleware(log))
...
```

### Handler level hookup:
```
log := log.New(os.Stdout, "app ", log.Lstdflags)

handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
    return von.Respond(ctx, w, "greeting", http.StatusOK)
}

app := von.New(shutdown)
app.Handle(http.MethodGet, "/greet", handler, GreeterMiddleware(log))
...
```
