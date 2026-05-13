# Day 11 — Tekton Debugging in Minikube

Today’s goal is simple:

> When a Tekton run fails, do not panic. Follow the trail.

In Tekton, a `PipelineRun` starts the pipeline, Tekton creates `TaskRuns` for the pipeline tasks, and each task runs inside Kubernetes Pods. Tekton’s own docs describe this chain: a `PipelineRun` executes tasks in order and automatically creates matching `TaskRuns`; logs are stored in the Pod containers that run the steps. ([Tekton][1])

---

## 1. Day 11 learning goals

By the end of Day 11, you should be able to:

1. Test your `slack-integration` project locally in Minikube.
2. Check whether Tekton itself is healthy.
3. Inspect a failing `PipelineRun`.
4. Find the related `TaskRun`.
5. Find the related Pod.
6. Inspect the exact failing step log.
7. Debug secret and service account issues.
8. Separate failures into three buckets:

   * configuration error
   * code error
   * infrastructure error
9. Use a calm debugging order instead of guessing.
10. Practice heap / priority queue basics in Go.

---

## 2. Quick revision of Days 1 to 10

Think of the previous days like this:

| Day    | Main idea                                                                               |
| ------ | --------------------------------------------------------------------------------------- |
| Day 1  | CI/CD means automatically checking, building, testing, and deploying code.              |
| Day 2  | Kubernetes runs containers using objects like Pods, Deployments, Services, and Secrets. |
| Day 3  | Tekton is a Kubernetes-native CI/CD system.                                             |
| Day 4  | A `Task` is one reusable unit of work.                                                  |
| Day 5  | A `Pipeline` connects many tasks together.                                              |
| Day 6  | A `PipelineRun` actually starts the pipeline.                                           |
| Day 7  | Workspaces help tasks share files.                                                      |
| Day 8  | Params pass values like repo URL, branch, Slack channel, or image name.                 |
| Day 9  | Secrets hold sensitive values like Slack webhook URLs or tokens.                        |
| Day 10 | Triggers can start a pipeline from an external event, such as a webhook.                |

Day 11 connects all of this.

Earlier you learned how to create things. Today you learn how to investigate things.

---

## 3. Debugging mindset for Tekton

A beginner usually sees this:

```text
Pipeline failed.
I do not know why.
Maybe Tekton is broken?
Maybe my YAML is wrong?
Maybe Slack is wrong?
Maybe Kubernetes is wrong?
```

A calm debugger thinks like this:

```text
Something failed.
I will find the exact layer where it failed.
Then I will read the exact error message.
Then I will fix only that issue.
```

Do not start by editing random YAML.

Do not start by deleting everything.

Do not start by reinstalling Tekton.

Instead, follow the evidence.

### Tekton debugging is like following a package delivery

Imagine you ordered a package.

You would not immediately blame the delivery person. You would check:

```text
Was the order placed?
Was the package packed?
Was it shipped?
Did it reach the local hub?
Did final delivery fail?
```

In Tekton, the same idea is:

```text
Did the trigger fire?
Was a PipelineRun created?
Did TaskRuns get created?
Did Pods start?
Which step failed?
What did the logs say?
```

---

## 4. Debug order: trigger -> PipelineRun -> TaskRun -> pod -> step logs

This is the most important Day 11 idea.

```text
Trigger
  ↓
PipelineRun
  ↓
TaskRun
  ↓
Pod
  ↓
Step container logs
```

For your `slack-integration` style project, this might mean:

```text
GitHub / curl webhook event
  ↓
EventListener receives event
  ↓
PipelineRun starts
  ↓
clone-repo TaskRun starts
  ↓
run-tests TaskRun starts
  ↓
send-slack-notification TaskRun starts
  ↓
step logs show success/failure
```

### Layer 1: Trigger

Question:

```text
Did the outside event create a PipelineRun?
```

If no `PipelineRun` exists, debug the trigger.

Common causes:

| Symptom                          | Possible cause                             |
| -------------------------------- | ------------------------------------------ |
| No `PipelineRun` created         | EventListener not running                  |
| Webhook gets connection refused  | Service not exposed                        |
| Trigger receives event but fails | Bad TriggerTemplate or missing params      |
| GitHub webhook fails             | wrong URL or Minikube tunnel not reachable |

### Layer 2: PipelineRun

Question:

```text
Did the pipeline start correctly?
```

A `PipelineRun` tells you the high-level status: running, failed, succeeded, or pending. Tekton tracks execution status in the `PipelineRun` status field. ([Tekton][1])

Common causes:

| Symptom                             | Possible cause                                               |
| ----------------------------------- | ------------------------------------------------------------ |
| `PipelineRun` failed immediately    | bad pipelineRef                                              |
| `PipelineRun` pending               | missing workspace, missing service account, waiting for task |
| `PipelineRun` failed after one task | one `TaskRun` failed                                         |

### Layer 3: TaskRun

Question:

```text
Which task failed?
```

Example tasks in your project:

```text
git-clone
go-test
build-image
send-slack-notification
```

If `send-slack-notification` fails, do not debug `git-clone`.

Debug the failing task only.

### Layer 4: Pod

Question:

```text
Did Kubernetes create and run the Pod?
```

Each Tekton `TaskRun` runs through Kubernetes Pods, and Tekton logs are available from the Pod containers. ([Tekton][2])

Common Pod problems:

| Pod status         | Meaning                                             |
| ------------------ | --------------------------------------------------- |
| `Pending`          | Pod cannot be scheduled or volume/secret is missing |
| `ImagePullBackOff` | Kubernetes cannot pull the container image          |
| `ErrImagePull`     | image name, tag, registry, or auth problem          |
| `CrashLoopBackOff` | container repeatedly crashes                        |
| `Completed`        | Pod finished                                        |
| `Error`            | one or more containers failed                       |

### Layer 5: Step logs

Question:

```text
What exact command failed?
```

This is usually where the answer is.

Examples:

```text
go test ./...
```

failed because code has a bug.

```text
curl -X POST "$SLACK_WEBHOOK_URL"
```

failed because the secret is missing or the URL is invalid.

```text
docker build ...
```

failed because image build context is wrong.

---

