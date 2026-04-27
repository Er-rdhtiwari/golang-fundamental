# Day 5 — Go Error Handling, Structured Logging, and Failure-Aware Code

Today’s focus is **how real backend systems handle failures clearly**.

In your `slack-integration` project, this means:

```text
CLI input
   ↓
PipelineEvent validation
   ↓
Router selects webhook
   ↓
Slack client sends message
   ↓
If something fails → return error + log useful context
```

---

# 1. Day 5 Learning Goals

By the end of Day 5, you should understand:

1. How Go handles errors using `error`
2. Why Go uses `if err != nil` so often
3. How Go error handling is different from Python exceptions
4. How to create beginner-friendly custom errors
5. How to wrap errors using `fmt.Errorf(... %w ...)`
6. How to use `errors.Is()` and `errors.As()`
7. How structured logging works with `zerolog`
8. Why structured logs are better than `fmt.Println`
9. What fields to log in your Slack notification project
10. How logs help debug Tekton pipeline failures
11. Basic linked list concepts in Go
12. How to build a simple log parser module

---

# 2. Quick Revision of Days 1 to 4

## Day 1 — CLI Basics

You learned how input enters a Go CLI app.

Example:

```bash
go run main.go \
  --event-type pr \
  --stage started \
  --status running
```

In Go, CLI flags are usually read using the `flag` package.

```go
eventType := flag.String("event-type", "", "event type")
flag.Parse()
```

Python equivalent:

```python
import argparse

parser = argparse.ArgumentParser()
parser.add_argument("--event-type")
args = parser.parse_args()
```

---

## Day 2 — Structs and Event Model

You learned that Go structs are like Python classes or dataclasses used mainly to hold data.

Go:

```go
type PipelineEvent struct {
	EventType string
	Stage     string
	Status    string
}
```

Python equivalent:

```python
from dataclasses import dataclass

@dataclass
class PipelineEvent:
    event_type: str
    stage: str
    status: str
```

---

## Day 3 — JSON, HTTP, Slack Webhook

You learned that a Go struct can become JSON using `json` tags.

```go
type SlackMessage struct {
	Text string `json:"text"`
}
```

Python equivalent:

```python
payload = {
    "text": "Pipeline started"
}
```

---

## Day 4 — Router Logic and Separation of Concerns

You learned that each package should have a clear responsibility.

```text
main.go
  → reads CLI input

model package
  → defines and validates PipelineEvent

router package
  → decides which Slack webhook to use

slack package
  → sends HTTP POST request to Slack
```

Today, we improve this by adding:

```text
errors + logs + failure context
```

---

# 3. Explain Go Error Handling Very Simply

In Go, errors are normal return values.

That means a function usually returns:

```go
result, error
```

Example:

```go
func divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("cannot divide by zero")
	}

	return a / b, nil
}
```

Using it:

```go
result, err := divide(10, 0)
if err != nil {
	fmt.Println("error:", err)
	return
}

fmt.Println("result:", result)
```

## Python comparison

In Python, you normally use exceptions:

```python
def divide(a, b):
    if b == 0:
        raise ValueError("cannot divide by zero")
    return a / b

try:
    result = divide(10, 0)
except ValueError as e:
    print("error:", e)
```

## Main difference

Go prefers this style:

```go
value, err := doSomething()
if err != nil {
	return err
}
```

Python prefers this style:

```python
try:
    value = do_something()
except Exception as e:
    handle_error(e)
```

Simple mental model:

```text
Python:
  Error jumps to except block

Go:
  Error is returned like normal data
```

---

# 4. Why `if err != nil` Is Used So Often

In Go, `nil` means “nothing” or “no value”.

For errors:

```go
err == nil
```

means:

```text
No error happened
```

And:

```go
err != nil
```

means:

```text
Something failed
```

Example:

```go
file, err := os.Open("config.json")
if err != nil {
	return err
}
defer file.Close()
```

This means:

