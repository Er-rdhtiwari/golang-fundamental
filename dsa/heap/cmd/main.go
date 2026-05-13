package main

import (
	"container/heap"
	"fmt"
)

type Notification struct{
	Message string
	Priority int
}

type NotificationQueue []Notification

func (q NotificationQueue) Len() int {
	return len(q)
}

func (q NotificationQueue) Less(i, j int) bool {
	return q[i].Priority < q[j].Priority

}

func (q NotificationQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q *NotificationQueue) Push(x any){
	item := x.(Notification)
	*q = append(*q, item)
}

func (q *NotificationQueue) Pop() any {
	old := *q
	n := len(old)
	item := old[n-1]
	*q = old[0:n-1]
	return item
}

func main(){

	queue := &NotificationQueue{}
	heap.Init(queue)

	heap.Push(queue, Notification{
		Message: "Go test Failed",
		Priority: 3,
	})
	heap.Push(queue, Notification{
		Message:  "Slack secret missing",
		Priority: 1,
	})

	heap.Push(queue, Notification{
		Message:  "Image pull failed",
		Priority: 2,
	})

	for queue.Len()>0{
		item := heap.Pop(queue).(Notification)
		fmt.Println(item.Message)
	}

}