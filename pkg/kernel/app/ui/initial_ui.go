package ui

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/files"
)

type InitialUI struct {
	font    *imgui.Font
	debugUI *DebugUI
	fbosUI  *FramebuffersUI
}

func NewInitialUI() (*InitialUI, error) {
	// create
	iui := &InitialUI{}

	// load custom font
	io := imgui.CurrentIO()
	iui.font = io.Fonts().AddFontFromFileTTF(files.GetFontPath("ui.ttf"))
	if !iui.font.IsLoaded() {
		return nil, fmt.Errorf("failed to load imgui font")
	}

	iui.customize()

	return iui, nil
}

func (ui *InitialUI) customize() {
	style := imgui.CurrentStyle()
	colors := style.Colors()

	// customize color
	// colors[imgui.ColWindowBg] = imgui.Vec4{0.302, 0.302, 0.302, 0.8}

	// customize other params
	style.SetWindowRounding(8)
	style.SetFrameRounding(5)
	style.SetTabRounding(5)

	style.SetColors(&colors)
}

func (ui *InitialUI) AttachDebugUI(d *DebugUI) {
	ui.debugUI = d
}

func (ui *InitialUI) AttachFramebuffersUI(f *FramebuffersUI) {
	ui.fbosUI = f
}

func (ui *InitialUI) OnKey(key glfw.Key, action glfw.Action) {
	if action == glfw.Press {
		if key == glfw.KeyF1 {
			ui.debugUI.Visible = !ui.debugUI.Visible
			ui.fbosUI.Visible = !ui.fbosUI.Visible
		}
	}
}

func (ui *InitialUI) GetDebugUI() *DebugUI {
	return ui.debugUI
}

func (ui *InitialUI) GetFramebuffersUI() *FramebuffersUI {
	return ui.fbosUI
}

func (ui *InitialUI) Draw() {
	// apply custom font
	imgui.PushFont(ui.font, 18.0)
	ui.debugUI.Show()
	ui.fbosUI.Show()
	// detach custom font
	imgui.PopFont()
}
