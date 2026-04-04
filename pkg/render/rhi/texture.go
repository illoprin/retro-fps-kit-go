// Package rhi provides high-level abstractions over OpenGL objects,
// simplifying the development of cross-platform rendering pipelines.
// 
// It implements the Render Hardware Interface (RHI) concept,
// shielding the user from low-level OpenGL state management and
// providing an idiomatic Go API for common graphics tasks.
// 
// Author: illoprin
// 
// 2026

package rhi

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// TextureConfig конфигурация для создания текстуры
type TextureConfig struct {
	Type            TextureType
	Width           int32
	Height          int32
	Depth           int32 // для TextureType2DArray - количество слоёв
	Format          TextureFormat
	FilterMin       TextureFilter
	FilterMag       TextureFilter
	WrapS           TextureWrap
	WrapT           TextureWrap
	WrapR           TextureWrap
	GenerateMipmaps bool
	Anisotropy      float32 // уровень анизотропной фильтрации (0 = выкл)
}

// DefaultTextureConfig возвращает конфигурацию по умолчанию
func DefaultTextureConfig(width, height int32) TextureConfig {
	return TextureConfig{
		Type:            TextureType2D,
		Width:           width,
		Height:          height,
		Depth:           1,
		Format:          FormatRGBA8,
		FilterMin:       FilterLinearMipmapLinear,
		FilterMag:       FilterLinear,
		WrapS:           WrapRepeat,
		WrapT:           WrapRepeat,
		WrapR:           WrapRepeat,
		GenerateMipmaps: true,
		Anisotropy:      0,
	}
}

// Texture обёртка над OpenGL текстурой
type Texture struct {
	ID     uint32
	Type   TextureType
	Config TextureConfig
}

// NewTexture создаёт новую текстуру с заданной конфигурацией
func NewTexture(config TextureConfig) (*Texture, error) {
	var id uint32
	gl.GenTextures(1, &id)
	if id == 0 {
		return nil, fmt.Errorf("failed to generate texture")
	}

	texture := &Texture{
		ID:     id,
		Type:   config.Type,
		Config: config,
	}

	texture.Bind()
	texture.setParams()
	texture.allocateStorage()
	texture.unbind()

	return texture, nil
}

// NewTextureFromImage создаёт текстуру из загруженного изображения
func NewTextureFromImage(path string, generateMipmaps bool, nearest bool) (*Texture, error) {
	imgFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %v", err)
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// create rgba and flip image
	rgba := image.NewRGBA(bounds)
	var x, y int
	// Iterate through pixels and swap top and bottom
	for i := 0; i < height*width; i++ {
		x = i % width
		y = i / height
		// Map the current pixel (x, y) to its vertically flipped position
		rgba.Set(x, height-1-y, img.At(x, y))
	}

	config := DefaultTextureConfig(int32(width), int32(height))
	config.GenerateMipmaps = generateMipmaps
	if generateMipmaps {
		if nearest {
			config.FilterMin = FilterNearestMipmapLinear
			config.FilterMag = FilterNearest
		} else {
			config.FilterMin = FilterLinearMipmapLinear
			config.FilterMag = FilterLinear
		}
	} else {
		if nearest {
			config.FilterMin = FilterNearest
			config.FilterMag = FilterNearest
		} else {
			config.FilterMin = FilterLinear
			config.FilterMag = FilterLinear
		}
	}

	texture, err := NewTexture(config)
	if err != nil {
		return nil, err
	}

	texture.Bind()
	texture.UploadRGBA(0, 0, int32(width), int32(height), rgba.Pix)

	if generateMipmaps {
		texture.GenerateMipmaps()
	}

	texture.unbind()

	return texture, nil
}

// NewFontAtlasTexture создаёт текстуру для шрифтового атласа (1 байт на пиксель)
func NewFontAtlasTexture(width, height int32, data []byte) (*Texture, error) {
	config := DefaultTextureConfig(width, height)
	config.Format = FormatR8
	config.FilterMin = FilterLinear
	config.FilterMag = FilterLinear
	config.GenerateMipmaps = false
	config.WrapS = WrapClampToEdge
	config.WrapT = WrapClampToEdge

	texture, err := NewTexture(config)
	if err != nil {
		return nil, err
	}

	texture.Bind()
	texture.UploadR8(0, 0, width, height, data)
	texture.unbind()

	return texture, nil
}

// NewFramebufferColorTexture создаёт текстуру для использования с Framebuffer
func NewFramebufferColorTexture(width, height int32, format TextureFormat) (*Texture, error) {
	config := DefaultTextureConfig(width, height)
	config.Format = format
	config.FilterMin = FilterNearest
	config.FilterMag = FilterNearest
	config.WrapS = WrapClampToEdge
	config.WrapT = WrapClampToEdge
	config.GenerateMipmaps = false

	return NewTexture(config)
}