```text
Try to open file.
If opening failed, stop and return the error.
If no error, continue.
```

## Python equivalent

```python
try:
    file = open("config.json")
except Exception as e:
    return e
```

## Why Go does this explicitly

Because Go wants failure handling to be visible.

In production systems, this is useful because you can clearly see:

```text
Where can this function fail?
What happens when it fails?
Do we retry?
Do we log?
Do we return?
Do we fallback?
```

In your Slack project, this matters because many things can fail:

```text
Invalid CLI input
Missing webhook
Unknown event type
Slack timeout
Slack returns 500
Slack returns 400
Network failure
JSON encoding failure
```

So Go wants you to handle these failures clearly.

---

# 5. Custom Errors with Simple Examples

A custom error is an error that gives your application a specific meaning.

Simple error:

```go
var ErrInvalidEvent = errors.New("invalid pipeline event")
```

Usage:

```go
func validateEvent(eventType string) error {
	if eventType == "" {
		return ErrInvalidEvent
	}

	return nil
}
```

Calling code:

```go
err := validateEvent("")
if err != nil {
	fmt.Println("validation failed:", err)
}
```

Output:

```text
validation failed: invalid pipeline event
```

---

## Python equivalent

```python
class InvalidEventError(Exception):
    pass

def validate_event(event_type):
    if not event_type:
        raise InvalidEventError("invalid pipeline event")
```

---

## Go Custom Error Style 1 — Sentinel Error

A sentinel error is a reusable predefined error.

```go
var ErrMissingWebhook = errors.New("missing webhook configuration")
```

You can compare it later using:

```go
errors.Is(err, ErrMissingWebhook)
```

Example:

```go
if errors.Is(err, ErrMissingWebhook) {
	fmt.Println("webhook is missing")
}
```

This is similar to checking exception type in Python:

```python
except MissingWebhookError:
    print("webhook is missing")
```

---

## Go Custom Error Style 2 — Error with Context

Sometimes the base error is not enough.

Bad:

```go
return errors.New("missing webhook")
```

Better:

```go
return fmt.Errorf("missing webhook for event type %q", eventType)
```

Even better with wrapping:

```go
return fmt.Errorf("route webhook for event type %q: %w", eventType, ErrMissingWebhook)
```

This means:

```text
Main error: ErrMissingWebhook
Extra context: event type was "job"
```

---

## Why wrapping matters

Assume this error happens:

```text
route webhook for event type "job": missing webhook configuration
```

You can still check:

```go
if errors.Is(err, ErrMissingWebhook) {
	// handle missing webhook case
}
```

So wrapping gives you both:

```text
Human-readable context
+
Machine-checkable error type
```

---

# 6. Structured Logging with zerolog

`zerolog` is a fast structured logging library for Go.

Install:

```bash
go get github.com/rs/zerolog
```

Basic example:

```go
package main

import (
	"os"

	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	logger.Info().
		Str("event_type", "pr").
		Str("stage", "started").
		Str("status", "running").
		Msg("pipeline notification started")
}
```

Output looks like JSON:

```json
{
  "level": "info",
  "event_type": "pr",
  "stage": "started",
  "status": "running",
  "time": "2026-04-27T10:15:30+05:30",
  "message": "pipeline notification started"
}
```

## Python equivalent

Python basic logging:

```python
import logging

logging.info("pipeline notification started")
```

Python structured style:

```python
logger.info(
    "pipeline notification started",
    extra={
        "event_type": "pr",
        "stage": "started",
        "status": "running"
    }
)
```

In Go with zerolog, structured fields are very natural:

```go
logger.Info().
	Str("event_type", "pr").
	Str("stage", "started").
	Str("status", "running").
	Msg("pipeline notification started")
```

---

# 7. Why Plain Print Statements Are Not Enough

Using `fmt.Println`:

```go
fmt.Println("Slack notification failed")
```

This is easy for beginners, but weak for real systems.

