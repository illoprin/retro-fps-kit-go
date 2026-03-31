package renderers

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/illoprin/retro-fps-kit-go/src/render"
	"github.com/illoprin/retro-fps-kit-go/src/window"
)

// create deferred fbo (color, normal)
// setup opengl state to render into deferred fbo

// DeferredRenderTarget describes deffered fbo and bindings for geometry rendering
type DeferredRenderTarget struct {
	DeferredFBO         *render.Framebuffer
	scConfig            *window.ScreenConfig
	lastResolutionRatio float32
	Wireframe           bool
}

func NewDeferredRenderTarget(
	scConfig *window.ScreenConfig,
) (*DeferredRenderTarget, error) {

	t := &DeferredRenderTarget{
		scConfig: scConfig,
	}

	// init deferred framebuffer
	if err := t.setupFramebuffer(); err != nil {
		return nil, err
	}

	return t, nil
}

func (t *DeferredRenderTarget) setupFramebuffer() error {

	// init new framebuffer
	fbWidth, fbHeight := t.scConfig.GetScreenSize()
	deferredFBO, err := render.NewFramebuffer(fbWidth, fbHeight)
	if err != nil {
		return err
	}

	// init color and depth attachments
	deferredFBO.Bind()

	// color
	err = deferredFBO.NewColorAttachment(render.FormatRGBA8)

	// normal
	err = deferredFBO.NewColorAttachment(render.FormatRGB16F)

	// position
	err = deferredFBO.NewColorAttachment(render.FormatRGB32F)

	deferredFBO.SetDrawBuffers([]int{0, 1, 2})

	// depth
	err = deferredFBO.NewDepthAttachment()
	if err != nil {
		deferredFBO.Delete()
		return err
	}

	// check framebuffer completness
	if !deferredFBO.Check() {
		deferredFBO.Delete()
		return fmt.Errorf("fbo not completed")
	}
	deferredFBO.Unbind()

	t.DeferredFBO = deferredFBO

	gl.ClearColor(0, 0, 0, 0)
	gl.CullFace(gl.BACK)
	gl.FrontFace(gl.CCW)

	return nil
}

func (t *DeferredRenderTarget) BindForNewFrame() {
	t.DeferredFBO.Bind()

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)

	// set wireframe if needed
	if t.Wireframe {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
		gl.Disable(gl.CULL_FACE)
	} else {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
	}

	// clear color and depth buffers
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (t *DeferredRenderTarget) ResizeCallback() {
	fbWidth, fbHeight := t.scConfig.GetScreenSize()
	t.DeferredFBO.Resize(fbWidth, fbHeight)
}

func (t *DeferredRenderTarget) Delete() {
	if t.DeferredFBO != nil {
		t.DeferredFBO.Delete()
	}
}
