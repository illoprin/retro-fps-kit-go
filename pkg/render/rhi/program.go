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
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Program represents a compiled and linked shader program
type Program struct {
	handle           uint32
	uniformNameCache map[string]int32
}

// NewProgram creates a new shader program from vertex and fragment shader files
func NewProgram(vertexPath, fragmentPath string) (*Program, error) {

	prog := &Program{
		handle:           0,
		uniformNameCache: make(map[string]int32),
	}

	// load vertex shader
	vertShader, err := loadShader(vertexPath, gl.VERTEX_SHADER)
	if err != nil {
		return nil, fmt.Errorf("failed to load vertex shader: %w", err)
	}

	// load fragment shader
	fragShader, err := loadShader(fragmentPath, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, fmt.Errorf("failed to load fragment shader: %w", err)
	}

	// create program
	program := gl.CreateProgram()
	gl.AttachShader(program, vertShader)
	gl.AttachShader(program, fragShader)
	gl.LinkProgram(program)

	// check link status
	var linkStatus int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &linkStatus)
	if linkStatus == gl.FALSE {
		// get log length
		var infoLogLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &infoLogLength)
		log := gl.Str(strings.Repeat("\x00", int(infoLogLength)))

		// print log and delete program
		gl.GetProgramInfoLog(program, infoLogLength, nil, log)
		gl.DeleteProgram(program)
		return nil, fmt.Errorf("program link error: %s", gl.GoStr(log))
	}
	prog.handle = program

	// release resources
	gl.DetachShader(program, vertShader)
	gl.DetachShader(program, fragShader)
	gl.DeleteShader(vertShader)
	gl.DeleteShader(fragShader)

	return prog, nil
}

func (p *Program) getUniformLocation(name string) int32 {
	// check uniform in cache
	if loc, exists := p.uniformNameCache[name]; exists {
		return loc
	}

	// get uniform location
	cName := gl.Str(name + "\x00")
	loc := gl.GetUniformLocation(p.handle, cName)

	// cache result (even if -1)
	p.uniformNameCache[name] = loc

	// print warning
	if loc == -1 {
		slog.Warn("undefined uniform name %s", name)
	}
	return loc
}

func (p *Program) Set1i(name string, v int32) {
	if loc := p.getUniformLocation(name); loc != -1 {
		gl.Uniform1i(loc, v)
	}
}
func (p *Program) Set1f(name string, v float32) {
	if loc := p.getUniformLocation(name); loc != -1 {
		gl.Uniform1f(loc, v)
	}
}

func (p *Program) Set2f(name string, v mgl32.Vec2) {
	if loc := p.getUniformLocation(name); loc != -1 {
		gl.Uniform2fv(loc, 1, &v[0])
	}
}

func (p *Program) Set3f(name string, v mgl32.Vec3) {
	if loc := p.getUniformLocation(name); loc != -1 {
		gl.Uniform3fv(loc, 1, &v[0])
	}
}

func (p *Program) Set4f(name string, v mgl32.Vec4) {
	if loc := p.getUniformLocation(name); loc != -1 {
		gl.Uniform4fv(loc, 1, &v[0])
	}
}

func (p *Program) SetMat2(name string, m mgl32.Mat2) {
	if loc := p.getUniformLocation(name); loc != -1 {
		gl.UniformMatrix2fv(loc, 1, false, &m[0])
	}
}

func (p *Program) SetMat4(name string, m mgl32.Mat4) {
	if loc := p.getUniformLocation(name); loc != -1 {
		gl.UniformMatrix4fv(loc, 1, false, &m[0])
	}
}

// Use activates this shader program for rendering
func (p *Program) Use() {
	gl.UseProgram(p.handle)
}

// Delete frees the shader program resources
func (p *Program) Delete() {
	if p.handle > 0 {
		gl.DeleteProgram(p.handle)
	}
}

// loadShader reads and compiles a single shader from file
func loadShader(filePath string, shaderType uint32) (uint32, error) {
	// read file
	srcBytes, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read shader file: %w", err)
	}

	// add eof terminator for C string
	src := string(srcBytes) + "\x00"
	glSrc, freeSrc := gl.Strs(src)
	defer freeSrc()

	// create and compile shader
	handle := gl.CreateShader(shaderType)
	gl.ShaderSource(handle, 1, glSrc, nil)
	gl.CompileShader(handle)

	// get compile status
	var status int32
	gl.GetShaderiv(handle, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		// get log length
		var logLen int32
		gl.GetShaderiv(handle, gl.INFO_LOG_LENGTH, &logLen)
		log := gl.Str(strings.Repeat("\x00", int(logLen)))

		// print log and delete shader
		gl.GetShaderInfoLog(handle, logLen, nil, log)
		gl.DeleteShader(handle)
		return 0, fmt.Errorf("shader compilation error: %s", gl.GoStr(log))
	}

	return handle, nil
}
