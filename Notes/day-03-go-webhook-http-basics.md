## 1. Day 3 Learning Goals

Today you will learn how your Go Slack notification project sends messages to Slack.

By the end, you should understand:

* What JSON is
* How Go structs become JSON
* Why Go uses `json` tags
* What HTTP POST means
* What Slack incoming webhook does
* How `PipelineEvent` becomes a Slack message
* How to write a small reusable Slack client
* How this connects later with Tekton pipeline notifications
* DSA topic: maps / hash tables
* Practice: character frequency problem in Go

---

## 2. Quick Revision of Days 1 and 2

### Day 1 Revision

* `main.go` is the entry point of the Go CLI.
* CLI flags collect input from the user.
* Example:

```bash
go run main.go --event-type pr --status failed
```

* Go program starts from:

```go
func main() {
}
```

* `go.mod` manages the Go module.
* Packages help organize real projects.
* CLI input is usually raw string data.

### Day 2 Revision

* Structs are like Python classes mainly used to hold data.
* Example:

```go
type PipelineEvent struct {
	EventType string
	Status    string
}
```

Python equivalent:

```python
class PipelineEvent:
    def __init__(self, event_type, status):
        self.event_type = event_type
        self.status = status
```

* Methods are functions attached to structs.
* Validation should belong near the model.

Example:

```go
func (e PipelineEvent) Validate() error {
	if e.EventType == "" {
		return errors.New("event type is required")
	}
	return nil
}
```

---

# 3. Beginner-Friendly Explanation of JSON

## What is JSON?

JSON means **JavaScript Object Notation**.

It is a common text format used to send data between systems.

Example JSON:

```json
{
  "event_type": "pr",
  "status": "failed",
  "repository": "cloud-resource-onboarding"
}
```

In Python, this is similar to a dictionary:

```python
event = {
    "event_type": "pr",
    "status": "failed",
    "repository": "cloud-resource-onboarding"
}
```

In Go, we usually represent this data using a struct:

```go
type PipelineEvent struct {
	EventType  string
	Status     string
	Repository string
}
```

---

## Why JSON is important in your Slack project

Your Go program cannot directly send a Go struct to Slack.

Slack expects HTTP request body in JSON format.

So the flow is:

```text
Go struct
   |
   v
JSON payload
   |
   v
HTTP POST
   |
   v
Slack webhook
```

---

# 4. Explain `json` Tags in Detail

## Toy Example First

Go struct:

```go
type User struct {
	Name  string
	Email string
}
```

If we convert this to JSON, Go uses field names:

```json
{
  "Name": "Radhe",
  "Email": "radhe@example.com"
}
```

But APIs usually prefer snake_case or lowercase names:

```json
{
  "name": "Radhe",
  "email": "radhe@example.com"
}
```

So we add JSON tags:

```go
type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
```

Now Go produces:

```json
{
  "name": "Radhe",
  "email": "radhe@example.com"
}
```

---

## Python Comparison

Python:

```python
user = {
    "name": "Radhe",
    "email": "radhe@example.com"
}
```

Python dictionaries already use the exact key names.

Go structs need tags because field names are Go identifiers.

---

## Important Go Rule

In Go, only exported fields can be converted to JSON.

Exported means field name starts with a capital letter.

Correct:

```go
type User struct {
	Name string `json:"name"`
}
```

Wrong:

```go
type User struct {
	name string `json:"name"`
}
```

The lowercase `name` is unexported, so JSON encoding will ignore it.

---

## Common JSON Tag Options

```go
type User struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}
```

`omitempty` means:

> If the field is empty, do not include it in JSON.

Example:

```go
type SlackMessage struct {
	Text string `json:"text"`
}
```

Slack expects this:

```json
{
  "text": "Build failed"
}
```

So the Go struct should be:

```go
type SlackMessage struct {
	Text string `json:"text"`
}
```

---

# 5. HTTP Request/Response Basics

HTTP is how systems talk over the web.

Example:

```text
Client sends request
Server sends response
```

ASCII flow:

```text
+-------------+        HTTP Request         +-------------+
| Go Program  | --------------------------> | Slack API   |
| HTTP Client |                             | Webhook URL |
+-------------+ <-------------------------- +-------------+
                  HTTP Response
```

