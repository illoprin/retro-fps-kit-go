package pipeline

import (
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/camera"
)

type Renderer interface {
	Prepare(width, height int, c *camera.Camera3D)
	Render(interface{})
	Shutdown()
}
