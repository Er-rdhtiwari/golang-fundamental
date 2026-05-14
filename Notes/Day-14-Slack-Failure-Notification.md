## 1. Day 14 learning goals

Today you are upgrading a simple Slack failure notification from:

> Build failed.

to something closer to:

> Pipeline failed in task `build-image`, step `docker-build`.
> Error: `failed to solve: process "/bin/sh -c go test ./..." did not complete successfully`
> Trace: last useful 20 lines, safely trimmed and secrets masked.

By the end of Day 14, you should understand:

1. How Tekton task failures appear in task status and container logs.
2. How to identify the failed step.
3. How to extract a useful error message from logs.
4. How to keep only a short trace instead of dumping huge logs.
5. How shell scripts and Go code can cooperate.
6. How to design a structured failure event for Slack.
7. How to avoid leaking secrets in notifications.
8. How to test the formatter logic.
9. Dynamic programming basics in Go.

---

## 2. Quick revision of Days 1 to 13

Here is the mental journey so far:

| Day | Concept                     | Why it mattered                              |
| --- | --------------------------- | -------------------------------------------- |
| 1   | Slack webhook basics        | Send a message from code to Slack            |
| 2   | Go project structure        | Keep code organized into packages            |
| 3   | Environment variables       | Configure webhook URL, namespace, channel    |
| 4   | JSON payloads               | Slack messages are structured JSON           |
| 5   | HTTP client in Go           | Send POST requests reliably                  |
| 6   | Error handling              | Do not ignore failed Slack sends             |
| 7   | Shell scripting basics      | Glue CI/CD tools together                    |
| 8   | Tekton basics               | Understand PipelineRun, TaskRun, steps       |
| 9   | Event-driven thinking       | Convert CI/CD events into notifications      |
| 10  | Message formatting          | Make Slack output readable                   |
| 11  | Config-driven behavior      | Avoid hardcoding project-specific values     |
| 12  | Testing formatters          | Test output without needing real Slack       |
| 13  | Cleaner project enhancement | Separate collector, model, formatter, sender |

Day 14 builds on all of this.

Earlier, your project could say:

```text
Pipeline failed.
```

Today, it should say:

```text
Pipeline failed at task build-image, step docker-build.
Reason: image build failed.
Trace:
  go test ./...
  FAIL github.com/example/app/pkg/api
  Error: expected 200, got 500
```

That is a real debugging upgrade.

---

## 3. Why normal “build failed” notifications are not enough

A weak notification says:

```text
Build failed.
```

That tells the team something broke, but not where or why.

The developer now has to:

1. Open Tekton dashboard.
2. Find the PipelineRun.
3. Find the failed TaskRun.
4. Find the failed step.
5. Open logs.
6. Scroll through hundreds or thousands of lines.
7. Search for the actual error.

That wastes time.

A useful notification should reduce the first debugging step.

It should answer:

```text
What failed?
Where did it fail?
What was the likely reason?
Where should I look next?
```

In real CI/CD systems, speed matters. A good failure notification can save several minutes every time a pipeline breaks.

---

## 4. What useful failure context looks like

A useful failure context usually contains:

```text
PipelineRun: user-service-pr-142
TaskRun: user-service-pr-142-build-image
Task: build-image
Step: docker-build
Exit code: 1
Reason: Error
Message: failed to build image
Short trace:
  #12 RUN go test ./...
  FAIL ./internal/api
  expected status 200, got 500
```

For Slack, you do not want the full log. You want the smallest useful summary.

A good failure message has:

1. **Identity**: Which pipeline, task, and step failed?
2. **Cause**: What error message was found?
3. **Trace**: What nearby log lines explain the error?
4. **Action hint**: Where should the developer look next?
5. **Safety**: No passwords, tokens, private keys, or secrets.

---

## 5. Failed step, error message, error trace, and log snippet

Let’s define the terms clearly.

### Failed step

In Tekton, a `TaskRun` contains one or more steps.

Example task:

```yaml
steps:
  - name: install
    image: golang:1.22
    script: go mod download

  - name: test
    image: golang:1.22
    script: go test ./...

  - name: build
    image: golang:1.22
    script: go build ./cmd/api
```

If `go test ./...` fails, then the failed step is:

```text
test
```

Tekton stores step status information in the `TaskRun` status. The corresponding Kubernetes pod also stores container logs for each step.

---

### Error message

The error message is usually the most important single line.

Example log:

```text
running tests
--- FAIL: TestCreateUser
    expected status 201, got 500
FAIL
exit status 1
```

A useful error message might be:

```text
expected status 201, got 500
```

or:

```text
FAIL: TestCreateUser
```

The collector should search for common error words:

```text
error
failed
fatal
exception
panic
denied
timeout
exit status
```

This is basic log parsing.

---

### Error trace

