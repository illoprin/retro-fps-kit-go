package prefab

import (
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/rhi"
)

type Prefab struct {
	Position         mgl.Vec3
	Scaling          mgl.Vec3
	Rotation         mgl.Vec3
	Pivot            mgl.Vec3
	Color            mgl.Vec3
	Textured         bool
	Diffuse          *rhi.Texture
	Emissive         *rhi.Texture
	Mesh             *rhi.Mesh
	EmissiveStrength float32
}

func NewPrefab(m *rhi.Mesh, t *rhi.Texture, e *rhi.Texture) *Prefab {
	p := &Prefab{
		Color:            mgl.Vec3{1, 1, 1},
		Scaling:          mgl.Vec3{1, 1, 1},
		Textured:         t != nil,
		Diffuse:          t,
		Emissive:         e,
		Mesh:             m,
		EmissiveStrength: 1.0,
	}

	return p
}

func (p *Prefab) Translate(x, y, z float32) {
	p.Position = p.Position.Add(mgl.Vec3{x, y, z})
}

func (p *Prefab) Rotate(x, y, z float32) {
	p.Rotation = p.Rotation.Add(mgl.Vec3{x, y, z})
}

func (p *Prefab) Scale(v float32) {
	p.Scaling = p.Scaling.Mul(v)
}
