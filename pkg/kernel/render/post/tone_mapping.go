package post

import (
	"fmt"

	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/files"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/window"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/context"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/pipeline"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/rhi"
)

const (
	ACESFilmTonemap  = "aces"
	UnchartedTonemap = "uncharted"
	ReinhardTonemap  = "reinhard"
)

var (
	ToneMapEnum = map[string]int{
		ACESFilmTonemap:  1,
		UnchartedTonemap: 2,
		ReinhardTonemap:  3,
	}
)

type ToneMappingConfig struct {
	Gamma   float32 `yaml:"gamma"`
	Tonemap string  `yaml:"tonemap"`
}

type ToneMappingPass struct {
	fbo       *rhi.Framebuffer
	program   *rhi.Program
	mesh      *rhi.Mesh
	screen    *window.ScreenConfig
	cfg       *ToneMappingConfig
	resources []rhi.Resource
}

func NewToneMappingPass(
	s PassSharedResources,
	cfg *ToneMappingConfig,
) (*ToneMappingPass, error) {

	p := &ToneMappingPass{
		screen:    s.ScreenConfig,
		mesh:      s.MeshQuad,
		cfg:       cfg,
		resources: make([]rhi.Resource, 0),
	}

	if err := p.initFramebuffer(); err != nil {
		return nil, err
	}

	if err := p.initProgram(); err != nil {
		return nil, err
	}

	p.resources = append(p.resources, p.fbo, p.program)

	return p, nil
}

func (p *ToneMappingPass) Use() bool {
	return true
}

func (p *ToneMappingPass) GetConfig() interface{} {
	return p.cfg
}

func (p *ToneMappingPass) GetColor() *rhi.Texture {
	return p.fbo.GetColorTexture(0)
}

func (p *ToneMappingPass) GetResultFramebuffer() *rhi.Framebuffer {
	return p.fbo
}

// GetDebugTextures implementing debug interface
func (p *ToneMappingPass) GetDebugTextures() []DebugTexture {
	return []DebugTexture{
		{"tonemapping.color", p.fbo.GetColorTexture(0)},
	}
}

func (p *ToneMappingPass) ResizeCallback() {
	w, h := p.screen.GetScreenSize()
	p.fbo.Resize(w, h)
}

func (p *ToneMappingPass) initFramebuffer() error {
	w, h := p.screen.GetScreenSize()

	fbo := rhi.NewFramebuffer(w, h)
	fbo.Bind()
	fbo.NewColorAttachment(rhi.FormatRGBA8, rhi.FilterNearest)

	if !fbo.Check() {
		fbo.Delete()
		return fmt.Errorf("framebuffer not complete")
	}

	p.fbo = fbo
	return nil
}

func (p *ToneMappingPass) initProgram() error {
	program, err := rhi.NewProgram(
		files.GetShaderPath("screen.vert"),
		files.GetShaderPath("tone_mapping.frag"),
	)
	if err != nil {
		return fmt.Errorf("failed to load shader - %w", err)
	}

	p.program = program
	return nil
}

func (p *ToneMappingPass) RenderPass(src *pipeline.DeferredRenderResult) {

	p.fbo.BindForDrawing()
	context.ClearColorBuffer()

	p.program.Use()

	// HDR color input
	src.Color.BindToUnit(0)
	p.program.Set1i("u_color", 0)

	// uniforms
	p.program.Set1f("u_gamma", p.cfg.Gamma)
	p.program.Set1i("u_tonemap_type", int32(ToneMapEnum[p.cfg.Tonemap]))

	p.mesh.Draw()
}

func (p *ToneMappingPass) Delete() {
	for _, r := range p.resources {
		if r != nil {
			r.Delete()
		}
	}
}
