# Day 1 — Golang CLI Foundations from a `slack-integration` Style Project

## 1. Day 1 learning goals

By the end of Day 1, you should understand:

* what kind of problem a `slack-integration` style CLI project solves
* how a Go project is usually organized
* what `package main`, `func main()`, `go.mod`, and `import` do
* how CLI input enters the program using flags
* how data flows from terminal input into Go code
* why this style of project looks like a real production tool
* the difference between arrays and slices in Go
* how to build a tiny module-based config loader / input parser

---

## 2. What I should already know before starting

You do **not** need much before Day 1.

Helpful basics:

* how to open terminal
* how to create files and folders
* basic programming idea: input → processing → output
* very basic Python understanding is enough

That is enough to begin.

---

## 3. Full beginner-friendly explanation of the project

Let us first understand the type of project you are learning from.

A `slack-integration` style project usually means:

* some event happens
* your program receives input about that event
* your program understands the input
* your program decides what message to build
* your program sends or prints a message

### Very simple real-world idea

Imagine this:

* A GitHub Pull Request is opened
* A pipeline starts
* Your system wants to notify Slack
* A CLI tool helps format, validate, or route that information

So the project is not just “a Go program.”
It is closer to a **small automation tool** used in DevOps workflows.

### What problem this project solves

Without such a tool, teams often do things manually:

* copy PR details manually
* build Slack messages manually
* handle pipeline notifications manually
* repeat the same steps again and again

With a CLI-based automation tool:

* input becomes standardized
* messages become consistent
* logic becomes reusable
* later integration with Tekton/Kubernetes/Slack becomes easier

### Why this matters for Cloud Resource Onboarding style POC

In a Cloud Resource Onboarding POC pattern, the same idea exists:

* user or system provides input
* tool validates that input
* tool converts input into a structured format
* tool triggers the next system action

So both patterns share the same backbone:

```text
Input → Parse → Validate → Transform → Output / Trigger Next Step
```

That is why Day 1 is important.
If you understand this flow now, later Tekton, Kubernetes, webhook, and Slack steps will feel much easier.

---

## 4. Project folder structure explanation

Here is a simple beginner-friendly folder structure inspired by a modular CLI project.

```text
slack-integration/
│
├── go.mod
├── main.go
├── README.md
│
├── cmd/
│   └── root.go
│
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── cli/
│   │   └── flags.go
│   ├── model/
│   │   └── event.go
│   ├── parser/
│   │   └── parser.go
│   └── output/
│       └── printer.go
│
└── sample/
    └── input.json
```

Now let us understand it simply.

### Meaning of each folder

* `go.mod`
  tells Go this is a module/project and manages dependencies

* `main.go`
  entry point of the program; execution starts here

* `cmd/`
  often used for CLI command setup in bigger projects

* `internal/config/`
  reads environment variables or default config values

* `internal/cli/`
  handles command-line flags like `--event`, `--user`

* `internal/model/`
  defines structured data types like Event, Config, Message

* `internal/parser/`
  converts raw input into clean structured data

* `internal/output/`
  prints output or later sends output to Slack/API

* `sample/`
  contains example test input files

### Why this is good

Because it separates concerns.

Instead of writing everything in one file, you split responsibilities.

That is how real production systems stay maintainable.

---

## 5. File-by-file explanation of the important files

## `go.mod`

This is the identity card of your project.

Example:

```go
module slack-integration

go 1.22
```

It tells Go:

* module name = `slack-integration`
* Go version = `1.22`

Later, if you use external libraries, they also get tracked here.

---

## `main.go`

This is the starting point.

It usually does this:

* load CLI inputs
* validate them
* call other functions/modules
* print or send output

Think of `main.go` like the **reception desk** of the application.

---

## `internal/cli/flags.go`

This file handles terminal inputs such as:

```bash
go run main.go --event pr_opened --user radhe --repo cloud-onboarding
```

It reads those values and stores them in a structured format.

---

## `internal/model/event.go`

This file defines your data structure.

Example:

```go
type Event struct {
    EventType string
    User      string
    Repo      string
}
```

This is better than passing random strings everywhere.

---

## `internal/parser/parser.go`

This file transforms raw CLI input into meaningful data.

Example:

* raw input says `pr_opened`
* parser may convert it into a clean event object

