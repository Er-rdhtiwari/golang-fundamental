# Day 9 — Tekton Basics for `slack-integration` Project

## 1. Day 9 learning goals

By the end of Day 9, you should understand:

1. What Tekton is and why CI/CD projects use it.
2. What **Task**, **Pipeline**, **PipelineRun**, and **TaskRun** mean.
3. How `go test`, `go build`, and validation commands become Tekton tasks.
4. How **params** pass input into a pipeline.
5. How **workspaces** share files between tasks.
6. How to debug a basic Tekton pipeline locally in Minikube.
7. Tree basics in DSA and one simple tree traversal problem in Go.
8. How to build a small **pipeline status tracker model** in Go.

Tekton runs on Kubernetes and defines CI/CD building blocks as Kubernetes custom resources, so after Tekton is installed you can manage pipelines using `kubectl`, similar to Pods, Deployments, and Secrets. ([Tekton][1])

---

## 2. Quick revision of Days 1 to 8

### Day 1 — Go CLI basics

You learned how input enters the program through CLI flags.

Example:

```bash
go run cmd/slack-notifier/main.go \
  --event-type pr \
  --status failed \
  --pipeline-name pr-validation
```

Simple idea:

```text
User input -> CLI flags -> Go variables -> event object -> Slack notification
```

Python comparison:

```python
# Python
event_type = input("Enter event type: ")
```

Go comparison:

```go
eventType := flag.String("event-type", "", "event type")
```

---

### Day 2 — Structs and event model

You learned that Go structs are used to group related data.

Go:

```go
type PipelineEvent struct {
    EventType string
    Status    string
}
```

Python equivalent:

```python
@dataclass
class PipelineEvent:
    event_type: str
    status: str
```

---

### Day 3 — JSON, HTTP, and Slack webhook

You learned how a Go struct becomes JSON and is sent to Slack using HTTP POST.

```text
PipelineEvent -> SlackMessage -> JSON -> HTTP POST -> Slack
```

---

### Day 4 — Router logic

You learned how routing decides which Slack webhook should receive the message.

```text
PR event  -> PR webhook
CD event  -> CD webhook
Job event -> Job webhook or fallback webhook
```

---

### Day 5 — Error handling and logging

You learned why real systems need useful errors and logs.

Go usually returns errors:

```go
if err != nil {
    return err
}
```

Python usually raises exceptions:

```python
try:
    send_message()
except Exception as err:
    print(err)
```

---

### Day 6 — Go testing

You learned `go test`.

```bash
go test ./...
```

Python comparison:

```bash
pytest
```

---

### Day 7 — Shell scripting

You learned helper scripts.

```bash
#!/usr/bin/env bash
set -euo pipefail

go test ./...
go build ./cmd/slack-notifier
```

---

### Day 8 — Kubernetes basics

You learned that Kubernetes runs containers inside Pods.

```text
Deployment -> ReplicaSet -> Pod -> Container
```

Tekton uses Kubernetes underneath. A Tekton Task eventually creates a Pod, and each step runs inside a container.

---

## 3. Explain Tekton in very simple language

Think of Tekton as a **CI/CD engine that runs your project commands inside Kubernetes**.

Without Tekton, you run commands manually:

```bash
go test ./...
go build ./cmd/slack-notifier
go run ./cmd/slack-notifier --dry-run=true
```

With Tekton, you describe those commands in YAML:

```text
Task 1: validate input
Task 2: run tests
Task 3: build binary
Task 4: send Slack notification
```

Then Kubernetes runs them in containers.

### Simple real-life analogy

Imagine cooking food.

| Real life                           | Tekton      |
| ----------------------------------- | ----------- |
| Recipe step: cut vegetables         | Task step   |
| Full recipe: make dinner            | Pipeline    |
| Actually cooking dinner today       | PipelineRun |
| Actual execution of one recipe step | TaskRun     |
| Kitchen counter shared by steps     | Workspace   |
| Recipe input like spice level       | Params      |

---

## 4. Explain Task vs Pipeline vs PipelineRun clearly

Tekton official docs describe a Task as a resource with one or more container **steps**, and Tasks can also define params and workspaces. ([Tekton][2])

### A. Task

A **Task** is one small job.

Example:

```text
Task: run Go tests
Command: go test ./...
```

Beginner meaning:

```text
Task = one reusable unit of work
```

Python analogy:

```python
def run_tests():
    os.system("pytest")
```

Go project example:

```text
Task: go-test
Step: run go test ./...
```

