package game

import (
	"fmt"
	"log"

	"github.com/go-gl/glfw/v3.3/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-kit-go/pkg/app"
	"github.com/illoprin/retro-fps-kit-go/pkg/app/controllers"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/files"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/model"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/pipeline"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
	"github.com/illoprin/retro-fps-kit-go/pkg/scene/prefab"
)

type DemoState struct {
	api        app.AppAPI
	prefabs    [](*prefab.Prefab)
	resources  []rhi.Resource
	renderer   *pipeline.PrefabRenderer
	controller *controllers.EditorController
}

func NewGame() *DemoState {
	return &DemoState{
		resources: make([]rhi.Resource, 0),
		prefabs:   make([]*prefab.Prefab, 0),
	}
}

func (g *DemoState) Init(api app.AppAPI) error {
	g.api = api

	g.controller = controllers.NewEditorController(
		api.GetInputManager(), mgl.Vec3{0, 0, 3}, 10.5, 0.25,
	)

	// init prefab renderer
	renderer, err := pipeline.NewPrefabRenderer()
	if err != nil {
		return fmt.Errorf("failed to create prefab renderer - %w", err)
	} else {
		g.renderer = renderer
	}

	// create obj parser
	parser := model.NewOBJParser()

	// shotgun model
	shotgunModel, err := parser.ParseFile(files.GetModelPath("shotgun.obj"))
	if err != nil {
		log.Printf("failed to load model %v", err)
	}

	// floor model
	floorModel, err := parser.ParseFile(files.GetModelPath("floor.obj"))
	if err != nil {
		log.Printf("failed to load model %v", err)
	}

	// walls model
	wallsModel, err := parser.ParseFile(files.GetModelPath("walls.obj"))
	if err != nil {
		log.Printf("failed to load model %v", err)
	}

	// ceiling model
	ceilingModel, err := parser.ParseFile(files.GetModelPath("ceiling.obj"))
	if err != nil {
		log.Printf("failed to load model %v", err)
	}

	// ceiling model
	tableModel, err := parser.ParseFile(files.GetModelPath("table.obj"))
	if err != nil {
		log.Printf("failed to load model %v", err)
	}

	// shotgun mesh
	meshShotgun := rhi.NewMesh()
	meshShotgun.SetupFromModel(shotgunModel, rhi.StaticDraw)

	// floor mesh
	meshFloor := rhi.NewMesh()
	meshFloor.SetupFromModel(floorModel, rhi.StaticDraw)

	// ceiling mesh
	meshCeiling := rhi.NewMesh()
	meshCeiling.SetupFromModel(ceilingModel, rhi.StaticDraw)

	// walls mesh
	meshWalls := rhi.NewMesh()
	meshWalls.SetupFromModel(wallsModel, rhi.StaticDraw)

	// walls mesh
	meshTable := rhi.NewMesh()
	meshTable.SetupFromModel(tableModel, rhi.StaticDraw)

	// colors texture
	texColors, err := rhi.NewTextureFromImage(files.GetTexturePath("colors.png"), true, true)
	if err != nil {
		log.Printf("failed to load texture %v", err)
	}

	// brick texture
	texBrick, err := rhi.NewTextureFromImage(files.GetTexturePath("dark_brick.png"), true, true)
	if err != nil {
		log.Printf("failed to load texture %v", err)
	}

	// rock texture
	texRock, err := rhi.NewTextureFromImage(files.GetTexturePath("gray_rock.png"), true, true)
	if err != nil {
		log.Printf("failed to load texture %v", err)
	}

	// rock texture
	texWood, err := rhi.NewTextureFromImage(files.GetTexturePath("wood.png"), true, true)
	if err != nil {
		log.Printf("failed to load texture %v", err)
	}

	// tiles texture
	texTiles, err := rhi.NewTextureFromImage(files.GetTexturePath("gray_tiles.png"), true, true)
	if err != nil {
		log.Printf("failed to load texture %v", err)
	}

	// add resources
	g.resources = append(
		g.resources,
		meshShotgun,
		meshFloor,
		meshWalls,
		meshCeiling,
		meshTable,
		texColors,
		texBrick,
		texRock,
		texWood,
		texTiles,
	)

	// add prefabs
	g.prefabs = append(
		g.prefabs,
		prefab.NewPrefab(meshShotgun, texColors),
		prefab.NewPrefab(meshFloor, texTiles),
		prefab.NewPrefab(meshWalls, texBrick),
		prefab.NewPrefab(meshCeiling, texRock),
		prefab.NewPrefab(meshTable, texWood),
	)

	return nil
}

func (g *DemoState) Update(deltaTime float32) {
	shotgun := g.prefabs[0]
	shotgun.Position = mgl.Vec3{0, 1.446, 0}
	shotgun.Scaling = mgl.Vec3{0.25, 0.25, 0.25}
	shotgun.Rotation[1] = -90

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

func (g *DemoState) RenderGeometry() {
	// render scene
	winConfig := g.api.GetWindow().GetConfig()
	g.renderer.Prepare(
		int(winConfig.Width),
		int(winConfig.Height),
		g.controller.GetCamera(),
	)
	for _, p := range g.prefabs {
		g.renderer.Render(p)
	}
}

func (g *DemoState) OnKey(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {
	window := g.api.GetWindow()

	if action == glfw.Press {
		window.ToggleCursor()
	}

}

func (g *DemoState) OnMouseButton(button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {

}

func (g *DemoState) OnMouseMove(dX, dY, posX, posY float64) {

}

func (g *DemoState) OnMouseScroll(dx, dy float64) {

}

func (g *DemoState) HasFPSController() bool {
	return true
}

func (g *DemoState) OnResize(w, h, sw, sh int32) {

}

func (g *DemoState) Destroy() {
	g.renderer.Delete()
	for _, r := range g.resources {
		if r != nil {
			r.Delete()
		}
	}
}
