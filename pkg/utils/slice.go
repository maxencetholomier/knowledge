package utils

func ItemInSlice(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}

	return false
}

func ANotInB(a, b []string) []string {
	var ids []string
	for _, value := range a {
		if !ItemInSlice(b, value) {
			ids = append(ids, value)
		}
	}
	return ids
}