---

## Common HTTP Methods

| Method | Meaning          |
| ------ | ---------------- |
| GET    | Read data        |
| POST   | Send/create data |
| PUT    | Replace data     |
| PATCH  | Update partially |
| DELETE | Delete data      |

For Slack webhook, we use:

```text
POST
```

Because we are sending a message to Slack.

---

## HTTP Request Contains

```text
URL
Method
Headers
Body
```

Example:

```text
POST https://hooks.slack.com/services/xxx
Content-Type: application/json

{
  "text": "Pipeline failed"
}
```

---

## HTTP Response Contains

```text
Status code
Headers
Body
```

Common response codes:

| Code | Meaning           |
| ---- | ----------------- |
| 200  | Success           |
| 400  | Bad request       |
| 401  | Unauthorized      |
| 403  | Forbidden         |
| 404  | Not found         |
| 429  | Too many requests |
| 500  | Server error      |

For Slack webhook, success usually means HTTP `200`.

---

# 6. What is a Slack Webhook?

A Slack incoming webhook is a special URL.

When you send JSON to that URL, Slack posts a message into a Slack channel.

Simple analogy:

```text
Webhook URL = Doorbell of Slack channel
Go program = Person pressing the doorbell
JSON message = Message note
Slack channel = Room where message appears
```

Example payload:

```json
{
  "text": "PR pipeline failed for branch feature/login"
}
```

---

## Why You Should Not Hardcode Webhook URLs

Bad:

```go
webhookURL := "https://hooks.slack.com/services/secret"
```

Why bad?

* It exposes secrets in code.
* It may get committed to GitHub.
* Different environments need different URLs.
* Security teams may flag it.
* Rotation becomes hard.

Better:

```go
webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
```

Run like this:

```bash
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/xxx"
go run main.go
```

---

# 7. Flow: Event → Slack Payload → Webhook → Response

Project flow:

```text
CLI flags
   |
   v
PipelineEvent struct
   |
   v
Validate event
   |
   v
Convert event to SlackMessage
   |
   v
Marshal SlackMessage to JSON
   |
   v
Send HTTP POST to Slack webhook
   |
   v
Check Slack response code
```

ASCII diagram:

```text
+-------------+
| CLI Input   |
+-------------+
       |
       v
+----------------+
| PipelineEvent  |
+----------------+
       |
       v
+----------------+
| SlackMessage   |
+----------------+
       |
       v
+----------------+
| JSON Payload   |
+----------------+
       |
       v
+----------------+
| HTTP POST      |
+----------------+
       |
       v
+----------------+
| Slack Channel  |
+----------------+
```

---

# 8. Pseudocode First

```text
START

Read Slack webhook URL from environment variable

Create pipeline event:
    event type
    status
    repository
    branch
    commit id

Validate event

Convert event into Slack message text

Create Slack payload:
    text = formatted message

Convert Slack payload into JSON

Create HTTP POST request:
    URL = webhook URL
    Body = JSON payload
    Header = Content-Type: application/json

Send request using http.Client

Check response:
    if status code is 200:
        print success
    else:
        return error

END
```

---

# 9. Real Go Code Example: Small Slack Client

## Suggested file

```text
pkg/notify/slack/client.go
```

## Code

```go
package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Message struct {
	Text string `json:"text"`
}

type Client struct {
	WebhookURL string
	HTTPClient *http.Client
}

func NewClient(webhookURL string) (*Client, error) {
	if webhookURL == "" {
		return nil, errors.New("slack webhook URL is required")
	}

	return &Client{
		WebhookURL: webhookURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (c *Client) SendMessage(message Message) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to convert slack message to JSON: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		c.WebhookURL,
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Slack: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack returned non-success status code: %d", resp.StatusCode)
	}

	return nil
}
```

---

# 10. Explain Every Important Line

```go
package slack
```

This file belongs to the `slack` package.

Python comparison:

```python
# slack/client.py
```

---

```go
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)
```

Go imports required packages.

Python equivalent:

```python
import json
import requests
import time
```

---

```go
type Message struct {
	Text string `json:"text"`
}
```

This represents Slack message payload.

It becomes:

```json
{
  "text": "some message"
}
```

Important:

