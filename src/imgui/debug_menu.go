package imgui

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/illoprin/retro-fps-kit-go/src/engine/global"
	"github.com/illoprin/retro-fps-kit-go/src/player"
)

type DebugMenu struct {
	Visible          bool
	DestroyingEffect bool
	c                *player.EditorController
}

func NewDebugMenu(c *player.EditorController) *DebugMenu {

	return &DebugMenu{true, false, c}
}

func (m *DebugMenu) Show() {
	// Create the window
	if m.Visible {
		imgui.SetNextWindowSizeConstraints(imgui.Vec2{200, 250}, imgui.Vec2{1000, 700})
		m.showDebugWindow()
	}
}

func (m *DebugMenu) showDebugWindow() {
	imgui.Begin("Debug")
	// tab bar
	if imgui.BeginTabBar("DebugTabBar") {

		if imgui.BeginTabItem("Stats") {
			m.barStats()
		}

		if imgui.BeginTabItem("Scene") {
			m.barScene()
		}

		if imgui.BeginTabItem("Textures") {
			m.barTextures()
		}

		if imgui.BeginTabItem("Framebuffers") {
			m.barFramebuffers()
		}

		imgui.EndTabBar()
	}

	imgui.End()
}

func (m *DebugMenu) getCameraStatsString() string {
	cam := m.c.GetCamera()
	p, y := cam.GetRotation()
	return fmt.Sprintf(
		"X: %.2f\nY: %.2f\nZ: %.2f\nPitch: %.2f\nYaw: %.2f\nFOV: %.2f",
		cam.Position[0], cam.Position[1], cam.Position[2], p, y, cam.Fov,
	)
}

func (m *DebugMenu) barStats() {

	io := imgui.CurrentIO()

	sVty := true
	if imgui.CollapsingHeaderBoolPtr("Performance", &sVty) {
		imgui.Text(fmt.Sprintf("FPS: %.2f\nFrame Time: %.3f ms", io.Framerate(), io.DeltaTime()))
	}
	dcVty := true
	if imgui.CollapsingHeaderBoolPtr("Draw Calls", &dcVty) {
		imgui.Text(fmt.Sprintf("Scene: %d\nImGUI: %d", global.LastDrawCalls, global.LastImguiDrawCalls))
	}
	vVty := true
	if imgui.CollapsingHeaderBoolPtr("Vertices", &vVty) {
		imgui.Text(fmt.Sprintf("Scene: %d\nImGUI: %d", global.LastVertices, global.LastImguiVertices))
	}
	cVty := true
	if imgui.CollapsingHeaderBoolPtr("Camera", &cVty) {
		imgui.Text(m.getCameraStatsString())
	}
	imgui.EndTabItem()

}

func (m *DebugMenu) barScene() {
	camVty := true
	if imgui.CollapsingHeaderBoolPtr("Controller", &camVty) {
		imgui.SliderFloat("Speed", &m.c.Speed, 1, 20)
		imgui.SliderFloat("Sensitivity", &m.c.Sensitivity, 0.01, 1)
	}
	if imgui.CollapsingHeaderBoolPtr("Game objects", &camVty) {
		imgui.Checkbox("Shotgun destroying effect", &m.DestroyingEffect)
	}
	imgui.EndTabItem()

}

func (m *DebugMenu) barTextures() {
	imgui.Text("nil")
	imgui.EndTabItem()
}

func (m *DebugMenu) barFramebuffers() {
	imgui.Text("nil")
	imgui.EndTabItem()
}
