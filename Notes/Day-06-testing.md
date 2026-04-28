# Day 6 — Go Testing Basics for `slack-integration` Project

## 1. Day 6 learning goals

Today you will learn how to test your Go project before connecting it deeply with Tekton or Slack.

By the end of Day 6, you should understand:

1. What testing means in Go.
2. Why Go test files end with `_test.go`.
3. How to run tests using `go test`.
4. How to test validation logic.
5. How table-driven tests work.
6. How to test Slack webhook logic without calling real Slack.
7. How to use `httptest` mock servers.
8. How to test routing logic.
9. How to test Slack message/payload formatting.
10. How to build and test a retry handler.
11. Basic recursion in Go.

---

## 2. Quick revision of Days 1 to 5

### Day 1 — CLI basics

You learned how input enters the Go program.

```text
CLI flags → Go variables → main.go
```

Example:

```bash
go run cmd/slack-notifier/main.go \
  --event-type pr \
  --stage validation \
  --status failed
```

Python comparison:

```python
# Python
argparse.ArgumentParser()
```

Go comparison:

```go
// Go
flag.String("event-type", "", "event type")
```

---

### Day 2 — Structs and model layer

You learned that Go structs group related data.

```go
type PipelineEvent struct {
    EventType    string
    Status       string
    PipelineName string
}
```

Python comparison:

```python
@dataclass
class PipelineEvent:
    event_type: str
    status: str
    pipeline_name: str
```

In Go, structs are commonly used instead of passing many loose values.

---

### Day 3 — JSON, HTTP, Slack webhook

You learned that Go structs can become JSON.

```go
type SlackMessage struct {
    Text string `json:"text"`
}
```

Python comparison:

```python
payload = {"text": "Build failed"}
requests.post(webhook_url, json=payload)
```

Go comparison:

```go
json.Marshal(payload)
http.Post(webhookURL, "application/json", body)
```

---

### Day 4 — Router logic

You learned that router logic decides where the notification should go.

```text
PR event  → PR Slack webhook
CD event  → CD Slack webhook
Job event → Job webhook or fallback CD webhook
```

---

### Day 5 — Error handling and logging

You learned this Go pattern:

```go
if err != nil {
    return err
}
```

Python comparison:

```python
try:
    do_something()
except Exception as e:
    return e
```

Go usually returns errors explicitly. Python usually uses exceptions.

---

## 3. Explain testing in Go very simply

Testing means checking whether your code behaves correctly.

Instead of manually running your CLI every time, you write small test functions.

### Manual checking

You run:

```bash
go run cmd/slack-notifier/main.go --event-type pr --status failed
```

Then you manually check:

```text
Did Slack message come?
Was the format correct?
Did routing work?
Did retry happen?
```

Problem: manual checking is slow and easy to forget.

---

### Automated testing

You write tests once:

```go
func TestPipelineEventValidate(t *testing.T) {
    // test code here
}
```

Then run:

```bash
go test ./...
```

Go checks everything automatically.

---

### Python comparison

Python usually uses:

```bash
pytest
```

Test file:

```python
def test_validate_event():
    assert validate_event(event) is True
```

Go uses:

```bash
go test
```

Test file:

```go
func TestValidateEvent(t *testing.T) {
    // checks
}
```

---

## 4. Explain how to run tests

### Run all tests

```bash
go test ./...
```

Meaning:

```text
Run tests in all packages under current project.
```

---

### Run tests in one package

```bash
go test ./pkg/notify/model
```

---

### Run tests with detailed output

```bash
go test -v ./...
```

`-v` means verbose.

---

### Run one specific test

```bash
go test -v ./pkg/notify/model -run TestPipelineEventValidate
```

---

### Check test coverage

```bash
go test -cover ./...
```

Coverage means:

```text
How much of your code is executed by tests?
```

---

## 5. Explain what unit testing means

A unit test checks one small piece of logic.

Example:

```text
Only test validation logic.
Only test routing logic.
Only test payload building.
Only test retry logic.
```

Unit testing should not depend on:

```text
Real Slack
Real Tekton
Real GitHub
Real Kubernetes cluster
Real network dependency
```

Why?

Because unit tests should be:

```text
Fast
Reliable
Repeatable
Safe
```

---

### In your project

Good unit test targets:

