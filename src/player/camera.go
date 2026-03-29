package player

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-kit-go/src/engine/global"
)

// Camera represents a first-person style camera
type Camera struct {
	// Position and orientation
	Position mgl32.Vec3
	front    mgl32.Vec3
	up       mgl32.Vec3
	right    mgl32.Vec3
	worldUp  mgl32.Vec3

	// Euler angles
	yaw   float32
	pitch float32

	// Field of view
	Fov float32
}

// NewCamera creates a new camera with default settings
func NewCamera(posX, posY, posZ float32) *Camera {
	cam := &Camera{
		Position: mgl32.Vec3{posX, posY, posZ},
		worldUp:  mgl32.Vec3{0, 1, 0},
		yaw:      -90, // Looking along -Z axis
		pitch:    0.0,
		Fov:      70.0,
	}

	cam.Update()
	return cam
}

// Update recalculates front, right, and up vectors based on yaw and pitch
func (c *Camera) Update() {
	// Calculate front vector

	frontX := float32(math.Cos(float64(mgl32.DegToRad(c.yaw))) * math.Cos(float64(mgl32.DegToRad(c.pitch))))
	frontY := float32(math.Sin(float64(mgl32.DegToRad(c.pitch))))
	frontZ := float32(math.Sin(float64(mgl32.DegToRad(c.yaw))) * math.Cos(float64(mgl32.DegToRad(c.pitch))))

	c.front = mgl32.Vec3{frontX, frontY, frontZ}.Normalize()

	// Calculate right and up vectors
	c.right = c.front.Cross(c.worldUp).Normalize()
	c.up = c.right.Cross(c.front).Normalize()
}

// GetViewMatrix returns the view matrix for the camera
func (c *Camera) GetViewMatrix() mgl32.Mat4 {
	return mgl32.LookAtV(c.Position, c.Position.Add(c.front), c.up)
}

// GetProjectionMatrix returns the perspective projection matrix
func (c *Camera) GetProjectionMatrix(width, height int) mgl32.Mat4 {
	aspect := float32(width) / float32(height)
	return mgl32.Perspective(mgl32.DegToRad(c.Fov), aspect, global.CamNear, global.CamFar)
}

// Getters
func (c *Camera) GetRotation() (float32, float32) {
	return c.pitch, c.yaw
}

func (c *Camera) GetFront() mgl32.Vec3 {
	return c.front
}

func (c *Camera) GetRight() mgl32.Vec3 {
	return c.right
}

func (c *Camera) AddPosition(p mgl32.Vec3) {
	c.Position = c.Position.Add(p)
}

func (c *Camera) SetRotation(pitch, yaw float32) {
	c.pitch = pitch
	c.yaw = yaw
}

func (c *Camera) Rotate(p, y float32) {
	c.pitch = mgl32.Clamp(c.pitch+p, -89.0, 89.0)
	c.yaw += y
}
