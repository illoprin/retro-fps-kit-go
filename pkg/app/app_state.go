package app

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/camera"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
)

type AppState interface {
	Init(e AppAPI) error
	// Update - update internal states
	Update(deltaTime float32)

	// OnKey - calls on key callback
	OnKey(key glfw.Key, action glfw.Action, mods glfw.ModifierKey)

	// OnMouseButton - calls on mouse button events
	OnMouseButton(button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey)

	// OnMouseMove - calls on mouse move
	OnMouseMove(x, y, dx, dy float64)

	// OnMouseScroll - calls on mouse scroll
	OnMouseScroll(dx, dy float64)

	// OnResize - calls on window resize
	// w, h - original widow size
	// wr, hr - "ratioed" window size
	OnResize(w, h, wr, hr int32)

	// Destroy - free state resources
	Destroy()
}

// Optional: Deferred Rendering
type GBufferDrawer interface {
	// RenderGBuffer - renders to gBuffer
	RenderGBuffer()
	GetCamera() *camera.Camera3D
}

// Optional: UI, Menu, HUD
type FlatDrawer interface {
	// RenderFlat - renders 2D on top of scene
	// last - framebuffer on which something can be drawn
	RenderFlat(last *rhi.Framebuffer)
}

// Optional: custom render pass (Forward Rendering, Skybox, Transperency)
type IndirectDrawer interface {
	// RenderIndirect - custom internal rendering
	RenderIndirect()
}

type FPSControlProvider interface {
	HasFPSController() bool
}
