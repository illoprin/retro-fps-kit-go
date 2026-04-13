package game

import (
	"fmt"

	"github.com/go-gl/glfw/v3.3/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/app"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/app/controllers"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/camera"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/logger"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/rhi"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/assets/level"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/systems/level"
)

type GameState struct {
	api        app.AppAPI
	builder    *levelasset.LevelBuilder
	renderer   *levelsys.LevelRenderer
	level      *levelsys.LevelSystem
	controller *controllers.EditorController
	resources  []rhi.Resource
}

func NewGameState() *GameState {
	s := &GameState{
		resources: make([]rhi.Resource, 0),
	}

	def := &levelasset.LevelDef{}

	s.builder = levelasset.NewLevelBuilder(def)
	s.level = levelsys.NewLevelSystem(s.builder)

	return s
}

func (g *GameState) Init(e app.AppAPI) error {
	g.api = e

	def := g.builder.GetDef()

	// build level mesh
	g.builder.BuildModel()

	// init level renderer (creates textures and vao)
	var err error
	g.renderer, err = levelsys.NewLevelRenderer(g.level)
	if err != nil {
		return fmt.Errorf("failed to create level renderer %w", err)
	}

	// create contoller
	playerStart, err := def.GetPlayerStart()
	if err != nil {
		logger.Warnf("player start not defined, init with default value")
		playerStart = &levelasset.EntityDef{
			Pos: mgl.Vec3{0, 0, 0},
		}
	}
	g.controller = controllers.NewEditorController(
		e.GetInputManager(), playerStart.Pos, 0.1, 0.1,
	)

	return nil
}

func (g *GameState) Update(deltaTime float32) {
	g.level.Update(deltaTime)

	input := g.api.GetInputManager()
	window := g.api.GetWindow()

	canUpdateController := false

	// update the controller only if game mode
	//
	// or lmb pressed
	if window.GetCursorDisabled() {
		canUpdateController = true
	} else {
		if !g.api.GUIWantCaptureMouse() && input.IsMouseButtonPressed(glfw.MouseButton1) {
			canUpdateController = true
		}
	}

	if canUpdateController {
		g.controller.Update(deltaTime)
	}
}

func (g *GameState) OnKey(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {
	window := g.api.GetWindow()

	if action == glfw.Press {
		if key == glfw.KeyF8 {
			window.ToggleCursor()
		}
	}
}

func (g *GameState) RenderGBuffer() {
	g.renderer.Render()
}

func (g *GameState) GetActiveCamera() *camera.Camera3D {
	return g.controller.GetCamera()
}

func (g *GameState) Destroy() {

}
