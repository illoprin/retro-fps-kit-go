package levelsystem

import "github.com/illoprin/retro-fps-toolkit-go/pkg/toolkit/assets/level"

type LevelSystem struct {
	builder *level.LevelBuilder
}

func (l *LevelSystem) GetBuilder() *level.LevelBuilder {
	return l.builder
}
