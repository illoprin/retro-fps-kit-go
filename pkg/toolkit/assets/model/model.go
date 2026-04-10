package modeldata

// ModelVertex describes one vertex of object (pos, uv, normal)
type ModelVertex struct {
	X, Y, Z    float32
	U, V       float32
	Nx, Ny, Nz float32
}

// Model describes 3D object mesh
type Model struct {
	Vertices []ModelVertex
	Indices  []uint32
}