---

### B. Pipeline

A **Pipeline** connects multiple Tasks in order. Tekton Pipelines define the Tasks that make up the pipeline, and can also define params, workspaces, retries, timeouts, and task ordering. ([Tekton][3])

Beginner meaning:

```text
Pipeline = ordered flow of multiple tasks
```

Example:

```text
validate-event -> go-test -> go-build -> slack-notify
```

Python analogy:

```python
def pipeline():
    validate_event()
    run_tests()
    build_app()
    send_slack_notification()
```

---

### C. PipelineRun

A **PipelineRun** is one actual execution of a Pipeline.

Pipeline is the template.

PipelineRun is the actual run.

Example:

```text
Pipeline: pr-validation-pipeline
PipelineRun: pr-validation-pipeline-run-001
```

Tekton docs state that a PipelineRun executes a Pipeline on-cluster and automatically creates TaskRuns for every Task in the Pipeline. ([Tekton][4])

Python analogy:

```python
pipeline(event_type="pr", status="failed")
```

---

### D. TaskRun

A **TaskRun** is one actual execution of one Task.

Tekton docs describe TaskRun as the object that executes a Task on-cluster; the Task defines steps, and the TaskRun runs those steps in order until success or failure. ([Tekton][5])

Example:

```text
PipelineRun creates:
- validate-event TaskRun
- go-test TaskRun
- go-build TaskRun
- slack-notify TaskRun
```

---

### Task vs Pipeline comparison

| Concept     | Beginner meaning               | Project example                     |
| ----------- | ------------------------------ | ----------------------------------- |
| Task        | One job                        | `go test ./...`                     |
| Pipeline    | Many jobs connected            | validate -> test -> build -> notify |
| PipelineRun | One execution of full pipeline | Run PR pipeline today               |
| TaskRun     | One execution of one task      | Actual `go-test` execution          |

---

## 5. Explain params and workspaces

### A. Params

Params are inputs to Tasks or Pipelines.

Example values:

```text
event_type = pr
status = failed
pipeline_name = pr-validation
repository = cloud-resource-onboarding
```

Tekton supports variable substitution using syntax such as `$(params.<param name>)`. Its docs also warn that Tekton does not escape variable contents automatically, so task authors must handle shell escaping carefully when using params in scripts. ([Tekton][6])

Example:

```yaml
params:
  - name: event-type
    type: string
```

Usage:

```bash
echo "Event type is $(params.event-type)"
```

Python comparison:

```python
def notify(event_type):
    print(event_type)
```

Go comparison:

```go
func Notify(eventType string) {
    fmt.Println(eventType)
}
```

Tekton comparison:

```yaml
$(params.event-type)
```

---

### B. Workspaces

A workspace is a shared folder.

Tekton docs describe workspaces as filesystem areas that Tasks declare and TaskRuns provide at runtime; they can be backed by an `emptyDir`, Secret, ConfigMap, PVC, or other volume types. ([Tekton][7])

Simple meaning:

```text
Workspace = shared folder between tasks
```

Example:

```text
prepare-source task writes code into workspace
go-test task reads code from same workspace
go-build task reads code from same workspace
```

Python analogy:

```python
# Function 1 writes file
write_source_code("/tmp/project")

# Function 2 reads same folder
run_tests("/tmp/project")
```

Tekton analogy:

```text
Task 1 -> /workspace/source
Task 2 -> /workspace/source
Task 3 -> /workspace/source
```

---

## 6. How this project’s Go validation/build steps become Tekton tasks

Your local project commands:

```bash
go test ./...
go build -o bin/slack-notifier ./cmd/slack-notifier
go run ./cmd/slack-notifier/main.go --event-type pr --status failed --dry-run=true
```

In Tekton, these become separate Tasks:

| Local command               | Tekton Task      |
| --------------------------- | ---------------- |
| `go test ./...`             | `go-test`        |
| `go build ...`              | `go-build`       |
| `go run ... --dry-run=true` | `validate-event` |
| Slack notification command  | `slack-notify`   |

### Why CI pipelines are broken into steps

CI pipelines are broken into steps because:

1. Each step has one clear responsibility.
2. Failure becomes easy to identify.
3. You can retry only the failed part.
4. Logs are cleaner.
5. Teams can reuse tasks across PR, CD, and job pipelines.
6. Security is easier because each task can use only the permissions it needs.
7. Long workflows become easier to debug.

Bad design:

```text
one huge script does everything
```

Better design:

