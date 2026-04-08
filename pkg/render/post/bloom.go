package post

import (
	"fmt"
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/files"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/window"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/context"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/pipeline"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
)

type BloomConfig struct {
	Use       bool
	Threshold float32
	Levels    int32
	MinRadius float32
	MaxRadius float32
	Tint      [3]float32
	Intensity float32
}

// Blur radiuses:
// sm = 1.4
// med = sm * 1.5
// lg = med * 1.4

// c - сила увеличения радиуса с количеством итераций
// итерационно вычислили, что это 0.7
// nextRadius = currLevel * с - с + baseRadius

type BloomPass struct {
	bright *rhi.Framebuffer
	blur   *rhi.Framebuffer
	ping   *rhi.Framebuffer
	pong   *rhi.Framebuffer
	result *rhi.Framebuffer

	brightProgram  *rhi.Program
	blurProgram    *rhi.Program
	combineProgram *rhi.Program
	addProgram     *rhi.Program

	mesh      *rhi.Mesh
	screenCfg *window.ScreenConfig
	cfg       *BloomConfig

	resources []rhi.Resource
}

func NewBloomPass(
	s PassSharedResources,
	cfg *BloomConfig,
) (*BloomPass, error) {

	p := &BloomPass{
		screenCfg: s.ScreenConfig,
		mesh:      s.MeshQuad,
		cfg:       cfg,
	}

	if err := p.initFramebuffers(); err != nil {
		return nil, err
	}

	if err := p.initPrograms(); err != nil {
		return nil, err
	}

	p.resources = append(p.resources,
		p.bright, p.result, p.ping, p.pong, p.blur,
		p.brightProgram, p.blurProgram, p.addProgram, p.combineProgram,
	)

	return p, nil
}

func (p *BloomPass) GetConfig() interface{} {
	return p.cfg
}

func (p *BloomPass) Use() bool {
	return p.cfg.Use
}

func (p *BloomPass) GetColor() *rhi.Texture {
	return p.result.GetColorTexture(0)
}

func (p *BloomPass) GetResultFramebuffer() *rhi.Framebuffer {
	return p.result
}

// GetDebugTextures implementing debug interface
func (p *BloomPass) GetDebugTextures() []DebugTexture {
	return []DebugTexture{
		{"bloom.blur", p.blur.GetColorTexture(0)},
		{"bloom.color", p.result.GetColorTexture(0)},
	}
}

func (p *BloomPass) ResizeCallback() {
	W, H := p.screenCfg.GetScreenSize()
	hW, hH := W/2, H/2

	p.bright.Resize(hW, hH)
	p.ping.Resize(hW, hH)
	p.pong.Resize(hW, hH)
	p.blur.Resize(hW, hH)
	p.result.Resize(W, H)
}

func (p *BloomPass) initFramebuffers() error {
	W, H := p.screenCfg.GetScreenSize()
	hW, hH := W/2, H/2

	makeFBO := func(w, h int32, ft rhi.TextureFilter) (*rhi.Framebuffer, error) {
		f := rhi.NewFramebuffer(w, h)
		f.Bind()
		f.NewColorAttachment(rhi.FormatRGB16F, ft)
		if !f.Check() {
			f.Delete()
			return nil, fmt.Errorf("bloom fbo not complete")
		}
		return f, nil
	}

	var err error
	if p.bright, err = makeFBO(hW, hH, rhi.FilterLinear); err != nil {
		return err
	}
	if p.blur, err = makeFBO(hW, hH, rhi.FilterLinear); err != nil {
		return err
	}
	if p.ping, err = makeFBO(hW, hH, rhi.FilterLinear); err != nil {
		return err
	}
	if p.pong, err = makeFBO(hW, hH, rhi.FilterLinear); err != nil {
		return err
	}
	if p.result, err = makeFBO(W, H, rhi.FilterNearest); err != nil {
		return err
	}

	return nil
}

func (p *BloomPass) initPrograms() error {
	var err error

	p.brightProgram, err = rhi.NewProgram(
		files.GetShaderPath("screen.vert"),
		files.GetShaderPath("bloom_bright.frag"),
	)
	if err != nil {
		return err
	}

	p.blurProgram, err = rhi.NewProgram(
		files.GetShaderPath("screen.vert"),
		files.GetShaderPath("blur.frag"),
	)
	if err != nil {
		return err
	}

	p.addProgram, err = rhi.NewProgram(
		files.GetShaderPath("screen.vert"),
		files.GetShaderPath("bloom_accumulate.frag"),
	)
	if err != nil {
		return err
	}

	p.combineProgram, err = rhi.NewProgram(
		files.GetShaderPath("screen.vert"),
		files.GetShaderPath("bloom_combine.frag"),
	)
	if err != nil {
		return err
	}

	return nil
}

func (p *BloomPass) performBlur(input *rhi.Texture, target *rhi.Framebuffer) {
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.ONE, gl.ONE)

	src := input

	for i := 0; i <= int(p.cfg.Levels); i++ {

		t := float32(i+1) / (float32(p.cfg.Levels-1) + 0.001)
		nextRadius := p.cfg.MinRadius * float32(math.Pow(2.0, float64(t*p.cfg.MaxRadius)))
		nextWeight := float32(math.Exp(-float64(i) * .1))

		// --- HORIZONTAL → ping
		p.ping.BindForDrawing()
		context.ClearColorBuffer()
		p.blurProgram.Use()

		p.blurProgram.Set1i("u_horizontal", 1)
		p.blurProgram.Set1f("u_radius", nextRadius)

		src.BindToUnit(0)
		p.blurProgram.Set1i("u_color", 0)

		p.mesh.Draw()

		// --- VERTICAL → pong
		p.pong.BindForDrawing()
		context.ClearColorBuffer()

		p.blurProgram.Set1i("u_horizontal", 0)

		p.ping.GetColorTexture(0).BindToUnit(0)
		p.mesh.Draw()

		// --- ADD → blurred
		target.BindForDrawing()

		p.addProgram.Use()

		p.pong.GetColorTexture(0).BindToUnit(0)
		p.addProgram.Set1i("u_add", 0)
		p.addProgram.Set1f("u_weight", nextWeight)

		p.mesh.Draw()

		// --- следующий уровень берёт уже размытую картинку
		src = p.pong.GetColorTexture(0)
	}

	gl.Disable(gl.BLEND)
}

func (p *BloomPass) RenderPass(src *pipeline.DeferredRenderResult) {

	// --- 1. Bright pass

	p.bright.BindForDrawing()
	context.ClearColorBuffer()
	p.brightProgram.Use()

	src.Color.BindToUnit(0)
	p.brightProgram.Set1i("u_color", 0)
	p.brightProgram.Set1f("u_threshold", p.cfg.Threshold)

	p.mesh.Draw()

	// --- 2. Blur

	p.blur.Bind()
	context.ClearColorBuffer()

	p.performBlur(p.bright.GetColorTexture(0), p.blur)

	// --- 3. Combine

	p.result.BindForDrawing()
	context.ClearColorBuffer()
	p.combineProgram.Use()

	src.Color.BindToUnit(0)
	p.combineProgram.Set1i("u_hdr", 0)

	p.blur.GetColorTexture(0).BindToUnit(1)
	p.combineProgram.Set1i("u_bloom", 1)

	p.combineProgram.Set1f("u_intensity", p.cfg.Intensity)
	p.combineProgram.Set3f("u_tint", p.cfg.Tint)

	p.mesh.Draw()
}

func (p *BloomPass) Delete() {
	for _, r := range p.resources {
		if r != nil {
			r.Delete()
		}
	}
}
