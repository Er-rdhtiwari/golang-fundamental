# Day 4 — Go CLI + Slack Notifier: Router Logic and Clean Project Architecture

## 1. Day 4 learning goals

Today you should understand:

1. Why routing logic is a separate layer.
2. Why `main.go` should stay small.
3. How `model`, `router`, and `slack client` packages work together.
4. How the project chooses the correct Slack webhook.
5. How fallback works when a webhook is missing.
6. How to organize a Go backend/CLI project cleanly.
7. How Go package structure compares with Python modules.
8. One DSA topic: **stack vs queue**.
9. One Go queue implementation problem.
10. One module-based practice task: **small task router**.

---

## 2. Quick revision of Days 1 to 3

### Day 1: CLI basics

You learned that the project starts from CLI input.

Example:

```bash
go run cmd/slack-notifier/main.go \
  --event-type pr \
  --status failed \
  --repository cloud-resource-onboarding
```

The CLI receives input from the user or pipeline.

Python comparison:

```python
# Python
import argparse
```

Go comparison:

```go
// Go
import "flag"
```

Both are used to read command-line input.

---

### Day 2: Structs, methods, packages

You learned that raw CLI input should become a typed Go struct.

Example:

```go
type PipelineEvent struct {
	EventType  string
	Status     string
	Repository string
}
```

Python equivalent:

```python
from dataclasses import dataclass

@dataclass
class PipelineEvent:
    event_type: str
    status: str
    repository: str
```

Main idea:

> Do not pass many loose variables everywhere. Create one clear object/struct.

---

### Day 3: JSON, HTTP, Slack webhook

You learned that Slack expects JSON over HTTP POST.

Go struct:

```go
type SlackMessage struct {
	Text string `json:"text"`
}
```

Python equivalent:

```python
payload = {
    "text": "Build failed"
}
```

Go sends HTTP POST using:

```go
http.Post(...)
```

Python sends HTTP POST using:

```python
requests.post(...)
```

---

## 3. Explain separation of concerns very simply

Separation of concerns means:

> Every file/package should have one clear job.

Think of a restaurant:

| Person       | Responsibility          |
| ------------ | ----------------------- |
| Receptionist | Receives customer       |
| Waiter       | Routes order to kitchen |
| Chef         | Prepares food           |
| Cashier      | Handles payment         |

You do not want the chef also managing the door, billing, and delivery.

In your Slack project:

| Layer     | Responsibility                       |
| --------- | ------------------------------------ |
| `main.go` | Read CLI input and connect packages  |
| `model`   | Define event data and validation     |
| `router`  | Decide which Slack webhook to use    |
| `slack`   | Send HTTP request to Slack           |
| `config`  | Store webhook URLs and routing rules |

### Messy all-in-one style

Bad style:

```go
func main() {
	// read flags
	// validate event
	// decide webhook
	// create Slack JSON
	// send HTTP request
	// retry
	// log output
}
```

Problem:

* Hard to read.
* Hard to test.
* Hard to debug.
* Hard to reuse.
* One small change can break everything.

### Clean modular style

Good style:

```go
func main() {
	event := buildEventFromFlags()
	event.Validate()

	route := router.Resolve(event)
	message := slack.BuildMessage(event)

	client.Send(route.WebhookURL, message)
}
```

This is much easier to understand.

---

## 4. Explain why `main.go` should stay small

`main.go` is the entry point.

Its job should be:

1. Read CLI flags.
2. Create `PipelineEvent`.
3. Validate event.
4. Ask router for webhook.
5. Ask Slack client to send message.
6. Print success or error.

`main.go` should **not** contain:

* Complex routing rules.
* Slack HTTP request details.
* Retry logic.
* JSON payload building details.
* Business validation logic.

### Python comparison

In Python, you may have:

```python
if __name__ == "__main__":
    main()
```

Usually this file should also stay small.

Bad Python style:

```python
# main.py
# 500 lines of flags, validation, routing, requests.post, retry, logging
```

Good Python style:

```python
# main.py
from router import resolve_route
from slack_client import send_message
from model import PipelineEvent
```

Same principle in Go.

---

