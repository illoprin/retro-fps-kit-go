package leveldata

import (
	"cmp"
	"fmt"
	"slices"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/lights"
)

type LevelEntityType uint32

// -1 means null surface
type SurfaceIndex int32

// -1 means null sector
type SectorIndex int32

const (
	PlayerStart LevelEntityType = iota
	Trigger
	Prop    // solid or hollow decorative
	Pickups // weapons, aid and other stuff
	NPC     // hostile or friendly alive
)

type LevelDef struct {
	// metadata
	Name   string
	Author string

	// sound
	Music    string
	Ambience string

	// geometry
	Surfaces []Surface // used for sector coloring
	Vertices []mgl.Vec2
	Walls    []Wall
	Sectors  []Sector

	// lights
	PointLights []lights.PointLight
	SpotLights  []lights.SpotLight
	Ambient     lights.AmbientLight

	// entities
	Props   map[string]PropDef
	Entites []EntityDef
}

// ======== Level Geometry and Surfaces ========

type Surface struct {
	DifFile     string   // diffuse texture file
	EmiFile     string   // emissive texture  file
	EmiStrength float32  // emissive strength
	Color       mgl.Vec3 // you can color the texture (or use just a color)
}

type Wall struct {
	V1, V2 int // index of vertex

	MSurf SurfaceIndex // middle surface
	LSurf SurfaceIndex // lower (for portal)
	USurf SurfaceIndex // Upper (for portal)

	FrontSector SectorIndex
	BackSector  SectorIndex

	Portal *Sector
}

type Sector struct {
	Sub []SectorIndex // holes or solids in sector (e. g. room)

	FloorHeight   float32
	CeilingHeight float32

	FloorSurf   SurfaceIndex
	CeilingSurf SurfaceIndex
	Dynamic     bool
	// if the sector is a hole...
	// the normals to the walls will point inside
	// otherwise, the normals will point outward.
	// Hole bool
}

// ======== Entities ========

type EntityDef struct {
	Name string
	Type LevelEntityType
	Def  string // entity rule defined in gamedef.yaml (game manifest/rules)
	// or you can use defined props (in level file)
	Pos mgl.Vec3
	Rot mgl.Vec3
	Scl mgl.Vec3
}

type PropDef struct {
	Obj         string
	Diff        string
	Emi         string
	EmiStrength float32
}

// GetPlayerStart returns entity type PlayerStart
func (l *LevelDef) GetPlayerStart() (*EntityDef, error) {
	index, found := slices.BinarySearchFunc(l.Entites, PlayerStart, func(e EntityDef, t LevelEntityType) int {
		return cmp.Compare(e.Type, t)
	})
	if found {
		return &l.Entites[index], nil
	}

	return nil, fmt.Errorf("not found")
}
