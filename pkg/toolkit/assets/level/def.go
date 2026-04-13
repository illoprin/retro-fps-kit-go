package levelasset

import (
	"cmp"
	"fmt"
	"slices"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/entities/lights"
)

type LevelEntityType uint32
type SurfaceDefIndex uint32

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
	Surfaces []SurfaceDef // used for sector coloring
	Vertices []mgl.Vec2
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

type SurfaceDef struct {
	DifFile     string   // diffuse texture file
	EmiFile     string   // emissive texture  file
	EmiStrength float32  // emissive strength
	Color       mgl.Vec3 // you can color the texture (or use just a color)
}

type Wall struct {
	V1, V2 int // index of vertex

	Surf  SurfaceDefIndex // regular wall
	LSurf SurfaceDefIndex // lower (portal)
	USurf SurfaceDefIndex // Upper (portal)

	Portal *Sector
}

type Sector struct {
	Walls []Wall

	FloorHeight   float32
	CeilingHeight float32

	FloorSurf   SurfaceDefIndex
	CeilingSurf SurfaceDefIndex
	Dynamic     bool
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