// NewFramebufferDepthTexture создаёт текстуру глубины для Framebuffer
func NewFramebufferDepthTexture(width, height int32, format TextureFormat) (*Texture, error) {
	config := DefaultTextureConfig(width, height)
	config.Format = format
	config.FilterMin = FilterNearest
	config.FilterMag = FilterNearest
	config.WrapS = WrapClampToEdge
	config.WrapT = WrapClampToEdge
	config.GenerateMipmaps = false

	return NewTexture(config)
}

// NewCubeMap создаёт Cubemap текстуру
func NewCubeMap(size int32, generateMipmaps bool) (*Texture, error) {
	config := TextureConfig{
		Type:            TextureTypeCubeMap,
		Width:           size,
		Height:          size,
		Depth:           1,
		Format:          FormatRGBA8,
		FilterMin:       FilterLinearMipmapLinear,
		FilterMag:       FilterLinear,
		WrapS:           WrapClampToEdge,
		WrapT:           WrapClampToEdge,
		WrapR:           WrapClampToEdge,
		GenerateMipmaps: generateMipmaps,
		Anisotropy:      0,
	}

	return NewTexture(config)
}

// NewCubeMapFromImages создаёт Cubemap из 6 изображений (в порядке: right, left, top, bottom, front, back)
func NewCubeMapFromImages(paths [6]string, generateMipmaps bool) (*Texture, error) {
	// Загружаем первое изображение для определения размера
	imgFile, err := os.Open(paths[0])
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %v", err)
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	bounds := img.Bounds()
	size := int32(bounds.Dx())
	if bounds.Dy() != int(size) {
		return nil, fmt.Errorf("cubemap images must be square")
	}

	texture, err := NewCubeMap(size, generateMipmaps)
	if err != nil {
		return nil, err
	}

	texture.Bind()

	// Загружаем все 6 граней
	cubeMapTargets := []uint32{
		gl.TEXTURE_CUBE_MAP_POSITIVE_X, // right
		gl.TEXTURE_CUBE_MAP_NEGATIVE_X, // left
		gl.TEXTURE_CUBE_MAP_POSITIVE_Y, // top
		gl.TEXTURE_CUBE_MAP_NEGATIVE_Y, // bottom
		gl.TEXTURE_CUBE_MAP_POSITIVE_Z, // front
		gl.TEXTURE_CUBE_MAP_NEGATIVE_Z, // back
	}

	for i, path := range paths {
		imgFile, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open cubemap image: %v", err)
		}

		img, _, err := image.Decode(imgFile)
		imgFile.Close()

		if err != nil {
			return nil, fmt.Errorf("failed to decode cubemap image: %v", err)
		}

		bounds := img.Bounds()
		if bounds.Dx() != int(size) || bounds.Dy() != int(size) {
			return nil, fmt.Errorf("all cubemap images must have same size")
		}

		rgba := image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

		gl.TexImage2D(cubeMapTargets[i], 0, gl.RGBA, size, size, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))
	}

	if generateMipmaps {
		texture.GenerateMipmaps()
	}

	texture.unbind()

	return texture, nil
}

// bind привязывает текстуру
func (t *Texture) Bind() {
	switch t.Type {
	case TextureType2D:
		gl.BindTexture(gl.TEXTURE_2D, t.ID)
	case TextureTypeCubeMap:
		gl.BindTexture(gl.TEXTURE_CUBE_MAP, t.ID)
	case TextureType2DArray:
		gl.BindTexture(gl.TEXTURE_2D_ARRAY, t.ID)
	}
}

// unbind отвязывает текстуру
func (t *Texture) unbind() {
	switch t.Type {
	case TextureType2D:
		gl.BindTexture(gl.TEXTURE_2D, 0)
	case TextureTypeCubeMap:
		gl.BindTexture(gl.TEXTURE_CUBE_MAP, 0)
	case TextureType2DArray:
		gl.BindTexture(gl.TEXTURE_2D_ARRAY, 0)
	}
}

// Bind привязывает текстуру к указанному юниту
func (t *Texture) BindToSlot(unit uint32) {
	gl.ActiveTexture(gl.TEXTURE0 + unit)
	t.Bind()
}

// Unbind отвязывает текстуру от указанного юнита
func (t *Texture) Unbind(unit uint32) {
	gl.ActiveTexture(gl.TEXTURE0 + unit)
	t.unbind()
}

// Delete удаляет текстуру
func (t *Texture) Delete() {
	if t.ID > 0 {
		gl.DeleteTextures(1, &t.ID)
	}
}

