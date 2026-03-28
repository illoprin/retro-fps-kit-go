package renderers

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
	postprocessing "github.com/illoprin/retro-fps-kit-go/src/post_processing"
	"github.com/illoprin/retro-fps-kit-go/src/render"
)

// create deferred fbo (color, normal)
// setup opengl state to render into deferred fbo

// DeferredRenderTarget describes deffered fbo and bindings for geometry rendering
type DeferredRenderTarget struct {
	DeferredFBO         *render.Framebuffer
	scConfig            *postprocessing.ScreenConfig
	lastResolutionRatio float32
}

func NewDeferredRenderTarget(
	scConfig *postprocessing.ScreenConfig,
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

	// init scene framebuffer
	deferredFBO, err := render.NewFramebuffer(
		int32(float32(t.scConfig.Width)*t.scConfig.ResolutionRatio),
		int32(float32(t.scConfig.Width)*t.scConfig.ResolutionRatio),
	)
	if err != nil {
		return err
	}

	// init color and depth attachments
	deferredFBO.Bind()

	// color
	err = deferredFBO.NewColorAttachment(render.FormatRGBA8)

	// normal
	err = deferredFBO.NewColorAttachment(render.FormatRGBA16F)
	deferredFBO.SetDrawBuffers([]int{0, 1})

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

	gl.CullFace(gl.BACK)
	gl.FrontFace(gl.CCW)

	return nil
}

func (t *DeferredRenderTarget) BindForNewFrame() {
	t.DeferredFBO.Bind()

	gl.Viewport(0, 0,
		int32(float32(t.scConfig.Width)*t.scConfig.ResolutionRatio),
		int32(float32(t.scConfig.Height)*t.scConfig.ResolutionRatio),
	)
	if t.scConfig.ResolutionRatio != t.lastResolutionRatio {
		t.resizeDeferredFBO()
	}

	gl.ClearColor(0, 0, 0, 0)

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)

	// set wireframe if needed
	if t.scConfig.Wireframe {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
		gl.Disable(gl.CULL_FACE)
	} else {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
	}

	// clear color and depth buffers
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (t *DeferredRenderTarget) ResizeCallback() {
	t.resizeDeferredFBO()
}

func (t *DeferredRenderTarget) resizeDeferredFBO() {
	t.DeferredFBO.Resize(
		int32(float32(t.scConfig.Width)*t.scConfig.ResolutionRatio),
		int32(float32(t.scConfig.Height)*t.scConfig.ResolutionRatio),
	)
}

func (t *DeferredRenderTarget) Delete() {
	if t.DeferredFBO != nil {
		t.DeferredFBO.Delete()
	}
}
