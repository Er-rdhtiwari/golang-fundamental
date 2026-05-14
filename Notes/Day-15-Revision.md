# Day 15 — Final End-to-End Revision and Confidence Build

Today is about seeing the whole project as **one connected system**.

Earlier days likely felt like separate parts:

* CLI commands
* request models
* routing logic
* Slack messages
* shell commands
* Tekton pipelines
* Kubernetes resources
* failure traces
* tests
* logs

Day 15 is where these become a **real project**.

Think of it like this:

> You are no longer just learning Go syntax.
> You are learning how backend systems are wired together.

---

# 1. Day 15 learning goals

By the end of Day 15, you should be able to explain and build this flow:

```text
CLI command
  -> Go model
  -> router
  -> Slack notification
  -> shell command
  -> Tekton trigger
  -> Kubernetes resource
  -> pipeline execution
  -> result notification
  -> failure trace capture
```

You should feel comfortable with:

* how a request enters the app
* how data moves between packages
* how errors are captured and returned
* how logs help debugging
* how Slack, shell, Tekton, and Kubernetes connect
* how to test individual modules
* how to design one final enhancement cleanly

Python comparison:

In Python, you may have used:

```python
argparse
dataclasses
requests
subprocess
logging
pytest
```

In Go, the equivalents often look like:

```go
cobra / flag
structs
net/http
os/exec
zerolog
testing
```

The biggest mindset shift:

> Python often uses exceptions.
> Go expects you to return and check errors explicitly.

---

# 2. Full revision of Days 1 to 14 in a structured way

Let’s revise the journey as if each day added one layer.

## Days 1–2: Go basics

You learned:

```go
package main

import "fmt"

func main() {
    fmt.Println("hello")
}
```

Python comparison:

```python
print("hello")
```

Key differences:

| Concept     | Python                  | Go                      |
| ----------- | ----------------------- | ----------------------- |
| Entry point | any script line can run | `func main()`           |
| Package     | module file             | `package main`          |
| Imports     | `import os`             | `import "fmt"`          |
| Types       | dynamic                 | mostly explicit/static  |
| Error style | exceptions              | returned `error` values |

---

## Days 3–4: Structs and models

You probably created request models.

Go:

```go
type DeployRequest struct {
    AppName   string `json:"app_name"`
    Namespace string `json:"namespace"`
    Image     string `json:"image"`
}
```

Python equivalent:

```python
from dataclasses import dataclass

@dataclass
class DeployRequest:
    app_name: str
    namespace: str
    image: str
```

Important Go idea:

> A `struct` is your main way to represent project data.

The JSON tags:

```go
`json:"app_name"`
```

tell Go how to encode/decode JSON.

---

## Days 5–6: Interfaces and services

You may have separated implementation from behavior.

```go
type SlackNotifier interface {
    SendMessage(ctx context.Context, msg SlackMessage) error
}
```

Python comparison:

```python
class SlackNotifier:
    def send_message(self, msg):
        raise NotImplementedError
```

In Go, interfaces are satisfied implicitly.

That means this works without saying “implements”:

```go
type SlackClient struct{}

func (s *SlackClient) SendMessage(ctx context.Context, msg SlackMessage) error {
    return nil
}
```

If the method matches, Go accepts it.

This is very important for testing.

---

## Days 7–8: CLI layer

The CLI is the user-facing entry point.

Example command:

```bash
slackctl trigger --app payments --namespace dev --image payments:v1
```

The CLI should not do all the work directly.

It should:

1. parse user input
2. build a request model
3. call the router or service
4. print the result

Good design:

```text
CLI = input adapter
router/service = business logic
clients = external systems
```

Python comparison:

| Python        | Go              |
| ------------- | --------------- |
| `argparse`    | `flag`, `cobra` |
| function call | function call   |
| dict payload  | struct payload  |
| exceptions    | `error` return  |

---

## Days 9–10: Slack integration

Slack is usually used for:

* deployment started
* deployment succeeded
* deployment failed
* error trace attached
* Tekton pipeline link shared

Simple mental model:

```text
Go app builds Slack message
  -> sends HTTP POST
  -> Slack channel receives notification
```

In Python, you might do:

```python
requests.post(webhook_url, json=payload)
```

In Go:

```go
req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, body)
```

More verbose, but safer and more explicit.

---

## Days 11–12: Shell and Tekton

Your Go app may call shell commands such as:

```bash
tkn pipeline start deploy-app
```

or:

```bash
kubectl apply -f resource.yaml
```

Python equivalent:

```python
subprocess.run(["kubectl", "apply", "-f", "resource.yaml"])
```

Go equivalent:

```go
cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", "resource.yaml")
out, err := cmd.CombinedOutput()
```

Important:

> Shell commands are powerful but risky.
> Always validate inputs and capture output.

---

## Days 13–14: Kubernetes, failures, and testing

You learned the project is not only about “happy path”.

Production systems fail because of:

* invalid config
* missing namespace
* Slack webhook failure
* Tekton pipeline failure
* Kubernetes permissions
* timeout
* bad image
* wrong service account
* bad YAML

So you added:

* structured logs
* error wrapping
* trace capture
* unit tests
* cleaner routing

Day 15 combines all of that.

---

# 3. Complete architecture in ASCII

```text
                         ┌────────────────────┐
                         │        User         │
                         │  slackctl trigger   │
                         └─────────┬──────────┘
                                   │
                                   ▼
                         ┌────────────────────┐
                         │        CLI         │
                         │ parse flags/input  │
                         └─────────┬──────────┘
                                   │ DeployRequest
                                   ▼
                         ┌────────────────────┐
                         │       Model        │
                         │ validate request   │
                         └─────────┬──────────┘
                                   │
                                   ▼
                         ┌────────────────────┐
                         │       Router       │
                         │ choose action      │
                         └──────┬───────┬─────┘
                                │       │
              start notification│       │ trigger pipeline
                                ▼       ▼
                    ┌────────────────┐  ┌────────────────┐
                    │ Slack Package  │  │ Shell Package  │
                    │ send messages  │  │ run commands   │
                    └───────┬────────┘  └───────┬────────┘
                            │                   │
                            ▼                   ▼
                    ┌────────────────┐  ┌────────────────┐
                    │ Slack Channel  │  │ Tekton CLI/API │
                    └────────────────┘  └───────┬────────┘
                                                │
                                                ▼
                                      ┌────────────────────┐
                                      │     Kubernetes     │
                                      │ PipelineRun, Pods  │
                                      └─────────┬──────────┘
                                                │
                                                ▼
                                      ┌────────────────────┐
                                      │  Result Collector  │
                                      │ success/failure    │
                                      └─────────┬──────────┘
                                                │
                         ┌──────────────────────┴──────────────────────┐
                         ▼                                             ▼
              ┌────────────────────┐                       ┌────────────────────┐
              │ Failure Trace       │                       │ Final Slack Update │
              │ logs/events/output  │                       │ success/failure    │
              └────────────────────┘                       └────────────────────┘
```

