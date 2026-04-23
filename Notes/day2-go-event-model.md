# Day 2 — Golang Structs, Methods, Packages, and Event Model

Today we move from “how input enters the CLI” to “how the program stores that input in a proper model.”

This is a very important day because real backend and DevOps tools do not pass random loose values everywhere. They usually convert input into a **typed object** first, validate it, and then pass that object through the system.

---

## 1. Day 2 learning goals

By the end of Day 2, you should understand:

* what a `struct` is in Go
* how methods work on structs
* what packages are
* exported vs unexported names
* how code is organized into folders in real projects
* what an event model is
* how raw CLI input becomes a typed `PipelineEvent`
* why validation should live in the model layer
* why passing one struct is better than passing many loose values
* basic string concepts for DSA
* one easy Go string problem

---

## 2. Quick revision of Day 1 in 5–8 points

Here is the Day 1 recap in simple language:

1. A Go program starts from `package main` and `func main()`.
2. `main.go` is usually the entry point of a CLI program.
3. CLI flags let the user pass input like `--event pr --status success`.
4. `go.mod` defines the module name and manages dependencies.
5. Imports are how one Go file uses another package.
6. A CLI tool reads input, processes it, and performs an action.
7. In a production CLI, the flow is usually: input -> parse -> validate -> build request -> execute.
8. Real tools should separate responsibilities instead of putting everything inside `main.go`.

---

## 3. Beginner-friendly explanation of structs using simple examples first

### What is a struct?

A `struct` is a way to group related data together.

Think like this:

If you want to represent one person, you do not want 5 separate variables floating around.

Bad style:

```go
name := "Radhe"
age := 28
city := "Chennai"
email := "radhe@example.com"
isActive := true
```

These values are related, but they are scattered.

Better style:

```go
type Person struct {
	Name     string
	Age      int
	City     string
	Email    string
	IsActive bool
}
```

Now one `Person` value can hold all related fields together.

Example:

```go
p := Person{
	Name:     "Radhe",
	Age:      28,
	City:     "Chennai",
	Email:    "radhe@example.com",
	IsActive: true,
}
```

### Real-life meaning

A struct is like a form.

For example:

* Student form
* Employee record
* Order details
* API request object
* Pipeline event

All of these are “one thing” with multiple related fields.

---

### Toy example 1: Book

```go
type Book struct {
	Title  string
	Author string
	Price  float64
}
```

Use it:

```go
book := Book{
	Title:  "Go Basics",
	Author: "John",
	Price:  499.0,
}
```

---

### Toy example 2: Car

```go
type Car struct {
	Brand string
	Model string
	Year  int
}
```

---

### Why structs matter in backend projects

In backend projects, we often need to represent:

* user requests
* database rows
* configuration
* events
* API responses
* pipeline metadata

In your project, the important struct is something like:

* `PipelineEvent`
* `NotificationRequest`

That struct becomes the clean typed form of raw input.

---

### Zero values in structs

This is very important in Go.

If you create a struct and do not assign values, Go gives default values automatically.

Example:

```go
type User struct {
	Name   string
	Age    int
	Active bool
}
```

```go
var u User
fmt.Println(u)
```

Output idea:

```go
{Name: Age:0 Active:false}
```

So default zero values are:

* `string` -> `""`
* `int` -> `0`
* `bool` -> `false`
* pointer -> `nil`
* slice -> `nil`
* map -> `nil`

### Why zero values matter in validation

Suppose user forgets to send `event type`.

Then this may happen:

```go
event.EventType == ""
```

That empty string is the zero value.

So validation checks often look for zero values.

---

## 4. Beginner-friendly explanation of methods

### What is a method?

A method is a function attached to a type.

Normal function:

```go
func Add(a int, b int) int {
	return a + b
}
```

Method on struct:

```go
type Person struct {
	Name string
	Age  int
}

func (p Person) Introduce() string {
	return "Hi, my name is " + p.Name
}
```

Use it:

```go
p := Person{Name: "Radhe", Age: 28}
fmt.Println(p.Introduce())
```

---

### Why use methods?

Because behavior should stay close to the data it belongs to.

Example:

If `PipelineEvent` needs validation, then this is clean:

```go
func (e PipelineEvent) Validate() error
```

instead of some random helper like:

```go
func ValidateEvent(eventType, status, repoURL, branch, author string) error
```

The method form is easier to read and maintain.

---

### Toy example: Rectangle with method

```go
package main

import "fmt"

type Rectangle struct {
	Width  float64
	Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func main() {
	rect := Rectangle{Width: 10, Height: 5}
	fmt.Println("Area:", rect.Area())
}
```

