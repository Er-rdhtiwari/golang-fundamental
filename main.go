package main

import (
	"flag"
	"fmt"
)

type Event struct {
	User      string
	EventType string
	Repo      string
}

func main() {
	user := flag.String("user", "guest", "user name")
	event := flag.String("event", "login", "event type")
	repo := flag.String("repo", "default", "repository name")

	flag.Parse()
	if *user == "" {
		fmt.Println("User name is required")
		return
	}

	fmt.Println("Hello from Go CLI", *user, *event, *repo)
	inputEvent := Event{
		User:      *user,
		EventType: *event,
		Repo:      *repo,
	}
	fmt.Printf("New Event: %s ", inputEvent)
	fmt.Printf("\n")
	fmt.Printf("%v", inputEvent)
	fmt.Printf("\n")
	fmt.Printf("%+v", inputEvent)
	fmt.Printf("\n")
}
