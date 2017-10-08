package glutils

import "C"
import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
)

func BasicProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func NewShader(vertFile, fragFile, geomFile string) (Shader, error) {
	var shader Shader
	vertSrc, err := readFile(vertFile)
	if err != nil {
		return shader, err
	}

	fragSrc, err := readFile(fragFile)
	if err != nil {
		return shader, err
	}

	var geomSrc []byte
	if geomFile != "" {
		geomSrc, err = readFile(geomFile)
		if err != nil {
			return shader, err
		}
	}

	p, err := createProgram(vertSrc, fragSrc, geomSrc)
	if err != nil {
		return shader, err
	}
	shader = setupShader(p)
	return shader, nil
}

func setupShader(program uint32) Shader {
	var (
		c int32
		i uint32
	)
	gl.UseProgram(program)
	uniforms := make(map[string]int32)
	attributes := map[string]uint32{} //make(map[string]uint32)

	gl.GetProgramiv(program, gl.ACTIVE_UNIFORMS, &c)
	for i = 0; i < uint32(c); i++ {
		var buf [256]byte
		gl.GetActiveUniform(program, i, 256, nil, nil, nil, &buf[0])
		loc := gl.GetUniformLocation(program, &buf[0])
		name := gl.GoStr(&buf[0])
		uniforms[name] = loc
	}

	gl.GetProgramiv(program, gl.ACTIVE_ATTRIBUTES, &c)
	for i = 0; i < uint32(c); i++ {
		var buf [256]byte
		gl.GetActiveAttrib(program, i, 256, nil, nil, nil, &buf[0])
		loc := gl.GetAttribLocation(program, &buf[0])
		name := gl.GoStr(&buf[0])
		attributes[name] = uint32(loc)
	}

	return Shader{
		Program:    program,
		Uniforms:   uniforms,
		Attributes: attributes,
	}
}

func createProgram(v, f, g []byte) (uint32, error) {
	var p, vertex, frag, geom uint32
	use_geom := false

	if val, err := compileShader(string(v)+"\x00", gl.VERTEX_SHADER); err != nil {
		return 0, err
	} else {
		vertex = val
		defer func(s uint32) { deleteShader(p, s) }(vertex)
	}

	if val, err := compileShader(string(f)+"\x00", gl.FRAGMENT_SHADER); err != nil {
		return 0, err
	} else {
		frag = val
		defer func(s uint32) { deleteShader(p, s) }(frag)
	}

	if len(g) > 0 {
		if val, err := compileShader(string(g)+"\x00", gl.GEOMETRY_SHADER); err != nil {
			return 0, err
		} else {
			geom = val
			defer func(s uint32) { deleteShader(p, s) }(geom)
		}
	}

	p, err := linkProgram(vertex, frag, geom, use_geom)
	if err != nil {
		return 0, err
	}

	return p, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func deleteShader(p, s uint32) {
	gl.DetachShader(p, s)
	gl.DeleteShader(s)
}

func linkProgram(v, f, g uint32, use_geom bool) (uint32, error) {
	program := gl.CreateProgram()
	gl.AttachShader(program, v)
	gl.AttachShader(program, f)
	if use_geom {
		gl.AttachShader(program, g)
	}

	gl.LinkProgram(program)
	// check for program linking errors
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	return program, nil
}

type Shader struct {
	Program    uint32
	Uniforms   map[string]int32
	Attributes map[string]uint32
}

func (s *Shader) Delete() {
	gl.DeleteProgram(s.Program)
}

type VertexArray struct {
	Data          []float32
	Indices       []uint32
	Stride        int32
	Normalized    bool
	DrawMode      uint32
	Attributes    AttributesMap
	Vao, Vbo, Ebo uint32
}

func (v *VertexArray) Setup() {
	gl.GenVertexArrays(1, &v.Vao)
	fillVbo := true
	// Vbo already set when VertexArray was instancied.
	// This is a secondary structure using the same Vbo and vertex
	// data but with a different shader and attributes
	if v.Vbo == 0 {
		gl.GenBuffers(1, &v.Vbo)
	} else {
		fillVbo = false
	}
	if len(v.Indices) > 0 {
		gl.GenBuffers(1, &v.Ebo)
	}

	gl.BindVertexArray(v.Vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, v.Vbo)
	if fillVbo {
		gl.BufferData(gl.ARRAY_BUFFER, len(v.Data)*GL_FLOAT32_SIZE, gl.Ptr(v.Data), v.DrawMode)
	}

	if len(v.Indices) > 0 {
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, v.Ebo)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(v.Indices)*GL_FLOAT32_SIZE, gl.Ptr(v.Indices), v.DrawMode)
	}

	for loc, ss := range v.Attributes {
		gl.EnableVertexAttribArray(loc)
		gl.VertexAttribPointer(loc, int32(ss[0]), gl.FLOAT, v.Normalized, v.Stride*GL_FLOAT32_SIZE, gl.PtrOffset(ss[1]*GL_FLOAT32_SIZE))
	}
	gl.BindVertexArray(0)
}

func (v *VertexArray) Delete() {
	gl.DeleteVertexArrays(1, &v.Vao)
	gl.DeleteBuffers(1, &v.Vbo)
	if len(v.Indices) > 0 {
		gl.DeleteBuffers(1, &v.Ebo)
	}
}

type AttributesMap map[uint32][2]int //map attrib loc to size / offset

func NewAttributesMap() AttributesMap {
	return make(AttributesMap)
}
func (am AttributesMap) Add(k uint32, size, offset int) {
	am[k] = [2]int{size, offset}
}