---

## `internal/config/config.go`

This file loads environment variables or defaults.

Example:

* Slack webhook URL
* app environment
* default channel name

For Day 1, we will keep it simple.

---

## `internal/output/printer.go`

For now, it may just print output.

Later, instead of printing, it could:

* send HTTP request
* call Slack webhook
* log to monitoring system

That is why printing is a useful first step.

---

## 6. How `main.go` works

Let us first see the mental model.

### Flow of `main.go`

```text
Program starts
   ↓
Read flags from terminal
   ↓
Validate the values
   ↓
Build structured input object
   ↓
Pass object to processing function
   ↓
Print result
```

### ASCII diagram

```text
+------------------+
|  Terminal Input  |
| --event --user   |
+---------+--------+
          |
          v
+------------------+
|     main.go      |
| entry point      |
+---------+--------+
          |
          v
+------------------+
|  CLI Flag Parser |
+---------+--------+
          |
          v
+------------------+
| Structured Event |
+---------+--------+
          |
          v
+------------------+
| Output / Message |
+------------------+
```

### Very small example

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello from Go CLI")
}
```

### Line-by-line

`package main`

* every Go file belongs to a package
* if you want to build an executable program, it must use `package main`

`import "fmt"`

* imports the formatting package
* used for printing text

`func main()`

* the first function that runs in an executable Go program

`fmt.Println(...)`

* prints a line to the terminal

---

## 7. Very simple Go basics first

Before CLI, let us understand a few core things.

## A. Package

A package is a group of related Go files.

Example:

* `package main` for executable program
* `package config` for config-related logic

Think of package as a folder-level identity.

---

## B. Function

A function is a reusable block of code.

```go
func greet() {
    fmt.Println("Hello")
}
```

---

## C. Variable

```go
name := "Radhe"
```

This creates a variable.

`:=` means Go infers the type automatically.

---

## D. Types

Common beginner types:

* `string`
* `int`
* `bool`

Example:

```go
user := "radhe"
count := 5
isValid := true
```

---

## E. Struct

Struct means a grouped data object.

```go
type Event struct {
    EventType string
    User      string
}
```

This is like a lightweight class-like data container.

---

## F. Import

You import packages when you want to use their functionality.

```go
import "fmt"
```

`fmt` is used for printing.

---

## G. `go.mod`

This tells Go:

* project name
* dependencies
* version info

Without it, modern Go project management becomes harder.

---

## 8. Then explain how CLI flags work in Go

A CLI flag is input passed from terminal.

Example:

```bash
go run main.go --user radhe --event pr_opened
```

Here:

* `--user` is a flag
* `radhe` is the value
* `--event` is another flag
* `pr_opened` is its value

### Why flags are useful

Because they make CLI tools flexible.

Instead of hardcoding values in code, user passes values from terminal.

### Very small idea

```text
Terminal command → program reads flags → program uses those values
```

### Go package for flags

Go provides built-in `flag` package.

Simple example:

```go
package main

import (
    "flag"
    "fmt"
)

func main() {
    user := flag.String("user", "guest", "name of the user")
    event := flag.String("event", "unknown", "type of event")

    flag.Parse()

    fmt.Println("User:", *user)
    fmt.Println("Event:", *event)
}
```

### Important point

`flag.String(...)` returns a **pointer**, so you read value using `*user` and `*event`.

For now, remember it like this:

* `flag.String(...)` gives you a stored reference
* `*user` gives you the actual value

We will keep pointer discussion light for Day 1.

---

## 9. Pseudocode first

Now let us build a small config loader / CLI parser mentally.

### Pseudocode

```text
START

Read user flag
Read event flag
Read repo flag

If user is empty
    print error
    stop

If event is empty
    print error
    stop

Create Event object with user, event, repo

Print event details

END
```

### Slightly more modular pseudocode

```text
main():
    config = readFlags()
    validate(config)
    event = buildEvent(config)
    printResult(event)
```

This is exactly how production systems grow:

* read input
* validate input
* convert to internal model
* process it
* output result

---

## 10. Then real Go code examples

## Example 1: simplest CLI

```go
package main

import (
    "flag"
    "fmt"
)

