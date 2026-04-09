package modeldata

import (
	"unsafe"

	"github.com/illoprin/retro-fps-toolkit-go/pkg/render/rhi"
)

func SetupMeshFromModel(msh *rhi.Mesh, mod *Model) {

	if len(mod.Vertices) == 0 || len(mod.Indices) < 3 {
		return
	}

	stride := unsafe.Sizeof(mod.Vertices[0])
	vertsSize := int(stride) * len(mod.Vertices)
	indicesSize := int(unsafe.Sizeof(mod.Indices[0])) * len(mod.Indices)

	msh.Bind() // CREATE BINDING

	// buffer for position, texcoord, normal
	vboIndex := msh.CreateVertexBuffer()
	msh.AllocateVertexBuffer(vboIndex, vertsSize, rhi.StaticDraw)
	msh.SetVertexBufferData(vboIndex, 0, vertsSize, unsafe.Pointer(&mod.Vertices[0]))

	// indices
	msh.CreateElementBuffer()
	msh.AllocateElementBuffer(indicesSize, rhi.StaticDraw)
	msh.SetElementBufferData(0, mod.Indices)

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

}
