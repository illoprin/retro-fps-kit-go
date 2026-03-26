package window

import (
	"github.com/go-gl/glfw/v3.3/glfw"
)

type Window struct {
	*glfw.Window
	width, height int
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
	return nil
}

func NewWindow(width, height int, title string) (*Window, error) {
	glfwW, err := glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		return nil, err
	}

	w := &Window{
		glfwW, width, height,
	}

	w.setupCallbacks()

	return w, nil
}

func (w *Window) setupCallbacks() {
	w.Window.SetFramebufferSizeCallback(func(_ *glfw.Window, width, height int) {
		w.width = width
		w.height = height
	})
}

func (w *Window) Center() {
	vidMode := glfw.GetPrimaryMonitor().GetVideoMode()
	w.SetPos(vidMode.Width/2-w.width/2, vidMode.Height/2-w.height/2)
}

func (w *Window) GetSize() (int, int) {
	return w.width, w.height
}
