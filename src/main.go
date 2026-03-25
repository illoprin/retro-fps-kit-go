package main

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"
	implglfw "github.com/AllenDang/cimgui-go/impl/glfw"
	implgl3 "github.com/AllenDang/cimgui-go/impl/opengl3"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/obj-scene-editor-go/src/assetmgr"
	"github.com/illoprin/obj-scene-editor-go/src/global"
	"github.com/illoprin/obj-scene-editor-go/src/model"
	"github.com/illoprin/obj-scene-editor-go/src/player"
	"github.com/illoprin/obj-scene-editor-go/src/render"
	"github.com/illoprin/obj-scene-editor-go/src/renderers"
	"github.com/illoprin/obj-scene-editor-go/src/scene"
	"github.com/illoprin/obj-scene-editor-go/src/window"
)

var (
	showingDebug              = true
	lastDrawCalls      uint32 = 0
	lastVertices       uint32 = 0
	lastImguiDrawCalls int32  = 0
	lastImguiVertices  int32  = 0
)

func init() {
	// use opengl in single thread
	runtime.LockOSThread()
}

func setupGL(w, h int) {
	gl.Viewport(0, 0, int32(w), int32(h))
	gl.ClearColor(0.176, 0.216, 0.302, 1.0)
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)
	gl.FrontFace(gl.CCW)
}

func showStatsWindow(
	io *imgui.IO, c *player.Camera,
) {
	// Create the window
	if showingDebug {
		r := c.GetRotation()
		cameraPosStr := fmt.Sprintf(
			"X: %.2f\nY: %.2f\nZ: %.2f\nPitch: %.2f\nYaw: %.2f\nFOV: %.2f",
			c.Position[0], c.Position[1], c.Position[2], r[0], r[1], c.Fov,
		)
		imgui.SetNextWindowSizeConstraints(imgui.Vec2{200, 250}, imgui.Vec2{1000, 600})
		imgui.Begin("Debug")
		// tab bar
		if imgui.BeginTabBar("DebugTabBar") {

			// stats bar
			if imgui.BeginTabItem("Stats") {
				sVty := true
				if imgui.CollapsingHeaderBoolPtr("Performance", &sVty) {
					imgui.Text(fmt.Sprintf("FPS: %.2f\nFrame Time: %.3f ms", io.Framerate(), io.DeltaTime()))
				}
				dcVty := true
				if imgui.CollapsingHeaderBoolPtr("Draw Calls", &dcVty) {
					imgui.Text(fmt.Sprintf("Scene: %d\nImGUI: %d", lastDrawCalls, lastImguiDrawCalls))
				}
				vVty := true
				if imgui.CollapsingHeaderBoolPtr("Vertices", &vVty) {
					imgui.Text(fmt.Sprintf("Scene: %d\nImGUI: %d", lastVertices, lastImguiVertices))
				}
				cVty := true
				if imgui.CollapsingHeaderBoolPtr("Camera", &cVty) {
					imgui.Text(cameraPosStr)
				}
				imgui.EndTabItem()
			}

			// scene bar
			if imgui.BeginTabItem("Scene") {
				imgui.Text("nil")
				imgui.EndTabItem()
			}

			// textures bar
			if imgui.BeginTabItem("Textures") {
				imgui.Text("nil")
				imgui.EndTabItem()
			}

			// fbo bar
			if imgui.BeginTabItem("Buffers") {
				imgui.Text("nil")
				imgui.EndTabItem()
			}

			imgui.EndTabBar()
		}

		imgui.End()
	}
}

func updateMetrics() {
	lastDrawCalls = global.DrawCalls
	lastVertices = global.DrawVertices
	global.DrawVertices = 0
	global.DrawCalls = 0
}

