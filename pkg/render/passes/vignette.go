package passes

import (
	"fmt"

	"github.com/illoprin/retro-fps-kit-go/pkg/core/files"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/window"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/context"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/pipeline"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
)

type VignetteConfig struct {
	Radius, Softness float32
	Use              bool
}

type VignettePass struct {
	cfg       *VignetteConfig
	program   *rhi.Program
	fbo       *rhi.Framebuffer
	mesh      *rhi.Mesh
	resources []rhi.Resource
	screenCfg *window.ScreenConfig
}

func NewVignettePass(
	screenCfg *window.ScreenConfig,
	quad *rhi.Mesh,
	cfg *VignetteConfig,
) (*VignettePass, error) {

	p := &VignettePass{
		screenCfg: screenCfg,
		resources: make([]rhi.Resource, 0),
		mesh:      quad,
		cfg:       cfg,
	}

	fbWidth, fbHeight := screenCfg.GetScreenSize()

	// create color framebuffer
	fbo, err := rhi.NewFramebuffer(fbWidth, fbHeight)
	if err != nil {
		return nil, fmt.Errorf("vignette pass - failed to create framebuffer %w", err)
	}
	fbo.Bind()
	if err := fbo.NewColorAttachment(rhi.FormatRGBA8); err != nil {
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
	program, err := rhi.NewProgram(
		files.GetShaderPath("screen.vert"),
		files.GetShaderPath("vignette.frag"),
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

// returns result  color
func (p *VignettePass) GetColor() *rhi.Texture {
	return p.fbo.GetColorTexture(0)
}

// returns result framebuffer
func (p *VignettePass) GetResultFramebuffer() *rhi.Framebuffer {
	return p.fbo
}

// RenderPass texture index 0 is color
func (p *VignettePass) RenderPass(src *pipeline.DeferredRenderResult) {
	p.fbo.Bind()
	p.program.Use()
	context.ClearColorBuffer()
	src.Color.BindToSlot(0)
	p.program.Set1i("u_color", 0)
	p.program.Set1f("u_radius", p.cfg.Radius)
	p.program.Set1f("u_softness", p.cfg.Softness)
	p.mesh.Draw()
}

func (p *VignettePass) ResizeCallback() {
	fbWidth, fbHeight := p.screenCfg.GetScreenSize()
	p.fbo.Resize(fbWidth, fbHeight)
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
