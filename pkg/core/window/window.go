package window

import (
	"fmt"
	"image"
	_ "image/png"
	"os"

	"github.com/go-gl/glfw/v3.3/glfw"
)

type ScreenConfig struct {
	Width, Height       int32
	ResolutionRatio     float32
	LastResolutionRatio float32
	Aspect              float32
}

func (s *ScreenConfig) GetScreenSize() (int32, int32) {
	return int32(float32(s.Width) * s.ResolutionRatio),
		int32(float32(s.Height) * s.ResolutionRatio)
}

type Window struct {
	*glfw.Window
	cfg              *ScreenConfig
	userSizeCallback glfw.FramebufferSizeCallback
	cursorDisabled   bool
}

func InitGLFW() error {
	// init glfw
	if err := glfw.Init(); err != nil {
		return err
	}

	// setup hints
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLDebugContext, glfw.True)
	return nil
}

func NewWindow(width, height int, title string, resolutionRatio float32) (*Window, error) {
	glfwW, err := glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		return nil, err
	}

	w := &Window{
		Window: glfwW,
		cfg: &ScreenConfig{
			Width:               int32(width),
			Height:              int32(height),
			Aspect:              float32(width) / float32(height),
			ResolutionRatio:     resolutionRatio,
			LastResolutionRatio: resolutionRatio,
		},
	}

	w.setupCallbacks()

	return w, nil
}

func (w *Window) GetConfig() *ScreenConfig {
	return w.cfg
}

func (w *Window) SetResizeCallback(f glfw.FramebufferSizeCallback) {
	w.userSizeCallback = f
}

func (w *Window) setupCallbacks() {
	w.Window.SetFramebufferSizeCallback(func(win *glfw.Window, width, height int) {
		w.cfg.Width = int32(width)
		w.cfg.Height = int32(height)
		w.cfg.Aspect = float32(width) / float32(height)
		if w.userSizeCallback != nil {
			w.userSizeCallback(win, width, height)
		}
	})
}

// ToggleCursor toggle cursor visibility
func (w *Window) ToggleCursor() {
	w.cursorDisabled = !w.cursorDisabled

	if w.cursorDisabled {
		w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
		w.SetInputMode(glfw.RawMouseMotion, glfw.True)
	} else {
		w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
		w.SetInputMode(glfw.RawMouseMotion, glfw.False)
	}
}

// SetIconFromFile sets window icon from *png* image
func (w *Window) SetIconFromFile(f string) error {
	file, err := os.Open(f)
	if err != nil {
		return fmt.Errorf("window - failed to open file")
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("window - failed to decode image")
	}

	w.SetIcon([]image.Image{img})
	return nil
}

// SetIconFromFile loads new glfw cursor
func (w *Window) LoadCursor(f string) (*glfw.Cursor, error) {

	file, err := os.Open(f)
	if err != nil {
		return nil, fmt.Errorf("window - failed to open file")
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("window - failed to decode image")
	}

	return glfw.CreateCursor(img, 0, 0), nil
}

func (w *Window) GetCursorDisabled() bool {
	return w.cursorDisabled
}

func (w *Window) Center() {
	vidMode := glfw.GetPrimaryMonitor().GetVideoMode()
	w.SetPos(vidMode.Width/2-int(w.cfg.Width)/2, vidMode.Height/2-int(w.cfg.Height)/2)
}
