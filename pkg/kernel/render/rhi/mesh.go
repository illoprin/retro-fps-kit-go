package rhi

import (
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/logger"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/context"
)

type vertexBuffer struct {
	handle    uint32
	allocated uint32
}

type Mesh struct {
	vao        uint32
	ebo        uint32
	vbos       []vertexBuffer
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
		vbos: make([]vertexBuffer, 0),
	}
	gl.GenVertexArrays(1, &m.vao)
	logger.Infof("mesh id=%d created", m.vao)
	return m
}

// CreateVertexBuffer - creates VBO object, returns *index*
func (m *Mesh) CreateVertexBuffer() int {
	var vbo uint32
	gl.GenBuffers(1, &vbo)

	m.vbos = append(m.vbos, vertexBuffer{handle: vbo})
	return len(m.vbos) - 1
}

// AllocateVertexBuffer - allocates memory for VBO
func (m *Mesh) AllocateVertexBufferWithData(index int, sizeBytes int, data unsafe.Pointer, bType BufferType) {
	if index < 0 || index >= len(m.vbos) {
		return
	}

	vbo := &m.vbos[index]

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo.handle)
	gl.BufferData(gl.ARRAY_BUFFER, sizeBytes, data, GetBufferType(bType))
	vbo.allocated = uint32(sizeBytes)
}

// SetVertexBufferData - updates data in VBO
func (m *Mesh) SetVertexBufferData(
	index int,
	offsetBytes int,
	sizeBytes int,
	data unsafe.Pointer,
) {
	if index < 0 || index >= len(m.vbos) {
		return
	}

	vbo := &m.vbos[index]

	if sizeBytes > int(vbo.allocated) {
		logger.Warnf("data size is greater than the allocated memory")
		return
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo.handle)
	gl.BufferSubData(gl.ARRAY_BUFFER, offsetBytes, sizeBytes, data)
}

// CreateElementBuffer - creates EBO object for VAO
// !!! BIND BEFORE USE
func (m *Mesh) CreateElementBuffer() {
	if m.ebo != 0 {
		return
	}

	gl.GenBuffers(1, &m.ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ebo)
}

// AllocateElementBuffer - allocates memory for EBO (indices) data
// !!! BE CAREFUL WITH VAO BINDING (Binding = m.VAO; Binding = 0)
func (m *Mesh) AllocateElementBuffer(sizeBytes int, bType BufferType) {
	if m.ebo == 0 {
		return
	}

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, sizeBytes, nil, GetBufferType(bType))
}

// SetElementBufferData - update EBO (indices) data
// !!! BE CAREFUL WITH VAO BINDING (Binding = m.VAO; Binding = 0)
func (m *Mesh) SetElementBufferData(offset int, data []uint32) {
	if m.ebo == 0 || len(data) == 0 {
		return
	}

	size := len(data) * int(unsafe.Sizeof(data[0]))

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ebo)
	gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, offset, size, gl.Ptr(data))

	m.indexCount = uint32(len(data))
}

// SetAttribute - set attribute pointer based on VBO object
// Allows you to set attributes for GL_FLOAT, GL_UNSIGNED_INT, and GL_INT
// !!! BIND BEFORE USE
func (m *Mesh) SetAttribute(vboIndex int, a VertexAttribute) {
	if vboIndex < 0 || vboIndex >= len(m.vbos) {
		return
	}

	vbo := &m.vbos[vboIndex]
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo.handle)

	if a.Type == Integer || a.Type == UnsignedInteger {
		gl.VertexAttribIPointer(
			a.Index,
			a.Components,
			GetDataType(a.Type),
			a.StrideBytes,
			unsafe.Pointer(uintptr(a.OffsetBytes)),
		)
	} else {
		gl.VertexAttribPointer(
			a.Index,
			a.Components,
			GetDataType(a.Type),
			false,
			a.StrideBytes,
			unsafe.Pointer(uintptr(a.OffsetBytes)),
		)
	}

	gl.EnableVertexAttribArray(a.Index)

	if a.Divisor > 0 {
		gl.VertexAttribDivisor(a.Index, a.Divisor)
	}
}

func (m *Mesh) Bind() {
	gl.BindVertexArray(m.vao)
}

func (m *Mesh) Unbind() {
	gl.BindVertexArray(0)
}

func (m *Mesh) Draw() {

	if m.vao <= 0 || m.indexCount < 3 || m.ebo == 0 {
		return
	}
	m.Bind()

	gl.DrawElements(gl.TRIANGLES, int32(m.indexCount), gl.UNSIGNED_INT, nil)

	context.Assert("Mesh.Draw")

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

	context.Assert("Mesh.Draw")

	// update stats
	FrameStats.DrawCalls++
	FrameStats.Vertices += uint64(m.indexCount) * uint64(instances)
	FrameStats.Triangles += (uint64(m.indexCount) / 3) * uint64(instances)
}

func (m *Mesh) Delete() {
	if m == nil {
		return
	}

	if m.vao != 0 {
		gl.DeleteVertexArrays(1, &m.vao)
	}
	if len(m.vbos) > 0 {
		for _, b := range m.vbos {
			gl.DeleteBuffers(1, &b.handle)
		}
	}
	if m.ebo != 0 {
		gl.DeleteBuffers(1, &m.ebo)
	}

	logger.Infof("mesh id=%d deleted", m.vao)

	m.vbos = make([]vertexBuffer, 0)
	m.vao = 0
	m.ebo = 0
	m.indexCount = 0
}
