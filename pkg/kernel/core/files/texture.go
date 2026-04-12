package files

import (
	"image"

	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/logger"
)

var (
	nullPath = GetTexturePath("null.png")
)

type RGBA8Data struct {
	W, H int32
	Data []byte
}

// LoadTexture helper function for loading RGBA8 textures
func LoadTexture(path string) (*RGBA8Data, error) {
	if path == "" {
		path = nullPath
	}

	// load file
	w, h, imgRaw, err := LoadImage(path)
	if err != nil {
		w, h, imgRaw, err = LoadImage(GetTexturePath("null.png"))
		if err != nil {
			logger.Errorf("failed to load null texture")
			return nil, err
		}
	}

	// build and flip image
	img := image.NewRGBA(imgRaw.Bounds())
	var x, y int
	for i := 0; i < w*h; i++ {
		x = i % w
		y = i / w
		img.Set(x, h-y-1, imgRaw.At(x, y))
	}

	return &RGBA8Data{
		W:    int32(w),
		H:    int32(h),
		Data: img.Pix,
	}, nil
}