Problem: it does not tell you enough.

```text
Slack notification failed
```

Which event failed?

```text
PR? CD? Job?
```

Which pipeline failed?

```text
deploy-prod? validate-pr? sync-job?
```

Which step failed?

```text
build? test? deploy? slack-notify?
```

Which webhook route was used?

```text
pr webhook? cd webhook? fallback webhook?
```

What was the status code?

```text
400? 403? 500?
```

Structured log is better:

```go
logger.Error().
	Str("event_type", event.EventType).
	Str("stage", event.Stage).
	Str("status", event.Status).
	Str("pipeline_name", event.PipelineName).
	Str("failed_step", event.FailedStep).
	Int("status_code", 500).
	Err(err).
	Msg("failed to send slack notification")
```

Now your logs contain searchable fields.

In Tekton debugging, this helps because you can search logs by:

```text
event_type=pr
pipeline_name=pr-validation
failed_step=unit-test
status=failed
```

---

# 8. How Failure Context Should Flow Through the Project

Your project should not only say:

```text
Something failed
```

It should carry context like:

```text
event_type
stage
status
repository
branch
commit_id
pipeline_name
pipeline_run_name
failed_step
error_message
route_key
webhook_name
retry_count
```

## Good failure flow

```text
main.go
  reads CLI flags

model.PipelineEvent
  stores event context

model.Validate()
  validates required fields

router.Resolve()
  decides webhook route

slack.SendMessage()
  sends HTTP request

if error happens:
  return wrapped error upward

main.go
  logs final error with event context
```

## ASCII Diagram

```text
CLI Flags
   |
   v
PipelineEvent
   |
   | validate error?
   v
Router
   |
   | route error?
   v
Slack Client
   |
   | HTTP error?
   v
Logger
   |
   v
Structured JSON logs for Tekton debugging
```

---

# 9. Pseudocode First for Logging + Error Return Flow

```text
START

read CLI flags

create PipelineEvent

create logger

log "notification started" with event fields

validate PipelineEvent
if validation fails:
    log validation error with event fields
    return error

resolve Slack webhook route
if route missing:
    log routing error with event fields
    return error

build Slack message

send Slack message
if send fails:
    log Slack send error with event fields and retry count
    return error

log "notification sent successfully"

END
```

---

# 10. Real Go Code Examples

## Example 1 — Custom Errors in Model Layer

File:

```text
pkg/notify/model/errors.go
```

```go
package model

import "errors"

var (
	ErrMissingEventType = errors.New("missing event type")
	ErrMissingStage     = errors.New("missing stage")
	ErrMissingStatus    = errors.New("missing status")
	ErrInvalidEvent     = errors.New("invalid pipeline event")
)
```

---

## Example 2 — PipelineEvent with Validation

File:

```text
pkg/notify/model/event.go
```

```go
package model

import "fmt"

type PipelineEvent struct {
	EventType       string
	Stage           string
	Status          string
	Repository      string
	Branch          string
	CommitID        string
	PipelineName    string
	PipelineRunName string
	FailedStep      string
	ErrorMessage    string
}

func (e PipelineEvent) Validate() error {
	if e.EventType == "" {
		return fmt.Errorf("validate pipeline event: %w", ErrMissingEventType)
	}

	if e.Stage == "" {
		return fmt.Errorf("validate pipeline event: %w", ErrMissingStage)
	}

	if e.Status == "" {
		return fmt.Errorf("validate pipeline event: %w", ErrMissingStatus)
	}

	return nil
}
```

## Python equivalent

```python
from dataclasses import dataclass

class MissingEventTypeError(Exception):
    pass

@dataclass
class PipelineEvent:
    event_type: str
    stage: str
    status: str

    def validate(self):
        if not self.event_type:
            raise MissingEventTypeError("missing event type")
```

## Key Go syntax differences

Go method:

```go
func (e PipelineEvent) Validate() error
```

Python method:

