package app

import (
	"fmt"
	"log"
	"runtime"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/illoprin/retro-fps-kit-go/pkg/app/config"
	"github.com/illoprin/retro-fps-kit-go/pkg/app/ui"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/files"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/monitor"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/window"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/context"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/passes"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/pipeline"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
)

var (
	cursorPointer *glfw.Cursor
	cursorWaiting *glfw.Cursor
	cursorCross   *glfw.Cursor
)

type CursorType int

const (
	CursorPointer CursorType = iota
	CursorWaiting
	CursorCross
)

type App struct {
	window  *window.Window
	input   *window.InputManager
	monitor *monitor.Monitor

	// initial imgui debug screen
	iUI *ui.InitialUI

	// application state (level, game menu, editor)
	activeState AppState

	// initial rhi resources
	resources []rhi.Resource

	// geometry render pass
	deferred *pipeline.DeferredRenderTarget

	// post processing render pass
	passPipeline []passes.PostProcessingPass
}

func NewApp() (*App, error) {
	runtime.LockOSThread()

	e := &App{
		resources: make([]rhi.Resource, 0),
		monitor:   monitor.NewMonitor(),
	}

	if err := e.initSystems(); err != nil {
		return nil, err
	}

	if err := e.initRenderingPipeline(); err != nil {
		return nil, fmt.Errorf("failed to init rendering pipeline - %v", err)
	}

	// setup custom gui
	e.iUI, _ = ui.NewInitialUI()
	e.initUI()

	// setup callbacks for input manager
	e.setupCallbacks()

	return e, nil
}

// ---- Callbacks

