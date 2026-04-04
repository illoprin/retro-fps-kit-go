package passes

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
)

const (
	noiseSize = int32(6)
)

// noise tex (random rotations for hemi-sphere)
func CreateNoiseTexture() (*rhi.Texture, error) {
	noiseSliceLen := noiseSize * noiseSize * 3
	noiseRaw := make([]byte, noiseSliceLen)
	for i := 0; i < int(noiseSliceLen/3); i++ {
		pix := (rand.Float32()*2 - 1) * math.Pi * .5
		noiseRaw[i*3] = byte(pix * 255.0)
		noiseRaw[i*3+1] = byte(pix * 255.0)
		noiseRaw[i*3+2] = 0
	}
	noiseTextureConfig := rhi.TextureConfig{
		Type:            rhi.TextureType2D,
		Width:           noiseSize,
		Height:          noiseSize,
		Format:          rhi.FormatRGB8,
		FilterMin:       rhi.FilterNearest,
		FilterMag:       rhi.FilterNearest,
		WrapS:           rhi.WrapRepeat,
		WrapT:           rhi.WrapRepeat,
		GenerateMipmaps: false,
		Anisotropy:      0,
	}
	noiseTexture, err := rhi.NewTexture(noiseTextureConfig)
	if err != nil {
		return nil, fmt.Errorf("ssao pass - failed to create noise texture %w", err)
	}
	noiseTexture.Bind()
	noiseTexture.UploadRGB(0, 0, noiseSize, noiseSize, noiseRaw)
	return noiseTexture, nil
}
