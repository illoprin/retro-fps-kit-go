package post

import (
	"fmt"

	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/files"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/window"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/context"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/pipeline"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/rhi"
)

type ChromaticConfig struct {
	Use      bool    `yaml:"use"`
	Radius   float32 `yaml:"radius"`
	Strength float32 `yaml:"strength"`
	Power    float32 `yaml:"power"`
}

type ChromaticPass struct {
	fbo     *rhi.Framebuffer
	program *rhi.Program
	mesh    *rhi.Mesh

	screen *window.ScreenConfig
	cfg    *ChromaticConfig

	resources []rhi.Resource
}

func NewChromaticPass(
	s PassSharedResources,
	cfg *ChromaticConfig,
) (*ChromaticPass, error) {

	p := &ChromaticPass{
		screen:    s.ScreenConfig,
		mesh:      s.MeshQuad,
		cfg:       cfg,
		resources: make([]rhi.Resource, 0),
	}

	W, H := p.screen.GetScreenSize()

	// create fbo
	fbo := rhi.NewFramebuffer(W, H)
	fbo.Bind()
	fbo.NewColorAttachment(rhi.FormatRGB8, rhi.FilterNearest) // LDR
	if !fbo.Check() {
		fbo.Delete()
		return nil, fmt.Errorf("fbo not completed")
	}
	p.fbo = fbo

	// load program
	program, err := rhi.NewProgram(
		files.GetShaderPath("initial/screen.vert"),
		files.GetShaderPath("initial/chromatic_abberation.frag"),
	)
	if err != nil {
		p.fbo.Delete()
		return nil, fmt.Errorf("failed to load program - %w", err)
	}
	p.program = program

	p.resources = append(p.resources, p.fbo, p.program)

	return p, nil
}

func (p *ChromaticPass) ResizeCallback() {
	W, H := p.screen.GetScreenSize()
	p.fbo.Resize(W, H)
}

func (p *ChromaticPass) RenderPass(src *pipeline.DeferredRenderResult) {

	p.fbo.BindForDrawing()
	context.ClearColorBuffer()

	p.program.Use()

	// color
	src.Color.BindToUnit(0)
	p.program.Set1i("u_color", 0)

	// params
	p.program.Set1f("u_strength", p.cfg.Strength)
	p.program.Set1f("u_radius", p.cfg.Radius)
	p.program.Set1f("u_power", p.cfg.Power)

	p.mesh.Draw()
}

func (p *ChromaticPass) GetColor() *rhi.Texture {
	return p.fbo.GetColorTexture(0)
}

func (p *ChromaticPass) GetResultFramebuffer() *rhi.Framebuffer {
	return p.fbo
}

func (p *ChromaticPass) GetConfig() interface{} {
	return p.cfg
}

func (p *ChromaticPass) Use() bool {
	return p.cfg.Use
}

func (p *ChromaticPass) Delete() {
	for _, r := range p.resources {
		r.Delete()
	}
}
