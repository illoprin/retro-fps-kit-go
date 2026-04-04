package passes

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/files"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/window"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/context"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/pipeline"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
)

const (
	distributionCoefficient = 4.0
	kernelSize              = 256
)

type CavityConfig struct {
	Use        bool
	Radius     float32
	DepthBias  float32
	Intensity  float32
	BlurSize   int32
	KernelSize int32 // max - 256
	BlackPoint float32
	WhitePoint float32
}

type CavityPass struct {
	cavity            *rhi.Framebuffer
	blur              *rhi.Framebuffer
	composition       *rhi.Framebuffer
	cavityProgram     *rhi.Program
	blurProgram       *rhi.Program
	compositorProgram *rhi.Program
	mesh              *rhi.Mesh
	samples           []mgl.Vec2
	projection        mgl.Mat4
	noise             *rhi.Texture
	resources         []rhi.Resource
	screenCfg         *window.ScreenConfig
	cfg               *CavityConfig
}

func NewCavityPass(
	screenCfg *window.ScreenConfig,
	quad *rhi.Mesh,
	cfg *CavityConfig,
	noise *rhi.Texture,
	blurProgram *rhi.Program,
	compositorProgram *rhi.Program,
) (*CavityPass, error) {
	p := &CavityPass{
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

	p.resources = append(p.resources, p.cavity, p.blur, p.composition, p.cavityProgram)
	return p, nil
}

func (p *CavityPass) GetName() string {
	return "cavity"
}

func (p *CavityPass) ResizeCallback() {

	fbWidth, fbHeight := p.screenCfg.GetScreenSize()
	// resize attachments
	p.cavity.Resize(fbWidth, fbHeight)
	p.composition.Resize(fbWidth, fbHeight)
}

func (p *CavityPass) initFramebuffers() error {

	fbWidth, fbHeight := p.screenCfg.GetScreenSize()
	var err error

	// init cavity buffer
	cavity, err := rhi.NewFramebuffer(fbWidth, fbHeight)
	if err != nil {
		return fmt.Errorf("cavity pass - failed to create cavity fbo - %w", err)
	}
	cavity.Bind()
	err = cavity.NewColorAttachment(rhi.FormatR8)
	if !cavity.Check() || err != nil {
		cavity.Delete()
		return fmt.Errorf("cavity pass - cavity fbo not completed %w", err)
	}
	p.cavity = cavity

	// init blur buffer
	blur, err := rhi.NewFramebuffer(fbWidth, fbHeight)
	if err != nil {
		return fmt.Errorf("cavity pass - failed to create blur fbo - %w", err)
	}
	blur.Bind()
	err = blur.NewColorAttachment(rhi.FormatR8)
	if !blur.Check() || err != nil {
		blur.Delete()
		return fmt.Errorf("cavity pass - blur fbo not completed %w", err)
	}
	p.blur = blur

	// init composition buffer
	composition, err := rhi.NewFramebuffer(fbWidth, fbHeight)
	if err != nil {
		return fmt.Errorf("cavity pass - failed to create composition fbo - %w", err)
	}
	composition.Bind()
	err = composition.NewColorAttachment(rhi.FormatRGBA8)
	if !composition.Check() || err != nil {
		composition.Delete()
		return fmt.Errorf("cavity pass - composition fbo not completed %w", err)
	}
	p.composition = composition
	return nil
}

func (p *CavityPass) initPrograms() error {
	// init cavity drawing program
	cavityProgram, err := rhi.NewProgram(
		files.GetShaderPath("screen.vert"),
		files.GetShaderPath("cavity.frag"),
	)
	if err != nil {
		return fmt.Errorf("cavity pass - failed to load cavity program - %w", err)
	}
	p.cavityProgram = cavityProgram
	return nil
}

func (p *CavityPass) createAndSendSamples() {
	// samples (random points [-1, 1] closer to center)
	p.cavityProgram.Use()
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
		p.cavityProgram.Set2f("u_samples["+strconv.Itoa(i)+"]", sample)
	}
}

func (p *CavityPass) SetProjectionMatrix(m mgl.Mat4) {
	p.projection = m
}

// returns result color
func (p *CavityPass) GetColor() *rhi.Texture {
	return p.composition.GetColorTexture(0)
}

// returns result fbo
func (p *CavityPass) GetResultFramebuffer() *rhi.Framebuffer {
	return p.composition
}

// returns blurred cavity color
func (p *CavityPass) GetBlur() *rhi.Texture {
	return p.blur.GetColorTexture(0)
}

// returns raw cavity color
func (p *CavityPass) GetOcclusion() *rhi.Texture {
	return p.cavity.GetColorTexture(0)
}

// RenderPass
// 0 - color
// 1 - normal
// 2 - depth
func (p *CavityPass) RenderPass(src *pipeline.DeferredRenderResult) {
	// -- cavity render pass
	p.cavity.Bind()
	context.ClearColorBuffer()
	p.cavityProgram.Use()

	// bind and send normal
	src.Normal.BindToSlot(0)
	p.cavityProgram.Set1i("u_normal", 0)
	// bind and send depth
	src.Depth.BindToSlot(1)
	p.cavityProgram.Set1i("u_depth", 1)
	// send noise texture (random samples rotations)
	p.noise.BindToSlot(2)
	p.cavityProgram.Set1i("u_noise", 2)

	// send radius
	p.cavityProgram.Set1f("u_radius", p.cfg.Radius)
	// send depth bias
	p.cavityProgram.Set1f("u_depthbias", p.cfg.DepthBias)
	// send intensity
	p.cavityProgram.Set1f("u_intensity", p.cfg.Intensity)
	// send kernel size
	p.cavityProgram.Set1i("u_kernel_size", p.cfg.KernelSize)
	// send noise scale
	noiseScale := mgl.Vec2{
		float32(p.screenCfg.Width) / float32(noiseSize),
		float32(p.screenCfg.Height) / float32(noiseSize),
	}
	p.cavityProgram.Set2f("u_noise_scale", noiseScale)
	// send inv projection
	p.cavityProgram.SetMat4("u_invprojection", p.projection.Inv())

	p.mesh.Draw()

	// -- Blur render pass
	p.blur.Bind()
	p.blurProgram.Use()
	context.ClearColorBuffer()
	// send uniforms
	//
	// cavity texture
	p.cavity.GetColorTexture(0).BindToSlot(0)
	p.blurProgram.Set1i("u_overlay", 0)
	// blur size
	p.blurProgram.Set1i("u_blur_size", p.cfg.BlurSize)

	p.mesh.Draw()

	// -- Compositor render pass
	p.composition.Bind()
	p.compositorProgram.Use()
	context.ClearColorBuffer()
	// send uniforms
	//
	// color texture
	src.Color.BindToSlot(0)
	p.compositorProgram.Set1i("u_color", 0)
	// occlusion texture
	p.blur.GetColorTexture(0).BindToSlot(1)
	p.compositorProgram.Set1i("u_overlay", 1)
	// levels cfg
	p.compositorProgram.Set1f("u_blackpoint", p.cfg.BlackPoint)
	p.compositorProgram.Set1f("u_whitepoint", p.cfg.WhitePoint)

	p.mesh.Draw()
}

func (p *CavityPass) Use() bool {
	return p.cfg.Use
}

func (p *CavityPass) GetConfig() interface{} {
	return p.cfg
}

func (p *CavityPass) Delete() {
	for _, r := range p.resources {
		if r != nil {
			r.Delete()
		}
	}
}
