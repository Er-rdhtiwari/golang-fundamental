package main

import (
	"fmt"
	"strings"
)

func countVowels(s string) int {
	s = strings.ToLower(s)
	count := 0
	for _, ch := range s {
		if ch == 'a' || ch == 'e' || ch == 'i' || ch == 'o' || ch == 'u' {
			count++
		}
	}
	return count

}

func main() {
	fmt.Println(countVowels("Golang"))

}
