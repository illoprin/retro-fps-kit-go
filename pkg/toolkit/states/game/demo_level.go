package game

import (
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/lights"
	leveldata "github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/assets/level"
)

var (
	demoLevel leveldata.LevelDef = leveldata.LevelDef{

		Name:     "Demo",
		Author:   "illoprin",
		Music:    "...",
		Ambience: "...",
		Surfaces: []leveldata.Surface{
			// 0
			leveldata.Surface{
				DifFile:     "round_bricks.png",
				EmiFile:     "",
				EmiStrength: 0.0,
				Color:       mgl.Vec3{1.0, 1.0, 1.0},
			},
			// 1
			leveldata.Surface{
				DifFile:     "gray_tiles.png",
				EmiFile:     "",
				EmiStrength: 0.0,
				Color:       mgl.Vec3{1.0, 1.0, 1.0},
			},
			// 2
			leveldata.Surface{
				DifFile:     "corrugate.png",
				EmiFile:     "",
				EmiStrength: 0.0,
				Color:       mgl.Vec3{1.0, 1.0, 1.0},
			},
			// 3
			leveldata.Surface{
				DifFile:     "tiny_tiles.png",
				EmiFile:     "",
				EmiStrength: 0.0,
				Color:       mgl.Vec3{1.0, 1.0, 1.0},
			},
			// 4
			leveldata.Surface{
				DifFile:     "star_wall.png",
				EmiFile:     "",
				EmiStrength: 0.0,
				Color:       mgl.Vec3{1.0, 1.0, 1.0},
			},
			// 5
			leveldata.Surface{
				DifFile:     "gray_rock.png",
				EmiFile:     "",
				EmiStrength: 0.0,
				Color:       mgl.Vec3{1.0, 1.0, 1.0},
			},
		},
		Vertices: []mgl.Vec2{
			// sector 0

			{9, 6},   // 0
			{15, 6},  // 1
			{18, 9},  // 2
			{18, 24}, // 3

			// sector 1

			{15, 27}, // 4
			{15, 50}, // 5
			{9, 50},  // 6
			{9, 27},  // 7

			// sector 0

			{6, 24}, // 8
			{6, 9},  // 9

			// sector 2 (nested in sector 0)

			{13.5, 13.5}, // 10
			{13.5, 19.5}, // 11
			{12, 21},     // 12
			{10.5, 19.5}, // 13
			{10.5, 13.5}, // 14
		},
		Walls: []leveldata.Wall{
			// 0-1 wall (sector 0)
			leveldata.Wall{
				V1:          0,
				V2:          1,
				MSurf:       0,
				LSurf:       -1,
				USurf:       -1,
				FrontSector: 0,
				BackSector:  -1,
			},
			// 1-2 wall (sector 0)
			leveldata.Wall{
				V1:          1,
				V2:          2,
				MSurf:       0,
				LSurf:       -1,
				USurf:       -1,
				FrontSector: 0,
				BackSector:  -1,
			},
			// 2-3 wall (sector 0)
			leveldata.Wall{
				V1:          2,
				V2:          3,
				MSurf:       0,
				LSurf:       -1,
				USurf:       -1,
				FrontSector: 0,
				BackSector:  -1,
			},
			// 3-4 wall (sector 0)
			leveldata.Wall{
				V1:          3,
				V2:          4,
				MSurf:       0,
				LSurf:       -1,
				USurf:       -1,
				FrontSector: 0,
				BackSector:  -1,
			},
			// 4-5 wall (sector 1)
			leveldata.Wall{
				V1:          4,
				V2:          5,
				MSurf:       2,
				LSurf:       -1,
				USurf:       -1,
				FrontSector: 1,
				BackSector:  -1,
			},
			// 7-4 wall (portal from sector 0 to 1)
			leveldata.Wall{
				V1:          7,
				V2:          4,
				MSurf:       -1,
				LSurf:       2,
				USurf:       0,
				FrontSector: 1,
				BackSector:  0,
			},
			// 5-6 wall (sector 1)
			leveldata.Wall{
				V1:          5,
				V2:          6,
				MSurf:       2,
				LSurf:       -1,
				USurf:       -1,
				FrontSector: 1,
				BackSector:  -1,
			},
			// 6-7 wall (sector 1)
			leveldata.Wall{
				V1:          6,
				V2:          7,
				MSurf:       2,
				LSurf:       -1,
				USurf:       -1,
				FrontSector: 1,
				BackSector:  -1,
			},
			// 7-8 wall (sector 0)
			leveldata.Wall{
				V1:          7,
				V2:          8,
				MSurf:       0,
				LSurf:       -1,
				USurf:       -1,
				FrontSector: 0,
				BackSector:  -1,
			},
			// 8-9 wall (sector 0)
			leveldata.Wall{
				V1:          8,
				V2:          9,
				MSurf:       0,
				LSurf:       -1,
				USurf:       -1,
				FrontSector: 0,
				BackSector:  -1,
			},
			// 9-0 wall (sector 0)
			leveldata.Wall{
				V1:          9,
				V2:          0,
				MSurf:       0,
				LSurf:       -1,
				USurf:       -1,
				FrontSector: 0,
				BackSector:  -1,
			},

			// 10-11 (sector 2 nested in sector 0)
			leveldata.Wall{
				V1:          10,
				V2:          11,
				MSurf:       -1,
				LSurf:       4,
				USurf:       2,
				FrontSector: 2,
				BackSector:  0,
			},
			// 11-12 (sector 2 nested in sector 0)
			leveldata.Wall{
				V1:          11,
				V2:          12,
				MSurf:       -1,
				LSurf:       4,
				USurf:       2,
				FrontSector: 2,
				BackSector:  0,
			},
			// 12-13 (sector 2 nested in sector 0)
			leveldata.Wall{
				V1:          12,
				V2:          13,
				MSurf:       -1,
				LSurf:       4,
				USurf:       2,
				FrontSector: 2,
				BackSector:  0,
			},
			// 13-14 (sector 2 nested in sector 0)
			leveldata.Wall{
				V1:          13,
				V2:          14,
				MSurf:       -1,
				LSurf:       4,
				USurf:       2,
				FrontSector: 2,
				BackSector:  0,
			},
			// 14-10 (sector 2 nested in sector 0)
			leveldata.Wall{
				V1:          14,
				V2:          10,
				MSurf:       -1,
				LSurf:       4,
				USurf:       2,
				FrontSector: 2,
				BackSector:  0,
			},

			// TODO: add column in sector 1
		},

		Sectors: []leveldata.Sector{
			// sector 0
			leveldata.Sector{
				Sub:           []leveldata.SectorIndex{2},
				FloorHeight:   0,
				CeilingHeight: 6,
				FloorSurf:     1,
				CeilingSurf:   5,
			},
			// sector 1
			leveldata.Sector{
				FloorHeight:   -.5,
				CeilingHeight: 3.4,
				FloorSurf:     3,
				CeilingSurf:   5,
			},
			// sector 2
			leveldata.Sector{
				FloorHeight:   -1,
				CeilingHeight: 3.5,
				FloorSurf:     3,
				CeilingSurf:   2,
			},
		},

		PointLights: []lights.PointLight{

			// ROOM LIGHT

			// right bottom
			lights.PointLight{
				Position: mgl.Vec4{15, 5, 9, 10},
				Color:    mgl.Vec4{0.502, 0.537, 0.871, 15},
			},
			// right top
			lights.PointLight{
				Position: mgl.Vec4{15, 5, 24, 10},
				Color:    mgl.Vec4{0.502, 0.537, 0.871, 15},
			},
			// left top
			lights.PointLight{
				Position: mgl.Vec4{9, 5, 24, 10},
				Color:    mgl.Vec4{0.502, 0.537, 0.871, 15},
			},
			// left bottom
			lights.PointLight{
				Position: mgl.Vec4{9, 5, 9, 10},
				Color:    mgl.Vec4{0.502, 0.537, 0.871, 15},
			},

			// TUNNEL light
			lights.PointLight{
				Position: mgl.Vec4{12, 2.5, 45, 7},
				Color:    mgl.Vec4{0.961, 0.953, 0.522, 40},
			},
		},

		Entites: []leveldata.EntityDef{
			leveldata.EntityDef{
				Name: "player_start",
				Type: leveldata.PlayerStart,
				Pos:  mgl.Vec3{12, 0, 9},
			},
		},
	}
)
