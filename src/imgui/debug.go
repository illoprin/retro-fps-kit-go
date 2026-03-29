package imgui

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/illoprin/retro-fps-kit-go/src/engine/global"
	"github.com/illoprin/retro-fps-kit-go/src/player"
	postprocessing "github.com/illoprin/retro-fps-kit-go/src/post_processing"
	"github.com/illoprin/retro-fps-kit-go/src/renderers"
	"github.com/illoprin/retro-fps-kit-go/src/window"
)

// DebugMenu manages all debug UI elements
type DebugMenu struct {
	Visible bool

	// Dependencies
	windowCfg       *window.ScreenConfig
	deferred        *renderers.DeferredRenderTarget
	controller      *player.EditorController
	colorGradingCfg *postprocessing.ColorGradingConfig
	vignetteCfg     *postprocessing.VignetteConfig
	ssaoCfg         *postprocessing.SSAOConfig

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
	showDeferredTex  bool
	showPPTex        bool
}

// NewDebugMenu creates a new debug menu instance
func NewDebugMenu(
	windowCfg *window.ScreenConfig,
	deferred *renderers.DeferredRenderTarget,
	controller *player.EditorController,
	ssaoCfg *postprocessing.SSAOConfig,
	colorGradingCfg *postprocessing.ColorGradingConfig,
	vignetteCfg *postprocessing.VignetteConfig,
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
		showDeferredTex:  true,
		showPPTex:        true,
	}
}

// Show renders the debug window if visible
func (m *DebugMenu) Show() {
	if m.Visible {
		imgui.SetNextWindowSizeConstraints(imgui.Vec2{200, 250}, imgui.Vec2{1000, 700})
		m.showDebugWindow()
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

		if imgui.BeginTabItem("Deferred Textures") {
			m.renderTextures(m.deferredTextures)
			imgui.EndTabItem()
		}

		if imgui.BeginTabItem("PP Textures") {
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
	imgui.SliderFloat("Resolution Ratio", &m.windowCfg.ResolutionRatio, 0.01, 1)

	if imgui.CollapsingHeaderBoolPtr("Color Grading", &m.showColorGrading) {
		imgui.Checkbox("Use Color Grading", &m.colorGradingCfg.Use)
		imgui.SliderFloat("Gamma", &m.colorGradingCfg.Gamma, 1, 3)
		imgui.SliderFloat("Exposure", &m.colorGradingCfg.Exposure, 0.5, 4)
		imgui.SliderFloat("Contrast", &m.colorGradingCfg.Contrast, 0.8, 3)
		imgui.SliderFloat("Saturation", &m.colorGradingCfg.Saturation, 0, 2)
		imgui.SliderFloat("Brightness", &m.colorGradingCfg.Brightness, 0.5, 10)
		imgui.ColorEdit3("Shadows", &m.colorGradingCfg.ShadowsColor)
		imgui.ColorEdit3("Midtones", &m.colorGradingCfg.MidColor)
		imgui.ColorEdit3("Highlights", &m.colorGradingCfg.HighlightColor)
		imgui.SliderFloat("Color Strength", &m.colorGradingCfg.ColorStrength, 0, 1.1)
	}

	if imgui.CollapsingHeaderBoolPtr("Vignette", &m.showVignette) {
		imgui.Checkbox("Use Vignette", &m.vignetteCfg.Use)
		imgui.SliderFloat("Radius", &m.vignetteCfg.Radius, 0.2, 2)
		imgui.SliderFloat("Softness", &m.vignetteCfg.Softness, 0.01, 1)
	}

	if imgui.CollapsingHeaderBoolPtr("SSAO", &m.showSSAO) {
		imgui.Checkbox("Use SSAO", &m.ssaoCfg.Use)
		imgui.SliderFloat("Radius ", &m.ssaoCfg.Radius, 0.02, 2)
		imgui.SliderFloat("Bias", &m.ssaoCfg.Bias, 0.0001, 0.1)
		imgui.SliderInt("Samples", &m.ssaoCfg.KernelSize, 4, 64)
	}
}

// renderTextures displays image textures slice
func (m *DebugMenu) renderTextures(textures []ImageTexture) {

	if len(textures) == 0 {
		imgui.Text("No textures available")
		return
	}

	for _, tex := range textures {
		imgui.Text(tex.Name)

		// Create texture reference for ImGui
		textureRef := imgui.NewEmptyTextureRef()
		textureRef.SetTexID(imgui.TextureID(tex.ID))

		// Calculate display size (maintain aspect ratio)
		displaySize := float32(512)
		aspect := float32(m.windowCfg.Aspect)

		imgui.ImageV(*textureRef,
			imgui.Vec2{aspect * displaySize, displaySize},
			imgui.Vec2{0, 1},
			imgui.Vec2{1, 0})
	}
}
