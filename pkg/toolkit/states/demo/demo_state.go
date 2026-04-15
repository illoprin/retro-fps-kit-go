package demo

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/glfw/v3.3/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/app"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/app/controllers"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/camera"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/files"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/rhi"
	modeldata "github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/assets/model"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/entities/prefab"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/systems/gui"
	prefabsystem "github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/systems/prefab"
	earcut "github.com/rclancey/go-earcut"
)

type DemoState struct {
	api        app.AppAPI
	prefabs    [](*prefab.Prefab)
	resources  []rhi.Resource
	renderer   *prefabsystem.PrefabRenderer
	canvas     *gui.GUICanvas
	controller *controllers.EditorController
	lastTime   time.Time
	drawGrid   bool
	drawShapes bool
	showUI     bool

	prefabEmissive *prefab.Prefab
}

func NewDemoState() *DemoState {
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

	// init gui canvas
	canvas, err := gui.NewGUICanvas()
	if err != nil {
		return fmt.Errorf("failed to create canvas - %w", err)
	}
	s.canvas = canvas
	s.resources = append(s.resources, s.canvas)

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

	// ceiling model
	emissiveModel, err := parser.ParseFile(files.GetModelPath("emissive_part.obj"))
	if err != nil {
		log.Printf("failed to load model %v", err)
	}

	// shotgun mesh
	meshShotgun := rhi.NewMesh()
	modeldata.SetupMeshFromModel(meshShotgun, shotgunModel)

	// floor mesh
	meshFloor := rhi.NewMesh()
	modeldata.SetupMeshFromModel(meshFloor, floorModel)

	// ceiling mesh
	meshCeiling := rhi.NewMesh()
	modeldata.SetupMeshFromModel(meshCeiling, ceilingModel)

	// walls mesh
	meshWalls := rhi.NewMesh()
	modeldata.SetupMeshFromModel(meshWalls, wallsModel)

	// walls mesh
	meshTable := rhi.NewMesh()
	modeldata.SetupMeshFromModel(meshTable, tableModel)

	// emissive part mesh
	meshEmissive := rhi.NewMesh()
	modeldata.SetupMeshFromModel(meshEmissive, emissiveModel)

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

	// tiles texture
	texEmissive, err := rhi.NewTextureFromImage(files.GetTexturePath("emissive.png"), texConfig)
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
		meshEmissive,
		texColors,
		texBrick,
		texRock,
		texWood,
		texTiles,
		texEmissive,
	)

	s.prefabEmissive = prefab.NewPrefab(meshEmissive, texEmissive, texEmissive)

	// add prefabs
	s.prefabs = append(
		s.prefabs,
		prefab.NewPrefab(meshShotgun, texColors, nil),
		prefab.NewPrefab(meshFloor, texTiles, nil),
		prefab.NewPrefab(meshWalls, texBrick, nil),
		prefab.NewPrefab(meshCeiling, texRock, nil),
		prefab.NewPrefab(meshTable, texWood, nil),
		s.prefabEmissive,
	)

	s.createConvexMesh(texBrick)
	return nil
}

func (s *DemoState) createConvexMesh(texture *rhi.Texture) {
	// triangulate
	shift := []float32{3, 4}
	scale := []float32{2, 2}
	verts := []float64{
		// flat
		2, 3,
		3, 2,
		5, 2,
		6, 3,
		6, 8,
		5, 9,
		3, 9,
		2, 8,

		// hole 1
		4.5, 4.5,
		4.5, 6.5,
		4, 7,
		3.5, 6.5,
		3.5, 4.5,
	} // CCW
	holes := []int{8} // CCW
	indices, _ := earcut.Earcut(verts, holes, 2)

	// create model vertices
	modelVertices := make([]modeldata.ModelVertex, len(verts)/2)
	for i, _ := range modelVertices {
		vertIndex := i * 2
		x := float32(verts[vertIndex])*scale[0] - shift[0]
		z := float32(verts[vertIndex+1])*scale[1] - shift[1]
		modelVertices[i] = modeldata.ModelVertex{
			X:  x,
			Y:  1.5,
			Z:  z,
			U:  x,
			V:  z,
			Nx: 0,
			Ny: 1,
			Nz: 0,
		}
	}

	// create model indices
	modelIndices := make([]uint32, len(indices))
	for i, _ := range indices {
		modelIndices[i] = uint32(indices[len(indices)-i-1])
	}

	// setup model
	model := modeldata.Model{
		Vertices: modelVertices,
		Indices:  modelIndices,
	}
	mesh := rhi.NewMesh()
	modeldata.SetupMeshFromModel(mesh, &model)

	// create prefab
	s.prefabs = append(s.prefabs,
		prefab.NewPrefab(mesh, texture, nil),
	)

	s.resources = append(s.resources, mesh)
}

