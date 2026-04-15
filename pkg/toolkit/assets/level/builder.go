package leveldata

import (
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/logger"
)

// 36 bytes
type LevelVertex struct {
	Position  mgl.Vec3
	TexCoord  mgl.Vec2
	Normal    mgl.Vec3
	SurfaceID int32
}

type LevelModel struct {
	Vertices []LevelVertex
	Indices  []uint32
}

type LevelBuilder struct {
	def         *LevelDef
	model       *LevelModel
	playerStart *EntityDef
	dirty       bool
}

type contour struct {
	points []float64 // x, y, x, y...
	isHole bool
}

func NewLevelBuilder(def *LevelDef) *LevelBuilder {
	b := &LevelBuilder{
		def:   def,
		dirty: true,
	}
	return b
}

func (b *LevelBuilder) GetDef() *LevelDef {
	return b.def
}

func (b *LevelBuilder) GetModel() *LevelModel {
	return b.model
}

func (b *LevelBuilder) IsDirty() bool {
	return b.dirty
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

	// build sector flats
	// for i, s := range b.def.Sectors {
	// 	b.buildSector(s, uint32(i))
	// }

	b.dirty = false

	logger.Infof("level model built")
}

func (b *LevelBuilder) getSector(idx SectorIndex) *Sector {
	if idx < 0 || int(idx) >= len(b.def.Sectors) {
		return nil
	}
	return &b.def.Sectors[idx]
}

func (b *LevelBuilder) buildSector(s Sector, sectorIndex uint32) {

}

func (b *LevelBuilder) buildWall(w Wall) {
	frontSector := b.getSector(w.FrontSector)
	backSector := b.getSector(w.BackSector)
	p1 := b.def.Vertices[w.V1]
	p2 := b.def.Vertices[w.V2]

	// portal wall
	if backSector != nil && frontSector != nil {
		// build upper wall
		if backSector.CeilingHeight > frontSector.CeilingHeight {
			// build wall from front.ceiling to back.ceiling
			b.addWallQuad(
				p1,
				p2,
				frontSector.CeilingHeight,
				backSector.CeilingHeight,
				w.USurf,
				true,
			)
		} else {
			// build wall from back.ceiling to front.ceiling
			b.addWallQuad(
				p1,
				p2,
				backSector.CeilingHeight,
				frontSector.CeilingHeight,
				w.USurf,
				false,
			)
		}

		// build lower wall
		if backSector.FloorHeight > frontSector.FloorHeight {
			// build wall from front.floor to back.floor
			b.addWallQuad(
				p1,
				p2,
				frontSector.FloorHeight,
				backSector.FloorHeight,
				w.LSurf,
				false,
			)
		} else {
			// build wall from back.floor to front.floor
			b.addWallQuad(
				p1,
				p2,
				backSector.FloorHeight,
				frontSector.FloorHeight,
				w.LSurf,
				true,
			)
		}

		// build mid wall
		if w.MSurf > -1 {
			b.addWallQuad(
				p1,
				p2,
				backSector.FloorHeight,
				backSector.CeilingHeight,
				w.MSurf,
				true,
			)
		}
		return
	}

	// solid wall
	if frontSector != nil && backSector == nil {
		// build mid wall
		b.addWallQuad(
			p1,
			p2,
			frontSector.FloorHeight,
			frontSector.CeilingHeight,
			w.MSurf,
			false,
		)
	}

}

func (b *LevelBuilder) addWallQuad(
	v1, v2 mgl.Vec2,
	h1, h2 float32,
	surfId SurfaceIndex,
	isExternal bool,
) {
	startIndex := uint32(len(b.model.Vertices))

	// wall normal
	dir := v2.Sub(v1).Normalize()
	normal := mgl.Vec3{dir.Y(), 0, -dir.X()}

	if !isExternal {
		normal = normal.Mul(-1)
	}

	// Vertices
	// 	left bottom
	// 	right bottom
	// 	right top
	// 	left top
	positions := []mgl.Vec3{
		{v2.X(), h1, v2.Y()},
		{v1.X(), h1, v1.Y()},
		{v1.X(), h2, v1.Y()},
		{v2.X(), h2, v2.Y()},
	}

	length := v2.Sub(v1).Len()
	height := h2 - h1

	uvs := []mgl.Vec2{
		{0, 0},
		{length, 0},
		{length, height},
		{0, height},
	}

	for i := 0; i < 4; i++ {
		b.model.Vertices = append(b.model.Vertices, LevelVertex{
			Position:  positions[i],
			TexCoord:  uvs[i],
			Normal:    normal,
			SurfaceID: int32(surfId),
		})
	}

	// add indices
	var indices []uint32

	if isExternal {
		// CW
		indices = []uint32{
			0, 1, 3,
			1, 2, 3,
		}
	} else {
		// CCW
		indices = []uint32{
			0, 3, 1,
			3, 2, 1,
		}
	}

	for _, i := range indices {
		b.model.Indices = append(b.model.Indices, startIndex+i)
	}
}