The beginner-friendly idea:

> The CLI starts the request.
> The router decides what to do.
> Slack tells humans what is happening.
> Shell/Tekton/Kubernetes do the actual infrastructure work.
> Logs and traces help you debug when things break.

---

# 4. Complete request/response flow

Example command:

```bash
slackctl trigger \
  --app payments \
  --namespace dev \
  --image registry.example.com/payments:v1
```

Flow:

```text
1. User runs CLI command

2. CLI reads flags:
   app = payments
   namespace = dev
   image = registry.example.com/payments:v1

3. CLI creates DeployRequest struct

4. Request is validated:
   - app name exists
   - namespace exists
   - image exists

5. Router receives request

6. Router sends "deployment started" Slack message

7. Router triggers Tekton pipeline through shell command

8. Tekton creates PipelineRun inside Kubernetes

9. Kubernetes starts pipeline pods

10. App waits/checks result

11. If success:
    - send success Slack message
    - return success response

12. If failure:
    - capture trace
    - send failure Slack message
    - return failure response
```

A possible response model:

```go
type TriggerResponse struct {
    RequestID   string `json:"request_id"`
    AppName     string `json:"app_name"`
    Namespace   string `json:"namespace"`
    PipelineRun string `json:"pipeline_run"`
    Status      string `json:"status"`
    Message     string `json:"message"`
}
```

Python equivalent:

```python
@dataclass
class TriggerResponse:
    request_id: str
    app_name: str
    namespace: str
    pipeline_run: str
    status: str
    message: str
```

---

# 5. Complete Slack notification flow

Slack should not just say “failed”.

Good Slack messages answer:

* What happened?
* Which app?
* Which namespace?
* Which pipeline?
* Who/what triggered it?
* What should I check next?

## Started message

```text
Deployment started

App: payments
Namespace: dev
Image: registry.example.com/payments:v1
Request ID: req-123
```

## Success message

```text
Deployment succeeded

App: payments
Namespace: dev
PipelineRun: payments-run-abc123
Duration: 2m 12s
```

## Failure message

```text
Deployment failed

App: payments
Namespace: dev
PipelineRun: payments-run-abc123
Reason: task build-image failed
Trace ID: trace-789
Suggested check: kubectl logs pod/<pod-name> -n dev
```

Slack flow:

```text
Router
  -> SlackNotifier.SendStarted()
  -> Tekton trigger
  -> result check
  -> SlackNotifier.SendSuccess()
       OR
     SlackNotifier.SendFailure()
```

Good package design:

```go
type Notifier interface {
    DeploymentStarted(ctx context.Context, req DeployRequest) error
    DeploymentSucceeded(ctx context.Context, res TriggerResponse) error
    DeploymentFailed(ctx context.Context, trace FailureTrace) error
}
```

Python comparison:

In Python, you might pass a class with methods.

In Go, you commonly define an interface so tests can use a fake notifier.

---

# 6. Complete Tekton pipeline flow

Tekton flow:

```text
Go app
  -> runs tkn command or calls Kubernetes API
  -> creates PipelineRun
  -> Tekton controller sees PipelineRun
  -> Tekton creates TaskRuns
  -> TaskRuns create Pods
  -> Pods execute steps
  -> status updates on PipelineRun
```

ASCII:

```text
┌────────────┐
│ Go Router  │
└─────┬──────┘
      │
      ▼
┌────────────┐
│ tkn start  │
└─────┬──────┘
      │
      ▼
┌────────────────┐
│ PipelineRun    │
│ Kubernetes CRD │
└─────┬──────────┘
      │
      ▼
┌────────────────┐
│ Tekton Control │
│ creates tasks  │
└─────┬──────────┘
      │
      ▼
┌──────────────┐
│ TaskRun Pods │
└─────┬────────┘
      │
      ▼
┌──────────────┐
│ Success/Fail │
└──────────────┘
```

Example Tekton Pipeline:

```yaml
apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: deploy-app
spec:
  params:
    - name: app-name
      type: string
    - name: image
      type: string

  tasks:
    - name: validate
      taskRef:
        name: validate-request

    - name: deploy
      runAfter:
        - validate
      taskRef:
        name: deploy-to-kubernetes
      params:
        - name: app-name
          value: $(params.app-name)
        - name: image
          value: $(params.image)

    - name: notify
      runAfter:
        - deploy
      taskRef:
        name: notify-slack
```

Important Day 15 connection:

> Tekton tasks are dependency-based.
> That connects directly to today’s DSA topic: topological sort.

---

# 7. Complete Kubernetes resource flow

When Tekton runs, Kubernetes resources are created.

```text
Pipeline
  -> PipelineRun
  -> TaskRun
  -> Pod
  -> Container
  -> Logs
  -> Status
```

Detailed flow:

```text
1. Pipeline already exists in cluster

2. App creates PipelineRun

3. Kubernetes stores PipelineRun as a custom resource

4. Tekton controller watches PipelineRun

5. Tekton creates TaskRuns

6. TaskRuns create Pods

7. Pods execute steps

8. Container logs are written

9. PipelineRun status becomes:
   - Succeeded
   - Failed
   - Running
   - Unknown
```

Example PipelineRun:

```yaml
apiVersion: tekton.dev/v1
kind: PipelineRun
metadata:
  generateName: deploy-payments-
  namespace: dev
spec:
  pipelineRef:
    name: deploy-app
  params:
    - name: app-name
      value: payments
    - name: image
      value: registry.example.com/payments:v1
```

The Go app can create this using either:

