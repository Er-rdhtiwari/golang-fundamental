package main

import "fmt"

func CountCharater(item string) map[rune]int {
	frequency := make(map[rune]int)
	for _, ch := range item {
		frequency[ch]++
	}
	return frequency
}

func main() {
	frequency := CountCharater("banana")

	for ch, count := range frequency {
		fmt.Printf("%c: %d\n", ch, count)
	}
	fmt.Printf("Testing %+v\n", frequency)
}
