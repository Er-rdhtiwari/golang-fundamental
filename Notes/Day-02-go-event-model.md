# Day 2 — Golang Structs, Methods, Packages, and Event Model

Today we move from “basic CLI flow” to “clean project structure.”

On Day 1, input came into the program through flags.
On Day 2, we learn how to **convert that raw input into a proper typed model** like `PipelineEvent`, validate it, and pass it cleanly through the project.

---

## 1. Day 2 learning goals

By the end of Day 2, you should understand:

1. What a **struct** is in Go
2. How **methods** work on structs
3. How **packages** help organize real backend/CLI code
4. What **exported vs unexported** names mean
5. How a project uses a model like `PipelineEvent`
6. How raw CLI input becomes a **typed event object**
7. Why validation should live in the **model layer**
8. Why a struct is better than passing many loose variables
9. How this pattern relates to real production backend projects
10. How Go ideas compare with Python classes, methods, and modules

---

## 2. Quick revision of Day 1 in 5–8 points

Here is the Day 1 revision in simple form:

1. A Go program starts from `package main` and `func main()`
2. `go.mod` defines the module name and dependencies
3. CLI flags are used to accept input from the terminal
4. Raw input first enters the program as strings, bools, ints, etc.
5. `main.go` usually acts as the **entry point**, not the full business logic
6. A real CLI should separate input parsing from validation and processing
7. Go is strongly typed, so we prefer clear structured data over random loose values
8. Production-style CLI tools are easier to maintain when logic is split into packages

---

## 3. Beginner-friendly explanation of structs using simple examples first

## What is a struct?

A **struct** is a way to group related data together into one typed object.

In Python, this feels like:

* a class with attributes
* or sometimes a dictionary, but more structured and safer

### Python example

```python
class User:
    def __init__(self, name, age, active):
        self.name = name
        self.age = age
        self.active = active
```

### Go equivalent

```go
type User struct {
	Name   string
	Age    int
	Active bool
}
```

This means:

* `User` is a new type
* it contains 3 fields
* every field has a fixed type

---

## Why structs matter

Without structs, you might pass values like this:

```go
name := "Radhe"
age := 30
active := true
```

That works for tiny examples.

But in a real project, imagine passing:

* event type
* status
* repo URL
* branch name
* author
* commit SHA
* environment
* timestamp

Passing all of them separately becomes messy very quickly.

That is why we use a struct.

---

## Toy example: Student

```go
package main

import "fmt"

type Student struct {
	Name   string
	Course string
	Age    int
}

func main() {
	s := Student{
		Name:   "Radhe",
		Course: "Go Basics",
		Age:    28,
	}

	fmt.Println(s.Name)
	fmt.Println(s.Course)
	fmt.Println(s.Age)
}
```

### Output

```go
Radhe
Go Basics
28
```

---

## Zero values in structs

This is very important in Go.

If you create a struct and do not set fields, Go gives **zero values** automatically.

```go
type Student struct {
	Name   string
	Age    int
	Active bool
}
```

If you write:

```go
var s Student
```

Then the fields become:

* `Name` → `""`
* `Age` → `0`
* `Active` → `false`

### Example

```go
package main

import "fmt"

type Student struct {
	Name   string
	Age    int
	Active bool
}

func main() {
	var s Student
	fmt.Printf("%q %d %v\n", s.Name, s.Age, s.Active)
}
```

### Output

```go
"" 0 false
```

### Python difference

In Python, if you do not define values properly, you often get:

* `AttributeError`
* or `None`
* or dynamic behavior

In Go, every field has a default zero value.
This is useful, but it can also hide bugs if you forget validation.

Example:

* `Status == ""` is possible
* `RepoURL == ""` is possible
* `PRNumber == 0` is possible

That is why validation matters.

---

## 4. Beginner-friendly explanation of methods

A **method** is a function attached to a type.

In Python:

```python
class User:
    def greet(self):
        print("hello")
```

In Go:

```go
type User struct {
	Name string
}

func (u User) Greet() {
	fmt.Println("Hello,", u.Name)
}
```

