package prefabsystem

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/camera"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/files"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/math"
	"github.com/illoprin/retro-fps-kit-go/pkg/kit/entities/prefab"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
)

type PrefabRenderer struct {
	p *rhi.Program
}

func NewPrefabRenderer() (*PrefabRenderer, error) {
	// init shader program
	program, err := rhi.NewProgram(
		files.GetShaderPath("d_basic.vert"),
		files.GetShaderPath("d_basic.frag"),
	)
	if err != nil {
		return nil, err
	}
	return &PrefabRenderer{program}, nil
}

func (r *PrefabRenderer) Prepare(w, h int, c *camera.Camera3D) {
	// prepare shader
	r.p.Use()

	// set global uniforms
	r.p.SetMat4("u_projection", c.GetProjectionMatrix(w, h))
	r.p.SetMat4("u_view", c.GetViewMatrix())
	r.p.Set3f("u_light_pos", c.Position)
	r.p.Set3f("u_light_color", mgl32.Vec3{0.761 / 2, 0.835 / 2, 0.988 / 2})
}

func (r *PrefabRenderer) Render(p *prefab.Prefab) {
	if p.Mesh == nil {
		return
	}

	// set uniforms
	r.p.SetMat4(
		"u_model",
		math.GetTransformMatrixWithOrder(
			math.OrderRTS,
			p.Position,
			p.Rotation,
			p.Scaling,
			p.Pivot,
		),
	)
	r.p.Set1i("u_useTexture", 0)
	r.p.Set1i("u_useEmissive", 0)
	if p.Textured {
		p.Diffuse.BindToUnit(0)
		r.p.Set1i("u_useTexture", 1)
		r.p.Set1i("u_diffuse", 0)
	}
	if p.Emissive != nil {
		p.Emissive.BindToUnit(1)
		r.p.Set1i("u_emissive", 1)
		r.p.Set1f("u_emissive_strength", p.EmissiveStrength)
		r.p.Set1i("u_useEmissive", 1)
	}
	r.p.Set3f("u_color", p.Color)

	// draw
	p.Mesh.Draw()
}

func (r *PrefabRenderer) Delete() {
	r.p.Delete()
}
