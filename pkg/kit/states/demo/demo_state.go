package demo

import (
	"fmt"
	"log"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-kit-go/pkg/app"
	"github.com/illoprin/retro-fps-kit-go/pkg/app/controllers"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/camera"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/files"
	modeldata "github.com/illoprin/retro-fps-kit-go/pkg/kit/assets/model"
	"github.com/illoprin/retro-fps-kit-go/pkg/kit/entities/prefab"
	prefabsystem "github.com/illoprin/retro-fps-kit-go/pkg/kit/systems/prefab"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
)

type DemoState struct {
	api        app.AppAPI
	prefabs    [](*prefab.Prefab)
	resources  []rhi.Resource
	renderer   *prefabsystem.PrefabRenderer
	controller *controllers.EditorController
	lastTime   time.Time
}

func NewDemo() *DemoState {
	return &DemoState{
		resources: make([]rhi.Resource, 0),
		prefabs:   make([]*prefab.Prefab, 0),
		lastTime:  time.Now(),
	}
}

func (s *DemoState) Init(api app.AppAPI) error {
	s.api = api

	s.controller = controllers.NewEditorController(
		api.GetInputManager(), mgl.Vec3{0, 2, 3}, 10.5, 0.1,
	)

	// init prefab renderer
	renderer, err := prefabsystem.NewPrefabRenderer()
	if err != nil {
		return fmt.Errorf("failed to create prefab renderer - %w", err)
	} else {
		s.renderer = renderer
	}

	// create obj parser
	parser := modeldata.NewOBJParser()

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
	modeldata.SetupMeshFromModel(meshShotgun, shotgunModel)

	// floor mesh
	meshFloor := rhi.NewMesh()
	modeldata.SetupMeshFromModel(meshShotgun, floorModel)

	// ceiling mesh
	meshCeiling := rhi.NewMesh()
	modeldata.SetupMeshFromModel(meshShotgun, ceilingModel)

	// walls mesh
	meshWalls := rhi.NewMesh()
	modeldata.SetupMeshFromModel(meshShotgun, wallsModel)

	// walls mesh
	meshTable := rhi.NewMesh()
	modeldata.SetupMeshFromModel(meshShotgun, tableModel)

	texConfig := rhi.DefaultTexture2DConfig(0, 0)

	// colors texture
	texColors, err := rhi.NewTextureFromImage(files.GetTexturePath("colors.png"), texConfig)
	if err != nil {
		log.Printf("failed to load texture %v", err)
	}

	// brick texture
	texBrick, err := rhi.NewTextureFromImage(files.GetTexturePath("dark_brick.png"), texConfig)
	if err != nil {
		log.Printf("failed to load texture %v", err)
	}

	// rock texture
	texRock, err := rhi.NewTextureFromImage(files.GetTexturePath("gray_rock.png"), texConfig)
	if err != nil {
		log.Printf("failed to load texture %v", err)
	}

	// rock texture
	texWood, err := rhi.NewTextureFromImage(files.GetTexturePath("wood.png"), texConfig)
	if err != nil {
		log.Printf("failed to load texture %v", err)
	}

	// tiles texture
	texTiles, err := rhi.NewTextureFromImage(files.GetTexturePath("gray_tiles.png"), texConfig)
	if err != nil {
		log.Printf("failed to load texture %v", err)
	}

	// add resources
	s.resources = append(
		s.resources,
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
	s.prefabs = append(
		s.prefabs,
		prefab.NewPrefab(meshShotgun, texColors),
		prefab.NewPrefab(meshFloor, texTiles),
		prefab.NewPrefab(meshWalls, texBrick),
		prefab.NewPrefab(meshCeiling, texRock),
		prefab.NewPrefab(meshTable, texWood),
	)

	return nil
}

func (s *DemoState) Update(deltaTime float32) {

	elapsed := time.Since(s.lastTime)

	if elapsed >= time.Second {
		s.lastTime = time.Now()
	}

	shotgun := s.prefabs[0]
	shotgun.Position = mgl.Vec3{0, 1.446, 0}
	shotgun.Scaling = mgl.Vec3{0.25, 0.25, 0.25}
	shotgun.Rotation[1] = -90

	input := s.api.GetInputManager()
	window := s.api.GetWindow()

	canUpdateController := false

	// update the controller only if game mode
	//
	// or lmb pressed
	if window.GetCursorDisabled() {
		canUpdateController = true
	} else {
		if !s.api.GUIWantCaptureMouse() && input.IsMouseButtonPressed(glfw.MouseButton1) {
			canUpdateController = true
		}
	}

	if canUpdateController {
		s.controller.Update(deltaTime)
	}
}

func (s *DemoState) RenderGBuffer() {
	// render scene
	winConfig := s.api.GetWindow().GetConfig()
	s.renderer.Prepare(
		int(winConfig.Width),
		int(winConfig.Height),
		s.controller.GetCamera(),
	)
	for _, p := range s.prefabs {
		s.renderer.Render(p)
	}
}

func (s *DemoState) GetCamera() *camera.Camera3D {
	return s.controller.GetCamera()
}

func (s *DemoState) OnKey(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {
	window := s.api.GetWindow()

	if action == glfw.Press {
		if key == glfw.KeyF8 {
			window.ToggleCursor()
		}
	}
}

func (s *DemoState) OnMouseButton(button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {

}

func (s *DemoState) OnMouseMove(dX, dY, posX, posY float64) {

}

func (s *DemoState) OnMouseScroll(dx, dy float64) {

}

func (s *DemoState) OnResize(w, h, sw, sh int32) {

}

func (s *DemoState) Destroy() {
	s.renderer.Delete()
	for _, r := range s.resources {
		if r != nil {
			r.Delete()
		}
	}
}
