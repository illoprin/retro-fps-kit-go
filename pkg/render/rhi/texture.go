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
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// TextureConfig config for texture create
type TextureConfig struct {
	Type            TextureType
	Width           int32
	Height          int32
	Depth           int32 // for TextureType2DArray
	Format          TextureFormat
	FilterMin       TextureFilter
	FilterMag       TextureFilter
	WrapS           TextureWrap
	WrapT           TextureWrap
	WrapR           TextureWrap
	GenerateMipmaps bool
	LodBias         float32
	Anisotropy      float32 // уровень анизотропной фильтрации (0 = выкл)
}

// Texture - wrapper over an OpenGL object
type Texture struct {
	ID     uint32
	Config TextureConfig
}

// NewTexture - creates new texture from config
func NewTexture(config TextureConfig) *Texture {
	var id uint32
	gl.GenTextures(1, &id)
	texture := &Texture{
		ID:     id,
		Config: config,
	}
	texture.Bind()
	texture.setParams()
	texture.allocateStorage()

	return texture
}

func (t *Texture) Bind() {
	switch t.Config.Type {
	case TextureType2D:
		gl.BindTexture(gl.TEXTURE_2D, t.ID)
	case TextureTypeCubeMap:
		gl.BindTexture(gl.TEXTURE_CUBE_MAP, t.ID)
	case TextureType2DArray:
		gl.BindTexture(gl.TEXTURE_2D_ARRAY, t.ID)
	}
}

func (t *Texture) unbind() {
	switch t.Config.Type {
	case TextureType2D:
		gl.BindTexture(gl.TEXTURE_2D, 0)
	case TextureTypeCubeMap:
		gl.BindTexture(gl.TEXTURE_CUBE_MAP, 0)
	case TextureType2DArray:
		gl.BindTexture(gl.TEXTURE_2D_ARRAY, 0)
	}
}

// BindToUnit - binds texture to the specified unit
func (t *Texture) BindToUnit(unit uint32) {
	gl.ActiveTexture(gl.TEXTURE0 + unit)
	t.Bind()
}

// Delete - deletes texture
func (t *Texture) Delete() {
	if t.ID > 0 {
		gl.DeleteTextures(1, &t.ID)
	}
}

// setParams - sets texture params
func (t *Texture) setParams() {
	target := GetTextureType(t.Config.Type)

	// -- filtering
	gl.TexParameteri(target, gl.TEXTURE_MIN_FILTER, GetMinFilter(t.Config.FilterMin))
	gl.TexParameteri(target, gl.TEXTURE_MAG_FILTER, GetFilter(t.Config.FilterMag))

	// -- wrapping
	gl.TexParameteri(target, gl.TEXTURE_WRAP_S, GetWrapMode(t.Config.WrapS))
	gl.TexParameteri(target, gl.TEXTURE_WRAP_T, GetWrapMode(t.Config.WrapT))

	// -- if cube map
	if t.Config.Type == TextureTypeCubeMap {
		wrapR := GetWrapMode(t.Config.WrapR)
		gl.TexParameteri(target, gl.TEXTURE_WRAP_R, wrapR)
	}

	// -- depth params
	if t.Config.Format == FormatDepth16 ||
		t.Config.Format == FormatDepth24 ||
		t.Config.Format == FormatDepth24Stencil8 ||
		t.Config.Format == FormatDepth32F {
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_COMPARE_MODE, gl.NONE)
	}

	// -- lod bias
	if t.Config.LodBias != 0 {
		gl.TexParameterf(target, gl.TEXTURE_LOD_BIAS, t.Config.LodBias)
	}

	// -- anisotropic filtering
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

func (t *Texture) GetPixels(format TextureFormat, mipMapLevel int32, ptr unsafe.Pointer) {
	target := GetTextureType(t.Config.Type)
	dataType := GetTextureDataType(format)
	channelData := GetFormat(format)
	t.Bind()
	gl.GetTexImage(target, mipMapLevel, channelData, dataType, ptr)
}

// allocateStorage - allocates memory
func (t *Texture) allocateStorage() {
	target := GetTextureType(t.Config.Type)
	internalFormat := GetInternalFormat(t.Config.Format)
	format := GetFormat(t.Config.Format)
	dataType := GetTextureDataType(t.Config.Format)

	switch t.Config.Type {
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

// Upload2D - uploads data to 2D texture
func (t *Texture) Upload2D(x, y int32, data unsafe.Pointer) {

	if t.Config.Type != TextureType2D {
		panic("Upload2D only works with TextureType2D")
	}

	target := GetTextureType(t.Config.Type)
	format := GetFormat(t.Config.Format)
	dataType := GetTextureDataType(t.Config.Format)

	gl.TexSubImage2D(target, 0, x, y, t.Config.Width, t.Config.Height, format, dataType, data)
}

// UploadLayer - uploads layer to texture array
func (t *Texture) UploadLayer(x, y, layer int32, data unsafe.Pointer) {

	if t.Config.Type != TextureType2DArray {
		panic("UploadLayer only works with TextureType2DArray")
	}

	target := GetTextureType(t.Config.Type)
	format := GetFormat(t.Config.Format)
	dataType := GetTextureDataType(t.Config.Format)

	gl.TexSubImage3D(target, 0, x, y, layer, t.Config.Width, t.Config.Height, 1, format, dataType, data)
}

// UploadFace - upload cube map face
func (t *Texture) UploadFace(face uint32, data unsafe.Pointer) {
	if t.Config.Type != TextureTypeCubeMap {
		panic("UploadFace only works with CubeMap")
	}
	gl.TexSubImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_X+face, 0, 0, 0, t.Config.Width, t.Config.Height,
		GetFormat(t.Config.Format), GetTextureDataType(t.Config.Format), data)
}

// GenerateMipmaps generates mip maps
func (t *Texture) GenerateMipmaps() {
	target := GetTextureType(t.Config.Type)
	t.Bind()
	gl.GenerateMipmap(target)
}

// GetWidth returns texture size
func (t *Texture) GetSize() (int32, int32) {
	return t.Config.Width, t.Config.Height
}

// UnbindFromUnit - unbinds texture from unit
func (t *Texture) UnbindFromUnit(unit uint32) {
	gl.ActiveTexture(gl.TEXTURE0 + unit)
	t.unbind()
}

// Resize - reallocates memory
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
