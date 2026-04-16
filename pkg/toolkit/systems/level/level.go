package levelsys

import leveldata "github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/assets/level"

type Level struct {
	builder *leveldata.LevelBuilder
}

func NewLevelSystem(b *leveldata.LevelBuilder) *Level {
	l := &Level{
		builder: b,
	}
	return l
}

func (l *Level) GetBuilder() *leveldata.LevelBuilder {
	return l.builder
}

func (l *Level) Update(dt float32) {

}
