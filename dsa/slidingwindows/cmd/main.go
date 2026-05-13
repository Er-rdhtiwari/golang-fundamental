package main

import "fmt"

func maxSumSubarray(nums []int, k int) int{
	if len(nums) <k || k<=0{
		return 0
	}

	windowSum :=0

	for i:=0; i<k; i++{
		windowSum += nums[i]
	}

	maxSum := windowSum

	for right :=k; right<len(nums); right++{
		left := right-k
		windowSum = windowSum-nums[left]+ nums[right]

		if windowSum > maxSum{
			maxSum= windowSum
		}
	}
	return maxSum
}

func main() {
    nums := []int{2, 1, 5, 1, 3, 2}
    k := 3

    result := maxSumSubarray(nums, k)

    fmt.Println(result)
}