A trace is a short group of lines around the failure.

Example:

```text
=== RUN   TestCreateUser
--- FAIL: TestCreateUser
    user_test.go:42: expected status 201, got 500
FAIL
FAIL github.com/example/user-service/internal/api 0.234s
```

This is better than one line because it gives nearby context.

---

### Log snippet

A log snippet is a trimmed part of logs.

The full log may be 5,000 lines.

The snippet might be the last 20 useful lines:

```text
#10 running go test ./...
--- FAIL: TestCreateUser
user_test.go:42: expected status 201, got 500
FAIL
```

Important rule:

```text
Slack should receive a useful summary, not a full log dump.
```

---

## 6. How shell scripts and Go code can work together here

For a beginner project, a nice split is:

```text
Shell script:
  Talks to kubectl/Tekton
  Gets TaskRun status
  Gets step logs
  Outputs structured JSON

Go code:
  Reads structured failure JSON
  Sanitizes text
  Truncates trace
  Formats Slack message
  Sends Slack webhook
```

This is practical because shell is good at calling command-line tools, while Go is better for structured logic, testing, formatting, and HTTP.

Python comparison:

```text
Python:
  dict, json.loads(), requests.post(), list slicing

Go:
  struct, json.Unmarshal(), http.Client.Do(), slices
```

In Python, you might pass dictionaries around:

```python
failure = {
    "task_run": "build-taskrun",
    "failed_step": "test",
    "error_message": "go test failed",
}
```

In Go, you usually define a struct:

```go
type FailureContext struct {
    TaskRun      string `json:"task_run"`
    FailedStep   string `json:"failed_step"`
    ErrorMessage string `json:"error_message"`
}
```

Go is more explicit. That can feel slower at first, but it makes larger CI/CD tools safer and easier to maintain.

---

## 7. Full failure capture flow in ASCII

```text
                ┌──────────────────────┐
                │ Tekton PipelineRun    │
                └──────────┬───────────┘
                           │
                           ▼
                ┌──────────────────────┐
                │ Tekton TaskRun fails  │
                └──────────┬───────────┘
                           │
                           ▼
                ┌──────────────────────┐
                │ Find failed step      │
                │ from TaskRun status   │
                └──────────┬───────────┘
                           │
                           ▼
                ┌──────────────────────┐
                │ Get logs for failed   │
                │ step container        │
                └──────────┬───────────┘
                           │
                           ▼
                ┌──────────────────────┐
                │ Parse logs for error  │
                │ words and trace       │
                └──────────┬───────────┘
                           │
                           ▼
                ┌──────────────────────┐
                │ Mask secrets          │
                │ token=****            │
                └──────────┬───────────┘
                           │
                           ▼
                ┌──────────────────────┐
                │ Truncate noisy trace  │
                │ max lines/chars       │
                └──────────┬───────────┘
                           │
                           ▼
                ┌──────────────────────┐
                │ Build structured      │
                │ FailureContext event  │
                └──────────┬───────────┘
                           │
                           ▼
                ┌──────────────────────┐
                │ Format Slack message  │
                └──────────┬───────────┘
                           │
                           ▼
                ┌──────────────────────┐
                │ Send to Slack webhook │
                └──────────────────────┘
```

The big idea:

```text
Raw Tekton logs → parsed summary → safe structured event → Slack message
```

---

## 8. Pseudocode first for error trace collection and Slack formatting

### Error trace collection pseudocode

```text
function collectFailure(taskRunName, namespace):
    taskRunJson = kubectl get taskrun taskRunName as JSON

    failedStep = find step where exitCode != 0

    if failedStep is empty:
        failedStep = "unknown"

    podName = taskRunJson.status.podName

    logs = kubectl logs podName container step-failedStep

    sanitizedLogs = maskSecrets(logs)

    errorMessage = findBestErrorLine(sanitizedLogs)

    trace = getLastUsefulLinesAroundError(sanitizedLogs)

    trace = truncate(trace, maxLines=20, maxChars=3000)

    return FailureContext{
        taskRun: taskRunName,
        namespace: namespace,
        failedStep: failedStep,
        errorMessage: errorMessage,
        trace: trace
    }
```

### Slack formatting pseudocode

```text
function formatSlackMessage(failure):
    title = "Tekton task failed"

    fields:
        PipelineRun
        TaskRun
        Failed step
        Exit code
        Error message

    traceBlock:
        code block containing short trace

    if trace was truncated:
        add note: "Trace trimmed for readability"

    return SlackPayload
```

### Secret masking pseudocode

```text
function maskSecrets(text):
    replace "password=anything" with "password=****"
    replace "token=anything" with "token=****"
    replace "Authorization: Bearer anything" with "Authorization: Bearer ****"
    replace private key blocks with "[REDACTED PRIVATE KEY]"
    return text
```

