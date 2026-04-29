package main

import "fmt"

func BinarySearch(numbers []int, targrt int) int{
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
	return -1
}

func main(){
	nums := []int{2, 4, 6, 8, 10, 12, 14}
	target := 10

	index := BinarySearch(nums, target)

	fmt.Println("Index:", index)
}