package player

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-kit-go/src/window"
)

type EditorController struct {
	camera *Camera

	i *window.InputManager

	// Movement speed
	Speed float32

	// Mouse sensitivity
	Sensitivity float32
}

func NewEditorController(i *window.InputManager, pos mgl32.Vec3, speed, sensitivity float32) *EditorController {
	c := NewCamera(pos[0], pos[1], pos[2])

	return &EditorController{
		c,
		i,
		speed,
		sensitivity,
	}
}

func (c *EditorController) Update(dt float32) {
	c.processKeyboard(dt)
	c.processMouse(c.i.GetMouseDelta())
	yOffset, _ := c.i.GetMouseScroll()
	c.processMouseScroll(yOffset)
	c.camera.Update()
}

// ProcessKeyboard handles keyboard input for camera movement
func (c *EditorController) processKeyboard(deltaTime float32) {
	speed := c.Speed * deltaTime

	if c.i.IsKeyPressed(glfw.KeyW) {
		c.camera.AddPosition(c.camera.GetFront().Mul(speed))
	}
	if c.i.IsKeyPressed(glfw.KeyS) {
		c.camera.AddPosition(c.camera.GetFront().Mul(-speed))
	}
	if c.i.IsKeyPressed(glfw.KeyA) {
		c.camera.AddPosition(c.camera.GetRight().Mul(-speed))
	}
	if c.i.IsKeyPressed(glfw.KeyD) {
		c.camera.AddPosition(c.camera.GetRight().Mul(speed))
	}
	if c.i.IsKeyPressed(glfw.KeySpace) {
		c.camera.AddPosition(mgl32.Vec3{0, 1, 0}.Mul(speed))
	}
	if c.i.IsKeyPressed(glfw.KeyLeftControl) {
		c.camera.AddPosition(mgl32.Vec3{0, 1, 0}.Mul(-speed))
	}
}

// processMouse handles mouse movement for camera rotation
func (c *EditorController) processMouse(dx, dy float64) {
	c.camera.Rotate(float32(-dy)*c.Sensitivity, float32(dx)*c.Sensitivity)
}

// processMouseScroll handles mouse wheel for zoom (FOV adjustment)
func (c *EditorController) processMouseScroll(yoffset float64) {
	c.camera.Fov -= float32(yoffset)
	if c.camera.Fov < 1.0 {
		c.camera.Fov = 1.0
	}
	if c.camera.Fov > 100.0 {
		c.camera.Fov = 100.0
	}
}

// GetCamera returns camera object
func (c *EditorController) GetCamera() *Camera {
	return c.camera
}
