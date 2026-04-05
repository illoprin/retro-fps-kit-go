package config

import "github.com/illoprin/retro-fps-kit-go/pkg/render/passes"

const (
	WindowWidth                    = 1600
	WindowHeight                   = 720
	WindowTitle                    = "Retro FPS Kit - Demo"
	DefaultResolutionRatio float32 = 0.5
)

var (
	SSAOConfig = &passes.SSAOConfig{
		Use:        false,
		KernelSize: 30,
		Radius:     0.5,
		Bias:       0.005,
		WhitePoint: 0.971,
		BlackPoint: 0.39,
		BlurSize:   2,
	}

	EyeAdaptionConfig = &passes.EyeAdaptionConfig{
		Use:           true,
		Radius:        300,
		AvgGray:       0.23,
		AdaptionSpeed: 0.005,
	}

	CavityConfig = &passes.CavityConfig{
		Use:        true,
		Radius:     17,
		DepthBias:  0.001,
		Intensity:  3.5,
		KernelSize: 32,
		WhitePoint: 1.0,
		BlackPoint: 0.26,
		BlurSize:   3,
	}

	ColorGradingConfig = &passes.ColorGradingConfig{
		Gamma:          1.9,
		Exposure:       0.7,
		Contrast:       1.18,
		Saturation:     0.85,
		Brightness:     1.4,
		ShadowsColor:   [3]float32{.063, .102, .576},
		MidColor:       [3]float32{.494, .294, .067},
		HighlightColor: [3]float32{.903, .402, .061},
		ColorStrength:  0.68,
		Use:            true,
	}

	VignetteConfig = &passes.VignetteConfig{
		Radius:   0.85,
		Softness: 0.535,
		Use:      true,
	}
)
