package postprocessing

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-kit-go/src/engine/assetmgr"
	"github.com/illoprin/retro-fps-kit-go/src/engine/global"
	"github.com/illoprin/retro-fps-kit-go/src/render"
	"github.com/illoprin/retro-fps-kit-go/src/window"
)

type ColorGrading struct {
	Gamma      float32
	Exposure   float32
	Contrast   float32
	Saturation float32
	Brightness float32
}

type SceneMixerConfig struct {
	lastResolutionRatio float32
	ResolutionRatio     float32
	Vignette            struct {
		Radius, Smooth float32
		Use            bool
	}
	Flickering struct {
		Frequency, Intensity float32
		Use                  bool
	}
	SceneClearColor mgl32.Vec3
	Wireframe       bool
}

type SceneMixer struct {
	cfg             *SceneMixerConfig
	cg              *ColorGrading
	w               *window.Window
	SceneFBO        *render.Framebuffer
	screenProgram   *render.Program
	screenQuad      *render.Mesh
	sceneRenderFunc func()
	resources       []render.Resource
}

func (m *SceneMixer) GetConfig() *SceneMixerConfig {
	return m.cfg
}

func (m *SceneMixer) GetColorGrading() *ColorGrading {
	return m.cg
}

func NewSceneMixer(win *window.Window, cfg *SceneMixerConfig, cg *ColorGrading) (*SceneMixer, error) {
	m := &SceneMixer{w: win, cg: cg, cfg: cfg}

	var err error
	err = m.initFramebuffers()
	err = m.setupScreen()
	if err != nil {
		return nil, err
	}
	m.setupCallbacks()

	m.cfg.lastResolutionRatio = m.cfg.ResolutionRatio

	m.resources = make([]render.Resource, 3)
	m.resources[0] = m.SceneFBO
	m.resources[1] = m.screenQuad
	m.resources[2] = m.screenProgram

	return m, nil
}

func (m *SceneMixer) initFramebuffers() error {

	winWidth, winHeight := m.w.GetSize()

	// init scene framebuffer
	sceneFBO, err := render.NewFramebuffer(render.FramebufferConfig{
		Width:           int32(float32(winWidth) * m.cfg.ResolutionRatio),
		Height:          int32(float32(winHeight) * m.cfg.ResolutionRatio),
		ColorFormat:     render.FormatRGB8,
		ColorFiltering:  render.FilterNearest,
		UseDepth:        true,
		UseDepthStencil: false,
		UseMultisample:  false,
	})
	if err != nil {
		return err
	}
	m.SceneFBO = sceneFBO

	return nil
}

func (m *SceneMixer) setupCallbacks() {

	prevCallback := m.w.SetFramebufferSizeCallback(nil)
	m.w.SetFramebufferSizeCallback(func(w *glfw.Window, width, height int) {
		m.resizeSceneFBO(width, height)
		if prevCallback != nil {
			prevCallback(w, width, height)
		}
	})
}

func (m *SceneMixer) resizeSceneFBO(w, h int) {
	m.SceneFBO.Resize(
		int32(float32(w)*m.cfg.ResolutionRatio),
		int32(float32(h)*m.cfg.ResolutionRatio),
	)
}

func (m *SceneMixer) setupScreen() error {

	// init screen quad mesh
	m.screenQuad = render.NewMesh()
	m.screenQuad.SetupBasicQuad()

	// init screen quad shader
	screenProg, err := render.NewProgram(assetmgr.GetShaderPath("screen_quad.vert"), assetmgr.GetShaderPath("screen_quad.frag"))
	if err != nil {
		return err
	}
	m.screenProgram = screenProg

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)
	gl.FrontFace(gl.CCW)
	return nil
}

func (m *SceneMixer) Render() {
	m.SceneFBO.Bind()
	m.newSceneFrame()
	m.sceneRenderFunc()
	m.SceneFBO.Unbind()
	m.renderSceneQuad()

	m.cfg.lastResolutionRatio = m.cfg.ResolutionRatio
}

func (m *SceneMixer) SetSceneRenderFunc(f func()) {
	m.sceneRenderFunc = f
}

func (m *SceneMixer) newSceneFrame() {

	w, h := m.w.GetSize()
	gl.Viewport(0, 0,
		int32(float32(w)*m.cfg.ResolutionRatio),
		int32(float32(h)*m.cfg.ResolutionRatio),
	)
	if m.cfg.ResolutionRatio != m.cfg.lastResolutionRatio {
		m.resizeSceneFBO(w, h)
	}

	gl.ClearColor(
		m.cfg.SceneClearColor[0],
		m.cfg.SceneClearColor[1],
		m.cfg.SceneClearColor[2],
		1.0,
	)

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
	if m.cfg.Wireframe {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
		gl.Disable(gl.CULL_FACE)
	} else {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
	}

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

}

func (m *SceneMixer) renderSceneQuad() {
	w, h := m.w.GetSize()

	gl.Disable(gl.DEPTH_TEST)
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
	gl.ClearColor(0, 0, 0, 1.0)
	gl.Viewport(0, 0, int32(w), int32(h))

	gl.Clear(gl.COLOR_BUFFER_BIT)

	m.screenProgram.Use()
	m.SceneFBO.ColorTexture.Bind(0)
	m.screenProgram.Set1i("u_color", 0)
	m.screenProgram.Set1f("u_vignette.radius", m.cfg.Vignette.Radius)
	m.screenProgram.Set1f("u_vignette.softness", m.cfg.Vignette.Smooth)
	m.screenProgram.Set1i("u_vignette.use", global.BoolToInt32(m.cfg.Vignette.Use))
	m.screenProgram.Set1f("u_color_grading.contrast", m.cg.Contrast)
	m.screenProgram.Set1f("u_color_grading.saturation", m.cg.Saturation)
	m.screenProgram.Set1f("u_color_grading.brightness", m.cg.Brightness)

	// m.screenProgram.Set1f("u_time", float32(glfw.GetTime()))
	m.screenQuad.Draw()

}

func (m *SceneMixer) Delete() {
	for _, r := range m.resources {
		if r != nil {
			r.Delete()
		}
	}
}
