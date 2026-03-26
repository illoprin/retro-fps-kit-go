package engine

import (
	"fmt"
	"log"
	"math"
	"math/rand"
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
	defaultWindowWidth     = 1470
	defaultWindowHeight    = 710
	defaultResolutionRatio = 0.5
	defaultTitle           = "Retro FPS Kit - Demo"
)

type PrefabState struct {
	LastEffectState bool
	Vertices        []model.ModelVertex
	Model           *model.Model
}

type Engine struct {
	window         *window.Window
	debugMenu      *imguimenus.DebugMenu
	input          *window.InputManager
	controller     *player.EditorController
	imguiFont      *imgui.Font
	prefabRenderer *renderers.PrefabRenderer
	resources      []render.Resource
	prefabs        []scene.Prefab
	prefabState    *PrefabState
	sceneFBO       *render.Framebuffer
	screenProgram  *render.Program
	screenQuad     *render.Mesh
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

	e.input = window.NewManager(e.window.Window)
	e.input.SetKeyCallback(func(key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
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
	})

	return nil
}

func (e *Engine) setupGL() error {
	width, height := e.window.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)
	gl.FrontFace(gl.CCW)

	// init scene framebuffer
	sceneFBO, err := render.NewFramebuffer(render.FramebufferConfig{
		Width:           int32(float32(width) * defaultResolutionRatio),
		Height:          int32(float32(height) * defaultResolutionRatio),
		ColorFormat:     render.FormatRGB8,
		ColorFiltering:  render.FilterNearest,
		UseDepth:        true,
		UseDepthStencil: false,
		UseMultisample:  false,
	})
	if err != nil {
		return err
	}
	e.sceneFBO = sceneFBO

	prevCallback := e.window.Window.SetFramebufferSizeCallback(nil)
	e.window.Window.SetFramebufferSizeCallback(func(w *glfw.Window, width, height int) {
		e.sceneFBO.Resize(
			int32(float32(width)*defaultResolutionRatio),
			int32(float32(height)*defaultResolutionRatio),
		)
		// gl.Viewport(0, 0,
		// 	int32(float32(width)*defaultResolutionRatio),
		// 	int32(float32(height)*defaultResolutionRatio),
		// )
		if prevCallback != nil {
			prevCallback(w, width, height)
		}
	})

	// init screen quad mesh
	e.screenQuad = render.NewMesh()
	e.screenQuad.SetupBasicQuad()

	// init screen quad shader
	screenProg, err := render.NewProgram(assetmgr.GetShaderPath("screen_quad.vert"), assetmgr.GetShaderPath("screen_quad.frag"))
	if err != nil {
		return err
	}
	e.screenProgram = screenProg

	e.resources = append(e.resources, sceneFBO, screenProg, e.screenQuad)

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

func (e *Engine) initMenus() {
	e.debugMenu = imguimenus.NewDebugMenu(e.controller)
}

func (e *Engine) Run() error {

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

	// model
	parser := model.NewOBJParser()
	shotgunModel, err := parser.ParseFile(assetmgr.GetModelPath("shotgun.obj"))
	if err != nil {
		log.Printf("failed to load model %v", err)
	}
	updatedModelVertices := make([]model.ModelVertex, len(shotgunModel.Vertices))

	// mesh
	meshShotgun := render.NewMesh()
	meshShotgun.SetupFromModel(shotgunModel, gl.DYNAMIC_DRAW)
	e.resources = append(e.resources, meshShotgun)
	// texture
	texColors, err := render.NewTextureFromImage(assetmgr.GetTexturePath("colors.png"), true)
	if err != nil {
		log.Printf("failed to load texture %v", err)
	} else {
		e.resources = append(e.resources, texColors)
	}
	// create prefab
	e.prefabs = append(e.prefabs, *scene.NewPrefab(meshShotgun, texColors))
	e.prefabState = &PrefabState{
		LastEffectState: false,
		Vertices:        updatedModelVertices,
		Model:           shotgunModel,
	}

	if err := e.setupGL(); err != nil {
		return err
	}
	e.initMenus()

	for !e.window.ShouldClose() {
		e.processInput()
		e.processImgui()
		// update scene
		e.update()
		// scene
		e.render()
		// render imgui on top of scene
		e.renderImgui()

		stats.UpdateGlobal()
	}

	return nil
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

	ps := e.prefabState
	shotgunMesh := e.prefabs[0].Mesh
	if e.debugMenu.DestroyingEffect {

		// update scene
		for i, v := range ps.Model.Vertices {
			normal := mgl32.Vec3{v.Nx, v.Ny, v.Nz}
			vert := mgl32.Vec3{v.X, v.Y, v.Z}

			factor := float32(math.Sin(glfw.GetTime()) * 0.1)
			factor = mgl32.Clamp(factor, 0, 1)
			offset := normal.Mul(factor + rand.Float32()*0.05)

			vert = vert.Add(offset)

			ps.Vertices[i] = v
			ps.Vertices[i].X = vert[0]
			ps.Vertices[i].Y = vert[1]
			ps.Vertices[i].Z = vert[2]
		}

		// update buffers
		shotgunMesh.UpdateVertexBuffer(0, ps.Vertices, shotgunMesh.GetCount())
		// update state
		e.prefabState.LastEffectState = true
		return
	}

	if e.prefabState.LastEffectState {
		shotgunMesh.UpdateVertexBuffer(0, ps.Model.Vertices, uint32(len(ps.Model.Indices)))
		e.prefabState.LastEffectState = false
	}

}

func (e *Engine) render() {

	e.sceneFBO.Bind()

	w, h := e.window.GetSize()
	gl.Viewport(0, 0, int32(float32(w)*defaultResolutionRatio), int32(float32(h)*defaultResolutionRatio))

	gl.ClearColor(0.176, 0.216, 0.302, 1.0)

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	// render scene
	e.prefabRenderer.Prepare(w, h, e.controller.GetCamera())
	for _, p := range e.prefabs {
		e.prefabRenderer.Render(&p)
	}

	e.sceneFBO.Unbind()

	gl.Disable(gl.DEPTH_TEST)

	gl.ClearColor(0, 0, 0, 1.0)

	gl.Viewport(0, 0, int32(w), int32(h))

	gl.Clear(gl.COLOR_BUFFER_BIT)

	e.screenProgram.Use()
	e.sceneFBO.ColorTexture.Bind(0)
	e.screenProgram.Set1i("u_color", 0)
	e.screenQuad.Draw()

}

func (e *Engine) renderImgui() {
	drawData := imgui.CurrentDrawData()
	stats.UpdateForImgui(drawData)
	implgl3.RenderDrawData(drawData)
	e.window.SwapBuffers()

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
