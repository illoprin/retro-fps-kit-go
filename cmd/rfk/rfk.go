package main

import (
	"github.com/illoprin/retro-fps-kit-go/pkg/app"
	"github.com/illoprin/retro-fps-kit-go/pkg/states/game"
)

func main() {
	e, err := app.NewApp()
	if err != nil {
		panic(err)
	}
	defer e.Destroy()

	game := game.NewGame()
	if game.Init(e) != nil {
		panic(err)
	}

	e.SetActiveState(game)

	e.Run()
}