---

### Method receiver

This part:

```go
func (r Rectangle) Area() float64
```

means `Area` is a method on `Rectangle`.

`r` is called the receiver.

You can think of it like:

* `rect.Area()`
* method works using `rect` data

---

### In project terms

Your event model may have methods like:

* `Validate() error`
* `Summary() string`
* `ToSlackMessage() string`

That makes the model smarter and more useful.

---

## 5. Explanation of packages and why real projects split code into packages

### What is a package?

A package is a way to organize Go code.

All files in the same folder usually belong to one package.

Example:

```go
package model
```

or

```go
package slack
```

---

### Why packages matter

If all code is inside `main.go`, the project becomes messy very quickly.

Real projects split code so each package has a clear job.

Example structure:

```text
slack-integration/
├── cmd/
│   └── app/
│       └── main.go
├── internal/
│   ├── model/
│   │   └── event.go
│   ├── router/
│   │   └── router.go
│   ├── slack/
│   │   └── client.go
│   └── validator/
│       └── validator.go
├── go.mod
```

---

### Simple package responsibility idea

* `main` -> entry point
* `model` -> structs and model methods
* `router` -> chooses where to send event
* `slack` -> sends message to Slack
* `validator` -> extra validation helpers if needed

---

### Why split code?

Because each folder should answer one question:

* model -> what data are we working with?
* router -> where should it go?
* slack -> how do we send it?
* main -> how does the program start?

This makes code:

* easier to read
* easier to test
* easier to debug
* easier to extend

---

## 6. Explain the event model in the project

### What is an event model?

An event model is a struct that represents one event in a clean typed form.

In your CLI-based DevOps style project, raw input may come like this:

```bash
./app --event pull_request --status success --repo onboarding --branch feature-x --author radhe
```

These are raw strings from CLI flags.

The program should convert them into one typed model:

```go
type PipelineEvent struct {
	EventType string
	Status    string
	RepoName  string
	Branch    string
	Author    string
}
```

Now the event is one proper object.

---

### Why is this called a model?

Because it models a real business object.

In your case, that business object is:

* pipeline event
* notification request
* PR state update
* CI/CD status event

---

### Where the event model sits in full architecture

```text
User CLI Input
    |
    v
main.go parses flags
    |
    v
Build PipelineEvent struct
    |
    v
Validate PipelineEvent
    |
    v
Pass event to router/service
    |
    v
Build message / send to Slack / print / log
```

---

### Full architecture view

```text
+------------------+
| CLI User         |
| --event success  |
| --repo demo      |
+------------------+
         |
         v
+------------------+
| main.go          |
| parse flags      |
| build struct     |
+------------------+
         |
         v
+------------------+
| model package    |
| PipelineEvent    |
| Validate()       |
+------------------+
         |
         v
+------------------+
| router/service   |
| business logic   |
+------------------+
         |
         v
+------------------+
| slack/client     |
| send message     |
+------------------+
```

---

## 7. Show how `main.go` should build a `PipelineEvent` or similar model

Below is the basic idea.

### Raw values come from flags

```go
eventType := flag.String("event", "", "event type")
status := flag.String("status", "", "pipeline status")
repo := flag.String("repo", "", "repository name")
branch := flag.String("branch", "", "branch name")
author := flag.String("author", "", "author name")
```

### Build struct after parsing

```go
event := model.PipelineEvent{
	EventType: *eventType,
	Status:    *status,
	RepoName:  *repo,
	Branch:    *branch,
	Author:    *author,
}
```

### Validate

```go
if err := event.Validate(); err != nil {
	fmt.Println("validation error:", err)
	os.Exit(1)
}
```

---

## 8. Explain why passing a struct is better than passing many loose values

This is one of the biggest lessons today.

---

### Loose values approach

```go
func ProcessEvent(eventType string, status string, repo string, branch string, author string) {
	// process
}
```

Problem:

* too many arguments
* hard to remember order
* easy to mix values
* not scalable when fields grow
* less readable
* validation becomes messy

Bad call:

```go
ProcessEvent("success", "pull_request", "main", "demo-repo", "radhe")
```

Can you quickly tell what each value means? Hard.

---

### Struct approach

```go
func ProcessEvent(event PipelineEvent) {
	// process
}
```

Call:

```go
ProcessEvent(event)
```

This is better because:

* related data stays together
* easier to pass around
* easier to validate
* easier to extend with new fields
* better readability
* better testability

---

### Best mental model

Loose variables are like carrying 8 separate papers in your hand.

A struct is like putting all those papers into one labeled folder.

Much safer. Much cleaner.

---

