package ui

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/core/camera"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/core/monitor"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/core/window"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/render/pipeline"
)

// DebugUI manages all debug UI elements
type DebugUI struct {
	Visible bool

	// Dependencies
	windowCfg    *window.ScreenConfig
	monitor      *monitor.Monitor
	deferred     *pipeline.DeferredRenderTarget
	passesUI     []ConfigUI
	passesShow   []bool
	activeCamera *camera.Camera3D

	// UI State
	showPerformance bool
	showCamera      bool
	showDrawCalls   bool
	showVertices    bool
}

// NewDebugUI creates a new debug menu instance
func NewDebugUI(
	windowCfg *window.ScreenConfig,
	monitor *monitor.Monitor,
	deferred *pipeline.DeferredRenderTarget,
	passesUI []ConfigUI,
) *DebugUI {
	d := &DebugUI{
		windowCfg: windowCfg,
		deferred:  deferred,
		monitor:   monitor,
		passesUI:  passesUI,

		// Initialize all collapsed states to true
		showPerformance: true,
		showCamera:      true,
		showDrawCalls:   true,
		showVertices:    true,

		// set visibility
		Visible: true,
	}

	// prepare passes array
	d.passesShow = make([]bool, len(passesUI))
	for i, _ := range d.passesShow {
		d.passesShow[i] = true
	}

	return d
}

// Show shows debug ui
func (m *DebugUI) Show() {

	if m.Visible {
		// -- Debug Window
		imgui.SetNextWindowSizeConstraints(
			imgui.Vec2{300, 600}, imgui.Vec2{500, 1000},
		)
		imgui.SetNextWindowPosV(
			imgui.Vec2{30, 30}, imgui.CondFirstUseEver, imgui.Vec2{0, 0},
		)
		m.showDebugWindow()
	}

}

// showDebugWindow creates the main debug window with tabs
func (m *DebugUI) showDebugWindow() {
	imgui.Begin("Debug")

	if imgui.BeginTabBar("DebugTabBar") {

		if imgui.BeginTabItem("Stats") {
			m.showStatsTab()
			imgui.EndTabItem()
		}

		if imgui.BeginTabItem("Camera") {
			m.showSceneTab()
			imgui.EndTabItem()
		}

		if imgui.BeginTabItem("Rendering") {
			m.showRenderingTab()
			imgui.EndTabItem()
		}

		imgui.EndTabBar()
	}

	imgui.End()

}

// renderStatsTab renders performance statistics
func (m *DebugUI) showStatsTab() {

	imgui.Text(fmt.Sprintf("FPS: %.2f", m.monitor.GetFPS()))
	imgui.Text(fmt.Sprintf("Frame Time: %.3f ms", m.monitor.GetFrameTime()*1000))

	imgui.Separator()

	imgui.Text(fmt.Sprintf("Draw Calls: %d", m.monitor.GetDrawCalls()))
	imgui.Text(fmt.Sprintf("Vertices: %d", m.monitor.GetVertices()))
	imgui.Text(fmt.Sprintf("Triangles: %d", m.monitor.GetTriangles()))
}

func (d *DebugUI) SetActiveCamera(c *camera.Camera3D) {
	d.activeCamera = c
}

// renderSceneTab renders scene and camera controls
func (m *DebugUI) showSceneTab() {

	if imgui.CollapsingHeaderBoolPtr("Camera", &m.showCamera) {

		if m.activeCamera == nil {
			imgui.Text("No active 3D camera data was received")
			return
		}

		imgui.PushItemFlag(imgui.ItemFlags(imgui.ItemFlagsReadOnly), true)
		front := m.activeCamera.GetFront()
		p, y := m.activeCamera.GetRotation()

		imgui.DragFloat3("Position", (*[3]float32)(&m.activeCamera.Position))
		imgui.DragFloat3("Front", (*[3]float32)(&front))

		rot := [2]float32{p, y}
		imgui.DragFloat2("Rotation", &rot)

		imgui.Text(fmt.Sprintf("FOV: %.2f", m.activeCamera.Fov))
		imgui.PopItemFlag()
	}
}

// renderPostProcessingTab renders post-processing effect controls
func (d *DebugUI) showRenderingTab() {
	// Wireframe control
	imgui.Checkbox("Wireframe", &d.deferred.Wireframe)

	// Resolution ratio
	imgui.SliderFloat("Resolution Ratio", &d.windowCfg.ResolutionRatio, 0.2, 1)

	// Passes UI
	for i, o := range d.passesUI {
		if imgui.CollapsingHeaderBoolPtr(o.GetName(), &d.passesShow[i]) {
			o.ShowUI()
		}
	}

}
