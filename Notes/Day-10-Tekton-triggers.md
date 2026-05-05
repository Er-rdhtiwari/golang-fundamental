# Day 10 — Tekton Triggers + PR Webhook Flow

Today’s mental model:

> **Yesterday, you manually started a PipelineRun. Today, GitHub/Postman will send an event, and Tekton will automatically create the PipelineRun.**

Tekton Triggers is the event-driven part of Tekton. It listens for external events, extracts useful fields, and creates Tekton resources such as `PipelineRun`. Official Tekton docs describe `TriggerBinding` as the resource that extracts fields from an event payload and binds them to params, and `TriggerTemplate` as the blueprint for creating resources like `PipelineRun`. ([Tekton][1])

---

## 1. Day 10 learning goals

By the end of Day 10, you should understand:

1. What Tekton Triggers are.
2. What `EventListener`, `TriggerBinding`, and `TriggerTemplate` do.
3. How GitHub/Postman webhook JSON becomes Tekton params.
4. How Tekton creates a `PipelineRun` automatically.
5. Where PR number, branch, sender, commit ID, and repo URL come from.
6. How this later connects to Slack notifications.
7. Graph basics and BFS in Go.
8. How to build a small webhook event parser in Go.

---

## 2. Quick revision of Days 1 to 9

### Days 1–4: Go Slack notifier foundation

You learned:

```text
CLI flags → PipelineEvent struct → validation → routing → Slack payload → webhook
```

Your Go app receives values like:

```bash
--event-type pr
--status failed
--pipeline-name pr-validation
--failed-step go-test
--sender rdh-tiwari
```

Then it builds a structured event and sends a Slack message.

---

### Days 5–6: Errors, logging, and tests

You learned that production code should not only “work”; it should also be:

```text
safe → testable → debuggable → observable
```

For example:

```go
if err != nil {
    logger.Error().Err(err).Msg("failed to send slack notification")
    return err
}
```

Python equivalent:

```python
try:
    send_slack_message()
except Exception as e:
    logger.error("failed to send slack notification", exc_info=e)
```

---

### Day 7: Shell scripting

You learned that shell scripts wrap repeated commands:

```bash
go test ./...
go build ./...
kubectl apply -f .tekton/
```

Shell scripts are useful because CI/CD systems run commands non-interactively.

---

### Day 8: Kubernetes basics

You learned:

```text
Cluster → Namespace → Pod → Container
```

Tekton runs on top of Kubernetes. A Tekton `TaskRun` eventually creates Kubernetes Pods.

---

### Day 9: Tekton basics

You learned:

```text
Task        = one reusable job
Pipeline    = ordered collection of tasks
PipelineRun = actual execution of a pipeline
TaskRun     = actual execution of a task
```

A `PipelineRun` instantiates and executes a `Pipeline`; Tekton automatically creates corresponding `TaskRuns` for the tasks in that pipeline. ([Tekton][2])

---

## 3. Explain Tekton Triggers in very simple language

Imagine your CI/CD system is a security gate.

Without triggers:

```text
You manually press the button → Pipeline starts
```

With triggers:

```text
GitHub sends webhook → Tekton receives it → Pipeline starts automatically
```

So Tekton Triggers means:

> “Start my pipeline when something happens outside Kubernetes.”

Examples:

```text
PR opened      → run validation pipeline
PR updated     → run tests again
Code merged    → run CD pipeline
Job completed  → send Slack notification
```

For your `slack-integration` project:

```text
GitHub PR event
   ↓
Tekton Trigger receives event
   ↓
Extracts PR number, branch, sender, commit ID
   ↓
Creates PR PipelineRun
   ↓
Pipeline runs validation/build
   ↓
Slack notification task sends result
```

---

## 4. EventListener, TriggerBinding, TriggerTemplate

Tekton Triggers has three beginner-important parts:

```text
EventListener    = receiver
TriggerBinding   = extractor
TriggerTemplate  = creator
```

### 4.1 EventListener

