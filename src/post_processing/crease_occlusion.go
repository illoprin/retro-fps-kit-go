package postprocessing

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-kit-go/src/engine/assetmgr"
	"github.com/illoprin/retro-fps-kit-go/src/render"
	"github.com/illoprin/retro-fps-kit-go/src/window"
)

type CreaseOcclusionConfig struct {
	Use       bool
	Radius    float32
	DepthBias float32
	Intensity float32
}

type CreaseOcclusionPass struct {
	crease            *render.Framebuffer
	composition       *render.Framebuffer
	creaseProgram     *render.Program
	compositorProgram *render.Program
	mesh              *render.Mesh
	resources         []render.Resource
	screenCfg         *window.ScreenConfig
	cfg               *CreaseOcclusionConfig
}

func NewCreaseOcclusionPass(
	screenCfg *window.ScreenConfig,
	quad *render.Mesh,
	cfg *CreaseOcclusionConfig,
) (*CreaseOcclusionPass, error) {
	p := &CreaseOcclusionPass{
		cfg:       cfg,
		screenCfg: screenCfg,
		mesh:      quad,
	}

	if err := p.initFramebuffers(); err != nil {
		return nil, err
	}

	if err := p.initPrograms(); err != nil {
		return nil, err
	}

	p.resources = append(p.resources, p.crease, p.composition, p.creaseProgram, p.compositorProgram)
	return p, nil
}

func (p *CreaseOcclusionPass) GetName() string {
	return "crease"
}

func (p *CreaseOcclusionPass) ResizeCallback() {

	realScreenSize := mgl.Vec3{
		float32(p.screenCfg.Width) * p.screenCfg.ResolutionRatio,
		float32(p.screenCfg.Height) * p.screenCfg.ResolutionRatio,
	}

	// resize attachments
	p.crease.Resize(int32(realScreenSize[0]), int32(realScreenSize[1]))
	p.composition.Resize(int32(realScreenSize[0]), int32(realScreenSize[1]))
}

func (p *CreaseOcclusionPass) initFramebuffers() error {

	realScreenSize := mgl.Vec3{
		float32(p.screenCfg.Width) * p.screenCfg.ResolutionRatio,
		float32(p.screenCfg.Height) * p.screenCfg.ResolutionRatio,
	}

	// init crease buffer
	crease, err := render.NewFramebuffer(int32(realScreenSize[0]), int32(realScreenSize[1]))
	if err != nil {
		return fmt.Errorf("crease pass - failed to create ssao fbo - %w", err)
	}
	crease.Bind()
	if err := crease.NewColorAttachment(render.FormatR8); err != nil {
		crease.Delete()
		return fmt.Errorf("crease pass - failed to create ssao fbo - %w", err)
	}
	if !crease.Check() {
		crease.Delete()
		return fmt.Errorf("crease pass - crease fbo not completed %w", err)
	}
	p.crease = crease

	// init composition buffer
	composition, err := render.NewFramebuffer(int32(realScreenSize[0]), int32(realScreenSize[1]))
	if err != nil {
		return fmt.Errorf("ssao pass - failed to create composition fbo - %w", err)
	}
	composition.Bind()
	if err := composition.NewColorAttachment(render.FormatRGBA8); err != nil {
		composition.Delete()
		return fmt.Errorf("ssao pass - failed to create composition fbo - %w", err)
	}
	if !composition.Check() {
		composition.Delete()
		return fmt.Errorf("ssao pass - composition fbo not completed %w", err)
	}
	p.composition = composition
	return nil
}

// func (p *CreaseOcclusionPass) initNoisy() error {
// 	// samples (hemi-sphere random points)
// 	p.samples = make([]mgl.Vec3, p.cfg.KernelSize)
// 	for i := 0; i < int(p.cfg.KernelSize); i++ {

// 		// generate random hemi-sphere sample
// 		sample := mgl.Vec3{
// 			rand.Float32()*2.0 - 1.0,
// 			rand.Float32()*2.0 - 1.0,
// 			rand.Float32(),
// 		}
// 		sample = sample.Normalize()
// 		sample = sample.Mul(rand.Float32())
// 		// distribution adjustment
// 		//
// 		// more samples closer to the center
// 		scale := float32(i) / 64.0
// 		scale = mathutils.Lerp(0.1, 1.0, scale*scale)
// 		sample = sample.Mul(scale)

// 		p.samples[i] = sample
// 	}
// 	return nil
// }

func (p *CreaseOcclusionPass) initPrograms() error {
	// init ssao drawing program
	creaseProgram, err := render.NewProgram(
		assetmgr.GetShaderPath("quad.vert"),
		assetmgr.GetShaderPath("crease_occlusion.frag"),
	)
	if err != nil {
		return fmt.Errorf("crease pass - failed to load crease program - %w", err)
	}
	p.creaseProgram = creaseProgram

	// init compositor program
	compositorProgram, err := render.NewProgram(
		assetmgr.GetShaderPath("quad.vert"),
		assetmgr.GetShaderPath("crease_compositor.frag"),
	)
	if err != nil {
		return fmt.Errorf("crease pass - failed to crease compositor load program - %w", err)
	}
	p.compositorProgram = compositorProgram
	return nil
}

func (p *CreaseOcclusionPass) GetColor() *render.Texture {
	return p.composition.ColorTextures[0]
}

func (p *CreaseOcclusionPass) GetOcclusion() *render.Texture {
	return p.crease.ColorTextures[0]
}

// RenderPass
// 0 - color
// 1 - normal
// 2 - depth
func (p *CreaseOcclusionPass) RenderPass(src []*render.Texture) {
	if len(src) < 3 {
		return
	}

	color := src[0]
	normal := src[1]
	depth := src[2]

	// -- Crease render pass
	p.crease.Bind()
	gl.Clear(gl.COLOR_BUFFER_BIT)
	p.creaseProgram.Use()
	// bind and send normal
	normal.Bind(0)
	p.creaseProgram.Set1i("u_normal", 0)
	// bind and send depth
	depth.Bind(1)
	p.creaseProgram.Set1i("u_depth", 1)
	// send radius
	p.creaseProgram.Set1f("u_radius", p.cfg.Radius)
	// send depth bias
	p.creaseProgram.Set1f("u_depthbias", p.cfg.DepthBias)
	// send intensity
	p.creaseProgram.Set1f("u_intensity", p.cfg.Intensity)

	p.mesh.Draw()

	// -- Compositor render pass

	p.composition.Bind()
	p.compositorProgram.Use()
	gl.Clear(gl.COLOR_BUFFER_BIT)
	// send uniforms
	//
	// color texture
	color.Bind(0)
	p.compositorProgram.Set1i("u_color", 0)
	// occlusion texture
	p.crease.ColorTextures[0].Bind(1)
	p.compositorProgram.Set1i("u_occlusion", 1)

	p.mesh.Draw()
}

func (p *CreaseOcclusionPass) Use() bool {
	return p.cfg.Use
}

func (p *CreaseOcclusionPass) GetConfig() interface{} {
	return p.cfg
}

func (p *CreaseOcclusionPass) Delete() {
	for _, r := range p.resources {
		if r != nil {
			r.Delete()
		}
	}
}
