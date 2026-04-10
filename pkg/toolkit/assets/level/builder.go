package level

import (
	mgl "github.com/go-gl/mathgl/mgl32"
)

// 68 bytes
type LevelVertex struct {
	Position         mgl.Vec3
	TexCoord         mgl.Vec2
	Normal           mgl.Vec3
	TexIndex         uint16
	EmiIndex         uint16
	Color            mgl.Vec3
	EmissiveStrength float32
	SectorID         uint32
}

type LevelModel struct {
	Vertices []LevelVertex
	Indices  []uint32
}

type LevelBuilder struct {
	def   *LevelDef
	model *LevelModel
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

}