## 5. Useful commands: `kubectl get`, `describe`, `logs`, and `tkn`

Minikube configures `kubectl` to talk to the Minikube cluster after `minikube start`; if local `kubectl` is not installed, Minikube can run it through `minikube kubectl --`. ([minikube][3])

Use this namespace variable throughout:

```bash
NS=slack-demo
```

Replace `slack-demo` with your actual namespace.

---

### A. `kubectl get`

Use `get` to answer:

```text
What exists?
What is its status?
What is its name?
```

Examples:

```bash
kubectl get namespaces
kubectl get pods -n $NS
kubectl get pipelineruns -n $NS
kubectl get taskruns -n $NS
kubectl get secrets -n $NS
kubectl get serviceaccounts -n $NS
```

Beginner translation:

```text
kubectl get = show me the list
```

Python mental model:

```python
items = list_resources(namespace="slack-demo")
print(items)
```

---

### B. `kubectl describe`

Use `describe` to answer:

```text
Why is this object unhappy?
What events happened?
What error message did Kubernetes record?
```

Kubernetes describes this command as showing detailed information, including related resources such as events or controllers. ([Kubernetes][4])

Examples:

```bash
kubectl describe pipelinerun <pipelinerun-name> -n $NS
kubectl describe taskrun <taskrun-name> -n $NS
kubectl describe pod <pod-name> -n $NS
kubectl describe secret slack-webhook-secret -n $NS
kubectl describe serviceaccount pipeline -n $NS
```

Beginner translation:

```text
kubectl describe = tell me the story of this object
```

Python mental model:

```python
details = get_object_details("pod", pod_name)
print(details.events)
```

---

### C. `kubectl logs`

Use `logs` to answer:

```text
What did the container print?
What exact command failed?
```

`kubectl logs` prints logs for a container in a Pod; if the Pod has multiple containers, you can choose one with `-c`. ([Kubernetes][5])

Examples:

```bash
kubectl logs <pod-name> -n $NS
kubectl logs <pod-name> -c step-run-tests -n $NS
kubectl logs <pod-name> --all-containers=true -n $NS
kubectl logs <pod-name> -c step-send-slack -n $NS
kubectl logs <pod-name> -c step-send-slack --previous -n $NS
```

Useful flags:

| Flag                    | Meaning                              |
| ----------------------- | ------------------------------------ |
| `-c <container>`        | show logs for one container          |
| `-f`                    | follow live logs                     |
| `--all-containers=true` | show logs from all containers        |
| `--tail=50`             | show last 50 lines                   |
| `--previous`            | show previous crashed container logs |
| `--since=10m`           | show logs from last 10 minutes       |

---

### D. `tkn`

`tkn` is the Tekton CLI. It is easier than raw `kubectl` for Tekton-specific resources.

Examples:

```bash
tkn pipelinerun list -n $NS
tkn pipelinerun describe <pipelinerun-name> -n $NS
tkn pipelinerun logs <pipelinerun-name> -n $NS
tkn pipelinerun logs <pipelinerun-name> -f -n $NS
tkn pipelinerun logs <pipelinerun-name> -a -n $NS
tkn taskrun list -n $NS
tkn taskrun describe <taskrun-name> -n $NS
```

The Tekton CLI docs show `tkn pipelinerun logs <name> -n <namespace>`, `tkn pr logs <name> -t <task>`, and `tkn pr logs <name> -a` for all tasks and steps. ([GitHub][6])

Beginner translation:

```text
kubectl = general Kubernetes tool
tkn = Tekton-focused helper tool
```

---

## 6. ASCII debugging decision flow

```text
START
  |
  v
Did you expect a run to start from a webhook/trigger?
  |
  +-- YES ------------------------------------------------+
  |                                                      |
  |                                                      v
  |                                      Is a PipelineRun created?
  |                                                      |
  |                         +----------------------------+------------------+
  |                         |                                               |
  |                        NO                                              YES
  |                         |                                               |
  |                         v                                               v
  |             Debug Trigger/EventListener                     Inspect PipelineRun
  |             - EventListener running?                         kubectl describe pr
  |             - Service reachable?                             tkn pr describe
  |             - TriggerTemplate valid?                         tkn pr logs
  |             - params mapped correctly?
  |
  +-- NO
        |
        v
  Did you manually create a PipelineRun?
        |
        v
  Inspect PipelineRun
        |
        v
  Is PipelineRun Failed?
        |
        +-- NO --> Is it Pending?
        |           |
        |           +-- YES --> Check workspace, SA, PVC, resources, Task refs
        |           |
        |           +-- NO --> Follow logs while it runs
        |
        +-- YES
              |
              v
       Which TaskRun failed?
              |
              v
       Describe TaskRun
              |
              v
       Find Pod
              |
              v
       Is Pod Pending/ImagePullBackOff?
              |
              +-- YES --> Kubernetes/infra/config issue
              |
              +-- NO
                    |
                    v
             Inspect step logs
                    |
                    v
             Is error from your app/test command?
                    |
                    +-- YES --> Code error
                    |
                    +-- NO --> Config/secret/RBAC/infra error
```

---

## 7. Pseudocode for how to debug a failing run

This is the mental algorithm.

```text
function debugTektonRun(namespace):

    check minikube status
    check Tekton controller pods

    list PipelineRuns in namespace

    if no PipelineRun exists:
        debug trigger
        check EventListener
        check service
        check TriggerTemplate
        check TriggerBinding
        stop

    choose latest PipelineRun

    describe PipelineRun

    if PipelineRun says missing pipeline/task/param/workspace:
        fix YAML/config
        stop

    list TaskRuns owned by PipelineRun

    find failed TaskRun

    describe failed TaskRun

    find podName from TaskRun status

    describe Pod

    if Pod is Pending:
        check secrets, service account, PVC, resources
        stop

    if Pod has ImagePullBackOff:
        check image name, image tag, registry auth
        stop

    list containers in Pod

    read logs for failed step container

    if logs show Go test failure:
        fix Go code

    else if logs show missing environment variable:
        fix Secret or Task env mapping

    else if logs show forbidden:
        fix ServiceAccount or RBAC

    else if logs show network timeout:
        check Minikube network, DNS, external service

    rerun PipelineRun

    compare old error with new result
```