```python
def validate(self):
```

In Go:

```go
return nil
```

means:

```text
No error
```

In Python, you may simply return nothing:

```python
return
```

---

## Example 3 — Router Errors

File:

```text
pkg/notify/router/errors.go
```

```go
package router

import "errors"

var (
	ErrMissingWebhook = errors.New("missing webhook configuration")
	ErrUnknownRoute   = errors.New("unknown notification route")
)
```

File:

```text
pkg/notify/router/router.go
```

```go
package router

import (
	"fmt"

	"slack-integration/pkg/notify/model"
)

type Config struct {
	PRWebhookURL string
	CDWebhookURL string
}

type Router struct {
	config Config
}

func NewRouter(config Config) Router {
	return Router{
		config: config,
	}
}

func (r Router) ResolveWebhook(event model.PipelineEvent) (string, error) {
	switch event.EventType {
	case "pr":
		if r.config.PRWebhookURL == "" {
			return "", fmt.Errorf("resolve pr webhook: %w", ErrMissingWebhook)
		}
		return r.config.PRWebhookURL, nil

	case "cd":
		if r.config.CDWebhookURL == "" {
			return "", fmt.Errorf("resolve cd webhook: %w", ErrMissingWebhook)
		}
		return r.config.CDWebhookURL, nil

	case "job":
		// Fallback rule:
		// If job webhook is not available, use CD webhook.
		if r.config.CDWebhookURL == "" {
			return "", fmt.Errorf("resolve job fallback webhook: %w", ErrMissingWebhook)
		}
		return r.config.CDWebhookURL, nil

	default:
		return "", fmt.Errorf("event type %q: %w", event.EventType, ErrUnknownRoute)
	}
}
```

---

## Example 4 — Basic Slack Client with Wrapped Errors

File:

```text
pkg/notify/slack/client.go
```

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

type Message struct {
	Text string `json:"text"`
}

func NewClient(timeout time.Duration) Client {
	return Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c Client) SendMessage(webhookURL string, message Message) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal slack message: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create slack request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send slack request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack returned non-success status code: %d", resp.StatusCode)
	}

	return nil
}
```

## Python equivalent

```python
import requests

def send_message(webhook_url, message):
    try:
        response = requests.post(webhook_url, json=message, timeout=10)
        response.raise_for_status()
    except Exception as e:
        raise RuntimeError("send slack request failed") from e
```

## Important comparison

Python:

```python
raise RuntimeError("send slack request failed") from e
```

Go:

```go
return fmt.Errorf("send slack request: %w", err)
```

Both preserve the original error.

---

# 11. Basic Logger Package Design

In real projects, you should not create logger setup repeatedly everywhere.

Better:

```text
pkg/logger/logger.go
```

```go
package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

func New(env string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	level := zerolog.InfoLevel

	if strings.EqualFold(env, "dev") {
		level = zerolog.DebugLevel
	}

	zerolog.SetGlobalLevel(level)

	return zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", "slack-notifier").
		Str("env", env).
		Logger()
}
```

Usage:

```go
log := logger.New("dev")

log.Info().Msg("application started")
```

Output:

```json
{
  "level": "info",
  "service": "slack-notifier",
  "env": "dev",
  "time": "2026-04-27T10:15:30+05:30",
  "message": "application started"
}
```

## Why this package is useful

Instead of every file doing this:

```go
zerolog.New(os.Stdout)
```

You centralize logger setup in one place.

This gives consistency:

```text
Same timestamp format
Same service name
Same environment field
Same log level behavior
```

---

# 12. How to Log Event Type, Stage, Status, Pipeline Name, Failed Step

You can create a helper function.

File:

```text
pkg/logger/event.go
```

```go
package logger

import (
	"github.com/rs/zerolog"

	"slack-integration/pkg/notify/model"
)

