package game

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/app"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/render/rhi"
)

type GameState struct {
	api       app.AppAPI
	resources []rhi.Resource
}

func NewSectorGameState() *GameState {
	return &GameState{
		resources: make([]rhi.Resource, 0),
	}
}

func (g *GameState) Init(e app.AppAPI) error {
	g.api = e

	return nil
}

func (g *GameState) Update(deltaTime float32) {

}

func (g *GameState) OnKey(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {

}

func (g *GameState) Destroy() {

}
