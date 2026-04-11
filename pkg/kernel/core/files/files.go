package files

import "path/filepath"

const (
	assetsFolder   = "assets"
	modelsFolder   = "models"
	texturesFolder = "textures"
	fontsFolder    = "fonts"
	shadersFolder  = "shaders"
	levelsFolder   = "levels"
)

func GetLevelTexturePath(levelName string, filename string) string {
	return filepath.Join(GetLevelPath(levelName), texturesFolder, filename)
}

func GetLevelModelPath(levelName string, filename string) string {
	return filepath.Join(GetLevelPath(levelName), modelsFolder, filename)
}

func GetLevelPath(levelName string) string {
	return filepath.Join(assetsFolder, levelsFolder, levelName)
}

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
