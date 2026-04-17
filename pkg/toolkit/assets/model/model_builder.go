package modeldata

import (
	"unsafe"

	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/rhi"
)

func NewMeshFromModel(mod *Model) (msh *rhi.Mesh) {

	if len(mod.Vertices) == 0 || len(mod.Indices) < 3 {
		return
	}

	stride := unsafe.Sizeof(mod.Vertices[0])
	vertsSize := int(stride) * len(mod.Vertices)
	indicesSize := int(unsafe.Sizeof(mod.Indices[0])) * len(mod.Indices)

	msh = rhi.NewMesh()

	// build layout
	msh.Bind()

	// create buffers
	vboIndex := msh.CreateVertexBuffer()
	msh.CreateElementBuffer()

	// position
	msh.SetAttribute(vboIndex, rhi.VertexAttribute{
		Index:       0,
		Components:  3, // vec3
		Type:        rhi.Float32,
		StrideBytes: int32(stride),
		OffsetBytes: 0,
	})

	// texcoord
	msh.SetAttribute(vboIndex, rhi.VertexAttribute{
		Index:       1,
		Components:  2, // vec2
		Type:        rhi.Float32,
		StrideBytes: int32(stride),
		OffsetBytes: 3 * rhi.SizeOfFloat32,
	})

	// normal
	msh.SetAttribute(vboIndex, rhi.VertexAttribute{
		Index:       2,
		Components:  3, // vec3
		Type:        rhi.Float32,
		StrideBytes: int32(stride),
		OffsetBytes: 5 * rhi.SizeOfFloat32,
	})

	msh.Unbind()

	msh.AllocateVertexBufferWithData(vboIndex, vertsSize, nil, rhi.StaticDraw)
	msh.SetVertexBufferData(vboIndex, 0, vertsSize, unsafe.Pointer(&mod.Vertices[0]))
	msh.AllocateElementBufferWithData(indicesSize, nil, rhi.StaticDraw)
	msh.SetElementBufferData(0, mod.Indices)

	return msh
}
