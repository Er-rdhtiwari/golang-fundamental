package main

import (
	"fmt"

)

func main(){
	message :=make(chan string)

	go func(){
		message <- "New Slack Message Received"
	}()
	msg := <-message
	fmt.Println(msg)
}