package pipeline

import (
	"fmt"

	"github.com/illoprin/retro-fps-kit-go/pkg/core/window"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/context"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
)

// create deferred fbo (color, normal)
// setup opengl state to render into deferred fbo

// DeferredRenderTarget describes deffered fbo and bindings for geometry rendering
type DeferredRenderTarget struct {
	fbo                 *rhi.Framebuffer
	scConfig            *window.ScreenConfig
	lastResolutionRatio float32
	Wireframe           bool
}

type DeferredRenderResult struct {
	Color    *rhi.Texture
	Normal   *rhi.Texture
	Position *rhi.Texture
	Depth    *rhi.Texture
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
	deferredFBO, err := rhi.NewFramebuffer(fbWidth, fbHeight)
	if err != nil {
		return err
	}

	// init color and depth attachments
	deferredFBO.Bind()

	// color
	err = deferredFBO.NewColorAttachment(rhi.FormatRGB32F) // WARN HDR
	if err != nil {
		deferredFBO.Delete()
		return err
	}

	// normal
	err = deferredFBO.NewColorAttachment(rhi.FormatRGB16F)
	if err != nil {
		deferredFBO.Delete()
		return err
	}

	// position
	err = deferredFBO.NewColorAttachment(rhi.FormatRGB32F)
	if err != nil {
		deferredFBO.Delete()
		return err
	}

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

	t.fbo = deferredFBO

	context.SetClearColor(0, 0, 0, 0)
	context.SetupForGeometry()

	return nil
}

func (t *DeferredRenderTarget) BindForNewFrame() {
	t.fbo.Bind()

	// set wireframe if needed
	if t.Wireframe {
		context.SetupForWireframe()
	}

	// clear color and depth buffers
	context.ClearDepthAndColorBuffers()
}

func (t *DeferredRenderTarget) ResizeCallback() {
	fbWidth, fbHeight := t.scConfig.GetScreenSize()
	t.fbo.Resize(fbWidth, fbHeight)
}

func (t *DeferredRenderTarget) GetResult() *DeferredRenderResult {
	return &DeferredRenderResult{
		Color:    t.fbo.GetColorTexture(0),
		Normal:   t.fbo.GetColorTexture(1),
		Position: t.fbo.GetColorTexture(2),
		Depth:    t.fbo.GetDepthTexture(),
	}
}

func (t *DeferredRenderTarget) GetFramebuffer() *rhi.Framebuffer {
	return t.fbo
}

func (t *DeferredRenderTarget) Delete() {
	if t.fbo != nil {
		t.fbo.Delete()
	}
}
