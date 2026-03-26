package main

import (
	"log"

	"github.com/illoprin/retro-fps-kit-go/src/engine"
)

func main() {

	e, err := engine.NewEngine()
	if err != nil {
		panic(err)
	}
	defer e.Destroy()

	if err := e.Run(); err != nil {
		log.Fatalln(err)
	}

}