Never trust logs blindly. CI logs often contain accidental secrets.

---

## 9. Real shell and Go code examples

### Example shell collector

This script collects basic failure information from a Tekton `TaskRun`.

File:

```text
scripts/collect-failure.sh
```

```bash
#!/usr/bin/env bash
set -euo pipefail

TASKRUN_NAME="${1:?Usage: collect-failure.sh <taskrun-name> [namespace]}"
NAMESPACE="${2:-default}"

taskrun_json="$(kubectl -n "$NAMESPACE" get taskrun "$TASKRUN_NAME" -o json)"

pod_name="$(echo "$taskrun_json" | jq -r '.status.podName // empty')"

failed_step="$(
  echo "$taskrun_json" |
    jq -r '
      .status.steps[]?
      | select((.terminated.exitCode // 0) != 0)
      | .name
    ' |
    head -n 1
)"

exit_code="$(
  echo "$taskrun_json" |
    jq -r '
      .status.steps[]?
      | select((.terminated.exitCode // 0) != 0)
      | .terminated.exitCode
    ' |
    head -n 1
)"

reason="$(
  echo "$taskrun_json" |
    jq -r '
      .status.conditions[]?
      | select(.type == "Succeeded")
      | .reason // "Unknown"
    ' |
    head -n 1
)"

message="$(
  echo "$taskrun_json" |
    jq -r '
      .status.conditions[]?
      | select(.type == "Succeeded")
      | .message // "No condition message"
    ' |
    head -n 1
)"

if [[ -z "$failed_step" ]]; then
  failed_step="unknown"
fi

if [[ -z "$exit_code" ]]; then
  exit_code="unknown"
fi

if [[ -z "$pod_name" ]]; then
  echo "Could not find pod name for TaskRun: $TASKRUN_NAME" >&2
  exit 1
fi

# Tekton step containers are usually named step-<stepName>.
container_name="step-${failed_step}"

raw_logs="$(
  kubectl -n "$NAMESPACE" logs "$pod_name" -c "$container_name" --tail=120 2>/dev/null || true
)"

# Basic secret masking.
safe_logs="$(
  printf "%s\n" "$raw_logs" |
    sed -E 's/(password|passwd|token|secret|api[_-]?key)=([^ ]+)/\1=****/Ig' |
    sed -E 's/(Authorization: Bearer )[A-Za-z0-9._~+\/=-]+/\1****/Ig'
)"

error_line="$(
  printf "%s\n" "$safe_logs" |
    grep -Ei 'error|failed|fatal|exception|panic|denied|timeout|exit status' |
    tail -n 1 ||
    true
)"

if [[ -z "$error_line" ]]; then
  error_line="$message"
fi

trace="$(
  printf "%s\n" "$safe_logs" |
    tail -n 30
)"

jq -n \
  --arg namespace "$NAMESPACE" \
  --arg task_run "$TASKRUN_NAME" \
  --arg pod "$pod_name" \
  --arg failed_step "$failed_step" \
  --arg exit_code "$exit_code" \
  --arg reason "$reason" \
  --arg error_message "$error_line" \
  --arg trace "$trace" \
  '{
    namespace: $namespace,
    task_run: $task_run,
    pod: $pod,
    failed_step: $failed_step,
    exit_code: $exit_code,
    reason: $reason,
    error_message: $error_message,
    trace: $trace
  }'
```

Run it like this:

```bash
chmod +x scripts/collect-failure.sh

./scripts/collect-failure.sh user-service-build-run default
```

Output:

```json
{
  "namespace": "default",
  "task_run": "user-service-build-run",
  "pod": "user-service-build-run-pod",
  "failed_step": "test",
  "exit_code": "1",
  "reason": "Failed",
  "error_message": "user_test.go:42: expected status 201, got 500",
  "trace": "--- FAIL: TestCreateUser\nuser_test.go:42: expected status 201, got 500\nFAIL"
}
```

Beginner note: this script is intentionally simple. In production, you would usually use the Kubernetes or Tekton API from Go instead of shelling out to `kubectl`.

---

### Go model and formatter

File:

```text
internal/model/failure.go
```

```go
package model

type FailureContext struct {
	Namespace    string `json:"namespace"`
	PipelineRun string `json:"pipeline_run,omitempty"`
	TaskRun     string `json:"task_run"`
	Pod         string `json:"pod"`
	FailedStep  string `json:"failed_step"`
	ExitCode    string `json:"exit_code"`
	Reason      string `json:"reason"`
	ErrorMessage string `json:"error_message"`
	Trace       string `json:"trace"`
	TraceTrimmed bool  `json:"trace_trimmed"`
}
```

Python comparison:

```python
failure = {
    "namespace": "default",
    "task_run": "build-run",
    "failed_step": "test",
}
```

Go uses a `struct` instead of a free-form dictionary. The backtick tags:

