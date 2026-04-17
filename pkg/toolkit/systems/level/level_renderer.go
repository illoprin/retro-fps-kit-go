package levelsys

import (
	"fmt"
	"unsafe"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/camera"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/files"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/lights"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/logger"
	mathutils "github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/math"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/rhi"
	leveldata "github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/assets/level"
)

const (
	maxSurfaces = 1024
	maxLights   = 256
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
	mesh        *rhi.Mesh
	program     *rhi.Program
	diffArr     *rhi.Texture
	emiArr      *rhi.Texture
	surfacesUbo *rhi.UniformBuffer
	pLightsUbo  *rhi.UniformBuffer
	sLightsUbo  *rhi.UniformBuffer
	surfaces    []SurfaceUniform
	level       *Level
	resources   []rhi.Resource
}

func NewLevelRenderer(l *Level) (*LevelRenderer, error) {
	r := &LevelRenderer{
		level:     l,
		resources: make([]rhi.Resource, 0, 5),
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
	r.createSurfaceUbo()

	// create mesh
	r.createMesh()

	// build lights ubos
	r.createLightsUbo()

	// update buffers
	r.Update()

	r.resources = append(r.resources,
		r.surfacesUbo,
		r.pLightsUbo,
		r.sLightsUbo,
		r.program,
		r.mesh,
		r.diffArr,
		r.emiArr,
	)

	return r, nil
}

func (r *LevelRenderer) createSurfaceUbo() {
	r.surfacesUbo = rhi.NewUniformBuffer(0)
	size := int(unsafe.Sizeof(SurfaceUniform{})) * len(r.surfaces)
	r.surfacesUbo.Bind()
	r.surfacesUbo.AllocateWithData(size, unsafe.Pointer(&r.surfaces[0]), rhi.StaticDraw)

	// create program binding
	r.program.SetAttachUniformBlock("SurfaceBlock", r.surfacesUbo)
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

	// load images
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
	diffArr, _ := rhi.NewTextureArray(diffImgs)
	// if err != nil {
	// 	return fmt.Errorf("failed to build diffuse texture array - %w", err)
	// }
	r.diffArr = diffArr

	// emission texture array
	emiArr, _ := rhi.NewTextureArray(emiImgs)
	// if err != nil {
	// 	return fmt.Errorf("failed to build emissive texture array - %w", err)
	// }
	r.emiArr = emiArr

	// create SurfaceUniform objects
	surfacesLen := mathutils.Min(len(def.Surfaces), maxSurfaces)
	r.surfaces = make([]SurfaceUniform, surfacesLen)

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

func (r *LevelRenderer) createLightsUbo() {
	r.pLightsUbo = rhi.NewUniformBuffer(1)
	r.sLightsUbo = rhi.NewUniformBuffer(2)

	// allocate
	var size int
	size = maxLights * int(unsafe.Sizeof(lights.PointLight{}))
	r.pLightsUbo.Bind()
	r.pLightsUbo.AllocateWithData(size, nil, rhi.StreamDraw)

	size = maxLights * int(unsafe.Sizeof(lights.SpotLight{}))
	r.sLightsUbo.Bind()
	r.sLightsUbo.AllocateWithData(size, nil, rhi.StreamDraw)

	// create attachmets
	r.program.SetAttachUniformBlock("PointLightsBlock", r.pLightsUbo)
	r.program.SetAttachUniformBlock("SpotLightsBlock", r.sLightsUbo)
}

func (r *LevelRenderer) updateLights() {
	def := r.level.builder.GetDef()

	if len(def.PointLights) > 0 {
		r.pLightsUbo.Bind()
		r.pLightsUbo.SetAllData(unsafe.Pointer(&def.PointLights[0]))
	}

	if len(def.SpotLights) > 0 {
		r.sLightsUbo.Bind()
		r.sLightsUbo.SetAllData(unsafe.Pointer(&def.SpotLights[0]))
	}

}

func (r *LevelRenderer) hasDiffuseArray() bool {
	return r.diffArr != nil
}

func (r *LevelRenderer) hasEmissiveArray() bool {
	return r.emiArr != nil
}

func (r *LevelRenderer) createMesh() {
	// create mesh
	r.mesh = rhi.NewMesh()

	// build layout

	// create buffers
	r.mesh.Bind()
	r.mesh.CreateVertexBuffer()
	r.mesh.CreateElementBuffer()

	// set attributes
	stride := int32(unsafe.Sizeof(leveldata.LevelVertex{}))
	texcoordOffset := int32(unsafe.Offsetof(leveldata.LevelVertex{}.TexCoord))
	normalOffset := int32(unsafe.Offsetof(leveldata.LevelVertex{}.Normal))
	surfaceIdOffset := int32(unsafe.Offsetof(leveldata.LevelVertex{}.SurfaceID))
	r.mesh.SetAttribute(0, rhi.VertexAttribute{
		Index:       0,
		Components:  3,
		Type:        rhi.Float32,
		StrideBytes: stride,
		OffsetBytes: 0,
	})
	r.mesh.SetAttribute(0, rhi.VertexAttribute{
		Index:       1,
		Components:  2,
		Type:        rhi.Float32,
		StrideBytes: stride,
		OffsetBytes: texcoordOffset,
	})
	r.mesh.SetAttribute(0, rhi.VertexAttribute{
		Index:       2,
		Components:  3,
		Type:        rhi.Float32,
		StrideBytes: stride,
		OffsetBytes: normalOffset,
	})
	r.mesh.SetAttribute(0, rhi.VertexAttribute{
		Index:       3,
		Components:  1,
		Type:        rhi.Integer,
		StrideBytes: stride,
		OffsetBytes: surfaceIdOffset,
	})

	r.mesh.Unbind()
}

func (r *LevelRenderer) Update() {

	builder := r.level.GetBuilder()
	model := builder.GetModel()

	// check dirty flag
	if builder.IsDirty() {
		logger.Warnf("before updating the mesh, build the level model - builder.BuildModel()")
		return
	}

	vertices := model.Vertices
	indices := model.Indices
	// check vertices
	if len(vertices) < 3 || len(indices) < 3 {
		logger.Warnf("empty level model")
		return
	}

	// get data byte size
	verticesSize := len(vertices) * int(unsafe.Sizeof(leveldata.LevelVertex{}))
	indicesSize := len(indices) * int(unsafe.Sizeof(uint32(0)))

	// allocate and set buffers
	r.mesh.AllocateVertexBufferWithData(0, verticesSize, unsafe.Pointer(&vertices[0]), rhi.StreamDraw)
	r.mesh.AllocateElementBuffer(indicesSize, rhi.StreamDraw)
	r.mesh.SetElementBufferData(0, indices)

	r.updateLights()
}

func (r *LevelRenderer) Render(
	w, h int, camera *camera.Camera3D, wireframe bool,
) {

	if camera == nil {
		return
	}

	r.program.Use()

	// projection
	r.program.SetMat4("u_projection", camera.GetProjection(w, h))
	// view
	r.program.SetMat4("u_view", camera.GetView())

	if r.hasDiffuseArray() {
		r.diffArr.BindToUnit(0)
		r.program.Set1i("u_diffuse", 0)
	}

	if r.hasEmissiveArray() {
		r.emiArr.BindToUnit(1)
		r.program.Set1i("u_emissive", 1)
	}

	r.program.Set1i("u_wireframe", mathutils.BoolToInt32(wireframe))

	// lights
	def := r.level.builder.GetDef()
	r.program.Set1ui("u_pointLightsNum", uint32(len(def.PointLights)))
	r.program.Set1ui("u_spotLightsNum", uint32(len(def.SpotLights)))

	r.mesh.Draw()
}

func (r *LevelRenderer) Delete() {
	for _, n := range r.resources {
		n.Delete()
	}
}
