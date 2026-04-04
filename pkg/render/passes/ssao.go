package passes

import (
	"fmt"
	"math/rand"
	"strconv"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/files"
	coremath "github.com/illoprin/retro-fps-kit-go/pkg/core/math"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/window"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/context"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/pipeline"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
)

type SSAOConfig struct {
	Use        bool
	KernelSize int32
	Radius     float32
	Bias       float32
	BlackPoint float32
	WhitePoint float32
	BlurSize   int32
}

type SSAOPass struct {
	ssao              *rhi.Framebuffer
	blur              *rhi.Framebuffer
	composition       *rhi.Framebuffer
	ssaoProgram       *rhi.Program
	blurProgram       *rhi.Program
	compositorProgram *rhi.Program
	noiseTexture      *rhi.Texture
	mesh              *rhi.Mesh
	resources         []rhi.Resource
	samples           []mgl.Vec3
	proj              mgl.Mat4
	screenCfg         *window.ScreenConfig
	cfg               *SSAOConfig
}

func NewSSAOPass(
	screenCfg *window.ScreenConfig,
	quad *rhi.Mesh,
	cfg *SSAOConfig,
	noiseTexture *rhi.Texture,
	blurProgram *rhi.Program,
	compositorProgram *rhi.Program,
) (*SSAOPass, error) {
	p := &SSAOPass{
		cfg:               cfg,
		screenCfg:         screenCfg,
		mesh:              quad,
		noiseTexture:      noiseTexture,
		blurProgram:       blurProgram,
		compositorProgram: compositorProgram,
	}

	if err := p.initFramebuffers(); err != nil {
		return nil, err
	}

	if err := p.initPrograms(); err != nil {
		return nil, err
	}

	p.initNoisy()

	p.resources = append(p.resources, p.ssao, p.blur, p.composition, p.ssaoProgram)
	return p, nil
}

func (p *SSAOPass) GetName() string {
	return "ssao"
}

func (p *SSAOPass) ResizeCallback() {
	// get fb size
	fbWidth, fbHeight := p.screenCfg.GetScreenSize()

	// resize attachments
	p.blur.Resize(fbWidth, fbHeight)
	p.ssao.Resize(fbWidth, fbHeight)
	p.composition.Resize(fbWidth, fbHeight)
}

