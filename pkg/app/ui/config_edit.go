package ui

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/passes"
)

type ConfigUI interface {
	GetName() string
	ShowUI()
}

type EyeAdaptionConfigUI struct {
	*passes.EyeAdaptionConfig
}

func (c *EyeAdaptionConfigUI) GetName() string {
	return "Eye Adaption"
}

func (c *EyeAdaptionConfigUI) ShowUI() {
	imgui.Checkbox("ea_Use", &c.Use)
	imgui.SliderFloat("ea_Radius", &c.Radius, 4, 100)
	imgui.SliderFloat("ea_Speed", &c.AdaptionSpeed, 0.005, 1)
	imgui.SliderFloat("ea_Gray", &c.AvgGray, 0.01, 0.3)
}

type SSAOConfigUI struct {
	*passes.SSAOConfig
}

func (c *SSAOConfigUI) GetName() string {
	return "SSAO"
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
	*passes.CavityConfig
}

func (c *CavityConfigUI) GetName() string {
	return "Cavity Occlusion"
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
	*passes.ColorGradingConfig
}

func (c *ColorGradingUI) GetName() string {
	return "Color Grading"
}

func (c *ColorGradingUI) ShowUI() {
	imgui.Checkbox("cg_Use", &c.Use)
	imgui.SliderFloat("cg_Gamma", &c.Gamma, 1, 3)
	imgui.SliderFloat("cg_Exposure", &c.Exposure, 0.5, 4)
	imgui.SliderFloat("cg_Contrast", &c.Contrast, 0.8, 3)
	imgui.SliderFloat("cg_Saturation", &c.Saturation, 0, 2)
	imgui.SliderFloat("cg_Brightness", &c.Brightness, 0.5, 10)
	imgui.ColorEdit3("cg_Shadows", &c.ShadowsColor)
	imgui.ColorEdit3("cg_Midtones", &c.MidColor)
	imgui.ColorEdit3("cg_Highlights", &c.HighlightColor)
	imgui.SliderFloat("cg_ColorStrength", &c.ColorStrength, 0, 1.1)
}

type VignetteUI struct {
	*passes.VignetteConfig
}

func (c *VignetteUI) GetName() string {
	return "Vignette"
}

func (c *VignetteUI) ShowUI() {
	imgui.Checkbox("v_Use", &c.Use)
	imgui.SliderFloat("v_Radius", &c.Radius, 0.2, 2)
	imgui.SliderFloat("v_Softness", &c.Softness, 0.01, 1)
}
