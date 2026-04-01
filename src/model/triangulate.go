package model

import (
	"math"
)

// vertex2D represents a 2D vertex with additional 3D info
type vertex2D struct {
	X, Y     float64
	Original ModelVertex
	Index    int
}

// Triangulate polygon using ear clipping algorithm
// Input: slice of vertices (can be convex or concave CCW )
// Output: vertices (with computed normals and texcoords) and indices for triangles
func TriangulatePolygon(vertices []ModelVertex) ([]ModelVertex, []uint32) {
	if len(vertices) < 3 {
		return nil, nil
	}

	// Create working copy with 2D coordinates (ignoring Z for triangulation)
	workingVerts := make([]vertex2D, len(vertices))
	for i, v := range vertices {
		workingVerts[i] = vertex2D{
			X:        float64(v.X),
			Y:        float64(v.Y),
			Original: v,
			Index:    i,
		}
	}

	// Ensure vertices are in counter-clockwise order
	if !isCounterClockwise(workingVerts) {
		reverseVertices(workingVerts)
	}

	// Ear clipping algorithm
	triangles := make([]uint32, 0)
	vertexList := make([]int, len(workingVerts))
	for i := range vertexList {
		vertexList[i] = i
	}

	for len(vertexList) > 3 {
		earFound := false

		for i := 0; i < len(vertexList); i++ {
			prev := vertexList[(i-1+len(vertexList))%len(vertexList)]
			curr := vertexList[i]
			next := vertexList[(i+1)%len(vertexList)]

			if isEar(workingVerts, prev, curr, next, vertexList) {
				// Add triangle
				triangles = append(triangles, uint32(prev), uint32(curr), uint32(next))

				// Remove current vertex
				vertexList = append(vertexList[:i], vertexList[i+1:]...)
				earFound = true
				break
			}
		}

		if !earFound {
			// Fallback: simple fan triangulation for remaining polygon
			for i := 1; i < len(vertexList)-1; i++ {
				triangles = append(triangles, uint32(vertexList[0]), uint32(vertexList[i]), uint32(vertexList[i+1]))
			}
			break
		}
	}

	// Add the last triangle
	if len(vertexList) == 3 {
		triangles = append(triangles, uint32(vertexList[0]), uint32(vertexList[1]), uint32(vertexList[2]))
	}

	// Generate texture coordinates and normals for output vertices
	result := make([]ModelVertex, len(vertices))
	copy(result, vertices)

	// Calculate bounding box for texture coordinates
	minX, minY := float64(vertices[0].X), float64(vertices[0].Y)
	maxX, maxY := minX, minY

	for _, v := range vertices {
		if float64(v.X) < minX {
			minX = float64(v.X)
		}
		if float64(v.Y) < minY {
			minY = float64(v.Y)
		}
		if float64(v.X) > maxX {
			maxX = float64(v.X)
		}
		if float64(v.Y) > maxY {
			maxY = float64(v.Y)
		}
	}

	// Generate UV coordinates (simple planar mapping)
	for i := range result {
		// UV based on bounding box
		if maxX > minX && maxY > minY {
			result[i].U = float32((float64(result[i].X) - minX) / (maxX - minX))
			result[i].V = float32((float64(result[i].Y) - minY) / (maxY - minY))
		}

		// Calculate normal (assuming flat polygon in XY plane)
		// Compute normal from first triangle
		if len(triangles) >= 3 {
			v0 := vertices[triangles[0]]
			v1 := vertices[triangles[1]]
			v2 := vertices[triangles[2]]

			edge1 := ModelVertex{X: v1.X - v0.X, Y: v1.Y - v0.Y, Z: v1.Z - v0.Z}
			edge2 := ModelVertex{X: v2.X - v0.X, Y: v2.Y - v0.Y, Z: v2.Z - v0.Z}

			normal := ModelVertex{
				Nx: edge1.Y*edge2.Z - edge1.Z*edge2.Y,
				Ny: edge1.Z*edge2.X - edge1.X*edge2.Z,
				Nz: edge1.X*edge2.Y - edge1.Y*edge2.X,
			}

			// Normalize
			length := float32(math.Sqrt(float64(normal.Nx*normal.Nx + normal.Ny*normal.Ny + normal.Nz*normal.Nz)))
			if length > 0 {
				normal.Nx /= length
				normal.Ny /= length
				normal.Nz /= length
			}

			// Apply same normal to all vertices (flat shading)
			for j := range result {
				result[j].Nx = normal.Nx
				result[j].Ny = normal.Ny
				result[j].Nz = normal.Nz
			}
		}
	}

	return result, triangles
}

// Check if polygon vertices are in counter-clockwise order
func isCounterClockwise(vertices []vertex2D) bool {
	sum := 0.0
	for i := 0; i < len(vertices); i++ {
		j := (i + 1) % len(vertices)
		sum += (vertices[j].X - vertices[i].X) * (vertices[j].Y + vertices[i].Y)
	}
	return sum < 0
}

// Reverse vertex order
func reverseVertices(vertices []vertex2D) {
	for i, j := 0, len(vertices)-1; i < j; i, j = i+1, j-1 {
		vertices[i], vertices[j] = vertices[j], vertices[i]
	}
}

// Check if vertex 'curr' is an ear
func isEar(vertices []vertex2D, prev, curr, next int, vertexList []int) bool {
	p := vertices[prev]
	c := vertices[curr]
	n := vertices[next]

	// Check if angle is convex (not reflex)
	if !isConvex(p, c, n) {
		return false
	}

	// Check if any other vertex lies inside triangle
	for _, vIdx := range vertexList {
		if vIdx == prev || vIdx == curr || vIdx == next {
			continue
		}

		if isPointInTriangle(vertices[vIdx], p, c, n) {
			return false
		}
	}

	return true
}

// Check if angle at vertex b is convex (less than 180 degrees)
func isConvex(a, b, c vertex2D) bool {
	// Cross product for 2D (z-component)
	cross := (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
	return cross > 0
}

// Check if point p is inside triangle (a,b,c)
func isPointInTriangle(p, a, b, c vertex2D) bool {
	// Barycentric coordinate method
	denom := (b.Y-c.Y)*(a.X-c.X) + (c.X-b.X)*(a.Y-c.Y)
	if denom == 0 {
		return false
	}

	u := ((b.Y-c.Y)*(p.X-c.X) + (c.X-b.X)*(p.Y-c.Y)) / denom
	v := ((c.Y-a.Y)*(p.X-c.X) + (a.X-c.X)*(p.Y-c.Y)) / denom
	w := 1 - u - v

	return u >= 0 && u <= 1 && v >= 0 && v <= 1 && w >= 0 && w <= 1
}

// Helper function to calculate polygon area
func polygonArea(vertices []ModelVertex) float64 {
	area := 0.0
	n := len(vertices)
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		area += float64(vertices[i].X) * float64(vertices[j].Y)
		area -= float64(vertices[j].X) * float64(vertices[i].Y)
	}
	return math.Abs(area) / 2.0
}