```text
PipelineEvent.Validate()
Router.WebhookFor()
BuildSlackMessage()
Retry.Do()
SlackClient.Send() with mock server
```

Avoid this in unit tests:

```text
Calling real Slack webhook
Calling real GitHub
Triggering real Tekton pipeline
```

---

## 6. Explain table-driven tests in a beginner-friendly way

A table-driven test means:

```text
Create many test cases in a list.
Loop over them.
Run the same test logic for each case.
```

This is very common in Go.

---

### Simple idea

Instead of writing:

```go
func TestValidEvent(t *testing.T) {}
func TestMissingEventType(t *testing.T) {}
func TestMissingStatus(t *testing.T) {}
```

You write one test:

```go
func TestPipelineEventValidate(t *testing.T) {
    tests := []struct {
        name    string
        event   PipelineEvent
        wantErr bool
    }{
        {"valid event", validEvent, false},
        {"missing event type", invalidEvent, true},
        {"missing status", invalidEvent, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

---

### Python comparison

Python `pytest` equivalent:

```python
@pytest.mark.parametrize("event,want_error", [
    (valid_event, False),
    (missing_event_type, True),
])
def test_validate_event(event, want_error):
    ...
```

Go equivalent:

```go
tests := []struct {
    name    string
    event   PipelineEvent
    wantErr bool
}{...}
```

---

## 7. Explain mock server testing using `httptest`

Your Slack client sends HTTP POST requests.

But in tests, you should not call real Slack.

### Why not use real Slack webhook in tests?

Because:

```text
It can spam real Slack channels.
It depends on internet/network.
Webhook may expire or fail.
Tests become slow.
Secrets may leak.
CI/Tekton tests should be safe.
```

Instead, use a fake HTTP server.

---

### Mock server idea

```text
Test code starts a fake server.
Slack client sends request to fake server.
Fake server checks request.
Fake server returns success or failure.
Test verifies behavior.
```

---

### Python comparison

Python options:

```python
responses
requests-mock
unittest.mock
```

Go option:

```go
httptest.NewServer(...)
```

---

## 8. Show how to test validation, routing, and formatting

In your project, this is the testing flow:

```text
Validation test:
PipelineEvent → Validate() → expect error or no error

Router test:
PipelineEvent → Router.WebhookFor() → expect correct webhook

Formatting test:
PipelineEvent → BuildSlackMessage() → expect correct text/fields

Slack client test:
SlackMessage → Send() → mock server receives request

Retry test:
Temporary failure → retry → final success/failure
```

---

## 9. Pseudocode first for test logic

### A. Validation test pseudocode

```text
Create test cases:
  - valid PR event
  - missing event type
  - missing status
  - failed event without failure reason

For each test case:
  call Validate()
  if want error but error is nil → fail test
  if no error expected but error exists → fail test
```

---

### B. Router test pseudocode

```text
Create router config:
  PR webhook = "pr-url"
  CD webhook = "cd-url"
  Job webhook = ""

Test cases:
  PR event should return PR webhook
  CD event should return CD webhook
  Job event should fallback to CD webhook
  Unknown event should return error
```

---

### C. Slack payload test pseudocode

```text
Create failed PipelineEvent
Call BuildSlackMessage(event)
Check:
  message text contains pipeline name
  color is danger/red
  fields include failed step
  fields include error message
```

---

### D. Mock Slack server pseudocode

```text
Start fake HTTP server
Fake server checks:
  method should be POST
  content type should be JSON
  body should contain expected text

Call SlackClient.Send(fakeServerURL, message)

If no error → test passes
```

---

### E. Retry handler pseudocode

```text
attempt = 0

retry 3 times:
  attempt++
  if attempt < 3:
      return temporary error
  else:
      return nil

Check:
  final error is nil
  total attempts are 3
```

---

# 10. Real Go test examples

Below is a simple project-focused version.

Recommended structure:

```text
slack-integration/
├── cmd/
│   └── slack-notifier/
│       └── main.go
└── pkg/
    └── notify/
        ├── model/
        │   ├── event.go
        │   └── event_test.go
        ├── router/
        │   ├── router.go
        │   └── router_test.go
        ├── slack/
        │   ├── message.go
        │   ├── client.go
        │   └── client_test.go
        └── retry/
            ├── retry.go
            └── retry_test.go
```

---

## 10.1 Validation logic

### `pkg/notify/model/event.go`

```go
package model

