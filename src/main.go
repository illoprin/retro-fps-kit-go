package main

import (
	"github.com/illoprin/retro-fps-kit-go/src/engine"
)

func main() {

	e, err := engine.NewEngine()
	if err != nil {
		panic(err)
	}
	defer e.Destroy()

	e.Run()

}