```go
`json:"task_run"`
```

tell Go how the field should appear in JSON.

---

### Go sanitizer and truncator

File:

```text
internal/failure/text.go
```

```go
package failure

import (
	"regexp"
	"strings"
)

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(password|passwd|token|secret|api[_-]?key)=([^ \n]+)`),
	regexp.MustCompile(`(?i)(Authorization:\s*Bearer\s+)[A-Za-z0-9._~+/=-]+`),
	regexp.MustCompile(`(?s)-----BEGIN [A-Z ]*PRIVATE KEY-----.*?-----END [A-Z ]*PRIVATE KEY-----`),
}

func MaskSecrets(input string) string {
	output := input

	for _, pattern := range secretPatterns {
		output = pattern.ReplaceAllStringFunc(output, func(match string) string {
			if strings.Contains(strings.ToUpper(match), "PRIVATE KEY") {
				return "[REDACTED PRIVATE KEY]"
			}

			if strings.HasPrefix(strings.ToLower(match), "authorization") {
				return regexp.MustCompile(`(?i)(Authorization:\s*Bearer\s+).*`).
					ReplaceAllString(match, `${1}****`)
			}

			return regexp.MustCompile(`(?i)^([^=]+)=.*$`).
				ReplaceAllString(match, `${1}=****`)
		})
	}

	return output
}

func TruncateTrace(trace string, maxLines int, maxChars int) (string, bool) {
	lines := strings.Split(trace, "\n")
	trimmed := false

	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
		trimmed = true
	}

	result := strings.Join(lines, "\n")

	if len(result) > maxChars {
		result = result[len(result)-maxChars:]
		trimmed = true
	}

	return strings.TrimSpace(result), trimmed
}

func FindErrorMessage(logText string) string {
	lines := strings.Split(logText, "\n")

	keywords := []string{
		"error",
		"failed",
		"fatal",
		"exception",
		"panic",
		"denied",
		"timeout",
		"exit status",
	}

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		lower := strings.ToLower(line)

		for _, keyword := range keywords {
			if strings.Contains(lower, keyword) {
				return line
			}
		}
	}

	return "No clear error line found"
}
```

Python comparison:

```python
lines = trace.split("\n")
last_20 = lines[-20:]
```

Go equivalent:

```go
lines := strings.Split(trace, "\n")
last20 := lines[len(lines)-20:]
```

But Go requires you to check lengths carefully. Python is more forgiving with slices. Go prefers explicit safety.

---

### Slack formatter

File:

```text
internal/slack/formatter.go
```

````go
package slack

import (
	"fmt"
	"strings"

	"slack-integration/internal/model"
)

type SlackPayload struct {
	Text   string        `json:"text"`
	Blocks []SlackBlock `json:"blocks"`
}

type SlackBlock struct {
	Type string       `json:"type"`
	Text *SlackText   `json:"text,omitempty"`
}

type SlackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func FormatFailureMessage(f model.FailureContext) SlackPayload {
	title := fmt.Sprintf(":x: Tekton task failed: `%s`", f.TaskRun)

	summary := fmt.Sprintf(
		"*Namespace:* `%s`\n*TaskRun:* `%s`\n*Failed step:* `%s`\n*Exit code:* `%s`\n*Reason:* `%s`\n*Error:* `%s`",
		emptyAsUnknown(f.Namespace),
		emptyAsUnknown(f.TaskRun),
		emptyAsUnknown(f.FailedStep),
		emptyAsUnknown(f.ExitCode),
		emptyAsUnknown(f.Reason),
		emptyAsUnknown(f.ErrorMessage),
	)

	if f.PipelineRun != "" {
		summary = fmt.Sprintf("*PipelineRun:* `%s`\n%s", f.PipelineRun, summary)
	}

	trace := strings.TrimSpace(f.Trace)
	if trace == "" {
		trace = "No trace available"
	}

	traceBlock := fmt.Sprintf("*Short trace:*\n```%s```", trace)

	if f.TraceTrimmed {
		traceBlock += "\n_Trace was trimmed to keep this Slack message readable._"
	}

	return SlackPayload{
		Text: fmt.Sprintf("Tekton task failed: %s", f.TaskRun),
		Blocks: []SlackBlock{
			{
				Type: "section",
				Text: &SlackText{
					Type: "mrkdwn",
					Text: title,
				},
			},
			{
				Type: "section",
				Text: &SlackText{
					Type: "mrkdwn",
					Text: summary,
				},
			},
			{
				Type: "section",
				Text: &SlackText{
					Type: "mrkdwn",
					Text: traceBlock,
				},
			},
		},
	}
}

func emptyAsUnknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return value
}
````

Go syntax notes:

```go
func FormatFailureMessage(f model.FailureContext) SlackPayload
```

means:

```text
function name: FormatFailureMessage
input: f of type model.FailureContext
returns: SlackPayload
```

Python equivalent:

```python
def format_failure_message(f: FailureContext) -> SlackPayload:
    ...