## 5. Difference between model, router, and Slack client layers

### A. Model layer

The model layer defines the data.

Example:

```go
type PipelineEvent struct {
	EventType  string
	Status     string
	Repository string
	Branch     string
	Sender     string
}
```

The model answers:

> What is a valid pipeline event?

It can validate required fields.

```go
func (e PipelineEvent) Validate() error {
	if e.EventType == "" {
		return fmt.Errorf("event type is required")
	}
	if e.Status == "" {
		return fmt.Errorf("status is required")
	}
	return nil
}
```

Python equivalent:

```python
@dataclass
class PipelineEvent:
    event_type: str
    status: str

    def validate(self):
        if not self.event_type:
            raise ValueError("event type is required")
```

Key difference:

* Python commonly uses exceptions.
* Go commonly returns `error`.

---

### B. Router layer

The router decides:

> Which webhook should receive this event?

Example:

```go
pr event  -> PR Slack webhook
cd event  -> CD Slack webhook
job event -> Job Slack webhook
```

The router should **not** send HTTP requests.

It only decides the destination.

---

### C. Slack client layer

The Slack client handles:

> How to send the message to Slack?

It knows about:

* HTTP POST.
* JSON encoding.
* Slack response status code.
* Retry logic.
* Timeout.

It should **not** decide whether the event is PR/CD/Job.

That belongs to the router.

---

## 6. Router logic using simple examples first

Imagine you have three support teams:

| Request type | Team           |
| ------------ | -------------- |
| billing      | Billing team   |
| technical    | Technical team |
| refund       | Refund team    |

Simple router logic:

```go
if requestType == "billing" {
	sendTo = "billing-team"
} else if requestType == "technical" {
	sendTo = "technical-team"
} else {
	sendTo = "general-support"
}
```

Python equivalent:

```python
if request_type == "billing":
    send_to = "billing-team"
elif request_type == "technical":
    send_to = "technical-team"
else:
    send_to = "general-support"
```

In Go, we often use `switch`:

```go
switch requestType {
case "billing":
	sendTo = "billing-team"
case "technical":
	sendTo = "technical-team"
default:
	sendTo = "general-support"
}
```

Go `switch` is similar to Python `match`:

```python
match request_type:
    case "billing":
        send_to = "billing-team"
    case "technical":
        send_to = "technical-team"
    case _:
        send_to = "general-support"
```

---

## 7. Project-based routing rules

For your Slack notifier project:

| Event type | Route key            | Webhook           |
| ---------- | -------------------- | ----------------- |
| `pr`       | `pr`                 | PR webhook        |
| `cd`       | `cd`                 | CD webhook        |
| `job`      | `job`                | Job webhook       |
| unknown    | maybe fallback/error | depends on config |

Example:

```text
Pull request failed  -> send to PR Slack channel
CD deployment failed -> send to CD Slack channel
Job sync failed      -> send to Job Slack channel
```

But sometimes you may not have a separate Job webhook.

So fallback rule can be:

```text
If job webhook is missing, send job notification to CD webhook.
```

This is practical because job events are often related to delivery/deployment flow.

---

## 8. Fallback behavior for missing webhook configuration

Fallback means:

> Use another safe option when the preferred option is missing.

Example:

```text
Event type: job
Preferred webhook: JOB_WEBHOOK
But JOB_WEBHOOK is empty
Fallback webhook: CD_WEBHOOK
```

Expected behavior:

```text
Route selected: cd
Fallback used: true
Reason: job webhook missing, using cd webhook
```

Bad behavior:

```text
panic: empty webhook URL
```

Good backend systems avoid sudden crashes when a safe fallback exists.

### Python equivalent

Python:

```python
job_webhook = os.getenv("JOB_WEBHOOK") or os.getenv("CD_WEBHOOK")
```

Go:

```go
if jobWebhook == "" {
	jobWebhook = cdWebhook
}
```

---

## 9. Pseudocode first for router logic

