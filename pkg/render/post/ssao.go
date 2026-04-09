package post

import (
	"fmt"
	"math/rand"
	"strconv"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/core/files"
	coremath "github.com/illoprin/retro-fps-toolkit-go/pkg/core/math"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/core/window"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/render/context"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/render/pipeline"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/render/rhi"
)

type SSAOConfig struct {
	Use        bool    `yaml:"use"`
	KernelSize int32   `yaml:"kernelSize"` // max - 64
	Radius     float32 `yaml:"radius"`
	Bias       float32 `yaml:"bias"`
	WhitePoint float32 `yaml:"whitePoint"`
	BlackPoint float32 `yaml:"blackPoint"`
	BlurSize   int32   `yaml:"blurSize"`
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
	screen            *window.ScreenConfig
	cfg               *SSAOConfig
}

func NewSSAOPass(
	s PassSharedResources,
	cfg *SSAOConfig,
	noiseTexture *rhi.Texture,
	blurProgram *rhi.Program,
	compositorProgram *rhi.Program,
) (*SSAOPass, error) {
	p := &SSAOPass{
		screen:            s.ScreenConfig,
		mesh:              s.MeshQuad,
		cfg:               cfg,
		noiseTexture:      noiseTexture,
		blurProgram:       blurProgram,
		compositorProgram: compositorProgram,
		resources:         make([]rhi.Resource, 0),
	}

	if err := p.initFramebuffers(); err != nil {
		return nil, err
	}

	if err := p.initPrograms(); err != nil {
		return nil, err
	}

	p.buildAndSendSamples()

	p.resources = append(p.resources, p.ssao, p.blur, p.composition, p.ssaoProgram)
	return p, nil
}

func (p *SSAOPass) ResizeCallback() {
	// get fb size
	fbWidth, fbHeight := p.screen.GetScreenSize()

	// resize attachments
	p.blur.Resize(fbWidth, fbHeight)
	p.ssao.Resize(fbWidth, fbHeight)
	p.composition.Resize(fbWidth, fbHeight)
}

func (p *SSAOPass) initFramebuffers() error {

	W, H := p.screen.GetScreenSize()
	hW, hH := W/2, H/2

	// init ssao buffer
	ssao := rhi.NewFramebuffer(hW, hH)
	ssao.Bind()
	ssao.NewColorAttachment(rhi.FormatR8, rhi.FilterLinear)
	if !ssao.Check() {
		ssao.Delete()
		return fmt.Errorf("raw fbo not completed")
	}
	p.ssao = ssao

	// init blur buffer
	blur := rhi.NewFramebuffer(hW, hH)
	blur.Bind()
	blur.NewColorAttachment(rhi.FormatR8, rhi.FilterLinear)
	if !blur.Check() {
		blur.Delete()
		return fmt.Errorf("blur fbo not completed")
	}
	p.blur = blur

	// init composition buffer
	composition := rhi.NewFramebuffer(W, H)
	composition.Bind()
	composition.NewColorAttachment(rhi.FormatRGB16F, rhi.FilterNearest)
	if !composition.Check() {
		composition.Delete()
		return fmt.Errorf("composition fbo not completed")
	}
	p.composition = composition
	return nil
}

func (p *SSAOPass) buildAndSendSamples() {
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
		return fmt.Errorf("failed to load ssao program - %w", err)
	}
	p.ssaoProgram = ssaoProgram
	return nil
}

// returns composition color
func (p *SSAOPass) GetColor() *rhi.Texture {
	return p.composition.GetColorTexture(0)
}

// GetDebugTextures implementing debug interface
func (p *SSAOPass) GetDebugTextures() []DebugTexture {
	return []DebugTexture{
		{"ssao.occlusion", p.ssao.GetColorTexture(0)},
		{"ssao.noise", p.noiseTexture},
		{"ssao.blurred", p.blur.GetColorTexture(0)},
	}
}

// returns result fbo
func (p *SSAOPass) GetResultFramebuffer() *rhi.Framebuffer {
	return p.composition
}

// RenderPass
func (p *SSAOPass) RenderPass(src *pipeline.DeferredRenderResult) {

	// -- SSAO render pass

	p.ssao.BindForDrawing()
	context.ClearColorBuffer()
	p.ssaoProgram.Use()

	// send uniforms
	//
	// normal texture
	src.Normal.BindToUnit(0)
	p.ssaoProgram.Set1i("u_normal", 0)
	// position texture
	src.Position.BindToUnit(1)
	p.ssaoProgram.Set1i("u_position", 1)
	// noise texture
	p.noiseTexture.BindToUnit(2)
	p.ssaoProgram.Set1i("u_noise", 2)
	// projection
	p.ssaoProgram.Set2f("u_proj_info", mgl.Vec2{p.proj.At(0, 0), p.proj.At(1, 1)})
	// samples
	if p.cfg.KernelSize > 64 {
		p.ssaoProgram.Set1i("u_kernel_size", 64)
	} else {
		p.ssaoProgram.Set1i("u_kernel_size", p.cfg.KernelSize)
	}
	// noise texture size
	noiseScale := mgl.Vec2{
		float32(p.screen.Width) / float32(noiseSize),
		float32(p.screen.Height) / float32(noiseSize),
	}
	p.ssaoProgram.Set2f("u_noise_scale", noiseScale)
	// radius and bias
	p.ssaoProgram.Set1f("u_radius", p.cfg.Radius)
	p.ssaoProgram.Set1f("u_bias", p.cfg.Bias)

	p.mesh.Draw()

	// -- Blur render pass

	p.blur.BindForDrawing()
	context.ClearColorBuffer()
	p.blurProgram.Use()
	// bind and send raw ssao data
	p.ssao.GetColorTexture(0).BindToUnit(0)
	p.blurProgram.Set1i("u_overlay", 0)
	// blur size
	p.blurProgram.Set1i("u_blur_size", p.cfg.BlurSize)

	p.mesh.Draw()

	// -- Compositor render pass

	p.composition.BindForDrawing()
	context.ClearColorBuffer()
	p.compositorProgram.Use()
	// send uniforms
	//
	// color texture
	src.Color.BindToUnit(0)
	p.compositorProgram.Set1i("u_color", 0)
	// ssao texture
	p.blur.GetColorTexture(0).BindToUnit(1)
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
