package controls

var UITextSizes = []int{16, 24, 32}

func ValidUITextSize(size int) bool {
	return size == 16 || size == 24 || size == 32
}