```text
FUNCTION ResolveRoute(event, config):

    IF event type is "pr":
        IF pr webhook exists:
            RETURN pr route
        ELSE:
            RETURN error "PR webhook missing"

    ELSE IF event type is "cd":
        IF cd webhook exists:
            RETURN cd route
        ELSE:
            RETURN error "CD webhook missing"

    ELSE IF event type is "job":
        IF job webhook exists:
            RETURN job route
        ELSE IF cd webhook exists:
            RETURN cd route with fallback=true
        ELSE:
            RETURN error "Job and CD webhook both missing"

    ELSE:
        IF default webhook exists:
            RETURN default route
        ELSE:
            RETURN error "Unsupported event type"
```

---

## 10. Real Go code examples

### Suggested project structure

```text
slack-integration/
│
├── cmd/
│   └── slack-notifier/
│       └── main.go
│
├── pkg/
│   └── notify/
│       ├── model/
│       │   └── event.go
│       │
│       ├── router/
│       │   └── router.go
│       │
│       └── slack/
│           └── client.go
│
└── go.mod
```

---

# File 1: `pkg/notify/model/event.go`

```go
package model

import "fmt"

const (
	EventTypePR  = "pr"
	EventTypeCD  = "cd"
	EventTypeJob = "job"
)

type PipelineEvent struct {
	EventType  string
	Status     string
	Repository string
	Branch     string
	Sender     string
}

func (e PipelineEvent) Validate() error {
	if e.EventType == "" {
		return fmt.Errorf("event type is required")
	}

	if e.Status == "" {
		return fmt.Errorf("status is required")
	}

	if e.Repository == "" {
		return fmt.Errorf("repository is required")
	}

	return nil
}
```

### Explanation

```go
package model
```

This file belongs to the `model` package.

Python equivalent:

```python
# model/event.py
```

---

```go
const (
	EventTypePR  = "pr"
	EventTypeCD  = "cd"
	EventTypeJob = "job"
)
```

These are constants.

Python equivalent:

```python
EVENT_TYPE_PR = "pr"
EVENT_TYPE_CD = "cd"
EVENT_TYPE_JOB = "job"
```

Go convention:

* Constants often use `CamelCase`.
* Exported names start with capital letters.
* `EventTypePR` is visible outside the package.
* `eventTypePR` would be private to the package.

---

```go
type PipelineEvent struct {
	EventType  string
	Status     string
	Repository string
	Branch     string
	Sender     string
}
```

This defines the event object.

Python equivalent:

```python
@dataclass
class PipelineEvent:
    event_type: str
    status: str
    repository: str
    branch: str
    sender: str
```

---

```go
func (e PipelineEvent) Validate() error
```

This is a method on `PipelineEvent`.

Python equivalent:

```python
def validate(self):
```

Go difference:

```go
(e PipelineEvent)
```

means this method belongs to the `PipelineEvent` struct.

---

# File 2: `pkg/notify/router/router.go`

```go
package router

import (
	"fmt"

	"slack-integration/pkg/notify/model"
)

type Config struct {
	PRWebhook      string
	CDWebhook      string
	JobWebhook     string
	DefaultWebhook string
}

type RouteResult struct {
	RouteName    string
	WebhookURL   string
	UsedFallback bool
	Reason       string
}

type Router struct {
	config Config
}

func NewRouter(config Config) Router {
	return Router{
		config: config,
	}
}

func (r Router) Resolve(event model.PipelineEvent) (RouteResult, error) {
	switch event.EventType {

	case model.EventTypePR:
		if r.config.PRWebhook == "" {
			return RouteResult{}, fmt.Errorf("PR webhook is missing")
		}

		return RouteResult{
			RouteName:  "pr",
			WebhookURL: r.config.PRWebhook,
			Reason:    "PR event routed to PR webhook",
		}, nil

	case model.EventTypeCD:
		if r.config.CDWebhook == "" {
			return RouteResult{}, fmt.Errorf("CD webhook is missing")
		}

		return RouteResult{
			RouteName:  "cd",
			WebhookURL: r.config.CDWebhook,
			Reason:    "CD event routed to CD webhook",
		}, nil

	case model.EventTypeJob:
		if r.config.JobWebhook != "" {
			return RouteResult{
				RouteName:  "job",
				WebhookURL: r.config.JobWebhook,
				Reason:    "Job event routed to Job webhook",
			}, nil
		}

		if r.config.CDWebhook != "" {
			return RouteResult{
				RouteName:    "cd",
				WebhookURL:   r.config.CDWebhook,
				UsedFallback: true,
				Reason:       "Job webhook missing, fallback to CD webhook",
			}, nil
		}

		return RouteResult{}, fmt.Errorf("job webhook missing and CD fallback webhook also missing")

	default:
		if r.config.DefaultWebhook != "" {
			return RouteResult{
				RouteName:    "default",
				WebhookURL:   r.config.DefaultWebhook,
				UsedFallback: true,
				Reason:       "Unknown event routed to default webhook",
			}, nil
		}

		return RouteResult{}, fmt.Errorf("unsupported event type: %s", event.EventType)
	}
}
```

