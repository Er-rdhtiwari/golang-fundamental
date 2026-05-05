package main

import "fmt"

func bfs(graph map[string][]string, start string) []string{

	visited := make(map[string]bool)
	queue := []string{start}
	order := []string{}

	visited[start] = true

	for len(queue) > 0{
		current := queue[0]
		queue = queue[1:]
		order = append(order,current)

		for _, neighbor := range graph[current]{
			if !visited[neighbor]{
				visited[neighbor]= true
				queue = append(queue, neighbor)
			}
		}
	}
	return order
}

func main() {
	graph := map[string][]string{
		"A": []string{"B", "C"},
		"B": []string{"D"},
		"C": []string{"D"},
		"D": []string{"E"},
		"E": []string{},
	}

	result := bfs(graph, "A")
	fmt.Println(result)
}