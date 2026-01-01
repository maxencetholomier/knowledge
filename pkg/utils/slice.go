package utils

func ItemInSlice(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}

	return false
}

func ANotInB(A, B []string) ([]string, error) {
	ids := []string{}

	for _, valueA := range A {
		if !ItemInSlice(B, valueA) {
			ids = append(ids, valueA)
		}
	}
	return ids, nil
}
