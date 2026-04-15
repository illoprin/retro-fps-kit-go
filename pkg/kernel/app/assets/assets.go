package assets

import (
	"fmt"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/camera"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/files"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/context"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/rhi"
)

type DefaultAssets struct {
	MeshQuad  *rhi.Mesh // 1 comp (location = 0) in_position [-1, 1]
	ProgGrid  *rhi.Program
	TexNull   *rhi.Texture
	resources []rhi.Resource
}

func NewDefaultAssets() (*DefaultAssets, error) {
	a := &DefaultAssets{}

	if err := a.initPrograms(); err != nil {
		return nil, err
	}

	if err := a.initTextures(); err != nil {
		a.Delete()
		return nil, err
	}

	a.initMeshes()

	return a, nil
}

func (a *DefaultAssets) initPrograms() (err error) {
	a.ProgGrid, err = rhi.NewProgram(
		files.GetShaderPath("initial/grid.vert"),
		files.GetShaderPath("initial/grid.frag"),
	)
	if err != nil {
		return fmt.Errorf("failed to load grid program - %w", err)
	}

	a.resources = append(a.resources, a.ProgGrid)
	return nil
}

func (a *DefaultAssets) initMeshes() {
	a.MeshQuad = rhi.NewMesh()
	rhi.SetupBasicQuadMesh(a.MeshQuad)
	a.resources = append(a.resources, a.MeshQuad)
}

// DrawGrid renders xz plane with grid pattern
// smallCellSize - size in units
// bigCellSize - size in units
// !!! CAMERA PROJECTION VIEW matrices must be updated
func (a *DefaultAssets) DrawGrid(camera *camera.Camera3D, gridY float32, smallCellSize float32, bigCellSize float32) {
	a.ProgGrid.Use()

	// projection * view
	pv := camera.Projection.Mul4(camera.View)
	a.ProgGrid.SetMat4("u_pv", pv)

	// grid pos
	gridPos := mgl.Vec3{camera.Position[0], gridY, camera.Position[2]}
	a.ProgGrid.Set3f("u_grid_pos", gridPos)

	// grid params
	a.ProgGrid.Set1f("u_bigcell", bigCellSize)
	a.ProgGrid.Set1f("u_smallcell", smallCellSize)

	context.SetFaceCulling(false)
	a.MeshQuad.Draw()
	context.SetFaceCulling(true)
}

func (a *DefaultAssets) initTextures() error {
	// a.resources = append(a.resources, a.TexNull)
	return nil
}

func (a *DefaultAssets) Delete() {
	for _, r := range a.resources {
		if r != nil {
			r.Delete()
		}
	}
}