---

## Line-by-line explanation

```go
package router
```

This package is responsible only for routing.

It should not send Slack messages.

---

```go
import (
	"fmt"

	"slack-integration/pkg/notify/model"
)
```

The router imports the model package because it needs to read `PipelineEvent`.

Python equivalent:

```python
from notify.model.event import PipelineEvent
```

---

```go
type Config struct {
	PRWebhook      string
	CDWebhook      string
	JobWebhook     string
	DefaultWebhook string
}
```

This stores webhook configuration.

Python equivalent:

```python
@dataclass
class Config:
    pr_webhook: str
    cd_webhook: str
    job_webhook: str
    default_webhook: str
```

---

```go
type RouteResult struct {
	RouteName    string
	WebhookURL   string
	UsedFallback bool
	Reason       string
}
```

Instead of returning only a string, we return a clear result object.

This helps debugging.

Example result:

```text
RouteName: cd
UsedFallback: true
Reason: Job webhook missing, fallback to CD webhook
```

---

```go
type Router struct {
	config Config
}
```

The router owns routing configuration.

Python equivalent:

```python
class Router:
    def __init__(self, config):
        self.config = config
```

---

```go
func NewRouter(config Config) Router
```

This is a constructor-style function.

Python equivalent:

```python
router = Router(config)
```

Go does not require constructors, but `NewRouter()` is a common Go convention.

---

```go
func (r Router) Resolve(event model.PipelineEvent) (RouteResult, error)
```

This method receives an event and returns:

1. `RouteResult`
2. `error`

Python equivalent:

```python
def resolve(self, event: PipelineEvent) -> RouteResult:
```

But in Python, errors are commonly raised:

```python
raise ValueError("missing webhook")
```

In Go, errors are returned:

```go
return RouteResult{}, fmt.Errorf("missing webhook")
```

---

## 11. ASCII diagram for package relationships

```text
                 CLI input / Tekton params / GitHub event data
                                  |
                                  v
                          cmd/slack-notifier
                              main.go
                                  |
                                  v
                  +-------------------------------+
                  |  model.PipelineEvent           |
                  |  - event type                  |
                  |  - status                      |
                  |  - repository                  |
                  |  - branch                      |
                  |  - sender                      |
                  +-------------------------------+
                                  |
                                  v
                  +-------------------------------+
                  |  router.Router                 |
                  |  - chooses PR/CD/Job webhook   |
                  |  - applies fallback logic      |
                  |  - returns RouteResult         |
                  +-------------------------------+
                                  |
                                  v
                  +-------------------------------+
                  |  slack.Client                  |
                  |  - builds JSON payload         |
                  |  - sends HTTP POST             |
                  |  - handles Slack response      |
                  +-------------------------------+
                                  |
                                  v
                            Slack Channel
```

---

## Why routing should not be mixed with Slack HTTP sending

Bad design:

```go
func SendSlackMessage(event PipelineEvent) error {
	if event.EventType == "pr" {
		webhook := os.Getenv("PR_WEBHOOK")
		http.Post(webhook, ...)
	}

	if event.EventType == "cd" {
		webhook := os.Getenv("CD_WEBHOOK")
		http.Post(webhook, ...)
	}
}
```

Problems:

* Slack client now has routing logic.
* Hard to test routing without making HTTP calls.
* Hard to reuse Slack client.
* Hard to add fallback rules cleanly.

