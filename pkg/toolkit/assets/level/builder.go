package leveldata

import (
	mgl "github.com/go-gl/mathgl/mgl32"
)

// 40 bytes
type LevelVertex struct {
	Position  mgl.Vec3
	TexCoord  mgl.Vec2
	Normal    mgl.Vec3
	SurfaceID uint32
	SectorID  uint32
}

type LevelModel struct {
	Vertices []LevelVertex
	Indices  []uint32
}

type LevelBuilder struct {
	def         *LevelDef
	model       *LevelModel
	playerStart *EntityDef
}

type contour struct {
	points []float64 // x, y, x, y...
	isHole bool
}

func NewLevelBuilder(def *LevelDef) *LevelBuilder {
	b := &LevelBuilder{
		def: def,
	}
	return b
}

func (b *LevelBuilder) GetDef() *LevelDef {
	return b.def
}

func (b *LevelBuilder) BuildModel() {
	b.model = &LevelModel{
		Vertices: make([]LevelVertex, 0),
		Indices:  make([]uint32, 0),
	}

	// build walls
	for _, w := range b.def.Walls {
		b.buildWall(w)
	}

	// build sectors
	for i, s := range b.def.Sectors {
		b.buildSector(s, uint32(i))
	}

}

func (b *LevelBuilder) buildSector(s Sector, sectorIndex uint32) {

}

func (b *LevelBuilder) buildWall(s Wall) {

}
