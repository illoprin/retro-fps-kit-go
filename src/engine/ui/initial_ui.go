package ui

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	implgl3 "github.com/AllenDang/cimgui-go/impl/opengl3"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/illoprin/retro-fps-kit-go/src/core/files"
)

type InitialUI struct {
	font      *imgui.Font
	debugMenu *DebugMenu
	stats     *ImguiStats
}

func NewInitialUI() (*InitialUI, error) {
	// create
	iui := &InitialUI{
		stats: &ImguiStats{},
	}

	// load custom font
	io := imgui.CurrentIO()
	iui.font = io.Fonts().AddFontFromFileTTF(files.GetFontPath("uifont.ttf"))
	if !iui.font.IsLoaded() {
		return nil, fmt.Errorf("failed to load imgui font")
	}

	// create debug menu

	return iui, nil
}

func (ui *InitialUI) OnKey(key glfw.Key, action glfw.Action) {
	if action == glfw.Press {
		if key == glfw.KeyF1 {
			ui.debugMenu.Visible = !ui.debugMenu.Visible
		}
	}
}

func (ui *InitialUI) NewFrame() {
	// start new frame
	NewFrame()
	// apply custom font
	imgui.PushFont(ui.font, 16.0)
	// imgui widgets
	ui.debugMenu.Show()
	// detach custom font
	imgui.PopFont()
	// finalize
	FinalizeFrame()
}

func (ui *InitialUI) Draw() {
	drawData := imgui.CurrentDrawData()
	ui.stats.Update(drawData)
	implgl3.RenderDrawData(drawData)
}
