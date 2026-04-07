package main

import (
	"github.com/illoprin/retro-fps-kit-go/pkg/app"
	"github.com/illoprin/retro-fps-kit-go/pkg/kit/states/game"
)

func main() {
	e, err := app.NewApp()
	if err != nil {
		panic(err)
	}
	defer e.Destroy()

	s := game.NewSectorGameState()
	if err := s.Init(e); err != nil {
		panic(err)
	}
	e.SetActiveState(s)

	// s := demo.NewDemo()
	// if err := s.Init(e); err != nil {
	// 	panic(err)
	// }
	// e.SetActiveState(s)

	e.Run()
}