### Full example

```go
package main

import "fmt"

type User struct {
	Name string
}

func (u User) Greet() {
	fmt.Println("Hello,", u.Name)
}

func main() {
	u := User{Name: "Radhe"}
	u.Greet()
}
```

### Output

```go
Hello, Radhe
```

---

## Method receiver

This part:

```go
func (u User) Greet()
```

means `Greet` is a method on `User`.

`u` is called the **receiver**.

---

## Value receiver vs pointer receiver

You do not need deep mastery today, but know this basic rule:

### Value receiver

Gets a copy.

```go
func (u User) PrintName() {
	fmt.Println(u.Name)
}
```

### Pointer receiver

Can modify the original struct.

```go
func (u *User) UpdateName(newName string) {
	u.Name = newName
}
```

### Example

```go
package main

import "fmt"

type User struct {
	Name string
}

func (u *User) UpdateName(newName string) {
	u.Name = newName
}

func main() {
	u := User{Name: "Old"}
	u.UpdateName("New")
	fmt.Println(u.Name)
}
```

### Output

```go
New
```

### Python comparison

Python objects are reference-like by default, so mutation feels natural.

Go makes this more explicit:

* value receiver = copy-like behavior
* pointer receiver = modify original

This explicitness is one of Go’s style differences.

---

## 5. Explanation of packages and why real projects split code into packages

## What is a package?

A package is a way to group related code files together.

In Python, this is similar to:

* modules
* packages
* folders with reusable logic

### Example mental mapping

| Python              | Go              |
| ------------------- | --------------- |
| `utils.py`          | package `utils` |
| `models.py`         | package `model` |
| `services/slack.py` | package `slack` |
| `main.py`           | `package main`  |

---

## Why split code into packages?

Because real projects should not keep everything inside one file.

That becomes hard to read, test, and maintain.

### Good separation

* `main.go` → input and wiring
* `model` → data structures and validation
* `router` → decision logic
* `slack` → Slack sending logic
* `githubclient` → GitHub API calls

---

## Example project layout

```text
slack-integration/
├── go.mod
├── cmd/
│   └── cli/
│       └── main.go
├── internal/
│   ├── model/
│   │   ├── pipeline_event.go
│   │   ├── validation.go
│   │   └── notification_request.go
│   ├── router/
│   │   └── router.go
│   └── slack/
│       └── client.go
```

You can also keep it simpler in the beginning:

```text
slack-integration/
├── go.mod
├── main.go
├── model/
│   ├── pipeline_event.go
│   └── validation.go
├── router/
│   └── router.go
└── slack/
    └── client.go
```

---

## Exported vs unexported names

This is a very important Go rule.

### Exported name

Starts with a capital letter.

```go
type PipelineEvent struct {
	EventType string
	Status    string
}
```

These can be accessed from other packages.

---

### Unexported name

Starts with a small letter.

```go
type pipelineConfig struct {
	retryCount int
}
```

These are private to the package.

---

## Example

Inside package `model`:

```go
package model

type PipelineEvent struct {
	EventType string
	Status    string
}

func (e PipelineEvent) Validate() error {
	return nil
}
```

From `main.go`, you can use:

```go
event := model.PipelineEvent{}
err := event.Validate()
```

But if the field or function name starts with lowercase, other packages cannot access it.

---

## Python comparison

Python has convention-based privacy:

* `_name` means “internal”
* but still accessible

Go has package-level visibility based on capitalization:

* `Name` = public/exported
* `name` = private/unexported

This is a major convention change from Python.

---

## 6. Explain the event model in the project

Now let us connect this to your CLI + Slack-style project.

The program receives raw input such as:

* event type
* pipeline name
* status
* repo
* branch
* commit SHA
* author
* message

That input is first **plain raw data**.

But the project should not pass these as loose strings everywhere.

Instead, it should convert them into one model object:

```go
type PipelineEvent struct {
	EventType string
	Pipeline  string
	Status    string
	Repo      string
	Branch    string
	CommitSHA string
	Author    string
	Message   string
}
```

