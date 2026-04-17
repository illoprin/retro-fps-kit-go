package camera

import (
	"math"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core"
)

var (
	worldUp = mgl.Vec3{0, 1, 0}
)

// Camera3D represents a first-person style camera
type Camera3D struct {
	// Position and orientation
	Position mgl.Vec3
	front    mgl.Vec3
	up       mgl.Vec3
	right    mgl.Vec3

	// Euler angles
	yaw   float32
	pitch float32
	roll  float32

	Projection mgl.Mat4
	View       mgl.Mat4

	// Field of view
	Fov float32
}

// NewCamera creates a new camera with default settings
func NewCamera(posX, posY, posZ float32) *Camera3D {
	cam := &Camera3D{
		Position: mgl.Vec3{posX, posY, posZ},
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
	right := c.front.Cross(worldUp).Normalize()
	up := c.right.Cross(c.front).Normalize()

	// Apply roll if needed
	if c.roll != 0 {
		// roll to radians
		rollRad := mgl.DegToRad(c.roll)

		// rotatation matrix around front
		rot := mgl.HomogRotate3D(rollRad, c.front)

		// rotate right and up
		right4 := rot.Mul4x1(right.Vec4(0))
		up4 := rot.Mul4x1(up.Vec4(0))

		c.right = right4.Vec3().Normalize()
		c.up = up4.Vec3().Normalize()
	} else {
		c.right = right
		c.up = up
	}
}

// GetViewMatrix returns the view matrix for the camera
func (c *Camera3D) GetView() mgl.Mat4 {
	c.View = mgl.LookAtV(c.Position, c.Position.Add(c.front), c.up)
	return c.View
}

// GetProjectionMatrix returns the perspective projection matrix
func (c *Camera3D) GetProjection(width, height int) mgl.Mat4 {
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

// Rotate - set pitch and yaw rotation
func (c *Camera3D) Rotate(p, y float32) {
	c.pitch = mgl.Clamp(c.pitch+p, -89.0, 89.0)
	c.yaw += y
}

// SetRoll - sets roll rotation
func (c *Camera3D) SetRoll(r float32) {
	c.roll = r
}
