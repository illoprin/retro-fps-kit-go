package main

import (
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/app"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/app/config"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/states/demo"
)

func main() {

	config, _ := config.LoadConfig("config.yaml")

	e, err := app.NewApp(config)
	if err != nil {
		panic(err)
	}
	defer e.Destroy()

	// s := game.NewGameState()
	// if err := s.Init(e); err != nil {
	// 	panic(err)
	// }
	// e.SetActiveState(s)

	s := demo.NewDemoState()
	if err := s.Init(e); err != nil {
		panic(err)
	}
	e.SetActiveState(s)

	e.Run()
}
