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

import "github.com/go-gl/gl/v4.1-core/gl"

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

type DrawType uint32

const (
	StaticDraw DrawType = iota
	DynamicDraw
	StreamDraw
)

const (
	sizeOfFloat uint32 = 4
)

type RenderStats struct {
	DrawCalls uint64
	Vertices  uint64
	Triangles uint64
}

func (r *RenderStats) Reset() {
	r.DrawCalls = 0
	r.Vertices = 0
	r.Triangles = 0
}

var FrameStats RenderStats

// GetInternalFormat - returns GPU store type
func GetInternalFormat(format TextureFormat) int32 {
	switch format {
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

// GetTextureType - get internal texture type
func GetTextureType(t TextureType) uint32 {
	switch t {
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

// GetFormat - returns texture format
func GetFormat(format TextureFormat) uint32 {
	switch format {
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

// GetDataType - returns source data format
func GetDataType(format TextureFormat) uint32 {
	switch format {
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

// GetMinFilter - returns min filter for texturing
func GetMinFilter(filter TextureFilter) int32 {
	switch filter {
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

// GetMagFilter - returns mag filter for texturing
func GetFilter(filter TextureFilter) int32 {
	switch filter {
	case FilterNearest:
		return gl.NEAREST
	case FilterLinear:
		return gl.LINEAR
	default:
		return gl.LINEAR
	}
}

// GetWrapMode - returns texture wrap mode
func GetWrapMode(mode TextureWrap) int32 {
	switch mode {
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