Python comparison:

```python
def debug_tekton_run(namespace):
    if not pipelinerun_exists(namespace):
        return debug_trigger(namespace)

    pr = latest_pipelinerun(namespace)
    if pr.failed:
        tr = failed_taskrun(pr)
        pod = pod_for_taskrun(tr)
        logs = logs_for_failed_step(pod)
        return classify_error(logs)
```

Go mindset difference:

In Python, you might write quick exploratory scripts. In Go, you usually make the control flow more explicit:

```go
if err != nil {
    return err
}
```

That same explicit habit is useful in debugging Tekton: check one layer, handle the error, then move to the next layer.

---

## 8. Real debugging command examples

Assume your project namespace is:

```bash
NS=slack-demo
```

---

### Step 1: Check Minikube

```bash
minikube status
kubectl config current-context
kubectl get nodes
```

Expected healthy output shape:

```text
host: Running
kubelet: Running
apiserver: Running

minikube

NAME       STATUS   ROLES           AGE   VERSION
minikube   Ready    control-plane    ...
```

If this fails, do not debug Tekton yet. Your local cluster itself is not healthy.

---

### Step 2: Check Tekton system pods

```bash
kubectl get pods -n tekton-pipelines
```

Expected output shape:

```text
NAME                                           READY   STATUS    RESTARTS
tekton-pipelines-controller-xxxxx             1/1     Running   0
tekton-pipelines-webhook-xxxxx                1/1     Running   0
```

If these are not running, your `PipelineRun` may not reconcile.

That means Tekton may not be creating `TaskRuns`.

---

### Step 3: Check project resources

```bash
kubectl get pipelines,tasks,pipelineruns,taskruns -n $NS
```

Expected output shape:

```text
NAME                                      AGE
pipeline.tekton.dev/slack-integration    2d

NAME                                  AGE
task.tekton.dev/git-clone             2d
task.tekton.dev/go-test               2d
task.tekton.dev/send-slack-message    2d

NAME                                                SUCCEEDED   REASON
pipelinerun.tekton.dev/slack-integration-run-abc    False       Failed
```

---

### Step 4: Describe the latest PipelineRun

```bash
kubectl describe pipelinerun slack-integration-run-abc -n $NS
```

Look for:

```text
Status:
Conditions:
  Type: Succeeded
  Status: False
  Reason: Failed
  Message: Tasks Completed: 2, Failed: 1
```

This tells you:

```text
The whole pipeline failed because one task failed.
```

Now find the failing `TaskRun`.

---

### Step 5: List TaskRuns

```bash
kubectl get taskruns -n $NS
```

Example:

```text
NAME                                  SUCCEEDED   REASON
slack-run-abc-git-clone               True        Succeeded
slack-run-abc-go-test                 False       Failed
slack-run-abc-send-slack-message      Unknown     TaskRunCancelled
```

Interpretation:

```text
git-clone worked.
go-test failed.
send-slack-message did not run because previous task failed.
```

So you should debug `go-test`.

---

### Step 6: Describe failed TaskRun

```bash
kubectl describe taskrun slack-run-abc-go-test -n $NS
```

Look for:

```text
Pod Name: slack-run-abc-go-test-pod
Steps:
  run-tests:
    Terminated:
      Exit Code: 1
```

This means:

```text
The run-tests step exited with code 1.
Now read the logs for that step.
```

---

### Step 7: Get the Pod

```bash
kubectl get pods -n $NS
```

Example:

```text
NAME                            READY   STATUS   RESTARTS
slack-run-abc-go-test-pod       0/1     Error    0
```

Now describe the Pod:

```bash
kubectl describe pod slack-run-abc-go-test-pod -n $NS
```

Look for Events:

```text
Events:
  Warning  FailedMount  secret "slack-webhook-secret" not found
```

That is a configuration error.

Or:

```text
Events:
  Warning  Failed     Failed to pull image "golang:badtag"
```

That is image / infra / config.

---

### Step 8: Inspect step logs

```bash
kubectl logs slack-run-abc-go-test-pod -c step-run-tests -n $NS
```

Example output:

```text
=== RUN   TestBuildSlackPayload
--- FAIL: TestBuildSlackPayload
    payload_test.go:18: expected channel #alerts, got empty channel
FAIL
```

This is a code error.

The pipeline worked. Kubernetes worked. Tekton worked. Your test failed.

---

### Step 9: Use `tkn` for easier logs

```bash
tkn pipelinerun logs slack-integration-run-abc -n $NS
```

Follow live logs:

```bash
tkn pipelinerun logs slack-integration-run-abc -f -n $NS
```

Show all logs:

```bash
tkn pipelinerun logs slack-integration-run-abc -a -n $NS
```

Show only one task:

```bash
tkn pr logs slack-integration-run-abc -t go-test -n $NS
```

---

## 9. Hands-on tasks

### Task A: Check your local cluster

Run:

```bash
minikube status
kubectl get nodes
kubectl get pods -A
```

Goal:

```text
Confirm Minikube and Kubernetes are healthy.
```

---

### Task B: Create or rerun a PipelineRun

Example:

```bash
kubectl apply -f tekton/pipelinerun.yaml -n $NS
```

Then watch:

```bash
kubectl get pipelineruns -n $NS -w
```

Stop watching with:

```text
Ctrl + C
```

---

### Task C: Find the failed TaskRun

```bash
kubectl get taskruns -n $NS
```

Then:

```bash
kubectl describe taskrun <failed-taskrun-name> -n $NS
```

Write down:

```text
TaskRun name:
Reason:
Message:
Pod name:
Failed step:
Exit code:
```

---

### Task D: Inspect logs

```bash
kubectl logs <pod-name> --all-containers=true -n $NS
```

Then inspect only the failing step:

```bash
kubectl logs <pod-name> -c <step-container-name> -n $NS
```

Example:

```bash
kubectl logs slack-run-abc-go-test-pod -c step-run-tests -n $NS
```

---

### Task E: Break and fix a secret issue

Temporarily rename your Slack secret reference in the Task:

```yaml
secretKeyRef:
  name: wrong-slack-secret
  key: webhook-url
```