```text
validate -> test -> build -> notify
```

---

## 7. Pipeline execution flow in ASCII

```text
Developer pushes code / PR opened
              |
              v
        Tekton PipelineRun
              |
              v
+-------------+-------------+
|                           |
v                           v
TaskRun: prepare-source     Status starts: Running
              |
              v
TaskRun: validate-event
              |
              v
TaskRun: go-test
              |
              v
TaskRun: go-build
              |
              v
TaskRun: slack-notify
              |
              v
PipelineRun status: Succeeded / Failed
```

More Kubernetes-style view:

```text
PipelineRun
   |
   | creates
   v
TaskRun 1 -> Pod -> Container step
TaskRun 2 -> Pod -> Container step
TaskRun 3 -> Pod -> Container step
TaskRun 4 -> Pod -> Container step
```

---

## 8. Pseudocode first for a simple pipeline

```text
START pipeline

INPUT:
  event_type
  status
  pipeline_name
  repository

TASK 1: prepare source
  create project files or clone repository

TASK 2: validate event
  check event_type is not empty
  check status is valid
  check pipeline_name is not empty

TASK 3: test Go project
  run go test ./...

TASK 4: build Go binary
  run go build

TASK 5: notify Slack
  if previous tasks passed:
      send success notification
  else:
      send failure notification

END pipeline
```

Python-style mental model:

```python
def pipeline(event_type, status, pipeline_name):
    prepare_source()
    validate_event(event_type, status, pipeline_name)
    run_go_tests()
    build_binary()
    notify_slack(status)
```

---

# 9. Real Tekton YAML examples with detailed explanation

Below is a beginner-friendly local example.

For real projects, your source code usually comes from Git using a clone task or from IBM Toolchain checkout. For beginner practice, this YAML creates a tiny Go project inside a workspace so you can understand the flow clearly.

> Use `tekton.dev/v1` for current Tekton examples. Some older clusters or managed environments may still use `tekton.dev/v1beta1`, so always check your cluster using:
>
> ```bash
> kubectl api-resources | grep -i tekton
> kubectl api-resources | grep -i task
> ```

---

## 9.1 Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: slack-integration-dev
```

Apply:

```bash
kubectl apply -f namespace.yaml
```

---

## 9.2 Task 1 — prepare Go source

File: `task-prepare-source.yaml`

```yaml
apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: prepare-go-source
  namespace: slack-integration-dev
spec:
  workspaces:
    - name: source
  steps:
    - name: create-source
      image: golang:1.22
      workingDir: $(workspaces.source.path)
      script: |
        #!/usr/bin/env sh
        set -eu

        echo "Creating small Go project..."

        mkdir -p cmd/slack-notifier
        mkdir -p pkg/status

        cat > go.mod <<'EOF'
        module slack-integration

        go 1.22
        EOF

        cat > pkg/status/status.go <<'EOF'
        package status

        type PipelineStatus struct {
            PipelineName string
            EventType    string
            Status       string
        }

        func (p PipelineStatus) IsFailed() bool {
            return p.Status == "failed"
        }
        EOF

        cat > pkg/status/status_test.go <<'EOF'
        package status

        import "testing"

        func TestIsFailed(t *testing.T) {
            event := PipelineStatus{
                PipelineName: "pr-validation",
                EventType: "pr",
                Status: "failed",
            }

            if !event.IsFailed() {
                t.Fatal("expected failed status")
            }
        }
        EOF

        cat > cmd/slack-notifier/main.go <<'EOF'
        package main

        import "fmt"

        func main() {
            fmt.Println("slack notifier build successful")
        }
        EOF

        echo "Source created successfully"
```

### Explanation

```yaml
kind: Task
```

This creates one Tekton Task.

```yaml
workspaces:
  - name: source
```

This Task needs a shared folder called `source`.

```yaml
image: golang:1.22
```

This step runs inside a Go container.

```yaml
workingDir: $(workspaces.source.path)
```

The command runs inside the shared workspace folder.

```yaml
script: |
```

This is the shell script that runs inside the container.

---

## 9.3 Task 2 — validate event params

File: `task-validate-event.yaml`

```yaml
apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: validate-event
  namespace: slack-integration-dev
