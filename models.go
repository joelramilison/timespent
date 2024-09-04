package main



// returns the activity name in a form that can be used in HTML as a selector value
func reduceActivity(name string) string {
	
	result := []rune{}
	for _, char := range name {
		if char != []rune(" ")[0] {
			result = append(result, char)
		}
	}
	return string(result)
}