* `Text` is capitalized so Go can export it to JSON.
* `json:"text"` tells Go to use `text` in JSON.

---

```go
type Client struct {
	WebhookURL string
	HTTPClient *http.Client
}
```

This is a reusable Slack client.

It stores:

* Slack webhook URL
* HTTP client

Python comparison:

```python
class Client:
    def __init__(self, webhook_url):
        self.webhook_url = webhook_url
```

---

```go
func NewClient(webhookURL string) (*Client, error)
```

This is a constructor-like function.

Python equivalent:

```python
client = Client(webhook_url)
```

Go convention:

```go
NewSomething()
```

is commonly used to create initialized objects.

---

```go
if webhookURL == "" {
	return nil, errors.New("slack webhook URL is required")
}
```

This validates that webhook URL is not empty.

In Python:

```python
if not webhook_url:
    raise ValueError("slack webhook URL is required")
```

---

```go
HTTPClient: &http.Client{
	Timeout: 10 * time.Second,
}
```

This creates an HTTP client with timeout.

Why timeout is important?

Without timeout, your program may hang forever if Slack does not respond.

Production rule:

```text
Every network call should have a timeout.
```

---

```go
payload, err := json.Marshal(message)
```

This converts Go struct to JSON.

Python equivalent:

```python
payload = json.dumps(message)
```

Example:

```go
Message{Text: "Build failed"}
```

becomes:

```json
{"text":"Build failed"}
```

---

```go
req, err := http.NewRequest(...)
```

This creates an HTTP request.

It does not send yet.

It only prepares the request.

---

```go
http.MethodPost
```

This means HTTP method is POST.

Same as:

```go
"POST"
```

But using constant is cleaner.

---

```go
bytes.NewBuffer(payload)
```

This converts JSON bytes into request body.

---

```go
req.Header.Set("Content-Type", "application/json")
```

This tells Slack:

```text
I am sending JSON data.
```

Without this, Slack may reject or misunderstand the payload.

---

```go
resp, err := c.HTTPClient.Do(req)
```

This sends the request.

Python equivalent:

```python
response = requests.post(url, json=payload)
```

---

```go
defer resp.Body.Close()
```

This closes the response body after function finishes.

Important Go convention:

```text
If you open/read an HTTP response, close the body.
```

---

```go
if resp.StatusCode < 200 || resp.StatusCode >= 300
```

This checks whether response is successful.

Success range:

```text
200 to 299
```

---

# 11. Slack Package Organization

Recommended structure:

```text
slack-integration/
├── go.mod
├── cmd/
│   └── slack-notifier/
│       └── main.go
├── pkg/
│   └── notify/
│       ├── model/
│       │   └── event.go
│       ├── formatter/
│       │   └── formatter.go
│       └── slack/
│           └── client.go
```

Meaning:

```text
model      -> event data and validation
formatter  -> converts event into Slack text
slack      -> sends message to Slack
main.go    -> wires everything together
```

---

## Example Model

```go
package model

type PipelineEvent struct {
	EventType  string
	Status     string
	Repository string
	Branch     string
	CommitID   string
	Sender     string
}
```

---

## Example Formatter

```go
package formatter

import (
	"fmt"

	"slack-integration/pkg/notify/model"
)

func FormatSlackText(event model.PipelineEvent) string {
	return fmt.Sprintf(
		"Pipeline event: %s\nStatus: %s\nRepo: %s\nBranch: %s\nCommit: %s\nSender: %s",
		event.EventType,
		event.Status,
		event.Repository,
		event.Branch,
		event.CommitID,
		event.Sender,
	)
}
```

---

## Example Main Flow

```go
package main

import (
	"fmt"
	"os"

	"slack-integration/pkg/notify/formatter"
	"slack-integration/pkg/notify/model"
	"slack-integration/pkg/notify/slack"
)

func main() {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")

	event := model.PipelineEvent{
		EventType:  "pr",
		Status:     "failed",
		Repository: "cloud-resource-onboarding",
		Branch:     "feature/slack-alert",
		CommitID:   "abc123",
		Sender:     "radheshyam",
	}

	text := formatter.FormatSlackText(event)

	client, err := slack.NewClient(webhookURL)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	err = client.SendMessage(slack.Message{
		Text: text,
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Slack notification sent successfully")
}
```

