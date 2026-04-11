package config

import (
	"os"

	yaml "github.com/goccy/go-yaml"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/post"
)

type Config struct {
	Window         DisplayConfig        `yaml:"display"`
	PostProcessing PostProcessingConfig `yaml:"post_processing"`
}

type DisplayConfig struct {
	Width  int     `yaml:"width"`
	Height int     `yaml:"height"`
	Ratio  float32 `yaml:"resolution_ratio"`
}

type PostProcessingConfig struct {
	ColorGrading *post.ColorGradingConfig `yaml:"color_grading"`
	EyeAdaption  *post.EyeAdaptionConfig  `yaml:"eye_adaption"`
	SSAO         *post.SSAOConfig         `yaml:"ssao"`
	Cavity       *post.CavityConfig       `yaml:"cavity"`
	Bloom        *post.BloomConfig        `yaml:"bloom"`
	Tonemapping  *post.ToneMappingConfig  `yaml:"tonemapping"`
	Vignette     *post.VignetteConfig     `yaml:"vignette"`
}

// LoadConfig creates new config object from yaml
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// SaveConfig saves config object to yaml by path
func SaveConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// validateConfig
func validateConfig(c *Config) error {
	// TODO
	return nil
}
