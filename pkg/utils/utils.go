package utils

func StrInSlice(strings []string, s string) bool {
	for _, x := range strings {
		if x == s {
			return true
		}
	}
	return false
}
