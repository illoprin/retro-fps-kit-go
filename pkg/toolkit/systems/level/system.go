package levelsys

import "github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/assets/level"

type LevelSystem struct {
	builder *levelasset.LevelBuilder
}

func NewLevelSystem(b *levelasset.LevelBuilder) *LevelSystem {
	l := &LevelSystem{
		builder: b,
	}
	return l
}

func (l *LevelSystem) GetBuilder() *levelasset.LevelBuilder {
	return l.builder
}

func (l *LevelSystem) Update(dt float32) {

}
