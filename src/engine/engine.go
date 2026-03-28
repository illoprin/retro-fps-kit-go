package engine

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"
	implglfw "github.com/AllenDang/cimgui-go/impl/glfw"
	implgl3 "github.com/AllenDang/cimgui-go/impl/opengl3"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-kit-go/src/engine/assetmgr"
	"github.com/illoprin/retro-fps-kit-go/src/engine/stats"
	imguimenus "github.com/illoprin/retro-fps-kit-go/src/imgui"
	"github.com/illoprin/retro-fps-kit-go/src/model"
	"github.com/illoprin/retro-fps-kit-go/src/player"
	postprocessing "github.com/illoprin/retro-fps-kit-go/src/post_processing"
	"github.com/illoprin/retro-fps-kit-go/src/render"
	"github.com/illoprin/retro-fps-kit-go/src/renderers"
	"github.com/illoprin/retro-fps-kit-go/src/scene"
	"github.com/illoprin/retro-fps-kit-go/src/window"
)

const (
	defaultWindowWidth  = 1470
	defaultWindowHeight = 710
	defaultTitle        = "Retro FPS Kit - Demo"
)

type Engine struct {
	window     *window.Window
	debugMenu  *imguimenus.DebugMenu
	input      *window.InputManager
	controller *player.EditorController
	imguiFont  *imgui.Font

	// geometry render pass
	resources      []render.Resource
	deferred       *renderers.DeferredRenderTarget
	prefabRenderer *renderers.PrefabRenderer
	prefabs        []scene.Prefab

	// post processing render pass
	ppPipeline []postprocessing.PostProcessingPass

	// screen render pass
	screen       *renderers.ScreenRenderer
	screenConfig *postprocessing.ScreenConfig
}

func NewEngine() (*Engine, error) {
	runtime.LockOSThread()

	e := &Engine{
		resources: make([]render.Resource, 0),
	}

	if err := window.InitGLFW(); err != nil {
		return nil, fmt.Errorf("failed to init glfw - %v", err)
	}

	if err := e.createWindow(); err != nil {
		return nil, fmt.Errorf("failed to init window - %v", err)
	}

	if err := gl.Init(); err != nil {
		return nil, fmt.Errorf("failed to init opengl - %v", err)
	}

	e.printUserData()

	if err := e.initImgui(); err != nil {
		return nil, fmt.Errorf("failed to init imgui context - %v", err)
	}

	if err := e.initImguiRenderer(); err != nil {
		return nil, fmt.Errorf("failed to init imgui renderer - %v", err)
	}

	if err := e.setupGame(); err != nil {
		return nil, fmt.Errorf("failed to init game - %v", err)
	}

	if err := e.setupRenderingPipeline(); err != nil {
		return nil, fmt.Errorf("failed to init rendering pipeline - %v", err)
	}

	e.initCustomImguiUI()

	return e, nil
}

func (e *Engine) createWindow() error {
	// create window
	var err error
	e.window, err = window.NewWindow(defaultWindowWidth, defaultWindowHeight, defaultTitle)
	if err != nil {
		return fmt.Errorf("failed init window %v", err)
	}
	e.window.MakeContextCurrent()
	e.window.Focus()
	e.window.Center()

	e.input = window.NewManager(e.window.Window)

	return nil
}

func (e *Engine) printUserData() {
	fmt.Println(gl.GoStr(gl.GetString(gl.RENDERER)))
	fmt.Println(gl.GoStr(gl.GetString(gl.VERSION)))
	fmt.Println(gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION)))
}

func (e *Engine) initImgui() error {

	// init imgui
	imgui.CreateContext()
	io := imgui.CurrentIO()
	io.SetConfigFlags(
		io.ConfigFlags() |
			imgui.ConfigFlagsNavEnableKeyboard |
			imgui.ConfigFlagsDockingEnable,
	)

	// load custom font
	e.imguiFont = io.Fonts().AddFontFromFileTTF(filepath.Join("assets", "fonts", "uifont.ttf"))
	if !e.imguiFont.IsLoaded() {
		return fmt.Errorf("failed to load imgui")
	}

	// dark theme
	imgui.StyleColorsDark()

	return nil
}

