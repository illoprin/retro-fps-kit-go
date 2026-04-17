package math

const (
	Epsilon = float32(1e-5)
)

func Lerp(a, b, f float32) float32 {
	return a + f*(b-a)
}

func BoolToInt32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func Max(a int, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func Min(a int, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
