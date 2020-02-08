package GCache

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func isPowerOfTwo(number int) bool {
	return (number & (number - 1)) == 0
}
