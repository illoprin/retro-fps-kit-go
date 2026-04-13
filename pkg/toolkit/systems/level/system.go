package levelsys

import "github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/assets/level"

type LevelSystem struct {
	builder *leveldata.LevelBuilder
}

func NewLevelSystem(b *leveldata.LevelBuilder) *LevelSystem {
	l := &LevelSystem{
		builder: b,
	}
	return l
}

func (l *LevelSystem) GetBuilder() *leveldata.LevelBuilder {
	return l.builder
}

func (l *LevelSystem) Update(dt float32) {

}