`EventListener` is like an HTTP server inside Kubernetes.

It waits for webhook requests.

```text
GitHub/Postman sends POST request
        ↓
EventListener receives it
```

Tekton creates Kubernetes resources behind the scenes for an `EventListener`, including a Deployment and Service, commonly prefixed with `el-`. ([GitHub][3])

Simple meaning:

> EventListener says: “I am ready to receive webhook events.”

---

### 4.2 TriggerBinding

`TriggerBinding` extracts values from webhook JSON.

Example GitHub sends this:

```json
{
  "pull_request": {
    "number": 42,
    "head": {
      "ref": "feature/add-validation",
      "sha": "abc123"
    },
    "base": {
      "ref": "main"
    }
  },
  "sender": {
    "login": "rdh-tiwari"
  }
}
```

TriggerBinding maps it like this:

```yaml
- name: pr-number
  value: $(body.pull_request.number)

- name: commit-id
  value: $(body.pull_request.head.sha)

- name: sender
  value: $(body.sender.login)
```

Simple meaning:

> TriggerBinding says: “From this big JSON, pick only the fields my pipeline needs.”

---

### 4.3 TriggerTemplate

`TriggerTemplate` uses the extracted values and creates a `PipelineRun`.

Simple meaning:

> TriggerTemplate says: “Use these params and create this PipelineRun.”

Example:

```yaml
params:
  - name: pr-number
    value: $(tt.params.pr-number)
```

This means:

```text
Use the PR number extracted by TriggerBinding
and pass it into the PipelineRun.
```

---

## 5. How GitHub/Postman webhook becomes a PipelineRun

### Manual PipelineRun

Yesterday’s style:

```bash
kubectl apply -f pr-pipelinerun.yaml
```

Flow:

```text
You create PipelineRun manually
        ↓
Tekton runs Pipeline
        ↓
Tasks execute
```

---

### Webhook-triggered PipelineRun

Today’s style:

```text
GitHub sends PR webhook
        ↓
EventListener receives webhook
        ↓
TriggerBinding extracts JSON fields
        ↓
TriggerTemplate creates PipelineRun
        ↓
Pipeline tasks execute
```

Main difference:

| Manual PipelineRun         | Webhook-triggered PipelineRun   |
| -------------------------- | ------------------------------- |
| You create it manually     | GitHub/Postman event creates it |
| Params are written in YAML | Params come from webhook JSON   |
| Good for local testing     | Good for real CI/CD             |
| Static values              | Dynamic values per PR           |

---

## 6. Full trigger flow in ASCII

```text
Developer opens or updates PR
            |
            v
+-------------------------+
|        GitHub           |
| PR webhook JSON payload |
+-------------------------+
            |
            | HTTP POST
            v
+-------------------------+
| Tekton EventListener    |
| Receives webhook event  |
+-------------------------+
            |
            v
+-------------------------+
| TriggerBinding          |
| Extracts values:        |
| - PR number             |
| - commit ID             |
| - source branch         |
| - target branch         |
| - sender                |
| - repo URL              |
+-------------------------+
            |
            v
+-------------------------+
| TriggerTemplate         |
| Creates PipelineRun     |
| using extracted params  |
+-------------------------+
            |
            v
+-------------------------+
| PipelineRun             |
| Runs PR validation      |
+-------------------------+
            |
            v
+-------------------------+
| Tasks                   |
| - clone repo            |
| - validate code         |
| - run tests             |
| - send Slack message    |
+-------------------------+
            |
            v
+-------------------------+
| Slack Notification      |
| PR #42 failed/succeeded |
+-------------------------+
```

---

## 7. JSON body path mapping

Tekton lets you read fields from the incoming webhook body using expressions like:

```text
$(body.pull_request.number)
```

Think of it like accessing nested dictionaries in Python.

### GitHub JSON

```json
{
  "pull_request": {
    "number": 42,
    "head": {
      "ref": "feature/add-validation",
      "sha": "abc123"
    },
    "base": {
      "ref": "main"
    }
  },
  "sender": {
    "login": "rdh-tiwari"
  }
}
```

