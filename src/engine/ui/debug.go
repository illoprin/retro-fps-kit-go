package ui

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/illoprin/retro-fps-kit-go/src/core/window"
	"github.com/illoprin/retro-fps-kit-go/src/engine/controllers"
	"github.com/illoprin/retro-fps-kit-go/src/render/passes"
	"github.com/illoprin/retro-fps-kit-go/src/render/pipeline"
)

// DebugMenu manages all debug UI elements
type DebugMenu struct {
	Visible bool

	// Dependencies
	windowCfg       *window.ScreenConfig
	deferred        *pipeline.DeferredRenderTarget
	controller      *controllers.EditorController
	colorGradingCfg *passes.ColorGradingConfig
	vignetteCfg     *passes.VignetteConfig
	ssaoCfg         *passes.SSAOConfig
	creaseCfg       *passes.CreaseOcclusionConfig

	// Textures for display
	deferredTextures []ImageTexture
	passTextures     []ImageTexture

	// UI State
	showPerformance  bool
	showDrawCalls    bool
	showVertices     bool
	showController   bool
	showGameObjects  bool
	showCamera       bool
	showColorGrading bool
	showVignette     bool
	showSSAO         bool
	showCrease       bool
	showDeferredTex  bool
	showPPTex        bool
}

// NewDebugMenu creates a new debug menu instance
func NewDebugMenu(
	windowCfg *window.ScreenConfig,
	deferred *pipeline.DeferredRenderTarget,
	controller *controllers.EditorController,
	ssaoCfg *passes.SSAOConfig,
	creaseCfg *passes.CreaseOcclusionConfig,
	colorGradingCfg *passes.ColorGradingConfig,
	vignetteCfg *passes.VignetteConfig,
	deferredTextures []ImageTexture,
	passTextures []ImageTexture,
) *DebugMenu {
	return &DebugMenu{
		Visible:          true,
		windowCfg:        windowCfg,
		deferred:         deferred,
		controller:       controller,
		colorGradingCfg:  colorGradingCfg,
		vignetteCfg:      vignetteCfg,
		ssaoCfg:          ssaoCfg,
		creaseCfg:        creaseCfg,
		deferredTextures: deferredTextures,
		passTextures:     passTextures,

		// Initialize all collapsed states to true
		showPerformance:  true,
		showDrawCalls:    true,
		showVertices:     true,
		showController:   true,
		showGameObjects:  true,
		showCamera:       true,
		showColorGrading: true,
		showVignette:     true,
		showSSAO:         true,
		showCrease:       true,
		showDeferredTex:  true,
		showPPTex:        true,
	}
}

// Show renders the debug window if visible
func (m *DebugMenu) Show() {
	if m.Visible {
		imgui.SetNextWindowSizeConstraints(imgui.Vec2{300, 600}, imgui.Vec2{500, 1000})
		imgui.SetNextWindowPosV(imgui.Vec2{30, 30}, imgui.CondFirstUseEver, imgui.Vec2{0, 0})
		m.showDebugWindow()
		imgui.SetNextWindowPosV(imgui.Vec2{1000, 30}, imgui.CondFirstUseEver, imgui.Vec2{0, 0})
		imgui.SetNextWindowSizeConstraints(imgui.Vec2{400, 600}, imgui.Vec2{1000, 700})
		m.showTexturesWindow()
	}
}

// showDebugWindow creates the main debug window with tabs
func (m *DebugMenu) showDebugWindow() {
	imgui.Begin("Debug")

	if imgui.BeginTabBar("DebugTabBar") {

		if imgui.BeginTabItem("Stats") {
			m.renderStatsTab()
			imgui.EndTabItem()
		}

		if imgui.BeginTabItem("Scene") {
			m.renderSceneTab()
			imgui.EndTabItem()
		}

		if imgui.BeginTabItem("Post Processing") {
			m.renderPostProcessingTab()
			imgui.EndTabItem()
		}

		imgui.EndTabBar()
	}

	imgui.End()

}

// window shows different render targets
func (m *DebugMenu) showTexturesWindow() {
	imgui.Begin("Render Targets")

	if imgui.BeginTabBar("TexturesTabBar") {
		if imgui.BeginTabItem("Deferred") {
			m.renderTextures(m.deferredTextures)
			imgui.EndTabItem()
		}

		if imgui.BeginTabItem("Post Processing") {
			m.renderTextures(m.passTextures)
			imgui.EndTabItem()
		}
		imgui.EndTabBar()
	}
	imgui.End()
}

// renderStatsTab renders performance statistics
func (m *DebugMenu) renderStatsTab() {
	io := imgui.CurrentIO()

	if imgui.CollapsingHeaderBoolPtr("Performance", &m.showPerformance) {
		imgui.Text(fmt.Sprintf("FPS: %.2f", io.Framerate()))
		imgui.Text(fmt.Sprintf("Frame Time: %.3f ms", io.DeltaTime()*1000))
		imgui.Text(fmt.Sprintf("Resolution: %dx%d (%.0f%%)",
			m.windowCfg.Width, m.windowCfg.Height,
			m.windowCfg.ResolutionRatio*100))
	}

	if imgui.CollapsingHeaderBoolPtr("Draw Calls", &m.showDrawCalls) {
		imgui.Text(fmt.Sprintf("Scene: %d", global.LastDrawCalls))
		imgui.Text(fmt.Sprintf("ImGUI: %d", global.LastImguiDrawCalls))
	}

	if imgui.CollapsingHeaderBoolPtr("Vertices", &m.showVertices) {
		imgui.Text(fmt.Sprintf("Scene: %d", global.LastVertices))
		imgui.Text(fmt.Sprintf("ImGUI: %d", global.LastImguiVertices))
	}
}