import (
	"fmt"
	"strings"
)

type PipelineEvent struct {
	EventType    string
	Stage        string
	Status       string
	PipelineName string
	Repository   string
	FailedStep   string
	ErrorMessage string
}

func (e PipelineEvent) Validate() error {
	if strings.TrimSpace(e.EventType) == "" {
		return fmt.Errorf("event type is required")
	}

	if strings.TrimSpace(e.Stage) == "" {
		return fmt.Errorf("stage is required")
	}

	if strings.TrimSpace(e.Status) == "" {
		return fmt.Errorf("status is required")
	}

	if strings.TrimSpace(e.PipelineName) == "" {
		return fmt.Errorf("pipeline name is required")
	}

	if e.Status == "failed" &&
		strings.TrimSpace(e.FailedStep) == "" &&
		strings.TrimSpace(e.ErrorMessage) == "" {
		return fmt.Errorf("failed event must include failed step or error message")
	}

	return nil
}
```

### Explanation

```go
func (e PipelineEvent) Validate() error
```

This means `Validate` is a method on `PipelineEvent`.

Python comparison:

```python
@dataclass
class PipelineEvent:
    def validate(self):
        ...
```

Go:

```go
return fmt.Errorf("event type is required")
```

Python:

```python
raise ValueError("event type is required")
```

But in Go, we return the error instead of throwing it.

---

## 10.2 Validation test

### `pkg/notify/model/event_test.go`

```go
package model

import "testing"

