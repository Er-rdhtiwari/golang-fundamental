package main

import "fmt"

func BubbleSort(numbers []int) {
	n := len(numbers)

	for i := 0; i < n; i++ {
		for j := 0; j < n-i-1; j++ {
			if numbers[j] > numbers[j+1] {
				numbers[j], numbers[j+1] = numbers[j+1], numbers[j]
			}
		}
	}
}

func main() {
	duration := []int{45, 0, 23, 65, 99}
	BubbleSort(duration)
	fmt.Println(duration)
}
