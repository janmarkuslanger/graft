package main

import (
	"fmt"

	"github.com/janmarkuslanger/graft/graft"
	"github.com/janmarkuslanger/graft/module"
	"github.com/janmarkuslanger/graft/router"
)

type MyLogger struct{}

func (l MyLogger) Test() string {
	return "TestLogger"
}

type UserDeps struct {
	Logger MyLogger
}

func main() {

	m := module.New("User", "/user", UserDeps{
		Logger: MyLogger{},
	})
	m.AddRoute(module.Route[UserDeps]{
		Path:   "/",
		Method: "GET",
		Handler: func(ctx router.Context, deps UserDeps) {
			fmt.Println(deps.Logger.Test())
		},
	})

	app := graft.New()
	app.UseModule(m)
	app.Run()
}
