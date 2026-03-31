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
	KernelSize int32 // max - 256
}

type CreaseOcclusionPass struct {
	crease            *render.Framebuffer
	composition       *render.Framebuffer
	creaseProgram     *render.Program
	compositorProgram *render.Program
	mesh              *render.Mesh
	samples           []mgl.Vec2
	projection        mgl.Mat4
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

	p.createAndSendSamples()

	p.resources = append(p.resources, p.crease, p.composition, p.creaseProgram, p.compositorProgram)
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

	// init crease buffer
	crease, err := render.NewFramebuffer(fbWidth, fbHeight)
	if err != nil {
		return fmt.Errorf("crease pass - failed to create crease fbo - %w", err)
	}
	crease.Bind()
	if err := crease.NewColorAttachment(render.FormatR8); err != nil {
		crease.Delete()
		return fmt.Errorf("crease pass - failed to create crease fbo - %w", err)
	}
	if !crease.Check() {
		crease.Delete()
		return fmt.Errorf("crease pass - crease fbo not completed %w", err)
	}
	p.crease = crease

	// init composition buffer
	composition, err := render.NewFramebuffer(fbWidth, fbHeight)
	if err != nil {
		return fmt.Errorf("crease pass - failed to create composition fbo - %w", err)
	}
	composition.Bind()
	if err := composition.NewColorAttachment(render.FormatRGBA8); err != nil {
		composition.Delete()
		return fmt.Errorf("crease pass - failed to create composition fbo - %w", err)
	}
	if !composition.Check() {
		composition.Delete()
		return fmt.Errorf("crease pass - composition fbo not completed %w", err)
	}
	p.composition = composition
	return nil
}

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
	position := src[3]

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
	// bind and send position
	position.Bind(2)
	p.creaseProgram.Set1i("u_position", 2)
	// send radius
	p.creaseProgram.Set1f("u_radius", p.cfg.Radius)
	// send depth bias
	p.creaseProgram.Set1f("u_depthbias", p.cfg.DepthBias)
	// send intensity
	p.creaseProgram.Set1f("u_intensity", p.cfg.Intensity)
	// send kernel size
	p.creaseProgram.Set1i("u_kernel_size", p.cfg.KernelSize)

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