---

# 12. Hands-on Tasks

## Task 1: Create Slack Message Struct

Create:

```go
type Message struct {
	Text string `json:"text"`
}
```

Then marshal it:

```go
msg := Message{Text: "Hello from Go"}
payload, _ := json.Marshal(msg)
fmt.Println(string(payload))
```

Expected output:

```json
{"text":"Hello from Go"}
```

---

## Task 2: Read Webhook URL from Environment

```go
webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
```

Run:

```bash
export SLACK_WEBHOOK_URL="your-webhook-url"
go run cmd/slack-notifier/main.go
```

---

## Task 3: Send Test Slack Message

Create message:

```go
slack.Message{
	Text: "Test notification from Go Slack client",
}
```

Send using:

```go
client.SendMessage(msg)
```

---

# 13. Expected Output

Terminal:

```text
Slack notification sent successfully
```

Slack channel:

```text
Pipeline event: pr
Status: failed
Repo: cloud-resource-onboarding
Branch: feature/slack-alert
Commit: abc123
Sender: radheshyam
```

---

# 14. Common Mistakes

## Mistake 1: Lowercase struct fields

Wrong:

```go
type Message struct {
	text string `json:"text"`
}
```

Correct:

```go
type Message struct {
	Text string `json:"text"`
}
```

---

## Mistake 2: Missing JSON header

Wrong:

```go
// no header
```

Correct:

```go
req.Header.Set("Content-Type", "application/json")
```

---

## Mistake 3: Hardcoding webhook URL

Wrong:

```go
webhookURL := "https://hooks.slack.com/services/xxx"
```

Correct:

```go
webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
```

---

## Mistake 4: No timeout

Wrong:

```go
client := &http.Client{}
```

Better:

```go
client := &http.Client{
	Timeout: 10 * time.Second,
}
```

---

## Mistake 5: Ignoring response code

Wrong:

```go
return nil
```

Correct:

```go
if resp.StatusCode < 200 || resp.StatusCode >= 300 {
	return fmt.Errorf("slack returned status code: %d", resp.StatusCode)
}
```

---

# 15. Debugging Tips

## Print JSON Payload

```go
fmt.Println(string(payload))
```

Check whether it looks like:

```json
{"text":"your message"}
```

---

## Check Webhook URL

```go
if webhookURL == "" {
	fmt.Println("SLACK_WEBHOOK_URL is missing")
}
```

---

## Print Response Code

```go
fmt.Println("Slack status:", resp.StatusCode)
```

---

## Test Webhook with curl

```bash
curl -X POST \
  -H 'Content-Type: application/json' \
  --data '{"text":"Hello from curl"}' \
  "$SLACK_WEBHOOK_URL"
```

If curl works but Go fails, problem is likely in Go request code.

---

# 16. DSA Topic: Maps / Hash Tables

## Simple Meaning

A map stores data as key-value pairs.

Python:

```python
user = {
    "name": "Radhe",
    "role": "developer"
}
```

Go:

```go
user := map[string]string{
	"name": "Radhe",
	"role": "developer",
}
```

---

## Why Maps Are Useful

Fast lookup.

Example:

```go
statusEmoji := map[string]string{
	"success": "✅",
	"failed":  "❌",
	"running": "⏳",
}
```

Use:

```go
fmt.Println(statusEmoji["failed"])
```

Output:

```text
❌
```

---

## Python vs Go Map Difference

Python:

```python
value = my_dict.get("key", "default")
```

Go:

```go
value, ok := myMap["key"]
if !ok {
	value = "default"
}
```

---

# 17. Go DSA Problem: Character Frequency

## Problem

Given a string, count how many times each character appears.

Input:

```text
banana
```

Output:

```text
b -> 1
a -> 3
n -> 2
```

---

## Go Solution

```go
package main

import "fmt"

func main() {
	text := "banana"

	frequency := make(map[rune]int)

	for _, ch := range text {
		frequency[ch]++
	}

	for ch, count := range frequency {
		fmt.Printf("%c -> %d\n", ch, count)
	}
}
```

---

## Explanation

```go
frequency := make(map[rune]int)
```

Creates a map.

Key type:

```go
rune
```

Value type:

```go
int
```

Why `rune`?