### Python equivalent

```python
pr_number = body["pull_request"]["number"]
commit_id = body["pull_request"]["head"]["sha"]
source_branch = body["pull_request"]["head"]["ref"]
target_branch = body["pull_request"]["base"]["ref"]
sender = body["sender"]["login"]
```

### Tekton equivalent

```yaml
$(body.pull_request.number)
$(body.pull_request.head.sha)
$(body.pull_request.head.ref)
$(body.pull_request.base.ref)
$(body.sender.login)
```

---

## 8. Pseudocode first for trigger flow

```text
START

GitHub sends webhook JSON to EventListener

EventListener receives request

TriggerBinding reads JSON:
    pr_number      = body.pull_request.number
    commit_id      = body.pull_request.head.sha
    source_branch  = body.pull_request.head.ref
    target_branch  = body.pull_request.base.ref
    sender         = body.sender.login
    repo_url       = body.repository.clone_url

TriggerTemplate creates PipelineRun:
    pass pr_number
    pass commit_id
    pass source_branch
    pass target_branch
    pass sender
    pass repo_url

PipelineRun starts PR pipeline

Pipeline runs tasks:
    validate code
    run tests
    build app
    send Slack notification

END
```

---

## 9. Real Tekton YAML examples

Below is a simple PR trigger setup.

Recommended files:

```text
.tekton/
  pr-binding.yaml
  pr-trigger-template.yaml
  pr-listener.yaml
  pr-pipeline.yaml
```

---

### 9.1 `pr-binding.yaml`

```yaml
apiVersion: triggers.tekton.dev/v1beta1
kind: TriggerBinding
metadata:
  name: pr-binding
  namespace: slack-integration-dev
spec:
  params:
    - name: repo-url
      value: $(body.repository.clone_url)

    - name: repo-full-name
      value: $(body.repository.full_name)

    - name: pr-number
      value: $(body.pull_request.number)

    - name: commit-id
      value: $(body.pull_request.head.sha)

    - name: source-branch
      value: $(body.pull_request.head.ref)

    - name: target-branch
      value: $(body.pull_request.base.ref)

    - name: sender
      value: $(body.sender.login)

    - name: action
      value: $(body.action)
```

### Explanation

```yaml
kind: TriggerBinding
```

This says:

```text
This file extracts data from webhook JSON.
```

```yaml
value: $(body.pull_request.number)
```

This means:

```text
Read pull_request.number from the incoming JSON body.
```

For GitHub PR events:

```text
PR number     → body.pull_request.number
commit ID     → body.pull_request.head.sha
source branch → body.pull_request.head.ref
target branch → body.pull_request.base.ref
sender        → body.sender.login
repo URL      → body.repository.clone_url
```

---

### 9.2 `pr-trigger-template.yaml`

```yaml
apiVersion: triggers.tekton.dev/v1beta1
kind: TriggerTemplate
metadata:
  name: pr-trigger-template
  namespace: slack-integration-dev
spec:
  params:
    - name: repo-url
    - name: repo-full-name
    - name: pr-number
    - name: commit-id
    - name: source-branch
    - name: target-branch
    - name: sender
    - name: action

  resourcetemplates:
    - apiVersion: tekton.dev/v1
      kind: PipelineRun
      metadata:
        generateName: pr-validation-run-
      spec:
        pipelineRef:
          name: pr-validation-pipeline

        params:
          - name: repo-url
            value: $(tt.params.repo-url)

          - name: repo-full-name
            value: $(tt.params.repo-full-name)

          - name: pr-number
            value: $(tt.params.pr-number)

          - name: commit-id
            value: $(tt.params.commit-id)

          - name: source-branch
            value: $(tt.params.source-branch)

          - name: target-branch
            value: $(tt.params.target-branch)

          - name: sender
            value: $(tt.params.sender)

          - name: action
            value: $(tt.params.action)

        workspaces:
          - name: shared-workspace
            emptyDir: {}
```