```

In Go, public functions start with a capital letter:

```go
FormatFailureMessage
```

Private helper functions usually start lowercase:

```go
emptyAsUnknown
```

That is a major Go convention shift from Python.

---

### Main program reading JSON from stdin

File:

```text
cmd/slack-notifier/main.go
```

```go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"slack-integration/internal/failure"
	"slack-integration/internal/model"
	"slack-integration/internal/slack"
)

func main() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read stdin: %v\n", err)
		os.Exit(1)
	}

	var ctx model.FailureContext
	if err := json.Unmarshal(input, &ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse failure JSON: %v\n", err)
		os.Exit(1)
	}

	ctx.Trace = failure.MaskSecrets(ctx.Trace)
	ctx.ErrorMessage = failure.MaskSecrets(ctx.ErrorMessage)

	ctx.Trace, ctx.TraceTrimmed = failure.TruncateTrace(ctx.Trace, 20, 3000)

	if ctx.ErrorMessage == "" || ctx.ErrorMessage == "No clear error line found" {
		ctx.ErrorMessage = failure.FindErrorMessage(ctx.Trace)
	}

	payload := slack.FormatFailureMessage(ctx)

	output, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal Slack payload: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}
```

Run it:

```bash
./scripts/collect-failure.sh user-service-build-run default \
  | go run ./cmd/slack-notifier
```

At this stage, the Go program prints the Slack payload. Later, you can connect it to your real Slack sender.

---

## 10. How to extend the event model

Before Day 14, your event might have looked like this:

```go
type PipelineEvent struct {
	PipelineRun string `json:"pipeline_run"`
	Status      string `json:"status"`
}
```

That is too small for rich failure notifications.

Extend it like this:

```go
type PipelineEvent struct {
	PipelineRun string          `json:"pipeline_run"`
	Namespace   string          `json:"namespace"`
	Status      string          `json:"status"`
	Failure     *FailureContext `json:"failure,omitempty"`
}
```

The pointer is important:

```go
Failure *FailureContext
```

It means failure details are optional.

Python comparison:

```python
event = {
    "pipeline_run": "user-service-pr-142",
    "status": "failed",
    "failure": None,
}
```

In Go:

```go
Failure: nil
```

is similar to Python:

```python
"failure": None
```

Example successful event:

```json
{
  "pipeline_run": "user-service-pr-142",
  "namespace": "default",
  "status": "Succeeded"
}
```

Example failed event:

```json
{
  "pipeline_run": "user-service-pr-142",
  "namespace": "default",
  "status": "Failed",
  "failure": {
    "task_run": "user-service-pr-142-build-image",
    "failed_step": "docker-build",
    "exit_code": "1",
    "error_message": "failed to solve image build",
    "trace": "short trace here"
  }
}
```

This is structured failure message design.

Do not force Slack formatting too early. Keep raw event data separate from display formatting.

Good separation:

```text
collector → model → formatter → sender
```

Bad separation:

```text
collector directly builds final Slack text everywhere
```

---

## 11. How to test the formatter

Testing the formatter is easier than testing real Slack.

You do not need to send a Slack message in tests.

You only check:

```text
Given this FailureContext,
does the Slack payload contain task name, step name, error, and trace?
```

File:

```text
internal/slack/formatter_test.go
```

```go
package slack

import (
	"strings"
	"testing"

	"slack-integration/internal/model"
)

