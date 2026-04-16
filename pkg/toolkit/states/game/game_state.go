package game

import (
	"fmt"

	"github.com/go-gl/glfw/v3.3/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/app"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/camera"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/logger"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/math"
	leveldata "github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/assets/level"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/entities/player"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/systems/gui"
	levelsys "github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/systems/level"
)

type GameState struct {
	api      app.AppAPI
	builder  *leveldata.LevelBuilder
	renderer *levelsys.LevelRenderer
	level    *levelsys.Level
	fps      *player.FPSController
	canvas   *gui.GUICanvas
}

func NewGameState() *GameState {
	s := &GameState{
		builder: leveldata.NewLevelBuilder(&demoLevel),
	}

	s.level = levelsys.NewLevelSystem(s.builder)

	return s
}

func (g *GameState) Init(a app.AppAPI) error {
	g.api = a

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
		playerStart = &leveldata.EntityDef{
			Pos: mgl.Vec3{0, 0, 0},
		}
	}
	g.fps = player.NewFPSController(
		g.api.GetInputManager(),
		playerStart.Pos,
		playerStart.Rot[1],
		0.1,
	)

	// create gui canvas
	canvas, err := gui.NewGUICanvas()
	if err != nil {
		return err
	}
	canvas.Circle(mgl.Vec2{0, 0}, 0.005, mgl.Vec4{1.0, 1.0, 1.0, 1.0}, 4)
	canvas.Update()
	g.canvas = canvas

	return nil
}

func (g *GameState) Update(deltaTime float32) {

	input := g.api.GetInputManager()
	window := g.api.GetWindow()

	canProcessPlayerControls := false

	// update the controller only if game mode
	//
	// or lmb pressed
	if window.GetCursorDisabled() {
		canProcessPlayerControls = true
	} else {
		if !g.api.GUIWantCaptureMouse() && input.IsMouseButtonPressed(glfw.MouseButton1) {
			canProcessPlayerControls = true
		}
	}
	if canProcessPlayerControls {
		g.fps.ProcessInput(deltaTime)
	}

	g.fps.Update(deltaTime)
	g.level.Update(deltaTime)
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
	scrConfig := g.api.GetWindow().GetConfig()

	g.renderer.Render(
		int(scrConfig.Width),
		int(scrConfig.Height),
		g.GetCamera(),
		g.api.GetDeferredRenderTarget().Wireframe,
	)
	g.api.GetDefaultAssets().DrawGrid(g.fps.GetCamera(), math.Epsilon, 1, 10)
}

func (g *GameState) GetCamera() *camera.Camera3D {
	return g.fps.GetCamera()
}

func (g *GameState) RenderFlat() {
	aspect := g.api.GetWindow().GetConfig().Aspect
	g.canvas.Render(aspect)
}

func (g *GameState) Destroy() {
	g.renderer.Delete()
	g.canvas.Delete()
}