### Explanation

```yaml
kind: TriggerTemplate
```

This means:

```text
This file defines what should be created after the webhook is received.
```

Here it creates:

```yaml
kind: PipelineRun
```

The `generateName` field means Kubernetes will generate a unique name like:

```text
pr-validation-run-x7abc
```

This is useful because every webhook should create a new PipelineRun.

---

### Important syntax

```yaml
$(tt.params.pr-number)
```

This means:

```text
Take the pr-number value received from TriggerBinding
and pass it into the PipelineRun.
```

---

### 9.3 `pr-listener.yaml`

```yaml
apiVersion: triggers.tekton.dev/v1beta1
kind: EventListener
metadata:
  name: pr-listener
  namespace: slack-integration-dev
spec:
  serviceAccountName: default

  triggers:
    - name: pr-trigger
      bindings:
        - ref: pr-binding
      template:
        ref: pr-trigger-template
```

### Explanation

```yaml
kind: EventListener
```

This creates the webhook receiver.

```yaml
bindings:
  - ref: pr-binding
```

This says:

```text
Use pr-binding to extract JSON values.
```

```yaml
template:
  ref: pr-trigger-template
```

This says:

```text
Use pr-trigger-template to create a PipelineRun.
```

---

### 9.4 `pr-pipeline.yaml`

```yaml
apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: pr-validation-pipeline
  namespace: slack-integration-dev
spec:
  params:
    - name: repo-url
      type: string
    - name: repo-full-name
      type: string
    - name: pr-number
      type: string
    - name: commit-id
      type: string
    - name: source-branch
      type: string
    - name: target-branch
      type: string
    - name: sender
      type: string
    - name: action
      type: string

  workspaces:
    - name: shared-workspace

  tasks:
    - name: print-pr-info
      taskSpec:
        params:
          - name: repo-url
          - name: pr-number
          - name: commit-id
          - name: source-branch
          - name: target-branch
          - name: sender
          - name: action
        steps:
          - name: print
            image: alpine:3.19
            script: |
              #!/bin/sh
              echo "PR Event Received"
              echo "Repo URL: $(params.repo-url)"
              echo "PR Number: $(params.pr-number)"
              echo "Commit ID: $(params.commit-id)"
              echo "Source Branch: $(params.source-branch)"
              echo "Target Branch: $(params.target-branch)"
              echo "Sender: $(params.sender)"
              echo "Action: $(params.action)"
      params:
        - name: repo-url
          value: $(params.repo-url)
        - name: pr-number
          value: $(params.pr-number)
        - name: commit-id
          value: $(params.commit-id)
        - name: source-branch
          value: $(params.source-branch)
        - name: target-branch
          value: $(params.target-branch)
        - name: sender
          value: $(params.sender)
        - name: action
          value: $(params.action)
```

This simple pipeline only prints values. Later, you can replace or extend it with:

```text
clone repo → validate Go code → run tests → build → Slack notify
```

---

## 10. Sample webhook JSON and mapping explanation

Use this JSON for Postman or `curl` testing:

```json
{
  "action": "opened",
  "repository": {
    "clone_url": "https://github.com/example/slack-integration.git",
    "full_name": "example/slack-integration"
  },
  "pull_request": {
    "number": 42,
    "head": {
      "ref": "feature/add-validation",
      "sha": "abc123def456"
    },
    "base": {
      "ref": "main"
    },
    "title": "Add validation logic"
  },
  "sender": {
    "login": "rdh-tiwari"
  }
}
```

### Mapping