func TestPipelineEventValidate(t *testing.T) {
	tests := []struct {
		name    string
		event   PipelineEvent
		wantErr bool
	}{
		{
			name: "valid pr event",
			event: PipelineEvent{
				EventType:    "pr",
				Stage:        "validation",
				Status:       "succeeded",
				PipelineName: "pr-check",
			},
			wantErr: false,
		},
		{
			name: "missing event type",
			event: PipelineEvent{
				Stage:        "validation",
				Status:       "succeeded",
				PipelineName: "pr-check",
			},
			wantErr: true,
		},
		{
			name: "missing status",
			event: PipelineEvent{
				EventType:    "pr",
				Stage:        "validation",
				PipelineName: "pr-check",
			},
			wantErr: true,
		},
		{
			name: "failed event without failure context",
			event: PipelineEvent{
				EventType:    "pr",
				Stage:        "validation",
				Status:       "failed",
				PipelineName: "pr-check",
			},
			wantErr: true,
		},
		{
			name: "failed event with failed step",
			event: PipelineEvent{
				EventType:    "pr",
				Stage:        "validation",
				Status:       "failed",
				PipelineName: "pr-check",
				FailedStep:   "go-test",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()

			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
```

### Important syntax

```go
tests := []struct {
    name string
}{...}
```

This means:

```text
Create a slice/list of anonymous structs.
```

Python comparison:

```python
tests = [
    {"name": "valid pr event", "event": event, "want_error": False}
]
```

Go uses structs instead of dictionaries for type safety.

---

## 10.3 Router logic

### `pkg/notify/router/router.go`

```go
package router

import (
	"fmt"

	"slack-integration/pkg/notify/model"
)

type Config struct {
	PRWebhook  string
	CDWebhook  string
	JobWebhook string
}

type Router struct {
	config Config
}

func NewRouter(config Config) Router {
	return Router{config: config}
}

func (r Router) WebhookFor(event model.PipelineEvent) (string, error) {
	switch event.EventType {
	case "pr":
		if r.config.PRWebhook == "" {
			return "", fmt.Errorf("pr webhook is not configured")
		}
		return r.config.PRWebhook, nil

	case "cd":
		if r.config.CDWebhook == "" {
			return "", fmt.Errorf("cd webhook is not configured")
		}
		return r.config.CDWebhook, nil

	case "job":
		if r.config.JobWebhook != "" {
			return r.config.JobWebhook, nil
		}

		if r.config.CDWebhook != "" {
			return r.config.CDWebhook, nil
		}

		return "", fmt.Errorf("job webhook and fallback cd webhook are not configured")

	default:
		return "", fmt.Errorf("unsupported event type: %s", event.EventType)
	}
}
```

### Project meaning

```text
PR event  → PR webhook
CD event  → CD webhook
Job event → Job webhook
Job event without Job webhook → CD webhook fallback
```

---

## 10.4 Router test

### `pkg/notify/router/router_test.go`

```go
package router

import (
	"testing"

	"slack-integration/pkg/notify/model"
)

func TestWebhookFor(t *testing.T) {
	r := NewRouter(Config{
		PRWebhook: "https://example.com/pr",
		CDWebhook: "https://example.com/cd",
		// JobWebhook intentionally empty to test fallback
	})

	tests := []struct {
		name        string
		event       model.PipelineEvent
		wantWebhook string
		wantErr     bool
	}{
		{
			name: "pr event goes to pr webhook",
			event: model.PipelineEvent{
				EventType: "pr",
			},
			wantWebhook: "https://example.com/pr",
			wantErr:     false,
		},
		{
			name: "cd event goes to cd webhook",
			event: model.PipelineEvent{
				EventType: "cd",
			},
			wantWebhook: "https://example.com/cd",
			wantErr:     false,
		},
		{
			name: "job event falls back to cd webhook",
			event: model.PipelineEvent{
				EventType: "job",
			},
			wantWebhook: "https://example.com/cd",
			wantErr:     false,
		},
		{
			name: "unknown event returns error",
			event: model.PipelineEvent{
				EventType: "unknown",
			},
			wantWebhook: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWebhook, err := r.WebhookFor(tt.event)

			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if gotWebhook != tt.wantWebhook {
				t.Fatalf("expected webhook %q, got %q", tt.wantWebhook, gotWebhook)
			}
		})
	}
}
```

---

## 10.5 Slack payload building

### `pkg/notify/slack/message.go`

```go
package slack

import (
	"fmt"

	"slack-integration/pkg/notify/model"
)

type Message struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Title  string  `json:"title"`
	Color  string  `json:"color"`
	Fields []Field `json:"fields"`
}

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

func BuildMessage(event model.PipelineEvent) Message {
	color := "good"

	if event.Status == "failed" {
		color = "danger"
	}

	fields := []Field{
		{
			Title: "Event Type",
			Value: event.EventType,
			Short: true,
		},
		{
			Title: "Stage",
			Value: event.Stage,
			Short: true,
		},
		{
			Title: "Status",
			Value: event.Status,
			Short: true,
		},
		{
			Title: "Pipeline",
			Value: event.PipelineName,
			Short: true,
		},
	}

	if event.FailedStep != "" {
		fields = append(fields, Field{
			Title: "Failed Step",
			Value: event.FailedStep,
			Short: true,
		})
	}

	if event.ErrorMessage != "" {
		fields = append(fields, Field{
			Title: "Error Message",
			Value: event.ErrorMessage,
			Short: false,
		})
	}

	return Message{
		Text: fmt.Sprintf("Pipeline %s: %s", event.Status, event.PipelineName),
		Attachments: []Attachment{
			{
				Title:  "Pipeline Notification",
				Color:  color,
				Fields: fields,
			},
		},
	}
}
```

---

## 10.6 Slack payload test

### `pkg/notify/slack/message_test.go`

```go
package slack

import (
	"testing"

	"slack-integration/pkg/notify/model"
)

func TestBuildMessageForFailedEvent(t *testing.T) {
	event := model.PipelineEvent{
		EventType:    "pr",
		Stage:        "validation",
		Status:       "failed",
		PipelineName: "pr-check",
		FailedStep:   "go-test",
		ErrorMessage: "unit tests failed",
	}

	msg := BuildMessage(event)

	if msg.Text != "Pipeline failed: pr-check" {
		t.Fatalf("unexpected text: %s", msg.Text)
	}

	if len(msg.Attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(msg.Attachments))
	}

	attachment := msg.Attachments[0]

	if attachment.Color != "danger" {
		t.Fatalf("expected color danger, got %s", attachment.Color)
	}

	if len(attachment.Fields) < 6 {
		t.Fatalf("expected failure fields, got %d fields", len(attachment.Fields))
	}
}
```

### What this test checks

```text
Failed event should create failure-style Slack message.
Color should be danger.
Failure context should be included.
```

This is important because Slack notification quality matters during Tekton failures.

---

## 10.7 Slack client with mock server

### `pkg/notify/slack/client.go`

```go
package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	return Client{
		httpClient: httpClient,
	}
}