```text
Option 1: shell command
  kubectl create -f pipelinerun.yaml

Option 2: Kubernetes Go client
  client.Create(ctx, pipelineRun)

Option 3: Tekton CLI
  tkn pipeline start deploy-app ...
```

For a beginner project, shell commands are okay.

For production, Kubernetes clients are better.

---

# 8. Complete failure trace capture flow

A beginner project often stops here:

```text
pipeline failed
```

A better project captures useful trace details.

Failure trace should include:

```go
type FailureTrace struct {
    TraceID     string `json:"trace_id"`
    RequestID   string `json:"request_id"`
    AppName     string `json:"app_name"`
    Namespace   string `json:"namespace"`
    PipelineRun string `json:"pipeline_run"`
    FailedTask  string `json:"failed_task"`
    Reason      string `json:"reason"`
    Logs        string `json:"logs"`
    Events      string `json:"events"`
}
```

Failure flow:

```text
Pipeline failed
  -> router calls TraceCollector
  -> TraceCollector gets:
       pipeline status
       failed task
       pod logs
       Kubernetes events
  -> trace is logged with zerolog
  -> Slack failure message includes trace ID
  -> CLI prints short failure response
```

ASCII:

```text
┌────────────────────┐
│ Tekton failure     │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│ Trace Collector    │
│ get logs/events    │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│ Structured Log     │
│ trace_id=req-123   │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│ Slack Failure Msg  │
│ useful summary     │
└────────────────────┘
```

Python comparison:

Python usually uses exceptions and traceback:

```python
try:
    run_pipeline()
except Exception as e:
    logger.exception("pipeline failed")
```

Go usually does:

```go
if err != nil {
    return fmt.Errorf("trigger pipeline: %w", err)
}
```

The `%w` keeps the original error wrapped inside the new error.

---

# 9. How every important package connects to the others

A clean project may look like this:

```text
slack-integration/
├── cmd/
│   └── slackctl/
│       └── main.go
├── internal/
│   ├── app/
│   │   └── service.go
│   ├── config/
│   │   └── config.go
│   ├── model/
│   │   └── deploy.go
│   ├── router/
│   │   └── router.go
│   ├── slack/
│   │   └── client.go
│   ├── shell/
│   │   └── runner.go
│   ├── tekton/
│   │   └── trigger.go
│   ├── k8s/
│   │   └── trace.go
│   ├── logger/
│   │   └── logger.go
│   └── planner/
│       └── topo.go
└── README.md
```

## `cmd/`

Entry point.

It parses command-line input.

```text
cmd -> model -> router
```

Python equivalent:

```text
main.py
```

---

## `internal/model`

Defines data.

```go
type DeployRequest struct {
    AppName   string
    Namespace string
    Image     string
}
```

Python equivalent:

```python
@dataclass
class DeployRequest:
    app_name: str
    namespace: str
    image: str
```

---

## `internal/router`

Decides which flow to run.

```text
router -> slack
router -> tekton
router -> trace collector
```

It should not contain low-level HTTP or shell details.

---

## `internal/slack`

Sends Slack messages.

```text
slack package -> Slack webhook/API
```

---

## `internal/shell`

Runs shell commands safely.

```text
shell package -> os/exec
```

---

## `internal/tekton`

Builds and starts Tekton pipeline commands.

```text
tekton package -> shell package
```

This keeps Tekton-specific command construction outside the router.

---

## `internal/k8s`

Reads Kubernetes status, logs, and events.

```text
k8s package -> kubectl or Kubernetes client
```

---

## `internal/logger`

Creates structured logger.

```text
all packages -> logger
```

With zerolog, logs become JSON-like and searchable.

---

## `internal/planner`

This is your final mini-project module.

It handles dependency ordering.

```text
planner -> topological sort
Tekton tasks -> dependency graph -> execution order
```

---

# 10. Pseudocode first for the final end-to-end flow

Start with simple thinking before Go code.

```text
function triggerDeployment(input):
    requestID = generateRequestID()

    request = build DeployRequest from input

    validate request
    if invalid:
        return error

    log "deployment requested"

    send Slack started notification
    if Slack fails:
        log warning but continue or fail based on config

    pipelineRun = trigger Tekton pipeline
    if trigger fails:
        trace = capture trigger failure
        send Slack failure notification
        return error

    result = wait for pipeline result

    if result succeeded:
        send Slack success notification
        return success response

    if result failed:
        trace = capture pipeline logs and events
        log trace
        send Slack failure notification
        return failure response
```

The main idea:

> Every external action can fail, so every step returns an error.

---

# 11. Real code snippets where useful

## Model

```go
package model

import "fmt"

type DeployRequest struct {
    RequestID string
    AppName   string
    Namespace string
    Image     string
}

func (r DeployRequest) Validate() error {
    if r.AppName == "" {
        return fmt.Errorf("app name is required")
    }
    if r.Namespace == "" {
        return fmt.Errorf("namespace is required")
    }
    if r.Image == "" {
        return fmt.Errorf("image is required")
    }
    return nil
}
```

Python comparison:

```python
@dataclass
class DeployRequest:
    request_id: str
    app_name: str
    namespace: str
    image: str

    def validate(self):
        if not self.app_name:
            raise ValueError("app name is required")
```

Important difference:

* Python raises `ValueError`
* Go returns `error`

---

## Logger using zerolog

```go
package logger

import (
    "os"
    "time"

    "github.com/rs/zerolog"
)

func New() zerolog.Logger {
    zerolog.TimeFieldFormat = time.RFC3339

    return zerolog.New(os.Stdout).
        With().
        Timestamp().
        Str("service", "slack-integration").
        Logger()
}
```

Usage:

```go
log.Info().
    Str("request_id", req.RequestID).
    Str("app", req.AppName).
    Str("namespace", req.Namespace).
    Msg("deployment requested")
```

Output:

```json
{
  "level": "info",
  "service": "slack-integration",
  "request_id": "req-123",
  "app": "payments",
  "namespace": "dev",
  "message": "deployment requested"
}
```

Python comparison:

```python
logger.info(
    "deployment requested",
    extra={"request_id": req.request_id, "app": req.app_name}
)
```

Go convention shift:

> Go logging often chains fields before the message.

---

## Router with clean dependencies