Run the pipeline.

Then inspect:

```bash
kubectl describe taskrun <taskrun-name> -n $NS
kubectl describe pod <pod-name> -n $NS
```

Expected idea:

```text
Kubernetes should complain that the secret does not exist.
```

Fix it back:

```yaml
secretKeyRef:
  name: slack-webhook-secret
  key: webhook-url
```

---

### Task F: Break and fix a Go test

In your Go project, intentionally make a test fail.

Example:

```go
if got != "#alerts" {
    t.Fatalf("expected #alerts, got %s", got)
}
```

Change expected value to something wrong:

```go
if got != "#wrong-channel" {
    t.Fatalf("expected #wrong-channel, got %s", got)
}
```

Run pipeline.

Expected failure:

```text
go test ./... fails
```

Then fix the test.

---

## 10. Expected output

### Healthy Minikube

```text
host: Running
kubelet: Running
apiserver: Running
```

### Healthy Tekton

```text
tekton-pipelines-controller   Running
tekton-pipelines-webhook      Running
```

### Successful PipelineRun

```text
NAME                         SUCCEEDED   REASON
slack-integration-run-abc    True        Succeeded
```

### Failed PipelineRun

```text
NAME                         SUCCEEDED   REASON
slack-integration-run-abc    False       Failed
```

### Failed TaskRun from code error

```text
NAME                     SUCCEEDED   REASON
slack-run-abc-go-test    False       Failed
```

Logs:

```text
FAIL    ./internal/slack
```

### Failed TaskRun from secret issue

Pod describe:

```text
Error: couldn't find key webhook-url in Secret slack-webhook-secret
```

or:

```text
secret "slack-webhook-secret" not found
```

### Failed TaskRun from service account issue

Pod or TaskRun event may show:

```text
forbidden
cannot get resource
serviceaccount not found
```

Meaning:

```text
Your Task is trying to do something Kubernetes does not allow.
```

---

## 11. Common mistakes

### Mistake 1: Debugging logs before checking whether a run exists

Bad flow:

```text
Pipeline failed? Immediately check pod logs.
```

Better flow:

```text
First check whether PipelineRun exists.
Then TaskRun.
Then Pod.
Then logs.
```

---

### Mistake 2: Confusing PipelineRun and Pipeline

A `Pipeline` is the recipe.

A `PipelineRun` is one execution of the recipe.

Python comparison:

```python
def pipeline():
    clone()
    test()
    notify()

pipeline()  # this call is like PipelineRun
```

Go comparison:

```go
func Pipeline() {
    Clone()
    Test()
    Notify()
}

func main() {
    Pipeline() // this execution is like a PipelineRun
}
```

---

### Mistake 3: Thinking every failure is a Tekton problem

Many failures are not Tekton problems.

| Error type          | Example              | Fix                                 |
| ------------------- | -------------------- | ----------------------------------- |
| Configuration error | wrong secret name    | fix YAML                            |
| Code error          | Go test fails        | fix Go code                         |
| Infra error         | Minikube not running | fix cluster/runtime                 |
| Auth error          | RBAC forbidden       | fix ServiceAccount/Role/RoleBinding |
| Image error         | `ImagePullBackOff`   | fix image/tag/registry auth         |

---

### Mistake 4: Not checking namespace

This is very common.

Wrong:

```bash
kubectl get pipelineruns
```

Better:

```bash
kubectl get pipelineruns -n $NS
```

Or check all namespaces:

```bash
kubectl get pipelineruns -A
```

---

### Mistake 5: Looking at the wrong Pod

A pipeline can create multiple Pods.

Always connect them:

```text
PipelineRun -> TaskRun -> Pod
```

Do not randomly pick a Pod.

---

### Mistake 6: Forgetting that secret values are base64 encoded

Check Secret exists:

```bash
kubectl get secret slack-webhook-secret -n $NS
```

Check keys:

```bash
kubectl get secret slack-webhook-secret -n $NS -o yaml
```

You may see:

```yaml
data:
  webhook-url: aHR0cHM6Ly9ob29rcy5zbGFjay5jb20v...
```

That value is encoded.

Do not paste secrets into screenshots or GitHub.

---

## 12. Debugging tips

### Tip 1: Always write down the exact failure

Instead of saying:

```text
Pipeline failed.
```

Say:

```text
PipelineRun slack-integration-run-abc failed because TaskRun go-test failed.
The Pod completed with Error.
The step-run-tests container exited with code 1.
The log says TestBuildSlackPayload failed.
```

That sentence tells you exactly what to fix.

---

### Tip 2: Read `Events` at the bottom of `describe`

For Kubernetes beginners, the bottom of `kubectl describe` is gold.

Example:

```bash
kubectl describe pod <pod-name> -n $NS
```

Scroll to:

```text
Events:
```

That often tells you:

```text
missing secret
cannot pull image
forbidden
failed mount
insufficient cpu
```

---

### Tip 3: Learn to classify errors

#### Configuration error

Something in YAML or Kubernetes object setup is wrong.

Examples:

```text
wrong taskRef
wrong pipelineRef
missing param
wrong secret name
wrong workspace name
wrong serviceAccountName
```

Fix:

```text
Edit YAML.
Reapply.
Rerun.
```

#### Code error

Your application/test/build command failed.

Examples:

```text
go test ./... failed
go build failed
JSON payload test failed
Slack message builder returned wrong output
```

Fix:

```text
Edit Go code.
Run tests locally.
Commit.
Rerun pipeline.
```

#### Infrastructure error

The platform underneath is unhealthy or unreachable.

Examples:

```text
minikube stopped
node NotReady
image registry unreachable
DNS failure
network timeout
not enough memory
```

Fix:

```text
Fix cluster, runtime, network, registry, or resources.
```

---

### Tip 4: Use local commands before blaming Tekton

Before running the full pipeline, test locally:

```bash
go test ./...
go build ./...
```

Python comparison:

```bash
pytest
python -m build
```

Go equivalent:

```bash
go test ./...
go build ./...
```

If it fails locally, it will probably fail in Tekton too.

---

### Tip 5: Use `kubectl get -o yaml` when `describe` is not enough

