package files

import "image"

var (
	nullPath = GetTexturePath("null.png")
)

type ImageData struct {
	W, H int32
	Data []byte
}

// LoadTexture helper function for loading RGBA8 textures
func LoadTexture(path string) (*ImageData, error) {
	if path == "" {
		path = nullPath
	}

	// load file
	w, h, imgRaw, err := LoadImage(path)
	if err != nil {
		// fallback на null
		w, h, imgRaw, err = LoadImage(nullPath)
		if err != nil {
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

	return &ImageData{
		W:    int32(w),
		H:    int32(h),
		Data: img.Pix,
	}, nil
}
