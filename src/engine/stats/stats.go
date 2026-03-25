package stats

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/illoprin/retro-fps-kit-go/src/engine/global"
)

func UpdateGlobal() {
	global.LastDrawCalls = global.DrawCalls
	global.LastVertices = global.DrawVertices
	global.DrawVertices = 0
	global.DrawCalls = 0
}

func UpdateForImgui(imguiDrawData *imgui.DrawData) {
	global.LastImguiVertices = imguiDrawData.TotalIdxCount()
	global.LastImguiDrawCalls = imguiDrawData.OwnerViewport().DrawData().CmdListsCount()
}