```bash
kubectl get pipelinerun <name> -n $NS -o yaml
kubectl get taskrun <name> -n $NS -o yaml
```

This shows full status fields.

Useful for finding:

```text
podName
conditions
reason
message
taskSpec
params
serviceAccountName
```

---

## 13. DSA topic: heap / priority queue basics

A priority queue is like a normal queue, but items with higher priority come out first.

Normal queue:

```text
first in -> first out
```

Priority queue:

```text
most important -> first out
```

Example for your project:

```text
Critical production failure     priority 1
Failed Slack notification       priority 2
Test failure                    priority 3
Lint warning                    priority 4
```

Lower number can mean higher priority.

### Real CI/CD use case

Imagine your pipeline collects failures:

```text
- Slack webhook secret missing
- Go tests failed
- Image pull failed
- Lint failed
```

You do not want to handle them randomly.

You want:

```text
1. Secret missing
2. Image pull failed
3. Go tests failed
4. Lint failed
```

Because some errors block everything else.

### Python equivalent

Python has `heapq`:

```python
import heapq

queue = []
heapq.heappush(queue, (1, "secret missing"))
heapq.heappush(queue, (3, "go test failed"))
heapq.heappush(queue, (2, "image pull failed"))

print(heapq.heappop(queue))
# (1, 'secret missing')
```

### Go equivalent

Go uses the `container/heap` package.

Main difference:

| Python                        | Go                        |
| ----------------------------- | ------------------------- |
| `heapq.heappush(queue, item)` | `heap.Push(&queue, item)` |
| tuple priority works easily   | define a struct           |
| dynamic list                  | custom slice type         |
| less boilerplate              | more explicit methods     |

In Go, you implement these methods:

```go
Len()
Less()
Swap()
Push()
Pop()
```

That feels verbose at first, but it teaches you exactly how the heap works.

---

## 14. One Go DSA problem: priority notification queue

### Problem

Build a priority queue for CI/CD notifications.

Each notification has:

```text
Message
Priority
```

Smaller priority number means more urgent.

Example:

```text
Priority 1: "Slack secret missing"
Priority 2: "Image pull failed"
Priority 3: "Go test failed"
```

Expected output:

```text
Slack secret missing
Image pull failed
Go test failed
```

---

### Go solution

```go
package main

import (
	"container/heap"
	"fmt"
)

// Notification is like a Python dict/object:
// {"message": "...", "priority": 1}
//
// In Go, we usually define a struct for this.
type Notification struct {
	Message  string
	Priority int
}

// NotificationQueue is a slice of Notification.
// Python equivalent: list[Notification]
type NotificationQueue []Notification

// Len is required by heap.Interface.
// Python equivalent: len(queue)
func (q NotificationQueue) Len() int {
	return len(q)
}

// Less decides priority order.
// q[i].Priority < q[j].Priority means smaller number comes first.
func (q NotificationQueue) Less(i, j int) bool {
	return q[i].Priority < q[j].Priority
}

// Swap swaps two items.
// Python lists do this internally; Go asks us to define it.
func (q NotificationQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

// Push adds an item.
// Important Go syntax note:
// receiver is *NotificationQueue because we modify the slice.
func (q *NotificationQueue) Push(x any) {
	item := x.(Notification)
	*q = append(*q, item)
}

// Pop removes the last item from the internal heap.
// The heap package moves the highest-priority item to the end before calling Pop.
func (q *NotificationQueue) Pop() any {
	old := *q
	n := len(old)
	item := old[n-1]
	*q = old[0 : n-1]
	return item
}

func main() {
	queue := &NotificationQueue{}

	heap.Init(queue)

	heap.Push(queue, Notification{
		Message:  "Go test failed",
		Priority: 3,
	})

	heap.Push(queue, Notification{
		Message:  "Slack secret missing",
		Priority: 1,
	})

	heap.Push(queue, Notification{
		Message:  "Image pull failed",
		Priority: 2,
	})

	for queue.Len() > 0 {
		item := heap.Pop(queue).(Notification)
		fmt.Println(item.Message)
	}
}
```

Expected output:

```text
Slack secret missing
Image pull failed
Go test failed
```

### Important Go concepts compared with Python

#### `struct`

Go:

```go
type Notification struct {
	Message  string
	Priority int
}
```

Python:

```python
@dataclass
class Notification:
    message: str
    priority: int
```

Go convention:

```text
Use PascalCase for exported fields: Message, Priority.
Use camelCase for private/internal fields: message, priority.
```

---

#### `any`

Go:

```go
func (q *NotificationQueue) Push(x any)
```

Python equivalent:

```python
def push(x: Any):
```

In Go, `any` means:

```text
this can hold any type
```

But then we must convert it back:

```go
item := x.(Notification)
```

Python usually does not need that because it is dynamically typed.

---

#### Pointer receiver

Go:

```go
func (q *NotificationQueue) Push(x any)
```

Why pointer?

Because `Push` changes the queue.

Python equivalent:

```python
queue.append(item)
```

Python lists are mutable by default. Go makes mutation more explicit.

---

## 15. Module-based practice task: build a priority notification/failure queue

Create this mini module inside your project:

```text
slack-integration/
  go.mod
  internal/
    failurequeue/
      queue.go
      queue_test.go
  cmd/
    failure-demo/
      main.go
```

Initialize module if needed:

```bash
go mod init slack-integration
```

---

### `internal/failurequeue/queue.go`

```go
package failurequeue

import "container/heap"

type Failure struct {
	Source   string
	Message  string
	Priority int
}

type Queue []Failure

func (q Queue) Len() int {
	return len(q)
}

func (q Queue) Less(i, j int) bool {
	return q[i].Priority < q[j].Priority
}

func (q Queue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q *Queue) Push(x any) {
	item := x.(Failure)
	*q = append(*q, item)
}

func (q *Queue) Pop() any {
	old := *q
	n := len(old)
	item := old[n-1]
	*q = old[:n-1]
	return item
}

func New() *Queue {
	q := &Queue{}
	heap.Init(q)
	return q
}

func Add(q *Queue, failure Failure) {
	heap.Push(q, failure)
}

func Next(q *Queue) Failure {
	return heap.Pop(q).(Failure)
}
```

---

