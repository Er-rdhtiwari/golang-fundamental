package main

import (
	"fmt"
	"flag"
)

func main() {
	user := flag.String("user", "guest", "user name")
	event := flag.String("event", "login", "event type")
	repo := flag.String("repo", "default", "repository name")
	flag.Parse()
	fmt.Println("Hello from Go CLI", *user, *event, *repo)
}