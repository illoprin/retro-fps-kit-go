package player

import (
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/camera"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/window"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/entities/rigidbody"
)

const (
	speed            = float32(0.53) // m/s
	sprintMultiplier = float32(2.0)
	bobSpeed         = 200.0
	bobYAmount       = 0.12
	bobXAmount       = 0.1
	strafeRollAmount = 2.0  // deg
	rollSpeed        = 12.5 // deg/s
)

type FPSController struct {
	i *window.InputManager

	camera    *camera.Camera3D
	rigidbody *rigidbody.Rigidbody
	pos       mgl.Vec3
	yaw       float32
	front     mgl.Vec3
	right     mgl.Vec3

	sprint bool
	sens   float32

	bobTime float32
	roll    float32

	height float32
}

func NewFPSController(
	i *window.InputManager,
	pos mgl.Vec3,
	yaw float32,
	sensitivity float32,
) *FPSController {
	c := &FPSController{
		i:         i,
		pos:       pos,
		yaw:       yaw,
		rigidbody: &rigidbody.Rigidbody{},
		sens:      sensitivity,
		height:    1.85, // in meters
	}

	c.camera = camera.NewCamera(
		pos[0],
		c.getEyesHeight(pos[1]),
		pos[2],
	)
	c.camera.SetRotation(0, yaw)
	c.updateVectors()

	return c
}

func (c *FPSController) ProcessInput(deltaTime float32) {
	c.updateVectors()
	c.processKeyboard(deltaTime)
	c.processMouse(c.i.GetMouseDelta())
}

func (c *FPSController) Update(deltaTime float32) {
	c.rigidbody.Update(deltaTime)
	c.pos = c.pos.Add(c.rigidbody.Velocity)
	c.updateCamera(deltaTime)
}

func (c *FPSController) updateCamera(dt float32) {
	vel := c.rigidbody.Velocity
	velLen := vel.Len()

	// camera bob
	c.bobTime += velLen * dt * bobSpeed
	bobY := float32(math.Sin(float64(c.bobTime))) * bobYAmount
	bobX := float32(math.Cos(float64(c.bobTime)*0.5)) * bobXAmount

	// strafe tilt
	var strafe float32 = 0
	var targetRoll float32 = 0
	if velLen > 0.012 {
		strafe = vel.Normalize().Dot(c.right)
	}
	targetRoll = -strafe * strafeRollAmount
	// apply roll
	c.roll += (targetRoll - c.roll) * rollSpeed * dt
	// clamp roll
	if math.Abs(float64(c.roll)) < 0.01 {
		c.roll = 0
	}

	// apply camera position
	y := c.getEyesHeight(c.pos[1]) + bobY
	c.camera.Position = mgl.Vec3{
		c.pos[0] + bobX,
		y,
		c.pos[2] + bobX,
	}

	// apply camera rotation
	c.camera.SetRoll(c.roll)

	// update camera matrices
	c.camera.Update()
}

func (c *FPSController) processKeyboard(dt float32) {

	target := mgl.Vec3{}
	s := speed * dt

	// sprint
	c.sprint = false
	if c.i.IsKeyPressed(glfw.KeyLeftControl) {
		c.sprint = true
	}

	if c.sprint {
		s *= sprintMultiplier
	}

	// movement
	if c.i.IsKeyPressed(glfw.KeyW) {
		target = target.Add(c.front.Mul(s))
	}
	if c.i.IsKeyPressed(glfw.KeyS) {
		target = target.Add(c.front.Mul(-s))
	}
	// strafe
	if c.i.IsKeyPressed(glfw.KeyA) {
		target = target.Add(c.right.Mul(-s))
	}
	if c.i.IsKeyPressed(glfw.KeyD) {
		target = target.Add(c.right.Mul(s))
	}

	// update velocity
	c.rigidbody.Velocity = c.rigidbody.Velocity.Add(target)
}

func (c *FPSController) getEyesHeight(y float32) float32 {
	return y + c.height*.8
}

func (c *FPSController) updateVectors() {
	x := math.Cos(float64(mgl.DegToRad(c.yaw)))
	z := math.Sin(float64(mgl.DegToRad(c.yaw)))
	c.front = mgl.Vec3{float32(x), 0, float32(z)}.Normalize()
	c.right = c.front.Cross(mgl.Vec3{0, 1, 0}).Normalize()
}

func (c *FPSController) processMouse(dx, dy float64) {
	c.yaw += float32(dx) * c.sens
	c.camera.Rotate(float32(-dy)*c.sens, float32(dx)*c.sens)
}

func (c *FPSController) GetCamera() *camera.Camera3D {
	return c.camera
}