```go
package router

import (
    "context"
    "fmt"

    "github.com/rs/zerolog"
    "slack-integration/internal/model"
)

type Notifier interface {
    DeploymentStarted(ctx context.Context, req model.DeployRequest) error
    DeploymentSucceeded(ctx context.Context, res model.TriggerResponse) error
    DeploymentFailed(ctx context.Context, trace model.FailureTrace) error
}

type PipelineTrigger interface {
    Trigger(ctx context.Context, req model.DeployRequest) (string, error)
    WaitForResult(ctx context.Context, namespace, pipelineRun string) (model.PipelineResult, error)
}

type TraceCollector interface {
    Capture(ctx context.Context, req model.DeployRequest, pipelineRun string, cause error) model.FailureTrace
}

type Router struct {
    notifier Notifier
    trigger  PipelineTrigger
    traces   TraceCollector
    log      zerolog.Logger
}

func New(
    notifier Notifier,
    trigger PipelineTrigger,
    traces TraceCollector,
    log zerolog.Logger,
) *Router {
    return &Router{
        notifier: notifier,
        trigger:  trigger,
        traces:   traces,
        log:      log,
    }
}

func (r *Router) TriggerDeployment(ctx context.Context, req model.DeployRequest) (model.TriggerResponse, error) {
    if err := req.Validate(); err != nil {
        return model.TriggerResponse{}, fmt.Errorf("validate request: %w", err)
    }

    r.log.Info().
        Str("request_id", req.RequestID).
        Str("app", req.AppName).
        Str("namespace", req.Namespace).
        Msg("starting deployment flow")

    if err := r.notifier.DeploymentStarted(ctx, req); err != nil {
        r.log.Warn().
            Err(err).
            Str("request_id", req.RequestID).
            Msg("failed to send start notification")
    }

    pipelineRun, err := r.trigger.Trigger(ctx, req)
    if err != nil {
        trace := r.traces.Capture(ctx, req, "", err)
        _ = r.notifier.DeploymentFailed(ctx, trace)

        return model.TriggerResponse{}, fmt.Errorf("trigger tekton pipeline: %w", err)
    }

    result, err := r.trigger.WaitForResult(ctx, req.Namespace, pipelineRun)
    if err != nil {
        trace := r.traces.Capture(ctx, req, pipelineRun, err)
        _ = r.notifier.DeploymentFailed(ctx, trace)

        return model.TriggerResponse{}, fmt.Errorf("wait for pipeline result: %w", err)
    }

    if result.Status != "Succeeded" {
        cause := fmt.Errorf("pipeline finished with status %s", result.Status)
        trace := r.traces.Capture(ctx, req, pipelineRun, cause)
        _ = r.notifier.DeploymentFailed(ctx, trace)

        return model.TriggerResponse{}, cause
    }

    response := model.TriggerResponse{
        RequestID:   req.RequestID,
        AppName:     req.AppName,
        Namespace:   req.Namespace,
        PipelineRun: pipelineRun,
        Status:      "Succeeded",
        Message:     "deployment completed successfully",
    }

    if err := r.notifier.DeploymentSucceeded(ctx, response); err != nil {
        r.log.Warn().
            Err(err).
            Str("request_id", req.RequestID).
            Msg("failed to send success notification")
    }

    return response, nil
}
```

This is the heart of the project.

Notice what the router does **not** know:

* how Slack HTTP works
* exact shell command details
* exact Kubernetes log command
* exact Tekton YAML format

That is clean design.

---

## Shell runner

```go
package shell

import (
    "context"
    "fmt"
    "os/exec"
)

type Runner struct{}

func (r Runner) Run(ctx context.Context, name string, args ...string) (string, error) {
    cmd := exec.CommandContext(ctx, name, args...)

    output, err := cmd.CombinedOutput()
    if err != nil {
        return string(output), fmt.Errorf("run command %s: %w", name, err)
    }

    return string(output), nil
}
```

Python equivalent:

```python
result = subprocess.run(
    ["kubectl", "get", "pods"],
    capture_output=True,
    text=True,
    check=True,
)
```

Go difference:

* `CombinedOutput()` returns stdout and stderr together
* error must be checked manually

---

## Tekton trigger package

```go
package tekton

import (
    "context"
    "fmt"

    "slack-integration/internal/model"
)

type ShellRunner interface {
    Run(ctx context.Context, name string, args ...string) (string, error)
}

type Client struct {
    shell ShellRunner
}

func New(shell ShellRunner) Client {
    return Client{shell: shell}
}

func (c Client) Trigger(ctx context.Context, req model.DeployRequest) (string, error) {
    pipelineRunName := fmt.Sprintf("deploy-%s-%s", req.AppName, req.RequestID)

    _, err := c.shell.Run(
        ctx,
        "tkn",
        "pipeline",
        "start",
        "deploy-app",
        "-n", req.Namespace,
        "-p", "app-name="+req.AppName,
        "-p", "image="+req.Image,
        "--use-param-defaults",
        "--showlog=false",
    )
    if err != nil {
        return "", fmt.Errorf("start tekton pipeline: %w", err)
    }

    return pipelineRunName, nil
}
```

In a real system, you would parse the actual PipelineRun name from `tkn` output or create the PipelineRun directly using Kubernetes APIs.

---

## Failure trace collector

```go
package k8s

import (
    "context"
    "fmt"

    "slack-integration/internal/model"
)

type ShellRunner interface {
    Run(ctx context.Context, name string, args ...string) (string, error)
}

type TraceCollector struct {
    shell ShellRunner
}

func NewTraceCollector(shell ShellRunner) TraceCollector {
    return TraceCollector{shell: shell}
}

func (c TraceCollector) Capture(
    ctx context.Context,
    req model.DeployRequest,
    pipelineRun string,
    cause error,
) model.FailureTrace {
    events, _ := c.shell.Run(
        ctx,
        "kubectl",
        "get",
        "events",
        "-n", req.Namespace,
        "--sort-by=.lastTimestamp",
    )

    return model.FailureTrace{
        TraceID:     "trace-" + req.RequestID,
        RequestID:   req.RequestID,
        AppName:     req.AppName,
        Namespace:   req.Namespace,
        PipelineRun: pipelineRun,
        Reason:      fmt.Sprintf("%v", cause),
        Events:      events,
    }
}
```

Beginner note:

This ignores errors from `kubectl get events`.

