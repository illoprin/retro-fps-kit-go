package monitor

import (
	"log"
	"sync/atomic"
	"time"
)

type Timer map[string]time.Time

type Monitor struct {
	timers         Timer
	lastTime       time.Time
	lastDrawCalls  atomic.Uint64
	lastVertices   atomic.Uint64
	lastTriangles  atomic.Uint64
	framesInSecond atomic.Uint64
	deltaTime      atomic.Value
	fps            atomic.Value
}

func NewMonitor(timers []string) (p *Monitor) {

	t := make(map[string]time.Time)
	for _, n := range timers {
		t[n] = time.Time{}
	}

	p = &Monitor{
		timers: t,
	}
	return nil
}

func (m *Monitor) StartTimer(name string) {
	if _, exist := m.timers[name]; exist {
		m.timers[name] = time.Now()
	}
	log.Printf("performance monitor - undefined timer - %s\n", name)
}

// EndTimer returns milliseconds since timer start
func (m *Monitor) EndTimer(name string) int64 {
	if _, exist := m.timers[name]; exist {
		t1 := m.timers[name]
		return time.Since(t1).Milliseconds()
	}
	log.Printf("performance monitor - undefined timer - %s\n", name)
	return 0
}

func (m *Monitor) GetTimerLastState(name string) time.Time {
	if state, exists := m.timers[name]; exists {
		return state
	}
	log.Printf("performance monitor - undefined timer - %s\n", name)
	return time.Time{}
}

// Update must call in the end of cycle before UpdateStates
func (m *Monitor) Update() {

	// FIX update draw calls
	now := time.Now()
	delta := time.Since(m.lastTime)
	if delta >= time.Second {
		frames := m.framesInSecond.Load()
		fps := float64(frames) / delta.Seconds()
		frameTime := (1 / float64(frames)) * 1000

		m.fps.Store(fps)
		m.deltaTime.Store(frameTime)

		m.framesInSecond.Store(0)
		m.framesInSecond.Store(0)
		m.lastTime = now
	}
	m.framesInSecond.Add(1)
}

func (m *Monitor) GetDeltaTime() float32 {
	return m.deltaTime.Load().(float32)
}

func (m *Monitor) GetFPS() float64 {
	return m.fps.Load().(float64)
}