func TestFormatFailureMessageIncludesFailureDetails(t *testing.T) {
	failure := model.FailureContext{
		Namespace:    "default",
		PipelineRun: "user-service-pr-142",
		TaskRun:     "user-service-pr-142-build-image",
		FailedStep:  "docker-build",
		ExitCode:    "1",
		Reason:      "Failed",
		ErrorMessage: "go test failed",
		Trace:       "FAIL github.com/example/user-service\nexit status 1",
		TraceTrimmed: true,
	}

	payload := FormatFailureMessage(failure)

	combined := payload.Text
	for _, block := range payload.Blocks {
		if block.Text != nil {
			combined += "\n" + block.Text.Text
		}
	}

	expectedParts := []string{
		"user-service-pr-142",
		"user-service-pr-142-build-image",
		"docker-build",
		"go test failed",
		"exit status 1",
		"Trace was trimmed",
	}

	for _, part := range expectedParts {
		if !strings.Contains(combined, part) {
			t.Fatalf("expected formatted message to contain %q, got:\n%s", part, combined)
		}
	}
}
```

Run:

```bash
go test ./...
```

Python comparison:

```python
assert "docker-build" in message
```

Go equivalent:

```go
if !strings.Contains(message, "docker-build") {
    t.Fatalf("expected message to contain docker-build")
}
```

Go tests are more verbose than Python tests, but they are very explicit.

---

## 12. Hands-on tasks

### Task 1: Create the failure model

Create:

```text
internal/model/failure.go
```

Add:

```go
type FailureContext struct {
	Namespace    string `json:"namespace"`
	TaskRun      string `json:"task_run"`
	Pod          string `json:"pod"`
	FailedStep   string `json:"failed_step"`
	ExitCode     string `json:"exit_code"`
	Reason       string `json:"reason"`
	ErrorMessage string `json:"error_message"`
	Trace        string `json:"trace"`
	TraceTrimmed bool   `json:"trace_trimmed"`
}
```

---

### Task 2: Add text utilities

Create:

```text
internal/failure/text.go
```

Implement:

```go
MaskSecrets()
TruncateTrace()
FindErrorMessage()
```

---

### Task 3: Add Slack failure formatter

Create:

```text
internal/slack/formatter.go
```

Implement:

```go
FormatFailureMessage()
```

---

### Task 4: Add shell collector

Create:

```text
scripts/collect-failure.sh
```

Use:

```bash
kubectl get taskrun
kubectl logs
jq
grep
tail
sed
```

---

### Task 5: Wire shell output into Go

Run:

```bash
./scripts/collect-failure.sh <taskrun-name> <namespace> \
  | go run ./cmd/slack-notifier
```

---

## 13. Expected output

Given this input:

```json
{
  "namespace": "default",
  "pipeline_run": "user-service-pr-142",
  "task_run": "user-service-pr-142-build-image",
  "pod": "user-service-pr-142-build-image-pod",
  "failed_step": "docker-build",
  "exit_code": "1",
  "reason": "Failed",
  "error_message": "failed to solve image build",
  "trace": "#12 RUN go test ./...\n--- FAIL: TestCreateUser\nexpected status 201, got 500\nFAIL"
}
```

Expected Slack-style message:

```text
:x: Tekton task failed: user-service-pr-142-build-image

PipelineRun: user-service-pr-142
Namespace: default
TaskRun: user-service-pr-142-build-image
Failed step: docker-build
Exit code: 1
Reason: Failed
Error: failed to solve image build

Short trace:
#12 RUN go test ./...
--- FAIL: TestCreateUser
expected status 201, got 500
FAIL
```

This is useful because the developer immediately knows:

```text
Task failed: build-image
Step failed: docker-build
Likely issue: test failure
Next action: open test logs or reproduce go test locally
```

---

## 14. Common mistakes

### Mistake 1: Sending the full log to Slack

Bad:

```text
Send 5,000 log lines to Slack.
```

Why bad:

```text
Noisy, slow, unreadable, may leak secrets.
```

Better:

```text
Send 20 to 40 useful lines.
```

---

### Mistake 2: Not masking secrets

Dangerous log:

```text
token=ghp_abc123secret
password=my-prod-password
Authorization: Bearer eyJhbGciOi...
```

Safe Slack output:

```text
token=****
password=****
Authorization: Bearer ****
```

---

### Mistake 3: Only showing PipelineRun name

A PipelineRun can contain many tasks.

This is not enough:

```text
PipelineRun failed: user-service-pr-142
```

Better:

```text
PipelineRun failed: user-service-pr-142
TaskRun: user-service-pr-142-build-image
Step: docker-build
```

---

### Mistake 4: Mixing collection and formatting too much

Bad design:

```go
func CollectFailureAndSendSlackAndFormatEverything()
```

Better design:

```text
Collect failure
Build model
Format message
Send message
```

Small functions are easier to test.

---

### Mistake 5: Assuming every failure has a clear error line

Sometimes logs are messy.

So your formatter should handle missing values:

```text
Error: No clear error line found
```

instead of crashing.

---

## 15. Debugging tips

When your failure collector does not work, debug one layer at a time.

### Check TaskRun JSON

```bash
kubectl -n default get taskrun <taskrun-name> -o json | jq '.status'
```

Look for:

```text
.status.steps
.status.conditions
.status.podName
```

---

### Check pod containers

```bash
kubectl -n default get pod <pod-name> \
  -o jsonpath='{.spec.containers[*].name}'
```

You may see:

```text
step-clone step-test step-build
```

Then logs should use:

```bash
kubectl -n default logs <pod-name> -c step-test
```

---

### Check failed step extraction

```bash
kubectl -n default get taskrun <taskrun-name> -o json |
  jq '.status.steps[] | {name, terminated}'
