// Package rhi provides high-level abstractions over OpenGL objects,
// simplifying the development of cross-platform rendering pipelines.
//
// It implements the Render Hardware Interface (RHI) concept,
// shielding the user from low-level OpenGL state management and
// providing an idiomatic Go API for common graphics tasks.
//
// Author: illoprin
//
// 2026

package rhi

import (
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/illoprin/retro-fps-kit-go/src/core/model"
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

func getVBOType(t DrawType) uint32 {
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
	gl.BufferData(gl.ARRAY_BUFFER, len(basicQuadVertices)*int(sizeOfFloat), gl.Ptr(basicQuadVertices), gl.STATIC_DRAW)

	// ebo
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(basicQuadIndices)*int(sizeOfFloat), gl.Ptr(basicQuadIndices), gl.STATIC_DRAW)

	// attribute pointers
	// position
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 2*int32(sizeOfFloat), unsafe.Pointer(uintptr(0)))
	gl.EnableVertexAttribArray(0)

	// unbind
	gl.BindVertexArray(0)

	m.count = uint32(len(basicQuadIndices))
}

func (m *Mesh) SetupFromModel(model *model.Model, t DrawType) {

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
		len(model.Vertices)*8*int(sizeOfFloat),
		gl.Ptr(model.Vertices), getVBOType(t),
	)

	// EBO
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ebo)
	gl.BufferData(
		gl.ELEMENT_ARRAY_BUFFER,
		len(model.Indices)*int(sizeOfFloat),
		gl.Ptr(model.Indices),
		gl.STATIC_DRAW,
	)

	// position (location = 0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 8*int32(sizeOfFloat), unsafe.Pointer(uintptr(0)))
	gl.EnableVertexAttribArray(0)

	// uv (location = 1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 8*int32(sizeOfFloat), unsafe.Pointer(uintptr(3*int(sizeOfFloat))))
	gl.EnableVertexAttribArray(1)

	// normal (location = 2)
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, 8*int32(sizeOfFloat), unsafe.Pointer(uintptr(5*int(sizeOfFloat))))
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
	gl.BufferSubData(gl.ARRAY_BUFFER, 8*int(sizeOfFloat)*vertexOffset, len(data)*int(sizeOfFloat)*8, gl.Ptr(data))
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	if m.count != verticesTotalCount {
		m.count = verticesTotalCount
	}
}

// NOTE придумать обёртку над glVertexAttribDivisior

func (m *Mesh) Draw() {
	if m.count <= 0 && m.vao <= 0 {
		return
	}
	gl.BindVertexArray(m.vao)

	// draw call!
	gl.DrawElements(gl.TRIANGLES, int32(m.count), gl.UNSIGNED_INT, nil)

	// update stats
	FrameStats.DrawCalls++
	FrameStats.Vertices += uint64(m.count)
	FrameStats.Triangles += uint64(m.count) / 3

	gl.BindVertexArray(0)
}

func (m *Mesh) Delete() {
	gl.DeleteBuffers(2, &m.vbo)
	gl.DeleteVertexArrays(1, &m.vao)
}

func (m *Mesh) GetCount() uint32 {
	return m.count
}
