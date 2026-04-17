package ui

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/post"
)

type ConfigUI interface {
	GetName() string
	ShowUI()
}

type EyeAdaptionConfigUI struct {
	*post.EyeAdaptionConfig
}

func (c *EyeAdaptionConfigUI) GetName() string {
	return "Eye Adaption (HDR)"
}

func (c *EyeAdaptionConfigUI) ShowUI() {
	imgui.Checkbox("ea_Use", &c.Use)
	imgui.SliderFloat("ea_Radius", &c.Radius, 4, 500)
	imgui.SliderFloat("ea_Speed", &c.AdaptionSpeed, 0.003, 0.06)
	imgui.SliderFloat("ea_Gray", &c.AvgGray, 0.01, 0.3)
	imgui.SliderFloat("ea_Exposure", &c.Exposure, 0.5, 3.0)
	imgui.SliderFloat("ea_ExposureMin", &c.ExposureMin, 0.1, 1.0)
	imgui.SliderFloat("ea_ExposureMax", &c.ExposureMax, 1.1, 20.0)
}

type BloomConfigUI struct {
	*post.BloomConfig
}

func (c *BloomConfigUI) GetName() string {
	return "Bloom (HDR)"
}

func (c *BloomConfigUI) ShowUI() {
	imgui.Checkbox("b_Use", &c.Use)
	imgui.SliderFloat("b_Intensity", &c.Intensity, 0.5, 10.0)
	imgui.SliderFloat("b_Threshold", &c.Threshold, 0.3, 10.0)
	imgui.SliderFloat("b_LensDirtIntensity", &c.LensDirtIntensity, 0, 10.0)
	imgui.SliderInt("b_Levels", &c.Levels, 1, 8)
	imgui.SliderFloat("b_MinRadius", &c.MinRadius, 0.5, 2)
	imgui.SliderFloat("b_MaxRadius", &c.MaxRadius, 1.0, 20.0)
	imgui.ColorEdit3("b_Tint", &c.Tint)
}

type ToneMappingConfigUI struct {
	*post.ToneMappingConfig
}

func (c *ToneMappingConfigUI) GetName() string {
	return "Tone Mapping (HDR -> LDR)"
}

var toneMapTypeItems = []string{
	post.ACESFilmTonemap,
	post.UnchartedTonemap,
	post.ReinhardTonemap,
}

func (c *ToneMappingConfigUI) ShowUI() {
	imgui.SliderFloat("tm_Gamma", &c.Gamma, 0.5, 4)

	// combo items

	// enum (1,2,3) → index (0,1,2)
	current := post.ToneMapEnum[c.Tonemap] - 1

	if imgui.BeginCombo("##combo", c.Tonemap) {
		for i, _ := range toneMapTypeItems {
			isSelected := current == (i - 1)
			if imgui.SelectableBool(toneMapTypeItems[i]) {
				current = i
				c.Tonemap = toneMapTypeItems[i]
			}
			if isSelected {
				imgui.SetItemDefaultFocus()
			}
		}
		imgui.EndCombo()
	}
}

type SSAOConfigUI struct {
	*post.SSAOConfig
}

func (c *SSAOConfigUI) GetName() string {
	return "SSAO (HDR)"
}

func (c *SSAOConfigUI) ShowUI() {
	imgui.Checkbox("ao_Use", &c.Use)
	imgui.SliderFloat("ao_Radius ", &c.Radius, 0.02, 2)
	imgui.SliderFloat("ao_Bias", &c.Bias, 0.0001, 0.5)
	imgui.SliderInt("ao_Samples", &c.KernelSize, 4, 64)
	imgui.SliderFloat("ao_Blackpoint", &c.BlackPoint, 0, 1)
	imgui.SliderFloat("ao_Whitepoint", &c.WhitePoint, 0, 1)
	imgui.SliderInt("ao_BlurSize", &c.BlurSize, 1, 8)
}

type CavityConfigUI struct {
	*post.CavityConfig
}

func (c *CavityConfigUI) GetName() string {
	return "Cavity Occlusion (HDR)"
}

func (c *CavityConfigUI) ShowUI() {
	imgui.Checkbox("c_Use", &c.Use)
	imgui.SliderFloat("c_Radius", &c.Radius, 5, 100)
	imgui.SliderFloat("c_Bias ", &c.DepthBias, 0.0001, 0.1)
	imgui.SliderFloat("c_Intensity", &c.Intensity, 0.01, 10)
	imgui.SliderInt("c_KernelSize", &c.KernelSize, 1, 256)
	imgui.SliderInt("c_BlurSize", &c.BlurSize, 1, 8)
	imgui.SliderFloat("c_WhitePoint", &c.WhitePoint, 0.0, 1.0)
	imgui.SliderFloat("c_BlackPoint", &c.BlackPoint, 0.0, 1.0)
}

type ColorGradingUI struct {
	*post.ColorGradingConfig
}

func (c *ColorGradingUI) GetName() string {
	return "Color Grading (LDR)"
}

func (c *ColorGradingUI) ShowUI() {
	imgui.Checkbox("cg_Use", &c.Use)
	imgui.SliderFloat("cg_Contrast", &c.Contrast, 0.8, 10)
	imgui.SliderFloat("cg_Saturation", &c.Saturation, 0, 2)
	imgui.SliderFloat("cg_Brightness", &c.Brightness, 0.5, 10)
	imgui.ColorEdit3("cg_Shadows", &c.ShadowsColor)
	imgui.ColorEdit3("cg_Midtones", &c.MidColor)
	imgui.ColorEdit3("cg_Highlights", &c.HighlightColor)
	imgui.SliderFloat("cg_ColorStrength", &c.ColorStrength, 0, 1.1)
}

type DitheringUI struct {
	*post.DitheringConfig
}

func (c *DitheringUI) GetName() string {
	return "Dithering (LDR)"
}

func (c *DitheringUI) ShowUI() {
	imgui.Checkbox("d_Use", &c.Use)
	imgui.SliderFloat("d_Min", &c.Min, 0.0, 0.99)
	imgui.SliderFloat("d_Max", &c.Max, 0.0, 1.0)
	imgui.SliderFloat("d_Speed", &c.Speed, 0.0, 100.0)
}

type ChromaticAbberationUI struct {
	*post.ChromaticConfig
}

func (c *ChromaticAbberationUI) GetName() string {
	return "Chromatic Abberation (LDR)"
}

func (c *ChromaticAbberationUI) ShowUI() {
	imgui.Checkbox("ca_Use", &c.Use)
	imgui.SliderFloat("ca_Radius", &c.Radius, 0.05, 1.0)
	imgui.SliderFloat("ca_Strength", &c.Strength, 0.01, 0.1)
	imgui.SliderFloat("ca_Power", &c.Power, 0.1, 5.0)
}

type VignetteUI struct {
	*post.VignetteConfig
}

func (c *VignetteUI) GetName() string {
	return "Vignette (LDR)"
}

func (c *VignetteUI) ShowUI() {
	imgui.Checkbox("v_Use", &c.Use)
	imgui.SliderFloat("v_Radius", &c.Radius, 0.2, 2)
	imgui.SliderFloat("v_Softness", &c.Softness, 0.01, 1)
}
