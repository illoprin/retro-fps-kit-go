package post

import (
	"fmt"

	"github.com/illoprin/retro-fps-kit-go/pkg/core/files"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/window"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/context"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/pipeline"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
)

type ColorGradingConfig struct {
	Contrast       float32
	Saturation     float32
	Brightness     float32
	ShadowsColor   [3]float32
	MidColor       [3]float32
	HighlightColor [3]float32
	ColorStrength  float32
	Use            bool
}

type ColorGradingPass struct {
	cfg       *ColorGradingConfig
	program   *rhi.Program
	mesh      *rhi.Mesh
	fbo       *rhi.Framebuffer
	resources []rhi.Resource
	screen    *window.ScreenConfig
}

func NewColorGradingPass(
	s PassSharedResources,
	cfg *ColorGradingConfig,
) (*ColorGradingPass, error) {

	p := &ColorGradingPass{
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
		files.GetShaderPath("color_grading.frag"),
	)
	if err != nil {
		fbo.Delete()
		return nil, err
	}
	p.program = program
	p.resources = append(p.resources, program, fbo)

	return p, nil
}

// get result color
func (p *ColorGradingPass) GetColor() *rhi.Texture {
	return p.fbo.GetColorTexture(0)
}

// returns result fbo
func (p *ColorGradingPass) GetResultFramebuffer() *rhi.Framebuffer {
	return p.fbo
}

// GetDebugTextures implementing debug interface
func (p *ColorGradingPass) GetDebugTextures() []DebugTexture {
	return []DebugTexture{
		{"color_grading.color", p.fbo.GetColorTexture(0)},
	}
}

// RenderPass texture index 0 - color
func (p *ColorGradingPass) RenderPass(src *pipeline.DeferredRenderResult) {
	p.fbo.BindForDrawing()
	p.program.Use()
	context.ClearColorBuffer()
	src.Color.BindToUnit(0)
	p.program.Set1i("u_color", 0)
	p.program.Set1f("u_brightness", p.cfg.Brightness)
	p.program.Set1f("u_saturation", p.cfg.Saturation)
	p.program.Set1f("u_contrast", p.cfg.Contrast)
	p.program.Set3f("u_shadow_color", p.cfg.ShadowsColor)
	p.program.Set3f("u_mid_color", p.cfg.MidColor)
	p.program.Set3f("u_highlight_color", p.cfg.HighlightColor)
	p.program.Set1f("u_color_strength", p.cfg.ColorStrength)
	p.mesh.Draw()
}

func (p *ColorGradingPass) ResizeCallback() {
	fbWidth, fbHeight := p.screen.GetScreenSize()
	p.fbo.Resize(fbWidth, fbHeight)
}

func (p *ColorGradingPass) GetConfig() interface{} {
	return p.cfg
}

func (p *ColorGradingPass) Use() bool {
	return p.cfg.Use
}

func (p *ColorGradingPass) Delete() {
	for _, r := range p.resources {
		if r != nil {
			r.Delete()
		}
	}
}
