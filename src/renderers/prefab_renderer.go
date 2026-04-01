package renderers

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-kit-go/src/engine/assetmgr"
	"github.com/illoprin/retro-fps-kit-go/src/math"
	"github.com/illoprin/retro-fps-kit-go/src/player"
	"github.com/illoprin/retro-fps-kit-go/src/render"
	"github.com/illoprin/retro-fps-kit-go/src/scene"
)

type PrefabRenderer struct {
	p *render.Program
}

func NewPrefabRenderer() (*PrefabRenderer, error) {
	// init shader program
	program, err := render.NewProgram(
		assetmgr.GetShaderPath("deferred_basic.vert"),
		assetmgr.GetShaderPath("deferred_basic.frag"),
	)
	if err != nil {
		return nil, err
	}
	return &PrefabRenderer{program}, nil
}

func (r *PrefabRenderer) Prepare(w, h int, c *player.Camera3D) {
	// prepare shader
	r.p.Use()

	// set global uniforms
	r.p.SetMat4("u_projection", c.GetProjectionMatrix(w, h))
	r.p.SetMat4("u_view", c.GetViewMatrix())
	r.p.Set3f("u_light_pos", c.Position)
	r.p.Set3f("u_light_color", mgl32.Vec3{0.761 / 2, 0.835 / 2, 0.988 / 2})
}

func (r *PrefabRenderer) Render(p *scene.Prefab) {
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
	if p.Textured {
		p.Texture.BindToSlot(0)
		r.p.Set1i("u_useTexture", 1)
		r.p.Set1i("u_texture", 0)
	}
	r.p.Set3f("u_color", p.Color)

	// draw
	p.Mesh.Draw()
}

func (r *PrefabRenderer) Delete() {
	r.p.Delete()
}