### `internal/failurequeue/queue_test.go`

```go
package failurequeue

import "testing"

func TestFailureQueueReturnsHighestPriorityFirst(t *testing.T) {
	q := New()

	Add(q, Failure{
		Source:   "go-test",
		Message:  "unit test failed",
		Priority: 3,
	})

	Add(q, Failure{
		Source:   "send-slack",
		Message:  "slack secret missing",
		Priority: 1,
	})

	Add(q, Failure{
		Source:   "build-image",
		Message:  "image pull failed",
		Priority: 2,
	})

	first := Next(q)

	if first.Message != "slack secret missing" {
		t.Fatalf("expected slack secret missing, got %s", first.Message)
	}
}
```

Run:

```bash
go test ./...
```

Expected:

```text
ok  	slack-integration/internal/failurequeue
```

---

### `cmd/failure-demo/main.go`

```go
package main

import (
	"fmt"

	"slack-integration/internal/failurequeue"
)

func main() {
	q := failurequeue.New()

	failurequeue.Add(q, failurequeue.Failure{
		Source:   "go-test",
		Message:  "Go tests failed",
		Priority: 3,
	})

	failurequeue.Add(q, failurequeue.Failure{
		Source:   "send-slack",
		Message:  "Slack webhook secret missing",
		Priority: 1,
	})

	failurequeue.Add(q, failurequeue.Failure{
		Source:   "build-image",
		Message:  "Image pull failed",
		Priority: 2,
	})

	for q.Len() > 0 {
		f := failurequeue.Next(q)
		fmt.Printf("[%s] %s\n", f.Source, f.Message)
	}
}
```

Run:

```bash
go run ./cmd/failure-demo
```

Expected:

```text
[send-slack] Slack webhook secret missing
[build-image] Image pull failed
[go-test] Go tests failed
```

This module connects directly to today’s Tekton lesson:

```text
Not all failures are equal.
Debug the highest-priority blocking failure first.
```

---

## 16. Revision checkpoint

Answer these without looking:

1. What is the difference between a `Pipeline` and a `PipelineRun`?
2. What creates `TaskRuns`?
3. Where do Tekton step logs live?
4. What command lists PipelineRuns?
5. What command explains why a Pod is unhappy?
6. What does `ImagePullBackOff` usually mean?
7. What does `forbidden` usually mean?
8. What does a missing Slack webhook secret indicate?
9. What is the difference between a configuration error and a code error?
10. In a priority queue, which item comes out first?

Expected answers:

```text
1. Pipeline is the recipe; PipelineRun is one execution.
2. PipelineRun creates TaskRuns.
3. In Pod container logs.
4. kubectl get pipelineruns -n <namespace>
5. kubectl describe pod <pod-name> -n <namespace>
6. Kubernetes cannot pull the image.
7. ServiceAccount/RBAC permission problem.
8. Configuration/secret issue.
9. Config error is setup/YAML/Kubernetes; code error is app/test/build logic.
10. The item with highest priority.
```

---

## 17. Homework

### CI/CD homework

Run one successful pipeline and one intentionally failing pipeline.

For the failing one, write a debug report:

```text
PipelineRun name:
PipelineRun status:
Failed TaskRun:
Pod name:
Failed step:
Exit code:
Exact log error:
Error category:
Fix applied:
Result after rerun:
```

Use these commands:

```bash
kubectl get pipelineruns -n $NS
kubectl describe pipelinerun <name> -n $NS
kubectl get taskruns -n $NS
kubectl describe taskrun <name> -n $NS
kubectl get pods -n $NS
kubectl describe pod <name> -n $NS
kubectl logs <pod-name> --all-containers=true -n $NS
tkn pipelinerun logs <pipelinerun-name> -n $NS
```

### Go DSA homework

Extend the failure queue so that if two failures have the same priority, the earlier one comes first.

Hint:

Add an `Order` field:

```go
type Failure struct {
	Source   string
	Message  string
	Priority int
	Order    int
}
```

Then update `Less`:

```go
func (q Queue) Less(i, j int) bool {
	if q[i].Priority == q[j].Priority {
		return q[i].Order < q[j].Order
	}
	return q[i].Priority < q[j].Priority
}
```

### Final Day 11 habit

Before fixing anything, say this out loud:

```text
I will not guess.
I will find the failing layer.
I will read the exact error.
I will fix one issue at a time.
```

[1]: https://tekton.dev/docs/pipelines/pipelineruns/ "Tekton"
[2]: https://tekton.dev/docs/pipelines/logs/ "Tekton"
[3]: https://minikube.sigs.k8s.io/docs/handbook/kubectl/ "Kubectl | minikube"
[4]: https://kubernetes.io/docs/reference/kubectl/generated/kubectl_describe/ "kubectl describe | Kubernetes"
[5]: https://kubernetes.io/docs/reference/kubectl/generated/kubectl_logs/ "kubectl logs | Kubernetes"
[6]: https://github.com/tektoncd/cli/blob/main/docs/cmd/tkn_pipelinerun_logs.md "cli/docs/cmd/tkn_pipelinerun_logs.md at main · tektoncd/cli · GitHub"
---
## Extra Notes for debugging
# Tekton + Kubernetes Learning Notes

## 1. Difference between `kubectl` and `tkn`

### `kubectl`

`kubectl` is used for Kubernetes resources:

```bash
kubectl get pods
kubectl get secrets
kubectl get serviceaccounts
kubectl get nodes
kubectl describe pod <pod-name>
```

### `tkn`

`tkn` is the Tekton CLI. It is used for Tekton resources:

```bash
tkn task list
tkn taskrun list
tkn pipeline list
tkn pipelinerun list
tkn taskrun logs <taskrun-name>
tkn pipelinerun logs <pipelinerun-name>
```

This command is wrong:

```bash
tkn get pods
```

Because `pods` are Kubernetes resources, not Tekton resources.

Use this instead:

```bash
kubectl get pods
```

---

# 2. Tekton Objects

## Task

A `Task` is only a definition.

Example:

```bash
kubectl get tasks
```

Output:

```text
NAME                    AGE
git-clone               24m
git-clone-repo          24m
git-set-commit-status   24m
slack-notify            24m
task-validate-check     24m
```