spec:
  params:
    - name: event-type
      type: string
    - name: status
      type: string
    - name: pipeline-name
      type: string
  steps:
    - name: validate
      image: alpine:3.20
      script: |
        #!/usr/bin/env sh
        set -eu

        echo "Validating pipeline event..."
        echo "Event type: $(params.event-type)"
        echo "Status: $(params.status)"
        echo "Pipeline name: $(params.pipeline-name)"

        if [ -z "$(params.event-type)" ]; then
          echo "event-type is required"
          exit 1
        fi

        if [ -z "$(params.status)" ]; then
          echo "status is required"
          exit 1
        fi

        if [ -z "$(params.pipeline-name)" ]; then
          echo "pipeline-name is required"
          exit 1
        fi

        case "$(params.status)" in
          running|succeeded|failed)
            echo "Status is valid"
            ;;
          *)
            echo "Invalid status. Allowed: running, succeeded, failed"
            exit 1
            ;;
        esac

        echo "Validation passed"
```

### Explanation

This task checks whether input is valid before running expensive build/test work.

In your real Go project, this can later become:

```bash
go run ./cmd/slack-notifier/main.go \
  --event-type "$(params.event-type)" \
  --status "$(params.status)" \
  --pipeline-name "$(params.pipeline-name)" \
  --dry-run=true
```

Python comparison:

```python
if status not in ["running", "succeeded", "failed"]:
    raise ValueError("Invalid status")
```

Shell/Tekton comparison:

```bash
case "$(params.status)" in
  running|succeeded|failed)
    echo "Status is valid"
    ;;
  *)
    exit 1
    ;;
esac
```

---

## 9.4 Task 3 — run Go tests

File: `task-go-test.yaml`

```yaml
apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: go-test
  namespace: slack-integration-dev
spec:
  workspaces:
    - name: source
  steps:
    - name: test
      image: golang:1.22
      workingDir: $(workspaces.source.path)
      script: |
        #!/usr/bin/env sh
        set -eu

        echo "Running Go tests..."
        go test ./...
```

### Explanation

This Tekton task is the same as running locally:

```bash
go test ./...
```

Python comparison:

```bash
pytest
```

Go difference:

```text
Go has built-in testing support with go test.
Python commonly uses pytest or unittest.
```

---

## 9.5 Task 4 — build Go binary

File: `task-go-build.yaml`

```yaml
apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: go-build
  namespace: slack-integration-dev
spec:
  workspaces:
    - name: source
  steps:
    - name: build
      image: golang:1.22
      workingDir: $(workspaces.source.path)
      script: |
        #!/usr/bin/env sh
        set -eu

        echo "Building Go binary..."
        mkdir -p bin
        go build -o bin/slack-notifier ./cmd/slack-notifier

        echo "Build completed"
        ls -lh bin/
```

### Explanation

This maps to your local command:

```bash
go build -o bin/slack-notifier ./cmd/slack-notifier
```

Python comparison:

```text
Python usually runs source files directly.
Go compiles source code into a binary.
```

Example:

```bash
python app.py
```

versus:

```bash
go build -o app
./app
```

---

## 9.6 Task 5 — fake Slack notification

File: `task-slack-notify.yaml`

```yaml
apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: slack-notify
  namespace: slack-integration-dev
spec:
  params:
    - name: event-type
      type: string
    - name: status
      type: string
    - name: pipeline-name
      type: string
  steps:
    - name: notify
      image: alpine:3.20
      script: |
        #!/usr/bin/env sh
        set -eu

        echo "Sending Slack notification..."
        echo "Pipeline: $(params.pipeline-name)"
        echo "Event type: $(params.event-type)"
        echo "Status: $(params.status)"
        echo "Slack notification simulated successfully"
```

In the real project, this would later call your Go CLI:

```bash
go run ./cmd/slack-notifier/main.go \
  --event-type "$(params.event-type)" \
  --status "$(params.status)" \
  --pipeline-name "$(params.pipeline-name)"
```

---

## 9.7 Pipeline

File: `pipeline-slack-integration.yaml`

```yaml
apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: slack-integration-pipeline
  namespace: slack-integration-dev
spec:
  params:
    - name: event-type
      type: string
      default: pr
    - name: status
      type: string
      default: succeeded
    - name: pipeline-name
      type: string
      default: pr-validation

  workspaces:
    - name: shared-source

  tasks:
    - name: prepare-source
      taskRef:
        name: prepare-go-source
      workspaces:
        - name: source
          workspace: shared-source

    - name: validate-event
      taskRef:
        name: validate-event
      runAfter:
        - prepare-source
      params:
        - name: event-type
          value: $(params.event-type)
        - name: status
          value: $(params.status)
        - name: pipeline-name
          value: $(params.pipeline-name)

    - name: run-tests
      taskRef:
        name: go-test
      runAfter:
        - validate-event
      workspaces:
        - name: source
          workspace: shared-source

    - name: build-binary
      taskRef:
        name: go-build
      runAfter:
        - run-tests
      workspaces:
        - name: source
          workspace: shared-source

    - name: notify-slack
      taskRef:
        name: slack-notify
      runAfter:
        - build-binary
      params:
        - name: event-type
          value: $(params.event-type)
        - name: status
          value: $(params.status)
        - name: pipeline-name
          value: $(params.pipeline-name)
