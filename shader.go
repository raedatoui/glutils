package glutils

import "C"
import (
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"strings"
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

func NewShader(vertFile, fragFile, geomFile string) (*Shader, error) {
	vertSrc, err := readFile(vertFile)
	if err != nil {
		return nil, err
	}

	fragSrc, err := readFile(fragFile)
	if err != nil {
		return nil, err
	}

	var geomSrc []byte
	if geomFile != "" {
		geomSrc, err = readFile(geomFile)
		if err != nil {
			return nil, err
		}
	}

	p, err := createProgram(vertSrc, fragSrc, geomSrc)
	if err != nil {
		return nil, err
	}

	return 	setupShader(p), nil
}

func setupShader(program uint32) *Shader {
	var (
		c, b, s int32
		i uint32
		n uint8
	)
	b = 255
	uniforms := make(map[string]int32)
	attributes := make(map[string]uint32)

	gl.GetProgramiv(program, gl.ACTIVE_UNIFORMS, &c)
	for i = 0; i < uint32(c); i++ {
		gl.GetActiveUniform(program, i, b, nil, &s, nil, &n)
		loc := gl.GetUniformLocation(program, &n)
		name := gl.GoStr(&n)
		fmt.Println(name, loc)
		uniforms[name] = loc
	}
	fmt.Println("---")
	gl.GetProgramiv(program, gl.ACTIVE_ATTRIBUTES, &c)
	for i = 0; i < uint32(c); i++ {
		gl.GetActiveAttrib(program, i, b, nil, nil, nil, &n)
		loc := gl.GetAttribLocation(program, &n)
		name := gl.GoStr(&n)
		fmt.Println(name, loc)
		attributes[name] = uint32(loc)
	}

	return &Shader{
		Program: program,
		Uniforms: uniforms,
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
		defer deleteShader(p, vertex)
	}

	if val, err := compileShader(string(f)+"\x00", gl.FRAGMENT_SHADER); err != nil {
		return 0, err
	} else {
		frag = val
		defer deleteShader(p, frag)
	}

	if len(g) > 0 {
		if val, err := compileShader(string(g)+"\x00", gl.GEOMETRY_SHADER); err != nil {
			return 0, err
		} else {
			geom = val
			defer deleteShader(p, geom)
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


type VertexArray struct {
	Data []float32
	Indices []uint32
	Stride int32
	Normalized bool
	DrawMode uint32
	Attributes map[uint32]int32 //map attrib loc to size
	Vao, vbo, ebo uint32
}

func (v *VertexArray) Setup () {
	gl.GenVertexArrays(1, &v.Vao)
	gl.GenBuffers(1, &v.vbo)

	gl.BindVertexArray(v.Vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, v.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(v.Data) * GL_FLOAT32_SIZE, gl.Ptr(v.Data), v.DrawMode)

	if len(v.Indices) > 0 {
		gl.GenBuffers(1, &v.ebo)
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, v.ebo)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(v.Indices) * GL_FLOAT32_SIZE, gl.Ptr(v.Indices), v.DrawMode)
	}

	i := 0
	for loc, size := range v.Attributes {
		gl.VertexAttribPointer(loc, size, gl.FLOAT, v.Normalized, v.Stride * GL_FLOAT32_SIZE, gl.PtrOffset(i * GL_FLOAT32_SIZE))
		gl.EnableVertexAttribArray(loc)
		i += int(size)
	}
	gl.BindVertexArray(0)
}