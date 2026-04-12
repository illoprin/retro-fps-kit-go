package resource

import (
	"sync"

	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/rhi"
)

type ResourceScope struct {
	resources []rhi.Resource
	mu        sync.Mutex
}

// Track - adds resource to delete list
func (s *ResourceScope) Track(r rhi.Resource) rhi.Resource {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources = append(s.resources, r)
	return r
}

// Purge - removes accumulated resources
func (s *ResourceScope) Purge() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := len(s.resources) - 1; i >= 0; i-- {
		s.resources[i].Delete()
	}
	s.resources = nil
}
