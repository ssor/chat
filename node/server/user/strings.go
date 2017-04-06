package user

func removeStringFromSlice(slice []string, index int) []string {
	if index >= len(slice) {
		return slice[:]
	}
	return append(slice[:index], slice[index+1:]...)
}
