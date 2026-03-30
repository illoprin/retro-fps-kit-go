package postprocessing

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/go-gl/gl/v4.1-core/gl"
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-kit-go/src/engine/assetmgr"
	mathutils "github.com/illoprin/retro-fps-kit-go/src/math"
	"github.com/illoprin/retro-fps-kit-go/src/render"
	"github.com/illoprin/retro-fps-kit-go/src/window"
)

type SSAOConfig struct {
	Use              bool
	NoiseTextureSize int32
	KernelSize       int32
	Radius           float32
	Bias             float32
	BlackPoint       float32
	WhitePoint       float32
	BlurSize         int32
}

type SSAOPass struct {
	ssao              *render.Framebuffer
	blur              *render.Framebuffer
	composition       *render.Framebuffer
	ssaoProgram       *render.Program
	blurProgram       *render.Program
	compositorProgram *render.Program
	noiseTexture      *render.Texture
	mesh              *render.Mesh
	resources         []render.Resource
	samples           []mgl.Vec3
	mprojection       mgl.Mat4
	screenCfg         *window.ScreenConfig
	cfg               *SSAOConfig
}

func NewSSAOPass(
	screenCfg *window.ScreenConfig,
	quad *render.Mesh,
	cfg *SSAOConfig,
) (*SSAOPass, error) {
	p := &SSAOPass{
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

	if err := p.initNoisy(); err != nil {
		return nil, err
	}

	p.resources = append(p.resources, p.ssao, p.blur, p.composition, p.ssaoProgram, p.blurProgram, p.compositorProgram, p.noiseTexture)
	return p, nil
}

func (p *SSAOPass) GetName() string {
	return "ssao"
}

func (p *SSAOPass) ResizeCallback() {

	realScreenSize := mgl.Vec3{
		float32(p.screenCfg.Width) * p.screenCfg.ResolutionRatio,
		float32(p.screenCfg.Height) * p.screenCfg.ResolutionRatio,
	}

	// resize attachments
	p.blur.Resize(int32(realScreenSize[0]), int32(realScreenSize[1]))
	p.ssao.Resize(int32(realScreenSize[0]), int32(realScreenSize[1]))
	p.composition.Resize(int32(realScreenSize[0]), int32(realScreenSize[1]))
}

func (p *SSAOPass) initFramebuffers() error {

	realScreenSize := mgl.Vec3{
		float32(p.screenCfg.Width) * p.screenCfg.ResolutionRatio,
		float32(p.screenCfg.Height) * p.screenCfg.ResolutionRatio,
	}

	// init ssao buffer
	ssao, err := render.NewFramebuffer(int32(realScreenSize[0]), int32(realScreenSize[1]))
	if err != nil {
		return fmt.Errorf("ssao pass - failed to create ssao fbo - %w", err)
	}
	ssao.Bind()
	if err := ssao.NewColorAttachment(render.FormatR8); err != nil {
		ssao.Delete()
		return fmt.Errorf("ssao pass - failed to create ssao fbo - %w", err)
	}
	if !ssao.Check() {
		ssao.Delete()
		return fmt.Errorf("ssao pass - ssao fbo not completed %w", err)
	}
	p.ssao = ssao

	// init blur buffer
	blur, err := render.NewFramebuffer(int32(realScreenSize[0]), int32(realScreenSize[1]))
	if err != nil {
		return fmt.Errorf("ssao pass - failed to create blur fbo - %w", err)
	}
	blur.Bind()
	if err := blur.NewColorAttachment(render.FormatR8); err != nil {
		blur.Delete()
		return fmt.Errorf("ssao pass - failed to create blur fbo - %w", err)
	}
	if !blur.Check() {
		blur.Delete()
		return fmt.Errorf("ssao pass - blur fbo not completed %w", err)
	}
	p.blur = blur

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

func (p *SSAOPass) initNoisy() error {
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
		scale = mathutils.Lerp(0.1, 1.0, scale*scale)
		sample = sample.Mul(scale)

		p.samples[i] = sample
	}

	p.ssaoProgram.Use()
	for i := 0; i < int(p.cfg.KernelSize); i++ {
		p.ssaoProgram.Set3f("u_samples["+strconv.Itoa(i)+"]", p.samples[i])
	}

	// noise tex (random rotations for hemi-sphere)
	noiseSliceLen := p.cfg.NoiseTextureSize * p.cfg.NoiseTextureSize * 3
	noiseRaw := make([]byte, noiseSliceLen)
	for i := 0; i < int(noiseSliceLen/3); i++ {
		pix := rand.Float32()*2 - 1
		noiseRaw[i*3] = byte(pix * 255.0)
		noiseRaw[i*3+1] = byte(pix * 255.0)
		noiseRaw[i*3+2] = 0
	}
	noiseTextureConfig := render.TextureConfig{
		Type:            render.TextureType2D,
		Width:           p.cfg.NoiseTextureSize,
		Height:          p.cfg.NoiseTextureSize,
		Format:          render.FormatRGB8,
		FilterMin:       render.FilterNearest,
		FilterMag:       render.FilterNearest,
		WrapS:           render.WrapRepeat,
		WrapT:           render.WrapRepeat,
		GenerateMipmaps: false,
		Anisotropy:      0,
	}
	noiseTexture, err := render.NewTexture(noiseTextureConfig)
	if err != nil {
		return fmt.Errorf("ssao pass - failed to create noise texture %w", err)
	}
	p.noiseTexture = noiseTexture
	p.noiseTexture.UploadRGB(0, 0, p.cfg.NoiseTextureSize, p.cfg.NoiseTextureSize, noiseRaw)
	return nil
}

func (p *SSAOPass) initPrograms() error {
	// init ssao drawing program
	ssaoProgram, err := render.NewProgram(
		assetmgr.GetShaderPath("quad.vert"),
		assetmgr.GetShaderPath("ssao.frag"),
	)
	if err != nil {
		return fmt.Errorf("ssao pass - failed to load ssao program - %w", err)
	}
	p.ssaoProgram = ssaoProgram

	// init blur ao program
	blurProgram, err := render.NewProgram(
		assetmgr.GetShaderPath("quad.vert"),
		assetmgr.GetShaderPath("ssao_blur.frag"),
	)
	if err != nil {
		return fmt.Errorf("ssao pass - failed to load blur program - %w", err)
	}
	p.blurProgram = blurProgram

	// init compositor program
	compositorProgram, err := render.NewProgram(
		assetmgr.GetShaderPath("quad.vert"),
		assetmgr.GetShaderPath("ssao_compositor.frag"),
	)
	if err != nil {
		return fmt.Errorf("ssao pass - failed to ssao compositor load program - %w", err)
	}
	p.compositorProgram = compositorProgram
	return nil
}

func (p *SSAOPass) GetColor() *render.Texture {
	return p.composition.ColorTextures[0]
}

func (p *SSAOPass) GetRawSSAO() *render.Texture {
	return p.ssao.ColorTextures[0]
}

func (p *SSAOPass) GetNoise() *render.Texture {
	return p.noiseTexture
}

func (p *SSAOPass) GetBlurSSAO() *render.Texture {
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

	color := src[0]
	normal := src[1]
	depth := src[2]

	// -- SSAO render pass

	p.ssao.Bind()
	p.ssaoProgram.Use()
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)

	// send uniforms
	//
	// normal texture
	normal.Bind(0)
	p.ssaoProgram.Set1i("u_normal", 0)
	// depth texture
	depth.Bind(1)
	p.ssaoProgram.Set1i("u_depth", 1)
	// noise texture
	p.noiseTexture.Bind(2)
	p.ssaoProgram.Set1i("u_noise", 2)
	// inv projection
	p.ssaoProgram.SetMat4("u_invprojection", p.mprojection.Inv())
	p.ssaoProgram.SetMat4("u_projection", p.mprojection)
	// samples
	p.ssaoProgram.Set1i("u_kernel_size", p.cfg.KernelSize)
	// noise texture size
	noiseScale := mgl.Vec2{
		float32(p.screenCfg.Width) / float32(p.cfg.NoiseTextureSize),
		float32(p.screenCfg.Height) / float32(p.cfg.NoiseTextureSize),
	}
	p.ssaoProgram.Set2f("u_noise_scale", noiseScale)
	// radius and bias
	p.ssaoProgram.Set1f("u_radius", p.cfg.Radius)
	p.ssaoProgram.Set1f("u_bias", p.cfg.Bias)

	p.mesh.Draw()

	// -- Blur render pass

	p.blur.Bind()
	p.blurProgram.Use()
	gl.Clear(gl.COLOR_BUFFER_BIT)
	// bind and send raw ssao data
	p.ssao.ColorTextures[0].Bind(0)
	p.blurProgram.Set1i("u_raw_ssao", 0)
	p.blurProgram.Set1f("u_blackpoint", p.cfg.BlackPoint)
	p.blurProgram.Set1f("u_whitepoint", p.cfg.WhitePoint)
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
	// ssao texture
	p.blur.ColorTextures[0].Bind(1)
	p.compositorProgram.Set1i("u_ssao", 1)

	p.mesh.Draw()
}

func (p *SSAOPass) SetProjectionMatrix(m mgl.Mat4) {
	p.mprojection = m
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
