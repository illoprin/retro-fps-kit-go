package config

import (
	"github.com/goccy/go-yaml"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/logger"
)

type Manager struct {
	current     *Config
	initialRaw  []byte
	storagePath string
}

func NewManager(initial *Config, path string) *Manager {
	// bytes copy of initial state
	initialRaw, err := yaml.Marshal(initial)
	if err != nil {
		logger.Warnf("failed to encode initial config state")
	}

	return &Manager{
		current:     initial,
		initialRaw:  initialRaw,
		storagePath: path,
	}
}

func (m *Manager) Save() error {
	return SaveConfig(m.storagePath, m.current)
}

func (m *Manager) Reset() {
	// set initial state
	err := yaml.Unmarshal(m.initialRaw, m.current)
	if err != nil {
		logger.Warnf("failed to decode initial config state")
	}
}

func (m *Manager) SetConfig(c *Config) {
	m.current = c
}

func (m *Manager) Config() *Config {
	return m.current
}