This object becomes the **event model**.

---

## Why is it called an event model?

Because it represents one meaningful thing that happened in the system.

Examples:

* PR opened
* PR merged
* pipeline started
* pipeline failed
* deployment succeeded

Instead of saying:

> here are 8 random strings

we say:

> here is one `PipelineEvent`

That is much cleaner.

---

## Where the event model sits in the architecture

```text
CLI Flags / Raw Input
        |
        v
    main.go
        |
        v
Build PipelineEvent struct
        |
        v
Validate event in model layer
        |
        v
Router decides where to send it
        |
        v
Slack client / other integrations
```

---

## Full architecture position

```text
+------------------+
| Terminal / User  |
+------------------+
         |
         v
+------------------+
| CLI flags        |
| --event          |
| --status         |
| --repo           |
+------------------+
         |
         v
+------------------+
| main.go          |
| parse input      |
| create model     |
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
| router package   |
| route by type    |
| route by status  |
+------------------+
         |
         v
+------------------+
| slack package    |
| send message     |
+------------------+
```

---

## 7. Show how `main.go` should build a `PipelineEvent` or similar model

The job of `main.go` should be:

1. Read flags
2. Put them into a struct
3. Validate the struct
4. Pass the struct to the next layer

### Good `main.go` thinking

```text
raw flags --> typed struct --> validate --> process
```

### Not good

```text
raw flags --> random if/else everywhere --> pass 8 strings --> more checks later --> confusion
```

---

## Example model fields

```go
type PipelineEvent struct {
	EventType string
	Status    string
	RepoURL   string
	Branch    string
	CommitSHA string
	Author    string
}
```

Then `main.go` can do:

```go
event := model.PipelineEvent{
	EventType: eventTypeFlag,
	Status:    statusFlag,
	RepoURL:   repoFlag,
	Branch:    branchFlag,
	CommitSHA: shaFlag,
	Author:    authorFlag,
}
```

Then:

```go
if err := event.Validate(); err != nil {
	fmt.Println("validation error:", err)
	return
}
```

This is the clean way.

---

## 8. Explain why passing a struct is better than passing many loose values

Let us compare.

## Loose values approach

```go
func SendNotification(eventType string, status string, repo string, branch string, sha string, author string) {
	// ...
}
```

Problem:

* hard to read
* hard to remember parameter order
* easier to pass wrong value in wrong place
* harder to extend later

Example bug:

```go
SendNotification("pipeline", "success", branch, repo, sha, author)
```

Oops. `branch` and `repo` got swapped.

---

## Struct approach

```go
func SendNotification(event model.PipelineEvent) {
	// ...
}
```

Benefits:

1. one clean object
2. field names are clear
3. easier to extend
4. easier to validate
5. easier to log
6. easier to test
7. better for future growth

---

## Python comparison

Python often uses:

* dict
* dataclass
* class instance

Go strongly prefers struct for this kind of modeling.

### Python dict version

```python
event = {
    "event_type": "pipeline",
    "status": "success",
    "repo": "my-repo"
}
```

Flexible, but less safe.

### Go struct version

```go
event := PipelineEvent{
	EventType: "pipeline",
	Status:    "success",
	RepoURL:   "my-repo",
}
```

More explicit and safer.

---

## 9. Pseudocode first for event creation and validation

## Pseudocode: create event from CLI

```text
START

read flags:
  eventType
  status
  repoURL
  branch
  commitSHA
  author

create PipelineEvent object using those values

validate the event:
  if eventType is empty -> error
  if status is empty -> error
  if repoURL is empty -> error
  if status is not one of allowed values -> error

if validation fails:
  print error
  stop program

if validation passes:
  send event to router
  router decides next action

END
```

---

## Pseudocode: model validation

```text
function Validate(event):
  if event.EventType is empty:
      return error "event type is required"

  if event.Status is empty:
      return error "status is required"

  if event.Status not in [started, success, failed]:
      return error "invalid status"

  if event.RepoURL is empty:
      return error "repo URL is required"

  return no error
```

