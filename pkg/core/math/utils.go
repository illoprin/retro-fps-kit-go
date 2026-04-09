package math

func Lerp(a, b, f float32) float32 {
	return a + f*(b-a)
}

func BoolToInt32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