func WithEvent(log zerolog.Logger, event model.PipelineEvent) zerolog.Context {
	return log.With().
		Str("event_type", event.EventType).
		Str("stage", event.Stage).
		Str("status", event.Status).
		Str("repository", event.Repository).
		Str("branch", event.Branch).
		Str("commit_id", event.CommitID).
		Str("pipeline_name", event.PipelineName).
		Str("pipeline_run_name", event.PipelineRunName).
		Str("failed_step", event.FailedStep)
}
```

Usage:

```go
eventLogger := logger.WithEvent(log, event).Logger()

eventLogger.Info().Msg("notification processing started")
```

Failure log:

```go
eventLogger.Error().
	Err(err).
	Str("error_message", event.ErrorMessage).
	Msg("notification processing failed")
```

---

## Example main.go Flow

File:

```text
cmd/slack-notifier/main.go
```

```go
package main

import (
	"errors"
	"flag"
	"os"
	"time"

	applogger "slack-integration/pkg/logger"
	"slack-integration/pkg/notify/model"
	"slack-integration/pkg/notify/router"
	"slack-integration/pkg/notify/slack"
)

func main() {
	eventType := flag.String("event-type", "", "event type: pr, cd, job")
	stage := flag.String("stage", "", "pipeline stage")
	status := flag.String("status", "", "pipeline status")
	pipelineName := flag.String("pipeline-name", "", "pipeline name")
	failedStep := flag.String("failed-step", "", "failed step")
	errorMessage := flag.String("error-message", "", "error message")
	env := flag.String("env", "dev", "environment")

	flag.Parse()

	log := applogger.New(*env)

	event := model.PipelineEvent{
		EventType:    *eventType,
		Stage:        *stage,
		Status:       *status,
		PipelineName: *pipelineName,
		FailedStep:   *failedStep,
		ErrorMessage: *errorMessage,
	}

	eventLogger := applogger.WithEvent(log, event).Logger()

	eventLogger.Info().Msg("notification processing started")

	if err := event.Validate(); err != nil {
		eventLogger.Error().
			Err(err).
			Msg("pipeline event validation failed")

		os.Exit(1)
	}

	rt := router.NewRouter(router.Config{
		PRWebhookURL: os.Getenv("PR_WEBHOOK_URL"),
		CDWebhookURL: os.Getenv("CD_WEBHOOK_URL"),
	})

	webhookURL, err := rt.ResolveWebhook(event)
	if err != nil {
		if errors.Is(err, router.ErrMissingWebhook) {
			eventLogger.Error().
				Err(err).
				Msg("webhook configuration missing")
		} else {
			eventLogger.Error().
				Err(err).
				Msg("failed to resolve webhook")
		}

		os.Exit(1)
	}

	client := slack.NewClient(10 * time.Second)

	message := slack.Message{
		Text: "Pipeline event: " + event.EventType + " | Status: " + event.Status,
	}

	if err := client.SendMessage(webhookURL, message); err != nil {
		eventLogger.Error().
			Err(err).
			Msg("failed to send slack notification")

		os.Exit(1)
	}

	eventLogger.Info().Msg("slack notification sent successfully")
}
```

---

# 13. Hands-On Tasks

## Task 1 — Add zerolog to your project

Run:

```bash
go get github.com/rs/zerolog
```

---

## Task 2 — Create logger package

Create:

```text
pkg/logger/logger.go
```

Add:

```go
package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func New(env string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	return zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", "slack-notifier").
		Str("env", env).
		Logger()
}
```

---

## Task 3 — Add event-based logging

Create:

```text
pkg/logger/event.go
```

Add helper:

```go
package logger

import (
	"github.com/rs/zerolog"

	"slack-integration/pkg/notify/model"
)