func main() {
    user := flag.String("user", "guest", "name of the user")
    event := flag.String("event", "unknown", "type of event")

    flag.Parse()

    fmt.Println("CLI Input Received")
    fmt.Println("User:", *user)
    fmt.Println("Event:", *event)
}
```

### Explain every important line

`import ("flag" "fmt")`

* `flag` reads CLI arguments
* `fmt` prints text

`user := flag.String("user", "guest", "name of the user")`

* creates a `user` flag
* if not passed, default value is `"guest"`
* description is `"name of the user"`

`flag.Parse()`

* tells Go to actually read terminal flags

`fmt.Println("User:", *user)`

* prints actual value of user flag

---

## Example 2: add validation

```go
package main

import (
    "flag"
    "fmt"
)

func main() {
    user := flag.String("user", "", "name of the user")
    event := flag.String("event", "", "type of event")
    repo := flag.String("repo", "", "repository name")

    flag.Parse()

    if *user == "" {
        fmt.Println("Error: --user is required")
        return
    }

    if *event == "" {
        fmt.Println("Error: --event is required")
        return
    }

    fmt.Println("Input looks good")
    fmt.Println("User:", *user)
    fmt.Println("Event:", *event)
    fmt.Println("Repo:", *repo)
}
```

### Why this matters

Production tools should not trust input blindly.

Validation is one of the first important engineering habits.

---

## Example 3: use a struct for cleaner design

```go
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
    user := flag.String("user", "", "name of the user")
    event := flag.String("event", "", "type of event")
    repo := flag.String("repo", "", "repository name")

    flag.Parse()

    if *user == "" || *event == "" {
        fmt.Println("Error: --user and --event are required")
        return
    }

    inputEvent := Event{
        User:      *user,
        EventType: *event,
        Repo:      *repo,
    }

    fmt.Println("Structured Event Created")
    fmt.Printf("%+v\n", inputEvent)
}
```

### What `%+v` does

It prints struct fields with names.

Example output:

```text
{User:radhe EventType:pr_opened Repo:cloud-onboarding}
```

---

## Example 4: module-based practice task — config loader / CLI input parser

Here is a simple modular structure.

### `main.go`

```go
package main

import (
    "fmt"
    "slack-integration/internal/cli"
    "slack-integration/internal/parser"
)

func main() {
    input := cli.ReadFlags()

    event, err := parser.BuildEvent(input)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    fmt.Println("Parsed Event:")
    fmt.Printf("%+v\n", event)
}
```

---

### `internal/cli/flags.go`

```go
package cli

import "flag"

type InputFlags struct {
    User  string
    Event string
    Repo  string
}

func ReadFlags() InputFlags {
    user := flag.String("user", "", "name of the user")
    event := flag.String("event", "", "type of event")
    repo := flag.String("repo", "", "repository name")

    flag.Parse()

    return InputFlags{
        User:  *user,
        Event: *event,
        Repo:  *repo,
    }
}
```

---

### `internal/parser/parser.go`

```go
package parser

import (
    "errors"
    "slack-integration/internal/cli"
)

type Event struct {
    User      string
    EventType string
    Repo      string
}

func BuildEvent(input cli.InputFlags) (Event, error) {
    if input.User == "" {
        return Event{}, errors.New("user is required")
    }

    if input.Event == "" {
        return Event{}, errors.New("event is required")
    }

    event := Event{
        User:      input.User,
        EventType: input.Event,
        Repo:      input.Repo,
    }

    return event, nil
}
```

### Why this is better

Because now:

* `cli` package only reads input
* `parser` package validates and builds event
* `main.go` only coordinates flow

This is how clean software starts.

---

## 11. Hands-on tasks for today

## Task 1 — Run a basic Go program

Create `main.go`:

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, Day 1 Go")
}
```

Run:

```bash
go run main.go
```

---

## Task 2 — Read one CLI flag

Update code to read `--user`.

---

## Task 3 — Read multiple flags

Read:

* `--user`
* `--event`
* `--repo`

---

## Task 4 — Add validation

If `--user` or `--event` is missing, print error.

---

## Task 5 — Create an `Event` struct

Store the CLI input in a structured object.

---

## Task 6 — Make a tiny module-based input parser

Split:

* flag reading into one package
* parsing/validation into another package

This is your Day 1 real-system-inspired task.

---