This does **not** show whether the task passed or failed.

---

## TaskRun

A `TaskRun` is the actual execution of a `Task`.

To check task execution status:

```bash
kubectl get taskruns -n slack-integration-dev
```

or:

```bash
kubectl get tr -n slack-integration-dev
```

Example:

```text
NAME                                                              SUCCEEDED   REASON
pipelinerun-xxx-validate-check                                    False       CreateContainerConfigError
pipelinerun-xxx-notify-pr-final                                   True        Succeeded
```

Meaning:

```text
SUCCEEDED=True    → task passed
SUCCEEDED=False   → task failed
REASON=Succeeded  → successful
REASON=Failed     → failed
REASON=CreateContainerConfigError → container could not start
```

---

## PipelineRun

A `PipelineRun` is the actual execution of a `Pipeline`.

To check pipeline status:

```bash
kubectl get pipelineruns -n slack-integration-dev
```

or:

```bash
tkn pipelinerun list -n slack-integration-dev
```

---

# 3. How to Check Task Status

Use:

```bash
kubectl get taskruns -n slack-integration-dev
```

For a better view:

```bash
kubectl get taskruns -n slack-integration-dev \
  -o custom-columns=NAME:.metadata.name,TASK:.spec.taskRef.name,STATUS:.status.conditions[0].status,REASON:.status.conditions[0].reason
```

To describe one failed TaskRun:

```bash
kubectl describe taskrun <taskrun-name> -n slack-integration-dev
```

Example:

```bash
kubectl describe taskrun pipelinerun-3268a300-f944-49c8-bb1f-a3c5f15caa72-validate-check \
  -n slack-integration-dev
```

---

# 4. How to See Logs

For a Tekton TaskRun:

```bash
tkn taskrun logs <taskrun-name> -n slack-integration-dev
```

Example:

```bash
tkn taskrun logs pipelinerun-3268a300-f944-49c8-bb1f-a3c5f15caa72-validate-check \
  -n slack-integration-dev
```

For a Kubernetes pod:

```bash
kubectl logs <pod-name> -n slack-integration-dev
```

For all containers in the pod:

```bash
kubectl logs <pod-name> -n slack-integration-dev --all-containers
```

Example:

```bash
kubectl logs pipelinerun-3268a300-f944-41524e318cf4c48a145a64e7fef522f8c-pod \
  -n slack-integration-dev \
  --all-containers
```

---

# 5. Important: `CreateContainerConfigError`

If you see this:

```text
CreateContainerConfigError
```

It means the container could not start.

So logs may be empty because the container never ran.

In that case, the most useful command is:

```bash
kubectl describe pod <pod-name> -n slack-integration-dev
```

Example:

```bash
kubectl describe pod pipelinerun-3268a300-f944-41524e318cf4c48a145a64e7fef522f8c-pod \
  -n slack-integration-dev
```

Check the bottom section:

```text
Events:
```

Common causes:

```text
secret not found
configmap not found
couldn't find key in Secret
invalid environment variable
volume mount issue
service account issue
image pull secret issue
```

---

# 6. How to Check Pods

List pods:

```bash
kubectl get pods -n slack-integration-dev
```

Example:

```text
NAME                                                              READY   STATUS
el-pr-listener-74b6c785bb-66v85                                   1/1     Running
pipelinerun-xxx-pod                                               0/1     CreateContainerConfigError
pipelinerun-yyy-pod                                               0/1     Completed
```

Meaning:

```text
Running                     → pod is running
Completed                   → pod finished successfully
CreateContainerConfigError  → pod could not start due to config issue
Error                       → pod failed
CrashLoopBackOff            → container keeps crashing
ImagePullBackOff            → image cannot be pulled
```

Describe a pod:

```bash
kubectl describe pod <pod-name> -n slack-integration-dev
```

View pod logs:

```bash
kubectl logs <pod-name> -n slack-integration-dev --all-containers
```

---

# 7. How to Check Kubernetes Context

This command is wrong:

```bash
kubectl get context
```

Because `context` is not a Kubernetes resource.

Use:

```bash
kubectl config current-context
```

To list all contexts:

```bash
kubectl config get-contexts
```

Expected minikube context:

```text
CURRENT   NAME       CLUSTER    AUTHINFO   NAMESPACE
*         minikube   minikube   minikube   default
```

Switch to minikube:

```bash
kubectl config use-context minikube
```

Verify:

```bash
kubectl config current-context
kubectl get nodes
```

Check minikube status:

```bash
minikube status
```

List minikube profiles:

```bash
minikube profile list
```

---

# 8. Minikube Resources

Original resources:

```bash
readonly MINIKUBE_CPUS=4
readonly MINIKUBE_MEMORY=8192
readonly MINIKUBE_DISK_SIZE=30g
```

Doubled resources:

```bash
readonly MINIKUBE_CPUS=8
readonly MINIKUBE_MEMORY=16384
readonly MINIKUBE_DISK_SIZE=60g
```

Start minikube:

```bash
minikube start \
  --cpus="${MINIKUBE_CPUS}" \
  --memory="${MINIKUBE_MEMORY}" \
  --disk-size="${MINIKUBE_DISK_SIZE}"
```

If minikube already exists and you want to recreate with new resources:

```bash
minikube stop
minikube delete

minikube start \
  --cpus=8 \
  --memory=16384 \
  --disk-size=60g
```

Verify resources:

```bash
minikube status
kubectl get nodes
kubectl describe node minikube
```

Check capacity:

```bash
kubectl describe node minikube | grep -A 8 "Capacity"
```

Important: only allocate these resources if your machine has enough CPU and RAM.

---

# 9. How to See Available Credentials

In Kubernetes, credentials are usually stored as **Secrets** and attached through **ServiceAccounts**.

List secrets:

```bash
kubectl get secrets -n slack-integration-dev
```

Example:

```text
NAME                   TYPE     DATA   AGE
git-credentials        Opaque   2      2m10s
slack-webhook-secret   Opaque   6      2m10s
```

List service accounts:

```bash
kubectl get serviceaccounts -n slack-integration-dev
```

or:

```bash
kubectl get sa -n slack-integration-dev
```

