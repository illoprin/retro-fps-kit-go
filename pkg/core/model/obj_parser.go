package model

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	mgl "github.com/go-gl/mathgl/mgl32"
)

// OBJParser - парсер OBJ файлов
type OBJParser struct {
	positions []mgl.Vec3
	texCoords []mgl.Vec2
	normals   []mgl.Vec3
	vertices  []ModelVertex
	indices   []uint32
	vertexMap map[string]uint32
}

func NewOBJParser() *OBJParser {
	o := &OBJParser{}
	o.cleanUp()
	return o
}

func (p *OBJParser) cleanUp() {
	p.positions = make([]mgl.Vec3, 0)
	p.texCoords = make([]mgl.Vec2, 0)
	p.normals = make([]mgl.Vec3, 0)
	p.vertices = make([]ModelVertex, 0)
	p.indices = make([]uint32, 0)
	p.vertexMap = make(map[string]uint32)
}

// ParseFile парсит OBJ файл и возвращает модель
func (p *OBJParser) ParseFile(filename string) (*Model, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Добавляем начальные пустые элементы для индексации с 1
	p.positions = append(p.positions, mgl.Vec3{0, 0, 0})
	p.texCoords = append(p.texCoords, mgl.Vec2{0, 0})
	p.normals = append(p.normals, mgl.Vec3{0, 0, 0})

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "v":
			// Вершина
			if len(parts) < 4 {
				return nil, fmt.Errorf("line %d: invalid vertex format", lineNum)
			}
			x, _ := strconv.ParseFloat(parts[1], 32)
			y, _ := strconv.ParseFloat(parts[2], 32)
			z, _ := strconv.ParseFloat(parts[3], 32)
			p.positions = append(p.positions, mgl.Vec3{float32(x), float32(y), float32(z)})

		case "vt":
			// Текстурная координата
			if len(parts) < 3 {
				return nil, fmt.Errorf("line %d: invalid texture coordinate format", lineNum)
			}
			u, _ := strconv.ParseFloat(parts[1], 32)
			v, _ := strconv.ParseFloat(parts[2], 32)
			p.texCoords = append(p.texCoords, mgl.Vec2{float32(u), float32(v)})

		case "vn":
			// Нормаль
			if len(parts) < 4 {
				return nil, fmt.Errorf("line %d: invalid normal format", lineNum)
			}
			nx, _ := strconv.ParseFloat(parts[1], 32)
			ny, _ := strconv.ParseFloat(parts[2], 32)
			nz, _ := strconv.ParseFloat(parts[3], 32)
			// Нормализуем нормаль
			normal := mgl.Vec3{float32(nx), float32(ny), float32(nz)}.Normalize()
			p.normals = append(p.normals, normal)

		case "f":
			// Грань (треугольник или многоугольник)
			if len(parts) < 4 {
				return nil, fmt.Errorf("line %d: invalid face format (need at least 3 vertices)", lineNum)
			}

			// Обрабатываем грань как треугольник (разбиваем многоугольники на треугольники)
			indices := make([]uint32, 0)
			for i := 1; i < len(parts); i++ {
				idx, err := p.parseVertexIndex(parts[i])
				if err != nil {
					return nil, fmt.Errorf("line %d: %v", lineNum, err)
				}
				indices = append(indices, idx)
			}

			// Разбиваем на треугольники (триангуляция многоугольника)
			for i := 1; i < len(indices)-1; i++ {
				p.indices = append(p.indices, indices[0], indices[i], indices[i+1])
			}

		default:
			// Игнорируем другие типы данных (группы, материалы и т.д.)
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	m := &Model{
		Vertices: p.vertices,
		Indices:  p.indices,
	}
	p.cleanUp()

	return m, nil
}

// parseVertexIndex парсит индекс вершины в формате v/vt/vn или v//vn и т.д.
func (p *OBJParser) parseVertexIndex(token string) (uint32, error) {
	// Сохраняем оригинальный токен для маппинга
	// origToken := token

	parts := strings.Split(token, "/")

	var vIdx, vtIdx, vnIdx int

	// Парсим индекс позиции
	if len(parts) >= 1 && parts[0] != "" {
		idx, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid vertex index: %s", parts[0])
		}
		vIdx = idx
	} else {
		return 0, fmt.Errorf("missing vertex index")
	}

	// Парсим индекс текстуры
	if len(parts) >= 2 && parts[1] != "" {
		idx, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, fmt.Errorf("invalid texture index: %s", parts[1])
		}
		vtIdx = idx
	}

	// Парсим индекс нормали
	if len(parts) >= 3 && parts[2] != "" {
		idx, err := strconv.Atoi(parts[2])
		if err != nil {
			return 0, fmt.Errorf("invalid normal index: %s", parts[2])
		}
		vnIdx = idx
	}

	// Преобразуем в индексы массивов (OBJ использует 1-based индексацию)
	if vIdx < 0 {
		vIdx = len(p.positions) + vIdx
	}
	if vtIdx < 0 {
		vtIdx = len(p.texCoords) + vtIdx
	}
	if vnIdx < 0 {
		vnIdx = len(p.normals) + vnIdx
	}

	// Создаем ключ для поиска существующей вершины
	key := fmt.Sprintf("%d/%d/%d", vIdx, vtIdx, vnIdx)

	// Если вершина уже существует, возвращаем её индекс
	if idx, exists := p.vertexMap[key]; exists {
		return idx, nil
	}

	// Создаем новую вершину
	vertex := ModelVertex{}

	// Позиция
	if vIdx >= 1 && vIdx < len(p.positions) {
		vertex.X = p.positions[vIdx].X()
		vertex.Y = p.positions[vIdx].Y()
		vertex.Z = p.positions[vIdx].Z()
	} else {
		return 0, fmt.Errorf("vertex index out of range: %d", vIdx)
	}

	// Текстурные координаты
	if vtIdx >= 1 && vtIdx < len(p.texCoords) {
		vertex.U = p.texCoords[vtIdx].X()
		vertex.V = p.texCoords[vtIdx].Y()
	}

	// Нормаль
	if vnIdx >= 1 && vnIdx < len(p.normals) {
		vertex.Nx = p.normals[vnIdx].X()
		vertex.Ny = p.normals[vnIdx].Y()
		vertex.Nz = p.normals[vnIdx].Z()
	}

	// Добавляем вершину
	newIdx := uint32(len(p.vertices))
	p.vertices = append(p.vertices, vertex)
	p.vertexMap[key] = newIdx

	return newIdx, nil
}