func WithEvent(log zerolog.Logger, event model.PipelineEvent) zerolog.Context {
	return log.With().
		Str("event_type", event.EventType).
		Str("stage", event.Stage).
		Str("status", event.Status).
		Str("pipeline_name", event.PipelineName).
		Str("failed_step", event.FailedStep)
}
```

---

## Task 4 — Log validation failure

In `main.go`:

```go
if err := event.Validate(); err != nil {
	eventLogger.Error().
		Err(err).
		Msg("pipeline event validation failed")

	os.Exit(1)
}
```

---

## Task 5 — Log Slack send failure

```go
if err := client.SendMessage(webhookURL, message); err != nil {
	eventLogger.Error().
		Err(err).
		Msg("failed to send slack notification")

	os.Exit(1)
}
```

---

# 14. Expected Output

## Success command

```bash
go run cmd/slack-notifier/main.go \
  --env dev \
  --event-type pr \
  --stage started \
  --status running \
  --pipeline-name pr-validation
```

Expected log:

```json
{
  "level": "info",
  "service": "slack-notifier",
  "env": "dev",
  "event_type": "pr",
  "stage": "started",
  "status": "running",
  "pipeline_name": "pr-validation",
  "time": "2026-04-27T10:15:30+05:30",
  "message": "notification processing started"
}
```

---

## Failure command

```bash
go run cmd/slack-notifier/main.go \
  --env dev \
  --event-type job \
  --stage failure \
  --status failed \
  --pipeline-name sync-job \
  --failed-step validate-input \
  --error-message "missing required config"
```

If webhook is missing:

```json
{
  "level": "error",
  "service": "slack-notifier",
  "env": "dev",
  "event_type": "job",
  "stage": "failure",
  "status": "failed",
  "pipeline_name": "sync-job",
  "failed_step": "validate-input",
  "error": "resolve job fallback webhook: missing webhook configuration",
  "time": "2026-04-27T10:15:30+05:30",
  "message": "webhook configuration missing"
}
```

This log is useful because it tells you:

```text
event_type = job
stage = failure
status = failed
pipeline_name = sync-job
failed_step = validate-input
actual error = missing webhook configuration
```

---

# 15. Common Mistakes

## Mistake 1 — Logging only the error

Weak:

```go
log.Error().Err(err).Msg("failed")
```

Better:

```go
log.Error().
	Err(err).
	Str("event_type", event.EventType).
	Str("pipeline_name", event.PipelineName).
	Str("failed_step", event.FailedStep).
	Msg("failed to send slack notification")
```

---

## Mistake 2 — Swallowing errors

Bad:

```go
if err != nil {
	fmt.Println(err)
}
```

The program continues even after failure.

Better:

```go
if err != nil {
	return err
}
```

Or in `main.go`:

```go
if err != nil {
	log.Error().Err(err).Msg("operation failed")
	os.Exit(1)
}
```

---

## Mistake 3 — Not wrapping errors

Weak:

```go
return err
```

Better:

```go
return fmt.Errorf("send slack request: %w", err)
```

This tells you where the error happened.

---

## Mistake 4 — Logging secrets

Never log:

```text
Slack webhook URL
API tokens
Passwords
Authorization headers
Private keys
```

Bad:

```go
log.Info().Str("webhook_url", webhookURL).Msg("using webhook")
```

Better:

```go
log.Info().Str("route", "pr").Msg("using resolved webhook route")
```

---

## Mistake 5 — Using inconsistent field names

Avoid mixing:

```text
eventType
event_type
event-type
event
```

Use one style:

```text
event_type
pipeline_name
failed_step
commit_id
```

This makes log searching easier.

---

# 16. Debugging Tips

## Tip 1 — Start from the first error log

In Tekton logs, check the first meaningful error.

```bash
kubectl logs -l eventlistener=pr-listener -f -n slack-integration-dev
```

Look for:

```json
"level":"error"
```

---

## Tip 2 — Search by pipeline name

If logs are large, search for:

```text
pipeline_name
pipeline_run_name
event_type
failed_step
```

Example:

```bash
kubectl logs pod-name -n slack-integration-dev | grep "pipeline_name"
```

---

## Tip 3 — Use event context everywhere

If the Slack send fails, you should know:

```text
Which event?
Which status?
Which pipeline?
Which step?
Which route?
```

---

## Tip 4 — Separate user error from system error

User/config error:

```text
missing webhook
invalid event type
missing status
```

System error:

```text
network timeout
Slack returned 500
JSON marshal failed
```

This helps decide:

```text
Should we fix config?
Should we retry?
Should we alert?
Should we fail the pipeline?
```

---

## Tip 5 — Keep logs readable in dev

For development, you can use console writer:

```go
output := zerolog.ConsoleWriter{
	Out:        os.Stdout,
	TimeFormat: time.RFC3339,
}

