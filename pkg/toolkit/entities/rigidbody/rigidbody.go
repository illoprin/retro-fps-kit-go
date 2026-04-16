package rigidbody

import (
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/math"
)

const (
	friction = float32(0.95)
)

type Rigidbody struct {
	Velocity mgl.Vec3
}

func (r *Rigidbody) Update(dt float32) {
	r.Velocity = r.Velocity.Mul(friction)
	if r.Velocity.Len() < math.Epsilon {
		r.Velocity = mgl.Vec3{}
	}
}