```

### Explanation

```yaml
params:
```

Pipeline-level inputs.

```yaml
workspaces:
  - name: shared-source
```

Pipeline-level shared folder.

```yaml
taskRef:
  name: go-test
```

This says: use the existing `go-test` Task.

```yaml
runAfter:
  - validate-event
```

This controls execution order.

```yaml
value: $(params.status)
```

This passes Pipeline param into Task param.

---

## 9.8 PipelineRun

File: `pipelinerun-slack-integration.yaml`

```yaml
apiVersion: tekton.dev/v1
kind: PipelineRun
metadata:
  generateName: slack-integration-run-
  namespace: slack-integration-dev
spec:
  pipelineRef:
    name: slack-integration-pipeline

  params:
    - name: event-type
      value: pr
    - name: status
      value: succeeded
    - name: pipeline-name
      value: pr-validation

  workspaces:
    - name: shared-source
      emptyDir: {}
```

### Explanation

```yaml
generateName:
```

Kubernetes will generate a unique name.

Example:

```text
slack-integration-run-abcde
```

```yaml
pipelineRef:
  name: slack-integration-pipeline
```

This tells Tekton which Pipeline to run.

```yaml
emptyDir: {}
```

This creates a temporary folder. It is deleted after the run finishes.

For real production pipelines, you may use a PVC or a Git clone task instead.

---

# 10. Hands-on tasks

## Step 1: Create namespace

```bash
kubectl apply -f namespace.yaml
```

---

## Step 2: Apply Tasks

```bash
kubectl apply -f task-prepare-source.yaml
kubectl apply -f task-validate-event.yaml
kubectl apply -f task-go-test.yaml
kubectl apply -f task-go-build.yaml
kubectl apply -f task-slack-notify.yaml
```

---

## Step 3: Apply Pipeline

```bash
kubectl apply -f pipeline-slack-integration.yaml
```

---

## Step 4: Start PipelineRun

```bash
kubectl create -f pipelinerun-slack-integration.yaml
```

---

## Step 5: Check PipelineRun

```bash
kubectl get pipelinerun -n slack-integration-dev
```

---

## Step 6: Check TaskRuns

```bash
kubectl get taskrun -n slack-integration-dev
```

---

## Step 7: Check Pods

```bash
kubectl get pods -n slack-integration-dev
```

---

## Step 8: View logs

Using Tekton CLI:

```bash
tkn pipelinerun logs -f -n slack-integration-dev
```

Using kubectl:

```bash
kubectl logs -n slack-integration-dev -l tekton.dev/pipelineRun=<PIPELINE_RUN_NAME> --all-containers=true
```

---

# 11. Expected output

You should see something like:

```text
Creating small Go project...
Source created successfully
```

Then:

```text
Validating pipeline event...
Event type: pr
Status: succeeded
Pipeline name: pr-validation
Status is valid
Validation passed
```

Then:

```text
Running Go tests...
ok  	slack-integration/pkg/status	0.002s
```

Then:

```text
Building Go binary...
Build completed
```

Then:

```text
Sending Slack notification...
Pipeline: pr-validation
Event type: pr
Status: succeeded
Slack notification simulated successfully
```

PipelineRun should show:

```text
Succeeded
```

---

# 12. Common mistakes

## Mistake 1: Wrong API version

Error:

```text
no matches for kind "Task" in version "tekton.dev/v1"
```

Fix:

```bash
kubectl api-resources | grep -i task
```

Your cluster may support:

```yaml
apiVersion: tekton.dev/v1
```

or older:

```yaml
apiVersion: tekton.dev/v1beta1
```

---

## Mistake 2: Forgetting namespace

Wrong:

```bash
kubectl get pipelinerun
```

Correct:

```bash
kubectl get pipelinerun -n slack-integration-dev
```

---

## Mistake 3: Workspace name mismatch

Pipeline has:

```yaml
workspaces:
  - name: shared-source
