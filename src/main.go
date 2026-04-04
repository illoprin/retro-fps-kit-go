package main

import (
	"github.com/illoprin/retro-fps-kit-go/src/engine"
	"github.com/illoprin/retro-fps-kit-go/src/states/game"
)

func main() {

	e, err := engine.NewEngine()
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
