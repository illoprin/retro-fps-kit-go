package app

import (
	"fmt"
	"unsafe"

	"github.com/illoprin/retro-fps-toolkit-go/pkg/app/config"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/core/files"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/core/logger"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/core/window"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/render/post"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/render/rhi"
)

const (
	WindowTitle = "Retro FPS Kit"
)

var (
	DefaultConfig = config.Config{
		Window: config.DisplayConfig{
			Width:  1600,
			Height: 720,
			Ratio:  .5,
		},
		PostProcessing: config.PostProcessingConfig{
			EyeAdaption: &post.EyeAdaptionConfig{
				Use:           true,
				Radius:        300,
				AvgGray:       0.18,
				AdaptionSpeed: 0.008,
				Exposure:      1,
			},

			SSAO: &post.SSAOConfig{
				Use:        false,
				KernelSize: 30,
				Radius:     0.5,
				Bias:       0.005,
				WhitePoint: 0.971,
				BlackPoint: 0.39,
				BlurSize:   2,
			},

			Cavity: &post.CavityConfig{
				Use:        true,
				Radius:     14,
				DepthBias:  0.001,
				Intensity:  2.48,
				KernelSize: 51,
				WhitePoint: 1.0,
				BlackPoint: 0.26,
				BlurSize:   3,
			},

			Bloom: &post.BloomConfig{
				Use:               true,
				Threshold:         1.124,
				Levels:            4,
				MinRadius:         0.6,
				MaxRadius:         4.2,
				LensDirtIntensity: 4.7,
				Tint:              [3]float32{0.65, 0.82, 1.0},
				Intensity:         0.7,
			},

			Tonemapping: &post.ToneMappingConfig{
				Gamma:   2.2,
				Tonemap: "aces",
			},

			ColorGrading: &post.ColorGradingConfig{
				Contrast:       2.28,
				Saturation:     0.84,
				Brightness:     1.57,
				ShadowsColor:   [3]float32{.16, .18, .3},
				MidColor:       [3]float32{.68, .518, .33},
				HighlightColor: [3]float32{.938, .641, .438},
				ColorStrength:  0.47,
				Use:            true,
			},

			Vignette: &post.VignetteConfig{
				Radius:   0.9,
				Softness: 0.535,
				Use:      true,
			},
		},
	}
)

var (
	EyeAdaptionID  post.PassID = "eye_adaption"
	SSAOID         post.PassID = "ssao"
	CavityID       post.PassID = "cavity"
	BloomID        post.PassID = "bloom"
	ToneMappingID  post.PassID = "tonemapping"
	ColorGradingID post.PassID = "color_grading"
	VignetteID     post.PassID = "vignette"
)

type DefaultPipeline struct {
	*post.PostProcessingPipeline
	resources []rhi.Resource
}