```

Task expects:

```yaml
workspaces:
  - name: source
```

Mapping must be correct:

```yaml
workspaces:
  - name: source
    workspace: shared-source
```

---

## Mistake 4: Running Go command in wrong folder

Wrong:

```yaml
workingDir: /
```

Correct:

```yaml
workingDir: $(workspaces.source.path)
```

---

## Mistake 5: Using an image without Go installed

Wrong:

```yaml
image: alpine
script: |
  go test ./...
```

Alpine does not include Go by default.

Correct:

```yaml
image: golang:1.22
```

---

## Mistake 6: Putting everything into one giant Task

Not ideal:

```text
validate + test + build + notify in one huge script
```

Better:

```text
validate task
test task
build task
notify task
```

This makes debugging easier.

---

# 13. Debugging tips

## Check PipelineRuns

```bash
kubectl get pipelinerun -n slack-integration-dev
```

---

## Describe failed PipelineRun

```bash
kubectl describe pipelinerun <PIPELINE_RUN_NAME> -n slack-integration-dev
```

---

## Check TaskRuns

```bash
kubectl get taskrun -n slack-integration-dev
```

---

## Describe failed TaskRun

```bash
kubectl describe taskrun <TASK_RUN_NAME> -n slack-integration-dev
```

---

## Check Pods

```bash
kubectl get pods -n slack-integration-dev
```

---

## Read Pod logs

```bash
kubectl logs <POD_NAME> -n slack-integration-dev --all-containers=true
```

---

## Check Kubernetes events

```bash
kubectl get events -n slack-integration-dev --sort-by=.lastTimestamp
```

---

## Check Tekton CRDs

```bash
kubectl get crd | grep -i tekton
kubectl get crd | grep -i task
kubectl get crd | grep -i pipeline
```

---

## Debugging mental model

```text
PipelineRun failed
      |
      v
Which TaskRun failed?
      |
      v
Which Pod failed?
      |
      v
Which step failed?
      |
      v
Read logs
      |
      v
Fix YAML / command / workspace / params
```

---

# 14. One DSA topic — Tree basics

## What is a tree?

A tree is a data structure that looks like a family tree.

Example:

```text
        A
       / \
      B   C
     / \
    D   E
```

## Important tree terms

| Term      | Meaning                        |
| --------- | ------------------------------ |
| Root      | Top node                       |
| Node      | One item in tree               |
| Edge      | Connection between nodes       |
| Parent    | Node above another node        |
| Child     | Node below another node        |
| Leaf      | Node with no children          |
| Height    | Longest path from root to leaf |
| Traversal | Visiting nodes in some order   |

In the example:

```text
A is root
B and C are children of A
D and E are children of B
C, D, E are leaf nodes
```

---

## Common tree traversals

### 1. Preorder

```text
Root -> Left -> Right
```

Example:

```text
A B D E C
```

### 2. Inorder

```text
Left -> Root -> Right
```

Example:

```text
D B E A C
```

### 3. Postorder

```text
Left -> Right -> Root
```

Example:

```text
D E B C A
```

### 4. Level order

```text
Level by level
```

Example:

```text
A B C D E
```

Python comparison:

```python
class TreeNode:
    def __init__(self, val):
        self.val = val
        self.left = None
        self.right = None
```

Go comparison:

```go
type TreeNode struct {
    Val   int
    Left  *TreeNode
    Right *TreeNode
}
```

Key Go difference:

```text
*TreeNode means pointer to another TreeNode.
In Python, object references are automatic.
In Go, you explicitly use pointers for linked structures.
```

---

# 15. One Go DSA problem — Preorder traversal

## Problem

Given a binary tree, return preorder traversal.

Preorder means:

```text
Root -> Left -> Right
```

Example tree:

```text
        1
       / \
      2   3
     / \
    4   5
```

Expected output:

```text
[1, 2, 4, 5, 3]
```

---

## Pseudocode

```text
function preorder(node):
    if node is nil:
        return

    add node value to result
    preorder(left child)
    preorder(right child)
```

---

## Go solution

```go
package main

import "fmt"

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func preorderTraversal(root *TreeNode) []int {
	result := []int{}

	var dfs func(node *TreeNode)

	dfs = func(node *TreeNode) {
		if node == nil {
			return
		}

		result = append(result, node.Val)
		dfs(node.Left)
		dfs(node.Right)
	}

	dfs(root)
	return result
}