## 12. Expected output for each task

## Task 1 output

```text
Hello, Day 1 Go
```

## Task 2 output

Command:

```bash
go run main.go --user radhe
```

Output:

```text
User: radhe
```

## Task 3 output

Command:

```bash
go run main.go --user radhe --event pr_opened --repo cloud-onboarding
```

Output:

```text
User: radhe
Event: pr_opened
Repo: cloud-onboarding
```

## Task 4 output when missing value

Command:

```bash
go run main.go --user radhe
```

Output:

```text
Error: --event is required
```

## Task 5 output

```text
Structured Event Created
{User:radhe EventType:pr_opened Repo:cloud-onboarding}
```

## Task 6 output

```text
Parsed Event:
{User:radhe EventType:pr_opened Repo:cloud-onboarding}
```

---

## 13. Common mistakes beginners make

## Mistake 1: forgetting `flag.Parse()`

If you do not call `flag.Parse()`, your flags will not be read properly.

---

## Mistake 2: forgetting `*` when printing flag value

Wrong:

```go
fmt.Println(user)
```

This prints pointer info.

Correct:

```go
fmt.Println(*user)
```

---

## Mistake 3: putting all code in one file

It works for tiny demos, but becomes hard to maintain.

---

## Mistake 4: not validating input

Production tools should never assume input is correct.

---

## Mistake 5: confusing package names and folder names

In Go, the package name inside the file should match your design clearly.

---

## Mistake 6: wrong module import path

If your module name in `go.mod` is:

```go
module slack-integration
```

then imports should use:

```go
"slack-integration/internal/cli"
```

---

## 14. Debugging tips

## Tip 1: print intermediate values

Use:

```go
fmt.Printf("%+v\n", myStruct)
```

This is very useful.

---

## Tip 2: test one layer at a time

First test:

* can flags be read?

Then test:

* can parsing work?

Then test:

* does `main.go` connect everything?

---

## Tip 3: check file/package mismatch

Example problem:

* file is inside `internal/cli`
* but package name is something else by mistake

---

## Tip 4: check `go.mod`

If imports fail, first inspect your module name.

---

## Tip 5: keep commands simple first

Do not start with advanced flags.
First confirm:

```bash
go run main.go --user radhe --event pr_opened
```

---

## 15. One small DSA topic for today

# Slices vs Arrays in Go

This is very important and very beginner-friendly.

## Array

An array has a fixed size.

Example:

```go
var nums [3]int = [3]int{10, 20, 30}
```

This means:

* size is exactly 3
* cannot easily grow

## Slice

A slice is flexible and more commonly used.

Example:

```go
nums := []int{10, 20, 30}
```

This means:

* dynamic-sized view over data
* can grow with `append`

Example:

```go
nums = append(nums, 40)
```

Now slice becomes:

```text
[10 20 30 40]
```

## Easy memory idea

* array = fixed-size box
* slice = expandable list

## Why slices matter in real systems

In CLI and DevOps tools, you often handle:

* list of repos
* list of environments
* list of changed files
* list of arguments

These are usually slices, not arrays.

### Tiny comparison

```go
package main

import "fmt"

func main() {
    arr := [3]int{1, 2, 3}
    slice := []int{1, 2, 3}

    fmt.Println("Array:", arr)
    fmt.Println("Slice:", slice)

    slice = append(slice, 4)
    fmt.Println("Slice after append:", slice)
}
```

---

## 16. One small Golang DSA practice problem

# Problem: Read numbers and print their sum using a slice

### Problem statement

Create a slice of integers and print the sum of all elements.

### Pseudocode

```text
Create slice with numbers
Set sum = 0
Loop through each number
    add number to sum
Print sum
```

### Code

```go
package main

import "fmt"

func main() {
    nums := []int{5, 10, 15, 20}
    sum := 0

    for _, num := range nums {
        sum += num
    }

    fmt.Println("Sum:", sum)
}
```

### Output

```text
Sum: 50
```

### Why this matters

This teaches:

* slice usage
* looping
* accumulation logic

These are basic but essential.

---

## 17. One module-based practice task inspired by real systems

# Practice task: Build a small config loader / CLI input parser

Your goal:

Create a small Go CLI that accepts:

* `--user`
* `--event`
* `--env`

Then:

