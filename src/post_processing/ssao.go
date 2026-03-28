package postprocessing

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/illoprin/retro-fps-kit-go/src/engine/assetmgr"
	"github.com/illoprin/retro-fps-kit-go/src/render"
	"github.com/illoprin/retro-fps-kit-go/src/window"
)

type SSAOConfig struct {
	Use bool
}

type SSAOPass struct {
	ssao        *render.Framebuffer
	blur        *render.Framebuffer
	ssaoProgram *render.Program
	blurProgram *render.Program
	mesh        *render.Mesh
	resources   []render.Resource
	screenCfg   *window.ScreenConfig
	cfg         *SSAOConfig
}

func NewSSAOPass(
	screenCfg *window.ScreenConfig,
	quad *render.Mesh,
	cfg *SSAOConfig,
) (*SSAOPass, error) {
	p := &SSAOPass{
		cfg:       cfg,
		screenCfg: screenCfg,
		mesh:      quad,
	}

	// init ssao buffer
	ssao, err := render.NewFramebuffer(screenCfg.Width, screenCfg.Height)
	if err != nil {
		return nil, fmt.Errorf("ssao pass - failed to create ssao fbo - %w", err)
	}
	ssao.Bind()
	if err := ssao.NewColorAttachment(render.FormatRGBA8); err != nil {
		ssao.Delete()
		return nil, fmt.Errorf("ssao pass - failed to create ssao fbo - %w", err)
	}
	if !ssao.Check() {
		ssao.Delete()
		return nil, fmt.Errorf("ssao pass - fbo not completed %w", err)
	}
	p.ssao = ssao

	// init blur buffer
	blur, err := render.NewFramebuffer(screenCfg.Width, screenCfg.Height)
	if err != nil {
		return nil, fmt.Errorf("ssao pass - failed to create blur fbo - %w", err)
	}
	p.blur = blur

	// blur.Bind()
	// blur.NewColorAttachment(render.FormatR8)

	// init ssao drawing program
	ssaoProgram, err := render.NewProgram(
		assetmgr.GetShaderPath("quad.vert"),
		assetmgr.GetShaderPath("ssao.frag"),
	)
	if err != nil {
		return nil, fmt.Errorf("ssao pass - failed to load program - %w", err)
	}
	p.ssaoProgram = ssaoProgram

	// init blur program
	blurProgram, err := render.NewProgram(
		assetmgr.GetShaderPath("quad.vert"),
		assetmgr.GetShaderPath("ssao_blur.frag"),
	)
	if err != nil {
		return nil, fmt.Errorf("ssao pass - failed to load program - %w", err)
	}
	p.blurProgram = blurProgram

	p.resources = append(p.resources, ssao, blur, ssaoProgram, blurProgram)
	return p, nil
}

func (p *SSAOPass) GetName() string {
	return "ssao"
}

func (p *SSAOPass) ResizeCallback() {
	// resize attachments
	p.blur.Resize(p.screenCfg.Width, p.screenCfg.Height)
	p.ssao.Resize(p.screenCfg.Width, p.screenCfg.Height)
}

func (p *SSAOPass) GetColor() *render.Texture {
	return p.ssao.ColorTextures[0]
}

// RenderPass
// 0 - color
// 1 - normal
// 2 - depth
func (p *SSAOPass) RenderPass(src []*render.Texture) {
	if len(src) < 3 {
		return
	}

	p.ssao.Bind()
	p.ssaoProgram.Use()
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
	src[0].Bind(0)
	p.ssaoProgram.Set1i("u_color", 0)
	p.mesh.Draw()

	// draw ssao
	// render quad
	// do blur
	// render quad
}

func (p *SSAOPass) Use() bool {
	return p.cfg.Use
}

func (p *SSAOPass) GetConfig() interface{} {
	return p.cfg
}

func (p *SSAOPass) Delete() {
	for _, r := range p.resources {
		if r != nil {
			r.Delete()
		}
	}
}
