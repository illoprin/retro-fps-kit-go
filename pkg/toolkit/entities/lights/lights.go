package lights

import mgl "github.com/go-gl/mathgl/mgl32"

type AmbientLight struct {
	Color     mgl.Vec3
	Intensity float32
}

type PointLight struct {
	Position  mgl.Vec3
	Intensity float32
	Radius    float32
	Color     mgl.Vec3
}

type SpotLight struct {
	Position   mgl.Vec3
	Forward    mgl.Vec3 // normalized
	Radius     float32
	Intensity  float32
	InnerAngle float32 // deg
	OuterAngle float32 // deg
}

func NewPointLight() PointLight {
	return PointLight{
		Position:  mgl.Vec3{0, 0, 0},
		Intensity: 4,
		Radius:    16,
		Color:     mgl.Vec3{.9, .2, .34},
	}
}

func NewSpotLight() SpotLight {
	return SpotLight{
		Forward:    mgl.Vec3{0, 0, 1},
		Radius:     8.3,
		Intensity:  10.2,
		InnerAngle: 26,
		OuterAngle: 60,
	}
}
