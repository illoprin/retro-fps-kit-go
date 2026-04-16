package post

import (
	"math"
	"math/rand"
	"unsafe"

	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/rhi"
)

const (
	noiseSize = int32(6)
)

// noise tex (random rotations for hemi-sphere)
func CreateNoiseTexture() *rhi.Texture {
	noiseSliceLen := noiseSize * noiseSize * 3
	noiseRaw := make([]byte, noiseSliceLen)
	for i := 0; i < int(noiseSliceLen/3); i++ {
		pix := (rand.Float32()*2 - 1) * math.Pi * .5
		noiseRaw[i*3] = byte(pix * 255.0)
		noiseRaw[i*3+1] = byte(pix * 255.0)
		noiseRaw[i*3+2] = 0
	}
	noiseTextureConfig := rhi.TextureConfig{
		Type:      rhi.TextureType2D,
		Width:     noiseSize,
		Height:    noiseSize,
		Format:    rhi.FormatRGB8,
		FilterMin: rhi.FilterNearest,
		FilterMag: rhi.FilterNearest,
		WrapS:     rhi.WrapRepeat,
		WrapT:     rhi.WrapRepeat,
	}
	noiseTexture := rhi.NewTexture(noiseTextureConfig)
	noiseTexture.Upload2D(0, 0, unsafe.Pointer(&noiseRaw[0]))
	return noiseTexture
}