That is acceptable for a first trace collector because the original failure is more important. In production, you would include trace collection errors too.

---

## Slack client

```go
package slack

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"

    "slack-integration/internal/model"
)

type Client struct {
    webhookURL string
    httpClient *http.Client
}

func New(webhookURL string) Client {
    return Client{
        webhookURL: webhookURL,
        httpClient: http.DefaultClient,
    }
}

type webhookPayload struct {
    Text string `json:"text"`
}

func (c Client) send(ctx context.Context, text string) error {
    body, err := json.Marshal(webhookPayload{Text: text})
    if err != nil {
        return fmt.Errorf("marshal slack payload: %w", err)
    }

    req, err := http.NewRequestWithContext(
        ctx,
        http.MethodPost,
        c.webhookURL,
        bytes.NewReader(body),
    )
    if err != nil {
        return fmt.Errorf("create slack request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")

    res, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("send slack request: %w", err)
    }
    defer res.Body.Close()

    if res.StatusCode >= 300 {
        return fmt.Errorf("slack returned status %d", res.StatusCode)
    }

    return nil
}

func (c Client) DeploymentStarted(ctx context.Context, req model.DeployRequest) error {
    return c.send(ctx, fmt.Sprintf(
        "Deployment started\nApp: %s\nNamespace: %s\nImage: %s\nRequest ID: %s",
        req.AppName,
        req.Namespace,
        req.Image,
        req.RequestID,
    ))
}

func (c Client) DeploymentSucceeded(ctx context.Context, res model.TriggerResponse) error {
    return c.send(ctx, fmt.Sprintf(
        "Deployment succeeded\nApp: %s\nNamespace: %s\nPipelineRun: %s",
        res.AppName,
        res.Namespace,
        res.PipelineRun,
    ))
}

func (c Client) DeploymentFailed(ctx context.Context, trace model.FailureTrace) error {
    return c.send(ctx, fmt.Sprintf(
        "Deployment failed\nApp: %s\nNamespace: %s\nReason: %s\nTrace ID: %s",
        trace.AppName,
        trace.Namespace,
        trace.Reason,
        trace.TraceID,
    ))
}
```

Python comparison:

Go:

```go
json.Marshal(payload)
http.NewRequestWithContext(...)
client.Do(req)
```

Python:

```python
requests.post(url, json=payload)
```

Go is more verbose, but you gain better control over context, cancellation, timeout, and errors.

---

## Unit test for router

```go
package router_test

import (
    "context"
    "testing"

    "github.com/rs/zerolog"
    "slack-integration/internal/model"
    "slack-integration/internal/router"
)

type fakeNotifier struct {
    started bool
    success bool
    failed  bool
}

func (f *fakeNotifier) DeploymentStarted(ctx context.Context, req model.DeployRequest) error {
    f.started = true
    return nil
}

func (f *fakeNotifier) DeploymentSucceeded(ctx context.Context, res model.TriggerResponse) error {
    f.success = true
    return nil
}

func (f *fakeNotifier) DeploymentFailed(ctx context.Context, trace model.FailureTrace) error {
    f.failed = true
    return nil
}

type fakeTrigger struct{}

func (f fakeTrigger) Trigger(ctx context.Context, req model.DeployRequest) (string, error) {
    return "deploy-payments-123", nil
}

func (f fakeTrigger) WaitForResult(ctx context.Context, namespace, pipelineRun string) (model.PipelineResult, error) {
    return model.PipelineResult{Status: "Succeeded"}, nil
}

type fakeTraceCollector struct{}

func (f fakeTraceCollector) Capture(
    ctx context.Context,
    req model.DeployRequest,
    pipelineRun string,
    cause error,
) model.FailureTrace {
    return model.FailureTrace{}
}

func TestTriggerDeploymentSuccess(t *testing.T) {
    notifier := &fakeNotifier{}

    r := router.New(
        notifier,
        fakeTrigger{},
        fakeTraceCollector{},
        zerolog.Nop(),
    )

    req := model.DeployRequest{
        RequestID: "123",
        AppName: "payments",
        Namespace: "dev",
        Image: "payments:v1",
    }

    res, err := r.TriggerDeployment(context.Background(), req)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    if res.Status != "Succeeded" {
        t.Fatalf("expected status Succeeded, got %s", res.Status)
    }

    if !notifier.started {
        t.Fatal("expected started notification")
    }

    if !notifier.success {
        t.Fatal("expected success notification")
    }

    if notifier.failed {
        t.Fatal("did not expect failure notification")
    }
}
```

Python comparison:

This is similar to using fake objects or mocks in `pytest`.

Go convention:

```go
if err != nil {
    t.Fatalf(...)
}
```

is very common.

---

# 12. Final mini project enhancement task

Your final enhancement:

## Build a dependency-aware Tekton notification trigger flow

Goal:

Before triggering Tekton, your app should understand the dependency order of pipeline stages.

Example input:

```text
validate -> build -> deploy -> notify
```

Or:

```json
{
  "tasks": [
    {"name": "validate", "depends_on": []},
    {"name": "build", "depends_on": ["validate"]},
    {"name": "deploy", "depends_on": ["build"]},
    {"name": "notify", "depends_on": ["deploy"]}
  ]
}
```

Your app should:

1. read pipeline task dependencies
2. calculate execution order using topological sort
3. log the planned order using zerolog
4. reject cycles
5. trigger Tekton only if the dependency graph is valid
6. send Slack notification with planned order
7. capture trace if planning fails

Example Slack message:

```text
Deployment plan created

App: payments
Execution order:
1. validate
2. build
3. deploy
4. notify
```

Why this is a great final enhancement:

* uses DSA in a real project
* improves confidence
* touches model, router, logger, tests, and Tekton
* prepares you for real CI/CD systems

---

# 13. Suggested refactoring improvements

## 1. Keep router small

Bad:

```go
func TriggerDeployment() {
    // parse CLI
    // send Slack
    // build shell command
    // read logs
    // format response
}
```

Good:

```text
CLI parses input
Router coordinates
Slack client sends messages
Tekton client triggers pipeline
Trace collector captures failure
Planner orders dependencies
```

---

## 2. Use interfaces at package boundaries

Useful interfaces:

