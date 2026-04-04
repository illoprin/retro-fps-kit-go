package main

import (
	"github.com/illoprin/retro-fps-kit-go/pkg/app"
	"github.com/illoprin/retro-fps-kit-go/pkg/states/demo"
)

func main() {
	e, err := app.NewApp()
	if err != nil {
		panic(err)
	}
	defer e.Destroy()

	demo := demo.NewDemo()
	if demo.Init(e) != nil {
		panic(err)
	}

	e.SetActiveState(demo)

	e.Run()
}
