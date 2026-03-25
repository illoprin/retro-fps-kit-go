package player

import (
	"github.com/go-gl/glfw/v3.3/glfw"
)

// KeyState represents the current state of a key
type KeyState struct {
	Pressed      bool
	JustPressed  bool
	JustReleased bool
}

// MouseState represents the current mouse state
type MouseState struct {
	X, Y             float64
	DeltaX, DeltaY   float64
	ScrollX, ScrollY float64
	Buttons          [glfw.MouseButtonLast + 1]bool
}

// InputManager handles all input events
type InputManager struct {
	window   *glfw.Window
	keys     map[glfw.Key]KeyState
	mouse    MouseState
	gameMode bool

	// User callbacks
	userKeyCallback func(key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey)

	// Prev Callbacks
	prevKeyCallback         glfw.KeyCallback
	prevMouseButtonCallback glfw.MouseButtonCallback
	prevCursorPosCallback   glfw.CursorPosCallback
	prevScrollCallback      glfw.ScrollCallback
}

// NewManager creates a new input manager
func NewManager(window *glfw.Window) *InputManager {
	m := &InputManager{
		window:   window,
		keys:     make(map[glfw.Key]KeyState),
		mouse:    MouseState{},
		gameMode: false,
	}

	// Setup callbacks
	m.setupCallbacks()

	return m
}

// setupCallbacks registers all input callbacks
func (m *InputManager) setupCallbacks() {
	// Keyboard callback
	m.prevKeyCallback = m.window.SetKeyCallback(m.handleKey)

	// Mouse button callback
	m.prevMouseButtonCallback = m.window.SetMouseButtonCallback(m.handleMouseButton)

	// Cursor position callback
	m.prevCursorPosCallback = m.window.SetCursorPosCallback(m.handleCursorPos)

	// Scroll callback
	m.prevScrollCallback = m.window.SetScrollCallback(m.handleScroll)
}

// handleKey processes keyboard events
func (m *InputManager) handleKey(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	state, exists := m.keys[key]
	if !exists {
		state = KeyState{}
	}

	switch action {
	case glfw.Press:
		state.Pressed = true
		state.JustPressed = true
		state.JustReleased = false
	case glfw.Release:
		state.Pressed = false
		state.JustPressed = false
		state.JustReleased = true
	}

	m.keys[key] = state

	// Call prev callback if set
	if m.prevKeyCallback != nil {
		m.prevKeyCallback(w, key, scancode, action, mods)
	}
	// call user callback if set
	if m.userKeyCallback != nil {
		m.userKeyCallback(key, scancode, action, mods)
	}
}

// handleMouseButton processes mouse button events
func (m *InputManager) handleMouseButton(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		m.mouse.Buttons[button] = true
	} else {
		m.mouse.Buttons[button] = false
	}

	if m.prevMouseButtonCallback != nil {
		m.prevMouseButtonCallback(w, button, action, mods)
	}
}

// handleCursorPos processes mouse movement
func (m *InputManager) handleCursorPos(w *glfw.Window, xpos, ypos float64) {
	// Calculate delta
	m.mouse.DeltaX = xpos - m.mouse.X
	m.mouse.DeltaY = ypos - m.mouse.Y

	// Update position
	m.mouse.X = xpos
	m.mouse.Y = ypos

	if m.prevCursorPosCallback != nil {
		m.prevCursorPosCallback(w, xpos, ypos)
	}
}

// handleScroll processes mouse wheel
func (m *InputManager) handleScroll(w *glfw.Window, xoff, yoff float64) {
	m.mouse.ScrollX = xoff
	m.mouse.ScrollY = yoff

	if m.prevScrollCallback != nil {
		m.prevScrollCallback(w, xoff, yoff)
	}
}

// Update resets frame-based states (called once per frame)
func (m *InputManager) Update() {
	// Reset just-pressed and just-released states
	for key := range m.keys {
		state := m.keys[key]
		state.JustPressed = false
		state.JustReleased = false
		m.keys[key] = state
	}

	// Reset scroll delta
	m.mouse.ScrollX = 0
	m.mouse.ScrollY = 0

	// Reset mouse delta (or keep it if you want continuous movement)
	m.mouse.DeltaX = 0
	m.mouse.DeltaY = 0
}

// IsKeyPressed returns true if the key is currently held down
func (m *InputManager) IsKeyPressed(key glfw.Key) bool {
	if state, exists := m.keys[key]; exists {
		return state.Pressed
	}
	return false
}

// IsKeyJustPressed returns true if the key was pressed this frame
func (m *InputManager) IsKeyJustPressed(key glfw.Key) bool {
	if state, exists := m.keys[key]; exists {
		return state.JustPressed
	}
	return false
}

// IsKeyJustReleased returns true if the key was released this frame
func (m *InputManager) IsKeyJustReleased(key glfw.Key) bool {
	if state, exists := m.keys[key]; exists {
		return state.JustReleased
	}
	return false
}

// GetMousePosition returns current mouse cursor position
func (m *InputManager) GetMousePosition() (float64, float64) {
	return m.mouse.X, m.mouse.Y
}

// GetMouseDelta returns mouse movement since last frame
func (m *InputManager) GetMouseDelta() (float64, float64) {
	return m.mouse.DeltaX, m.mouse.DeltaY
}

// GetMouseScroll returns scroll wheel movement
func (m *InputManager) GetMouseScroll() (float64, float64) {
	return m.mouse.ScrollX, m.mouse.ScrollY
}

// IsMouseButtonPressed returns true if the mouse button is held
func (m *InputManager) IsMouseButtonPressed(button glfw.MouseButton) bool {
	return m.mouse.Buttons[button]
}

// SetCursorMode changes cursor behavior (normal, hidden, disabled)
func (m *InputManager) SetCursorMode(mode int) {
	m.window.SetInputMode(glfw.CursorMode, mode)
}

func (m *InputManager) ToggleGameMode() {
	m.gameMode = !m.gameMode

	if m.gameMode {
		m.SetCursorMode(glfw.CursorDisabled)
	} else {
		m.SetCursorMode(glfw.CursorNormal)
	}
}

func (m *InputManager) GetGameMode() bool {
	return m.gameMode
}

// SetKeyCallback sets a custom callback for keyboard events
func (m *InputManager) SetKeyCallback(callback func(key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey)) {
	m.userKeyCallback = callback
}
