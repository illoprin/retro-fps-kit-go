package rhi

import (
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/logger"
)

// UniformBuffer  - stores uniforms
type UniformBuffer struct {
	handle    uint32 // id
	sizeBytes int    // buffer size in bytes
	binding   uint32 // binding point (layout = binding)
}

// NewUniformBuffer - creates empty UBO with determined binding
func NewUniformBuffer(binding uint32) *UniformBuffer {
	var ubo uint32
	gl.GenBuffers(1, &ubo)
	logger.Infof("ubo id=%d created", ubo)

	return &UniformBuffer{
		handle:  ubo,
		binding: binding,
	}
}

// Bind - bind uniform buffer to GL_UNIFORM_BUFFER
func (u *UniformBuffer) Bind() {
	gl.BindBuffer(gl.UNIFORM_BUFFER, u.handle)
}

// Unbind - debug (not for runtime)
func (u *UniformBuffer) Unbind() {
	gl.BindBuffer(gl.UNIFORM_BUFFER, 0)
}

// Allocate - allocates ubo buffer WITHOUT data
func (u *UniformBuffer) Allocate(sizeBytes int, drawType BufferType) {
	u.sizeBytes = sizeBytes
	u.Bind()
	gl.BufferData(gl.UNIFORM_BUFFER, sizeBytes, nil, GetBufferType(drawType))

	// Привязка к binding point
	u.BindToShader(u.binding)
}

// SetData - set data to UBO
// offsetBytes - which byte to write from
// sizeBytes - how many bytes
func (u *UniformBuffer) SetData(offsetBytes int, sizeBytes int, data unsafe.Pointer) {
	if u.handle == 0 || sizeBytes <= 0 || offsetBytes < 0 {
		return
	}

	u.Bind()
	gl.BufferSubData(gl.UNIFORM_BUFFER, offsetBytes, sizeBytes, data)
}

// SetAllData - updates whole buffer
func (u *UniformBuffer) SetAllData(data unsafe.Pointer) {
	if u.handle == 0 || u.sizeBytes == 0 {
		return
	}
	u.Bind()
	gl.BufferSubData(gl.UNIFORM_BUFFER, 0, u.sizeBytes, data)
}

// BindToShader - bind UBO to shader binding point
// shaderBindingIndex - index в shader layout(binding = X)
func (u *UniformBuffer) BindToShader(shaderBindingIndex uint32) {
	gl.BindBufferBase(gl.UNIFORM_BUFFER, shaderBindingIndex, u.handle)
}

// Delete - удаляет UBO
func (u *UniformBuffer) Delete() {
	if u.handle != 0 {
		gl.DeleteBuffers(1, &u.handle)
		u.handle = 0
		u.sizeBytes = 0
	}
}
