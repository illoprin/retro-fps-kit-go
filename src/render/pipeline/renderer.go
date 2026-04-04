package pipeline

import (
	"github.com/illoprin/retro-fps-kit-go/src/core/camera"
)

type Renderer interface {
	Prepare(int, int, *camera.Camera3D)
	Render(interface{})
	Shutdown()
}
