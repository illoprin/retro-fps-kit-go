package pipeline

import (
	"github.com/illoprin/retro-fps-kit-go/pkg/core/camera"
)

type Renderer interface {
	Prepare(int, int, *camera.Camera3D)
	Render(interface{})
	Shutdown()
}
