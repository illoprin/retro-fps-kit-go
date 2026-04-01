package render

import (
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/illoprin/retro-fps-kit-go/src/engine/global"
	"github.com/illoprin/retro-fps-kit-go/src/model"
)

var (
	basicQuadVertices = []float32{
		-1, -1,
		1, -1,
		1, 1,
		-1, 1,
	}
	basicQuadIndices = []uint32{
		0, 1, 2,
		2, 3, 0,
	}
)

type MeshType uint32

const (
	StaticDraw MeshType = iota
	DynamicDraw
	StreamDraw
)

func getVBOType(t MeshType) uint32 {
	switch t {
	case StaticDraw:
		return gl.STATIC_DRAW
	case DynamicDraw:
		return gl.DYNAMIC_DRAW
	case StreamDraw:
		return gl.STREAM_DRAW
	default:
		return gl.STATIC_DRAW
	}
}

type Mesh struct {
	vbo, ebo, vao, count, instances uint32
}

func NewMesh() *Mesh {
	m := &Mesh{
		count:     0,
		instances: 0,
	}
	gl.GenBuffers(2, &m.vbo)
	gl.GenVertexArrays(1, &m.vao)

	return m
}

func (m *Mesh) SetupBasicQuad() {

	gl.BindVertexArray(m.vao)

	// vbo
	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(basicQuadVertices)*int(global.SizeOfFloat), gl.Ptr(basicQuadVertices), gl.STATIC_DRAW)

	// ebo
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(basicQuadIndices)*int(global.SizeOfFloat), gl.Ptr(basicQuadIndices), gl.STATIC_DRAW)

	// attribute pointers
	// position
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 2*int32(global.SizeOfFloat), unsafe.Pointer(uintptr(0)))
	gl.EnableVertexAttribArray(0)

	// unbind
	gl.BindVertexArray(0)

	m.count = uint32(len(basicQuadIndices))
}

func (m *Mesh) SetupFromModel(model *model.Model, t MeshType) {

	// check model
	if model == nil {
		return
	}

	// check triangles
	if len(model.Indices) < 3 || len(model.Vertices) < 3 {
		return
	}

	gl.BindVertexArray(m.vao)

	// VBO
	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbo)
	gl.BufferData(
		gl.ARRAY_BUFFER,
		len(model.Vertices)*8*int(global.SizeOfFloat),
		gl.Ptr(model.Vertices), getVBOType(t),
	)

	// EBO
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ebo)
	gl.BufferData(
		gl.ELEMENT_ARRAY_BUFFER,
		len(model.Indices)*int(global.SizeOfFloat),
		gl.Ptr(model.Indices),
		gl.STATIC_DRAW,
	)

	// position (location = 0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 8*int32(global.SizeOfFloat), unsafe.Pointer(uintptr(0)))
	gl.EnableVertexAttribArray(0)

	// uv (location = 1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 8*int32(global.SizeOfFloat), unsafe.Pointer(uintptr(3*int(global.SizeOfFloat))))
	gl.EnableVertexAttribArray(1)

	// normal (location = 2)
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, 8*int32(global.SizeOfFloat), unsafe.Pointer(uintptr(5*int(global.SizeOfFloat))))
	gl.EnableVertexAttribArray(2)

	gl.BindVertexArray(0)

	m.count = uint32(len(model.Indices))
}

func (m *Mesh) UpdateVertexBuffer(vertexOffset int, data []model.ModelVertex, verticesTotalCount uint32) {
	if len(data) <= 0 || vertexOffset < 0 {
		return
	}
	gl.BindVertexArray(m.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbo)
	gl.BufferSubData(gl.ARRAY_BUFFER, 8*int(global.SizeOfFloat)*vertexOffset, len(data)*int(global.SizeOfFloat)*8, gl.Ptr(data))
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	if m.count != verticesTotalCount {
		m.count = verticesTotalCount
	}
}

func (m *Mesh) Draw() {
	if m.count <= 0 && m.vao <= 0 {
		return
	}
	gl.BindVertexArray(m.vao)
	// NOTE придумать обёртку над draw mode (triangles, points и тд)
	gl.DrawElements(gl.TRIANGLES, int32(m.count), gl.UNSIGNED_INT, nil)
	global.DrawCalls++
	global.DrawVertices += m.count
	gl.BindVertexArray(0)
}

func (m *Mesh) Delete() {
	gl.DeleteBuffers(2, &m.vbo)
	gl.DeleteVertexArrays(1, &m.vao)
}

func (m *Mesh) GetCount() uint32 {
	return m.count
}
