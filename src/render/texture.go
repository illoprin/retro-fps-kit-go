package render

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// TextureType определяет тип текстуры
type TextureType int

const (
	TextureType2D TextureType = iota
	TextureTypeCubeMap
	TextureType2DArray
)

// TextureFilter определяет режим фильтрации
type TextureFilter int

const (
	FilterNearest TextureFilter = iota
	FilterLinear
	FilterNearestMipmapNearest
	FilterNearestMipmapLinear
	FilterLinearMipmapNearest
	FilterLinearMipmapLinear
)

// TextureWrap определяет режим оборачивания
type TextureWrap int

const (
	WrapRepeat TextureWrap = iota
	WrapMirroredRepeat
	WrapClampToEdge
	WrapClampToBorder
)

// TextureFormat определяет внутренний формат текстуры
type TextureFormat int

const (
	// Форматы для цветных текстур
	FormatRGBA8 TextureFormat = iota
	FormatRGBA16F
	FormatRGBA32F
	FormatRGB8
	FormatRGB16F
	FormatRGB32F

	// Форматы для одноканальных текстур
	FormatR8
	FormatR16F
	FormatR32F

	// Форматы для глубины
	FormatDepth16
	FormatDepth24
	FormatDepth32F
	FormatDepth24Stencil8
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

	texture.bind()
	texture.setParams()
	texture.allocateStorage()
	texture.unbind()

	return texture, nil
}

// NewTextureFromImage создаёт текстуру из загруженного изображения
func NewTextureFromImage(path string, generateMipmaps bool) (*Texture, error) {
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
	config.FilterMin = FilterLinearMipmapLinear
	config.FilterMag = FilterLinear

	texture, err := NewTexture(config)
	if err != nil {
		return nil, err
	}

	texture.bind()
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

	texture.bind()
	texture.UploadR8(0, 0, width, height, data)
	texture.unbind()

	return texture, nil
}

// NewFramebufferColorTexture создаёт текстуру для использования с Framebuffer
func NewFramebufferColorTexture(width, height int32, format TextureFormat, filter TextureFilter) (*Texture, error) {
	config := DefaultTextureConfig(width, height)
	config.Format = format
	config.FilterMin = FilterLinear
	config.FilterMag = FilterLinear
	config.WrapS = WrapClampToEdge
	config.WrapT = WrapClampToEdge
	config.GenerateMipmaps = false

	return NewTexture(config)
}

