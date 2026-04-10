package main

import (
	"github.com/illoprin/retro-fps-toolkit-go/pkg/app"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/app/config"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/states/game"
)

func main() {

	config, err := config.LoadConfig("config.yaml")
	if err != nil {
		panic(err)
	}

	e, err := app.NewApp(config)
	if err != nil {
		panic(err)
	}
	defer e.Destroy()

	s := game.NewGameState()
	if err := s.Init(e); err != nil {
		panic(err)
	}
	e.SetActiveState(s)

	// s := demo.NewDemoState()
	// if err := s.Init(e); err != nil {
	// 	panic(err)
	// }
	// e.SetActiveState(s)

	e.Run()
}