```go
type Notifier interface {
    DeploymentStarted(ctx context.Context, req model.DeployRequest) error
}

type PipelineTrigger interface {
    Trigger(ctx context.Context, req model.DeployRequest) (string, error)
}

type ShellRunner interface {
    Run(ctx context.Context, name string, args ...string) (string, error)
}
```

Why?

Because then tests can use fake implementations.

---

## 3. Keep models boring

Models should not know too much.

Good:

```go
type DeployRequest struct {
    AppName string
}
```

Avoid putting Slack, Tekton, and Kubernetes logic inside models.

---

## 4. Prefer explicit error wrapping

Good:

```go
return fmt.Errorf("trigger tekton pipeline: %w", err)
```

Bad:

```go
return err
```

Better error messages create better debugging.

---

## 5. Make package names simple

Good Go package names:

```text
model
router
slack
shell
tekton
k8s
planner
logger
```

Avoid:

```text
utils
helpers
common
misc
```

`utils` often becomes a messy drawer.

---

# 14. Suggested production improvements

For production, improve these areas.

## Security

* validate CLI input
* avoid raw shell string building
* do not print secrets
* read Slack webhook from environment or secret manager
* use Kubernetes service accounts with limited permissions

Bad:

```go
exec.Command("sh", "-c", userInput)
```

Better:

```go
exec.CommandContext(ctx, "tkn", "pipeline", "start", "deploy-app", "-p", "image="+safeImage)
```

---

## Reliability

Add:

* timeouts
* retries for Slack
* retry only safe operations
* context cancellation
* idempotency using request IDs

Example:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()
```

Python comparison:

Similar to setting timeout in `requests` or `subprocess`.

---

## Observability

Add:

* structured logs
* trace IDs
* request IDs
* pipeline run IDs
* latency fields
* failure reasons

Example:

```go
log.Error().
    Err(err).
    Str("request_id", req.RequestID).
    Str("pipeline_run", pipelineRun).
    Msg("deployment failed")
```

---

## Kubernetes-native implementation

Instead of shelling out to:

```bash
kubectl
tkn
```

use:

* Kubernetes Go client
* Tekton client libraries
* informers/watchers

For learning, shell is okay.

For production, native clients are cleaner.

---

## Better testing

Add:

* unit tests for validation
* unit tests for router
* unit tests for planner
* HTTP tests for Slack client
* fake shell runner tests
* integration test with a test namespace

---

# 15. Hands-on tasks

Do these in order.

## Task 1: Draw your architecture

Write your own version of:

```text
CLI -> model -> router -> slack -> tekton -> k8s -> trace
```

Do not skip this. Drawing helps you understand ownership.

---

## Task 2: Add structured logging

Add logs for:

* request received
* validation failed
* Slack message sent
* Tekton trigger started
* PipelineRun created
* pipeline succeeded
* pipeline failed
* trace captured

---

## Task 3: Add error wrapping

Search for:

```go
return err
```

Replace important ones with:

```go
return fmt.Errorf("useful context: %w", err)
```

---

## Task 4: Add planner package

Create:

```text
internal/planner/topo.go
internal/planner/topo_test.go
```

The planner should return pipeline execution order.

---

## Task 5: Connect planner to router

Before triggering Tekton:

```text
router -> planner -> planned order -> Slack -> Tekton
```

---

## Task 6: Add failure trace for planner error

If dependencies contain a cycle:

```text
do not trigger Tekton
send Slack failure message
log cycle error
return error
```

---

# 16. Expected output

## Successful CLI run

```bash
slackctl trigger --app payments --namespace dev --image payments:v1
```

Expected terminal output:

```text
Request ID: req-123
App: payments
Namespace: dev
PipelineRun: deploy-payments-req-123
Status: Succeeded
Message: deployment completed successfully
```

Expected logs:

```json
{"level":"info","request_id":"req-123","app":"payments","message":"deployment requested"}
{"level":"info","request_id":"req-123","message":"dependency plan created"}
{"level":"info","request_id":"req-123","pipeline_run":"deploy-payments-req-123","message":"tekton pipeline triggered"}
{"level":"info","request_id":"req-123","status":"Succeeded","message":"deployment completed"}
```

Expected Slack messages:

```text
Deployment started
App: payments
Namespace: dev
Image: payments:v1
```

```text
Deployment plan created
Execution order:
1. validate
2. build
3. deploy
4. notify
```

```text
Deployment succeeded
App: payments
PipelineRun: deploy-payments-req-123
```

---

## Failed run because of dependency cycle

Example bad graph:

```text
build depends on deploy
deploy depends on build
```

Expected terminal output:

```text
Status: Failed
Message: invalid pipeline dependency graph: cycle detected
Trace ID: trace-req-123
```

Expected Slack message:

```text
Deployment failed before Tekton trigger

Reason: invalid pipeline dependency graph: cycle detected
Trace ID: trace-req-123
```

Expected important behavior:

> Tekton should not be triggered if the dependency graph is invalid.

---

# 17. Common mistakes

## Mistake 1: Putting everything in `main.go`

Bad:

```text
main.go has CLI, Slack, Tekton, Kubernetes, logging, tests
```

Better:

```text
main.go only wires dependencies together
```

---

## Mistake 2: Ignoring errors

Bad:

```go
notifier.DeploymentStarted(ctx, req)
```

Good:

```go
if err := notifier.DeploymentStarted(ctx, req); err != nil {
    log.Warn().Err(err).Msg("failed to send notification")
}
```

---

## Mistake 3: Using shell unsafely

Bad:

```go
exec.Command("sh", "-c", "kubectl apply -f " + fileName)
```

Better:

```go
exec.CommandContext(ctx, "kubectl", "apply", "-f", fileName)
```

---

## Mistake 4: Slack failure hides real failure

If Slack failure happens after pipeline failure, do not replace the original error.

Bad:

```go
return slackErr
```

Better:

```go
log.Warn().Err(slackErr).Msg("failed to send failure notification")
return originalPipelineErr
```

---

## Mistake 5: No request ID

Without request IDs, debugging becomes hard.

Always pass:

```text
request_id
trace_id
pipeline_run
```

through logs and messages.

---

## Mistake 6: Planner does not detect cycles

A pipeline dependency graph can be invalid.

Example:

```text
build -> deploy
deploy -> build
```

Your planner must reject this.

---

# 18. Final debugging checklist

When something fails, check in this order:

```text
1. Did the CLI receive correct flags?