log := zerolog.New(output).With().Timestamp().Logger()
```

This gives prettier local logs.

For production/Tekton, JSON logs are usually better.

---

# 17. DSA Topic — Linked List Basics

A linked list is a chain of nodes.

Each node has:

```text
data + pointer to next node
```

Example:

```text
10 → 20 → 30 → nil
```

Each node knows only the next node.

---

## Python mental model

Python class:

```python
class Node:
    def __init__(self, value):
        self.value = value
        self.next = None
```

Go version:

```go
type Node struct {
	Value int
	Next  *Node
}
```

Important Go difference:

```go
Next *Node
```

means:

```text
Next is a pointer to another Node
```

In Python, almost everything is reference-like by default.

In Go, you explicitly use `*Node` for pointer.

---

## Basic Linked List in Go

```go
package main

import "fmt"

type Node struct {
	Value int
	Next  *Node
}

func main() {
	first := &Node{Value: 10}
	second := &Node{Value: 20}
	third := &Node{Value: 30}

	first.Next = second
	second.Next = third

	current := first

	for current != nil {
		fmt.Println(current.Value)
		current = current.Next
	}
}
```

Output:

```text
10
20
30
```

---

## Linked List Diagram

```text
first
  |
  v
+-------+------+     +-------+------+     +-------+------+
|  10   | next | --> |  20   | next | --> |  30   | nil  |
+-------+------+     +-------+------+     +-------+------+
```

---

# 18. DSA Practice Problem in Go

## Problem: Count Nodes in a Linked List

Given the head of a linked list, count how many nodes exist.

Example:

```text
10 → 20 → 30 → nil
```

Answer:

```text
3
```

---

## Pseudocode

```text
count = 0
current = head

while current is not nil:
    count = count + 1
    current = current.next

return count
```

---

## Go Solution

```go
package main

import "fmt"

type Node struct {
	Value int
	Next  *Node
}

func CountNodes(head *Node) int {
	count := 0
	current := head

	for current != nil {
		count++
		current = current.Next
	}

	return count
}

func main() {
	head := &Node{Value: 10}
	head.Next = &Node{Value: 20}
	head.Next.Next = &Node{Value: 30}

	result := CountNodes(head)

	fmt.Println("Total nodes:", result)
}
```

Output:

```text
Total nodes: 3
```

---

## Python equivalent

```python
class Node:
    def __init__(self, value):
        self.value = value
        self.next = None

def count_nodes(head):
    count = 0
    current = head

    while current is not None:
        count += 1
        current = current.next

    return count
```

---

# 19. Module-Based Practice Task — Build a Small Log Parser

## Goal

Build a small Go program that extracts:

```text
level
message
timestamp
```

from a structured JSON log line.

Example log:

```json
{"level":"error","time":"2026-04-27T10:15:30+05:30","message":"failed to send slack notification"}
```

Expected output:

```text
Level: error
Time: 2026-04-27T10:15:30+05:30
Message: failed to send slack notification
```

---

## Suggested Folder Structure

```text
log-parser/
  go.mod
  main.go
  parser/
    parser.go
```

---

## parser/parser.go

```go
package parser

import (
	"encoding/json"
	"fmt"
)

type LogEntry struct {
	Level   string `json:"level"`
	Time    string `json:"time"`
	Message string `json:"message"`
}