| Required value | GitHub JSON path        | Tekton expression               |
| -------------- | ----------------------- | ------------------------------- |
| PR number      | `pull_request.number`   | `$(body.pull_request.number)`   |
| Commit ID      | `pull_request.head.sha` | `$(body.pull_request.head.sha)` |
| Source branch  | `pull_request.head.ref` | `$(body.pull_request.head.ref)` |
| Target branch  | `pull_request.base.ref` | `$(body.pull_request.base.ref)` |
| Sender         | `sender.login`          | `$(body.sender.login)`          |
| Repo URL       | `repository.clone_url`  | `$(body.repository.clone_url)`  |
| Action         | `action`                | `$(body.action)`                |

---

## 11. Hands-on tasks

### Task 1: Create namespace

```bash
kubectl create namespace slack-integration-dev
```

---

### Task 2: Apply YAML files

```bash
kubectl apply -f .tekton/pr-pipeline.yaml
kubectl apply -f .tekton/pr-binding.yaml
kubectl apply -f .tekton/pr-trigger-template.yaml
kubectl apply -f .tekton/pr-listener.yaml
```

---

### Task 3: Check resources

```bash
kubectl get eventlistener -n slack-integration-dev
kubectl get triggerbinding -n slack-integration-dev
kubectl get triggertemplate -n slack-integration-dev
kubectl get pipeline -n slack-integration-dev
```

Expected:

```text
pr-listener
pr-binding
pr-trigger-template
pr-validation-pipeline
```

---

### Task 4: Port-forward EventListener service

The EventListener usually creates a service named like:

```text
el-pr-listener
```

Check service:

```bash
kubectl get svc -n slack-integration-dev
```

Then port-forward:

```bash
kubectl port-forward svc/el-pr-listener 8080:8080 -n slack-integration-dev
```

---

### Task 5: Send webhook using curl

Open another terminal:

```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: pull_request" \
  -d '{
    "action": "opened",
    "repository": {
      "clone_url": "https://github.com/example/slack-integration.git",
      "full_name": "example/slack-integration"
    },
    "pull_request": {
      "number": 42,
      "head": {
        "ref": "feature/add-validation",
        "sha": "abc123def456"
      },
      "base": {
        "ref": "main"
      },
      "title": "Add validation logic"
    },
    "sender": {
      "login": "rdh-tiwari"
    }
  }'
```

---

### Task 6: Watch PipelineRun

```bash
kubectl get pipelinerun -n slack-integration-dev
```

Then:

```bash
kubectl get taskrun -n slack-integration-dev
kubectl get pods -n slack-integration-dev
```

Logs:

```bash
kubectl logs -l tekton.dev/pipelineRun=<PIPELINERUN_NAME> -n slack-integration-dev
```

Or with Tekton CLI:

```bash
tkn pipelinerun logs -f -n slack-integration-dev
```

---

## 12. Expected output

When the pipeline runs, logs should show:

```text
PR Event Received
Repo URL: https://github.com/example/slack-integration.git
PR Number: 42
Commit ID: abc123def456
Source Branch: feature/add-validation
Target Branch: main
Sender: rdh-tiwari
Action: opened
```

This proves:

```text
Webhook JSON → TriggerBinding → TriggerTemplate → PipelineRun → Task logs
```

---

## 13. Common mistakes

### Mistake 1: Wrong API group

Wrong:

```yaml
apiVersion: tekton.dev/v1beta1
kind: TriggerBinding
```

Better for triggers:

```yaml
apiVersion: triggers.tekton.dev/v1beta1
kind: TriggerBinding
```

Core Tekton Pipeline resources use `tekton.dev`, while trigger resources use `triggers.tekton.dev`.

---

### Mistake 2: Wrong JSON path

Wrong:

```yaml
$(body.pr.number)
```

Correct for GitHub PR payload:

```yaml
$(body.pull_request.number)
```

---

### Mistake 3: Forgetting namespace

You apply files to one namespace but check another namespace.

Use:

```bash
kubectl get pipelinerun -n slack-integration-dev
```

---

### Mistake 4: EventListener service not found

Check:

```bash
kubectl get svc -n slack-integration-dev
```

Look for:

```text
el-pr-listener
```

---

### Mistake 5: PipelineRun created but task fails

Then the trigger worked. Now debug the pipeline/task.

