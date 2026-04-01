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
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-kit-go/src/engine/assetmgr"
	"github.com/illoprin/retro-fps-kit-go/src/engine/stats"
	imguimenus "github.com/illoprin/retro-fps-kit-go/src/imgui"
	"github.com/illoprin/retro-fps-kit-go/src/model"
	"github.com/illoprin/retro-fps-kit-go/src/player"
	postprocessing "github.com/illoprin/retro-fps-kit-go/src/post_processing"
	"github.com/illoprin/retro-fps-kit-go/src/render"
	"github.com/illoprin/retro-fps-kit-go/src/render/context"
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

	// application state (level, game menu, editor)
	currentState EngineState

	// geometry render pass
	resources      []render.Resource
	deferred       *renderers.DeferredRenderTarget
	prefabRenderer *renderers.PrefabRenderer
	prefabs        []scene.Prefab

	// post processing render pass
	passPipeline []postprocessing.PostProcessingPass
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

	if err := context.InitContext(); err != nil {
		return nil, err
	}

	context.LogUserHardware()
	context.SetupDebugOutput()

	if err := e.initImgui(); err != nil {
		return nil, fmt.Errorf("failed to init imgui context - %v", err)
	}

	if err := e.initImguiRenderer(); err != nil {
		return nil, fmt.Errorf("failed to init imgui renderer - %v", err)
	}

	if err := e.initGame(); err != nil {
		return nil, fmt.Errorf("failed to init game - %v", err)
	}

	if err := e.initRenderingPipeline(); err != nil {
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

	deferredResult := e.deferred.GetResult()

	deferredTextures := []imguimenus.ImageTexture{
		{
			ID:   deferredResult.Color.ID,
			Name: "Color",
		},
		{
			ID:   deferredResult.Normal.ID,
			Name: "Normal",
		},
		{
			ID:   deferredResult.Position.ID,
			Name: "Position",
		},
		{
			ID:   deferredResult.Depth.ID,
			Name: "Depth",
		},
	}

	// Post-processing textures (add more as needed)
	passTextures := make([]imguimenus.ImageTexture, 0)
	for _, p := range e.passPipeline {

		if p.GetName() == "ssao" {
			ssao := p.(*postprocessing.SSAOPass)
			rawSSAO := imguimenus.ImageTexture{
				ID:   ssao.GetOcclusion().ID,
				Name: "ssao.raw",
			}
			noiseSSAO := imguimenus.ImageTexture{
				ID:   ssao.GetNoise().ID,
				Name: "ssao.noise",
			}
			blurSSAO := imguimenus.ImageTexture{
				ID:   ssao.GetBlur().ID,
				Name: "ssao.blur",
			}
			passTextures = append(passTextures, rawSSAO, noiseSSAO, blurSSAO)
		}

		if p.GetName() == "crease" {
			crease := p.(*postprocessing.CreaseOcclusionPass)
			rawCrease := imguimenus.ImageTexture{
				ID:   crease.GetOcclusion().ID,
				Name: "crease.raw",
			}
			creaseBlur := imguimenus.ImageTexture{
				ID:   crease.GetBlur().ID,
				Name: "crease.blur",
			}
			passTextures = append(passTextures, rawCrease, creaseBlur)
		}

		t := imguimenus.ImageTexture{
			ID:   p.GetColor().ID,
			Name: p.GetName(),
		}
		passTextures = append(passTextures, t)
	}

	ssao := e.passPipeline[0].(*postprocessing.SSAOPass)
	crease := e.passPipeline[1].(*postprocessing.CreaseOcclusionPass)
	colorGrading := e.passPipeline[2].(*postprocessing.ColorGradingPass)
	vignette := e.passPipeline[3].(*postprocessing.VignettePass)

	e.debugMenu = imguimenus.NewDebugMenu(
		e.window.GetConfig(), // window config
		e.deferred,           // deferred render target (for wireframe)
		e.controller,         // player controller
		ssao.GetConfig().(*postprocessing.SSAOConfig),
		crease.GetConfig().(*postprocessing.CreaseOcclusionConfig),
		colorGrading.GetConfig().(*postprocessing.ColorGradingConfig),
		vignette.GetConfig().(*postprocessing.VignetteConfig),
		deferredTextures, // deferred textures
		passTextures,     // post-processing textures
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

func (e *Engine) initGame() error {

	e.controller = player.NewEditorController(
		e.input, mgl32.Vec3{0, 0, 3}, 10.5, 0.25,
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
	meshShotgun.SetupFromModel(shotgunModel, render.StaticDraw)

	// floor mesh
	meshFloor := render.NewMesh()
	meshFloor.SetupFromModel(floorModel, render.StaticDraw)

	// ceiling mesh
	meshCeiling := render.NewMesh()
	meshCeiling.SetupFromModel(ceilingModel, render.StaticDraw)

	// walls mesh
	meshWalls := render.NewMesh()
	meshWalls.SetupFromModel(wallsModel, render.StaticDraw)

	// walls mesh
	meshTable := render.NewMesh()
	meshTable.SetupFromModel(tableModel, render.StaticDraw)

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

func (e *Engine) initRenderingPipeline() error {

	// init prefab renderer
	prefabRenderer, err := renderers.NewPrefabRenderer()
	if err != nil {
		return fmt.Errorf("failed to create prefab renderer - %w", err)
	} else {
		e.prefabRenderer = prefabRenderer
	}

	// setup deferred render target
	deferred, err := renderers.NewDeferredRenderTarget(e.window.GetConfig())
	if err != nil {
		return err
	}
	e.deferred = deferred

	// init screen quad mesh
	meshQuad := render.NewMesh()
	meshQuad.SetupBasicQuad()

	// auxiliary resources texture
	noiseTexture, err := postprocessing.CreateNoiseTexture()
	if err != nil {
		return err
	}
	// blur - uses ssao and cavity
	overlayBlurProgram, err := render.NewProgram(
		assetmgr.GetShaderPath("quad.vert"),
		assetmgr.GetShaderPath("mask_blur.frag"),
	)
	if err != nil {
		return err
	}
	// compositor - uses ssao and cavity
	overlayCompositorProgram, err := render.NewProgram(
		assetmgr.GetShaderPath("quad.vert"),
		assetmgr.GetShaderPath("mask_compositor.frag"),
	)
	if err != nil {
		return err
	}

	// setup effects
	// setup post processing effects

	// -- ssao
	ssaoConfig := &postprocessing.SSAOConfig{
		Use:        true,
		KernelSize: 30,
		Radius:     0.5,
		Bias:       0.005,
		WhitePoint: 0.971,
		BlackPoint: 0.39,
		BlurSize:   2,
	}
	ssaoPass, err := postprocessing.NewSSAOPass(
		e.window.GetConfig(), meshQuad, ssaoConfig,
		noiseTexture, overlayBlurProgram, overlayCompositorProgram,
	)
	if err != nil {
		return err
	}

	// -- crease occlusion
	creaseConfig := &postprocessing.CreaseOcclusionConfig{
		Use:        false,
		Radius:     17,
		DepthBias:  0.001,
		Intensity:  3.5,
		KernelSize: 32,
		WhitePoint: 1.0,
		BlackPoint: 0.26,
		BlurSize:   3,
	}
	creasePass, err := postprocessing.NewCreaseOcclusionPass(
		e.window.GetConfig(), meshQuad, creaseConfig,
		noiseTexture, overlayBlurProgram, overlayCompositorProgram,
	)
	if err != nil {
		return err
	}

	// -- color grading
	colorGradingConfig := &postprocessing.ColorGradingConfig{
		Gamma:          1.9,
		Exposure:       1.6,
		Contrast:       1.18,
		Saturation:     0.85,
		Brightness:     1.4,
		ShadowsColor:   [3]float32{.063, .102, .576},
		MidColor:       [3]float32{.494, .294, .067},
		HighlightColor: [3]float32{.903, .402, .061},
		ColorStrength:  0.68,
		Use:            true,
	}
	colorGradingPass, err := postprocessing.NewColorGradingPass(
		e.window.GetConfig(), meshQuad, colorGradingConfig,
	)
	if err != nil {
		return err
	}

	// -- vignette
	vignetteConfig := &postprocessing.VignetteConfig{
		Radius:   0.85,
		Softness: 0.535,
		Use:      true,
	}
	vignettePass, err := postprocessing.NewVignettePass(
		e.window.GetConfig(), meshQuad, vignetteConfig,
	)
	if err != nil {
		return err
	}

	// create post processing pipeline
	e.passPipeline = append(
		e.passPipeline,
		ssaoPass, creasePass, colorGradingPass, vignettePass,
	)

	// append resources
	e.resources = append(e.resources,
		meshQuad,
		deferred,
		prefabRenderer,
		noiseTexture,
		overlayBlurProgram,
		overlayCompositorProgram,
	)

	e.window.SetResizeCallback(e.framebufferSizeCallback)
	return nil
}

func (e *Engine) framebufferSizeCallback(w *glfw.Window, width, height int) {
	e.resizeRenderTargets()
}

func (e *Engine) resizeRenderTargets() {
	// resize all render targets
	// deferred
	e.deferred.ResizeCallback()
	for _, t := range e.passPipeline {
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
		if e.window.GetConfig().LastResolutionRatio != e.window.GetConfig().ResolutionRatio {
			e.resizeRenderTargets()
		}

		// render geometry
		e.deferred.BindForNewFrame()
		e.render()
		context.BindFramebuffer(nil)

		// setup context for 2D rendering
		context.SetupForPolygonFill()
		context.DisableDepth()

		// perform post processing
		_, resultFbo := e.performPostProcessingPipeline()

		// blit to main framebuffer
		resultFbo.Blit(0, e.window.GetConfig().Width, e.window.GetConfig().Height)

		// render imgui on top of screen
		e.renderImgui()

		e.window.SwapBuffers()

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
		int(e.window.GetConfig().Width),
		int(e.window.GetConfig().Height),
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
}

// performPostProcessingPipeline applies post processing and returns result texture and framebuffer
func (e *Engine) performPostProcessingPipeline() (*render.Texture, *render.Framebuffer) {
	// get deferred textures
	deferredRes := e.deferred.GetResult()

	var resultFbo *render.Framebuffer = e.deferred.GetFramebuffer()

	// get actual camera
	cam := e.controller.GetCamera()

	// perform
	for _, node := range e.passPipeline {
		if !node.Use() {
			continue
		}
		if node.GetName() == "ssao" {
			ssaoPass := node.(*postprocessing.SSAOPass)
			ssaoPass.SetProjectionMatrix(cam.Projection)
		} else if node.GetName() == "crease" {
			creasePass := node.(*postprocessing.CreaseOcclusionPass)
			creasePass.SetProjectionMatrix(cam.Projection)
		}
		node.RenderPass(deferredRes)
		deferredRes.Color = node.GetColor()
		resultFbo = node.GetResultFramebuffer()
	}

	return deferredRes.Color, resultFbo
}

func (e *Engine) Destroy() {

	// clear post process pipeline
	for _, n := range e.passPipeline {
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
