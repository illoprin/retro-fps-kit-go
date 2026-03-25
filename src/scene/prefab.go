package scene

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-kit-go/src/render"
)

type Prefab struct {
	Position mgl32.Vec3
	Scaling  mgl32.Vec3
	Rotation mgl32.Vec3
	Pivot    mgl32.Vec3
	Color    mgl32.Vec3
	Textured bool
	Texture  *render.Texture
	Mesh     *render.Mesh
}

func NewPrefab(m *render.Mesh, t *render.Texture) *Prefab {
	p := &Prefab{
		Color:    mgl32.Vec3{1, 1, 1},
		Scaling:  mgl32.Vec3{1, 1, 1},
		Textured: t != nil,
		Texture:  t,
		Mesh:     m,
	}

	return p
}

func (p *Prefab) Translate(x, y, z float32) {
	p.Position = p.Position.Add(mgl32.Vec3{x, y, z})
}

func (p *Prefab) Rotate(x, y, z float32) {
	p.Rotation = p.Rotation.Add(mgl32.Vec3{x, y, z})
}

func (p *Prefab) Scale(v float32) {
	p.Scaling = p.Scaling.Mul(v)
}
