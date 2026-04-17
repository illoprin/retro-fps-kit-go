package geometry

import (
	"math"
	"sort"

	mgl "github.com/go-gl/mathgl/mgl32"
	lmath "github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/math"
)

// Segment represents 2D line
type Segment struct {
	P1 mgl.Vec2
	P2 mgl.Vec2
}

func almostEqual(a, b mgl.Vec2) bool {
	return math.Abs(float64(a.X()-b.X())) < float64(lmath.Epsilon) &&
		math.Abs(float64(a.Y()-b.Y())) < float64(lmath.Epsilon)
}

// delete
func uniquePoints(segments []Segment) []mgl.Vec2 {
	var points []mgl.Vec2

	for _, s := range segments {
		found := false
		for _, p := range points {
			if almostEqual(p, s.P1) {
				found = true
				break
			}
		}
		if !found {
			points = append(points, s.P1)
		}
	}

	return points
}

// compute centroid
func centroid(points []mgl.Vec2) mgl.Vec2 {
	var c mgl.Vec2
	for _, p := range points {
		c = c.Add(p)
	}
	return c.Mul(1.0 / float32(len(points)))
}

// polygon dot angle relative to center of mass
func angle(c, p mgl.Vec2) float32 {
	return float32(math.Atan2(
		float64(p.Y()-c.Y()),
		float64(p.X()-c.X()),
	))
}

// BuildPolygon - returns CCW ordered points of polygon
// segments - segments that make up the polygon in random order
func BuildPolygon(segments []Segment) []mgl.Vec2 {
	points := uniquePoints(segments)

	c := centroid(points)

	// sort by angle
	sort.Slice(points, func(i, j int) bool {
		return angle(c, points[i]) < angle(c, points[j])
	})

	return points
}
