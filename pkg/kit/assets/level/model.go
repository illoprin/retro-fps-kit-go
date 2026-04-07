package level

import (
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-kit-go/pkg/kit/entities/lights"
)

type LevelEntityType uint32

const (
	PlayerStart LevelEntityType = iota
	Trigger
	Prop    // solid or hollow decorative
	Pickups // weapons, aid and other stuff
	NPC     // hostile or friendly alive
)

type LevelModel struct {
	Name     string
	Author   string
	Music    string
	Ambience string

	Vertices []mgl.Vec2
	Walls    []Wall
	Sectors  []Sector

	Surfaces []SurfaceDef
	Models   map[string]PropDef

	PointLights []lights.PointLight
	SpotLights  []lights.SpotLight
	Ambient     lights.AmbientLight

	Entites []EntityDef
}

type EntityDef struct {
	Name string
	Type LevelEntityType
	Def  string // entity rule defined in gamedef.yaml (game manifest/rules)
	Surf *SurfaceDef
	Pos  mgl.Vec3
	Rot  mgl.Vec3
	Scl  mgl.Vec3
}

type PropDef struct {
	File string
	Surf *SurfaceDef
}

type SurfaceDef struct {
	DifFile     string  // diffuse texture file
	Frames      uint8   // if it is sprite
	EmiFile     string  // emissive texture  file
	EmiStrength float32 // emissive strength
}

type Wall struct {
	V1, V2 int

	Tex  *SurfaceDef // regular wall
	LTex *SurfaceDef // lower (portal)
	UTex *SurfaceDef // Upper (portal)

	Portal *Sector
}

type Sector struct {
	Walls []*Wall

	FloorHeight   float32
	CeilingHeight float32

	FloorTex   *SurfaceDef
	CeilingTex *SurfaceDef
}