Good design:

```go
route, err := router.Resolve(event)
err = slackClient.Send(route.WebhookURL, message)
```

Now:

* Router decides destination.
* Slack client sends message.
* Both can be tested separately.

---

# File 3: `pkg/notify/slack/client.go`

```go
package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Message struct {
	Text string `json:"text"`
}

type Client struct {
	httpClient *http.Client
}

func NewClient() Client {
	return Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c Client) Send(webhookURL string, message Message) error {
	if webhookURL == "" {
		return fmt.Errorf("webhook URL is empty")
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to encode Slack message: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create Slack request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Slack returned non-success status: %d", resp.StatusCode)
	}

	return nil
}
```

### Important point

This file does not know about:

```go
pr
cd
job
```

That is good.

The Slack client only knows:

```text
Given a webhook URL and message, send it to Slack.
```

Python equivalent:

```python
import requests

def send(webhook_url, message):
    response = requests.post(webhook_url, json=message, timeout=10)
    response.raise_for_status()
```

---

# File 4: `cmd/slack-notifier/main.go`

```go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"slack-integration/pkg/notify/model"
	"slack-integration/pkg/notify/router"
	"slack-integration/pkg/notify/slack"
)

func main() {
	eventType := flag.String("event-type", "", "event type: pr, cd, job")
	status := flag.String("status", "", "event status: running, succeeded, failed")
	repository := flag.String("repository", "", "repository name")
	branch := flag.String("branch", "", "branch name")
	sender := flag.String("sender", "", "sender name")

	flag.Parse()

	event := model.PipelineEvent{
		EventType:  *eventType,
		Status:     *status,
		Repository: *repository,
		Branch:     *branch,
		Sender:     *sender,
	}

	if err := event.Validate(); err != nil {
		log.Fatalf("invalid event: %v", err)
	}

	routeConfig := router.Config{
		PRWebhook:      os.Getenv("SLACK_PR_WEBHOOK"),
		CDWebhook:      os.Getenv("SLACK_CD_WEBHOOK"),
		JobWebhook:     os.Getenv("SLACK_JOB_WEBHOOK"),
		DefaultWebhook: os.Getenv("SLACK_DEFAULT_WEBHOOK"),
	}

	eventRouter := router.NewRouter(routeConfig)

	route, err := eventRouter.Resolve(event)
	if err != nil {
		log.Fatalf("failed to resolve route: %v", err)
	}

	message := slack.Message{
		Text: fmt.Sprintf(
			"Event: %s\nStatus: %s\nRepo: %s\nBranch: %s\nSender: %s\nRoute: %s\nFallback: %t\nReason: %s",
			event.EventType,
			event.Status,
			event.Repository,
			event.Branch,
			event.Sender,
			route.RouteName,
			route.UsedFallback,
			route.Reason,
		),
	}

	client := slack.NewClient()

	if err := client.Send(route.WebhookURL, message); err != nil {
		log.Fatalf("failed to send Slack message: %v", err)
	}

	log.Println("Slack notification sent successfully")
}
```

---

## `main.go` flow explained simply

```text
Read CLI flags
    ↓
Create PipelineEvent
    ↓
Validate event
    ↓
Load webhook config from environment variables
    ↓
Ask router to choose webhook
    ↓
Create Slack message
    ↓
Ask Slack client to send message
```

This is exactly what clean backend code should look like.

---

## Go syntax comparison with Python

### Reading CLI input

Go:

```go
eventType := flag.String("event-type", "", "event type")
flag.Parse()
```

Python:

```python
parser.add_argument("--event-type")
args = parser.parse_args()
```

Important Go detail:

```go
eventType
```

is a pointer because `flag.String()` returns `*string`.

So we use:

```go
*eventType
```

to get the actual value.

---

### Error handling

Go:

```go
if err != nil {
	log.Fatalf("error: %v", err)
}
```

Python:

```python
try:
    do_something()
except Exception as e:
    print(e)
```

Go convention:

> Check errors immediately.

---

### Environment variables

Go:

```go
os.Getenv("SLACK_PR_WEBHOOK")
```

Python:

```python
os.getenv("SLACK_PR_WEBHOOK")
```

