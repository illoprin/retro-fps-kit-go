package app

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/camera"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
)

// AppState defines the lifecycle and event
// handling for a specific application logic block.
// Implement this interface to create
// game levels, menus, or specialized tools like editors.
type AppState interface {
	Init(e AppAPI) error

	// Update - update internal states
	Update(deltaTime float32)

	// OnKey - calls on key callback
	OnKey(key glfw.Key, action glfw.Action, mods glfw.ModifierKey)

	// Destroy - free state resources
	Destroy()
}

// Optional: state can handle framebuffer resize
type ResizeHandler interface {
	// OnResize - calls on window resize
	// w, h - original widow size
	// wr, hr - "ratioed" window size
	OnResize(w, h, wr, hr int32)
}

// Optional: state can handle mouse buttons
type MouseButtonsHandler interface {
	// OnMouseButton - calls on mouse button events
	OnMouseButton(button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey)
}

// Optional: state can handle mouse movement
type MouseMoveHandler interface {
	// OnMouseMove - calls on mouse move
	OnMouseMove(x, y, dx, dy float64)
}

// Optional: state can handle mouse scroll
type MouseScrollHandler interface {
	// OnMouseScroll - calls on mouse scroll
	OnMouseScroll(dx, dy float64)
}

// Optional: Deferred Rendering
type GBufferDrawer interface {
	// RenderGBuffer - renders to gBuffer
	RenderGBuffer()
	GetCamera() *camera.Camera3D
}

// Optional: If state uses imgui objects
type UIDrawer interface {
	// RenderImgui - renders imgui objects
	ShowImgui()
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
