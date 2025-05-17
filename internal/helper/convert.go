package helper

func ConvertInt32SliceToInt(src []int32) []int {
	dst := make([]int, len(src))
	for i, v := range src {
		dst[i] = int(v)
	}
	return dst
}