package monitor

import (
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"

	mathutils "github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/math"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/rhi"
)

const (
	smoothedStep = float32(0.1)
	maxDeltaTime = 1.0 / 60.0
)

type Monitor struct {
	// timers for heavy operations
	timers     map[string]time.Time
	timersLock sync.RWMutex

	// last frame data
	lastFPS       atomic.Value // float64
	lastFrameTime atomic.Value // float64 (seconds)
	lastDrawCalls atomic.Uint64
	lastVertices  atomic.Uint64
	lastTriangles atomic.Uint64

	// delta time
	lastTime          time.Time
	deltaTime         float32
	smoothedDeltaTime float32

	// capacitors to compute within second
	frameCount   atomic.Uint64
	lastTickTime time.Time
}

func NewMonitor() *Monitor {
	m := &Monitor{
		timers:       make(map[string]time.Time),
		lastTickTime: time.Now(),
	}
	// init default values (to avoid panic on first read)
	m.lastFPS.Store(float64(0))
	m.lastFrameTime.Store(float64(0))
	return m
}

func (m *Monitor) StartTimer(name string) {
	m.timersLock.Lock()
	defer m.timersLock.Unlock()
	m.timers[name] = time.Now()
}

func (m *Monitor) EndTimer(name string) int64 {
	m.timersLock.RLock()
	t1, exist := m.timers[name]
	m.timersLock.RUnlock()

	if exist {
		return time.Since(t1).Milliseconds()
	}
	log.Printf("monitor - undefined timer - %s\n", name)
	return 0
}

func (m *Monitor) NewFrame() {
	now := time.Now()
	lastDeltaTime := m.deltaTime
	m.deltaTime = float32(time.Since(m.lastTime).Seconds())
	m.smoothedDeltaTime = float32(
		math.Min(
			float64(mathutils.Lerp(lastDeltaTime, m.deltaTime, smoothedStep)),
			maxDeltaTime,
		),
	)
	m.lastTime = now
}

// Update - calls one time in the end of the cycle
func (m *Monitor) Update() {
	m.frameCount.Add(1)

	// get stats from rhi FrameStats
	// IMPORTANT rhi stats must available on this stage
	m.lastDrawCalls.Store(rhi.FrameStats.DrawCalls)
	m.lastVertices.Store(rhi.FrameStats.Vertices)
	m.lastTriangles.Store(rhi.FrameStats.Triangles)

	// compute fps, frame time and other
	now := time.Now()
	elapsed := time.Since(m.lastTickTime)

	if elapsed >= time.Second {
		frames := m.frameCount.Load()
		seconds := elapsed.Seconds()

		fps := float64(frames) / seconds
		avgFrameTime := (seconds / float64(frames))

		m.lastFPS.Store(fps)
		m.lastFrameTime.Store(avgFrameTime)

		m.frameCount.Store(0)
		m.lastTickTime = now
	}
}

// Getters
func (m *Monitor) GetFPS() float64               { return m.lastFPS.Load().(float64) }
func (m *Monitor) GetFrameTime() float64         { return m.lastFrameTime.Load().(float64) }
func (m *Monitor) GetDrawCalls() uint64          { return m.lastDrawCalls.Load() }
func (m *Monitor) GetVertices() uint64           { return m.lastVertices.Load() }
func (m *Monitor) GetTriangles() uint64          { return m.lastTriangles.Load() }
func (m *Monitor) GetDeltaTime() float32         { return m.deltaTime }
func (m *Monitor) GetSmoothedDeltaTime() float32 { return m.smoothedDeltaTime }