// NewDefaultPipeline - allocates memory and create
// resources for default post processing pipeline
func NewDefaultPipeline(screen *window.ScreenConfig, cfg *config.PostProcessingConfig) (*DefaultPipeline, error) {
	// create default pipeline object
	p := &DefaultPipeline{
		PostProcessingPipeline: post.NewPipeline(),
	}

	// create auxiliary resources

	// init screen quad mesh
	quad := rhi.NewMesh()
	rhi.SetupBasicQuadMesh(quad)

	noiseTexture := post.CreateNoiseTexture()

	// blur - uses ssao and cavity
	blurProg, err := rhi.NewProgram(
		files.GetShaderPath("screen.vert"),
		files.GetShaderPath("overlay_blur.frag"),
	)
	if err != nil {
		return nil, err
	}

	// compositor - uses ssao and cavity
	compProg, err := rhi.NewProgram(
		files.GetShaderPath("screen.vert"),
		files.GetShaderPath("overlay_compositor.frag"),
	)
	if err != nil {
		return nil, err
	}

	// lens dirt texture

	var lensDirt *rhi.Texture = loadLensDirtTexture()

	p.resources = append(p.resources, quad, noiseTexture, blurProg, compProg)

	// create pass objects

	// shared
	shared := post.PassSharedResources{
		ScreenConfig: screen,
		MeshQuad:     quad,
	}

	// helper function to create pass

	// -- eye adaption pass
	eyeAdaptionPass, err := post.NewEyeAdaptionPass(shared, cfg.EyeAdaption)
	if err != nil {
		p.Delete()
		return nil, fmt.Errorf("eye adaption pass - %w", err)
	}

	// -- ssao
	ssaoPass, err := post.NewSSAOPass(shared, cfg.SSAO,
		noiseTexture, blurProg, compProg)
	if err != nil {
		p.Delete()
		return nil, fmt.Errorf("ssao pass - %w", err)
	}

	// -- cavity occlusion
	cavityPass, err := post.NewCavityPass(shared, cfg.Cavity,
		blurProg, compProg)
	if err != nil {
		p.Delete()
		return nil, fmt.Errorf("cavity pass - %w", err)
	}

	// -- bloomPass
	bloomPass, err := post.NewBloomPass(shared, cfg.Bloom, lensDirt)
	if err != nil {
		p.Delete()
		return nil, fmt.Errorf("bloom pass - %w", err)
	}

	// -- tone mapping
	toneMappingPass, err := post.NewToneMappingPass(shared, cfg.Tonemapping)
	if err != nil {
		p.Delete()
		return nil, fmt.Errorf("tone mapping pass - %w", err)
	}

	// -- color grading
	colorGradingPass, err := post.NewColorGradingPass(shared, cfg.ColorGrading)
	if err != nil {
		p.Delete()
		return nil, fmt.Errorf("color grading pass - %w", err)
	}

	// -- vignettePass
	vignettePass, err := post.NewVignettePass(shared, cfg.Vignette)
	if err != nil {
		p.Delete()
		return nil, fmt.Errorf("vignette pass - %w", err)
	}

	// create pipeline

	p.AddPass(&post.PassDescriptor{
		ID:    SSAOID,
		After: EyeAdaptionID,
		Pass:  ssaoPass,
	})
	p.AddPass(&post.PassDescriptor{
		ID:    CavityID,
		After: SSAOID,
		Pass:  cavityPass,
	})
	p.AddPass(&post.PassDescriptor{
		ID:    EyeAdaptionID,
		After: "",
		Pass:  eyeAdaptionPass,
	})
	p.AddPass(&post.PassDescriptor{
		ID:    BloomID,
		After: CavityID,
		Pass:  bloomPass,
	})
	p.AddPass(&post.PassDescriptor{
		ID:    ToneMappingID,
		After: BloomID,
		Pass:  toneMappingPass,
	})
	p.AddPass(&post.PassDescriptor{
		ID:    ColorGradingID,
		After: ToneMappingID,
		Pass:  colorGradingPass,
	})
	p.AddPass(&post.PassDescriptor{
		ID:    VignetteID,
		After: ColorGradingID,
		Pass:  vignettePass,
	})

	return p, nil
}

func loadLensDirtTexture() (lensDirt *rhi.Texture) {
	W, H, img, err := files.LoadImage(files.GetTexturePath("lens_dirt.png"))

	if err != nil {
		logger.Warnf("lens dirt not found")
		return nil
	}

	var x, y int
	lensDirtPix := make([]uint8, W*H)
	for i := 0; i < W*H; i++ {
		x = i % W
		y = i / W
		r, _, _, _ := img.At(x, y).RGBA()
		lensDirtPix[i] = uint8(r >> 8)
	}

	lensDirtConfig := rhi.DefaultTexture2DConfig(int32(W), int32(H))
	lensDirtConfig.Format = rhi.FormatR8
	lensDirtConfig.FilterMin = rhi.FilterLinear
	lensDirtConfig.FilterMag = rhi.FilterLinear
	lensDirt = rhi.NewTexture(lensDirtConfig)
	lensDirt.Upload2D(0, 0, unsafe.Pointer(&lensDirtPix[0]))
	return lensDirt
}

func (p *DefaultPipeline) Delete() {
	// clear resources
	for _, r := range p.resources {
		if r != nil {
			r.Delete()
		}
	}
	p.PostProcessingPipeline.Delete()
}
