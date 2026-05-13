package main

import (
	"fmt"
	"sync"
)

type NotificationJob struct{
	ID int
	UserID string
	Message string
}


func ProcessNotification(workerID int, job NotificationJob){
	fmt.Printf(
		"Worker %d semding notifcation %d to user %s: %s\n",
		workerID,
		job.ID,
		job.UserID,
		job.Message,
	)
}

func Worker(id int, jobs<-chan NotificationJob, wg *sync.WaitGroup){
	defer wg.Done()

	for job := range jobs{
		ProcessNotification(id, job)
	}
}

func main(){
	jobs := make(chan NotificationJob)

	var wg sync.WaitGroup
	workerCount :=3

	for i:=1;i<=workerCount; i++{
		wg.Add(1)
		go Worker(i, jobs, &wg)
	}

	notificationJobs := []NotificationJob{
        {ID: 1, UserID: "U001", Message: "Welcome to the workspace"},
        {ID: 2, UserID: "U002", Message: "Your report is ready"},
        {ID: 3, UserID: "U003", Message: "You have a new mention"},
        {ID: 4, UserID: "U004", Message: "Daily summary is available"},
        {ID: 5, UserID: "U005", Message: "Password changed successfully"},
    }
	for _, job :=range notificationJobs{
		jobs<- job
	}

	close(jobs)

	wg.Wait()

	fmt.Println("All notifications processed")

}