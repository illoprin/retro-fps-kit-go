package engine

import (
	"fmt"
	"runtime"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/illoprin/retro-fps-kit-go/src/app"
	"github.com/illoprin/retro-fps-kit-go/src/core/config"
	"github.com/illoprin/retro-fps-kit-go/src/core/files"
	"github.com/illoprin/retro-fps-kit-go/src/core/monitor"
	"github.com/illoprin/retro-fps-kit-go/src/core/window"
	"github.com/illoprin/retro-fps-kit-go/src/engine/ui"
	"github.com/illoprin/retro-fps-kit-go/src/render/context"
	"github.com/illoprin/retro-fps-kit-go/src/render/passes"
	"github.com/illoprin/retro-fps-kit-go/src/render/pipeline"
	"github.com/illoprin/retro-fps-kit-go/src/render/rhi"
)

type Engine struct {
	window  *window.Window
	input   *window.InputManager
	monitor *monitor.Monitor

	// initial imgui debug screen
	iUI *ui.InitialUI

	// application state (level, game menu, editor)
	activeState app.AppState

	// initial rhi resources
	resources []rhi.Resource

	// geometry render pass
	deferred *pipeline.DeferredRenderTarget

	// post processing render pass
	passPipeline []passes.PostProcessingPass
}

func NewEngine() (*Engine, error) {
	runtime.LockOSThread()

	e := &Engine{
		resources: make([]rhi.Resource, 0),
	}

	if err := e.initSystems(); err != nil {
		return nil, err
	}

	if err := e.initRenderingPipeline(); err != nil {
		return nil, fmt.Errorf("failed to init rendering pipeline - %v", err)
	}

	// setup custom gui
	e.iUI, _ = ui.NewInitialUI()

	// setup callbacks for input manager
	e.setupCallbacks()

	return e, nil
}

// ---- Callbacks

func (e *Engine) setupCallbacks() {
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

func (e *Engine) initRenderingPipeline() error {

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
		files.GetShaderPath("quad.vert"),
		files.GetShaderPath("mask_blur.frag"),
	)
	if err != nil {
		return err
	}
	// compositor - uses ssao and cavity
	compProg, err := rhi.NewProgram(
		files.GetShaderPath("quad.vert"),
		files.GetShaderPath("mask_compositor.frag"),
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
		e.window.GetConfig(), meshQuad, ssaoConfig,
		noiseTexture, blurProg, compProg)); err != nil {
		return fmt.Errorf("ssao pass - %w", err)
	}

	// -- crease occlusion

	if err := addPass(passes.NewCreaseOcclusionPass(
		e.window.GetConfig(), meshQuad, creaseConfig,
		noiseTexture, blurProg, compProg,
	)); err != nil {
		return fmt.Errorf("crease pass - %w", err)
	}

	// -- color grading

	if err := addPass(passes.NewColorGradingPass(
		e.window.GetConfig(), meshQuad, colorGradingConfig,
	)); err != nil {
		return fmt.Errorf("color grading pass - %w", err)
	}

	// -- vignette

	if err := addPass(passes.NewVignettePass(
		e.window.GetConfig(), meshQuad, vignetteConfig,
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

func (e *Engine) resizeRenderTargets() {
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

func (e *Engine) Run() {

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

		e.iUI.NewFrame()

		// update current context
		if e.activeState != nil {
			e.activeState.Update()
		}

		// resize targets if screen ratio changes
		if e.window.GetConfig().LastResolutionRatio != e.window.GetConfig().ResolutionRatio {
			e.resizeRenderTargets()
		}

		// -- Render

		context.BindFramebuffer(nil)

		// perform custom indirect rendering of current state (if needs)
		if state, ok := e.activeState.(app.IndirectDrawer); ok {
			context.SaveState() // WARN not notimplemented
			state.RenderIndirect()
			context.RestoreState()
		}

		// perform gbuffer rendering of current state (if needs)
		var lastRenderTarget *rhi.Framebuffer
		if state, ok := e.activeState.(app.GBufferDrawer); ok {
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
		if state, ok := e.activeState.(app.FlatDrawer); ok {
			// render flat of current app state
			state.RenderFlat(lastRenderTarget)
		}

		// blit last render target to initial framebuffer
		lastRenderTarget.Blit(0, e.window.GetConfig().Width, e.window.GetConfig().Height, rhi.FilterNearest)

		// render imgui on top of screen
		e.iUI.Draw()

		e.window.SwapBuffers()

		// -- End Frame

		e.monitor.Update()
		rhi.FrameStats.Reset()
	}
}

// performPostProcessingPipeline applies post processing
// and returns result texture and framebuffer
func (e *Engine) performPostProcessingPipeline() (*rhi.Texture, *rhi.Framebuffer) {
	s, ok := e.activeState.(app.GBufferDrawer)
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

func (e *Engine) Close() {
	e.window.SetShouldClose(true)
}

func (e *Engine) SetActiveState(s app.AppState) {
	e.activeState = s
}

func (e *Engine) GetConfig() *config.Config {
	return nil
}

func (e *Engine) GetWindow() *window.Window {
	return e.window
}

func (e *Engine) GetInputManager() *window.InputManager {
	return e.input
}

func (e *Engine) GetMonitor() *monitor.Monitor {
	return e.monitor
}

func (e *Engine) GetGBuffer() *rhi.Framebuffer {
	return e.deferred.GetFramebuffer()
}

func (e *Engine) GUIWantCaptureMouse() bool {
	io := imgui.CurrentIO()
	return io.WantCaptureMouse()
}

func (e *Engine) GetTime() float64 {
	return glfw.GetTime()
}

// ---- Auxiliary methods

func (e *Engine) initSystems() error {
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

func (e *Engine) createWindow() error {
	// create window
	var err error
	e.window, err = window.NewWindow(
		WindowWidth,
		WindowHeight,
		WindowTitle,
		DefaultResolutionRatio,
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

func (e *Engine) initImguiContext() error {
	// init imgui context
	ui.InitImgui()

	// init imgui renderer
	if err := ui.InitImguiRenderer(e.window); err != nil {
		return fmt.Errorf("failed to init imgui renderer - %v", err)
	}

	return nil
}

// -- Destroy

func (e *Engine) Destroy() {

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

	ui.DestroyImgui()
	e.window.Destroy()
	glfw.Terminate()
}
