package renderers

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/obj-scene-editor-go/src/assetmgr"
	"github.com/illoprin/obj-scene-editor-go/src/math"
	"github.com/illoprin/obj-scene-editor-go/src/player"
	"github.com/illoprin/obj-scene-editor-go/src/render"
	"github.com/illoprin/obj-scene-editor-go/src/scene"
)

type PrefabRenderer struct {
	p *render.Program
}

func NewPrefabRenderer() (*PrefabRenderer, error) {
	// init shader program
	program, err := render.NewProgram(
		assetmgr.GetShaderPath("basic.vert"),
		assetmgr.GetShaderPath("basic.frag"),
	)
	if err != nil {
		return nil, err
	}
	return &PrefabRenderer{program}, nil
}

func (r *PrefabRenderer) Prepare(w, h int, c *player.Camera) {
	// prepare shader
	r.p.Use()

	// set global uniforms
	r.p.SetMat4("u_projection", c.GetProjectionMatrix(w, h))
	r.p.SetMat4("u_view", c.GetViewMatrix())
	r.p.Set3f("u_lightPos", c.Position)
	r.p.Set3f("u_lightColor", mgl32.Vec3{0.761 / 2, 0.835 / 2, 0.988 / 2})
}

func (r *PrefabRenderer) Render(p *scene.Prefab) {
	if p.Mesh == nil {
		return
	}

	// set uniforms
	r.p.SetMat4(
		"u_model",
		math.GetTransformMatrixWithOrder(
			math.OrderSRT,
			p.Position,
			p.Rotation,
			p.Scaling,
			p.Pivot,
		),
	)
	r.p.Set1i("u_useTexture", 0)
	if p.Textured {
		p.Texture.Bind(0)
		r.p.Set1i("u_useTexture", 1)
		r.p.Set1i("u_texture", 0)
	}
	r.p.Set3f("u_color", p.Color)

	// draw
	p.Mesh.Draw()
}

func (r *PrefabRenderer) Shutdown() {
	r.p.Delete()
}
