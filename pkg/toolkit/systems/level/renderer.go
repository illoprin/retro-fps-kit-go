package levelsys

import (
	"fmt"
	"unsafe"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/files"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/rhi"
)

// 32 bytes
type SurfaceUniform struct {
	DiffIndex  int32
	EmiIndex   int32
	EmiStength float32
	_          float32 // padding
	Color      mgl.Vec4
}

type LevelRenderer struct {
	mesh      *rhi.Mesh
	program   *rhi.Program
	diffArr   *rhi.Texture
	emiArr    *rhi.Texture
	ubo       *rhi.UniformBuffer
	surfaces  []SurfaceUniform
	level     *LevelSystem
	resources []rhi.Resource
}

func NewLevelRenderer(l *LevelSystem) (*LevelRenderer, error) {
	r := &LevelRenderer{
		level: l,
	}

	// load level program
	program, err := rhi.NewProgram(
		files.GetShaderPath("deferred/d_level.vert"),
		files.GetShaderPath("deferred/d_level.frag"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load level shader - %w", err)
	}
	r.program = program

	// build texture array
	if err := r.buildSurfaces(); err != nil {
		return nil, fmt.Errorf("failed to build textures - %w", err)
	}

	// build surface ubo
	r.createUBO()

	// create and upload mesh
	r.createMesh()

	r.resources = append(r.resources, r.ubo, r.program, r.diffArr, r.emiArr, r.mesh)

	return r, nil
}

func (r *LevelRenderer) createUBO() {
	r.ubo = rhi.NewUniformBuffer(0)
	size := int(unsafe.Sizeof(SurfaceUniform{})) * len(r.surfaces)
	r.ubo.Allocate(size, rhi.StaticDraw)
	r.ubo.SetAllData(unsafe.Pointer(&r.surfaces[0]))
}

func (r *LevelRenderer) buildSurfaces() error {
	def := r.level.GetBuilder().GetDef()

	// collect all images
	diffImgs := make([]*files.RGBA8Data, 0, len(def.Surfaces))
	emiImgs := make([]*files.RGBA8Data, 0, len(def.Surfaces))

	// build a relations map (texture -> surface)
	diffMap := map[string]int{}
	emiMap := map[string]int{}

	var diffW, diffH int32 = 0, 0         // diffuse texture size
	var emiW, emiH int32 = 0, 0           // emissive texture size
	var difData, emiData *files.RGBA8Data // textures rgba8 data
	var err error                         // error
	var hasDiff, hasEmi bool              // has surface diffuse or emissive
	for i, s := range def.Surfaces {

		hasDiff = s.DifFile != ""
		hasEmi = s.EmiFile != ""

		// load diffuse
		if hasDiff {
			if _, ok := diffMap[s.DifFile]; !ok {
				// load
				difFile := files.GetTexturePath(s.DifFile)
				difData, err = files.LoadTexture(difFile)
				if err != nil {
					return err
				}

				// add
				diffMap[s.DifFile] = len(diffImgs)
				diffImgs = append(diffImgs, difData)

				// check
				if len(diffImgs) == 1 {
					diffW, diffH = difData.W, difData.H
				} else {
					if difData.W != diffW || difData.H != diffH {
						return fmt.Errorf("diffuse size (surface id=%d) mismatch", i)
					}
				}
			}
		}

		// load emission
		if hasEmi {
			if _, ok := emiMap[s.EmiFile]; !ok {
				// load
				emiFile := files.GetTexturePath(s.EmiFile)
				emiData, err = files.LoadTexture(emiFile)
				if err != nil {
					return err
				}

				// add
				emiMap[s.EmiFile] = len(emiImgs)
				emiImgs = append(emiImgs, emiData)

				// check
				if len(emiImgs) == 1 {
					emiW, emiH = emiData.W, emiData.H
				} else {
					if emiData.W != emiW || emiData.H != emiH {
						return fmt.Errorf("emissive size (surface id=%d) mismatch", i)
					}
				}
			}
		}

	}

	// diffuse texture array
	diffArr, err := rhi.NewTextureArray(diffImgs)
	r.diffArr = diffArr

	// emission texture array
	emiArr, err := rhi.NewTextureArray(emiImgs)
	r.emiArr = emiArr

	// create SurfaceUniform objects
	r.surfaces = make([]SurfaceUniform, len(def.Surfaces))

	for i, s := range def.Surfaces {
		// find texture dependency
		var dif, emi int32 = -1, -1

		// if has diffuse map - give texture index
		if diffIndex, ok := diffMap[s.DifFile]; ok {
			dif = int32(diffIndex)
		}

		// if has emission map - give texture index
		if emiIndex, ok := emiMap[s.EmiFile]; ok {
			emi = int32(emiIndex)
		}

		r.surfaces[i] = SurfaceUniform{
			DiffIndex:  dif,
			EmiIndex:   emi,
			EmiStength: s.EmiStrength,
			Color:      mgl.Vec4{s.Color[0], s.Color[1], s.Color[2], 1.0},
		}
	}

	return nil
}

func (r *LevelRenderer) hasDiffuseArray() bool {
	return r.diffArr != nil
}

func (r *LevelRenderer) hasEmissiveArray() bool {
	return r.emiArr != nil
}

func (r *LevelRenderer) createMesh() {
	r.mesh = rhi.NewMesh()
	r.Update()
}

func (r *LevelRenderer) Update() {
	r.mesh.Bind()

}

func (r *LevelRenderer) Render() {
	r.program.Use()

	if r.hasDiffuseArray() {
		r.diffArr.BindToUnit(0)
		r.program.Set1i("u_diffuse", 0)
	}

	if r.hasEmissiveArray() {
		r.emiArr.BindToUnit(1)
		r.program.Set1i("u_emissive", 1)
	}

	r.program.SetAttachUniformBlock("SurfaceBlock", r.ubo)

	// TODO lights

	r.mesh.Draw()
}

func (r *LevelRenderer) Delete() {
	for _, n := range r.resources {
		if n != nil {
			n.Delete()
		}
	}
}
