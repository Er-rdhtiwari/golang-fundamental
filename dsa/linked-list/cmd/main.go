package main

import "fmt"

type Node struct {
	Value int
	Next  *Node
}

func CountNodes(head *Node) int {
	count := 0
	current := head

	for current != nil {
		count++
		current = current.Next
	}
	return count
}

func main() {
	head := &Node{Value: 10}
	second := &Node{Value: 20}
	third := &Node{Value: 30}

	head.Next = second
	second.Next = third
	current := head
	result := CountNodes(head)

	fmt.Println("Total nodes:", result)

	for current != nil {
		fmt.Println(current.Value)
		current = current.Next
	}
}