---

## Why validation belongs in model layer

Because the model knows its own rules.

`PipelineEvent` should know:

* which fields are required
* which values are allowed
* what makes the event valid

This keeps validation close to the data.

### Good design

* `main.go` parses input
* `model` validates meaning
* `router` routes
* `slack` sends

### Bad design

Put validation everywhere:

* some in `main.go`
* some in router
* some in Slack client
* repeated checks everywhere

That becomes messy very fast.

---

## 10. Real code examples with full explanation

We will now build a beginner-friendly version.

---

## File: `model/pipeline_event.go`

```go
package model

import (
	"errors"
	"fmt"
	"strings"
)

type PipelineEvent struct {
	EventType string
	Status    string
	RepoURL   string
	Branch    string
	CommitSHA string
	Author    string
	Message   string
}

// Normalize cleans input before validation or routing.
func (e *PipelineEvent) Normalize() {
	e.EventType = strings.TrimSpace(strings.ToLower(e.EventType))
	e.Status = strings.TrimSpace(strings.ToLower(e.Status))
	e.RepoURL = strings.TrimSpace(e.RepoURL)
	e.Branch = strings.TrimSpace(e.Branch)
	e.CommitSHA = strings.TrimSpace(e.CommitSHA)
	e.Author = strings.TrimSpace(e.Author)
	e.Message = strings.TrimSpace(e.Message)
}

// Validate checks whether the event has the minimum required valid data.
func (e PipelineEvent) Validate() error {
	if e.EventType == "" {
		return errors.New("event type is required")
	}

	if e.Status == "" {
		return errors.New("status is required")
	}

	if e.RepoURL == "" {
		return errors.New("repo URL is required")
	}

	allowedStatus := map[string]bool{
		"started": true,
		"success": true,
		"failed":  true,
	}

	if !allowedStatus[e.Status] {
		return fmt.Errorf("invalid status: %s", e.Status)
	}

	return nil
}
```

---

## Explanation

### `type PipelineEvent struct`

This defines the model.

### `Normalize()`

This is useful because user input may contain:

* extra spaces
* uppercase/lowercase mismatch

Example:

* `" Success "` should become `"success"`

### Why pointer receiver in `Normalize()`?

Because it changes the original struct.

```go
func (e *PipelineEvent) Normalize()
```

This means the method updates the actual event.

### Why value receiver in `Validate()`?

Because validation only reads values, it does not change them.

```go
func (e PipelineEvent) Validate() error
```

---

## File: `model/notification_request.go`

```go
package model

type NotificationRequest struct {
	Channel string
	Text    string
	Event    PipelineEvent
}
```

---

## Explanation

This is another model.

Why have this?

Because sometimes your internal event model and your outgoing notification payload are not the same thing.

That is a very real backend pattern.

* `PipelineEvent` = input/business event
* `NotificationRequest` = outgoing notification object

---

## File: `main.go`

```go
package main

import (
	"flag"
	"fmt"

	"slack-integration/model"
)

func main() {
	eventType := flag.String("event", "", "event type like pipeline or pr")
	status := flag.String("status", "", "status like started, success, failed")
	repoURL := flag.String("repo", "", "repository URL")
	branch := flag.String("branch", "", "branch name")
	commitSHA := flag.String("sha", "", "commit sha")
	author := flag.String("author", "", "author name")
	message := flag.String("message", "", "custom message")

	flag.Parse()

	event := model.PipelineEvent{
		EventType: *eventType,
		Status:    *status,
		RepoURL:   *repoURL,
		Branch:    *branch,
		CommitSHA: *commitSHA,
		Author:    *author,
		Message:   *message,
	}

	event.Normalize()

	if err := event.Validate(); err != nil {
		fmt.Println("validation error:", err)
		return
	}

	fmt.Println("event created successfully")
	fmt.Printf("%+v\n", event)
}
```

---

## Explanation of `main.go`

### Step 1: define flags

```go
eventType := flag.String("event", "", "event type like pipeline or pr")
```

