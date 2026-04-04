package context

import (
	"log"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
)

func InitContext() error {
	if err := gl.Init(); err != nil {
		log.Printf("failed to ini opengl context - %v\n", err)
		return err
	}
	return nil
}

func SetupDebugOutput() {
	gl.Enable(gl.DEBUG_OUTPUT)
	// real-time single thread debugging
	gl.Enable(gl.DEBUG_OUTPUT_SYNCHRONOUS)
	gl.DebugMessageCallback(gl.DebugProc(func(
		source uint32,
		gltype uint32,
		id uint32,
		severity uint32,
		length int32,
		message string,
		userParam unsafe.Pointer,
	) {
		if severity == gl.DEBUG_SEVERITY_NOTIFICATION {
			return // trash filter
		}

		log.Printf("[GL][%d][%d] severity=%d: %s\n",
			source, gltype, severity, message,
		)
	}), nil)
}

func SetupForGeometry() {
	// gl.Enable(gl.BLEND)
	// gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.Enable(gl.DEPTH_TEST)

	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)
	gl.FrontFace(gl.CCW)

	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
}

func SetClearColor(r, g, b, a float32) {
	gl.ClearColor(r, g, b, a)
}

func SetupForWireframe() {
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	gl.Disable(gl.CULL_FACE)
}

func SetupForFlat() {
	gl.Disable(gl.DEPTH_TEST)
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
}

func ClearDepthAndColorBuffers() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func ClearDepthColorAndStencilBuffers() {
	gl.Clear(gl.COLOR_BUFFER_BIT |
		gl.DEPTH_BUFFER_BIT |
		gl.STENCIL_BUFFER_BIT,
	)
}

func ClearColorBuffer() {
	gl.Clear(gl.COLOR_BUFFER_BIT)
}

func LogUserHardware() {
	log.Println(gl.GoStr(gl.GetString(gl.RENDERER)))
	log.Println(gl.GoStr(gl.GetString(gl.VERSION)))
	log.Println(gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION)))
}

func SetupViewport(xo, yo, w, h int32) {
	gl.Viewport(xo, yo, w, h)
}

func BindFramebuffer(f *rhi.Framebuffer) {
	if f != nil {
		f.Bind()
		return
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

func SaveState() {
	// FIX Implement
}

func RestoreState() {
	// FIX Implement
}
