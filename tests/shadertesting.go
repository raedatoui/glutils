package main

import (
	"log"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/raedatoui/glutils"
	"runtime"
	"os"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/raedatoui/learn-opengl-golang/sections"
)

const WIDTH = 800
const HEIGHT = 800

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

func init() {
	dir, err := glutils.ImportPathToDir("github.com/raedatoui/glutils/tests")
	if err != nil {
		log.Fatalln("Unable to find Go package in your GOPATH, it's needed to load assets:", err)
	}
	if err := os.Chdir(dir); err != nil {
		log.Panicln("os.Chdir:", err)
	}
}

func main () {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, gl.TRUE)

	window, err := glfw.CreateWindow(WIDTH, HEIGHT, "", nil, nil)
	if err != nil {
		log.Fatalf("cant create window %v", err)
	}
	window.MakeContextCurrent()

	// Initialize Glow - this is the equivalent of glew
	if err := gl.Init(); err != nil {
		log.Fatalf("cant init gl %v", err)
	}

	vertices := []float32{
		-0.5, -0.5, -0.5, 0.0, 0.0,
		0.5, -0.5, -0.5, 1.0, 0.0,
		0.5, 0.5, -0.5, 1.0, 1.0,
		0.5, 0.5, -0.5, 1.0, 1.0,
		-0.5, 0.5, -0.5, 0.0, 1.0,
		-0.5, -0.5, -0.5, 0.0, 0.0,

		-0.5, -0.5, 0.5, 0.0, 0.0,
		0.5, -0.5, 0.5, 1.0, 0.0,
		0.5, 0.5, 0.5, 1.0, 1.0,
		0.5, 0.5, 0.5, 1.0, 1.0,
		-0.5, 0.5, 0.5, 0.0, 1.0,
		-0.5, -0.5, 0.5, 0.0, 0.0,

		-0.5, 0.5, 0.5, 1.0, 0.0,
		-0.5, 0.5, -0.5, 1.0, 1.0,
		-0.5, -0.5, -0.5, 0.0, 1.0,
		-0.5, -0.5, -0.5, 0.0, 1.0,
		-0.5, -0.5, 0.5, 0.0, 0.0,
		-0.5, 0.5, 0.5, 1.0, 0.0,

		0.5, 0.5, 0.5, 1.0, 0.0,
		0.5, 0.5, -0.5, 1.0, 1.0,
		0.5, -0.5, -0.5, 0.0, 1.0,
		0.5, -0.5, -0.5, 0.0, 1.0,
		0.5, -0.5, 0.5, 0.0, 0.0,
		0.5, 0.5, 0.5, 1.0, 0.0,

		-0.5, -0.5, -0.5, 0.0, 1.0,
		0.5, -0.5, -0.5, 1.0, 1.0,
		0.5, -0.5, 0.5, 1.0, 0.0,
		0.5, -0.5, 0.5, 1.0, 0.0,
		-0.5, -0.5, 0.5, 0.0, 0.0,
		-0.5, -0.5, -0.5, 0.0, 1.0,

		-0.5, 0.5, -0.5, 0.0, 1.0,
		0.5, 0.5, -0.5, 1.0, 1.0,
		0.5, 0.5, 0.5, 1.0, 0.0,
		0.5, 0.5, 0.5, 1.0, 0.0,
		-0.5, 0.5, 0.5, 0.0, 0.0,
		-0.5, 0.5, -0.5, 0.0, 1.0,
	}

	rotationAxis := mgl32.Vec3{1.0, 0.3, 0.5}.Normalize()
	cubePositions := []mgl32.Mat4{
		mgl32.Translate3D(0.0, 0.0, 0.0),
		mgl32.Translate3D(2.0, 5.0, -15.0),
		mgl32.Translate3D(-1.5, -2.2, -2.5),
		mgl32.Translate3D(-3.8, -2.0, -12.3),
		mgl32.Translate3D(2.4, -0.4, -3.5),
		mgl32.Translate3D(-1.7, 3.0, -7.5),
		mgl32.Translate3D(1.3, -2.0, -2.5),
		mgl32.Translate3D(1.5, 2.0, -2.5),
		mgl32.Translate3D(1.5, 0.2, -1.5),
		mgl32.Translate3D(-1.3, 1.0, -1.5),
	}

	shader, err := glutils.NewShader(
		"basic.vs",
		"basic.frag",
		"")
	if  err != nil {
		log.Fatalf("cant create shader %v", err)
	}

	attr := make(glutils.AttributesMap)
	attr[shader.Attributes["texCoord"]] = [2]int{2, 3}
	attr[shader.Attributes["position"]] = [2]int{3, 0}


	v := glutils.VertexArray {
		Data: vertices,
		DrawMode: gl.STATIC_DRAW,
		Stride: 5,
		Attributes: attr,
		Normalized: false,
	}

	v.Setup()


	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.ClearColor(1.0, 1.0, 1.0, 1.0)

		// Activate shader
		gl.UseProgram(shader.Program)

		// Create transformations
		view := mgl32.Translate3D(0.0, 0.0, -3.0)
		projection := mgl32.Perspective(45.0, sections.RATIO, 0.1, 100.0)

		// Get their uniform location
		modelLoc := shader.Uniforms["model"]
		viewLoc := shader.Uniforms["view"]
		projLoc := shader.Uniforms["projection"]

		// Pass the matrices to the shader

		gl.UniformMatrix4fv(viewLoc, 1, false, &view[0])
		// Note: currently we set the projection matrix each frame,
		// but since the projection matrix rarely changes it's often best practice to set it outside the main loop only once.
		gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])

		// Draw container
		gl.BindVertexArray(v.Vao)

		for i := 0; i < 10; i++ {
			// Calculate the model matrix for each object and pass it to shader before drawing
			model := cubePositions[i]

			angle := float32(glfw.GetTime()) * float32(i+1)

			model = model.Mul4(mgl32.HomogRotate3D(angle, rotationAxis))
			gl.UniformMatrix4fv(modelLoc, 1, false, &model[0])
			gl.DrawArrays(gl.TRIANGLES, 0, 36)
		}
		gl.BindVertexArray(0)

		window.SwapBuffers()
		// Poll Events
		glfw.PollEvents()
	}

}