This returns a pointer to string.

That is why later we use `*eventType`.

### Step 2: parse flags

```go
flag.Parse()
```

Now the terminal values are loaded.

### Step 3: build a typed model

```go
event := model.PipelineEvent{ ... }
```

This is the key step of Day 2.

### Step 4: normalize

```go
event.Normalize()
```

Clean input.

### Step 5: validate

```go
if err := event.Validate(); err != nil
```

If bad data comes in, stop early.

### Step 6: continue with routing later

For now we just print the event.

---

## Example run

```bash
go run main.go \
  --event pipeline \
  --status success \
  --repo https://github.com/example/repo \
  --branch main \
  --sha abc123 \
  --author radhe \
  --message "build passed"
```

### Expected output

```go
event created successfully
{EventType:pipeline Status:success RepoURL:https://github.com/example/repo Branch:main CommitSHA:abc123 Author:radhe Message:build passed}
```

---

## Example invalid run

```bash
go run main.go \
  --event pipeline \
  --status done \
  --repo https://github.com/example/repo
```

### Output

```go
validation error: invalid status: done
```

---

## 11. File-by-file explanation of the model package

Let us explain the model package as if it is a small real project.

---

## File 1: `model/pipeline_event.go`

**Purpose:**
This is the core event model for the program.

**What it contains:**

* `PipelineEvent` struct
* fields for event-related data
* `Normalize()` method
* `Validate()` method

**Why it matters:**
This file defines the “shape” of the event and the rules for a valid event.

This is similar to:

* Python class
* or Pydantic/dataclass-style model thinking

But in Go, we usually write validation manually unless using external libraries.

---

## File 2: `model/notification_request.go`

**Purpose:**
Represents the outgoing notification payload.

**Why separate it from `PipelineEvent`?**

Because one internal model may produce multiple outputs later:

* Slack message
* email payload
* log entry
* webhook payload

So separation gives flexibility.

---

## File 3: `model/validation.go` (optional split)

In small projects, validation can stay inside `pipeline_event.go`.

In slightly larger projects, you may split validation rules into another file.

Example:

```go
package model

import (
	"errors"
	"fmt"
)

func validateRequired(value string, fieldName string) error {
	if value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

func (e PipelineEvent) Validate() error {
	if err := validateRequired(e.EventType, "event type"); err != nil {
		return err
	}
	if err := validateRequired(e.Status, "status"); err != nil {
		return err
	}
	if err := validateRequired(e.RepoURL, "repo URL"); err != nil {
		return err
	}

	allowedStatus := map[string]bool{
		"started": true,
		"success": true,
		"failed":  true,
	}

	if !allowedStatus[e.Status] {
		return errors.New("status must be one of: started, success, failed")
	}

	return nil
}
```

**Why split?**

* keeps model definition shorter
* keeps validation readable
* easier to grow later

For beginners, keeping both in one file is perfectly fine.

---

## 12. Hands-on tasks for today

Do these in order.

### Task 1

Create a simple `Student` struct with:

* `Name`
* `Course`
* `Age`

Print one student.

---

### Task 2

Add a method:

```go
func (s Student) PrintDetails()
```

Use it to print the student info.

---

### Task 3

Create a `PipelineEvent` struct with fields:

* `EventType`
* `Status`
* `RepoURL`
* `Branch`

Create one event and print it.

---

### Task 4

Add a `Validate()` method that checks:

* `EventType` not empty
* `Status` not empty
* `RepoURL` not empty

---

### Task 5

Add `Normalize()` so status becomes lowercase and trimmed.

---

### Task 6

Use CLI flags in `main.go` to fill the event model.

---

### Task 7

Print validation error if input is bad.

---

## 13. Expected output

If you run valid input:

```bash
go run main.go --event pipeline --status success --repo repo-url --branch main
```

Expected:

```go
event created successfully
{EventType:pipeline Status:success RepoURL:repo-url Branch:main CommitSHA: Author: Message:}
```

Notice the empty fields.

That is because of **zero values**.

* missing strings become `""`

