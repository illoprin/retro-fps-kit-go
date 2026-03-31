package imgui

import "github.com/AllenDang/cimgui-go/imgui"

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
