package postprocessing

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"

	"github.com/go-gl/gl/v4.1-core/gl"
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-kit-go/src/engine/assetmgr"
	"github.com/illoprin/retro-fps-kit-go/src/render"
	"github.com/illoprin/retro-fps-kit-go/src/window"
)

const (
	distributionCoefficient = 4.0
	kernelSize              = 256
)

type CreaseOcclusionConfig struct {
	Use        bool
	Radius     float32
	DepthBias  float32
	Intensity  float32
	BlurSize   int32
	KernelSize int32 // max - 256
	BlackPoint float32
	WhitePoint float32
}

type CreaseOcclusionPass struct {
	crease            *render.Framebuffer
	blur              *render.Framebuffer
	composition       *render.Framebuffer
	creaseProgram     *render.Program
	blurProgram       *render.Program
	compositorProgram *render.Program
	mesh              *render.Mesh
	samples           []mgl.Vec2
	projection        mgl.Mat4
	noise             *render.Texture
	resources         []render.Resource
	screenCfg         *window.ScreenConfig
	cfg               *CreaseOcclusionConfig
}

func NewCreaseOcclusionPass(
	screenCfg *window.ScreenConfig,
	quad *render.Mesh,
	cfg *CreaseOcclusionConfig,
	noise *render.Texture,
	blurProgram *render.Program,
	compositorProgram *render.Program,
) (*CreaseOcclusionPass, error) {
	p := &CreaseOcclusionPass{
		cfg:               cfg,
		screenCfg:         screenCfg,
		mesh:              quad,
		noise:             noise,
		blurProgram:       blurProgram,
		compositorProgram: compositorProgram,
	}

	if err := p.initFramebuffers(); err != nil {
		return nil, err
	}

	if err := p.initPrograms(); err != nil {
		return nil, err
	}

	p.createAndSendSamples()

	p.resources = append(p.resources, p.crease, p.blur, p.composition, p.creaseProgram)
	return p, nil
}

func (p *CreaseOcclusionPass) GetName() string {
	return "crease"
}

func (p *CreaseOcclusionPass) ResizeCallback() {

	fbWidth, fbHeight := p.screenCfg.GetScreenSize()
	// resize attachments
	p.crease.Resize(fbWidth, fbHeight)
	p.composition.Resize(fbWidth, fbHeight)
}

func (p *CreaseOcclusionPass) initFramebuffers() error {

	fbWidth, fbHeight := p.screenCfg.GetScreenSize()
	var err error

	// init crease buffer
	crease, err := render.NewFramebuffer(fbWidth, fbHeight)
	if err != nil {
		return fmt.Errorf("crease pass - failed to create crease fbo - %w", err)
	}
	crease.Bind()
	err = crease.NewColorAttachment(render.FormatR8)
	if !crease.Check() || err != nil {
		crease.Delete()
		return fmt.Errorf("crease pass - crease fbo not completed %w", err)
	}
	p.crease = crease

	// init blur buffer
	blur, err := render.NewFramebuffer(fbWidth, fbHeight)
	if err != nil {
		return fmt.Errorf("crease pass - failed to create blur fbo - %w", err)
	}
	blur.Bind()
	err = blur.NewColorAttachment(render.FormatR8)
	if !blur.Check() || err != nil {
		blur.Delete()
		return fmt.Errorf("crease pass - blur fbo not completed %w", err)
	}
	p.blur = blur

	// init composition buffer
	composition, err := render.NewFramebuffer(fbWidth, fbHeight)
	if err != nil {
		return fmt.Errorf("crease pass - failed to create composition fbo - %w", err)
	}
	composition.Bind()
	err = composition.NewColorAttachment(render.FormatRGBA8)
	if !composition.Check() || err != nil {
		composition.Delete()
		return fmt.Errorf("crease pass - composition fbo not completed %w", err)
	}
	p.composition = composition
	return nil
}

func (p *CreaseOcclusionPass) initPrograms() error {
	// init crease drawing program
	creaseProgram, err := render.NewProgram(
		assetmgr.GetShaderPath("quad.vert"),
		assetmgr.GetShaderPath("crease_occlusion.frag"),
	)
	if err != nil {
		return fmt.Errorf("crease pass - failed to load crease program - %w", err)
	}
	p.creaseProgram = creaseProgram
	return nil
}

func (p *CreaseOcclusionPass) createAndSendSamples() {
	// samples (random points [-1, 1] closer to center)
	p.creaseProgram.Use()
	p.samples = make([]mgl.Vec2, kernelSize)
	for i := 0; i < kernelSize; i++ {
		u := rand.Float64()
		v := rand.Float64()

		a := u * 2 * math.Pi
		r := math.Pow(v, distributionCoefficient)

		sample := mgl.Vec2{
			float32(math.Cos(a) * r),
			float32(math.Sin(a) * r),
		}

		p.samples[i] = sample
		p.creaseProgram.Set2f("u_samples["+strconv.Itoa(i)+"]", sample)
	}
}

func (p *CreaseOcclusionPass) SetProjectionMatrix(m mgl.Mat4) {
	p.projection = m
}

func (p *CreaseOcclusionPass) GetColor() *render.Texture {
	return p.composition.ColorTextures[0]
}

func (p *CreaseOcclusionPass) GetBlur() *render.Texture {
	return p.blur.ColorTextures[0]
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
	// send noise texture (random samples rotations)
	p.noise.Bind(2)
	p.creaseProgram.Set1i("u_noise", 2)

	// send radius
	p.creaseProgram.Set1f("u_radius", p.cfg.Radius)
	// send depth bias
	p.creaseProgram.Set1f("u_depthbias", p.cfg.DepthBias)
	// send intensity
	p.creaseProgram.Set1f("u_intensity", p.cfg.Intensity)
	// send kernel size
	p.creaseProgram.Set1i("u_kernel_size", p.cfg.KernelSize)
	// send noise scale
	noiseScale := mgl.Vec2{
		float32(p.screenCfg.Width) / float32(noiseSize),
		float32(p.screenCfg.Height) / float32(noiseSize),
	}
	p.creaseProgram.Set2f("u_noise_scale", noiseScale)
	// send inv projection
	p.creaseProgram.SetMat4("u_invprojection", p.projection.Inv())

	p.mesh.Draw()

	// -- Blur render pass
	p.blur.Bind()
	p.blurProgram.Use()
	gl.Clear(gl.COLOR_BUFFER_BIT)
	// send uniforms
	//
	// crease texture
	p.crease.ColorTextures[0].Bind(0)
	p.blurProgram.Set1i("u_overlay", 0)
	// blur size
	p.blurProgram.Set1i("u_blur_size", p.cfg.BlurSize)

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
	p.blur.ColorTextures[0].Bind(1)
	p.compositorProgram.Set1i("u_overlay", 1)
	// levels cfg
	p.compositorProgram.Set1f("u_blackpoint", p.cfg.BlackPoint)
	p.compositorProgram.Set1f("u_whitepoint", p.cfg.WhitePoint)

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