## 9. Pseudocode first for event creation and validation

### Pseudocode

```text
START

Read CLI flags:
  event type
  status
  repo name
  branch
  author

Create PipelineEvent object with these values

Call Validate on PipelineEvent

If validation fails:
  print error
  exit program

If validation passes:
  print event summary
  send to next layer

END
```

---

### Validation pseudocode

```text
FUNCTION Validate(event):
  IF event type is empty:
      return error

  IF status is empty:
      return error

  IF repo name is empty:
      return error

  IF status is not one of [started, success, failed]:
      return error

  return nil
```

---

## 10. Real code examples with full explanation

---

### A. Toy example first

```go
package main

import "fmt"

type Student struct {
	Name  string
	Age   int
	City  string
}

func (s Student) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if s.Age <= 0 {
		return fmt.Errorf("age must be greater than 0")
	}
	return nil
}

func main() {
	st := Student{
		Name: "Radhe",
		Age:  25,
		City: "Chennai",
	}

	if err := st.Validate(); err != nil {
		fmt.Println("validation error:", err)
		return
	}

	fmt.Println("student is valid:", st)
}
```

### What this teaches

* `Student` is a struct
* `Validate()` is a method
* validation checks zero or invalid values
* method keeps logic close to the data

---

### B. Project-style example

#### File: `internal/model/event.go`

```go
package model

import (
	"fmt"
	"strings"
)

type PipelineEvent struct {
	EventType string
	Status    string
	RepoName  string
	Branch    string
	Author    string
}

func (e PipelineEvent) Validate() error {
	if strings.TrimSpace(e.EventType) == "" {
		return fmt.Errorf("event type is required")
	}
	if strings.TrimSpace(e.Status) == "" {
		return fmt.Errorf("status is required")
	}
	if strings.TrimSpace(e.RepoName) == "" {
		return fmt.Errorf("repo name is required")
	}
	if strings.TrimSpace(e.Branch) == "" {
		return fmt.Errorf("branch is required")
	}
	if strings.TrimSpace(e.Author) == "" {
		return fmt.Errorf("author is required")
	}

	switch e.Status {
	case "started", "success", "failed":
		return nil
	default:
		return fmt.Errorf("invalid status: %s", e.Status)
	}
}

func (e PipelineEvent) Summary() string {
	return fmt.Sprintf(
		"event=%s status=%s repo=%s branch=%s author=%s",
		e.EventType, e.Status, e.RepoName, e.Branch, e.Author,
	)
}
```

---

#### Full explanation

##### `package model`

This file belongs to the `model` package.

##### Struct

```go
type PipelineEvent struct {
	EventType string
	Status    string
	RepoName  string
	Branch    string
	Author    string
}
```

This defines the event shape.

##### Why these names start with capital letters

Because capitalized names are **exported** in Go.

That means other packages can use them.

So `main.go` can do this:

```go
model.PipelineEvent{}
```

If the name was lowercase like `pipelineEvent`, it would not be visible outside the package.

---

#### Exported vs unexported names

Very important rule:

* Capital letter -> exported
* Small letter -> unexported

Example:

```go
type PipelineEvent struct{}   // exported
type pipelineEvent struct{}   // unexported
```

```go
func BuildEvent() {}          // exported
func buildEvent() {}          // unexported
```

Use exported names when another package must use them.

Use unexported names for internal helpers.

---

#### Why `Validate()` is in the model

Because validation is directly about the model’s correctness.

The model itself should know:

* what fields are required
* which values are valid
* what makes it broken

That is why this is good:

```go
event.Validate()
```

instead of putting all validation randomly in `main.go`.

---

### C. `main.go` example

#### File: `cmd/app/main.go`

```go
package main

import (
	"flag"
	"fmt"
	"os"

	"slack-integration/internal/model"
)

func main() {
	eventType := flag.String("event", "", "event type")
	status := flag.String("status", "", "pipeline status")
	repo := flag.String("repo", "", "repository name")
	branch := flag.String("branch", "", "branch name")
	author := flag.String("author", "", "author name")

	flag.Parse()

	event := model.PipelineEvent{
		EventType: *eventType,
		Status:    *status,
		RepoName:  *repo,
		Branch:    *branch,
		Author:    *author,
	}

	if err := event.Validate(); err != nil {
		fmt.Println("validation error:", err)
		os.Exit(1)
	}

	fmt.Println("event created successfully")
	fmt.Println(event.Summary())
}
```

---

### Explanation of flow

```text
Flags -> raw string values -> PipelineEvent struct -> Validate() -> Summary()
```

This is exactly the shape of real CLI/backend tools.

---

### Example run

