package game

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/illoprin/retro-fps-kit-go/pkg/app"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
)

type SectorGameState struct {
	api       app.AppAPI
	resources []rhi.Resource
}

func NewSectorGameState() *SectorGameState {
	return &SectorGameState{
		resources: make([]rhi.Resource, 0),
	}
}

func (g *SectorGameState) Init(e app.AppAPI) error {
	g.api = e

	return nil
}

func (g *SectorGameState) Update(deltaTime float32) {

}

func (g *SectorGameState) OnKey(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {

}

func (g *SectorGameState) Destroy() {

}
