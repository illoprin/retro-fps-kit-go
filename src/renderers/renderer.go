package renderers

import "github.com/illoprin/obj-scene-editor-go/src/player"

type Renderer interface {
	Prepare(*player.Camera)
	Render(interface{})
	Destroy()
}