func main() {
	if err := window.InitGLFW(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	// create window
	win, err := window.NewWindow(1470, 710, "Game")
	if err != nil {
		panic(err)
	}
	defer win.Destroy()
	win.MakeContextCurrent()
	win.Focus()
	win.Center()

	input := player.NewManager(win.Window)
	controller := player.NewEditorController(
		input, mgl32.Vec3{0, 0, 3}, 1.2, 0.1,
	)

	// init opengl
	if err := gl.Init(); err != nil {
		panic(err)
	}

	// init imgui
	imgui.CreateContext()
	defer imgui.DestroyContext()
	io := imgui.CurrentIO()
	io.SetConfigFlags(io.ConfigFlags() | imgui.ConfigFlagsNavEnableKeyboard | imgui.ConfigFlagsDockingEnable)
	// load custom font
	imguiCustomFont := io.Fonts().AddFontFromFileTTF(filepath.Join("assets", "fonts", "uifont.ttf"))
	if !imguiCustomFont.IsLoaded() {
		log.Fatalf("could not load font")
	}
	imgui.StyleColorsDark()

	// setup imgui renderer
	// a crutch to get the C pointer to GLFWWindow
	type glfwWindow struct {
		data unsafe.Pointer
	}
	ptr := (*glfwWindow)(unsafe.Pointer(win.Window))
	glfwWindowCPtr := ptr.data

	// init imgui window
	imguiWinGLFW := implglfw.NewGLFWwindowFromC(glfwWindowCPtr)
	implglfw.InitForOpenGL(imguiWinGLFW, true)
	defer implglfw.Shutdown()

	// init imgui gl renderer
	implgl3.InitV("#version 430 core")
	defer implgl3.Shutdown()

	setupGL(win.GetSize())

	// init prefab renderer
	prefabRenderer, err := renderers.NewPrefabRenderer()
	if err != nil {
		panic(err)
	}
	defer prefabRenderer.Shutdown()

	// init assets
	// model
	parser := model.NewOBJParser()
	shotgunModel, err := parser.ParseFile(assetmgr.GetModelPath("shotgun.obj"))
	if err != nil {
		log.Printf("failed to load model %v", err)
	}
	// mesh
	meshShotgun := render.NewMesh()
	meshShotgun.SetupFromModel(shotgunModel, gl.STATIC_DRAW)
	// texture
	texColors, err := render.NewTextureFromImage(assetmgr.GetTexturePath("colors.png"), true)
	if err != nil {
		log.Printf("failed to load texture %v", err)
	} else {
		defer texColors.Delete()
	}
	shotgunPrefab := scene.NewPrefab(meshShotgun, texColors)

	for !win.ShouldClose() {
		// process input
		input.Update()
		glfw.PollEvents()

		// update our scene
		if input.IsKeyJustPressed(glfw.KeyEscape) {
			win.SetShouldClose(true)
		}
		if input.IsKeyJustPressed(glfw.KeyG) {
			input.ToggleGameMode()
		}
		if input.IsKeyJustPressed(glfw.KeyF1) {
			showingDebug = !showingDebug
		}
		if input.GetGameMode() || input.IsMouseButtonPressed(glfw.MouseButton1) && !io.WantCaptureMouse() {
			controller.Update(io.DeltaTime())
		}

		// begin imgui frame
		implgl3.NewFrame()
		implglfw.NewFrame()
		imgui.NewFrame()

		// imgui widgets

		// apply custom font
		imgui.PushFont(imguiCustomFont, 15.0)

		showStatsWindow(io, controller.GetCamera())

		imgui.PopFont()

		// finalize imgui frame
		imgui.Render()

		// render scene
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		w, h := win.GetSize()
		prefabRenderer.Prepare(w, h, controller.GetCamera())
		prefabRenderer.Render(shotgunPrefab)

		// render imgui on top of scene
		imguiDrawData := imgui.CurrentDrawData()
		lastImguiVertices = imguiDrawData.TotalIdxCount()
		lastImguiDrawCalls = imguiDrawData.OwnerViewport().DrawData().CmdListsCount()
		implgl3.RenderDrawData(imguiDrawData)

		updateMetrics()

		win.SwapBuffers()
	}
}