func ParseLogLine(line string) (LogEntry, error) {
	var entry LogEntry

	if line == "" {
		return entry, fmt.Errorf("log line is empty")
	}

	err := json.Unmarshal([]byte(line), &entry)
	if err != nil {
		return entry, fmt.Errorf("parse log line json: %w", err)
	}

	return entry, nil
}
```

---

## main.go

```go
package main

import (
	"fmt"

	"log-parser/parser"
)

func main() {
	line := `{"level":"error","time":"2026-04-27T10:15:30+05:30","message":"failed to send slack notification"}`

	entry, err := parser.ParseLogLine(line)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Level:", entry.Level)
	fmt.Println("Time:", entry.Time)
	fmt.Println("Message:", entry.Message)
}
```

---

## Run

```bash
go run main.go
```

Expected output:

```text
Level: error
Time: 2026-04-27T10:15:30+05:30
Message: failed to send slack notification
```

---

## Python equivalent

```python
import json

line = '{"level":"error","time":"2026-04-27T10:15:30+05:30","message":"failed to send slack notification"}'

entry = json.loads(line)

print("Level:", entry["level"])
print("Time:", entry["time"])
print("Message:", entry["message"])
```

## Go difference

Python:

```python
entry["level"]
```

Go:

```go
entry.Level
```

Because in Go, JSON is decoded into a typed struct.

---

# 20. Revision Checkpoint

You are ready for Day 5 if you can answer these:

1. What does `err != nil` mean?
2. Why does Go return errors instead of throwing exceptions?
3. What is the difference between `errors.New()` and `fmt.Errorf()`?
4. Why is `%w` used in `fmt.Errorf()`?
5. What does `errors.Is()` do?
6. Why is structured logging better than `fmt.Println()`?
7. What fields should we log for a pipeline failure?
8. Why should we not log webhook URLs?
9. How does logging help debug Tekton failures?
10. What is a linked list node?
11. Why does linked list use `*Node` in Go?
12. How does a log parser convert JSON text into a Go struct?

---

# 21. Homework

## Homework 1 — Add Validation Errors

Add these errors in your model package:

```go
ErrMissingPipelineName
ErrMissingRepository
ErrMissingBranch
```

Use them inside `Validate()`.

---

## Homework 2 — Add Structured Logs

Add structured logs for:

```text
notification started
event validation failed
webhook resolved
slack send started
slack send failed
slack send success
```

---

## Homework 3 — Add Route Field

When router resolves webhook, log only route name:

```text
route=pr
route=cd
route=job_fallback_cd
```

Do not log actual webhook URL.

---

## Homework 4 — Improve Log Parser

Enhance the log parser to also extract:

```text
event_type
pipeline_name
failed_step
```

Example input:

```json
{"level":"error","time":"2026-04-27T10:15:30+05:30","event_type":"job","pipeline_name":"sync-job","failed_step":"validate-input","message":"failed to send slack notification"}
```

Expected output:

```text
Level: error
Time: 2026-04-27T10:15:30+05:30
Event Type: job
Pipeline Name: sync-job
Failed Step: validate-input
Message: failed to send slack notification
```

---

## Homework 5 — Linked List Practice

Create a linked list:

```text
5 → 10 → 15 → 20
```

Write a function:

```go
func SumNodes(head *Node) int
```

Expected answer:

```text
50
```

---

# Simple Day 5 Summary

Go does not hide failures. It makes you handle them directly using `error`.

In Python, you usually think:

```text
try → except
```

In Go, think:

```text
call function → check err → return/log/fallback
```

For your Slack notification project, good error handling and structured logging make the system production-friendly. When Tekton pipeline notification fails, your logs should clearly answer:

```text
What failed?
Where did it fail?
Which pipeline was affected?
Which step failed?
Which event type was being processed?
Was it validation, routing, or Slack delivery?
```

That is the real purpose of Day 5.
