package post

import (
	"fmt"

	"github.com/illoprin/retro-fps-toolkit-go/pkg/core/files"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/core/window"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/render/context"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/render/pipeline"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/render/rhi"
)

type VignetteConfig struct {
	Radius   float32 `yaml:"radius"`
	Softness float32 `yaml:"softness"`
	Use      bool    `yaml:"use"`
}

type VignettePass struct {
	cfg       *VignetteConfig
	program   *rhi.Program
	fbo       *rhi.Framebuffer
	mesh      *rhi.Mesh
	resources []rhi.Resource
	screen    *window.ScreenConfig
}

func NewVignettePass(
	s PassSharedResources,
	cfg *VignetteConfig,
) (*VignettePass, error) {

	p := &VignettePass{
		screen:    s.ScreenConfig,
		mesh:      s.MeshQuad,
		cfg:       cfg,
		resources: make([]rhi.Resource, 0),
	}

	fbWidth, fbHeight := p.screen.GetScreenSize()

	// create color framebuffer
	fbo := rhi.NewFramebuffer(fbWidth, fbHeight)
	fbo.Bind()
	fbo.NewColorAttachment(rhi.FormatRGB8, rhi.FilterNearest)
	if !fbo.Check() {
		fbo.Delete()
		return nil, fmt.Errorf("fbo not completed")
	}
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

// returns result  color
func (p *VignettePass) GetColor() *rhi.Texture {
	return p.fbo.GetColorTexture(0)
}

// returns result framebuffer
func (p *VignettePass) GetResultFramebuffer() *rhi.Framebuffer {
	return p.fbo
}

// GetDebugTextures implementing debug interface
func (p *VignettePass) GetDebugTextures() []DebugTexture {
	return []DebugTexture{
		{"vignette.color", p.fbo.GetColorTexture(0)},
	}
}

// RenderPass texture index 0 is color
func (p *VignettePass) RenderPass(src *pipeline.DeferredRenderResult) {
	p.fbo.BindForDrawing()
	p.program.Use()
	context.ClearColorBuffer()
	src.Color.BindToUnit(0)
	p.program.Set1i("u_color", 0)
	p.program.Set1f("u_radius", p.cfg.Radius)
	p.program.Set1f("u_softness", p.cfg.Softness)
	p.mesh.Draw()
}

func (p *VignettePass) ResizeCallback() {
	fbWidth, fbHeight := p.screen.GetScreenSize()
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