Very similar.

---

## 12. Hands-on tasks

### Task 1: Create package structure

Create this folder structure:

```text
slack-integration/
├── cmd/slack-notifier/main.go
├── pkg/notify/model/event.go
├── pkg/notify/router/router.go
├── pkg/notify/slack/client.go
└── go.mod
```

---

### Task 2: Add event model

Create:

```text
pkg/notify/model/event.go
```

Add:

* `PipelineEvent`
* `Validate()`
* constants for `pr`, `cd`, `job`

---

### Task 3: Add router logic

Create:

```text
pkg/notify/router/router.go
```

Add:

* `Config`
* `RouteResult`
* `Router`
* `NewRouter()`
* `Resolve()`

---

### Task 4: Add Slack client

Create:

```text
pkg/notify/slack/client.go
```

Add:

* `Message`
* `Client`
* `NewClient()`
* `Send()`

---

### Task 5: Test fallback locally

Export only CD webhook and skip Job webhook:

```bash
export SLACK_CD_WEBHOOK="https://hooks.slack.com/services/xxx/yyy/zzz"
export SLACK_JOB_WEBHOOK=""
```

Run:

```bash
go run cmd/slack-notifier/main.go \
  --event-type job \
  --status failed \
  --repository cloud-resource-onboarding \
  --branch feature/slack-router \
  --sender radheshyam
```

Expected behavior:

```text
Job webhook missing, fallback to CD webhook
```

---

## 13. Expected output

For PR event:

```bash
export SLACK_PR_WEBHOOK="https://hooks.slack.com/services/pr"
```

Command:

```bash
go run cmd/slack-notifier/main.go \
  --event-type pr \
  --status failed \
  --repository cloud-resource-onboarding \
  --branch feature/pr-check \
  --sender radheshyam
```

Expected route:

```text
Route: pr
Fallback: false
Reason: PR event routed to PR webhook
Slack notification sent successfully
```

---

For Job event with missing Job webhook:

```bash
export SLACK_CD_WEBHOOK="https://hooks.slack.com/services/cd"
export SLACK_JOB_WEBHOOK=""
```

Command:

```bash
go run cmd/slack-notifier/main.go \
  --event-type job \
  --status failed \
  --repository cloud-resource-onboarding \
  --branch feature/job-sync \
  --sender radheshyam
```

Expected route:

```text
Route: cd
Fallback: true
Reason: Job webhook missing, fallback to CD webhook
Slack notification sent successfully
```

---

For missing PR webhook:

```bash
export SLACK_PR_WEBHOOK=""
```

Command:

```bash
go run cmd/slack-notifier/main.go \
  --event-type pr \
  --status failed \
  --repository cloud-resource-onboarding
```

Expected error:

```text
failed to resolve route: PR webhook is missing
```

---

## 14. Common mistakes

### Mistake 1: Putting all logic in `main.go`

Bad:

```go
main.go has flags + validation + routing + HTTP + retry
```

Better:

```text
main.go coordinates only.
```

---

### Mistake 2: Router sends Slack message

Bad:

```go
func ResolveAndSend(event PipelineEvent) error
```

Better:

```go
func Resolve(event PipelineEvent) (RouteResult, error)
```

Router should decide, not send.

---

### Mistake 3: Slack client knows event type

Bad:

```go
func Send(event PipelineEvent) error {
	if event.EventType == "pr" {
		...
	}
}
```

Better:

```go
func Send(webhookURL string, message Message) error
```

Slack client should stay generic.

---

### Mistake 4: No fallback clarity

Bad:

```go
if jobWebhook == "" {
	webhook = cdWebhook
}
```

Better:

```go
RouteResult{
	UsedFallback: true,
	Reason: "Job webhook missing, fallback to CD webhook",
}
```

This helps debugging.

---

### Mistake 5: Ignoring errors

Bad:

```go
client.Send(webhookURL, message)
```

Good:

```go
if err := client.Send(webhookURL, message); err != nil {
	log.Fatalf("failed to send message: %v", err)
}
```

---

## 15. Debugging tips

### Tip 1: Print route result before sending

