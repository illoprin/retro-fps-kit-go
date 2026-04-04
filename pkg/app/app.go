package app

import (
	"fmt"
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

func (e *App) setupCallbacks() {
	e.input.SetKeyCallback(
		func(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {
			if action == glfw.Press {
				if key == glfw.KeyEscape {
					e.window.SetShouldClose(true)
				}
			}

			e.iUI.OnKey(key, action)

			if e.activeState != nil {
				e.activeState.OnKey(key, action, mods)
			}
		},
	)

	e.input.SetMouseButtonCallback(
		func(button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
			if e.activeState != nil {
				e.activeState.OnMouseButton(button, action, mods)
			}
		},
	)
	e.input.SetMouseScrollCallback(
		func(xOffset, yOffset float64) {
			if e.activeState != nil {
				e.activeState.OnMouseScroll(xOffset, yOffset)
			}
		},
	)
	e.input.SetMouseMoveCallback(
		func(x, y, dx, dy float64) {
			if e.activeState != nil {
				e.activeState.OnMouseMove(x, y, dx, dy)
			}
		},
	)
	e.window.SetResizeCallback(
		func(win *glfw.Window, width, height int) {
			e.resizeRenderTargets()
		},
	)
}

// ---- Renderer

func (e *App) initRenderingPipeline() error {

	// setup deferred render target
	deferred, err := pipeline.NewDeferredRenderTarget(e.window.GetConfig())
	if err != nil {
		return err
	}
	e.deferred = deferred

	// init screen quad mesh
	meshQuad := rhi.NewMesh()
	meshQuad.SetupBasicQuad()

	// auxiliary resources texture
	noiseTexture, err := passes.CreateNoiseTexture()
	if err != nil {
		return err
	}
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
		e.passPipeline = append(e.passPipeline, p)
		return nil
	}

	// -- ssao

	if err := addPass(passes.NewSSAOPass(
		e.window.GetConfig(), meshQuad, config.SSAOConfig,
		noiseTexture, blurProg, compProg)); err != nil {
		return fmt.Errorf("ssao pass - %w", err)
	}

	// -- crease occlusion

	if err := addPass(passes.NewCavityPass(
		e.window.GetConfig(), meshQuad, config.CavityConfig,
		noiseTexture, blurProg, compProg,
	)); err != nil {
		return fmt.Errorf("crease pass - %w", err)
	}

	// -- color grading

	if err := addPass(passes.NewColorGradingPass(
		e.window.GetConfig(), meshQuad, config.ColorGradingConfig,
	)); err != nil {
		return fmt.Errorf("color grading pass - %w", err)
	}

	// -- vignette

	if err := addPass(passes.NewVignettePass(
		e.window.GetConfig(), meshQuad, config.VignetteConfig,
	)); err != nil {
		return fmt.Errorf("vignette pass - %w", err)
	}

	// append resources
	e.resources = append(e.resources,
		meshQuad,
		deferred,
		noiseTexture,
		blurProg,
		compProg,
	)

	return nil
}

func (e *App) resizeRenderTargets() {
	// deferred
	e.deferred.ResizeCallback()
	// resize all render targets
	for _, t := range e.passPipeline {
		t.ResizeCallback()
	}
	// resize current state
	if e.activeState != nil {
		sw, sh := e.window.GetConfig().GetScreenSize()
		w, h := e.window.GetConfig().Width, e.window.GetConfig().Height
		e.activeState.OnResize(w, h, sw, sh)
	}
}

// ---- Game cycle

func (e *App) Run() {

	for !e.window.ShouldClose() {

		// -- Input

		io := imgui.CurrentIO()
		if e.window.GetCursorDisabled() {
			io.SetWantCaptureMouse(false)
			io.SetWantCaptureKeyboard(false)
		}
		e.input.Update()
		glfw.PollEvents()

		// -- Update

		// - imgui
		// start new frame
		ui.NewFrame()
		// draw custom ui
		e.iUI.Draw()
		// finalize
		ui.FinalizeFrame()
		// ------

		// update current context
		if e.activeState != nil {
			e.activeState.Update(float32(e.monitor.GetFrameTime()))
		}

		// resize targets if screen ratio changes
		if e.window.GetConfig().LastResolutionRatio != e.window.GetConfig().ResolutionRatio {
			e.resizeRenderTargets()
		}

		// -- Render

		context.BindFramebuffer(nil)

		// perform custom indirect rendering of current state (if needs)
		if state, ok := e.activeState.(IndirectDrawer); ok {
			context.SaveState() // WARN not notimplemented
			state.RenderIndirect()
			context.RestoreState()
		}

		// perform gbuffer rendering of current state (if needs)
		var lastRenderTarget *rhi.Framebuffer
		if state, ok := e.activeState.(GBufferDrawer); ok {
			// setup for geometry rendering
			context.SetupForGeometry()
			// render geometry
			e.deferred.BindForNewFrame()
			// render geometry of current app state
			state.RenderGBuffer()
			// setup context for 2D rendering
			context.SetupForFlat()
			// perform post processing
			_, lastRenderTarget = e.performPostProcessingPipeline()
		}

		context.BindFramebuffer(lastRenderTarget)
		// perform flat rendering of current state (if needs)
		if state, ok := e.activeState.(FlatDrawer); ok {
			// render flat of current app state
			// on top of last render target
			state.RenderFlat(lastRenderTarget)
		}

		// blit last render target to initial framebuffer
		lastRenderTarget.Blit(0, e.window.GetConfig().Width, e.window.GetConfig().Height, rhi.FilterNearest)

		// render imgui on top of screen
		ui.Render()

		// -- End Frame

		// update monitor stats
		e.monitor.Update()

		// reset rhi stats
		rhi.FrameStats.Reset()

		e.window.SwapBuffers()
	}
}

// performPostProcessingPipeline applies post processing
// and returns result texture and framebuffer
func (e *App) performPostProcessingPipeline() (*rhi.Texture, *rhi.Framebuffer) {
	s, ok := e.activeState.(GBufferDrawer)
	if !ok {
		return nil, nil
	}

	// get deferred textures
	res := e.deferred.GetResult()
	fbo := e.deferred.GetFramebuffer()

	// get actual camera
	cam := s.GetCamera()

	// perform
	for _, node := range e.passPipeline {
		if !node.Use() {
			continue
		}
		if p, ok := node.(passes.HasProjection); ok {
			p.SetProjectionMatrix(cam.Projection)
		}
		node.RenderPass(res)
		res.Color = node.GetColor()
		fbo = node.GetResultFramebuffer()
	}

	return res.Color, fbo
}

// ---- AppProvider implementations

func (e *App) Close() {
	e.window.SetShouldClose(true)
}

func (e *App) SetActiveState(s AppState) {
	e.activeState = s
}

func (e *App) GetConfig() *config.Config {
	return nil
}

func (e *App) GetWindow() *window.Window {
	return e.window
}

func (e *App) GetInputManager() *window.InputManager {
	return e.input
}

func (e *App) GetMonitor() *monitor.Monitor {
	return e.monitor
}

func (e *App) GetGBuffer() *rhi.Framebuffer {
	return e.deferred.GetFramebuffer()
}

func (e *App) GUIWantCaptureMouse() bool {
	io := imgui.CurrentIO()
	return io.WantCaptureMouse()
}

func (e *App) GetTime() float64 {
	return glfw.GetTime()
}

// ---- Auxiliary methods

func (e *App) initSystems() error {
	if err := window.InitGLFW(); err != nil {
		return fmt.Errorf("failed to init glfw - %v", err)
	}

	if err := e.createWindow(); err != nil {
		return fmt.Errorf("failed to init window - %v", err)
	}

	if err := context.InitContext(); err != nil {
		return err
	}

	context.LogUserHardware()
	context.SetupDebugOutput()

	if err := e.initImguiContext(); err != nil {
		return fmt.Errorf("failed to init imgui context - %v", err)
	}

	return nil
}

func (e *App) createWindow() error {
	// create window
	var err error
	e.window, err = window.NewWindow(
		config.WindowWidth,
		config.WindowHeight,
		config.WindowTitle,
		config.DefaultResolutionRatio,
	)
	if err != nil {
		return fmt.Errorf("failed init window %v", err)
	}
	e.window.MakeContextCurrent()
	e.window.Focus()
	e.window.Center()

	e.input = window.NewManager(e.window.Window)

	return nil
}

func (e *App) initImguiContext() error {
	// init imgui context
	ui.Init()

	// init imgui renderer
	if err := ui.InitImgui(e.window); err != nil {
		return fmt.Errorf("failed to init imgui renderer - %v", err)
	}
	return nil
}

func (e *App) initUI() {

	// prepare editors for ConfigUI
	// (post process options editor)
	editors := make([]ui.ConfigUI, len(e.passPipeline))
	for i, n := range e.passPipeline {
		var editor ui.ConfigUI

		switch p := n.(type) {
		case *passes.SSAOPass:
			editor = &ui.SSAOConfigUI{SSAOConfig: p.GetConfig().(*passes.SSAOConfig)}
		case *passes.CavityPass:
			editor = &ui.CavityConfigUI{CavityConfig: p.GetConfig().(*passes.CavityConfig)}
		case *passes.ColorGradingPass:
			editor = &ui.ColorGradingUI{ColorGradingConfig: p.GetConfig().(*passes.ColorGradingConfig)}
		case *passes.VignettePass:
			editor = &ui.VignetteUI{VignetteConfig: p.GetConfig().(*passes.VignetteConfig)}
		}

		editors[i] = editor
	}

	e.iUI.AttachDebugUI(
		ui.NewDebugUI(
			e.window.GetConfig(),
			e.monitor,
			e.deferred,
			editors,
		),
	)

	e.iUI.AttachFramebuffersUI(
		ui.NewFramebuffersUI(
			e.deferred.GetResult(),
			e.passPipeline,
			e.window.GetConfig(),
		),
	)
}

// -- Destroy

func (e *App) Destroy() {

	if e.activeState != nil {
		e.activeState.Destroy()
	}

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

	ui.Destroy()
	e.window.Destroy()
	glfw.Terminate()
}
