package main

import "fmt"

func SumN(n int) int {
	if n == 1{
		return 1
	}
	if n <= 0{
		return 0
	}
	return n + SumN(n-1)
}

func main(){
	result := SumN(5)
	fmt.Println("Sum: ", result)
}