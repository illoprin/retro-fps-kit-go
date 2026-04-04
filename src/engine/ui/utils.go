package ui

import (
	"fmt"
	"unsafe"

	implglfw "github.com/AllenDang/cimgui-go/impl/glfw"
	implgl3 "github.com/AllenDang/cimgui-go/impl/opengl3"
	"github.com/illoprin/retro-fps-kit-go/src/core/window"

	"github.com/AllenDang/cimgui-go/imgui"
)

// ImguiStats vertices, draw calls, triangles stats
type ImguiStats struct {
	LastVertices  int32
	LastTriangles int32
	LastDrawCalls int32
}

func (s *ImguiStats) Update(data *imgui.DrawData) {
	s.LastVertices = data.TotalIdxCount()
	s.LastTriangles = data.TotalIdxCount() / 3
	s.LastDrawCalls = data.OwnerViewport().DrawData().CmdListsCount()
}

func InitImgui() {
	// init imgui
	imgui.CreateContext()
	io := imgui.CurrentIO()
	io.SetConfigFlags(
		io.ConfigFlags() |
			imgui.ConfigFlagsNavEnableKeyboard |
			imgui.ConfigFlagsDockingEnable,
	)
	// dark theme
	imgui.StyleColorsDark()
}

func InitImguiRenderer(win *window.Window) error {
	// setup imgui renderer
	// a crutch to get the C pointer to GLFWWindow
	type glfwWindow struct {
		data unsafe.Pointer
	}
	ptr := (*glfwWindow)(unsafe.Pointer(win))

	// init imgui window
	imguiWinGLFW := implglfw.NewGLFWwindowFromC(ptr.data)
	if imguiWinGLFW == nil {
		return fmt.Errorf("could not attach imgui to glfw window")
	}

	if !implglfw.InitForOpenGL(imguiWinGLFW, true) {
		return fmt.Errorf("could not init imgui renderer")
	}

	// init imgui gl renderer
	implgl3.InitV("#version 430 core")

	return nil
}

func NewFrame() {
	// begin imgui frame
	implgl3.NewFrame()
	implglfw.NewFrame()
	imgui.NewFrame()
}

func FinalizeFrame() {
	// finalize imgui frame
	imgui.Render()
}

// ImageTexture represents a texture for display in ImGui
type ImageTexture struct {
	ID   uint32
	Name string
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

func DestroyImgui() {
	implgl3.Shutdown()
	implglfw.Shutdown()
	imgui.DestroyContext()
}
