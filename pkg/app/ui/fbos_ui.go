package ui

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/window"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/passes"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/pipeline"
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
	renderPasses []passes.PostProcessingPass,
	screenCfg *window.ScreenConfig,
) *FramebuffersUI {

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

	// prepare pass textures
	f.passTextures = make([]ImageTexture, 0)
	for _, p := range renderPasses {

		if p.GetName() == "ssao" {
			ssao := p.(*passes.SSAOPass)
			rawSSAO := ImageTexture{
				ID:   ssao.GetOcclusion().ID,
				Name: "ssao.raw",
			}
			noiseSSAO := ImageTexture{
				ID:   ssao.GetNoise().ID,
				Name: "ssao.noise",
			}
			blurSSAO := ImageTexture{
				ID:   ssao.GetBlur().ID,
				Name: "ssao.blur",
			}
			f.passTextures = append(f.passTextures, rawSSAO, noiseSSAO, blurSSAO)
		}

		if p.GetName() == "crease" {
			crease := p.(*passes.CavityPass)
			rawCrease := ImageTexture{
				ID:   crease.GetOcclusion().ID,
				Name: "crease.raw",
			}
			creaseBlur := ImageTexture{
				ID:   crease.GetBlur().ID,
				Name: "crease.blur",
			}
			f.passTextures = append(f.passTextures, rawCrease, creaseBlur)
		}

		t := ImageTexture{
			ID:   p.GetColor().ID,
			Name: p.GetName(),
		}
		f.passTextures = append(f.passTextures, t)
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
