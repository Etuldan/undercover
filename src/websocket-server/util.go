package main

func checkInList(list []int, element int) bool {
	for _, value := range list {
		if value == element {
			return true
		}
	}
	return false
}
