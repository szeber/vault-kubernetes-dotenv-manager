package helper

func StringInSlice(haystack []string, needle string) bool {
	for _, x := range haystack {
		if x == needle {
			return true
		}
	}

	return false
}
