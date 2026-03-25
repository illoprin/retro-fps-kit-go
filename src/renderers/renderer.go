package renderers

import "github.com/illoprin/retro-fps-kit-go/src/player"

type Renderer interface {
	Prepare(int, int, *player.Camera)
	Render(interface{})
	Shutdown()
}