Because Go strings are UTF-8, and `rune` safely represents characters.

---

Python equivalent:

```python
text = "banana"
frequency = {}

for ch in text:
    frequency[ch] = frequency.get(ch, 0) + 1

print(frequency)
```

---

# 18. Module-Based Practice Task

## Task: Build Notification Formatter

Create this file:

```text
pkg/notify/formatter/formatter.go
```

Code:

```go
package formatter

import (
	"fmt"

	"slack-integration/pkg/notify/model"
)

func FormatSlackText(event model.PipelineEvent) string {
	statusEmoji := map[string]string{
		"succeeded": "✅",
		"failed":    "❌",
		"running":   "⏳",
	}

	emoji, ok := statusEmoji[event.Status]
	if !ok {
		emoji = "ℹ️"
	}

	return fmt.Sprintf(
		"%s Pipeline Notification\nEvent: %s\nStatus: %s\nRepository: %s\nBranch: %s\nCommit: %s\nSender: %s",
		emoji,
		event.EventType,
		event.Status,
		event.Repository,
		event.Branch,
		event.CommitID,
		event.Sender,
	)
}
```

---

## Why This Is Useful

This keeps formatting separate from Slack sending.

Good design:

```text
formatter package -> prepares message
slack package     -> sends message
model package     -> stores event data
```

Bad design:

```text
main.go does everything
```

---

# 19. Revision Checkpoint

You should now be able to answer:

1. What is JSON?
2. Why does Go need `json` tags?
3. Why should struct fields be capitalized for JSON?
4. What does `json.Marshal()` do?
5. Why do we use HTTP POST for Slack?
6. What is `http.Client`?
7. Why is timeout important?
8. Why should webhook URL come from environment variable?
9. What does `Content-Type: application/json` mean?
10. How does `PipelineEvent` become a Slack message?
11. What is a map in Go?
12. How is Go map similar to Python dictionary?

---

# 20. Homework

## Homework 1

Create this struct:

```go
type SlackMessage struct {
	Text string `json:"text"`
}
```

Convert it to JSON and print it.

---

## Homework 2

Create a formatter function:

```go
func FormatSlackText(event PipelineEvent) string
```

It should return:

```text
PR pipeline failed for repo cloud-resource-onboarding on branch feature/test
```

---

## Homework 3

Create a status emoji map:

```go
map[string]string{
	"succeeded": "✅",
	"failed": "❌",
	"running": "⏳",
}
```

Use it inside your formatter.

---

## Homework 4

Solve character frequency for:

```text
slacknotification
```

---

## Homework 5

Explain in your own words:

```text
PipelineEvent -> Formatter -> SlackMessage -> JSON -> HTTP POST -> Slack
```

This is the most important Day 3 flow.
---
## Core idea

`GET` is used to **read/fetch data**.
`POST` is used to **send/create data**.

---

## 1. GET curl

### Simple GET

```bash
curl "https://example.com/api/users"
```

Meaning:

```text
Method: GET
URL: https://example.com/api/users
Body: no body
```

### GET with header

```bash
curl -X GET "https://example.com/api/users" \
  -H "Accept: application/json"
```

### GET with query params

```bash
curl "https://example.com/api/users?status=active&page=1"
```

Query params go in the URL.

---

## 2. POST curl

### Simple POST with JSON body

```bash
curl -X POST "https://example.com/api/users" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Radhe",
    "role": "developer"
  }'
```

Meaning:

```text
Method: POST
URL: https://example.com/api/users
Header: Content-Type: application/json
Body: JSON data
```

---

## 3. Difference in one table

| Part          | GET              | POST             |
| ------------- | ---------------- | ---------------- |
| Purpose       | Read data        | Send/create data |
| Body          | Usually no body  | Usually has body |
| Data location | URL/query params | Request body     |
| Example use   | Get user list    | Create user      |
| Slack webhook | Not used         | Used             |

---

## 4. Slack webhook POST curl

```bash
curl -X POST "$SLACK_WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Hello from curl"
  }'
```

Slack webhook needs `POST` because you are **sending a message**.

---

## 5. Easy memory trick

```text
GET  = give me data
POST = take this data
```

Example:

```text
GET  /users        -> give me users
POST /users        -> create this new user
```
