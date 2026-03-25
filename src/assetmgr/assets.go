package assetmgr

import "path/filepath"

const (
	assetsFolder   = "assets"
	modelsFolder   = "models"
	texturesFolder = "textures"
	fontsFolder    = "fonts"
	shadersFolder  = "shaders"
)

func GetModelPath(filename string) string {
	return filepath.Join(assetsFolder, modelsFolder, filename)
}

func GetFontPath(filename string) string {
	return filepath.Join(assetsFolder, fontsFolder, filename)
}

func GetTexturePath(filename string) string {
	return filepath.Join(assetsFolder, texturesFolder, filename)
}

func GetShaderPath(filename string) string {
	return filepath.Join(assetsFolder, shadersFolder, filename)
}
