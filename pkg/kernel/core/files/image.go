package files

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
)

func LoadImage(f string) (int, int, image.Image, error) {
	file, err := os.Open(f)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to open file")
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to decode image")
	}

	W, H := img.Bounds().Dx(), img.Bounds().Dy()

	return W, H, img, nil
}