func (p *SSAOPass) initFramebuffers() error {

	fbWidth, fbHeight := p.screenCfg.GetScreenSize()

	// init ssao buffer
	ssao, err := rhi.NewFramebuffer(fbWidth, fbHeight)
	if err != nil {
		return fmt.Errorf("ssao pass - failed to create ssao fbo - %w", err)
	}
	ssao.Bind()
	if err := ssao.NewColorAttachment(rhi.FormatR8); err != nil {
		ssao.Delete()
		return fmt.Errorf("ssao pass - failed to create ssao fbo - %w", err)
	}
	if !ssao.Check() {
		ssao.Delete()
		return fmt.Errorf("ssao pass - ssao fbo not completed %w", err)
	}
	p.ssao = ssao

	// init blur buffer
	blur, err := rhi.NewFramebuffer(fbWidth, fbHeight)
	if err != nil {
		return fmt.Errorf("ssao pass - failed to create blur fbo - %w", err)
	}
	blur.Bind()
	if err := blur.NewColorAttachment(rhi.FormatR8); err != nil {
		blur.Delete()
		return fmt.Errorf("ssao pass - failed to create blur fbo - %w", err)
	}
	if !blur.Check() {
		blur.Delete()
		return fmt.Errorf("ssao pass - blur fbo not completed %w", err)
	}
	p.blur = blur

	// init composition buffer
	composition, err := rhi.NewFramebuffer(fbWidth, fbHeight)
	if err != nil {
		return fmt.Errorf("ssao pass - failed to create composition fbo - %w", err)
	}
	composition.Bind()
	if err := composition.NewColorAttachment(rhi.FormatRGBA8); err != nil {
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

func (p *SSAOPass) initNoisy() {
	// samples (hemi-sphere random points)
	p.samples = make([]mgl.Vec3, p.cfg.KernelSize)
	for i := 0; i < int(p.cfg.KernelSize); i++ {

		// generate random hemi-sphere sample
		sample := mgl.Vec3{
			rand.Float32()*2.0 - 1.0,
			rand.Float32()*2.0 - 1.0,
			rand.Float32(),
		}
		sample = sample.Normalize()
		sample = sample.Mul(rand.Float32())

		// distribution adjustment
		//
		// more samples closer to the center
		scale := float32(i) / 64.0
		scale = coremath.Lerp(0.1, 1.0, scale*scale)
		sample = sample.Mul(scale)

		p.samples[i] = sample
	}

	p.ssaoProgram.Use()
	for i := 0; i < int(p.cfg.KernelSize); i++ {
		p.ssaoProgram.Set3f("u_samples["+strconv.Itoa(i)+"]", p.samples[i])
	}
}

func (p *SSAOPass) initPrograms() error {
	// init ssao drawing program
	ssaoProgram, err := rhi.NewProgram(
		files.GetShaderPath("screen.vert"),
		files.GetShaderPath("ssao.frag"),
	)
	if err != nil {
		return fmt.Errorf("ssao pass - failed to load ssao program - %w", err)
	}
	p.ssaoProgram = ssaoProgram
	return nil
}

// returns composition color
func (p *SSAOPass) GetColor() *rhi.Texture {
	return p.composition.GetColorTexture(0)
}

// returns raw ssao color
func (p *SSAOPass) GetOcclusion() *rhi.Texture {
	return p.ssao.GetColorTexture(0)
}

// returns kernel rotation noise
func (p *SSAOPass) GetNoise() *rhi.Texture {
	return p.noiseTexture
}

// returns blurred ssao color
func (p *SSAOPass) GetBlur() *rhi.Texture {
	return p.blur.GetColorTexture(0)
}

// returns result fbo
func (p *SSAOPass) GetResultFramebuffer() *rhi.Framebuffer {
	return p.composition
}

// RenderPass
func (p *SSAOPass) RenderPass(src *pipeline.DeferredRenderResult) {

	// -- SSAO render pass

	p.ssao.Bind()
	context.ClearColorBuffer()
	p.ssaoProgram.Use()

	// send uniforms
	//
	// normal texture
	src.Normal.BindToSlot(0)
	p.ssaoProgram.Set1i("u_normal", 0)
	// position texture
	src.Position.BindToSlot(1)
	p.ssaoProgram.Set1i("u_position", 1)
	// noise texture
	p.noiseTexture.BindToSlot(2)
	p.ssaoProgram.Set1i("u_noise", 2)
	// projection
	p.ssaoProgram.Set2f("u_proj_info", mgl.Vec2{p.proj.At(0, 0), p.proj.At(1, 1)})
	// samples
	p.ssaoProgram.Set1i("u_kernel_size", p.cfg.KernelSize)
	// noise texture size
	noiseScale := mgl.Vec2{
		float32(p.screenCfg.Width) / float32(noiseSize),
		float32(p.screenCfg.Height) / float32(noiseSize),
	}
	p.ssaoProgram.Set2f("u_noise_scale", noiseScale)
	// radius and bias
	p.ssaoProgram.Set1f("u_radius", p.cfg.Radius)
	p.ssaoProgram.Set1f("u_bias", p.cfg.Bias)

	p.mesh.Draw()

	// -- Blur render pass

	p.blur.Bind()
	context.ClearColorBuffer()
	p.blurProgram.Use()
	// bind and send raw ssao data
	p.ssao.GetColorTexture(0).BindToSlot(0)
	p.blurProgram.Set1i("u_overlay", 0)
	// blur size
	p.blurProgram.Set1i("u_blur_size", p.cfg.BlurSize)

	p.mesh.Draw()

	// -- Compositor render pass

	p.composition.Bind()
	context.ClearColorBuffer()
	p.compositorProgram.Use()
	// send uniforms
	//
	// color texture
	src.Color.BindToSlot(0)
	p.compositorProgram.Set1i("u_color", 0)
	// ssao texture
	p.blur.GetColorTexture(0).BindToSlot(1)
	p.compositorProgram.Set1i("u_overlay", 1)
	// levels cfg
	p.compositorProgram.Set1f("u_blackpoint", p.cfg.BlackPoint)
	p.compositorProgram.Set1f("u_whitepoint", p.cfg.WhitePoint)

	p.mesh.Draw()
}

func (p *SSAOPass) SetProjectionMatrix(m mgl.Mat4) {
	p.proj = m
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