func (c Client) Send(webhookURL string, message Message) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create slack request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send slack request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("slack returned non-success status: %d", resp.StatusCode)
	}

	return nil
}
```

---

## 10.8 Test Slack client using `httptest`

### `pkg/notify/slack/client_test.go`

```go
package slack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientSendSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST request, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("expected application/json content type")
		}

		var msg Message
		err := json.NewDecoder(r.Body).Decode(&msg)
		if err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if msg.Text != "hello from test" {
			t.Fatalf("unexpected message text: %s", msg.Text)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.Client())

	err := client.Send(server.URL, Message{
		Text: "hello from test",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
```

### Explanation

```go
httptest.NewServer(...)
```

Creates fake server.

```go
server.URL
```

Fake webhook URL.

```go
server.Client()
```

HTTP client connected to fake server.

Python comparison:

```python
# Python idea
with responses.RequestsMock() as rsps:
    rsps.add(responses.POST, fake_url, status=200)
```

Go equivalent:

```go
server := httptest.NewServer(...)
```

---

## 10.9 Test Slack server failure

```go
func TestClientSendFailureStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.Client())

	err := client.Send(server.URL, Message{
		Text: "this should fail",
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
```

This test checks:

```text
If Slack/fake server returns 500,
client should return error.
```

This is important before adding retry.

---

# 11. Hands-on tasks

## Task 1 — Create validation tests

Create:

```text
pkg/notify/model/event_test.go
```

Test these cases:

```text
valid PR event
missing event type
missing stage
missing status
missing pipeline name
failed event without failed step/error message
failed event with failed step
```

---

## Task 2 — Create router tests

Create:

```text
pkg/notify/router/router_test.go
```

Test:

```text
PR → PR webhook
CD → CD webhook
Job → Job webhook
Job without Job webhook → CD webhook fallback
Unknown event → error
```

---

## Task 3 — Create Slack message formatting tests

Create:

```text
pkg/notify/slack/message_test.go
```

Test:

```text
Success event uses color good
Failed event uses color danger
Failed event includes failed step
Failed event includes error message
```

---

## Task 4 — Create Slack client mock server test

Create:

```text
pkg/notify/slack/client_test.go
```

Test:

```text
HTTP method is POST
Content-Type is application/json
Request body contains expected text
200 response returns nil error
500 response returns error
```

---

## Task 5 — Build retry handler and test it

This is your module-based practice task for Day 6.

---

# 12. Expected output

Run:

```bash
go test ./...
```

Expected output:

```text
ok  	slack-integration/pkg/notify/model	0.003s
ok  	slack-integration/pkg/notify/router	0.004s
ok  	slack-integration/pkg/notify/slack	0.006s
ok  	slack-integration/pkg/notify/retry	0.003s
```

Run verbose:

```bash
go test -v ./...
```

Expected style:

```text
=== RUN   TestPipelineEventValidate
=== RUN   TestPipelineEventValidate/valid_pr_event
=== RUN   TestPipelineEventValidate/missing_event_type
--- PASS: TestPipelineEventValidate (0.00s)
    --- PASS: TestPipelineEventValidate/valid_pr_event (0.00s)
    --- PASS: TestPipelineEventValidate/missing_event_type (0.00s)
PASS
```

---

# 13. Common mistakes

## Mistake 1 — Test file name is wrong

Wrong:

```text
eventtest.go
event_testfile.go
```

Correct:

```text
event_test.go
```

Go only detects test files ending with:

```text
_test.go
```

---

## Mistake 2 — Test function name is wrong

Wrong:

```go
func testValidate(t *testing.T) {}
```

Correct:

```go
func TestValidate(t *testing.T) {}
```

Test function must start with capital `Test`.

---

## Mistake 3 — Forgetting `*testing.T`

Wrong:

```go
func TestValidate() {}
```

Correct:

```go
func TestValidate(t *testing.T) {}
```

---

## Mistake 4 — Calling real Slack webhook in test

Avoid:

```go
client.Send("https://hooks.slack.com/services/real/webhook", msg)
```

Use:

```go
server := httptest.NewServer(...)
client.Send(server.URL, msg)
```

---

## Mistake 5 — Not checking both error cases

Weak test:

```go
if err != nil {
    t.Fatal(err)
}
```

Better test:

```go
if wantErr && err == nil {
    t.Fatalf("expected error, got nil")
}

if !wantErr && err != nil {
    t.Fatalf("expected no error, got %v", err)
}
```

---

## Mistake 6 — Sleeping too much in retry tests

Avoid:

```go
retry.Do(3, 5*time.Second, fn)
```

That makes tests slow.

Use:

```go
retry.Do(3, 0, fn)
```

---

# 14. Debugging tips for failing tests

## Tip 1 — Run one package only

```bash
go test -v ./pkg/notify/model
```

---

## Tip 2 — Run one test only

```bash
go test -v ./pkg/notify/model -run TestPipelineEventValidate
```

---

## Tip 3 — Print temporary debug output

```go
t.Logf("got webhook: %s", gotWebhook)
```

Then run:

```bash
go test -v ./pkg/notify/router
```

---

## Tip 4 — Read the failure message carefully

Example:

```text
expected webhook "https://example.com/pr", got "https://example.com/cd"
```

Meaning:

```text
Router logic is sending PR event to wrong webhook.
```

---

## Tip 5 — Check package name

If file is inside:

```text
pkg/notify/model
```

Usually use:

```go
package model
```

Not:

```go
package main
```

---

## Tip 6 — Check import path

If your module name in `go.mod` is:

```go
module github.com/yourname/slack-integration
```

Then import should be:

```go
import "github.com/yourname/slack-integration/pkg/notify/model"
```

If your module name is:

```go
module slack-integration
```

Then import should be:

```go
import "slack-integration/pkg/notify/model"
```

---

# 15. One DSA topic — Recursion

## Simple definition

Recursion means:

```text
A function calls itself to solve a smaller version of the same problem.
```

---

## Real-life example

Imagine you are standing in a queue and want to know your position.

You ask the person in front:

```text
What is your position?
```

That person asks the next person.

Eventually, the first person says:

```text
I am position 1.
```

Then everyone adds `1`.

That is recursion.

---

## Recursion has two parts

### 1. Base case

This stops recursion.

```text
If n == 1, stop.
```

### 2. Recursive case

This calls the same function again.

```text
n + sum(n-1)
```

---

## Python comparison

Python:

```python
def sum_n(n):
    if n == 1:
        return 1
    return n + sum_n(n - 1)
```

Go:

```go
func SumN(n int) int {
	if n == 1 {
		return 1
	}

	return n + SumN(n-1)
}
```

---

## Key syntax differences

Python:

```python
if n == 1:
    return 1
```

Go:

```go
if n == 1 {
    return 1
}
```

Python uses indentation. Go uses curly braces.

---

# 16. One Go DSA problem — Sum of numbers from 1 to N using recursion

## Problem

Write a function that returns:

```text
1 + 2 + 3 + ... + n
```

Example:

```text
n = 5
answer = 15
```

Because:

```text
1 + 2 + 3 + 4 + 5 = 15
```

---

## Pseudocode

```text
function SumN(n):
    if n <= 0:
        return 0

    return n + SumN(n - 1)
```

---

## Go solution

```go
package main

import "fmt"

func SumN(n int) int {
	if n <= 0 {
		return 0
	}

	return n + SumN(n-1)
}

func main() {
	result := SumN(5)
	fmt.Println(result)
}
```

Expected output:

```text
15
```

---

## Dry run

```text
SumN(5)
= 5 + SumN(4)
= 5 + 4 + SumN(3)
= 5 + 4 + 3 + SumN(2)
= 5 + 4 + 3 + 2 + SumN(1)
= 5 + 4 + 3 + 2 + 1 + SumN(0)
= 5 + 4 + 3 + 2 + 1 + 0
= 15
```

---

## Test for recursion function

```go
package main

import "testing"

func TestSumN(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want int
	}{
		{
			name: "sum of 5",
			n:    5,
			want: 15,
		},
		{
			name: "sum of 1",
			n:    1,
			want: 1,
		},
		{
			name: "sum of 0",
			n:    0,
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SumN(tt.n)

			if got != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}
}
```

---

# 17. One module-based practice task — Build a retry handler and test it

In your Slack integration project, retry is useful because Slack/webhook/network calls may fail temporarily.

Example failures:

```text
Temporary network issue
Slack returns 500
Request timeout
DNS issue
```

Instead of failing immediately, we retry a few times.

---

## 17.1 Retry handler code

Create:

```text
pkg/notify/retry/retry.go
```

```go
package retry

import (
	"fmt"
	"time"
)

func Do(maxAttempts int, delay time.Duration, fn func() error) error {
	if maxAttempts <= 0 {
		return fmt.Errorf("max attempts must be greater than zero")
	}

	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		lastErr = fn()

		if lastErr == nil {
			return nil
		}

		if attempt < maxAttempts && delay > 0 {
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", maxAttempts, lastErr)
}
```

---

## 17.2 Explanation

```go
func Do(maxAttempts int, delay time.Duration, fn func() error) error
```

This function accepts another function:

```go
fn func() error
```

Python comparison:

```python
def retry(max_attempts, delay, fn):
    ...
```

In Python, passing function:

```python
retry(3, 1, send_slack_message)
```

In Go:

```go
retry.Do(3, time.Second, sendSlackMessage)
```

---

## 17.3 Retry test — success after retry

Create:

```text
pkg/notify/retry/retry_test.go
```

```go
package retry

import (
	"fmt"
	"testing"
	"time"
)

func TestDoSucceedsAfterRetry(t *testing.T) {
	attempts := 0

	err := Do(3, 0*time.Second, func() error {
		attempts++

		if attempts < 3 {
			return fmt.Errorf("temporary failure")
		}

		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}
```

---

## 17.4 Retry test — final failure

```go
func TestDoFailsAfterMaxAttempts(t *testing.T) {
	attempts := 0

	err := Do(3, 0*time.Second, func() error {
		attempts++
		return fmt.Errorf("always failing")
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}
```

---

## 17.5 Retry test — invalid attempts

```go
func TestDoWithInvalidAttempts(t *testing.T) {
	err := Do(0, 0*time.Second, func() error {
		return nil
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
```

---

## 17.6 How retry connects to Slack client

Later you can use it like this:

```go
err := retry.Do(3, 2*time.Second, func() error {
	return slackClient.Send(webhookURL, message)
})

if err != nil {
	return fmt.Errorf("failed to send slack notification after retries: %w", err)
}
```

Project flow:

```text
PipelineEvent
   ↓
Validate
   ↓
Build Slack Message
   ↓
Route Webhook
   ↓
Retry Send
   ↓
Slack Webhook
```

---

# 18. Revision checkpoint

Before moving to Day 7, you should be able to answer these:

1. Why do Go test files end with `_test.go`?
2. What does `go test ./...` do?
3. What is a unit test?
4. Why should real Slack webhook calls not be used in tests?
5. What is a table-driven test?
6. What does `t.Run()` do?
7. What does `t.Fatalf()` do?
8. What does `httptest.NewServer()` do?
9. How can you test router logic?
10. How can you test Slack payload formatting?
11. Why are tests important before Tekton automation?
12. What is recursion?
13. What is a base case?
14. What is a recursive case?
15. Why should retry tests use `0*time.Second` delay?

---

# 19. Homework

## Homework 1 — Add validation tests

Add tests for:

```text
empty repository
invalid event type
invalid status
failed event with only error message
failed event with both failed step and error message
```

---

## Homework 2 — Add router tests

Add tests for:

```text
missing PR webhook
missing CD webhook
job webhook present
job webhook missing but CD fallback present
job webhook missing and CD fallback missing
```

---

## Homework 3 — Add Slack payload tests

Add tests for:

```text
successful PR message
failed PR message
failed CD message
failed Job message
message includes pipeline name
message includes failed step
message includes error message
```

---

## Homework 4 — Add mock server tests

Add tests for:

```text
server returns 200
server returns 400
server returns 500
server receives invalid method
server receives invalid JSON
```

---

## Homework 5 — Improve retry handler

Enhance retry logic to support:

```text
attempt number logging
last error wrapping
optional max delay
future exponential backoff
```

Simple future version idea:

```text
attempt 1 → wait 1 second
attempt 2 → wait 2 seconds
attempt 3 → wait 4 seconds
```

---

# Final Day 6 mental model

```text
Manual checking:
"I ran it once and it looked okay."

Automated testing:
"Every important rule is checked again and again safely."

Go testing:
Small test functions + go test command.

Table-driven testing:
Many test cases + one loop.

httptest:
Fake Slack server instead of real Slack.

Before Tekton:
Test logic locally first, then automate with confidence.
```

Day 6 takeaway:

```text
Testing protects your project before automation makes mistakes faster.
```