```go
log.Printf("route selected: %+v", route)
```

Output:

```text
route selected: {RouteName:cd WebhookURL:xxx UsedFallback:true Reason:Job webhook missing, fallback to CD webhook}
```

---

### Tip 2: Check environment variables

```bash
echo $SLACK_PR_WEBHOOK
echo $SLACK_CD_WEBHOOK
echo $SLACK_JOB_WEBHOOK
```

---

### Tip 3: Test routing without Slack first

Temporarily comment this:

```go
client.Send(route.WebhookURL, message)
```

And print:

```go
fmt.Println(message.Text)
```

This helps you verify routing before HTTP sending.

---

### Tip 4: Use fake webhook value during routing test

```bash
export SLACK_CD_WEBHOOK="dummy-cd-webhook"
```

Then test only router behavior.

---

### Tip 5: Add clear error messages

Bad:

```go
return err
```

Better:

```go
return fmt.Errorf("failed to resolve route for event type %s: %w", event.EventType, err)
```

---

## 16. DSA topic: Stack vs Queue

### Stack

A stack means:

> Last In, First Out.

Like a stack of plates.

You put plate A, then B, then C.

You remove C first.

```text
Push A
Push B
Push C

Pop -> C
Pop -> B
Pop -> A
```

Common use cases:

* Undo feature.
* Browser back button.
* Function call stack.
* Expression validation.

Python list as stack:

```python
stack = []
stack.append("A")
stack.append("B")
stack.pop()
```

---

### Queue

A queue means:

> First In, First Out.

Like people standing in a line.

First person enters first, first person leaves first.

```text
Enqueue A
Enqueue B
Enqueue C

Dequeue -> A
Dequeue -> B
Dequeue -> C
```

Common use cases:

* Job processing.
* Message queues.
* Print queue.
* Request handling.
* Pipeline tasks.

Python queue using list:

```python
queue = []
queue.append("A")
queue.append("B")
first = queue.pop(0)
```

But `pop(0)` can be inefficient for large lists.

---

### Stack vs Queue table

| Concept          | Stack              | Queue               |
| ---------------- | ------------------ | ------------------- |
| Rule             | Last In, First Out | First In, First Out |
| Example          | Plates             | Waiting line        |
| Add operation    | Push               | Enqueue             |
| Remove operation | Pop                | Dequeue             |
| Use case         | Undo               | Job processing      |

---

## 17. Small Go DSA problem: Queue implementation

### Problem

Create a simple queue in Go that supports:

1. `Enqueue(item string)`
2. `Dequeue() (string, error)`
3. `IsEmpty() bool`

Use it to process pipeline events in order.

---

### Pseudocode

```text
Create Queue struct with items list

FUNCTION Enqueue(item):
    Add item to end of list

FUNCTION Dequeue():
    IF queue is empty:
        return error
    Take first item
    Remove first item from list
    Return first item

FUNCTION IsEmpty():
    return length of list == 0
```

---

### Go solution

```go
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
			fmt.Println("error:", err)
			return
		}

		fmt.Println("Processing:", event)
	}
}
```

---

### Expected output

```text
Processing: pr-event
Processing: cd-event
Processing: job-event
```

---

### Python equivalent

```python
class Queue:
    def __init__(self):
        self.items = []

    def enqueue(self, item):
        self.items.append(item)

    def dequeue(self):
        if not self.items:
            raise Exception("queue is empty")
        return self.items.pop(0)

    def is_empty(self):
        return len(self.items) == 0
```

---

### Key Go syntax differences

| Python          | Go                         |
| --------------- | -------------------------- |
| `self`          | receiver like `(q *Queue)` |
| `list.append()` | `append(slice, item)`      |
| Exceptions      | `error` return             |
| `len(list)`     | `len(slice)`               |
| Class           | Struct + methods           |

---

## 18. Module-based practice task: Build a small task router

### Goal

Build a small router that routes tasks by type.

Task types:

| Task type | Handler        |
| --------- | -------------- |
| `build`   | Build handler  |
| `deploy`  | Deploy handler |
| `notify`  | Notify handler |
| unknown   | Error          |

---

### Folder structure