Check:

```bash
kubectl describe pipelinerun <name> -n slack-integration-dev
kubectl describe taskrun <name> -n slack-integration-dev
kubectl logs <pod-name> -n slack-integration-dev
```

---

### Mistake 6: Expecting Slack notification too early

First verify this flow:

```text
Webhook → PipelineRun → printed params
```

Then add:

```text
Slack notify task
```

This keeps debugging simple.

---

## 14. Debugging tips

### Check trigger resources

```bash
kubectl get eventlistener -n slack-integration-dev
kubectl get triggerbinding -n slack-integration-dev
kubectl get triggertemplate -n slack-integration-dev
```

---

### Describe EventListener

```bash
kubectl describe eventlistener pr-listener -n slack-integration-dev
```

---

### Check EventListener pod

```bash
kubectl get pods -n slack-integration-dev
```

Look for a pod related to:

```text
el-pr-listener
```

Logs:

```bash
kubectl logs -l eventlistener=pr-listener -n slack-integration-dev
```

---

### Check if PipelineRun was created

```bash
kubectl get pipelinerun -n slack-integration-dev
```

If no PipelineRun appears, problem is likely in:

```text
EventListener / TriggerBinding / TriggerTemplate / webhook request
```

If PipelineRun appears but fails, problem is likely in:

```text
Pipeline / Task / image / script / workspace / params
```

---

## 15. One DSA topic — Graph basics

A graph is a collection of:

```text
nodes + edges
```

Example:

```text
A --- B
|     |
C --- D
```

Here:

```text
Nodes = A, B, C, D
Edges = A-B, A-C, B-D, C-D
```

### Real-world examples

| Real world             | Graph meaning                       |
| ---------------------- | ----------------------------------- |
| Cities and roads       | Cities = nodes, roads = edges       |
| People and friendships | People = nodes, friendships = edges |
| Web pages and links    | Pages = nodes, links = edges        |
| CI/CD tasks            | Tasks = nodes, dependencies = edges |

Tekton Pipeline is also graph-like:

```text
clone → test → build → notify
```

Each task is like a node. Dependency between tasks is like an edge.

---

### Graph representation

In Python, you may write:

```python
graph = {
    "A": ["B", "C"],
    "B": ["A", "D"],
    "C": ["A", "D"],
    "D": ["B", "C"],
}
```

In Go:

```go
graph := map[string][]string{
    "A": {"B", "C"},
    "B": {"A", "D"},
    "C": {"A", "D"},
    "D": {"B", "C"},
}
```

But in real Go syntax, for readability:

```go
graph := map[string][]string{
    "A": []string{"B", "C"},
    "B": []string{"A", "D"},
    "C": []string{"A", "D"},
    "D": []string{"B", "C"},
}
```

Newer Go also allows the shorter composite form in many cases:

```go
graph := map[string][]string{
    "A": {"B", "C"},
}
```

But as a beginner, use the explicit version first.

---

### BFS meaning

BFS means **Breadth-First Search**.

Simple idea:

> Visit nearby nodes first, then go deeper.

BFS uses a queue.

```text
Queue = first in, first out
```

Python equivalent:

```python
from collections import deque
queue = deque(["A"])
```

Go equivalent:

```go
queue := []string{"A"}
```

---

## 16. One Go DSA problem — Easy BFS

### Problem

Given a graph and a starting node, return the BFS visiting order.

Example:

```text
Graph:
A -> B, C
B -> D
C -> D
D -> E

Start: A

Output:
A B C D E
```

---

### Go solution

```go
package main

import "fmt"

func bfs(graph map[string][]string, start string) []string {
	visited := make(map[string]bool)
	queue := []string{start}
	order := []string{}

	visited[start] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		order = append(order, current)

		for _, neighbor := range graph[current] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	return order
}

func main() {
	graph := map[string][]string{
		"A": []string{"B", "C"},
		"B": []string{"D"},
		"C": []string{"D"},
		"D": []string{"E"},
		"E": []string{},
	}

	result := bfs(graph, "A")
	fmt.Println(result)
}
```