// NewFramebufferDepthTexture создаёт текстуру глубины для Framebuffer
func NewFramebufferDepthTexture(width, height int32) (*Texture, error) {
	config := DefaultTextureConfig(width, height)
	config.Format = FormatDepth32F
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

	texture.bind()

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
func (t *Texture) bind() {
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
func (t *Texture) Bind(unit uint32) {
	gl.ActiveTexture(gl.TEXTURE0 + unit)
	t.bind()
}

// Unbind отвязывает текстуру от указанного юнита
func (t *Texture) Unbind(unit uint32) {
	gl.ActiveTexture(gl.TEXTURE0 + unit)
	t.unbind()
}

// Delete удаляет текстуру
func (t *Texture) Delete() {
	gl.DeleteTextures(1, &t.ID)
}

// setParams устанавливает параметры текстуры
func (t *Texture) setParams() {
	target := t.getTarget()

	// Установка фильтрации
	minFilter := t.getMinFilter()
	magFilter := t.getMagFilter()
	gl.TexParameteri(target, gl.TEXTURE_MIN_FILTER, minFilter)
	gl.TexParameteri(target, gl.TEXTURE_MAG_FILTER, magFilter)

	// Установка оборачивания
	wrapS := t.getWrapMode(t.Config.WrapS)
	wrapT := t.getWrapMode(t.Config.WrapT)
	gl.TexParameteri(target, gl.TEXTURE_WRAP_S, wrapS)
	gl.TexParameteri(target, gl.TEXTURE_WRAP_T, wrapT)

	if t.Type == TextureTypeCubeMap {
		wrapR := t.getWrapMode(t.Config.WrapR)
		gl.TexParameteri(target, gl.TEXTURE_WRAP_R, wrapR)
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
	target := t.getTarget()
	internalFormat := t.getInternalFormat()

	switch t.Type {
	case TextureType2D:
		gl.TexImage2D(target, 0, internalFormat, t.Config.Width, t.Config.Height, 0, t.getFormat(), t.getDataType(), nil)

	case TextureTypeCubeMap:
		for i := 0; i < 6; i++ {
			gl.TexImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_X+uint32(i), 0, internalFormat,
				t.Config.Width, t.Config.Height, 0, t.getFormat(), t.getDataType(), nil)
		}

	case TextureType2DArray:
		gl.TexImage3D(target, 0, internalFormat, t.Config.Width, t.Config.Height, t.Config.Depth,
			0, t.getFormat(), t.getDataType(), nil)
	}
}

// UploadRGBA загружает RGBA данные в текстуру
func (t *Texture) UploadRGBA(x, y, width, height int32, data []byte) {
	target := t.getTarget()
	gl.TexSubImage2D(target, 0, x, y, width, height, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data))
}

// UploadR8 загружает одноканальные данные (8 бит) в текстуру
func (t *Texture) UploadR8(x, y, width, height int32, data []byte) {
	target := t.getTarget()
	gl.TexSubImage2D(target, 0, x, y, width, height, gl.RED, gl.UNSIGNED_BYTE, gl.Ptr(data))
}

// UploadDepth загружает данные глубины
func (t *Texture) UploadDepth(x, y, width, height int32, data []float32) {
	target := t.getTarget()
	gl.TexSubImage2D(target, 0, x, y, width, height, gl.DEPTH_COMPONENT, gl.FLOAT, gl.Ptr(data))
}

// UploadLayer загружает данные в конкретный слой для Texture2DArray
func (t *Texture) UploadLayer(layer, x, y, width, height int32, data []byte) {
	if t.Type != TextureType2DArray {
		panic("UploadLayer only works with TextureType2DArray")
	}
	gl.TexSubImage3D(gl.TEXTURE_2D_ARRAY, 0, x, y, layer, width, height, 1,
		t.getFormat(), t.getDataType(), gl.Ptr(data))
}

// GenerateMipmaps генерирует мип-карты
func (t *Texture) GenerateMipmaps() {
	target := t.getTarget()
	gl.GenerateMipmap(target)
}

// GetID возвращает ID текстуры
func (t *Texture) GetID() uint32 {
	return t.ID
}

// GetWidth возвращает ширину текстуры
func (t *Texture) GetWidth() int32 {
	return t.Config.Width
}

// GetHeight возвращает высоту текстуры
func (t *Texture) GetHeight() int32 {
	return t.Config.Height
}

// Resize изменяет размер текстуры (только для Framebuffer текстур)
func (t *Texture) Resize(width, height int32) {
	if t.Config.Width == width && t.Config.Height == height {
		return
	}

	t.Config.Width = width
	t.Config.Height = height

	t.bind()
	t.allocateStorage()
	t.unbind()
}

// Вспомогательные методы
func (t *Texture) getTarget() uint32 {
	switch t.Type {
	case TextureType2D:
		return gl.TEXTURE_2D
	case TextureTypeCubeMap:
		return gl.TEXTURE_CUBE_MAP
	case TextureType2DArray:
		return gl.TEXTURE_2D_ARRAY
	default:
		return gl.TEXTURE_2D
	}
}

func (t *Texture) getInternalFormat() int32 {
	switch t.Config.Format {
	case FormatRGBA8:
		return gl.RGBA8
	case FormatRGBA16F:
		return gl.RGBA16F
	case FormatRGBA32F:
		return gl.RGBA32F
	case FormatRGB8:
		return gl.RGB8
	case FormatRGB16F:
		return gl.RGB16F
	case FormatRGB32F:
		return gl.RGB32F
	case FormatR8:
		return gl.R8
	case FormatR16F:
		return gl.R16F
	case FormatR32F:
		return gl.R32F
	case FormatDepth16:
		return gl.DEPTH_COMPONENT16
	case FormatDepth24:
		return gl.DEPTH_COMPONENT24
	case FormatDepth32F:
		return gl.DEPTH_COMPONENT32F
	case FormatDepth24Stencil8:
		return gl.DEPTH24_STENCIL8
	default:
		return gl.RGBA8
	}
}

func (t *Texture) getFormat() uint32 {
	switch t.Config.Format {
	case FormatRGBA8, FormatRGBA16F, FormatRGBA32F:
		return gl.RGBA
	case FormatRGB8, FormatRGB16F, FormatRGB32F:
		return gl.RGB
	case FormatR8, FormatR16F, FormatR32F:
		return gl.RED
	case FormatDepth16, FormatDepth24, FormatDepth32F:
		return gl.DEPTH_COMPONENT
	case FormatDepth24Stencil8:
		return gl.DEPTH_STENCIL
	default:
		return gl.RGBA
	}
}

func (t *Texture) getDataType() uint32 {
	switch t.Config.Format {
	case FormatRGBA8, FormatRGB8, FormatR8:
		return gl.UNSIGNED_BYTE
	case FormatRGBA16F, FormatRGB16F, FormatR16F:
		return gl.HALF_FLOAT
	case FormatRGBA32F, FormatRGB32F, FormatR32F, FormatDepth32F:
		return gl.FLOAT
	case FormatDepth16:
		return gl.UNSIGNED_SHORT
	case FormatDepth24:
		return gl.UNSIGNED_INT
	case FormatDepth24Stencil8:
		return gl.UNSIGNED_INT_24_8
	default:
		return gl.UNSIGNED_BYTE
	}
}

func (t *Texture) getMinFilter() int32 {
	switch t.Config.FilterMin {
	case FilterNearest:
		return gl.NEAREST
	case FilterLinear:
		return gl.LINEAR
	case FilterNearestMipmapNearest:
		return gl.NEAREST_MIPMAP_NEAREST
	case FilterNearestMipmapLinear:
		return gl.NEAREST_MIPMAP_LINEAR
	case FilterLinearMipmapNearest:
		return gl.LINEAR_MIPMAP_NEAREST
	case FilterLinearMipmapLinear:
		return gl.LINEAR_MIPMAP_LINEAR
	default:
		return gl.LINEAR_MIPMAP_LINEAR
	}
}

func (t *Texture) getMagFilter() int32 {
	switch t.Config.FilterMag {
	case FilterNearest:
		return gl.NEAREST
	case FilterLinear:
		return gl.LINEAR
	default:
		return gl.LINEAR
	}
}

func (t *Texture) getWrapMode(wrap TextureWrap) int32 {
	switch wrap {
	case WrapRepeat:
		return gl.REPEAT
	case WrapMirroredRepeat:
		return gl.MIRRORED_REPEAT
	case WrapClampToEdge:
		return gl.CLAMP_TO_EDGE
	case WrapClampToBorder:
		return gl.CLAMP_TO_BORDER
	default:
		return gl.REPEAT
	}
}
