package postprocessing

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/illoprin/retro-fps-kit-go/src/engine/assetmgr"
	"github.com/illoprin/retro-fps-kit-go/src/render"
	"github.com/illoprin/retro-fps-kit-go/src/window"
)

type VignetteConfig struct {
	Radius, Softness float32
	Use              bool
}

type VignettePass struct {
	cfg       *VignetteConfig
	program   *render.Program
	fbo       *render.Framebuffer
	mesh      *render.Mesh
	resources []render.Resource
	screenCfg *window.ScreenConfig
}

func NewVignettePass(
	screenCfg *window.ScreenConfig,
	quad *render.Mesh,
	cfg *VignetteConfig,
) (*VignettePass, error) {

	p := &VignettePass{
		screenCfg: screenCfg,
		resources: make([]render.Resource, 0),
		mesh:      quad,
		cfg:       cfg,
	}

	// create color framebuffer
	fbo, err := render.NewFramebuffer(screenCfg.Width, screenCfg.Height)
	if err != nil {
		return nil, fmt.Errorf("vignette pass - failed to create framebuffer %w", err)
	}
	fbo.Bind()
	if err := fbo.NewColorAttachment(render.FormatRGBA8); err != nil {
		fbo.Delete()
		return nil, fmt.Errorf("vignette pass - failed to create color attachment %w", err)
	}
	if !fbo.Check() {
		fbo.Delete()
		return nil, fmt.Errorf("vignette pass - fbo not completed %w", err)
	}
	fbo.Unbind()
	p.fbo = fbo

	// create program
	program, err := render.NewProgram(
		assetmgr.GetShaderPath("quad.vert"),
		assetmgr.GetShaderPath("vignette.frag"),
	)
	if err != nil {
		fbo.Delete()
		return nil, err
	}
	p.program = program
	p.resources = append(p.resources, program, fbo)

	return p, nil
}

func (p *VignettePass) GetName() string {
	return "vignette"
}

// get framebuffer color
func (p *VignettePass) GetColor() *render.Texture {
	return p.fbo.ColorTextures[0]
}

// RenderPass texture index 0 is color
func (p *VignettePass) RenderPass(src []*render.Texture) {
	if len(src) == 0 {
		return
	}
	p.fbo.Bind()
	p.program.Use()
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
	src[0].Bind(0)
	p.program.Set1i("u_color", 0)
	p.program.Set1f("u_radius", p.cfg.Radius)
	p.program.Set1f("u_softness", p.cfg.Softness)
	p.mesh.Draw()
}

func (p *VignettePass) ResizeCallback() {
	p.fbo.Resize(p.screenCfg.Width, p.screenCfg.Height)
}

func (p *VignettePass) Use() bool {
	return p.cfg.Use
}

func (p *VignettePass) GetConfig() interface{} {
	return p.cfg
}

func (p *VignettePass) Delete() {
	for _, r := range p.resources {
		if r != nil {
			r.Delete()
		}
	}
}
