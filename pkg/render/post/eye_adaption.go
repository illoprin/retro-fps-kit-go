package post

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"unsafe"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/files"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/window"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/context"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/pipeline"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
)

const (
	lumaSamples = 16
)

type EyeAdaptionConfig struct {
	Use           bool
	Radius        float32
	AvgGray       float32
	AdaptionSpeed float32
	Exposure      float32
}

type EyeAdaptionPass struct {
	luma        *rhi.Framebuffer
	ldr         *rhi.Framebuffer
	lumaProgram *rhi.Program
	ldrProgram  *rhi.Program
	samples     []mgl.Vec2
	resources   []rhi.Resource
	mesh        *rhi.Mesh
	screen      *window.ScreenConfig
	cfg         *EyeAdaptionConfig

	prevExposure float32
}

func NewEyeAdaptionPass(
	s PassSharedResources,
	cfg *EyeAdaptionConfig,
) (*EyeAdaptionPass, error) {
	p := &EyeAdaptionPass{
		cfg:       cfg,
		screen:    s.ScreenConfig,
		mesh:      s.MeshQuad,
		resources: make([]rhi.Resource, 4),
	}

	// create luma framebuffer
	luma := rhi.NewFramebuffer(1, 1)
	luma.Bind()
	luma.NewColorAttachment(rhi.FormatR32F, rhi.FilterNearest)
	if !luma.Check() {
		luma.Delete()
		return nil, fmt.Errorf("incomplete luma fbo")
	}
	p.luma = luma

	// create result framebuffer
	sW, sH := p.screen.GetScreenSize()
	ldr := rhi.NewFramebuffer(sW, sH)
	ldr.Bind()
	ldr.NewColorAttachment(rhi.FormatRGB16F, rhi.FilterNearest)
	if !ldr.Check() {
		ldr.Delete()
		return nil, fmt.Errorf("incomplete ldr fbo")
	}
	p.ldr = ldr

	// create luma program
	lumaProgram, err := rhi.NewProgram(
		files.GetShaderPath("screen.vert"),
		files.GetShaderPath("average_luma.frag"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create luma program - %w", err)
	}
	p.lumaProgram = lumaProgram

	// create ldr (composition) program
	ldrProgram, err := rhi.NewProgram(
		files.GetShaderPath("screen.vert"),
		files.GetShaderPath("eye_adaption.frag"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ldr program - %w", err)
	}
	p.ldrProgram = ldrProgram

	p.resources = append(p.resources, luma, ldr, lumaProgram, ldrProgram)

	p.buildAndSendSamples()

	return p, nil

}

func (p *EyeAdaptionPass) buildAndSendSamples() {
	// samples (random points [-1, 1] closer to center)
	p.lumaProgram.Use()
	p.samples = make([]mgl.Vec2, lumaSamples)
	for i := 0; i < lumaSamples; i++ {
		u := rand.Float64()
		v := rand.Float64()

		a := u * 2 * math.Pi
		r := math.Pow(v, distributionCoefficient)

		sample := mgl.Vec2{
			float32(math.Cos(a) * r),
			float32(math.Sin(a) * r),
		}

		p.samples[i] = sample
		p.lumaProgram.Set2f("u_samples["+strconv.Itoa(i)+"]", sample)
	}
}

func (p *EyeAdaptionPass) RenderPass(src *pipeline.DeferredRenderResult) {

	// -- LUMA render pass (get average luma)

	p.luma.BindForDrawing()
	context.ClearColorBuffer()
	p.lumaProgram.Use()
	// send radius
	p.lumaProgram.Set1f("u_radius", p.cfg.Radius)
	// send source color
	src.Color.BindToUnit(0)
	p.lumaProgram.Set1i("u_color", 0)

	p.mesh.Draw()

	// -- Read luma and calculate smoothed exposure

	// save luma value
	// WARN gpu stall
	var currentLuma float32
	p.luma.ReadPixels(0, 0, 1, 1, 0, rhi.FormatR32F, unsafe.Pointer(&currentLuma))
	alpha := p.cfg.AdaptionSpeed
	currentExposure := p.cfg.AvgGray / max(currentLuma, 0.001)
	smoothedExposure := p.prevExposure*(1.0-alpha) + currentExposure*alpha
	smoothedExposure = mgl.Clamp(smoothedExposure, 0.5, 3.0)
	p.prevExposure = smoothedExposure

	// -- Result render pass (apply average exposure)

	// send average luma
	p.ldr.BindForDrawing()
	context.ClearColorBuffer()
	p.ldrProgram.Use()
	// send color
	src.Color.BindToUnit(0)
	p.ldrProgram.Set1i("u_color", 0)
	// apply exposure
	p.ldrProgram.Set1f("u_exposure", smoothedExposure*p.cfg.Exposure)

	p.mesh.Draw()

}

func (p *EyeAdaptionPass) ResizeCallback() {
	sW, sH := p.screen.GetScreenSize()

	p.ldr.Resize(sW, sH)
}

func (p *EyeAdaptionPass) GetColor() *rhi.Texture {
	return p.ldr.GetColorTexture(0)
}

func (p *EyeAdaptionPass) GetResultFramebuffer() *rhi.Framebuffer {
	return p.ldr
}

func (p *EyeAdaptionPass) GetConfig() interface{} {
	return p.cfg
}

func (p *EyeAdaptionPass) Use() bool {
	return p.cfg.Use
}

// GetDebugTextures implementing debug interface
func (p *EyeAdaptionPass) GetDebugTextures() []DebugTexture {
	return []DebugTexture{
		{"eye_adaption .color", p.ldr.GetColorTexture(0)},
	}
}

func (p *EyeAdaptionPass) Delete() {
	for _, r := range p.resources {
		if r != nil {
			r.Delete()
		}
	}
}
