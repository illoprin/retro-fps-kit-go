package app

import (
	"github.com/illoprin/retro-fps-kit-go/pkg/app/config"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/monitor"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/window"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
)

// AppAPI provides access to engine-level subsystems and shared resources.
// It is passed to AppState during initialization to allow states to
// interact with the window, input, rendering pipeline, and performance monitoring.
type AppAPI interface {
	// SetActiveState - switches the current application state to a new one.
	SetActiveState(AppState)

	// GetTime - returns the total time elapsed since the application started.
	GetTime() float64

	// GetInputManager - provides access to the input system for checking key and mouse states.
	GetInputManager() *window.InputManager

	// GetMonitor - returns the performance monitor containing FPS and frame timing data.
	GetMonitor() *monitor.Monitor

	// GetGBuffer - provides access to the geometry buffer for custom rendering operations.
	GetGBuffer() *rhi.Framebuffer

	// GetConfig - returns the current engine and application configuration settings.
	GetConfig() *config.Config

	// GetWindow - provides access to the native window instance and its properties.
	GetWindow() *window.Window

	// GUIWantCaptureMouse - returns true if the UI (ImGui) is currently capturing mouse input.
	GUIWantCaptureMouse() bool

	// SetCursor - sets current cursor appearance
	SetCursor(CursorType)

	// Close - signals the application to terminate and close the window.
	Close()
}
