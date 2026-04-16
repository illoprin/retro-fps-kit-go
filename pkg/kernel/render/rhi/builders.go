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
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/files"
)

// SetupBasicQuadMesh - setups buffers for basic quad mesh
// 1 attribute - (location = 0) vec2 in_position
func SetupBasicQuadMesh(m *Mesh) {
	var (
		basicQuadVertices = []float32{
			-1, -1,
			1, -1,
			1, 1,
			-1, 1,
		}
		basicQuadIndices = []uint32{
			0, 1, 2,
			2, 3, 0,
		}
	)
	sizeVerts := len(basicQuadVertices) * int(unsafe.Sizeof(basicQuadVertices[0]))
	sizeIndices := len(basicQuadIndices) * int(unsafe.Sizeof(basicQuadIndices[0]))

	// build layout
	m.Bind()
	m.CreateVertexBuffer()
	m.CreateElementBuffer()
	m.SetAttribute(0, VertexAttribute{
		Index:       0,
		Components:  2,
		Type:        Float32,
		StrideBytes: 2 * SizeOfFloat32,
		OffsetBytes: 0,
		Divisor:     0,
	})
	m.Unbind()

	// allocate buffers
	m.AllocateVertexBuffer(0, sizeVerts, StaticDraw)
	m.SetVertexBufferData(0, 0, sizeVerts, unsafe.Pointer(&basicQuadVertices[0]))
	m.AllocateElementBuffer(sizeIndices, StaticDraw)
	m.SetElementBufferData(0, basicQuadIndices)
}

// DefaultTexture2DConfig - returns default texture config
func DefaultTexture2DConfig(width, height int32) TextureConfig {
	return TextureConfig{
		Type:            TextureType2D,
		Width:           width,
		Height:          height,
		Depth:           1,
		Format:          FormatRGBA8,
		FilterMin:       FilterLinearMipmapLinear,
		FilterMag:       FilterNearest,
		WrapS:           WrapRepeat,
		WrapT:           WrapRepeat,
		WrapR:           WrapRepeat,
		GenerateMipmaps: true,
		LodBias:         0,
	}
}

// NewTextureFromImage creates texture from image
// width & height will be re-written
func NewTextureFromImage(path string, config TextureConfig) (*Texture, error) {
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

	config.Width = int32(width)
	config.Height = int32(height)

	t := NewTexture(config)
	t.Bind()
	t.Upload2D(0, 0, unsafe.Pointer(&rgba.Pix[0]))

	if config.GenerateMipmaps {
		t.GenerateMipmaps()
	}

	t.unbind()

	return t, nil
}

// NewTextureArray builds RGBA8 Nearest texture array
// and mipmaps for that
func NewTextureArray(imgs []*files.RGBA8Data) (*Texture, error) {

	if len(imgs) <= 0 {
		return nil, fmt.Errorf("empty texture set")
	}

	tex := NewTexture(TextureConfig{
		Type:            TextureType2DArray,
		Width:           imgs[0].W,
		Height:          imgs[0].H,
		Depth:           int32(len(imgs)),
		Format:          FormatRGBA8,
		FilterMin:       FilterLinearMipmapLinear,
		FilterMag:       FilterNearest,
		WrapS:           WrapRepeat,
		WrapT:           WrapRepeat,
		GenerateMipmaps: true,
	})

	tex.Bind()

	for i, img := range imgs {
		tex.UploadLayer(0, 0, int32(i), unsafe.Pointer(&img.Data[0]))
	}

	tex.GenerateMipmaps()

	return tex, nil
}

// NewFramebufferColorTexture создаёт текстуру для использования с Framebuffer
func NewFramebufferColorTexture(
	width, height int32,
	format TextureFormat, filter TextureFilter,
) *Texture {
	config := DefaultTexture2DConfig(width, height)
	config.Format = format
	config.FilterMin = filter
	config.FilterMag = filter
	config.WrapS = WrapClampToEdge
	config.WrapT = WrapClampToEdge
	config.GenerateMipmaps = false

	return NewTexture(config)
}

// NewFramebufferDepthTexture создаёт текстуру глубины для Framebuffer
func NewFramebufferDepthTexture(width, height int32, format TextureFormat) *Texture {
	config := DefaultTexture2DConfig(width, height)
	config.Format = format
	config.FilterMin = FilterNearest
	config.FilterMag = FilterNearest
	config.WrapS = WrapClampToEdge
	config.WrapT = WrapClampToEdge
	config.GenerateMipmaps = false

	return NewTexture(config)
}

// NewCubeMap создаёт Cubemap текстуру
func NewCubeMap(size int32, generateMipmaps bool) *Texture {
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
	}

	return NewTexture(config)
}

// NewCubeMapFromImages creates cubemap from 6 images (order: right, left, top, bottom, front, back)
func NewCubeMapFromImages(paths [6]string, generateMipmaps bool) (*Texture, error) {
	// load first image to determine size
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

	texture := NewCubeMap(size, generateMipmaps)

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
