package render

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/illoprin/retro-fps-kit-go/src/engine/global"
	"github.com/illoprin/retro-fps-kit-go/src/model"
)

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

func (m *Mesh) SetupFromModel(model *model.Model, usage uint32) {

	// check model
	if model == nil {
		return
	}

	// check triangles
	if len(model.Indices) < 3 || len(model.Vertices) < 3 {
		return
	}

	gl.BindVertexArray(m.vao)

	sizeOfFloat := 4

	// VBO
	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(model.Vertices)*8*sizeOfFloat, gl.Ptr(model.Vertices), usage)

	// EBO
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(model.Indices)*sizeOfFloat, gl.Ptr(model.Indices), gl.STATIC_DRAW)

	// position (location = 0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 8*int32(sizeOfFloat), gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// uv (location = 1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 8*int32(sizeOfFloat), gl.PtrOffset(3*sizeOfFloat))
	gl.EnableVertexAttribArray(1)

	// normal (location = 2)
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, 8*int32(sizeOfFloat), gl.PtrOffset(5*sizeOfFloat))
	gl.EnableVertexAttribArray(2)

	gl.BindVertexArray(0)

	m.count = uint32(len(model.Indices))
}

func (m *Mesh) Draw() {
	gl.BindVertexArray(m.vao)
	gl.DrawElements(gl.TRIANGLES, int32(m.count), gl.UNSIGNED_INT, nil)
	global.DrawCalls++
	global.DrawVertices += m.count
	gl.BindVertexArray(0)
}

func (m *Mesh) Delete() {
	gl.DeleteBuffers(2, &m.vbo)
	gl.DeleteVertexArrays(1, &m.vao)
}
