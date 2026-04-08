package app

import (
	"fmt"

	"github.com/illoprin/retro-fps-kit-go/pkg/app/config"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/files"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/window"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/post"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
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
func NewDefaultPipeline(screen *window.ScreenConfig) (*DefaultPipeline, error) {
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

	p.resources = append(p.resources, quad, noiseTexture, blurProg, compProg)

	// create pass objects

	// shared
	shared := post.PassSharedResources{
		ScreenConfig: screen,
		MeshQuad:     quad,
	}

	// helper function to create pass

	// -- eye adaption pass
	eyeAdaptionPass, err := post.NewEyeAdaptionPass(shared, config.EyeAdaptionConfig)
	if err != nil {
		p.Delete()
		return nil, fmt.Errorf("eye adaption pass - %w", err)
	}

	// -- ssao
	ssaoPass, err := post.NewSSAOPass(shared, config.SSAOConfig,
		noiseTexture, blurProg, compProg)
	if err != nil {
		p.Delete()
		return nil, fmt.Errorf("ssao pass - %w", err)
	}

	// -- cavity occlusion
	cavityPass, err := post.NewCavityPass(shared, config.CavityConfig,
		blurProg, compProg)
	if err != nil {
		p.Delete()
		return nil, fmt.Errorf("cavity pass - %w", err)
	}

	// -- bloomPass
	bloomPass, err := post.NewBloomPass(shared, config.BloomConfig)
	if err != nil {
		p.Delete()
		return nil, fmt.Errorf("bloom pass - %w", err)
	}

	// -- tone mapping
	toneMappingPass, err := post.NewToneMappingPass(shared, config.ToneMappingConfig)
	if err != nil {
		p.Delete()
		return nil, fmt.Errorf("tone mapping pass - %w", err)
	}

	// -- color grading
	colorGradingPass, err := post.NewColorGradingPass(shared, config.ColorGradingConfig)
	if err != nil {
		p.Delete()
		return nil, fmt.Errorf("color grading pass - %w", err)
	}

	// -- vignettePass
	vignettePass, err := post.NewVignettePass(shared, config.VignetteConfig)
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
	err = p.Build()
	if err != nil {
		p.Delete()
		return nil, err
	}

	return p, nil
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