func main() {
	root := &TreeNode{Val: 1}
	root.Left = &TreeNode{Val: 2}
	root.Right = &TreeNode{Val: 3}
	root.Left.Left = &TreeNode{Val: 4}
	root.Left.Right = &TreeNode{Val: 5}

	output := preorderTraversal(root)

	fmt.Println(output)
}
```

---

## Line-by-line beginner explanation

```go
type TreeNode struct
```

Defines one tree node.

```go
Val int
```

Stores node value.

```go
Left *TreeNode
Right *TreeNode
```

Stores links to left and right child nodes.

```go
result := []int{}
```

Creates an empty integer slice.

Python equivalent:

```python
result = []
```

```go
var dfs func(node *TreeNode)
```

Declares a recursive helper function.

```go
if node == nil
```

Base condition.

Python equivalent:

```python
if node is None:
```

```go
result = append(result, node.Val)
```

Adds current node value.

Python equivalent:

```python
result.append(node.val)
```

---

# 16. Module-based practice task — Build a pipeline status tracker model

## Goal

Create a small Go model that tracks pipeline status.

This model will help your `slack-integration` project understand:

```text
pipeline name
event type
stage
status
failed step
error message
start/end state
```

---

## Suggested file

```text
pkg/status/tracker.go
```

---

## Go code

```go
package status

import (
	"fmt"
	"time"
)

type PipelineStatus string

const (
	StatusRunning   PipelineStatus = "running"
	StatusSucceeded PipelineStatus = "succeeded"
	StatusFailed    PipelineStatus = "failed"
)

type PipelineTracker struct {
	PipelineName string
	EventType    string
	Stage        string
	Status       PipelineStatus
	FailedStep   string
	ErrorMessage string
	StartedAt    time.Time
	FinishedAt   time.Time
}

func NewPipelineTracker(pipelineName, eventType, stage string) PipelineTracker {
	return PipelineTracker{
		PipelineName: pipelineName,
		EventType:    eventType,
		Stage:        stage,
		Status:       StatusRunning,
		StartedAt:    time.Now(),
	}
}

func (p *PipelineTracker) MarkSucceeded() {
	p.Status = StatusSucceeded
	p.FinishedAt = time.Now()
}

func (p *PipelineTracker) MarkFailed(failedStep, errorMessage string) {
	p.Status = StatusFailed
	p.FailedStep = failedStep
	p.ErrorMessage = errorMessage
	p.FinishedAt = time.Now()
}

func (p PipelineTracker) IsFailed() bool {
	return p.Status == StatusFailed
}

func (p PipelineTracker) Summary() string {
	if p.IsFailed() {
		return fmt.Sprintf(
			"Pipeline %s failed at step %s: %s",
			p.PipelineName,
			p.FailedStep,
			p.ErrorMessage,
		)
	}

	return fmt.Sprintf(
		"Pipeline %s completed with status %s",
		p.PipelineName,
		p.Status,
	)
}
```

---

## Python equivalent

```python
from dataclasses import dataclass
from datetime import datetime
from enum import Enum

class PipelineStatus(Enum):
    RUNNING = "running"
    SUCCEEDED = "succeeded"
    FAILED = "failed"

@dataclass
class PipelineTracker:
    pipeline_name: str
    event_type: str
    stage: str
    status: PipelineStatus = PipelineStatus.RUNNING
    failed_step: str = ""
    error_message: str = ""
```

---

## Go vs Python learning points

| Concept           | Python                | Go                                  |
| ----------------- | --------------------- | ----------------------------------- |
| Class-like data   | `class` / `dataclass` | `struct`                            |
| Enum              | `Enum`                | `const` with custom type            |
| Method receiver   | `self`                | `(p PipelineTracker)`               |
| Mutable method    | normal object method  | pointer receiver `*PipelineTracker` |
| Current time      | `datetime.now()`      | `time.Now()`                        |
| String formatting | f-string              | `fmt.Sprintf`                       |
| Public field      | normal naming         | Capitalized field name              |

Important Go convention:

```go
PipelineName string
```

Capital `P` means exported/public.

```go
pipelineName string
```

Small `p` means unexported/package-private.

---

## Practice test file

File:

```text
pkg/status/tracker_test.go
```

Code:

```go
package status

import "testing"

func TestPipelineTrackerMarkFailed(t *testing.T) {
	tracker := NewPipelineTracker("pr-validation", "pr", "validation")

	tracker.MarkFailed("go-test", "unit test failed")

	if !tracker.IsFailed() {
		t.Fatal("expected tracker to be failed")
	}

	if tracker.FailedStep != "go-test" {
		t.Fatalf("expected failed step go-test, got %s", tracker.FailedStep)
	}
}

