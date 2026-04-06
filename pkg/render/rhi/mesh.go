package rhi

import (
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type Mesh struct {
	vao        uint32
	ebo        uint32
	vbos       []uint32
	indexCount uint32
}

type VertexAttribute struct {
	Index       uint32
	Components  int32
	Type        DataType
	StrideBytes int32
	OffsetBytes int32
	Divisor     uint32
}

func NewMesh() *Mesh {
	m := &Mesh{
		vbos: make([]uint32, 0),
	}
	gl.GenVertexArrays(1, &m.vao)
	return m
}

// CreateVertexBuffer - creates VBO object, returns *index*
func (m *Mesh) CreateVertexBuffer() int {
	var vbo uint32
	gl.GenBuffers(1, &vbo)

	m.vbos = append(m.vbos, vbo)
	return len(m.vbos) - 1
}

// AllocateVertexBuffer - allocates memory for VBO
// !!! BIND BEFORE USE
func (m *Mesh) AllocateVertexBuffer(index int, sizeBytes int, bType BufferType) {
	if index < 0 || index >= len(m.vbos) {
		return
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbos[index])
	gl.BufferData(gl.ARRAY_BUFFER, sizeBytes, nil, GetBufferType(bType))
}

// SetVertexBufferData - updates data in VBO
// !!! BIND BEFORE USE
func (m *Mesh) SetVertexBufferData(
	index int,
	offsetBytes int,
	sizeBytes int,
	data unsafe.Pointer,
) {
	if index < 0 || index >= len(m.vbos) {
		return
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbos[index])
	gl.BufferSubData(gl.ARRAY_BUFFER, offsetBytes, sizeBytes, data)
}

// CreateElementBuffer - creates EBO object
// !!! BIND BEFORE USE
func (m *Mesh) CreateElementBuffer() {
	if m.ebo != 0 {
		return
	}

	gl.GenBuffers(1, &m.ebo)
}

// AllocateElementBuffer - allocates memory for EBO (indices) data
// !!! BIND BEFORE USE
func (m *Mesh) AllocateElementBuffer(sizeBytes int, bType BufferType) {
	if m.ebo == 0 {
		return
	}

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, sizeBytes, nil, GetBufferType(bType))
}

// SetElementBufferData - update EBO (indices) data
// !!! BIND BEFORE USE
func (m *Mesh) SetElementBufferData(offset int, data []uint32) {
	if m.ebo == 0 || len(data) == 0 {
		return
	}

	size := len(data) * int(unsafe.Sizeof(data[0]))

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ebo)
	gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, offset, size, gl.Ptr(data))

	m.indexCount = uint32(len(data))
}

// SetAttribute - set attribute pointer
// !!! BIND BEFORE USE
func (m *Mesh) SetAttribute(vboIndex int, a VertexAttribute) {
	if vboIndex < 0 || vboIndex >= len(m.vbos) {
		return
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbos[vboIndex])

	gl.VertexAttribPointer(
		a.Index,
		a.Components,
		GetDataType(a.Type),
		false,
		a.StrideBytes,
		unsafe.Pointer(uintptr(a.OffsetBytes)),
	)
	gl.EnableVertexAttribArray(a.Index)

	if a.Divisor > 0 {
		gl.VertexAttribDivisor(a.Index, a.Divisor)
	}
}

func (m *Mesh) Bind() {
	gl.BindVertexArray(m.vao)
}

func (m *Mesh) Draw() {

	if m.vao <= 0 || m.indexCount < 3 || m.ebo == 0 {
		return
	}
	m.Bind()

	gl.DrawElements(gl.TRIANGLES, int32(m.indexCount), gl.UNSIGNED_INT, nil)

	// update stats
	FrameStats.DrawCalls++
	FrameStats.Vertices += uint64(m.indexCount)
	FrameStats.Triangles += uint64(m.indexCount) / 3

}

func (m *Mesh) DrawInstanced(instances int32) {
	if m.vao <= 0 || m.indexCount < 3 || m.ebo == 0 {
		return
	}
	m.Bind()

	gl.DrawElementsInstanced(gl.TRIANGLES, int32(m.indexCount), gl.UNSIGNED_INT, nil, instances)

	// update stats
	FrameStats.DrawCalls++
	FrameStats.Vertices += uint64(m.indexCount) * uint64(instances)
	FrameStats.Triangles += (uint64(m.indexCount) / 3) * uint64(instances)
}

func (m *Mesh) Delete() {
	if m.vao != 0 {
		gl.DeleteVertexArrays(1, &m.vao)
	}
	if len(m.vbos) > 0 {
		gl.DeleteBuffers(int32(len(m.vbos)), &m.vbos[0])
	}
	if m.ebo != 0 {
		gl.DeleteBuffers(1, &m.ebo)
	}

	m.vbos = make([]uint32, 0)
	m.vao = 0
	m.ebo = 0
	m.indexCount = 0
}