This is normal in Go.

---

## 14. Common mistakes

Here are beginner mistakes you will likely hit.

### 1. Forgetting capital letters for exported names

Wrong:

```go
type PipelineEvent struct {
	eventType string
}
```

If another package needs it, this will fail.

Right:

```go
type PipelineEvent struct {
	EventType string
}
```

---

### 2. Forgetting to dereference flag pointers

Wrong:

```go
EventType: eventType,
```

Right:

```go
EventType: *eventType,
```

---

### 3. Putting too much logic in `main.go`

Bad habit:

* parsing
* validation
* routing
* Slack sending
* formatting

all in one file

Better:

* `main.go` stays small
* model validates itself
* routing happens elsewhere

---

### 4. Ignoring zero values

If user forgets `--status`, Go does not crash.
It gives `""`.

That means validation is required.

---

### 5. Confusing function and method

Function:

```go
func ValidateEvent(e PipelineEvent) error
```

Method:

```go
func (e PipelineEvent) Validate() error
```

Both are valid, but method is more natural here because validation belongs to the model.

---

### 6. Using loose values everywhere

This becomes unreadable quickly.

Prefer one struct.

---

## 15. Debugging tips

### Tip 1: print the struct

Use:

```go
fmt.Printf("%+v\n", event)
```

This shows field names too.

---

### Tip 2: print values before validation

```go
fmt.Printf("raw event: %+v\n", event)
```

This helps you see empty fields.

---

### Tip 3: test invalid input on purpose

Try:

* missing `--status`
* wrong status like `done`
* extra spaces like `" Success "`

This helps you understand normalization and validation.

---

### Tip 4: check package name and import path

If import fails, verify:

* folder name
* package name
* module path in `go.mod`

---

### Tip 5: read compiler errors slowly

Go errors are often direct and useful.

Example:

* undefined field
* cannot use `*string` as `string`
* unused variable

These usually tell you exactly what is wrong.

---

## 16. One DSA topic — string basics simply

Today’s DSA topic is **strings**.

## What is a string?

A string is a sequence of characters.

Examples:

```go
name := "radhe"
status := "success"
```

---

## Common string operations in Go

### Length

```go
fmt.Println(len("go"))
```

Output:

```go
2
```

---

### Compare strings

```go
if status == "success" {
	fmt.Println("ok")
}
```

---

### Join strings

```go
message := "build " + "passed"
```

---

### Trim spaces

```go
strings.TrimSpace(" success ")
```

---

### Convert case

```go
strings.ToLower("SUCCESS")
```

---

## Python comparison

### Python

```python
name.lower()
name.strip()
len(name)
```

### Go

```go
strings.ToLower(name)
strings.TrimSpace(name)
len(name)
```

### Main difference

Python strings have methods on the object.
Go uses package functions from `strings`.

---

## Important note for beginners

Go strings are UTF-8 text.
For advanced character handling, runes matter.

But for now, basic CLI/backend use usually starts with:

* compare
* trim
* lowercase
* contains
* split

That is enough for Day 2.

---

## 17. One easy Go DSA problem

## Problem: Count vowels in a string

Given a string, count how many vowels it contains.

### Example

Input:

```go
"golang"
```

Output:

```go
2
```

Because vowels are `o` and `a`.

---

## Go solution

```go
package main

import (
	"fmt"
	"strings"
)

func countVowels(s string) int {
	s = strings.ToLower(s)
	count := 0

	for _, ch := range s {
		if ch == 'a' || ch == 'e' || ch == 'i' || ch == 'o' || ch == 'u' {
			count++
		}
	}

	return count
}

func main() {
	fmt.Println(countVowels("Golang")) // 2
}
```

---

## Explanation

* `strings.ToLower` makes comparison easy
* `for _, ch := range s` loops through characters
* if character is vowel, increment count

---

## 18. One module-based practice task

Create a small model package with a `NotificationRequest` or `PipelineEvent` style object.

## Practice requirement

Create this folder structure:

```text
practice/
├── go.mod
├── main.go
└── model/
    └── notification_request.go
```

