package main

// Preorder  -> N->L->R
// Inorder   -> L->N->R
// Postorder -> L->R->N

import "fmt"

type TreeNode struct {
	val   int
	Left  *TreeNode
	Right *TreeNode
}

func PreorderTraversal(root *TreeNode) []int {
	result := []int{}

	var dfs func(node *TreeNode)

	dfs = func(node *TreeNode) {
		if node == nil {
			return
		}
		result = append(result, node.val)
		dfs(node.Left)
		dfs(node.Right)
	}
	dfs(root)
	return result
}

func InorderTraversal(root *TreeNode) []int {
	result := []int{}

	var dfs func(node *TreeNode)

	dfs = func(node *TreeNode) {
		if node == nil {
			return
		}

		dfs(node.Left)
		result = append(result, node.val)
		dfs(node.Right)
	}
	dfs(root)
	return result
}

func PostorderTraversal(root *TreeNode) []int {
	result := []int{}

	var dfs func(node *TreeNode)

	dfs = func(node *TreeNode) {
		if node == nil {
			return
		}

		dfs(node.Left)
		dfs(node.Right)
		result = append(result, node.val)
	}
	dfs(root)
	return result
}

func main() {
	root := &TreeNode{val: 1}
	root.Left = &TreeNode{val: 2}
	root.Right = &TreeNode{val: 3}
	root.Left.Left = &TreeNode{val: 4}
	root.Right.Right = &TreeNode{val: 5}

	output := PreorderTraversal(root)
	fmt.Println(output)

	Inoutput := InorderTraversal(root)
	fmt.Println(Inoutput)

	Postoutput := PostorderTraversal(root)
	fmt.Println(Postoutput)

}