func (a *App) setupCallbacks() {
	a.input.SetKeyCallback(
		func(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {
			if action == glfw.Press {
				if key == glfw.KeyEscape {
					a.window.SetShouldClose(true)
				}
			}

			a.iUI.OnKey(key, action)

			if a.activeState != nil {
				a.activeState.OnKey(key, action, mods)
			}
		},
	)

	a.input.SetMouseButtonCallback(
		func(button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
			if a.activeState != nil {
				a.activeState.OnMouseButton(button, action, mods)
			}
		},
	)
	a.input.SetMouseScrollCallback(
		func(xOffset, yOffset float64) {
			if a.activeState != nil {
				a.activeState.OnMouseScroll(xOffset, yOffset)
			}
		},
	)
	a.input.SetMouseMoveCallback(
		func(x, y, dx, dy float64) {
			if a.activeState != nil {
				a.activeState.OnMouseMove(x, y, dx, dy)
			}
		},
	)
	a.window.SetResizeCallback(
		func(win *glfw.Window, width, height int) {
			a.resizeRenderTargets()
		},
	)
}

// ---- Renderer

func (a *App) initRenderingPipeline() error {

	// setup deferred render target
	deferred, err := pipeline.NewDeferredRenderTarget(a.window.GetConfig())
	if err != nil {
		return err
	}
	a.deferred = deferred

	// init screen quad mesh
	meshQuad := rhi.NewMesh()
	rhi.SetupBasicQuadMesh(meshQuad)

	// auxiliary resources
	noiseTexture := passes.CreateNoiseTexture()

	// blur - uses ssao and cavity
	blurProg, err := rhi.NewProgram(
		files.GetShaderPath("screen.vert"),
		files.GetShaderPath("overlay_blur.frag"),
	)
	if err != nil {
		return err
	}
	// compositor - uses ssao and cavity
	compProg, err := rhi.NewProgram(
		files.GetShaderPath("screen.vert"),
		files.GetShaderPath("overlay_compositor.frag"),
	)
	if err != nil {
		return err
	}

	// setup post processing effects

	// helper function to create pass
	addPass := func(p passes.PostProcessingPass, err error) error {
		if err != nil {
			return err
		}
		a.passPipeline = append(a.passPipeline, p)
		return nil
	}

	// -- eye adaption pass

	if err := addPass(passes.NewEyeAdaptionPass(
		a.window.GetConfig(), meshQuad, config.EyeAdaptionConfig),
	); err != nil {
		return fmt.Errorf("eye adaption pass - %w", err)
	}

	// -- ssao

	if err := addPass(passes.NewSSAOPass(
		a.window.GetConfig(), meshQuad, config.SSAOConfig,
		noiseTexture, blurProg, compProg)); err != nil {
		return fmt.Errorf("ssao pass - %w", err)
	}

	// -- cavity occlusion

	if err := addPass(passes.NewCavityPass(
		a.window.GetConfig(), meshQuad, config.CavityConfig,
		blurProg, compProg,
	)); err != nil {
		return fmt.Errorf("cavity pass - %w", err)
	}

	// -- bloom

	if err := addPass(passes.NewBloomPass(
		a.window.GetConfig(), meshQuad, config.BloomConfig,
	)); err != nil {
		return fmt.Errorf("bloom pass - %w", err)
	}

	// -- tone mapping

	if err := addPass(passes.NewToneMappingPass(
		a.window.GetConfig(), meshQuad, config.ToneMappingConfig,
	)); err != nil {
		return fmt.Errorf("tone mapping pass - %w", err)
	}

	// -- color grading

	if err := addPass(passes.NewColorGradingPass(
		a.window.GetConfig(), meshQuad, config.ColorGradingConfig,
	)); err != nil {
		return fmt.Errorf("color grading pass - %w", err)
	}

	// -- vignette

	if err := addPass(passes.NewVignettePass(
		a.window.GetConfig(), meshQuad, config.VignetteConfig,
	)); err != nil {
		return fmt.Errorf("vignette pass - %w", err)
	}

	// append resources
	a.resources = append(a.resources,
		meshQuad,
		deferred,
		noiseTexture,
		blurProg,
		compProg,
	)

	return nil
}

func (a *App) resizeRenderTargets() {
	// deferred
	a.deferred.ResizeCallback()
	// resize all render targets
	for _, t := range a.passPipeline {
		t.ResizeCallback()
	}
	// resize current state
	if a.activeState != nil {
		sw, sh := a.window.GetConfig().GetScreenSize()
		w, h := a.window.GetConfig().Width, a.window.GetConfig().Height
		a.activeState.OnResize(w, h, sw, sh)
	}
}

// ---- Game cycle

func (a *App) Run() {

	a.SetCursor(CursorPointer)

	for !a.window.ShouldClose() {

		a.monitor.NewFrame()

		// -- Input
		io := imgui.CurrentIO()
		if a.window.GetCursorDisabled() {
			io.SetWantCaptureMouse(false)
			io.SetWantCaptureKeyboard(false)
		}
		a.input.Update()
		glfw.PollEvents()

		// -- Update

		// - imgui
		// start new frame
		ui.NewFrame()
		// draw custom ui
		a.iUI.Draw()
		// render state ui (if needs)
		if s, ok := a.activeState.(UIDrawer); ok {
			s.DrawImgui()
		}
		// finalize
		ui.FinalizeFrame()
		// ------

		// update current context
		if a.activeState != nil {
			a.activeState.Update(a.monitor.GetDeltaTime())
		}

		// resize targets if screen ratio changes
		if a.window.GetConfig().LastResolutionRatio != a.window.GetConfig().ResolutionRatio {
			a.resizeRenderTargets()
		}

		// -- Render

		context.BindFramebuffer(nil) // bind initial (0) framebuffer
		context.ClearColorBuffer()

		// perform custom indirect rendering of current state (if needs)
		if state, ok := a.activeState.(IndirectDrawer); ok {
			// save context state
			context.CaptureState()
			// perform user's rendering
			state.RenderIndirect()
			// restore context state
			context.RestoreState()
		}

		// perform gbuffer rendering of current state (if needs)
		var lastRenderTarget *rhi.Framebuffer
		if state, ok := a.activeState.(GBufferDrawer); ok {
			// setup for geometry rendering
			context.SetupForGeometry()
			// render geometry
			a.deferred.BindForNewFrame()
			// render geometry of current app state
			state.RenderGBuffer()
			// setup context for 2D rendering
			context.SetupForFlat()
			// perform post processing
			_, lastRenderTarget = a.performPostProcessingPipeline()
			// send camera info to debug ui
			a.iUI.GetDebugUI().SetActiveCamera(state.GetCamera())
		}

		context.BindFramebuffer(lastRenderTarget)
		// perform flat rendering of current state (if needs)
		if state, ok := a.activeState.(FlatDrawer); ok {
			// render flat of current app state
			// on top of last render target
			state.RenderFlat(lastRenderTarget)
		}

		if lastRenderTarget != nil {
			// blit last render target to initial framebuffer
			lastRenderTarget.Blit(
				0, a.window.GetConfig().Width, a.window.GetConfig().Height,
				rhi.FilterNearest,
			)
		}

		// render imgui on top of screen
		ui.Render()

		// -- End Frame

		// update monitor stats
		a.monitor.Update()

		// reset rhi stats
		rhi.FrameStats.Reset()

		a.window.SwapBuffers()
	}
}

// performPostProcessingPipeline applies post processing
// and returns result texture and framebuffer
func (a *App) performPostProcessingPipeline() (*rhi.Texture, *rhi.Framebuffer) {
	s, ok := a.activeState.(GBufferDrawer)
	if !ok {
		return nil, nil
	}

	// get deferred textures
	res := a.deferred.GetResult()
	fbo := a.deferred.GetFramebuffer()

	// get actual camera
	cam := s.GetCamera()

	// perform
	for _, node := range a.passPipeline {
		if !node.Use() {
			continue
		}
		if p, ok := node.(passes.HasProjection); ok {
			p.SetProjectionMatrix(cam.Projection)
		}
		if p, ok := node.(passes.HasDeltaTime); ok {
			p.SetDeltaTime(a.monitor.GetDeltaTime())
		}
		node.RenderPass(res)
		res.Color = node.GetColor()
		fbo = node.GetResultFramebuffer()
	}

	return res.Color, fbo
}

// ---- AppProvider implementations

func (a *App) Close() {
	a.window.SetShouldClose(true)
}

func (a *App) SetActiveState(s AppState) {
	a.activeState = s
}

func (a *App) GetConfig() *config.Config {
	return nil
}

func (a *App) GetWindow() *window.Window {
	return a.window
}

func (a *App) GetInputManager() *window.InputManager {
	return a.input
}

func (a *App) GetMonitor() *monitor.Monitor {
	return a.monitor
}

func (a *App) GetGBuffer() *rhi.Framebuffer {
	return a.deferred.GetFramebuffer()
}

func (a *App) GUIWantCaptureMouse() bool {
	io := imgui.CurrentIO()
	return io.WantCaptureMouse()
}

func (a *App) GetTime() float64 {
	return glfw.GetTime()
}

func (a *App) SetCursor(t CursorType) {
	var currentCursor *glfw.Cursor

	switch t {
	case CursorPointer:
		currentCursor = cursorPointer
	case CursorWaiting:
		currentCursor = cursorWaiting
	case CursorCross:
		currentCursor = cursorCross
	}

	if currentCursor == nil {
		log.Printf("app.SetCursor - cursor not found\n")
		return
	}

	a.window.SetCursor(currentCursor)
}

// ---- Auxiliary methods

func (a *App) initSystems() error {
	if err := window.InitGLFW(); err != nil {
		return fmt.Errorf("failed to init glfw - %v", err)
	}

	if err := a.createWindow(); err != nil {
		return fmt.Errorf("failed to init window - %v", err)
	}

	if err := context.InitContext(); err != nil {
		return err
	}

	context.LogUserHardware()
	context.SetupDebugOutput()

	if err := a.initImguiContext(); err != nil {
		return fmt.Errorf("failed to init imgui context - %v", err)
	}

	return nil
}

func (a *App) createWindow() error {
	// create window
	var err error
	a.window, err = window.NewWindow(
		config.WindowWidth,
		config.WindowHeight,
		config.WindowTitle,
		config.DefaultResolutionRatio,
	)
	if err != nil {
		return fmt.Errorf("failed init window %v", err)
	}
	a.window.MakeContextCurrent()
	a.window.Focus()
	a.window.Center()

	// load icon
	err = a.window.SetIconFromFile(files.GetTexturePath("initial/logo.png"))
	if err != nil {
		log.Printf("app - failed to load window icon - %v\n", err)
	}

	// load cursors
	//
	// -- pointer
	cursorPointer, err = a.window.LoadCursor(files.GetTexturePath("initial/pointer.png"))
	if err != nil {
		log.Printf("app - failed to load cursor - %v\n", err)
	}
	// -- cross
	cursorCross, err = a.window.LoadCursor(files.GetTexturePath("initial/cross.png"))
	if err != nil {
		log.Printf("app - failed to load cursor - %v\n", err)
	}
	// -- waiting
	// later

	a.input = window.NewManager(a.window.Window)

	return nil
}

func (a *App) initImguiContext() error {
	// init imgui context
	ui.Init()

	// init imgui renderer
	if err := ui.InitImguiRenderer(a.window.Window); err != nil {
		return fmt.Errorf("failed to init imgui renderer - %v", err)
	}
	return nil
}

func (a *App) initUI() {

	// prepare editors for ConfigUI
	// (post process options editor)
	editors := make([]ui.ConfigUI, 0)
	for _, n := range a.passPipeline {
		var editor ui.ConfigUI

		switch p := n.(type) {
		case *passes.EyeAdaptionPass:
			editor = &ui.EyeAdaptionConfigUI{EyeAdaptionConfig: p.GetConfig().(*passes.EyeAdaptionConfig)}
		case *passes.SSAOPass:
			editor = &ui.SSAOConfigUI{SSAOConfig: p.GetConfig().(*passes.SSAOConfig)}
		case *passes.CavityPass:
			editor = &ui.CavityConfigUI{CavityConfig: p.GetConfig().(*passes.CavityConfig)}
		case *passes.BloomPass:
			editor = &ui.BloomConfigUI{BloomConfig: p.GetConfig().(*passes.BloomConfig)}
		case *passes.ToneMappingPass:
			editor = &ui.ToneMappingConfigUI{ToneMappingConfig: p.GetConfig().(*passes.ToneMappingConfig)}
		case *passes.ColorGradingPass:
			editor = &ui.ColorGradingUI{ColorGradingConfig: p.GetConfig().(*passes.ColorGradingConfig)}
		case *passes.VignettePass:
			editor = &ui.VignetteUI{VignetteConfig: p.GetConfig().(*passes.VignetteConfig)}
		default:
			continue
		}

		editors = append(editors, editor)
	}

	a.iUI.AttachDebugUI(
		ui.NewDebugUI(
			a.window.GetConfig(),
			a.monitor,
			a.deferred,
			editors,
		),
	)

	a.iUI.AttachFramebuffersUI(
		ui.NewFramebuffersUI(
			a.deferred.GetResult(),
			a.passPipeline,
			a.window.GetConfig(),
		),
	)
}

// -- Destroy

func (a *App) Destroy() {

	if a.activeState != nil {
		a.activeState.Destroy()
	}

	// clear post process pipeline
	for _, n := range a.passPipeline {
		if n != nil {
			n.Delete()
		}
	}

	// clear resources
	for _, r := range a.resources {
		if r != nil {
			r.Delete()
		}
	}

	ui.Destroy()
	a.window.Destroy()
	glfw.Terminate()
}
