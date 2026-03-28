package global

var (
	DrawCalls          uint32 = 0
	DrawVertices       uint32 = 0
	LastDrawCalls      uint32 = 0
	LastVertices       uint32 = 0
	LastImguiVertices  int32  = 0
	SizeOfFloat        int32  = 4
	LastImguiDrawCalls int32  = 0
)

const (
	DefaultResolutionRatio float32 = 0.5
)

func BoolToInt32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
