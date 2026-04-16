package post

import (
	"fmt"
	"unsafe"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/files"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/window"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/pipeline"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/rhi"
)

var (
	bayer_2x2 = []float32{
		0 / 16, 2 / 16,
		3 / 16, 1 / 16,
	}

	bayer_4x4 = []float32{
		0 / 16, 8 / 16, 2 / 16, 10 / 16,
		12 / 16, 4 / 16, 14 / 16, 6 / 16,
		3 / 16, 11 / 16, 1 / 16, 9 / 16,
		15 / 16, 7 / 16, 13 / 16, 5 / 16,
	}
)

type DitheringConfig struct {
	Use   bool
	Speed float32
	Min   float32
	Max   float32
}

type DitheringPass struct {
	fbo       *rhi.Framebuffer
	matrix    *rhi.Texture
	program   *rhi.Program
	mesh      *rhi.Mesh
	cfg       *DitheringConfig
	screen    *window.ScreenConfig
	resources []rhi.Resource
}

// NewDitheringPass creates new render pass object
// LDR effect
func NewDitheringPass(
	s PassSharedResources,
	cfg *DitheringConfig,
) (*DitheringPass, error) {
	d := &DitheringPass{
		cfg:    cfg,
		screen: s.ScreenConfig,
	}

	W, H := s.ScreenConfig.GetScreenSize()

	// init framebuffer
	fbo := rhi.NewFramebuffer(W, H)
	fbo.NewColorAttachment(rhi.FormatRGB8, rhi.FilterNearest)
	if !fbo.Check() {
		return nil, fmt.Errorf("fbo not completed")
	}
	d.fbo = fbo

	// load program
	prog, err := rhi.NewProgram(
		files.GetShaderPath("initial/screen.vert"),
		files.GetShaderPath("initial/dithering.frag"),
	)
	if err != nil {
		d.Delete()
		return nil, fmt.Errorf("failed to load program")
	}
	d.program = prog

	// create matrix
	d.createMatrix()

	d.resources = append(d.resources, d.fbo, d.program)

	return d, nil
}

func (d *DitheringPass) createMatrix() {
	matrixSize := 2
	matrix := bayer_2x2

	config := rhi.TextureConfig{
		Type:      rhi.TextureType2D,
		Format:    rhi.FormatR32F,
		Width:     int32(matrixSize),
		Height:    int32(matrixSize),
		FilterMin: rhi.FilterNearest,
		FilterMag: rhi.FilterNearest,
		WrapS:     rhi.WrapRepeat,
		WrapT:     rhi.WrapRepeat,
	}

	tex := rhi.NewTexture(config)
	tex.Upload2D(0, 0, unsafe.Pointer(&matrix[0]))
}

func (d *DitheringPass) GetColor() *rhi.Texture {
	return d.fbo.GetColorTexture(0)
}

func (d *DitheringPass) GetResultFramebuffer() *rhi.Framebuffer {
	return d.fbo
}

func (d *DitheringPass) GetConfig() interface{} {
	return d.cfg
}

func (d *DitheringPass) Use() bool {
	return d.cfg.Use
}

func (d *DitheringPass) RenderPass(src *pipeline.DeferredRenderResult) {
	d.program.Use()

	d.fbo.BindForDrawing()

	// color
	src.Color.BindToUnit(0)
	d.program.Set1i("u_color", 0)

	// matrix
	d.matrix.BindToUnit(1)
	d.program.Set1i("u_matrix", 0)

	// params
	time := float32(glfw.GetTime())
	d.program.Set1i("u_matrix_size", d.matrix.Config.Width)
	d.program.Set1f("u_time", time)
	d.program.Set1f("u_min", d.cfg.Min)
	d.program.Set1f("u_max", d.cfg.Max)

	d.mesh.Draw()

}

func (d *DitheringPass) ResizeCallback() {
	W, H := d.screen.GetScreenSize()
	d.fbo.Resize(W, H)
}

func (d *DitheringPass) Delete() {
	for _, r := range d.resources {
		r.Delete()
	}
}
