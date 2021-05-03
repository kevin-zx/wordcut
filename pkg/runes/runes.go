package runes

func Equal(a, b []rune) bool {
	if len(a) == len(b) {
		for i := 0; i < len(a); i++ {
			if a[i] != b[i] {
				return false
			}
		}

		return true
	}

	return false
}

//
// if a < b return -1
// if a == b return 0
// if a > b return 1
func Compare(a, b []rune) int {
	l := len(a)
	if len(a) > len(b) {
		l = len(b)
	}
	for i := 0; i < l; i++ {
		if a[i] > b[i] {
			return 1
		}
		if a[i] < b[i] {
			return -1
		}
	}
	if len(a) == len(b) {
		return 0
	}
	if len(a) > len(b) {
		return 1
	}
	return -1
}