func TestPipelineTrackerMarkSucceeded(t *testing.T) {
	tracker := NewPipelineTracker("cd-pipeline", "cd", "build")

	tracker.MarkSucceeded()

	if tracker.Status != StatusSucceeded {
		t.Fatalf("expected status succeeded, got %s", tracker.Status)
	}
}
```

Run:

```bash
go test ./pkg/status
```

Expected:

```text
ok  	slack-integration/pkg/status
```

---

# 17. Revision checkpoint

You should now be able to answer these.

## Concept questions

1. What is Tekton?
2. What is a Task?
3. What is a Pipeline?
4. What is a PipelineRun?
5. What is a TaskRun?
6. Why do we use params?
7. Why do we use workspaces?
8. Why should `go test` and `go build` be separate tasks?
9. What command shows PipelineRuns?
10. What command shows TaskRuns?

---

## Quick answers

1. Tekton is a Kubernetes-native CI/CD system.
2. Task is one reusable job.
3. Pipeline is a flow of multiple tasks.
4. PipelineRun is one execution of a Pipeline.
5. TaskRun is one execution of a Task.
6. Params pass input values.
7. Workspaces share files between tasks.
8. Separate tasks make debugging and reuse easier.
9. `kubectl get pipelinerun -n <namespace>`
10. `kubectl get taskrun -n <namespace>`

---

## Mental model to remember

```text
Task = function definition
Pipeline = function orchestration
PipelineRun = function call
TaskRun = actual execution of one function
Workspace = shared folder
Params = function arguments
```

Python comparison:

```python
def validate():
    pass

def test():
    pass

def build():
    pass

def pipeline():
    validate()
    test()
    build()

pipeline()
```

Tekton comparison:

```text
Task: validate
Task: test
Task: build
Pipeline: validate -> test -> build
PipelineRun: execute now
```

---

# 18. Homework

## Homework 1 — Tekton concept revision

Write a 5-line explanation for each:

```text
Task
Pipeline
PipelineRun
TaskRun
Params
Workspace
```

---

## Homework 2 — Modify pipeline status

Run PipelineRun with failed status:

```yaml
params:
  - name: event-type
    value: pr
  - name: status
    value: failed
  - name: pipeline-name
    value: pr-validation
```

Observe logs.

---

## Homework 3 — Add one more validation

In `validate-event` task, allow only these event types:

```text
pr
cd
job
```

Add shell validation:

```bash
case "$(params.event-type)" in
  pr|cd|job)
    echo "Event type is valid"
    ;;
  *)
    echo "Invalid event type"
    exit 1
    ;;
esac
```

---

## Homework 4 — Extend Go pipeline tracker

Add this method:

```go
func (p PipelineTracker) Duration() time.Duration {
	return p.FinishedAt.Sub(p.StartedAt)
}
```

Then write a test for it.

---

## Homework 5 — DSA tree practice

Implement inorder traversal in Go.

Expected order:

```text
Left -> Root -> Right
```

For this tree:

```text
        1
       / \
      2   3
     / \
    4   5
```

Expected output:

```text
[4, 2, 5, 1, 3]
```

---

## Final Day 9 takeaway

Tekton is just a clean way to run your project commands inside Kubernetes.

For your `slack-integration` project, remember this flow:

```text
Go CLI command locally
        |
        v
Tekton Task
        |
        v
Pipeline connects tasks
        |
        v
PipelineRun executes pipeline
        |
        v
TaskRuns create Pods
        |
        v
Logs show validation/test/build/notify result
```

Best mental shortcut:

```text
Task = one command group
Pipeline = ordered command flow
PipelineRun = actual run
TaskRun = actual task execution
Params = inputs
Workspace = shared project folder
```

[1]: https://tekton.dev/docs/pipelines/?utm_source=chatgpt.com "Tasks and Pipelines - Tekton"
[2]: https://tekton.dev/docs/pipelines/tasks/ "Tekton"
[3]: https://tekton.dev/docs/pipelines/pipelines/ "Tekton"
[4]: https://tekton.dev/docs/pipelines/pipelineruns/ "Tekton"
[5]: https://tekton.dev/docs/pipelines/taskruns/ "Tekton"
[6]: https://tekton.dev/docs/pipelines/variables/ "Tekton"
[7]: https://tekton.dev/docs/pipelines/workspaces/ "Tekton"

