package ui

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/window"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/pipeline"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/post"
)

type FramebuffersUI struct {
	Visible bool

	// Dependencies
	screen *window.ScreenConfig

	// Textures for display
	deferredTextures []ImageTexture
	passTextures     []ImageTexture
}

func NewFramebuffersUI(
	res *pipeline.DeferredRenderResult,
	p *post.PostProcessingPipeline,
	screenCfg *window.ScreenConfig,
) *FramebuffersUI {

	passes := p.GetExecutionList()

	f := &FramebuffersUI{
		screen:  screenCfg,
		Visible: true,
	}

	// get deferred textures
	f.deferredTextures = []ImageTexture{
		{
			ID:   res.Color.ID,
			Name: "Color",
		},
		{
			ID:   res.Normal.ID,
			Name: "Normal",
		},
		{
			ID:   res.Position.ID,
			Name: "Position",
		},
		{
			ID:   res.Depth.ID,
			Name: "Depth",
		},
	}

	// build passes textures
	for _, pass := range passes {
		if dbg, ok := pass.(post.DebuggablePass); ok {
			for _, tex := range dbg.GetDebugTextures() {
				f.passTextures = append(f.passTextures, ImageTexture{
					ID:   tex.Texture.ID,
					Name: tex.Name,
				})
			}
			continue
		}

		// fallback
		f.passTextures = append(f.passTextures, ImageTexture{
			ID:   pass.GetColor().ID,
			Name: "pass",
		})
	}

	return f
}

// window shows different render targets
func (f *FramebuffersUI) showTexturesWindow() {
	imgui.Begin("Render Targets")

	if imgui.BeginTabBar("TexturesTabBar") {
		if imgui.BeginTabItem("Deferred") {
			showTextures(f.deferredTextures, f.screen.Aspect, 512)
			imgui.EndTabItem()
		}

		if imgui.BeginTabItem("Post Processing") {
			showTextures(f.passTextures, f.screen.Aspect, 512)
			imgui.EndTabItem()
		}
		imgui.EndTabBar()
	}

	imgui.End()
}

func (f *FramebuffersUI) Show() {
	if f.Visible {
		// -- Textures Window
		imgui.SetNextWindowPosV(
			imgui.Vec2{1000, 30}, imgui.CondFirstUseEver, imgui.Vec2{0, 0},
		)
		imgui.SetNextWindowSizeConstraints(
			imgui.Vec2{400, 600}, imgui.Vec2{1000, 700},
		)
		f.showTexturesWindow()
	}
}