---

## File: `model/notification_request.go`

```go
package model

import (
	"errors"
	"strings"
)

type NotificationRequest struct {
	Channel string
	Text    string
	Source  string
}

func (n *NotificationRequest) Normalize() {
	n.Channel = strings.TrimSpace(strings.ToLower(n.Channel))
	n.Text = strings.TrimSpace(n.Text)
	n.Source = strings.TrimSpace(strings.ToLower(n.Source))
}

func (n NotificationRequest) Validate() error {
	if n.Channel == "" {
		return errors.New("channel is required")
	}
	if n.Text == "" {
		return errors.New("text is required")
	}
	if n.Source == "" {
		return errors.New("source is required")
	}
	return nil
}
```

---

## File: `main.go`

```go
package main

import (
	"fmt"

	"practice/model"
)

func main() {
	req := model.NotificationRequest{
		Channel: " Deployments ",
		Text:    " Build succeeded ",
		Source:  " CLI ",
	}

	req.Normalize()

	if err := req.Validate(); err != nil {
		fmt.Println("validation error:", err)
		return
	}

	fmt.Printf("valid request: %+v\n", req)
}
```

---

## What this teaches

1. custom model type
2. package usage
3. exported fields
4. method usage
5. normalization
6. validation
7. struct-based design

This is exactly the same pattern your real project will use.

---

## 19. Revision checkpoint

Before moving to Day 3, make sure you can answer these:

1. What is a struct in Go?
2. How is a struct different from passing loose variables?
3. What is a method?
4. What is the difference between value receiver and pointer receiver?
5. What is a package?
6. Why do real projects split code into packages?
7. What does exported vs unexported mean?
8. What is a zero value in Go?
9. Why is validation needed?
10. Why should validation belong in the model layer?
11. How does raw CLI input become a typed event object?
12. Where does the event model sit in the architecture?

If you can explain these in simple words, Day 2 is successful.

---

## 20. Homework

Do these carefully.

### Homework Part A — toy level

Create a `Book` struct with:

* `Title`
* `Author`
* `Pages`

Add a method:

```go
func (b Book) PrintInfo()
```

---

### Homework Part B — project level

Create a `PipelineEvent` model with these fields:

* `EventType`
* `Status`
* `RepoURL`
* `Branch`
* `Author`

Add:

* `Normalize()`
* `Validate()`

Validation rules:

* all fields required except `Author`
* status must be one of `started`, `success`, `failed`

---

### Homework Part C — CLI level

Update `main.go` to accept flags and build the struct from terminal input.

---

### Homework Part D — practice thinking

Answer in your own words:

1. Why is a struct better than many loose values?
2. Why should Slack client not receive random strings one by one?
3. Why is model validation safer than checking fields everywhere?

---

# Final mental model for Day 2

Think of Day 2 like this:

```text
Day 1:
CLI input enters the program

Day 2:
That raw input is converted into a clean typed model

Later:
That model will move through router, integrations, and workflows
```

Or even shorter:

```text
raw input -> struct -> validate -> route -> act
```

That is one of the most important backend patterns you can learn.

---

# Python-to-Go mental mapping for today

| Python idea                         | Go idea                             |
| ----------------------------------- | ----------------------------------- |
| class with attributes               | struct                              |
| instance method                     | method with receiver                |
| module/package                      | package                             |
| `_private` convention               | lowercase unexported                |
| public name                         | Capitalized exported                |
| dynamic dict input                  | typed struct                        |
| validation via class/Pydantic logic | validation method on struct         |
| `None`/missing cases                | zero values like `""`, `0`, `false` |

---

# Tiny summary

Today you learned:

* structs group related data
* methods attach behavior to data
* packages organize code
* capitalization controls visibility
* zero values are real and important
* raw CLI input should become a typed event model
* validation belongs near the model
* this pattern is exactly how real backend/CLI projects stay clean

Ask for **Day 3**, and I will teach you how this validated event moves into **router/service/integration layers**, with more project-style flow.
