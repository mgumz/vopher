package main

func first_not_empty(parts ...string) (result string) {
	for i := range parts {
		if len(parts[i]) > 0 {
			return parts[i]
		}
	}
	return
}