### Output

```text
[A B C D E]
```

### Python comparison

Python:

```python
visited = set()
queue = ["A"]
```

Go:

```go
visited := make(map[string]bool)
queue := []string{"A"}
```

Python list append:

```python
queue.append(neighbor)
```

Go slice append:

```go
queue = append(queue, neighbor)
```

Python loop:

```python
for neighbor in graph[current]:
```

Go loop:

```go
for _, neighbor := range graph[current] {
}
```

Key Go convention shift:

> Go prefers explicit types, explicit error handling, and simple loops.

---

## 17. Module-based practice task — Build webhook event parser in Go

### Goal

Build a small Go module that receives GitHub PR webhook JSON and extracts:

```text
PR number
commit ID
source branch
target branch
sender
repo URL
action
```

This directly matches today’s Tekton TriggerBinding logic.

---

### Folder structure

```text
webhook-parser/
  go.mod
  main.go
```

---

### `go.mod`

```go
module webhook-parser

go 1.22
```

---

### `main.go`

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
)

type GitHubPREvent struct {
	Action     string     `json:"action"`
	Repository Repository `json:"repository"`
	PullRequest PullRequest `json:"pull_request"`
	Sender     Sender     `json:"sender"`
}

type Repository struct {
	CloneURL string `json:"clone_url"`
	FullName string `json:"full_name"`
}

type PullRequest struct {
	Number int    `json:"number"`
	Head   Branch `json:"head"`
	Base   Branch `json:"base"`
	Title  string `json:"title"`
}

type Branch struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

type Sender struct {
	Login string `json:"login"`
}

type PipelineEvent struct {
	EventType    string
	Action       string
	RepoURL      string
	RepoFullName string
	PRNumber     int
	CommitID     string
	SourceBranch string
	TargetBranch string
	Sender        string
}

func ConvertToPipelineEvent(event GitHubPREvent) PipelineEvent {
	return PipelineEvent{
		EventType:    "pr",
		Action:       event.Action,
		RepoURL:      event.Repository.CloneURL,
		RepoFullName: event.Repository.FullName,
		PRNumber:     event.PullRequest.Number,
		CommitID:     event.PullRequest.Head.SHA,
		SourceBranch: event.PullRequest.Head.Ref,
		TargetBranch: event.PullRequest.Base.Ref,
		Sender:        event.Sender.Login,
	}
}

func main() {
	rawJSON := []byte(`{
		"action": "opened",
		"repository": {
			"clone_url": "https://github.com/example/slack-integration.git",
			"full_name": "example/slack-integration"
		},
		"pull_request": {
			"number": 42,
			"head": {
				"ref": "feature/add-validation",
				"sha": "abc123def456"
			},
			"base": {
				"ref": "main",
				"sha": "base456"
			},
			"title": "Add validation logic"
		},
		"sender": {
			"login": "rdh-tiwari"
		}
	}`)

	var githubEvent GitHubPREvent

	err := json.Unmarshal(rawJSON, &githubEvent)
	if err != nil {
		log.Fatal("failed to parse webhook JSON:", err)
	}

	pipelineEvent := ConvertToPipelineEvent(githubEvent)

	fmt.Println("Pipeline Event Created")
	fmt.Println("Event Type:", pipelineEvent.EventType)
	fmt.Println("Action:", pipelineEvent.Action)
	fmt.Println("Repo URL:", pipelineEvent.RepoURL)
	fmt.Println("Repo Full Name:", pipelineEvent.RepoFullName)
	fmt.Println("PR Number:", pipelineEvent.PRNumber)
	fmt.Println("Commit ID:", pipelineEvent.CommitID)
	fmt.Println("Source Branch:", pipelineEvent.SourceBranch)
	fmt.Println("Target Branch:", pipelineEvent.TargetBranch)
	fmt.Println("Sender:", pipelineEvent.Sender)
}
```

---

### Expected output

```text
Pipeline Event Created
Event Type: pr
Action: opened
Repo URL: https://github.com/example/slack-integration.git
Repo Full Name: example/slack-integration
PR Number: 42
Commit ID: abc123def456
Source Branch: feature/add-validation
Target Branch: main
Sender: rdh-tiwari
```

---

### Go vs Python comparison

Python version would look like:

```python
import json

