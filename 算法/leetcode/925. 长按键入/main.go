package main

func isLongPressedName(name string, typed string) bool {
	if len(typed) < len(name) {
		return false
	}
	i, j  := 0, 0
	for i < len(name) {
		flag := name[i]
		count := 0
		for i < len(name) {
			if name[i] != flag {
				break
			} else {
				count++
			}
			i++
		}

		for j < len(typed) {
			if typed[j] != flag {
				break
			} else {
				count--
			}
			j++
		}
		if count > 0 {
			return false
		}
	}
	if j != len(typed) {
		return false
	}
	return true
}