// setParams устанавливает параметры текстуры
func (t *Texture) setParams() {
	target := GetTextureType(t.Type)

	// Установка фильтрации
	minFilter := GetMinFilter(t.Config.FilterMin)
	magFilter := GetFilter(t.Config.FilterMag)
	gl.TexParameteri(target, gl.TEXTURE_MIN_FILTER, minFilter)
	gl.TexParameteri(target, gl.TEXTURE_MAG_FILTER, magFilter)

	// Установка оборачивания
	wrapS := GetWrapMode(t.Config.WrapS)
	wrapT := GetWrapMode(t.Config.WrapT)
	gl.TexParameteri(target, gl.TEXTURE_WRAP_S, wrapS)
	gl.TexParameteri(target, gl.TEXTURE_WRAP_T, wrapT)

	if t.Type == TextureTypeCubeMap {
		wrapR := GetWrapMode(t.Config.WrapR)
		gl.TexParameteri(target, gl.TEXTURE_WRAP_R, wrapR)
	}

	if t.Config.Format == FormatDepth16 ||
		t.Config.Format == FormatDepth24 ||
		t.Config.Format == FormatDepth24Stencil8 ||
		t.Config.Format == FormatDepth32F {
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_COMPARE_MODE, gl.NONE)
	}

	// Анизотропная фильтрация (если доступна)
	if t.Config.Anisotropy > 0 {
		var maxAnisotropy float32
		gl.GetFloatv(gl.MAX_TEXTURE_MAX_ANISOTROPY, &maxAnisotropy)
		anisotropy := t.Config.Anisotropy
		if anisotropy > maxAnisotropy {
			anisotropy = maxAnisotropy
		}
		gl.TexParameterf(target, gl.TEXTURE_MAX_ANISOTROPY, anisotropy)
	}
}

// allocateStorage выделяет память для текстуры
func (t *Texture) allocateStorage() {
	target := GetTextureType(t.Type)
	internalFormat := GetInternalFormat(t.Config.Format)
	format := GetFormat(t.Config.Format)
	dataType := GetDataType(t.Config.Format)

	switch t.Type {
	case TextureType2D:
		gl.TexImage2D(target, 0, internalFormat, t.Config.Width, t.Config.Height, 0, format, dataType, nil)

	case TextureTypeCubeMap:
		for i := 0; i < 6; i++ {
			gl.TexImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_X+uint32(i), 0, internalFormat,
				t.Config.Width, t.Config.Height, 0, format, dataType, nil)
		}

	case TextureType2DArray:
		gl.TexImage3D(target, 0, internalFormat, t.Config.Width, t.Config.Height, t.Config.Depth,
			0, format, dataType, nil)
	}
}

// UploadRGBA загружает RGBA данные в текстуру
func (t *Texture) UploadRGBA(x, y, width, height int32, data []byte) {
	target := GetTextureType(t.Type)
	gl.TexSubImage2D(target, 0, x, y, width, height, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data))
}

// UploadRGBA загружает RGB данные в текстуру
func (t *Texture) UploadRGB(x, y, width, height int32, data []byte) {
	target := GetTextureType(t.Type)
	gl.TexSubImage2D(target, 0, x, y, width, height, gl.RGB, gl.UNSIGNED_BYTE, gl.Ptr(data))
}

// UploadR8 загружает одноканальные данные (8 бит) в текстуру
func (t *Texture) UploadR8(x, y, width, height int32, data []byte) {
	target := GetTextureType(t.Type)
	gl.TexSubImage2D(target, 0, x, y, width, height, gl.RED, gl.UNSIGNED_BYTE, gl.Ptr(data))
}

// UploadDepth загружает данные глубины
func (t *Texture) UploadDepth(x, y, width, height int32, data []float32) {
	target := GetTextureType(t.Type)
	gl.TexSubImage2D(target, 0, x, y, width, height, gl.DEPTH_COMPONENT, gl.FLOAT, gl.Ptr(data))
}

// UploadLayer загружает данные в конкретный слой для Texture2DArray
func (t *Texture) UploadLayer(layer, x, y, width, height int32, data []byte) {
	if t.Type != TextureType2DArray {
		panic("UploadLayer only works with TextureType2DArray")
	}
	gl.TexSubImage3D(gl.TEXTURE_2D_ARRAY, 0, x, y, layer, width, height, 1,
		GetFormat(t.Config.Format), GetDataType(t.Config.Format), gl.Ptr(data))
}

// GenerateMipmaps генерирует мип-карты
func (t *Texture) GenerateMipmaps() {
	target := GetTextureType(t.Type)
	gl.GenerateMipmap(target)
}

// GetWidth возвращает размер текстуры
func (t *Texture) GetSize() (int32, int32) {
	return t.Config.Width, t.Config.Height
}

// Resize изменяет размер текстуры (только для Framebuffer текстур)
func (t *Texture) Resize(width, height int32) {
	if t.Config.Width == width && t.Config.Height == height {
		return
	}

	t.Config.Width = width
	t.Config.Height = height

	t.Bind()
	t.allocateStorage()
	t.unbind()
}