```

---

### Check raw logs first

```bash
kubectl -n default logs <pod-name> -c step-test --tail=100
```

Only after that should you debug parsing.

---

### Check JSON output from shell

```bash
./scripts/collect-failure.sh <taskrun-name> default | jq .
```

If this fails, the Go code is not the problem yet.

---

### Check Go formatting separately

Save sample JSON:

```bash
cat > sample-failure.json <<'EOF'
{
  "namespace": "default",
  "task_run": "demo-taskrun",
  "failed_step": "test",
  "exit_code": "1",
  "reason": "Failed",
  "error_message": "go test failed",
  "trace": "FAIL\nexit status 1"
}
EOF
```

Run:

```bash
cat sample-failure.json | go run ./cmd/slack-notifier
```

---

## 16. One DSA topic: dynamic programming basics

Dynamic programming, or DP, is a technique for solving problems by reusing answers to smaller subproblems.

The beginner-friendly idea:

```text
Do not solve the same smaller problem again and again.
Store the answer and reuse it.
```

Classic example:

```text
Fibonacci numbers
```

Naive recursion:

```text
fib(5)
= fib(4) + fib(3)

fib(4)
= fib(3) + fib(2)
```

Notice `fib(3)` gets calculated multiple times.

DP says:

```text
Calculate fib(0), fib(1), fib(2), ...
Store them.
Reuse them.
```

Python style:

```python
dp = [0] * (n + 1)
dp[0] = 0
dp[1] = 1

for i in range(2, n + 1):
    dp[i] = dp[i - 1] + dp[i - 2]
```

Go style:

```go
dp := make([]int, n+1)
dp[0] = 0
dp[1] = 1

for i := 2; i <= n; i++ {
	dp[i] = dp[i-1] + dp[i-2]
}
```

Important syntax differences:

| Python                      | Go                           |
| --------------------------- | ---------------------------- |
| `dp = [0] * (n + 1)`        | `dp := make([]int, n+1)`     |
| `for i in range(2, n + 1):` | `for i := 2; i <= n; i++ {}` |
| indentation creates blocks  | braces `{}` create blocks    |
| dynamic typing              | explicit/static typing       |

DP appears in CI/CD thinking too.

For example, when parsing logs, you may process lines and keep useful state:

```text
Have I seen an error?
What were the previous few lines?
What is the best error line so far?
```

That is not exactly DP, but the habit is similar: reuse remembered information instead of reprocessing everything blindly.

---

## 17. One Go DSA problem: climbing stairs

### Problem

You are climbing a staircase with `n` steps.

Each time, you can climb either:

```text
1 step
or
2 steps
```

How many distinct ways can you reach the top?

Example:

```text
n = 3
Ways:
1 + 1 + 1
1 + 2
2 + 1

Answer: 3
```

### DP thinking

To reach step `n`, your last move was either:

```text
from step n-1 using 1 step
from step n-2 using 2 steps
```

So:

```text
ways[n] = ways[n-1] + ways[n-2]
```

Base cases:

```text
ways[0] = 1
ways[1] = 1
```

Why `ways[0] = 1`?

Because there is one way to stay at the ground: do nothing.

### Go solution

```go
package main

import "fmt"

func climbStairs(n int) int {
	if n < 0 {
		return 0
	}

	dp := make([]int, n+1)

	dp[0] = 1

	if n >= 1 {
		dp[1] = 1
	}

	for i := 2; i <= n; i++ {
		dp[i] = dp[i-1] + dp[i-2]
	}

	return dp[n]
}

func main() {
	fmt.Println(climbStairs(1)) // 1
	fmt.Println(climbStairs(2)) // 2
	fmt.Println(climbStairs(3)) // 3
	fmt.Println(climbStairs(4)) // 5
	fmt.Println(climbStairs(5)) // 8
}
```

Python comparison:

```python
def climb_stairs(n):
    dp = [0] * (n + 1)
    dp[0] = 1

    if n >= 1:
        dp[1] = 1

    for i in range(2, n + 1):
        dp[i] = dp[i - 1] + dp[i - 2]

    return dp[n]
```

Same logic, different syntax.

Go convention uses:

```go
climbStairs
```

Python convention uses:

```python
climb_stairs
```

That is another important naming shift.

---

## 18. Module-based practice task: error trace collector / summary generator

Build a small local module that does not need real Tekton at first.

### Goal

Input:

```text
A log file
```

Output:

```text
A short JSON failure summary
```

### Suggested structure

```text
error-summary/
  go.mod
  cmd/
    error-summary/
      main.go
  internal/
    summary/
      parser.go
      parser_test.go
```

Initialize:

```bash
mkdir error-summary
cd error-summary
go mod init error-summary
mkdir -p cmd/error-summary internal/summary
```

---

### Parser code

File:

```text
internal/summary/parser.go
```

```go
package summary

import (
	"strings"
)

type Result struct {
	ErrorMessage string `json:"error_message"`
	Trace       string `json:"trace"`
	Trimmed     bool   `json:"trimmed"`
}

