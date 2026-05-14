package main

import "fmt"


type NumArray struct{
	prefix []int
}

func cunstructor(nums []int) NumArray{
	prefix := make([]int,len(nums)+1)

	for i :=0; i<len(nums); i++{
		prefix[i+1]= prefix[i]+ nums[i]

	}
	return NumArray{
		prefix: prefix,
	}
}

func (n NumArray) SumRange(left int, right int) int{
	return n.prefix[right+1]- n.prefix[left]
}
	
func main(){
	nums:= []int{2,4,1,3}

	numArray :=cunstructor(nums)

	fmt.Println(numArray.SumRange(1,3))
	fmt.Println(numArray.SumRange(0,2))
	fmt.Println(numArray.SumRange(2,2))

}