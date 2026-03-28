package postprocessing

import "github.com/illoprin/retro-fps-kit-go/src/render"

type VignetteConfig struct {
	Radius, Smooth float32
}

type VignettePass struct {
	cfg       *VignetteConfig
	program   *render.Program
	fbo       *render.Framebuffer
	resources []render.Resource
	screenCfg *ScreenConfig
}

func NewVignettePass(
	screenCfg *ScreenConfig,
	quad *render.Mesh,
	cfg *VignetteConfig,
) (*VignettePass, error) {
	// create color framebuffer
	return nil, nil
}

// get framebuffer color
func (p *VignettePass) GetColor() *render.Texture {
	return nil
}

// RenderPass texture index 0 is color
func (p *VignettePass) RenderPass(src []*render.Texture) {
	if len(src) < 1 {
		return
	}
}

func (p *VignettePass) ResizeCallback() {
}

func (p *VignettePass) Delete() {
	for _, r := range p.resources {
		if r != nil {
			r.Delete()
		}
	}
}
