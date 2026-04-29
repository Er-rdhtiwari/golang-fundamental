package main

import "fmt"

func InsertPosition(numbers []int, targrt int) int{
	left :=0
	right := len(numbers)-1
	for left <= right{
		mid := left + (right-left)/2

		if numbers[mid] == targrt{
			return mid
		}
		if numbers[mid] < targrt{
			left = mid+1
		} else {
			right = mid-1
		}
	}
	return left
}

func main(){
	nums := []int{2, 4, 6, 8, 10, 12, 14}
	target := 11

	index := InsertPosition(nums, target)

	fmt.Println("Index:", index)
}