func (e *Engine) initImguiRenderer() error {
	// setup imgui renderer
	// a crutch to get the C pointer to GLFWWindow
	type glfwWindow struct {
		data unsafe.Pointer
	}
	ptr := (*glfwWindow)(unsafe.Pointer(e.window.Window))

	// init imgui window
	imguiWinGLFW := implglfw.NewGLFWwindowFromC(ptr.data)
	if imguiWinGLFW == nil {
		return fmt.Errorf("could not attach imgui to glfw window")
	}

	if !implglfw.InitForOpenGL(imguiWinGLFW, true) {
		return fmt.Errorf("could not init imgui renderer")
	}

	// init imgui gl renderer
	implgl3.InitV("#version 430 core")

	return nil
}

func (e *Engine) initCustomImguiUI() {

	buffersImages := []imguimenus.ImageTexture{
		{
			ID:   e.deferred.DeferredFBO.ColorTextures[0].ID,
			Name: "Color",
		},
		{
			ID:   e.deferred.DeferredFBO.ColorTextures[1].ID,
			Name: "Normal",
		},
		{
			ID:   e.deferred.DeferredFBO.DepthTexture.ID,
			Name: "Depth",
		},
	}

	// FIX
	e.debugMenu = imguimenus.NewDebugMenu(
		e.controller,
		e.window,
		e.screen.GetConfig(),
		e.screen.GetColorGrading(),
		buffersImages, nil,
	)
}

func (e *Engine) keyCallback(key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		if key == glfw.KeyEscape {
			e.window.SetShouldClose(true)
		}
		if key == glfw.KeyF8 {
			e.input.ToggleGameMode()
		}
		if key == glfw.KeyF1 {
			e.debugMenu.Visible = !e.debugMenu.Visible
		}
	}
}

