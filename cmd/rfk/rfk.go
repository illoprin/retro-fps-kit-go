package main

import (
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/app"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/app/config"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/states/game"
)

func main() {

	config, _ := config.LoadConfig("config.yaml")

	a, err := app.NewApp(config)
	if err != nil {
		panic(err)
	}
	defer a.Destroy()

	s := game.NewGameState()
	if err := s.Init(a); err != nil {
		panic(err)
	}
	defer s.Destroy()
	a.SetActiveState(s)

	// s := demo.NewDemoState()
	// if err := s.Init(e); err != nil {
	// panic(err)
	// }
	// e.SetActiveState(s)

	a.Run()
}