func (s *DemoState) Update(deltaTime float32) {

	// time

	elapsed := time.Since(s.lastTime)

	if elapsed >= time.Second {
		s.lastTime = time.Now()
	}

	// prefabs

	shotgun := s.prefabs[0]
	shotgun.Position = mgl.Vec3{0, 1.446, 0}
	shotgun.Scaling = mgl.Vec3{0.25, 0.25, 0.25}
	shotgun.Rotation[1] = -90

	// input

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

	// gui

	// clear canvas
	s.canvas.Clear()

	// demo shapes
	if s.drawShapes {
		t := s.api.GetTime()
		radiusDelta := .1
		radiusBase := .25
		radius := radiusBase + math.Sin(t)*(radiusDelta/2)

		s.canvas.Circle(mgl.Vec2{-.25, -.25}, float32(radius), mgl.Vec4{0.89, 0.706, 0.102, .5}, 16)
		s.canvas.Rect(mgl.Vec2{.25, -.25}, mgl.Vec2{.48, .25}, mgl.Vec4{0.161, 0.486, 0.878, 0.9})
		s.canvas.Line(mgl.Vec2{-.5, .7}, mgl.Vec2{.5, .6}, mgl.Vec4{0.929, 0.059, 0.361, .7}, 0.02)
	}

	// crosshair dot
	s.canvas.Circle(mgl.Vec2{0, 0}, 0.005, mgl.Vec4{1, 1, 1, .8}, 4)

	s.canvas.Update()
}

func (s *DemoState) RenderGBuffer() {
	defaultAssets := s.api.GetDefaultAssets()

	// render scene
	winConfig := s.api.GetWindow().GetConfig()
	s.renderer.Prepare(
		int(winConfig.Width),
		int(winConfig.Height),
		s.controller.GetCamera(),
		float32(s.api.GetTime()),
	)
	for _, p := range s.prefabs {
		s.renderer.Render(p)
	}
	if s.drawGrid {
		defaultAssets.DrawGrid(s.controller.GetCamera(), 1, 1.0, 10.0)
	}
}

func (s *DemoState) ShowImgui() {
	if s.showUI {
		imgui.Begin("Scene")

		imgui.SliderFloat("Emissive Strength", &s.prefabEmissive.EmissiveStrength, 0.0, 100.0)
		imgui.Checkbox("Draw Grid", &s.drawGrid)
		imgui.Checkbox("Dithering", &s.renderer.Dithering)
		imgui.Checkbox("2D Shapes", &s.drawShapes)

		imgui.End()
	}
}

func (s *DemoState) RenderFlat(_ *rhi.Framebuffer) {
	aspect := s.api.GetWindow().GetConfig().Aspect
	s.canvas.Draw(aspect)
}

func (s *DemoState) GetCamera() *camera.Camera3D {
	s.api.GetWindow().GetConfig()
	return s.controller.GetCamera()
}

func (s *DemoState) OnKey(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {
	window := s.api.GetWindow()

	if action == glfw.Press {
		if key == glfw.KeyF8 {
			window.ToggleCursor()
		}
		if key == glfw.KeyF2 {
			s.showUI = !s.showUI
		}
	}
}

func (s *DemoState) Destroy() {
	s.renderer.Delete()
	for _, r := range s.resources {
		if r != nil {
			r.Delete()
		}
	}
}
