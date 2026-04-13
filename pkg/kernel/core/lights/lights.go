package lights

import (
	"math"

	mgl "github.com/go-gl/mathgl/mgl32"
)

type AmbientLight struct {
	Color     mgl.Vec3
	Intensity float32
}

// PointLight struct prepared for UBO injection
// Position:
//
//		xyz: world space pos
//	 w: radius
//
// Color:
//
//		rgb: color
//	 w: intensity
//
// Size:
//
//	32 bytes
type PointLight struct {
	Position mgl.Vec4 // w is radius
	Color    mgl.Vec4 // w is intensity
}

// SpotLight struct prepared for UBO injection
// Size
//
//	64 bytes
type SpotLight struct {
	Position mgl.Vec4 // w is radius
	Color    mgl.Vec4 // w is intensity
	Forward  mgl.Vec4 // forward.xyz + cosOuter
	Params   mgl.Vec4 // cosInner + padding
}

func DefaultPointLight() PointLight {
	return PointLight{
		Position: mgl.Vec4{0, 0, 0, 16},
		Color:    mgl.Vec4{.9, .2, .34, 4},
	}
}

func DefaultSpotLight() SpotLight {

	return SpotLight{
		Position: mgl.Vec4{0, 0, 0, 8.3},
		Color:    mgl.Vec4{1, 1, 1, 10.2},
		Forward:  mgl.Vec4{0, 0, 1, float32(math.Cos(float64(mgl.DegToRad(26))))},
		Params:   mgl.Vec4{float32(math.Cos(float64(mgl.DegToRad(60)))), 0, 0, 0},
	}
}
