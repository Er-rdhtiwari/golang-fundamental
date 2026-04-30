package main

import (
	"fmt"
)

type Queue struct {
	items []string
}

func (q *Queue) Enqueue(item string) {
	q.items = append(q.items, item)
}

func (q *Queue) Dequeue() (string, error) {
	if q.IsEmpty() {
		return "", fmt.Errorf("queue is empty")
	}
	firstItem := q.items[0]
	q.items = q.items[1:]
	return firstItem, nil

}

func (q *Queue) IsEmpty() bool {
	return len(q.items) == 0
}

func main() {
	queue := Queue{}
	queue.Enqueue("pr-event")
	queue.Enqueue("cd-event")
	queue.Enqueue("job-event")

	for !queue.IsEmpty() {
		event, err := queue.Dequeue()
		if err != nil {
			fmt.Println("error: ", err)
			return
		}
		fmt.Println("Processing:", event)
	}
}
