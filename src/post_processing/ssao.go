package postprocessing

import (
	"github.com/illoprin/retro-fps-kit-go/src/engine/assetmgr"
	"github.com/illoprin/retro-fps-kit-go/src/render"
)

type SSAOConfig struct {
}

type SSAOPass struct {
	ssao        *render.Framebuffer
	blur        *render.Framebuffer
	ssaoProgram *render.Program
	blurProgram *render.Program
	mesh        *render.Mesh
	resources   []render.Resource
	screenCfg   *ScreenConfig
	cfg         *SSAOConfig
}

func NewSSAOPass(
	screenCfg *ScreenConfig,
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
		return nil, err
	}
	ssao.Bind()
	ssao.NewColorAttachment(render.FormatR8)

	// init blur buffer
	blur, err := render.NewFramebuffer(screenCfg.Width, screenCfg.Height)
	if err != nil {
		return nil, err
	}
	blur.Bind()
	blur.NewColorAttachment(render.FormatR8)

	// init ssao drawing program
	ssaoProgram, err := render.NewProgram(
		assetmgr.GetShaderPath("ssao.frag"),
		assetmgr.GetShaderPath("screen.vert"),
	)
	if err != nil {
		return nil, err
	}

	// init blur program
	blurProgram, err := render.NewProgram(
		assetmgr.GetShaderPath("ssao_blur.frag"),
		assetmgr.GetShaderPath("screen.vert"),
	)
	if err != nil {
		return nil, err
	}

	p.ssao, p.blur = ssao, blur
	p.resources = append(p.resources, ssao, blur, ssaoProgram, blurProgram)
	return p, nil
}

func (p *SSAOPass) ResizeCallback() {
	// resize attachments
	p.blur.Resize(p.screenCfg.Width, p.screenCfg.Height)
	p.ssao.Resize(p.screenCfg.Width, p.screenCfg.Height)
}

func (p *SSAOPass) GetColor() *render.Texture {
	return p.blur.ColorTextures[0]
}

// RenderPass
// 0 - color
// 1 - normal
// 2 - depth
func (p *SSAOPass) RenderPass(src []*render.Texture) {
	if len(src) < 3 {
		return
	}

	// draw ssao
	// render quad
	// do blur
	// render quad
}

func (p *SSAOPass) Delete() {
	for _, r := range p.resources {
		if r != nil {
			r.Delete()
		}
	}
}
