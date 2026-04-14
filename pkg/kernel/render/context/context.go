package context

import (
	"log"
	"log/slog"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/rhi"
)

type ContextState struct {
	EnableDepthTest bool
	EnableCullFace  bool
	EnableBlend     bool
	DepthFunc       int32
	BlendSrc        int32
	BlendDst        int32
	CullFaceMode    int32
	PolygonMode     int32
}

var state ContextState

// CaptureState queries the current OpenGL state.
func CaptureState() {
	state.EnableDepthTest = gl.IsEnabled(gl.DEPTH_TEST)
	state.EnableCullFace = gl.IsEnabled(gl.CULL_FACE)
	state.EnableBlend = gl.IsEnabled(gl.BLEND)

	gl.GetIntegerv(gl.DEPTH_FUNC, &state.DepthFunc)
	gl.GetIntegerv(gl.BLEND_SRC_RGB, &state.BlendSrc)
	gl.GetIntegerv(gl.BLEND_DST_RGB, &state.BlendDst)
	gl.GetIntegerv(gl.CULL_FACE_MODE, &state.CullFaceMode)

	var polygonMode [2]int32
	gl.GetIntegerv(gl.POLYGON_MODE, &polygonMode[0])
	state.PolygonMode = polygonMode[0]
}

// RestoreState applies the saved state back to the OpenGL machine.
func RestoreState() {
	if state.EnableDepthTest {
		gl.Enable(gl.DEPTH_TEST)
	} else {
		gl.Disable(gl.DEPTH_TEST)
	}
	if state.EnableCullFace {
		gl.Enable(gl.CULL_FACE)
	} else {
		gl.Disable(gl.CULL_FACE)
	}
	if state.EnableBlend {
		gl.Enable(gl.BLEND)
	} else {
		gl.Disable(gl.BLEND)
	}

	gl.DepthFunc(uint32(state.DepthFunc))
	gl.BlendFunc(uint32(state.BlendSrc), uint32(state.BlendDst))
	gl.CullFace(uint32(state.CullFaceMode))
	gl.PolygonMode(gl.FRONT_AND_BACK, uint32(state.PolygonMode))
}

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
	gl.Enable(gl.DEPTH_TEST)

	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)
	gl.FrontFace(gl.CCW)

	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
}

func SetDepth(v bool) {
	if v {
		gl.Enable(gl.DEPTH_TEST)
	} else {
		gl.Disable(gl.DEPTH_TEST)
	}
}

func SetFaceCulling(v bool) {
	if v {
		gl.Enable(gl.CULL_FACE)
	} else {
		gl.Disable(gl.CULL_FACE)
	}
}

func SetBlending(v bool) {
	if v {
		gl.Enable(gl.BLEND)
	} else {
		gl.Disable(gl.BLEND)
	}
}

func SetupBlending() {
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
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
	slog.Info("Renderer - " + gl.GoStr(gl.GetString(gl.RENDERER)))
	slog.Info("OpenGL Version - " + gl.GoStr(gl.GetString(gl.VERSION)))
	slog.Info("GLSL Version - " + gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION)))
}

func SetupViewport(xo, yo, w, h int32) {
	gl.Viewport(xo, yo, w, h)
}

func BindFramebuffer(f *rhi.Framebuffer) {
	if f != nil {
		f.BindForDrawing()
		return
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}
