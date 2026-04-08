package config

import (
	"github.com/illoprin/retro-fps-kit-go/pkg/render/post"
)

const (
	WindowWidth                    = 1600
	WindowHeight                   = 720
	WindowTitle                    = "Retro FPS Kit - Demo"
	DefaultResolutionRatio float32 = 0.5
)

var (
	EyeAdaptionConfig = &post.EyeAdaptionConfig{
		Use:           true,
		Radius:        300,
		AvgGray:       0.18,
		AdaptionSpeed: 0.008,
		Exposure:      1,
	}

	SSAOConfig = &post.SSAOConfig{
		Use:        false,
		KernelSize: 30,
		Radius:     0.5,
		Bias:       0.005,
		WhitePoint: 0.971,
		BlackPoint: 0.39,
		BlurSize:   2,
	}

	CavityConfig = &post.CavityConfig{
		Use:        true,
		Radius:     17,
		DepthBias:  0.001,
		Intensity:  3.5,
		KernelSize: 32,
		WhitePoint: 1.0,
		BlackPoint: 0.26,
		BlurSize:   3,
	}

	BloomConfig = &post.BloomConfig{
		Use:       true,
		Threshold: 1.124,
		Levels:    4,
		MinRadius: 0.6,
		MaxRadius: 4.2,
		Tint:      [3]float32{0.65, 0.82, 1.0},
		Intensity: 1.6,
	}

	ToneMappingConfig = &post.ToneMappingConfig{
		Gamma:   2.2,
		Tonemap: post.ACESFilmTonemap,
	}

	ColorGradingConfig = &post.ColorGradingConfig{
		Contrast:       2.28,
		Saturation:     0.84,
		Brightness:     1.57,
		ShadowsColor:   [3]float32{.16, .18, .3},
		MidColor:       [3]float32{.68, .518, .33},
		HighlightColor: [3]float32{.938, .641, .438},
		ColorStrength:  0.47,
		Use:            true,
	}

	VignetteConfig = &post.VignetteConfig{
		Radius:   0.9,
		Softness: 0.535,
		Use:      true,
	}
)
