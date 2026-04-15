package gui

import (
	"fmt"
	"math"
	"unsafe"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/files"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/rhi"
)

type guiVertex struct {
	Pos   mgl.Vec2
	UV    mgl.Vec2
	Color mgl.Vec4
}

type GUICanvas struct {
	// gpu
	mesh    *rhi.Mesh
	program *rhi.Program

	// mesh data
	vertices []guiVertex
	indices  []uint32

	dirty bool
}

func NewGUICanvas() (*GUICanvas, error) {

	// init program
	program, err := rhi.NewProgram(
		files.GetShaderPath("flat/gui_canvas.vert"),
		files.GetShaderPath("flat/gui_canvas.frag"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load gui canvas program - %w", err)
	}

	// create mesh for drawing
	m := rhi.NewMesh()

	m.Bind()

	m.CreateVertexBuffer()
	m.AllocateVertexBuffer(0, 1024*1024, rhi.StreamDraw) // reserve 1MB

	m.CreateElementBuffer()
	m.AllocateElementBuffer(1024*1024, rhi.StreamDraw) // reserve 1MB

	// position
	m.SetAttribute(0, rhi.VertexAttribute{
		Index:       0,
		Components:  2,
		Type:        rhi.Float32,
		StrideBytes: int32(unsafe.Sizeof(guiVertex{})),
		OffsetBytes: 0,
	})

	// uv
	m.SetAttribute(0, rhi.VertexAttribute{
		Index:       1,
		Components:  2,
		Type:        rhi.Float32,
		StrideBytes: int32(unsafe.Sizeof(guiVertex{})),
		OffsetBytes: 8,
	})

	// color
	m.SetAttribute(0, rhi.VertexAttribute{
		Index:       2,
		Components:  4,
		Type:        rhi.Float32,
		StrideBytes: int32(unsafe.Sizeof(guiVertex{})),
		OffsetBytes: 16,
	})

	return &GUICanvas{
		mesh:    m,
		program: program,
		dirty:   true,
	}, nil
}

func (c *GUICanvas) pushQuad(v0, v1, v2, v3 guiVertex) {
	base := uint32(len(c.vertices))

	c.vertices = append(c.vertices, v0, v1, v2, v3)

	c.indices = append(c.indices,
		base+0, base+1, base+2,
		base+0, base+2, base+3,
	)
}

// Line - draws line from p0 to p1, uses NDC coordinate system (-1, 1)
func (c *GUICanvas) Line(p0, p1 mgl.Vec2, color mgl.Vec4, thickness float32) {
	dir := p1.Sub(p0).Normalize()
	normal := mgl.Vec2{-dir.Y(), dir.X()}

	offset := normal.Mul(thickness * 0.5)

	v0 := guiVertex{Pos: p0.Add(offset), Color: color}
	v1 := guiVertex{Pos: p1.Add(offset), Color: color}
	v2 := guiVertex{Pos: p1.Sub(offset), Color: color}
	v3 := guiVertex{Pos: p0.Sub(offset), Color: color}

	c.pushQuad(v3, v2, v1, v0)
	c.dirty = true
}

// Circle - draws circle in center with radius, uses NDC coordinate system (-1, 1)
func (c *GUICanvas) Circle(center mgl.Vec2, radius float32, color mgl.Vec4, segments int) {
	base := uint32(len(c.vertices))

	c.vertices = append(c.vertices, guiVertex{
		Pos:   center,
		Color: color,
	})

	for i := 0; i <= segments; i++ {
		angle := float32(i) / float32(segments) * (2.0 * 3.1415926)

		p := mgl.Vec2{
			center.X() + float32(math.Cos(float64(angle)))*radius,
			center.Y() + float32(math.Sin(float64(angle)))*radius,
		}

		c.vertices = append(c.vertices, guiVertex{
			Pos:   p,
			Color: color,
		})
	}

	for i := 1; i <= segments; i++ {
		c.indices = append(c.indices,
			base,
			base+uint32(i),
			base+uint32(i+1),
		)
	}

	c.dirty = true
}

// Rect - draws rect from p0 to p1, uses NDC coordinate system (-1, 1)
func (c *GUICanvas) Rect(p0, p1 mgl.Vec2, color mgl.Vec4) {
	v0 := guiVertex{Pos: mgl.Vec2{p0.X(), p0.Y()}, Color: color, UV: mgl.Vec2{0, 0}}
	v1 := guiVertex{Pos: mgl.Vec2{p1.X(), p0.Y()}, Color: color, UV: mgl.Vec2{1, 0}}
	v2 := guiVertex{Pos: mgl.Vec2{p1.X(), p1.Y()}, Color: color, UV: mgl.Vec2{1, 1}}
	v3 := guiVertex{Pos: mgl.Vec2{p0.X(), p1.Y()}, Color: color, UV: mgl.Vec2{0, 1}}

	c.pushQuad(v0, v1, v2, v3)
	c.dirty = true
}

// Update - updates mesh, removes dirty flag
func (c *GUICanvas) Update() {
	if !c.dirty {
		return
	}

	if len(c.vertices) == 0 {
		return
	}

	c.mesh.Bind()

	c.mesh.SetVertexBufferData(
		0,
		0,
		len(c.vertices)*int(unsafe.Sizeof(guiVertex{})),
		unsafe.Pointer(&c.vertices[0]),
	)

	c.mesh.SetElementBufferData(0, c.indices)

	c.dirty = false
}

// Draw - render canvas in active framebuffer
func (c *GUICanvas) Draw(aspect float32) {
	if len(c.indices) == 0 {
		return
	}

	c.program.Use()

	proj := mgl.Ortho(-aspect, aspect, -1, 1, -1, 1)
	c.program.SetMat4("u_projection", proj)

	c.mesh.Draw()
}

// Clear - clear canvas data
// (not to be confused with delete)
func (c *GUICanvas) Clear() {
	c.vertices = c.vertices[:0]
	c.indices = c.indices[:0]
	c.dirty = true
}

func (c *GUICanvas) Delete() {
	c.mesh.Delete()
	c.program.Delete()
}
