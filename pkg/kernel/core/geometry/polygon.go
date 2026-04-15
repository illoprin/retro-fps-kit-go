package geometry

import (
	"fmt"
	"math"
	"sort"

	mgl "github.com/go-gl/mathgl/mgl32"
	lmath "github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/math"
)

type Segment struct {
	P1 mgl.Vec2
	P2 mgl.Vec2
}

// почти равенство (из-за float)
func almostEqual(a, b mgl.Vec2) bool {
	return math.Abs(float64(a.X()-b.X())) < float64(lmath.Epsilon) &&
		math.Abs(float64(a.Y()-b.Y())) < float64(lmath.Epsilon)
}

// delte
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

// центр масс
func centroid(points []mgl.Vec2) mgl.Vec2 {
	var c mgl.Vec2
	for _, p := range points {
		c = c.Add(p)
	}
	return c.Mul(1.0 / float32(len(points)))
}

// угол точки относительно центра
func angle(c, p mgl.Vec2) float32 {
	return float32(math.Atan2(
		float64(p.Y()-c.Y()),
		float64(p.X()-c.X()),
	))
}

// основная функция
func BuildPolygon(segments []Segment) []mgl.Vec2 {
	points := uniquePoints(segments)

	c := centroid(points)

	// сортировка по углу (CCW)
	sort.Slice(points, func(i, j int) bool {
		return angle(c, points[i]) < angle(c, points[j])
	})

	return points
}

func main() {
	segments := []Segment{
		{mgl.Vec2{0, 0}, mgl.Vec2{1, 0}},
		{mgl.Vec2{1, 0}, mgl.Vec2{1, 1}},
		{mgl.Vec2{1, 1}, mgl.Vec2{0, 1}},
		{mgl.Vec2{0, 1}, mgl.Vec2{0, 0}},
	}

	polygon := BuildPolygon(segments)

	for i, p := range polygon {
		fmt.Printf("%d: %v\n", i, p)
	}
}
