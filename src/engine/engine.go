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
	window         *window.Window
	debugMenu      *imguimenus.DebugMenu
	input          *player.InputManager
	controller     *player.EditorController
	imguiFont      *imgui.Font
	prefabRenderer *renderers.PrefabRenderer
	resources      []render.Resource
	prefabs        []scene.Prefab
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

	if err := e.initImgui(); err != nil {
		return nil, fmt.Errorf("failed to init imgui context - %v", err)
	}

	// setup input manager
	if err := e.initImguiRenderer(); err != nil {
		return nil, fmt.Errorf("failed to init imgui renderer - %v", err)
	}

	if err := e.initScene(); err != nil {
		return nil, fmt.Errorf("failed to init scene - %v", err)
	}

	e.setupGL()

	e.initMenus()

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
	e.input = player.NewManager(e.window.Window)

	return nil
}

func (e *Engine) setupGL() {
	width, height := e.window.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	gl.ClearColor(0.176, 0.216, 0.302, 1.0)
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)
	gl.FrontFace(gl.CCW)
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

func (e *Engine) initMenus() {
	e.debugMenu = imguimenus.NewDebugMenu(e.controller)
}

func (e *Engine) initScene() error {
	e.controller = player.NewEditorController(
		e.input, mgl32.Vec3{0, 0, 3}, 10.5, 0.085,
	)

	// init prefab renderer
	prefabRenderer, err := renderers.NewPrefabRenderer()
	if err != nil {
		return fmt.Errorf("failed to create prefab renderer")
	} else {
		e.prefabRenderer = prefabRenderer
	}

	// init assets
	// model
	parser := model.NewOBJParser()
	shotgunModel, err := parser.ParseFile(assetmgr.GetModelPath("shotgun.obj"))
	if err != nil {
		log.Printf("failed to load model %v", err)
	}
	// mesh
	meshShotgun := render.NewMesh()
	meshShotgun.SetupFromModel(shotgunModel, gl.STATIC_DRAW)
	e.resources = append(e.resources, meshShotgun)
	// texture
	texColors, err := render.NewTextureFromImage(assetmgr.GetTexturePath("colors.png"), true)
	if err != nil {
		log.Printf("failed to load texture %v", err)
	} else {
		e.resources = append(e.resources, texColors)
	}
	e.prefabs = append(e.prefabs, *scene.NewPrefab(meshShotgun, texColors))
	return nil
}

func (e *Engine) Run() {
	for !e.window.ShouldClose() {
		e.processInput()
		e.processImgui()
		e.render()
		stats.UpdateGlobal()
	}
}

func (e *Engine) processInput() {
	e.input.Update()
	glfw.PollEvents()

	io := imgui.CurrentIO()

	// process input and update scene
	if e.input.IsKeyJustPressed(glfw.KeyEscape) {
		e.window.SetShouldClose(true)
	}
	if e.input.IsKeyJustPressed(glfw.KeyF8) {
		e.input.ToggleGameMode()
	}
	if e.input.IsKeyJustPressed(glfw.KeyF1) {
		e.debugMenu.Visible = !e.debugMenu.Visible
	}

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
	imgui.PushFont(e.imguiFont, 12.0)

	// imgui widgets
	e.debugMenu.Show()

	imgui.PopFont()

	// finalize imgui frame
	imgui.Render()
}

func (e *Engine) Destroy() {

	// clear resources
	for _, res := range e.resources {
		res.Delete()
	}

	e.prefabRenderer.Shutdown()
	implgl3.Shutdown()
	implglfw.Shutdown()
	imgui.DestroyContext()
	e.window.Destroy()
	glfw.Terminate()
}

func (e *Engine) update() {

}

func (e *Engine) render() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	// render scene
	w, h := e.window.GetSize()
	e.prefabRenderer.Prepare(w, h, e.controller.GetCamera())
	for _, p := range e.prefabs {
		e.prefabRenderer.Render(&p)
	}

	// render imgui on top of scene
	e.renderImgui()

	e.window.SwapBuffers()
}

func (e *Engine) renderImgui() {
	drawData := imgui.CurrentDrawData()
	stats.UpdateForImgui(drawData)
	implgl3.RenderDrawData(drawData)
}
