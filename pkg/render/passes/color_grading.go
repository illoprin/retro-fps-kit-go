package passes

import (
	"fmt"

	"github.com/illoprin/retro-fps-kit-go/pkg/core/files"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/window"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/context"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/pipeline"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
)

type ColorGradingConfig struct {
	Gamma          float32
	Exposure       float32
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
	screenCfg *window.ScreenConfig
}

func NewColorGradingPass(
	screenCfg *window.ScreenConfig,
	quad *rhi.Mesh,
	cfg *ColorGradingConfig,
) (*ColorGradingPass, error) {

	p := &ColorGradingPass{
		screenCfg: screenCfg,
		resources: make([]rhi.Resource, 0),
		mesh:      quad,
		cfg:       cfg,
	}

	fbWidth, fbHeight := p.screenCfg.GetScreenSize()

	// create color framebuffer
	fbo, err := rhi.NewFramebuffer(fbWidth, fbHeight)
	if err != nil {
		return nil, fmt.Errorf("color grading pass - failed to create framebuffer %w", err)
	}
	fbo.Bind()
	if err := fbo.NewColorAttachment(rhi.FormatRGBA8); err != nil {
		fbo.Delete()
		return nil, fmt.Errorf("color grading pass - failed to create color attachment %w", err)
	}
	if !fbo.Check() {
		fbo.Delete()
		return nil, fmt.Errorf("color grading pass - fbo not completed %w", err)
	}
	fbo.Unbind()
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

func (p *ColorGradingPass) GetName() string {
	return "color_grading"
}

// get result color
func (p *ColorGradingPass) GetColor() *rhi.Texture {
	return p.fbo.GetColorTexture(0)
}

// returns result fbo
func (p *ColorGradingPass) GetResultFramebuffer() *rhi.Framebuffer {
	return p.fbo
}

// RenderPass texture index 0 - color
func (p *ColorGradingPass) RenderPass(src *pipeline.DeferredRenderResult) {
	p.fbo.Bind()
	p.program.Use()
	context.ClearColorBuffer()
	src.Color.BindToSlot(0)
	p.program.Set1i("u_color", 0)
	p.program.Set1f("u_brightness", p.cfg.Brightness)
	p.program.Set1f("u_saturation", p.cfg.Saturation)
	p.program.Set1f("u_contrast", p.cfg.Contrast)
	p.program.Set3f("u_shadow_color", p.cfg.ShadowsColor)
	p.program.Set3f("u_mid_color", p.cfg.MidColor)
	p.program.Set3f("u_highlight_color", p.cfg.HighlightColor)
	p.program.Set1f("u_color_strength", p.cfg.ColorStrength)
	p.program.Set1f("u_gamma", p.cfg.Gamma)
	p.program.Set1f("u_exposure", p.cfg.Exposure)
	p.mesh.Draw()
}

func (p *ColorGradingPass) ResizeCallback() {
	fbWidth, fbHeight := p.screenCfg.GetScreenSize()
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