```text
task-router/
├── main.go
└── router/
    └── router.go
```

---

### `router/router.go`

```go
package router

import "fmt"

type Task struct {
	Type string
	Name string
}

type RouteResult struct {
	HandlerName string
	Reason      string
}

func RouteTask(task Task) (RouteResult, error) {
	switch task.Type {
	case "build":
		return RouteResult{
			HandlerName: "build-handler",
			Reason:      "Build task routed to build handler",
		}, nil

	case "deploy":
		return RouteResult{
			HandlerName: "deploy-handler",
			Reason:      "Deploy task routed to deploy handler",
		}, nil

	case "notify":
		return RouteResult{
			HandlerName: "notify-handler",
			Reason:      "Notify task routed to notify handler",
		}, nil

	default:
		return RouteResult{}, fmt.Errorf("unsupported task type: %s", task.Type)
	}
}
```

---

### `main.go`

```go
package main

import (
	"fmt"
	"log"

	"task-router/router"
)

func main() {
	task := router.Task{
		Type: "deploy",
		Name: "deploy-to-dev",
	}

	result, err := router.RouteTask(task)
	if err != nil {
		log.Fatalf("failed to route task: %v", err)
	}

	fmt.Println("Task Name:", task.Name)
	fmt.Println("Handler:", result.HandlerName)
	fmt.Println("Reason:", result.Reason)
}
```

---

### Expected output

```text
Task Name: deploy-to-dev
Handler: deploy-handler
Reason: Deploy task routed to deploy handler
```

---

### How this connects to your Slack project

This small task router is the same concept as your Slack router.

```text
Task type -> Handler
Event type -> Slack webhook
```

Same pattern.

```text
build  -> build-handler
deploy -> deploy-handler
notify -> notify-handler
```

Your project:

```text
pr  -> PR webhook
cd  -> CD webhook
job -> Job webhook or CD fallback
```

---

## 19. Revision checkpoint

You should now be able to answer these:

1. What is separation of concerns?
2. Why should `main.go` stay small?
3. What belongs in the `model` package?
4. What belongs in the `router` package?
5. What belongs in the `slack` package?
6. Why should router logic not send HTTP requests?
7. What is fallback logic?
8. What happens if `job` webhook is missing?
9. What is the difference between stack and queue?
10. Why is queue useful for pipeline/job processing?

---

## 20. Homework

### Homework 1: Add route logging

Add this after route resolution:

```go
log.Printf("selected route=%s fallback=%t reason=%s", route.RouteName, route.UsedFallback, route.Reason)
```

---

### Homework 2: Add validation for supported event types

Inside `event.go`, update `Validate()`:

```go
func (e PipelineEvent) Validate() error {
	if e.EventType == "" {
		return fmt.Errorf("event type is required")
	}

	switch e.EventType {
	case EventTypePR, EventTypeCD, EventTypeJob:
		// valid event type
	default:
		return fmt.Errorf("unsupported event type: %s", e.EventType)
	}

	if e.Status == "" {
		return fmt.Errorf("status is required")
	}

	if e.Repository == "" {
		return fmt.Errorf("repository is required")
	}

	return nil
}
```

---

### Homework 3: Add dry-run mode

Dry run means:

> Print the message and route, but do not send Slack notification.

Example CLI:

```bash
go run cmd/slack-notifier/main.go \
  --event-type job \
  --status failed \
  --repository cloud-resource-onboarding \
  --dry-run true
```

Expected output:

```text
DRY RUN MODE
Route: cd
Fallback: true
Reason: Job webhook missing, fallback to CD webhook
Message will not be sent to Slack
```

---

### Homework 4: Extend task router

Add one more task type:

```text
test -> test-handler
```

Expected output:

```text
Task Name: run-unit-tests
Handler: test-handler
Reason: Test task routed to test handler
```

---

# Day 4 final mental model

```text
main.go
  = coordinator

model
  = defines and validates data

router
  = decides destination

slack client
  = sends HTTP request

fallback
  = safe backup route

queue
  = process events in order
```

The clean project mindset is:

> Do not put everything in one file. Give each package one job, connect them clearly, and make every decision easy to test and debug.