func (e *Engine) setupGame() error {

	e.controller = player.NewEditorController(
		e.input, mgl32.Vec3{0, 0, 3}, 10.5, 0.085,
	)

	e.input.SetKeyCallback(e.keyCallback)

	// create obj parser
	parser := model.NewOBJParser()

	// shotgun model
	shotgunModel, err := parser.ParseFile(assetmgr.GetModelPath("shotgun.obj"))
	if err != nil {
		log.Printf("failed to load model %v", err)
	}

	// floor model
	floorModel, err := parser.ParseFile(assetmgr.GetModelPath("floor.obj"))
	if err != nil {
		log.Printf("failed to load model %v", err)
	}

	// walls model
	wallsModel, err := parser.ParseFile(assetmgr.GetModelPath("walls.obj"))
	if err != nil {
		log.Printf("failed to load model %v", err)
	}

	// ceiling model
	ceilingModel, err := parser.ParseFile(assetmgr.GetModelPath("ceiling.obj"))
	if err != nil {
		log.Printf("failed to load model %v", err)
	}

	// ceiling model
	tableModel, err := parser.ParseFile(assetmgr.GetModelPath("table.obj"))
	if err != nil {
		log.Printf("failed to load model %v", err)
	}

	// shotgun mesh
	meshShotgun := render.NewMesh()
	meshShotgun.SetupFromModel(shotgunModel, gl.STATIC_DRAW)

	// floor mesh
	meshFloor := render.NewMesh()
	meshFloor.SetupFromModel(floorModel, gl.STATIC_DRAW)

	// ceiling mesh
	meshCeiling := render.NewMesh()
	meshCeiling.SetupFromModel(ceilingModel, gl.STATIC_DRAW)

	// walls mesh
	meshWalls := render.NewMesh()
	meshWalls.SetupFromModel(wallsModel, gl.STATIC_DRAW)

	// walls mesh
	meshTable := render.NewMesh()
	meshTable.SetupFromModel(tableModel, gl.STATIC_DRAW)

	// colors texture
	texColors, err := render.NewTextureFromImage(assetmgr.GetTexturePath("colors.png"), true, true)
	if err != nil {
		log.Printf("failed to load texture %v", err)
	}

	// brick texture
	texBrick, err := render.NewTextureFromImage(assetmgr.GetTexturePath("dark_brick.png"), true, true)
	if err != nil {
		log.Printf("failed to load texture %v", err)
	}

	// rock texture
	texRock, err := render.NewTextureFromImage(assetmgr.GetTexturePath("gray_rock.png"), true, true)
	if err != nil {
		log.Printf("failed to load texture %v", err)
	}

	// rock texture
	texWood, err := render.NewTextureFromImage(assetmgr.GetTexturePath("wood.png"), true, true)
	if err != nil {
		log.Printf("failed to load texture %v", err)
	}

	// tiles texture
	texTiles, err := render.NewTextureFromImage(assetmgr.GetTexturePath("gray_tiles.png"), true, true)
	if err != nil {
		log.Printf("failed to load texture %v", err)
	}

	// add resources
	e.resources = append(
		e.resources,
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
	e.prefabs = append(
		e.prefabs,
		*scene.NewPrefab(meshShotgun, texColors),
		*scene.NewPrefab(meshFloor, texTiles),
		*scene.NewPrefab(meshWalls, texBrick),
		*scene.NewPrefab(meshCeiling, texRock),
		*scene.NewPrefab(meshTable, texWood),
	)

	return nil
}

func (e *Engine) setupRenderingPipeline() error {

	// init prefab renderer
	prefabRenderer, err := renderers.NewPrefabRenderer()
	if err != nil {
		return fmt.Errorf("failed to create prefab renderer")
	} else {
		e.prefabRenderer = prefabRenderer
	}

	// setup screen
	w, h := e.window.GetSize()
	e.screenConfig = &postprocessing.ScreenConfig{
		Width:           int32(w),
		Height:          int32(h),
		ResolutionRatio: 0.5,
	}

	// setup deferred render target
	deferred, err := renderers.NewDeferredRenderTarget(e.screenConfig)
	if err != nil {
		return err
	}
	e.deferred = deferred

	// init screen quad mesh
	meshQuad := render.NewMesh()
	meshQuad.SetupBasicQuad()

	screen, err := renderers.NewScreen(meshQuad, e.screenConfig)
	if err != nil {
		return err
	}
	e.screen = screen

	// setup effects
	// setup post processing effects

	// -- ssao
	ssaoConfig := &postprocessing.SSAOConfig{}
	ssaoPass, err := postprocessing.NewSSAOPass(
		e.screenConfig, meshQuad, ssaoConfig,
	)
	if err != nil {
		return err
	}

	// -- color grading
	colorGradingConfig := &postprocessing.ColorGradingConfig{
		Gamma:      1.714,
		Exposure:   1.194,
		Contrast:   1.28,
		Saturation: 0.96,
		Brightness: 1.180,
	}
	colorGradingPass, err := postprocessing.NewColorGradingPass(
		e.screenConfig, meshQuad, colorGradingConfig,
	)
	if err != nil {
		return err
	}

	// -- vignette
	vignetteConfig := &postprocessing.VignetteConfig{
		Radius: 0.85,
		Smooth: 0.535,
	}
	vignettePass, err := postprocessing.NewVignettePass(
		e.screenConfig, meshQuad, vignetteConfig,
	)
	if err != nil {
		return err
	}

	// create post processing pipeline
	e.ppPipeline = append(
		e.ppPipeline,
		ssaoPass, colorGradingPass, vignettePass,
	)

	// append resources
	e.resources = append(e.resources, meshQuad, screen, deferred, prefabRenderer)

	e.window.SetResizeCallback(e.framebufferSizeCallback)
	return nil
}

func (e *Engine) framebufferSizeCallback(w *glfw.Window, width, height int) {
	e.screenConfig.Width = int32(width)
	e.screenConfig.Height = int32(height)
	e.screenConfig.Aspect = float32(width) / float32(height)
	e.resizeRenderTargets()
}

func (e *Engine) resizeRenderTargets() {
	// resize all render targets
	// deferred
	e.deferred.ResizeCallback()
	for _, t := range e.ppPipeline {
		t.ResizeCallback()
	}
}

func (e *Engine) Run() {

	for !e.window.ShouldClose() {
		// process current context input
		e.processInput()
		e.processImgui()

		// update current context
		e.update()

		// resize targets if resolution changes
		if e.screenConfig.LastResolutionRatio != e.screenConfig.ResolutionRatio {
			e.resizeRenderTargets()
		}

		// render geometry
		e.deferred.BindForNewFrame()
		e.render()

		// perform post processing
		result := e.performPostProcessingPipeline()

		// render screen quad
		e.screen.RenderScreenQuad(result)

		// render imgui on top of screen
		e.renderImgui()

		// update global post rendering states
		stats.UpdateGlobal()
	}

}

func (e *Engine) processInput() {
	e.input.Update()
	glfw.PollEvents()

	io := imgui.CurrentIO()

	// update the controller only if ImGui does not want to capture the mouse.
	if e.input.GetGameMode() {
		// disable imgui input im game mode
		io.SetWantCaptureKeyboard(false)
		e.controller.Update(io.DeltaTime())
	} else {
		// update controller if has no imgui clicks
		if !io.WantCaptureMouse() && e.input.IsMouseButtonPressed(glfw.MouseButton1) {
			e.controller.Update(io.DeltaTime())
		}
	}
}

func (e *Engine) processImgui() {
	// begin imgui frame
	implgl3.NewFrame()
	implglfw.NewFrame()
	imgui.NewFrame()

	// apply custom font
	imgui.PushFont(e.imguiFont, 16.0)

	// imgui widgets
	e.debugMenu.Show()

	imgui.PopFont()

	// finalize imgui frame
	imgui.Render()
}

func (e *Engine) update() {
	shotgun := &e.prefabs[0]
	shotgun.Position = mgl32.Vec3{0, 1.446, 0}
	shotgun.Scaling = mgl32.Vec3{0.25, 0.25, 0.25}
	shotgun.Rotation[1] = -90
}

func (e *Engine) render() {
	// render scene
	e.prefabRenderer.Prepare(
		int(e.screenConfig.Width),
		int(e.screenConfig.Height),
		e.controller.GetCamera(),
	)
	for _, p := range e.prefabs {
		e.prefabRenderer.Render(&p)
	}
}

func (e *Engine) renderImgui() {
	drawData := imgui.CurrentDrawData()
	stats.UpdateForImgui(drawData)
	implgl3.RenderDrawData(drawData)
	e.window.SwapBuffers()
}

func (e *Engine) performPostProcessingPipeline() *render.Texture {
	// get deferred textures
	color := e.deferred.DeferredFBO.ColorTextures[0]
	normal := e.deferred.DeferredFBO.ColorTextures[1]
	depth := e.deferred.DeferredFBO.DepthTexture

	lastColor := color

	// perform
	for _, node := range e.ppPipeline {
		node.RenderPass([]*render.Texture{lastColor, normal, depth})
		lastColor = node.GetColor()
	}

	return lastColor
}

func (e *Engine) Destroy() {

	// clear post process pipeline
	for _, n := range e.ppPipeline {
		if n != nil {
			n.Delete()
		}
	}

	// clear resources
	for _, r := range e.resources {
		if r != nil {
			r.Delete()
		}
	}

	implgl3.Shutdown()
	implglfw.Shutdown()
	imgui.DestroyContext()
	e.window.Destroy()
	glfw.Terminate()
}