func Generate(logText string, maxLines int) Result {
	errorMessage := findError(logText)
	trace, trimmed := lastLines(logText, maxLines)

	return Result{
		ErrorMessage: errorMessage,
		Trace:       trace,
		Trimmed:     trimmed,
	}
}

func findError(logText string) string {
	lines := strings.Split(logText, "\n")

	keywords := []string{"error", "failed", "fatal", "panic", "exception", "timeout"}

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		lower := strings.ToLower(line)

		for _, keyword := range keywords {
			if strings.Contains(lower, keyword) {
				return line
			}
		}
	}

	return "No clear error found"
}

func lastLines(logText string, maxLines int) (string, bool) {
	lines := strings.Split(strings.TrimSpace(logText), "\n")

	if len(lines) <= maxLines {
		return strings.Join(lines, "\n"), false
	}

	return strings.Join(lines[len(lines)-maxLines:], "\n"), true
}
```

---

### Main program

File:

```text
cmd/error-summary/main.go
```

```go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"error-summary/internal/summary"
)

func main() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read stdin: %v\n", err)
		os.Exit(1)
	}

	result := summary.Generate(string(input), 20)

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode summary: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}
```

---

### Try it

Create a sample log:

```bash
cat > failed.log <<'EOF'
starting build
downloading modules
running tests
=== RUN   TestCreateUser
--- FAIL: TestCreateUser
    user_test.go:42: expected status 201, got 500
FAIL
exit status 1
EOF
```

Run:

```bash
cat failed.log | go run ./cmd/error-summary
```

Expected output:

```json
{
  "error_message": "exit status 1",
  "trace": "starting build\ndownloading modules\nrunning tests\n=== RUN   TestCreateUser\n--- FAIL: TestCreateUser\n    user_test.go:42: expected status 201, got 500\nFAIL\nexit status 1",
  "trimmed": false
}
```

Then improve it:

```text
Instead of choosing "exit status 1",
try choosing "--- FAIL: TestCreateUser"
or "expected status 201, got 500".
```

That is where real log parsing becomes interesting.

---

## 19. Revision checkpoint

Answer these before moving on:

1. What is the difference between a failed `TaskRun` and a failed step?
2. Why is `Build failed` not enough?
3. Where do step logs usually come from?
4. Why should Slack receive a trace snippet instead of full logs?
5. What kinds of secrets should be masked?
6. What is the purpose of a structured `FailureContext` model?
7. Why should collector, formatter, and sender be separate?
8. What does `json:"failed_step"` mean in a Go struct?
9. What is the Go equivalent of Python’s `None`?
10. What is the basic DP formula for climbing stairs?

Expected answers:

```text
1. TaskRun is the whole task execution; step is one container/action inside it.
2. It does not explain where or why the failure happened.
3. From the Kubernetes pod container logs created by Tekton steps.
4. Full logs are noisy, unreadable, and may expose secrets.
5. Tokens, passwords, API keys, Authorization headers, private keys.
6. To pass clean failure data between collector, formatter, and sender.
7. Separation makes the code easier to test and maintain.
8. It controls the JSON field name.
9. nil.
10. ways[n] = ways[n-1] + ways[n-2].
```

---

## 20. Homework

For homework, enhance your Slack integration with safe failure summaries.

### Part A: Collector

Create:

```text
scripts/collect-failure.sh
```

It should output JSON with:

```json
{
  "namespace": "default",
  "task_run": "example-taskrun",
  "pod": "example-pod",
  "failed_step": "test",
  "exit_code": "1",
  "reason": "Failed",
  "error_message": "some useful error",
  "trace": "short trace"
}
```

---

### Part B: Go formatter

Create or update:

```text
internal/model/failure.go
internal/failure/text.go
internal/slack/formatter.go
```

Your Slack message must include:

```text
TaskRun
failed step
exit code
error message
short trace
trimmed trace note
```

---

### Part C: Safety

Add masking for at least these:

```text
password=
token=
secret=
api_key=
Authorization: Bearer
PRIVATE KEY blocks
```

---

### Part D: Tests

Write tests for:

```text
FormatFailureMessage
MaskSecrets
TruncateTrace
FindErrorMessage
```

Minimum test cases:

```text
1. Formatter includes failed step.
2. Formatter includes error message.
3. Secrets are masked.
4. Long trace is trimmed.
5. Missing error line returns fallback message.
```

---

### Part E: DSA

Implement climbing stairs in Go twice:

1. Using a `dp` slice.
2. Using two variables only.

Two-variable version hint:

```go
prev2 := 1
prev1 := 1

for i := 2; i <= n; i++ {
	current := prev1 + prev2
	prev2 = prev1
	prev1 = current
}
```

The big Day 14 takeaway:

```text
A good CI/CD notification does not just say that something failed.
It gives the developer the safest, shortest, most useful path to begin debugging.
```