```bash
go run ./cmd/app --event pull_request --status success --repo onboarding --branch main --author radhe
```

Expected output:

```text
event created successfully
event=pull_request status=success repo=onboarding branch=main author=radhe
```

---

### Invalid example

```bash
go run ./cmd/app --event pull_request --status done --repo onboarding --branch main --author radhe
```

Output:

```text
validation error: invalid status: done
```

---

## 11. File-by-file explanation of the model package

Let us imagine a small model package.

```text
internal/model/
├── event.go
└── notification.go
```

---

### File 1: `event.go`

Purpose:

* define `PipelineEvent`
* define validation rules
* define helper methods like `Summary()`

Example responsibilities:

* event type
* status
* repo
* branch
* author

This file is the heart of “typed event input.”

---

### File 2: `notification.go`

Purpose:

* define `NotificationRequest`
* maybe convert event into a message-ready format

Example:

```go
package model

import (
	"fmt"
	"strings"
)

type NotificationRequest struct {
	Channel string
	Message string
}

func (n NotificationRequest) Validate() error {
	if strings.TrimSpace(n.Channel) == "" {
		return fmt.Errorf("channel is required")
	}
	if strings.TrimSpace(n.Message) == "" {
		return fmt.Errorf("message is required")
	}
	return nil
}
```

This shows how a second model can exist for the next step.

---

### Why this separation is useful

`PipelineEvent` = raw business event
`NotificationRequest` = message sending request

So one model represents “what happened,” and another represents “what to send.”

That is clean architecture thinking.

---

## 12. Hands-on tasks for today

Do these in order.

### Task 1 — Create a toy struct

Create a `Student` struct with:

* Name
* Age
* City

Then print one student object.

---

### Task 2 — Add a method

Add a method:

```go
func (s Student) Introduce() string
```

Print something like:

```text
Hi, I am Radhe from Chennai
```

---

### Task 3 — Create `PipelineEvent`

Create a `PipelineEvent` struct with:

* EventType
* Status
* RepoName
* Branch
* Author

---

### Task 4 — Add `Validate()` method

Rules:

* all fields required
* status must be one of:

  * started
  * success
  * failed

---

### Task 5 — Build event from CLI flags

Read flags from `main.go`, build `PipelineEvent`, validate it, then print summary.

---

### Task 6 — Add `NotificationRequest`

Create another model:

* Channel
* Message

Add `Validate()` method.

---

## 13. Expected output

### Task 1 output idea

```text
{Name:Radhe Age:25 City:Chennai}
```

---

### Task 2 output idea

```text
Hi, I am Radhe from Chennai
```

---

### Task 5 valid run

Command:

```bash
go run . --event pull_request --status success --repo onboarding --branch main --author radhe
```

Output:

```text
event created successfully
event=pull_request status=success repo=onboarding branch=main author=radhe
```

---

### Task 5 invalid run

Command:

```bash
go run . --event pull_request --status done --repo onboarding --branch main --author radhe
```

Output:

```text
validation error: invalid status: done
exit status 1
```

---

## 14. Common mistakes

Here are the most common beginner mistakes today:

### 1. Forgetting capital letters for exported names

Wrong if used outside package:

```go
type pipelineEvent struct {}
```

Right:

```go
type PipelineEvent struct {}
```

---

### 2. Forgetting to call `flag.Parse()`

If you do not call it, your flags may stay empty.

---

### 3. Confusing struct field names with local variables

This is okay:

```go
event := model.PipelineEvent{
	EventType: *eventType,
}
```

But beginners sometimes mix field name and variable name.

---

### 4. Ignoring zero values

If a user does not provide a flag, the value may become empty string `""`.

You must validate it.

---

### 5. Putting all validation in `main.go`

This makes `main.go` too big and messy.

---

### 6. Passing too many loose variables

This makes functions hard to understand.

---

### 7. Using package names wrongly

Folder name and package name should make sense.

---

## 15. Debugging tips

### Tip 1 — Print the struct

Before validation, print the struct:

```go
fmt.Printf("%+v\n", event)
```

This shows field names and values.

Example:

```text
{EventType:pull_request Status:success RepoName:onboarding Branch:main Author:radhe}
```

---

### Tip 2 — Print raw flag values

```go
fmt.Println("event:", *eventType)
fmt.Println("status:", *status)
```

This helps confirm CLI parsing is working.

---

### Tip 3 — Check zero values

If something prints blank, it may be `""`.

---

### Tip 4 — Validate one rule at a time

Do not write 20 rules at once.

Start with:

* event required
* status required

Then add the rest.

---

### Tip 5 — Test invalid input deliberately

