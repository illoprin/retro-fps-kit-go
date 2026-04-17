package leveldata

import (
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/geometry"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/logger"
	"github.com/rclancey/go-earcut"
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

	// build sector flats (floors and ceilings)
	for i, _ := range b.def.Sectors {
		b.buildSectorFlats(SectorIndex(i))
	}

	b.dirty = false

	logger.Infof("level model built")
}

func (b *LevelBuilder) getSector(idx SectorIndex) *Sector {
	if idx < 0 || int(idx) >= len(b.def.Sectors) {
		return nil
	}
	return &b.def.Sectors[idx]
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
			b.addWallQuad(p1, p2, frontSector.CeilingHeight, backSector.CeilingHeight, w.USurf, true)
		} else {
			// build wall from back.ceiling to front.ceiling
			b.addWallQuad(p1, p2, backSector.CeilingHeight, frontSector.CeilingHeight, w.USurf, false)
		}

		// build lower wall
		if backSector.FloorHeight > frontSector.FloorHeight {
			// build wall from front.floor to back.floor
			b.addWallQuad(p1, p2, frontSector.FloorHeight, backSector.FloorHeight, w.LSurf, false)
		} else {
			// build wall from back.floor to front.floor
			b.addWallQuad(p1, p2, backSector.FloorHeight, frontSector.FloorHeight, w.LSurf, true)
		}

		// build mid wall
		if w.MSurf > -1 {
			b.addWallQuad(p1, p2, backSector.FloorHeight, backSector.CeilingHeight, w.MSurf, true)
		}
		return
	}

	// solid wall
	if frontSector != nil && backSector == nil {
		// build mid wall
		b.addWallQuad(p1, p2, frontSector.FloorHeight, frontSector.CeilingHeight, w.MSurf, false)
	}

}

func (b *LevelBuilder) getSectorSegments(idx SectorIndex) []geometry.Segment {

	s := b.getSector(idx)
	segments := make([]geometry.Segment, 0)

	nested := s.Sub
	for _, w := range b.def.Walls {
		isNested := false
		if len(nested) > 0 {
			for _, n := range nested {
				if idx == n {
					logger.Errorf("the pointer to the nested sector points to the parent")
					continue
				}
				if w.FrontSector == SectorIndex(n) || w.BackSector == SectorIndex(n) {
					isNested = true
				}
			}
		}

		isSectorWall := (w.FrontSector == idx || w.BackSector == idx) && !isNested
		if isSectorWall {
			var s geometry.Segment
			if w.FrontSector == idx {
				s = geometry.Segment{
					P1: b.def.Vertices[w.V1],
					P2: b.def.Vertices[w.V2],
				}
			} else if w.BackSector == idx {
				s = geometry.Segment{
					P1: b.def.Vertices[w.V2],
					P2: b.def.Vertices[w.V1],
				}
			}
			segments = append(segments, s)
		}
	}
	return segments
}

func (b *LevelBuilder) buildSectorFlats(idx SectorIndex) {
	// get sector
	s := b.getSector(idx)

	// has no floor and ceiling
	if s.CeilingSurf < 0 && s.FloorSurf < 0 {
		return
	}

	// build sector countour
	parentSegments := b.getSectorSegments(idx)

	if len(parentSegments) < 3 {
		return
	}

	// retruns points of polygon in CCW
	contour := geometry.BuildPolygon(parentSegments)

	// build sector holes (from nested sectors)
	var floorHoles [][]mgl.Vec2
	var ceilHoles [][]mgl.Vec2

	for _, subIdx := range s.Sub {
		sub := b.def.Sectors[subIdx]

		// get sub-sector countour
		// NOTE: I hope earcut will digest CCW for holes
		subSegments := b.getSectorSegments(subIdx)
		holeContour := geometry.BuildPolygon(subSegments)

		// need a hole in the floor?
		if sub.FloorHeight < s.FloorHeight {
			floorHoles = append(floorHoles, holeContour)
		}

		// need a hole in ceil?
		if sub.CeilingHeight > s.CeilingHeight {
			ceilHoles = append(ceilHoles, holeContour)
		}
	}

	// add flats
	b.addFlat(contour, floorHoles, s.FloorHeight, s.FloorSurf, true)
	b.addFlat(contour, ceilHoles, s.CeilingHeight, s.CeilingSurf, false)

}

func (b *LevelBuilder) addFlat(
	contour []mgl.Vec2,
	holes [][]mgl.Vec2,
	height float32,
	surf SurfaceIndex,
	isFloor bool,
) {
	if surf < 0 {
		return
	}

	startIndex := uint32(len(b.model.Vertices))

	// build vertices list
	var verticesRaw []float64
	var holeIndices []int
	var allPoints []mgl.Vec2 // save vec2 to create LevelVertex

	// add external contour
	for _, v := range contour {
		verticesRaw = append(verticesRaw, float64(v[0]), float64(v[1]))
		allPoints = append(allPoints, v)
	}

	// build holes list
	for _, hole := range holes {
		// remember where this hole in verticesRaw array begins
		holeIndices = append(holeIndices, len(verticesRaw)/2)
		for _, v := range hole {
			verticesRaw = append(verticesRaw, float64(v[0]), float64(v[1]))
			allPoints = append(allPoints, v)
		}
	}

	// triangulate
	indices, _ := earcut.Earcut(verticesRaw, holeIndices, 2)

	// normal
	normal := mgl.Vec3{0, 1, 0}
	if !isFloor {
		normal = normal.Mul(-1)
	}

	// positions
	for _, v := range allPoints {
		vertex := LevelVertex{
			Position:  mgl.Vec3{v[0], height, v[1]},
			TexCoord:  mgl.Vec2{v[0], v[1]},
			Normal:    normal,
			SurfaceID: int32(surf),
		}
		b.model.Vertices = append(b.model.Vertices, vertex)
	}

	// add indices
	for i := 0; i < len(indices); i += 3 {
		i1, i2, i3 := uint32(indices[i]), uint32(indices[i+1]), uint32(indices[i+2])
		if isFloor {
			// invert indices order for floor (CW -> CCW)
			b.model.Indices = append(b.model.Indices, startIndex+i1, startIndex+i3, startIndex+i2)
		} else {
			b.model.Indices = append(b.model.Indices, startIndex+i1, startIndex+i2, startIndex+i3)
		}
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
	normal := mgl.Vec3{-dir.Y(), 0, dir.X()}

	if isExternal {
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

	// add positions
	for i := 0; i < 4; i++ {
		b.model.Vertices = append(b.model.Vertices, LevelVertex{
			Position:  positions[i],
			TexCoord:  uvs[i],
			Normal:    normal,
			SurfaceID: int32(surfId),
		})
	}

	// build indices
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

	// add indices
	for _, i := range indices {
		b.model.Indices = append(b.model.Indices, startIndex+i)
	}
}
