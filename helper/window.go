package helper

import (
	"log"

	"github.com/go-gl/glfw/v3.3/glfw"
)

type WindowScaleHandler struct {
	width  float64
	height float64
	scale  float64
	lock   int
}

func NewWindowScaleHandler(window *glfw.Window) *WindowScaleHandler {
	h := &WindowScaleHandler{scale: 1.0}
	window.SetScrollCallback(h.scrollCallback)
	return h
}

func (h *WindowScaleHandler) locked() bool {
	return h.lock < 2
}

func (h *WindowScaleHandler) tryUnlock() {
	if h.locked() {
		h.lock++
	}
}

func (h *WindowScaleHandler) Reset(window *glfw.Window) {
	h.tryUnlock()
	h.scale = 1.0
	width, height := window.GetSize()
	h.width, h.height = float64(width), float64(height)
}

func (h *WindowScaleHandler) applyScale(w *glfw.Window) {
	width := int(h.width * h.scale)
	height := int(h.height * h.scale)

	// adjust center
	cw, ch := w.GetSize()
	px, py := w.GetPos()
	w.SetPos(px+(cw-width)/2, py+(ch-height)/2)

	w.SetSize(width, height)
}

func (h *WindowScaleHandler) scrollCallback(w *glfw.Window, xoff, yoff float64) {
	if h.locked() {
		return
	}

	h.scale += yoff / 200
	if h.scale < .1 {
		h.scale = .1
	}
	if h.scale > 2.0 {
		h.scale = 2.0
	}
	h.applyScale(w)
}

type WindowMoveHandler struct {
	clicked bool
	offsetX float64
	offsetY float64
}

func NewWindowMoveHandler(window *glfw.Window) *WindowMoveHandler {
	h := new(WindowMoveHandler)
	window.SetCursorPosCallback(h.cursorPosCallback)
	window.SetMouseButtonCallback(h.mouseButtonCallback)
	return h
}

func (h *WindowMoveHandler) reset() {
	h.offsetX = 0
	h.offsetY = 0
}

func (h *WindowMoveHandler) mouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if button != glfw.MouseButtonLeft {
		return
	}

	h.clicked = action != glfw.Release
	if !h.clicked {
		h.reset()
	}
}

func (h *WindowMoveHandler) cursorPosCallback(w *glfw.Window, xpos, ypos float64) {
	if !h.clicked {
		return
	}

	if h.offsetX == 0 && h.offsetY == 0 {
		h.offsetX = xpos
		h.offsetY = ypos
	}

	x, y := w.GetPos()
	x += int(xpos - h.offsetX)
	y += int(ypos - h.offsetY)
	w.SetPos(x, y)
}

type WindowCloseHandler struct {
	key glfw.Key
}

func NewWindowCloseHandler(w *glfw.Window, key glfw.Key) *WindowCloseHandler {
	h := &WindowCloseHandler{key: key}
	w.SetKeyCallback(h.keyCallback)
	return h
}

func (h *WindowCloseHandler) keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == h.key && action == glfw.Press {
		window.SetShouldClose(true)
	}
}

type TextureReplacer interface {
	ReplaceTexture(*glfw.Window, string) error
}

type WindowDropHandler struct {
	replacer TextureReplacer
}

func NewWindowDropHandler(w *glfw.Window, tr TextureReplacer) *WindowDropHandler {
	h := &WindowDropHandler{
		replacer: tr,
	}
	w.SetDropCallback(h.dropCallback)
	return h
}

func (h *WindowDropHandler) dropCallback(w *glfw.Window, names []string) {
	if len(names) == 0 {
		return
	}
	if err := h.replacer.ReplaceTexture(w, names[0]); err != nil {
		log.Printf("Error: %v\n", err)
	}
}
