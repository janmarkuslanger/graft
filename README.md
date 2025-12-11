<p align="center"><img src="/logo.svg" alt="GRAFT Logo" /></p>

---

<p align="center">
  <a href="https://codecov.io/gh/janmarkuslanger/graft"><img src="https://codecov.io/gh/janmarkuslanger/graft/graph/badge.svg?token=SY8BCTMFEL" alt="Code coverage"></a>
  <a href="https://goreportcard.com/report/github.com/janmarkuslanger/graft"><img src="https://goreportcard.com/badge/github.com/janmarkuslanger/graft" alt="Go Report"></a>
  <a href="https://github.com/janmarkuslanger/graft/releases"><img src="https://img.shields.io/github/release/janmarkuslanger/graft" alt="Latest Release"></a>
  <a href="https://github.com/janmarkuslanger/graft/actions"><img src="https://github.com/janmarkuslanger/graft/actions/workflows/ci.yml/badge.svg" alt="Build Status"></a>
  <a href="https://github.com/janmarkuslanger/graft/archive/refs/heads/main.zip"><img src="https://img.shields.io/badge/Download-ZIP-blue" alt="Download ZIP"></a>
</p>

---

Graft is a tiny, opinionated HTTP toolkit that helps you structure Go services around modules. It keeps the standard library ergonomics you already know, adds a light router wrapper, and lets every module declare typed dependencies so the compiler has your back.

## Features

- Modular route definitions with strongly typed dependency injection
- Middleware chaining that still looks and feels like `net/http`
- Helpers for trailing slashes and serving static assets
- Batteries-included testability using Go's `httptest` helpers

## Installation

```sh
go get github.com/janmarkuslanger/graft@latest
```

You can then import the packages you need:

```go
import (
    "github.com/janmarkuslanger/graft/graft"
    "github.com/janmarkuslanger/graft/module"
    "github.com/janmarkuslanger/graft/router"
)
```

## Quick Start

Create a module, register it on the app, run the server:

```go
package main

import (
    "net/http"

    "github.com/janmarkuslanger/graft/graft"
    "github.com/janmarkuslanger/graft/module"
    "github.com/janmarkuslanger/graft/router"
)

type Deps struct {
    Greeter func() string
}

func main() {
    app := graft.New()

    hello := module.Module[Deps]{
        Name:     "hello",
        BasePath: "/hello",
        Deps: Deps{
            Greeter: func() string { return "hello from graft" },
        },
        Routes: []module.Route[Deps]{
            {
                Method:  http.MethodGet,
                Path:    "/",
                Handler: func(ctx router.Context, deps Deps) {
                    ctx.Writer.Write([]byte(deps.Greeter()))
                },
            },
        },
    }

    app.UseModule(&hello)
    app.Run() // serves on :8080
}
```

## Working With Modules

- `module.Module[T]` is generic over the dependency type you expect in handlers.
- `Routes` holds the HTTP verb, relative path, and handler.
- `Middlewares` lets you apply router middlewares to every route.
- Handlers receive both the request context (`router.Context`) and your typed dependency value.

Example with dependencies and middleware:

```go
type AuthDeps struct {
    Users    UserService
    Sessions SessionService
}

auth := module.Module[AuthDeps]{
	Name:     "auth",
	BasePath: "/auth",
	Deps: AuthDeps{Users: users, Sessions: sessions},
	Middlewares: []router.Middleware{
        requestLogger,
    },
    Routes: []module.Route[AuthDeps]{
        {
            Method: http.MethodPost,
            Path:   "/login",
            Handler: func(ctx router.Context, deps AuthDeps) {
                token, err := deps.Sessions.Login(ctx.Request.Context())
                if err != nil {
                    http.Error(ctx.Writer, err.Error(), http.StatusUnauthorized)
                    return
                }
                ctx.Writer.Write([]byte(token))
            },
        },
	},
}
```

Hooks let a module prepare dependencies or kick off background work:

```go
auth := module.Module[AuthDeps]{
    // ...
    Hooks: module.Hooks[AuthDeps]{
        OnUse: func(deps *AuthDeps) {
            deps.Users = users.WithCache()
        },
        OnStart: func(deps *AuthDeps) {
            log.Println("auth module ready", deps.Users.Count())
        },
    },
}
```

`OnUse` runs when the module is registered with `UseModule`. `OnStart` runs right before the server starts, after all modules have been registered. Both receive a pointer to your dependency struct so you can update it in-place.

## Router Toolbox

- `router.New()` returns a wrapper around `http.ServeMux`.
- `AddHandler("GET /ping", handler, middleware...)` registers method-aware routes.
- `Handle` automatically registers both `/path` and `/path/` variants.
- `Static(prefix, dir, middleware...)` is a convenience for `http.FileServer`.
- Middleware signature: `func(ctx router.Context, next router.HandlerFunc)`.

Static assets in one line:

```go
r := router.New()
r.Static("/assets", "./public", loggingMiddleware)
```

## Running & Testing

- `graft.Run()` listens on `:8080` by default.
- Use Go's standard tools: `go test ./...` runs the full suite; `go test ./... -coverprofile=coverage.txt` mirrors the CI setup and feeds Codecov.
- The router and modules are intentionally `httptest` friendly—spin up a router, register modules, and exercise handlers directly.

## Contributing

Bug reports, feature ideas, and pull requests are always welcome. Please format code (`gofmt`) and run the tests before submitting. If you plan a larger change, open an issue first so we can figure out the best direction together.

The project is MIT licensed—see `LICENSE` for the legal bits.