Test cases:

* missing author
* empty repo
* bad status

This helps you trust your validation logic.

---

## 16. One DSA topic — String basics in simple language

Today’s DSA topic: **strings in Go**

### What is a string?

A string is text.

Example:

```go
name := "Radhe"
```

---

### Common string basics

#### Length

```go
fmt.Println(len(name))
```

#### Compare

```go
if name == "Radhe" {
	fmt.Println("matched")
}
```

#### Join strings

```go
full := "Hello " + name
```

#### Loop through characters

Simple way:

```go
for i := 0; i < len(name); i++ {
	fmt.Println(string(name[i]))
}
```

For beginners, this is enough for now.

---

### Why strings matter in backend work

You will use strings for:

* event type
* repo name
* branch name
* author
* messages
* Slack content
* JSON fields
* log messages

So string basics are directly useful in your project.

---

## 17. One easy Go DSA problem

### Problem: Count vowels in a string

Given a string, count how many vowels are present.

Example:

```text
Input:  "radhe"
Output: 2
```

Because vowels are `a` and `e`.

---

### Simple solution in Go

```go
package main

import "fmt"

func countVowels(s string) int {
	count := 0

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == 'a' || ch == 'e' || ch == 'i' || ch == 'o' || ch == 'u' ||
			ch == 'A' || ch == 'E' || ch == 'I' || ch == 'O' || ch == 'U' {
			count++
		}
	}

	return count
}

func main() {
	fmt.Println(countVowels("radhe"))
}
```

Output:

```text
2
```

---

### Why this problem is good for Day 2

Because it helps you practice:

* strings
* loops
* conditions
* function writing

---

## 18. One module-based practice task

Create this small package:

```text
internal/model/
└── notification_request.go
```

### Requirement

Create a model:

```go
type NotificationRequest struct {
	Channel string
	Message string
}
```

Add methods:

* `Validate() error`
* `Summary() string`

### Rules

* `Channel` cannot be empty
* `Message` cannot be empty

---

### Sample code

```go
package model

import (
	"fmt"
	"strings"
)

type NotificationRequest struct {
	Channel string
	Message string
}

func (n NotificationRequest) Validate() error {
	if strings.TrimSpace(n.Channel) == "" {
		return fmt.Errorf("channel is required")
	}
	if strings.TrimSpace(n.Message) == "" {
		return fmt.Errorf("message is required")
	}
	return nil
}

func (n NotificationRequest) Summary() string {
	return fmt.Sprintf("channel=%s message=%s", n.Channel, n.Message)
}
```

### Practice from `main.go`

```go
req := model.NotificationRequest{
	Channel: "slack-devops",
	Message: "Pipeline completed successfully",
}

if err := req.Validate(); err != nil {
	fmt.Println("validation error:", err)
	return
}

fmt.Println(req.Summary())
```

---

## 19. Revision checkpoint

Before ending Day 2, make sure you can answer these:

### Structs

* What is a struct?
* Why is struct better than loose variables?
* What are zero values?

### Methods

* What is a method?
* Why attach `Validate()` to a struct?

### Packages

* Why do we split code into packages?
* What is exported vs unexported?

### Project architecture

* How does raw CLI input become a typed model?
* Why does validation belong in the model layer?

### DSA

* How do you loop through a string in Go?
* How would you count vowels?

---

## 20. Homework

Do these carefully.

### Homework 1

Create a `PipelineEvent` struct and `Validate()` method.

---

### Homework 2

Add a `Summary()` method.

Example output:

```text
event=pull_request status=started repo=demo branch=feature-1 author=radhe
```

---

### Homework 3

Create a second model called `NotificationRequest`.

Fields:

* Channel
* Message

Add validation.

---

### Homework 4

From `main.go`, first build `PipelineEvent`, validate it, then create `NotificationRequest` from it.

Flow:

```text
CLI input -> PipelineEvent -> Validate -> build message -> NotificationRequest -> Validate
```

---

### Homework 5

Solve this string problem in Go:

**Count how many times the letter `a` appears in a string**

Example:

```text
Input:  "banana"
Output: 3
```

---

# Final mental model for Day 2

Today’s biggest idea is this:

```text
raw input is messy
typed struct is clean
validation protects the system
packages keep code organized
methods keep behavior near data
```

And in project form:

```text
CLI flags
   |
   v
main.go
   |
   v
build PipelineEvent
   |
   v
event.Validate()
   |
   v
pass clean event to next layer
   |
   v
router / notifier / Slack client
```

That is exactly how a real backend or DevOps CLI starts becoming production-friendly.

When you are ready, send **Day 3 prompt**.
