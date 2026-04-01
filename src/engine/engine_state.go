package engine

import "github.com/go-gl/glfw/v3.3/glfw"

type EngineState interface {
	Init(e *Engine) error
	Update(deltaTime float32)
	Render()
	OnKey(key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey)
	OnMouseButton(button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey)
	OnMouseMove(dX, dY, posX, posY float64)
	OnMouseScroll(dx, dy float64)
	Destroy()
}
