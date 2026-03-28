package postprocessing

type ScreenConfig struct {
	Width, Height       int32
	ResolutionRatio     float32
	LastResolutionRatio float32
	Aspect              float32
	Wireframe           bool
}
