package app

import (
	"github.com/illoprin/retro-fps-kit-go/pkg/app/config"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/monitor"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/window"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
)

type AppAPI interface {
	SetActiveState(AppState)
	GetTime() float64
	GetInputManager() *window.InputManager
	GetMonitor() *monitor.Monitor
	GetGBuffer() *rhi.Framebuffer
	GetConfig() *config.Config
	GetWindow() *window.Window
	GUIWantCaptureMouse() bool
	Close()
}
