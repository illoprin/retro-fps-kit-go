package renderers

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/illoprin/retro-fps-kit-go/src/engine/assetmgr"
	postprocessing "github.com/illoprin/retro-fps-kit-go/src/post_processing"
	"github.com/illoprin/retro-fps-kit-go/src/render"
)

type ScreenRenderer struct {
	cfg     *postprocessing.ScreenConfig
	program *render.Program
	mesh    *render.Mesh
}

func (s *ScreenRenderer) GetConfig() *postprocessing.ScreenConfig {
	return s.cfg
}

func NewScreen(
	screenQuad *render.Mesh,
	cfg *postprocessing.ScreenConfig,
) (*ScreenRenderer, error) {
	s := &ScreenRenderer{
		cfg:  cfg,
		mesh: screenQuad,
	}

	// init screen quad shader
	screenProg, err := render.NewProgram(
		assetmgr.GetShaderPath("basic_quad.vert"),
		assetmgr.GetShaderPath("screen.frag"),
	)
	if err != nil {
		s.mesh.Delete()
		return nil, err
	}
	s.program = screenProg

	return s, nil
}

func (s *ScreenRenderer) RenderScreenQuad(color *render.Texture) {
	// bind initial framebuffer
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	gl.Disable(gl.DEPTH_TEST)
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
	gl.ClearColor(0, 0, 0, 1.0)
	gl.Viewport(0, 0, int32(s.cfg.Width), int32(s.cfg.Height))

	gl.Clear(gl.COLOR_BUFFER_BIT)

	s.program.Use()
	color.Bind(0)
	s.program.Set1i("u_color", 0)

	s.mesh.Draw()
}

func (s *ScreenRenderer) Delete() {
	s.program.Delete()
}