Describe default service account:

```bash
kubectl describe sa default -n slack-integration-dev
```

Check which service account a TaskRun is using:

```bash
kubectl get taskrun <taskrun-name> \
  -n slack-integration-dev \
  -o jsonpath='{.spec.serviceAccountName}{"\n"}'
```

Then describe that service account:

```bash
kubectl describe sa <service-account-name> -n slack-integration-dev
```

---

# 10. How to See Each Secret

You had these secrets:

```text
git-credentials
slack-webhook-secret
```

Describe secret without showing actual values:

```bash
kubectl describe secret git-credentials -n slack-integration-dev
kubectl describe secret slack-webhook-secret -n slack-integration-dev
```

Show YAML with base64-encoded values:

```bash
kubectl get secret git-credentials -n slack-integration-dev -o yaml
kubectl get secret slack-webhook-secret -n slack-integration-dev -o yaml
```

Show only keys using `jq`:

```bash
kubectl get secret git-credentials -n slack-integration-dev -o jsonpath='{.data}' | jq 'keys'
kubectl get secret slack-webhook-secret -n slack-integration-dev -o jsonpath='{.data}' | jq 'keys'
```

Without `jq`:

```bash
kubectl get secret git-credentials -n slack-integration-dev -o jsonpath='{.data}'
kubectl get secret slack-webhook-secret -n slack-integration-dev -o jsonpath='{.data}'
```

Decode one secret key:

```bash
kubectl get secret <secret-name> -n slack-integration-dev \
  -o jsonpath='{.data.<key-name>}' | base64 --decode
```

Example:

```bash
kubectl get secret git-credentials -n slack-integration-dev \
  -o jsonpath='{.data.username}' | base64 --decode
```

Example:

```bash
kubectl get secret git-credentials -n slack-integration-dev \
  -o jsonpath='{.data.password}' | base64 --decode
```

For Slack secret, first check the real key names:

```bash
kubectl describe secret slack-webhook-secret -n slack-integration-dev
```

Then decode a key:

```bash
kubectl get secret slack-webhook-secret -n slack-integration-dev \
  -o jsonpath='{.data.<key-name>}' | base64 --decode
```

Do not paste decoded secret values in chat or screenshots.

---

# 11. Slack Notify Task Status

In Tekton, a Slack notification task is usually placed in the `finally` section.

Example:

```yaml
finally:
  - name: slack-notify
    taskRef:
      name: slack-notify
    params:
      - name: status
        value: "$(tasks.status)"
```

`$(tasks.status)` gives the overall pipeline status.

Possible values:

```text
Succeeded
Failed
Completed
None
```

To get the status of a specific task inside `finally`:

```yaml
value: "$(tasks.task-validate-check.status)"
```

Example:

```yaml
finally:
  - name: slack-notify
    taskRef:
      name: slack-notify
    params:
      - name: pipeline-status
        value: "$(tasks.status)"
      - name: validate-check-status
        value: "$(tasks.task-validate-check.status)"
```

---

# 12. Useful Debugging Flow

When something fails, follow this order.

## Step 1: Check pods

```bash
kubectl get pods -n slack-integration-dev
```

## Step 2: Check TaskRuns

```bash
kubectl get taskruns -n slack-integration-dev
```

## Step 3: Find failed TaskRun

Look for:

```text
SUCCEEDED=False
```

or:

```text
REASON=CreateContainerConfigError
```

## Step 4: Describe failed TaskRun

```bash
kubectl describe taskrun <taskrun-name> -n slack-integration-dev
```

## Step 5: Describe failed pod

```bash
kubectl describe pod <pod-name> -n slack-integration-dev
```

Check the `Events` section.

## Step 6: Check logs

```bash
tkn taskrun logs <taskrun-name> -n slack-integration-dev
```

or:

```bash
kubectl logs <pod-name> -n slack-integration-dev --all-containers
```

## Step 7: Check secrets

```bash
kubectl get secrets -n slack-integration-dev
kubectl describe secret <secret-name> -n slack-integration-dev
```

## Step 8: Check service account

```bash
kubectl get sa -n slack-integration-dev
kubectl describe sa <service-account-name> -n slack-integration-dev
```

---

# 13. Most Important Commands Cheat Sheet

```bash
# Check current Kubernetes context
kubectl config current-context

# List all contexts
kubectl config get-contexts

# Switch to minikube
kubectl config use-context minikube

# Check minikube
minikube status

# Check nodes
kubectl get nodes

# Check pods
kubectl get pods -n slack-integration-dev

# Check Tekton tasks
kubectl get tasks -n slack-integration-dev

# Check Tekton task executions
kubectl get taskruns -n slack-integration-dev

# Check Tekton pipeline executions
kubectl get pipelineruns -n slack-integration-dev

# Describe failed TaskRun
kubectl describe taskrun <taskrun-name> -n slack-integration-dev

# Describe failed pod
kubectl describe pod <pod-name> -n slack-integration-dev

# View TaskRun logs
tkn taskrun logs <taskrun-name> -n slack-integration-dev

# View pod logs
kubectl logs <pod-name> -n slack-integration-dev --all-containers

# List secrets
kubectl get secrets -n slack-integration-dev

# Describe secret
kubectl describe secret <secret-name> -n slack-integration-dev

# List service accounts
kubectl get sa -n slack-integration-dev
```

---

# 14. Key Learning Summary

`kubectl get tasks` shows Task definitions, not pass or fail status.

To see task pass or fail status, use:

```bash
kubectl get taskruns -n slack-integration-dev
```

To debug a failed task, use:

```bash
kubectl describe taskrun <taskrun-name> -n slack-integration-dev
```

To debug `CreateContainerConfigError`, use:

```bash
kubectl describe pod <pod-name> -n slack-integration-dev
```

For pods, always use `kubectl`, not `tkn`.

For Tekton TaskRuns and PipelineRuns, you can use either `kubectl` or `tkn`.

For credentials, check:

```bash
kubectl get secrets -n slack-integration-dev
kubectl get sa -n slack-integration-dev
```

Most failures like `CreateContainerConfigError` are usually caused by missing secrets, wrong secret keys, missing config maps, bad environment variables, or service account configuration issues.
