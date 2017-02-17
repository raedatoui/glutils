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

	program, err := createProgram(vertSrc, fragSrc, geomSrc)
	if err != nil {
		return nil, err
	}

	var count int32
	var i uint32
	var s, b  int32
	b = 255
	var t uint32
	var n uint8

	uniforms := make(map[string]int32)
	attributes := make(map[string]uint32)

	gl.GetProgramiv(program, gl.ACTIVE_UNIFORMS, &count)
	for i = 0; i < uint32(count); i++ {
		gl.GetActiveUniform(program, i, b, nil, &s, &t, &n)
		loc := gl.GetUniformLocation(program, &n)
		name := gl.GoStr(&n)
		fmt.Println(name, loc)
		uniforms[name] = loc
	}

	gl.GetProgramiv(program, gl.ACTIVE_ATTRIBUTES, &count)
	for i = 0; i < uint32(count); i++ {
		gl.GetActiveAttrib(program, i, b, nil, &s, &t, &n)
		loc := gl.GetAttribLocation(program, &n)
		name := gl.GoStr(&n)
		fmt.Println(name, loc)
		attributes[name] = uint32(loc)
	}

	sh := &Shader{
		Program: program,
		Uniforms: uniforms,
		Attributes: attributes,
	}
	return sh, nil
}


func createProgram(vertexSource, fragementSource, geometrySource string) (uint32, error) {
	vertexShader, err := compileShader(string(vertexSource)+"\x00", gl.VERTEX_SHADER)
	if err != nil {
		return nil, err
	}

	fragmentShader, err := compileShader(string(fragementSource)+"\x00", gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, err
	}

	var geometryShader uint32
	if geometrySource != "" {
		geometryShader, err = compileShader(string(geometrySource), gl.GEOMETRY_SHADER)
		if err != nil {
			return nil, err
		}
	}

	program, err := linkProgram(vertexShader, fragmentShader, geometryShader)

	if err != nil {
		return nil, err
	}
	defer deleteShader(program, vertexShader)
	defer deleteShader(program, fragmentShader)


	if geometrySource != "" {
		defer deleteShader(program, geometryShader)
	}

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
	defer gl.DetachShader(p, s)
	defer gl.DeleteShader(p)
}

func linkProgram(vertexShader, fragmentShader, geometryShader uint32) (uint32, error) {
	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	if geometryShader != 0 {
		gl.AttachShader(program, geometryShader)
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
