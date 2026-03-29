package math

func Lerp(a, b, f float32) float32 {
	return a + f*(b-a)
}