// renderSceneTab renders scene and camera controls
func (m *DebugMenu) renderSceneTab() {
	if imgui.CollapsingHeaderBoolPtr("Controller", &m.showController) {
		imgui.SliderFloat("Speed", &m.controller.Speed, 1, 20)
		imgui.SliderFloat("Sensitivity", &m.controller.Sensitivity, 0.01, 1)
	}

	if imgui.CollapsingHeaderBoolPtr("Game Objects", &m.showGameObjects) {
		imgui.Text("Game objects list coming soon...")
	}

	if imgui.CollapsingHeaderBoolPtr("Camera", &m.showCamera) {
		imgui.PushItemFlag(imgui.ItemFlags(imgui.ItemFlagsReadOnly), true)
		cam := m.controller.GetCamera()
		front := cam.GetFront()
		p, y := cam.GetRotation()

		imgui.DragFloat3("Position", (*[3]float32)(&cam.Position))
		imgui.DragFloat3("Front", (*[3]float32)(&front))

		rot := [2]float32{p, y}
		imgui.DragFloat2("Rotation", &rot)

		imgui.Text(fmt.Sprintf("FOV: %.2f", cam.Fov))
		imgui.PopItemFlag()
	}
}

// renderPostProcessingTab renders post-processing effect controls
func (m *DebugMenu) renderPostProcessingTab() {
	// Wireframe control
	imgui.Checkbox("Wireframe", &m.deferred.Wireframe)

	// Resolution ratio
	imgui.SliderFloat("Resolution Ratio", &m.windowCfg.ResolutionRatio, 0.2, 1)

	if imgui.CollapsingHeaderBoolPtr("SSAO", &m.showSSAO) {
		imgui.Checkbox("ao_Use", &m.ssaoCfg.Use)
		imgui.SliderFloat("ao_Radius ", &m.ssaoCfg.Radius, 0.02, 2)
		imgui.SliderFloat("ao_Bias", &m.ssaoCfg.Bias, 0.0001, 0.5)
		imgui.SliderInt("ao_Samples", &m.ssaoCfg.KernelSize, 4, 64)
		imgui.SliderFloat("ao_Blackpoint", &m.ssaoCfg.BlackPoint, 0, 1)
		imgui.SliderFloat("ao_Whitepoint", &m.ssaoCfg.WhitePoint, 0, 1)
		imgui.SliderInt("ao_BlurSize", &m.ssaoCfg.BlurSize, 1, 8)
	}

	if imgui.CollapsingHeaderBoolPtr("Crease", &m.showCrease) {
		imgui.Checkbox("c_Use", &m.creaseCfg.Use)
		imgui.SliderFloat("c_Radius", &m.creaseCfg.Radius, 5, 100)
		imgui.SliderFloat("c_Bias ", &m.creaseCfg.DepthBias, 0.0001, 0.1)
		imgui.SliderFloat("c_Intensity", &m.creaseCfg.Intensity, 0.01, 10)
		imgui.SliderInt("c_KernelSize", &m.creaseCfg.KernelSize, 1, 256)
		imgui.SliderInt("c_BlurSize", &m.creaseCfg.BlurSize, 1, 8)
		imgui.SliderFloat("c_WhitePoint", &m.creaseCfg.WhitePoint, 0.0, 1.0)
		imgui.SliderFloat("c_BlackPoint", &m.creaseCfg.BlackPoint, 0.0, 1.0)
	}

	if imgui.CollapsingHeaderBoolPtr("Color Grading", &m.showColorGrading) {
		imgui.Checkbox("cg_Use", &m.colorGradingCfg.Use)
		imgui.SliderFloat("cg_Gamma", &m.colorGradingCfg.Gamma, 1, 3)
		imgui.SliderFloat("cg_Exposure", &m.colorGradingCfg.Exposure, 0.5, 4)
		imgui.SliderFloat("cg_Contrast", &m.colorGradingCfg.Contrast, 0.8, 3)
		imgui.SliderFloat("cg_Saturation", &m.colorGradingCfg.Saturation, 0, 2)
		imgui.SliderFloat("cg_Brightness", &m.colorGradingCfg.Brightness, 0.5, 10)
		imgui.ColorEdit3("cg_Shadows", &m.colorGradingCfg.ShadowsColor)
		imgui.ColorEdit3("cg_Midtones", &m.colorGradingCfg.MidColor)
		imgui.ColorEdit3("cg_Highlights", &m.colorGradingCfg.HighlightColor)
		imgui.SliderFloat("cg_ColorStrength", &m.colorGradingCfg.ColorStrength, 0, 1.1)
	}

	if imgui.CollapsingHeaderBoolPtr("Vignette", &m.showVignette) {
		imgui.Checkbox("v_Use", &m.vignetteCfg.Use)
		imgui.SliderFloat("v_Radius", &m.vignetteCfg.Radius, 0.2, 2)
		imgui.SliderFloat("v_Softness", &m.vignetteCfg.Softness, 0.01, 1)
	}

}