2. Did request validation pass?

3. Is request_id visible in logs?

4. Was Slack started notification attempted?

5. Was dependency planning successful?

6. Was Tekton trigger command built correctly?

7. Did tkn/kubectl command return output?

8. Was PipelineRun created?

9. What is PipelineRun status?

10. Which TaskRun failed?

11. Which Pod failed?

12. What do pod logs say?

13. What do Kubernetes events say?

14. Was failure trace captured?

15. Was Slack failure notification sent?

16. Did the CLI return a useful final message?
```

Useful commands:

```bash
kubectl get pipelineruns -n dev
kubectl describe pipelinerun <name> -n dev
kubectl get taskruns -n dev
kubectl get pods -n dev
kubectl logs <pod-name> -n dev
kubectl get events -n dev --sort-by=.lastTimestamp
```

Tekton CLI commands:

```bash
tkn pipelinerun list -n dev
tkn pipelinerun describe <name> -n dev
tkn pipelinerun logs <name> -n dev
```

---

# 19. DSA topic: Topological sort

Topological sort sounds scary, but the idea is simple.

It answers this question:

> Given tasks with dependencies, what order should we run them in?

Example:

```text
validate must run before build
build must run before deploy
deploy must run before notify
```

Valid order:

```text
validate -> build -> deploy -> notify
```

This is exactly how pipelines work.

Tekton example:

```yaml
tasks:
  - name: validate

  - name: build
    runAfter:
      - validate

  - name: deploy
    runAfter:
      - build

  - name: notify
    runAfter:
      - deploy
```

## Graph thinking

Tasks are nodes:

```text
validate
build
deploy
notify
```

Dependencies are arrows:

```text
validate -> build -> deploy -> notify
```

Topological sort returns an order that respects arrows.

## Cycle problem

This is invalid:

```text
build -> deploy
deploy -> build
```

Why?

Because:

```text
build waits for deploy
deploy waits for build
```

Nobody can start.

So topological sort also helps detect impossible pipelines.

## Simple algorithm: Kahn’s algorithm

Beginner version:

```text
1. Count how many dependencies each task has.

2. Start with tasks that have zero dependencies.

3. Remove one zero-dependency task from the queue.

4. Add it to the result.

5. Reduce dependency count of tasks that depend on it.

6. If new tasks now have zero dependencies, add them to the queue.

7. If result contains all tasks, success.

8. If not, there is a cycle.
```

Pipeline meaning:

```text
zero dependencies = task can run now
```

---

# 20. One Go DSA problem

## Problem: Pipeline execution order

You are given pipeline tasks and their dependencies.

Return a valid execution order.

If there is a cycle, return an error.

Example input:

```go
tasks := []Task{
    {Name: "validate", DependsOn: []string{}},
    {Name: "build", DependsOn: []string{"validate"}},
    {Name: "deploy", DependsOn: []string{"build"}},
    {Name: "notify", DependsOn: []string{"deploy"}},
}
```

Expected output:

```text
validate, build, deploy, notify
```

## Go solution

```go
package planner

import "fmt"

type Task struct {
    Name      string
    DependsOn []string
}

func ExecutionOrder(tasks []Task) ([]string, error) {
    indegree := make(map[string]int)
    graph := make(map[string][]string)

    for _, task := range tasks {
        if _, exists := indegree[task.Name]; !exists {
            indegree[task.Name] = 0
        }
    }

    for _, task := range tasks {
        for _, dependency := range task.DependsOn {
            graph[dependency] = append(graph[dependency], task.Name)
            indegree[task.Name]++
        }
    }

    queue := make([]string, 0)

    for taskName, count := range indegree {
        if count == 0 {
            queue = append(queue, taskName)
        }
    }

    order := make([]string, 0, len(tasks))

    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]

        order = append(order, current)

        for _, next := range graph[current] {
            indegree[next]--

            if indegree[next] == 0 {
                queue = append(queue, next)
            }
        }
    }

    if len(order) != len(tasks) {
        return nil, fmt.Errorf("cycle detected in pipeline dependencies")
    }

    return order, nil
}
```

Python equivalent:

```python
from collections import defaultdict, deque

def execution_order(tasks):
    indegree = {}
    graph = defaultdict(list)

    for task in tasks:
        indegree[task["name"]] = 0

    for task in tasks:
        for dep in task["depends_on"]:
            graph[dep].append(task["name"])
            indegree[task["name"]] += 1

    queue = deque([name for name, count in indegree.items() if count == 0])
    order = []

    while queue:
        current = queue.popleft()
        order.append(current)

        for nxt in graph[current]:
            indegree[nxt] -= 1
            if indegree[nxt] == 0:
                queue.append(nxt)

    if len(order) != len(tasks):
        raise ValueError("cycle detected")

    return order
```

Key Go syntax differences:

| Idea              | Python             | Go                            |
| ----------------- | ------------------ | ----------------------------- |
| dictionary        | `dict`             | `map[string]int`              |
| list              | `list`             | `[]string`                    |
| append            | `list.append(x)`   | `slice = append(slice, x)`    |
| error             | `raise ValueError` | `return nil, fmt.Errorf(...)` |
| loop              | `for x in items`   | `for _, x := range items`     |
| empty queue check | `while queue:`     | `for len(queue) > 0`          |

---

# 21. Final module-based practice task

Build:

```text
internal/planner
```

## Goal

Create a pipeline dependency tracker/execution planner.

It should:

1. accept tasks with dependencies
2. return valid execution order
3. detect cycles
4. log the plan
5. connect to router before Tekton trigger

## Suggested files

```text
internal/planner/task.go
internal/planner/planner.go
internal/planner/planner_test.go
```

## `task.go`

```go
package planner

type Task struct {
    Name      string
    DependsOn []string
}
```

## `planner.go`

```go
package planner

import "fmt"

type Planner struct{}

func New() Planner {
    return Planner{}
}

