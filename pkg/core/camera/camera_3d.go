package camera

import (
	"math"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-kit-go/pkg/core"
)

// Camera3D represents a first-person style camera
type Camera3D struct {
	// Position and orientation
	Position mgl.Vec3
	front    mgl.Vec3
	up       mgl.Vec3
	right    mgl.Vec3
	worldUp  mgl.Vec3

	// Euler angles
	yaw   float32
	pitch float32

	Projection mgl.Mat4
	View       mgl.Mat4

	// Field of view
	Fov float32
}

// NewCamera creates a new camera with default settings
func NewCamera(posX, posY, posZ float32) *Camera3D {
	cam := &Camera3D{
		Position: mgl.Vec3{posX, posY, posZ},
		worldUp:  mgl.Vec3{0, 1, 0},
		yaw:      -90, // Looking along -Z axis
		pitch:    0.0,
		Fov:      70.0,
	}
	cam.Update()
	return cam
}

// Update recalculates front, right, and up vectors based on yaw and pitch
func (c *Camera3D) Update() {
	// Calculate front vector

	frontX := float32(math.Cos(float64(mgl.DegToRad(c.yaw))) * math.Cos(float64(mgl.DegToRad(c.pitch))))
	frontY := float32(math.Sin(float64(mgl.DegToRad(c.pitch))))
	frontZ := float32(math.Sin(float64(mgl.DegToRad(c.yaw))) * math.Cos(float64(mgl.DegToRad(c.pitch))))

	c.front = mgl.Vec3{frontX, frontY, frontZ}.Normalize()

	// Calculate right and up vectors
	c.right = c.front.Cross(c.worldUp).Normalize()
	c.up = c.right.Cross(c.front).Normalize()
}

// GetViewMatrix returns the view matrix for the camera
func (c *Camera3D) GetViewMatrix() mgl.Mat4 {
	c.View = mgl.LookAtV(c.Position, c.Position.Add(c.front), c.up)
	return c.View
}

// GetProjectionMatrix returns the perspective projection matrix
func (c *Camera3D) GetProjectionMatrix(width, height int) mgl.Mat4 {
	aspect := float32(width) / float32(height)
	c.Projection = mgl.Perspective(mgl.DegToRad(c.Fov), aspect, core.CamNear, core.CamFar)
	return c.Projection
}

// Getters
func (c *Camera3D) GetRotation() (float32, float32) {
	return c.pitch, c.yaw
}

func (c *Camera3D) GetFront() mgl.Vec3 {
	return c.front
}

func (c *Camera3D) GetRight() mgl.Vec3 {
	return c.right
}

func (c *Camera3D) AddPosition(p mgl.Vec3) {
	c.Position = c.Position.Add(p)
}

func (c *Camera3D) SetRotation(pitch, yaw float32) {
	c.pitch = pitch
	c.yaw = yaw
}

func (c *Camera3D) Rotate(p, y float32) {
	c.pitch = mgl.Clamp(c.pitch+p, -89.0, 89.0)
	c.yaw += y
}