data = json.loads(raw_json)

pr_number = data["pull_request"]["number"]
commit_id = data["pull_request"]["head"]["sha"]
sender = data["sender"]["login"]
```

Go version uses structs:

```go
type GitHubPREvent struct {
	Action string `json:"action"`
}
```

Important difference:

| Python                        | Go                              |
| ----------------------------- | ------------------------------- |
| Dynamic dictionaries          | Static structs                  |
| `data["sender"]["login"]`     | `event.Sender.Login`            |
| Runtime key mistakes possible | Compile-time structure helps    |
| Less boilerplate              | More explicit                   |
| Easy for quick scripts        | Better for production contracts |

For your project, Go structs are useful because webhook payloads become typed, validated objects before sending Slack notifications.

---

## 18. Revision checkpoint

You should now be able to answer these:

1. What does an `EventListener` do?
2. What does a `TriggerBinding` do?
3. What does a `TriggerTemplate` do?
4. What is the difference between manual `PipelineRun` and webhook-triggered `PipelineRun`?
5. Where does PR number come from?
6. Where does commit ID come from?
7. Where does sender come from?
8. Why should we first print params before adding Slack notification?
9. How is Tekton trigger mapping similar to Python dictionary access?
10. Why is Go struct parsing useful for webhook events?

Simple answers:

```text
EventListener receives the webhook.
TriggerBinding extracts values from JSON.
TriggerTemplate creates the PipelineRun.
PipelineRun executes the Pipeline.
Slack notification comes after the pipeline knows success/failure context.
```

---

## 19. Homework

### Homework 1: Tekton trigger practice

Create these files:

```text
.tekton/pr-binding.yaml
.tekton/pr-trigger-template.yaml
.tekton/pr-listener.yaml
.tekton/pr-pipeline.yaml
```

Then test with:

```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: pull_request" \
  -d '<sample-json>'
```

Verify:

```bash
kubectl get pipelinerun -n slack-integration-dev
tkn pipelinerun logs -f -n slack-integration-dev
```

---

### Homework 2: Add one more field

Add PR title.

GitHub JSON path:

```text
body.pull_request.title
```

TriggerBinding:

```yaml
- name: pr-title
  value: $(body.pull_request.title)
```

Then pass it into:

```text
TriggerTemplate → PipelineRun → Pipeline → Task logs
```

---

### Homework 3: Extend Go parser

Add this field:

```go
Title string `json:"title"`
```

Then print:

```text
PR Title: Add validation logic
```

---

### Homework 4: BFS practice

Modify the BFS problem to check whether a path exists between two nodes.

Example:

```text
Start: A
Target: E
Output: true
```

---

## Final Day 10 mental model

```text
GitHub event
   ↓
EventListener receives it
   ↓
TriggerBinding extracts values
   ↓
TriggerTemplate creates PipelineRun
   ↓
Pipeline validates/builds/tests
   ↓
Slack task sends notification
```

For your `slack-integration` project, Day 10 is the bridge between:

```text
manual CI/CD testing
```

and

```text
real event-driven automation
```

[1]: https://tekton.dev/docs/triggers/triggerbindings/?utm_source=chatgpt.com "TriggerBindings - Tekton"
[2]: https://tekton.dev/docs/pipelines/pipelineruns/?utm_source=chatgpt.com "PipelineRuns - Tekton"
[3]: https://github.com/tektoncd/triggers/blob/main/docs/eventlisteners.md?utm_source=chatgpt.com "triggers/docs/eventlisteners.md at main · tektoncd/triggers"