func (p Planner) Plan(tasks []Task) ([]string, error) {
    indegree := make(map[string]int)
    graph := make(map[string][]string)

    for _, task := range tasks {
        if task.Name == "" {
            return nil, fmt.Errorf("task name is required")
        }

        indegree[task.Name] = 0
    }

    for _, task := range tasks {
        for _, dep := range task.DependsOn {
            if _, exists := indegree[dep]; !exists {
                return nil, fmt.Errorf("unknown dependency %q for task %q", dep, task.Name)
            }

            graph[dep] = append(graph[dep], task.Name)
            indegree[task.Name]++
        }
    }

    queue := make([]string, 0)

    for name, count := range indegree {
        if count == 0 {
            queue = append(queue, name)
        }
    }

    order := make([]string, 0, len(tasks))

    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]

        order = append(order, current)

        for _, next := range graph[current] {
            indegree[next]--

            if indegree[next] == 0 {
                queue = append(queue, next)
            }
        }
    }

    if len(order) != len(tasks) {
        return nil, fmt.Errorf("cycle detected in pipeline dependencies")
    }

    return order, nil
}
```

## `planner_test.go`

```go
package planner_test

import (
    "testing"

    "slack-integration/internal/planner"
)

func TestPlanSuccess(t *testing.T) {
    p := planner.New()

    tasks := []planner.Task{
        {Name: "validate"},
        {Name: "build", DependsOn: []string{"validate"}},
        {Name: "deploy", DependsOn: []string{"build"}},
        {Name: "notify", DependsOn: []string{"deploy"}},
    }

    got, err := p.Plan(tasks)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    want := []string{"validate", "build", "deploy", "notify"}

    if len(got) != len(want) {
        t.Fatalf("expected %d tasks, got %d", len(want), len(got))
    }

    for i := range want {
        if got[i] != want[i] {
            t.Fatalf("at index %d expected %s, got %s", i, want[i], got[i])
        }
    }
}

func TestPlanCycle(t *testing.T) {
    p := planner.New()

    tasks := []planner.Task{
        {Name: "build", DependsOn: []string{"deploy"}},
        {Name: "deploy", DependsOn: []string{"build"}},
    }

    _, err := p.Plan(tasks)
    if err == nil {
        t.Fatal("expected cycle error, got nil")
    }
}
```

## Connect planner to router

Add interface:

```go
type ExecutionPlanner interface {
    Plan(tasks []planner.Task) ([]string, error)
}
```

Router flow becomes:

```text
validate request
plan dependencies
send plan Slack message
trigger Tekton
wait for result
send final Slack message
```

This is a complete module-based enhancement.

---

# 22. Final revision checklist

Use this as your Day 15 completion checklist.

## Go basics

* [ ] I understand `package main`
* [ ] I understand `func main()`
* [ ] I understand structs
* [ ] I understand methods
* [ ] I understand interfaces
* [ ] I understand slices
* [ ] I understand maps
* [ ] I understand `context.Context`
* [ ] I understand explicit error returns

## Project structure

* [ ] CLI only parses input
* [ ] model defines data
* [ ] router coordinates flow
* [ ] Slack package sends notifications
* [ ] shell package runs commands
* [ ] Tekton package triggers pipelines
* [ ] Kubernetes package captures traces
* [ ] planner package orders dependencies
* [ ] logger package creates zerolog logger

## End-to-end flow

* [ ] CLI creates request
* [ ] request is validated
* [ ] router sends start notification
* [ ] planner creates execution order
* [ ] Tekton pipeline is triggered
* [ ] Kubernetes resources are created
* [ ] result is checked
* [ ] success notification is sent
* [ ] failure trace is captured when needed
* [ ] failure notification is sent when needed

## Testing

* [ ] model validation test exists
* [ ] planner success test exists
* [ ] planner cycle test exists
* [ ] router success test exists
* [ ] router failure test exists
* [ ] fake Slack client used in tests
* [ ] fake Tekton trigger used in tests
* [ ] fake shell runner used in tests

## Production thinking

* [ ] logs include request ID
* [ ] errors are wrapped
* [ ] shell input is safe
* [ ] secrets are not logged
* [ ] timeouts exist
* [ ] Slack failure does not hide pipeline failure
* [ ] failure trace is useful

---

# 23. Next-step learning suggestions

You are now at a very good point.

Your next learning path can be:

## 1. Go depth

Focus on:

* interfaces
* context
* goroutines
* channels
* testing
* table-driven tests
* dependency injection
* error wrapping

Especially learn table-driven tests:

```go
tests := []struct {
    name    string
    input   []Task
    wantErr bool
}{
    {
        name: "valid chain",
        input: []Task{
            {Name: "validate"},
            {Name: "build", DependsOn: []string{"validate"}},
        },
        wantErr: false,
    },
}
```

This is a very common Go testing convention.

---

## 2. Kubernetes client-go

After shell-based learning, move toward native Kubernetes clients.

Instead of:

```bash
kubectl get pods
```

your app can call Kubernetes APIs directly.

This is more production-like.

---

## 3. Tekton controller/operator style

Later, you can build a small controller that watches custom resources.

That would move you from:

```text
CLI triggers pipeline
```

to:

```text
Kubernetes resource triggers pipeline automatically
```

That is how many real cloud-native systems work.

---

## 4. Better Slack app integration

Instead of only incoming webhooks, explore:

* Slack slash commands
* Slack interactive buttons
* approval workflows
* deployment approval from Slack

Example future flow:

```text
User clicks "Approve Deploy" in Slack
  -> Go service receives Slack event
  -> Tekton pipeline starts
  -> Slack receives status update
```

---

## 5. More DSA connected to systems

Useful DSA topics for backend/cloud work:

| DSA topic             | Real project connection       |
| --------------------- | ----------------------------- |
| Topological sort      | pipeline dependency order     |
| BFS/DFS               | resource traversal            |
| Hash maps             | fast lookup of services/tasks |
| Queues                | async job processing          |
| Heaps                 | priority scheduling           |
| Tries                 | command autocomplete          |
| Union-find            | dependency grouping           |
| Graph cycle detection | invalid pipelines             |

---

# Final encouragement

You have now reached the most important stage of project learning:

> seeing how small pieces become a real system.

The big lesson from Day 15 is not just Go syntax.

It is this:

```text
A production-style backend is a chain of small, testable modules.
```

Your final mental model should be:

```text
CLI receives intent
Model structures it
Router coordinates it
Slack communicates it
Tekton executes it
Kubernetes runs it
Trace captures failures
Logs explain everything
Tests protect behavior
```

That is a real engineering mindset.
