package widget

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/ww24/desktop-image-widget/asset"
	"github.com/ww24/desktop-image-widget/gfx"
	"github.com/ww24/desktop-image-widget/helper"
)

const (
	defaultWindowSize = 400
)

type Widget struct {
	texture       *gfx.Texture
	resetHandlers []func(*glfw.Window)
}

func NewWidget() *Widget {
	return new(Widget)
}

func init() {
	// GLFW event handling must be run on the main OS thread.
	runtime.LockOSThread()
}

func (w *Widget) Run(ctx context.Context) (err error) {
	defer func() {
		e := w.recover()
		if err == nil {
			err = e
		}
	}()

	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.TransparentFramebuffer, glfw.True)
	glfw.WindowHint(glfw.Floating, glfw.True)
	glfw.WindowHint(glfw.Decorated, glfw.False)
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Samples, 8)
	window, err := glfw.CreateWindow(defaultWindowSize, defaultWindowSize, "image widget", nil, nil)
	if err != nil {
		log.Fatalln("failed to create window:", err)
	}

	window.MakeContextCurrent()
	helper.NewWindowMoveHandler(window)
	helper.NewWindowCloseHandler(window, glfw.KeyEscape)
	helper.NewWindowDropHandler(window, w)
	scaleHandler := helper.NewWindowScaleHandler(window)
	w.resetHandlers = append(w.resetHandlers, func(w *glfw.Window) { scaleHandler.Reset(w) })

	if err := gl.Init(); err != nil {
		log.Fatalln("failed to init gl:", err)
	}

	gl.Enable(gl.MULTISAMPLE)

	// the linked shader program determines how the data will be rendered
	vertShader, err := gfx.NewShader(gfx.Vertex, gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	fragShader, err := gfx.NewShader(gfx.Fragment, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	shaderProgram, err := gfx.NewProgram(vertShader, fragShader)
	if err != nil {
		return err
	}
	defer shaderProgram.Delete()

	var s float32 = 1.0
	vertices := []float32{
		// top left
		-s, s, 0.0, // position
		0.0, 0.0, // texture coordinates

		// top right
		s, s, 0.0,
		1.0, 0.0,

		// bottom right
		s, -s, 0.0,
		1.0, 1.0,

		// bottom left
		-s, -s, 0.0,
		0.0, 1.0,
	}

	indices := []uint32{
		// rectangle
		0, 1, 2, // top triangle
		0, 2, 3, // bottom triangle
	}
	vao := w.createVAO(vertices, indices)

	if err := w.SetDefaultTexture(window); err != nil {
		return err
	}

	for !window.ShouldClose() {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := w.programLoop(window, shaderProgram, vao); err != nil {
				log.Fatalf("Error: %v\n", err)
			}
		}
	}

	return nil
}

func (w *Widget) recover() error {
	if cause := recover(); cause != nil {
		if e, ok := cause.(error); ok {
			return e
		} else {
			return fmt.Errorf("%v", cause)
		}
	}
	return nil
}

func (w *Widget) programLoop(window *glfw.Window, shaderProgram *gfx.Program, vao uint32) error {
	// poll events and call their registered callbacks
	glfw.PollEvents()

	// background color
	gl.ClearColor(.0, .0, .0, .0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	// draw vertices
	shaderProgram.Use()

	// set texture0 to uniform0 in the fragment shader
	w.texture.Bind(gl.TEXTURE0)
	if err := w.texture.SetUniform(shaderProgram.GetUniformLocation("texture0")); err != nil {
		return err
	}

	gl.BindVertexArray(vao)
	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
	gl.BindVertexArray(0)

	w.texture.UnBind()

	// swap in the rendered buffer
	window.SwapBuffers()

	return nil
}

func (w *Widget) reset(window *glfw.Window) {
	for _, f := range w.resetHandlers {
		f(window)
	}
}

func (w *Widget) SetDefaultTexture(window *glfw.Window) error {
	img, _, err := image.Decode(bytes.NewReader(asset.DefaultImage))
	if err != nil {
		return err
	}

	if err := w.replaceTexture(window, img); err != nil {
		return err
	}

	return nil
}

func (w *Widget) ReplaceTexture(window *glfw.Window, imageFile string) error {
	img, err := w.loadImage(imageFile)
	if err != nil {
		return err
	}

	if err := w.replaceTexture(window, img); err != nil {
		return err
	}

	return nil
}

func (w *Widget) replaceTexture(window *glfw.Window, img image.Image) error {
	texture, err := gfx.NewTexture(img, gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		return err
	}

	w.texture = texture

	window.SetSize(img.Bounds().Dx(), img.Bounds().Dy())
	w.reset(window)

	return nil
}

func (w *Widget) loadImage(file string) (image.Image, error) {
	f, err := os.Open(file)
	if os.IsNotExist(err) {
		file = filepath.Join(filepath.Dir(os.Args[0]), file)
		f, err = os.Open(file)
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// createCAO creates the Vertex Array Object for a triangle.
func (w *Widget) createVAO(vertices []float32, indices []uint32) uint32 {
	var VAO uint32
	gl.GenVertexArrays(1, &VAO)

	var VBO uint32
	gl.GenBuffers(1, &VBO)

	var EBO uint32
	gl.GenBuffers(1, &EBO)

	// Bind the Vertex Array Object first, then bind and set vertex buffer(s) and attribute pointers()
	gl.BindVertexArray(VAO)

	// copy vertices data into VBO (it needs to be bound first)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// copy indices into element buffer
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, EBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// size of one whole vertex (sum of attrib sizes)
	const stride = 3*4 + 2*4
	var offset int

	// position
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(0)
	offset += 3 * 4

	// texture position
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(1)

	// unbind the VAO (safe practice so we don't accidentally (mis)configure it later)
	gl.BindVertexArray(0)

	return VAO
}