* validate required fields
* store values in a struct
* print structured config

## Suggested structure

```text
day1-cli/
├── go.mod
├── main.go
└── internal/
    ├── cli/
    │   └── flags.go
    └── config/
        └── config.go
```

## Example code

### `internal/cli/flags.go`

```go
package cli

import "flag"

type RawInput struct {
    User  string
    Event string
    Env   string
}

func ReadFlags() RawInput {
    user := flag.String("user", "", "user name")
    event := flag.String("event", "", "event type")
    env := flag.String("env", "dev", "environment name")

    flag.Parse()

    return RawInput{
        User:  *user,
        Event: *event,
        Env:   *env,
    }
}
```

### `internal/config/config.go`

```go
package config

import (
    "errors"
    "day1-cli/internal/cli"
)

type AppConfig struct {
    User      string
    EventType string
    Env       string
}

func Load(input cli.RawInput) (AppConfig, error) {
    if input.User == "" {
        return AppConfig{}, errors.New("user is required")
    }

    if input.Event == "" {
        return AppConfig{}, errors.New("event is required")
    }

    cfg := AppConfig{
        User:      input.User,
        EventType: input.Event,
        Env:       input.Env,
    }

    return cfg, nil
}
```

### `main.go`

```go
package main

import (
    "fmt"
    "day1-cli/internal/cli"
    "day1-cli/internal/config"
)

func main() {
    raw := cli.ReadFlags()

    cfg, err := config.Load(raw)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    fmt.Println("Loaded config successfully")
    fmt.Printf("%+v\n", cfg)
}
```

## Example run

```bash
go run main.go --user radhe --event pr_opened --env dev
```

## Output

```text
Loaded config successfully
{User:radhe EventType:pr_opened Env:dev}
```

## Why this is inspired by real systems

Because real systems rarely work directly with raw input.
They usually:

* read raw input
* normalize it
* validate it
* build structured config
* pass clean config to business logic

That is exactly what you are doing here.

---

## 18. Revision checklist for Day 1

Use this checklist before ending today.

### I understand:

* [ ] what a `slack-integration` style CLI project does
* [ ] the basic input → parse → validate → output flow
* [ ] what `package main` means
* [ ] what `func main()` does
* [ ] what `go.mod` is for
* [ ] how `import` works
* [ ] how to read CLI flags using `flag`
* [ ] why validation is important
* [ ] how to store input in a struct
* [ ] why modular folder structure is useful
* [ ] the difference between arrays and slices
* [ ] how a small config loader works

---

## 19. Small homework

Do these 4 small tasks.

### Homework 1

Create a CLI that accepts:

* `--user`
* `--event`
* `--repo`
* `--branch`

Print them in a struct.

---

### Homework 2

Make `--user` and `--event` mandatory.

---

### Homework 3

If `--env` is not passed, default it to `dev`.

---

### Homework 4

Add one function:

```go
func PrintSummary(...)
```

It should print:

```text
User radhe triggered event pr_opened in repo cloud-onboarding on env dev
```

---

# How today connects to your `slack-integration` learning journey

Today may look simple, but it is the foundation of everything later.

## Today you learned:

* how input enters a Go CLI
* how Go project structure starts
* how to convert raw input into structured objects

## Later in Slack integration, this becomes:

* CLI input becomes GitHub/Tekton/webhook input
* simple validation becomes production validation
* printed output becomes Slack message payload
* config loader becomes environment/secret loader
* struct becomes pipeline event model

## Later in Tekton/Kubernetes/Slack integration, this helps because:

* Tekton tasks pass parameters
* Kubernetes workloads use environment/config
* Slack integration needs clean payload building
* production systems need modular design, not one-file scripts

### Future mental bridge

```text
Day 1:
CLI flags → struct → print

Later:
Webhook/Tekton params → event model → Slack payload/API call
```

That is the same engineering pattern, only with bigger integrations.

---

# One final simple summary

Day 1 is about learning this single powerful idea:

```text
A production CLI is just a clean pipeline:
take input, understand it, validate it, structure it, and prepare it for action.
```

That is the foundation of your `slack-integration` and Cloud Resource Onboarding journey.

If you want, next I can give you **Day 2** in the same style, focused on structs, methods, modules, validation flow, and better project design.
