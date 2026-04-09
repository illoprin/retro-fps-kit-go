package app

import (
	"fmt"
	"log"
	"log/slog"
	"runtime"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/app/config"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/app/ui"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/core/files"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/core/logger"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/core/monitor"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/core/window"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/render/context"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/render/pipeline"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/render/post"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/render/rhi"
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
	cfg     *config.Config

	// initial imgui debug screen
	iUI *ui.InitialUI

	// application state (level, game menu, editor)
	activeState AppState

	// geometry render pass
	deferred *pipeline.DeferredRenderTarget

	// post processing render pass
	pipeline *DefaultPipeline
}

func NewApp(cfg *config.Config) (*App, error) {
	runtime.LockOSThread()

	e := &App{
		monitor: monitor.NewMonitor(),
		cfg:     cfg,
	}

	if err := e.initSystems(); err != nil {
		return nil, err
	}

	if err := e.initRenderingPipeline(); err != nil {
		logger.Errorf("failed to init rendering pipeline - %v", err)
		return nil, fmt.Errorf("failed to init rendering pipeline - %v", err)
	}

	// setup custom gui
	e.iUI, _ = ui.NewInitialUI()
	e.initUI()

	// setup callbacks for input manager
	e.setupCallbacks()

	logger.Infof("engine started")

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
			if s, ok := a.activeState.(MouseButtonsHandler); ok {
				s.OnMouseButton(button, action, mods)
			}
		},
	)
	a.input.SetMouseScrollCallback(
		func(xOffset, yOffset float64) {
			if s, ok := a.activeState.(MouseScrollHandler); ok {
				s.OnMouseScroll(xOffset, yOffset)
			}
		},
	)
	a.input.SetMouseMoveCallback(
		func(x, y, dx, dy float64) {
			if s, ok := a.activeState.(MouseMoveHandler); ok {
				s.OnMouseMove(x, y, dx, dy)
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
	var err error
	a.deferred, err = pipeline.NewDeferredRenderTarget(a.window.GetConfig())
	if err != nil {
		return err
	}

	// setup post processing pipeline
	a.pipeline, err = NewDefaultPipeline(
		a.window.GetConfig(),
		&a.cfg.PostProcessing,
	)
	if err != nil {
		logger.Errorf("failed to create post processing pipeline %v", err)
		return err
	}

	// build pipeline
	if err := a.pipeline.Build(); err != nil {
		logger.Errorf("post processing pipeline building error - %v", err)
		return err
	}

	return nil
}

func (a *App) resizeRenderTargets() {
	// deferred
	a.deferred.ResizeCallback()

	// resize all render targets
	a.pipeline.ResizeCallback()

	// call resize callback of active state
	if s, ok := a.activeState.(ResizeHandler); ok {
		sw, sh := a.window.GetConfig().GetScreenSize()
		w, h := a.window.GetConfig().Width, a.window.GetConfig().Height
		s.OnResize(w, h, sw, sh)
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
		// draw app ui
		a.iUI.Draw()
		// render state ui (if needs)
		if s, ok := a.activeState.(UIDrawer); ok {
			s.ShowImgui()
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
			ctx := post.PostProcessContext{
				GBuffer:   a.deferred.GetFramebuffer(),
				Result:    a.deferred.GetResult(),
				Camera:    state.GetCamera(),
				DeltaTime: a.monitor.GetDeltaTime(),
			}
			_, lastRenderTarget = a.pipeline.Execute(ctx)

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

// ---- AppAPI implementations

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
	logger.Init(slog.LevelDebug)

	if a.cfg == nil {
		logger.Warnf("failed to load config -> we init with default one")
		a.cfg = &DefaultConfig
	}

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
		a.cfg.Window.Width,
		a.cfg.Window.Height,
		WindowTitle,
		a.cfg.Window.Ratio,
	)
	if err != nil {
		return fmt.Errorf("failed init window %v", err)
	}
	a.window.MakeContextCurrent()
	a.window.Focus()
	a.window.Center()

	// load icon
	err = a.window.SetIconFromFile(files.GetTexturePath("initial/icon.png"))
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
	for _, n := range a.pipeline.GetExecutionList() {
		var editor ui.ConfigUI

		switch p := n.(type) {
		case *post.EyeAdaptionPass:
			editor = &ui.EyeAdaptionConfigUI{EyeAdaptionConfig: p.GetConfig().(*post.EyeAdaptionConfig)}
		case *post.SSAOPass:
			editor = &ui.SSAOConfigUI{SSAOConfig: p.GetConfig().(*post.SSAOConfig)}
		case *post.CavityPass:
			editor = &ui.CavityConfigUI{CavityConfig: p.GetConfig().(*post.CavityConfig)}
		case *post.BloomPass:
			editor = &ui.BloomConfigUI{BloomConfig: p.GetConfig().(*post.BloomConfig)}
		case *post.ToneMappingPass:
			editor = &ui.ToneMappingConfigUI{ToneMappingConfig: p.GetConfig().(*post.ToneMappingConfig)}
		case *post.ColorGradingPass:
			editor = &ui.ColorGradingUI{ColorGradingConfig: p.GetConfig().(*post.ColorGradingConfig)}
		case *post.VignettePass:
			editor = &ui.VignetteUI{VignetteConfig: p.GetConfig().(*post.VignetteConfig)}
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
			a.pipeline.PostProcessingPipeline,
			a.window.GetConfig(),
		),
	)
}

// -- Destroy

func (a *App) Destroy() {

	if a.activeState != nil {
		a.activeState.Destroy()
	}

	// clear deferred fbo
	a.deferred.Delete()

	// clear post process pipeline
	a.pipeline.Delete()

	ui.Destroy()
	a.window.Destroy()
	glfw.Terminate()
}
