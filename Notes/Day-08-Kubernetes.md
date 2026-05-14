# Day 8 — Kubernetes Fundamentals for `slack-integration`

## 1. Day 8 learning goals

By the end of Day 8, you should understand:

1. What Kubernetes is and why teams use it.
2. What a cluster, node, pod, namespace, secret, deployment, and service account mean.
3. How your Slack notifier project can run inside Kubernetes.
4. How Tekton uses Kubernetes underneath.
5. How to inspect resources locally using Minikube and `kubectl`.
6. Why Kubernetes basics are required before Tekton makes sense.
7. How to write a simple Go config loader that reads environment variables with defaults.
8. Binary search basics with one Go DSA problem.

Kubernetes is an open-source system for automating deployment, scaling, and management of containerized applications. ([Kubernetes][1])

---

## 2. Quick revision of Days 1 to 7

### Day 1 — Go CLI basics

You learned:

```text
User input -> CLI flags -> Go program -> build event object
```

Example:

```bash
go run cmd/slack-notifier/main.go \
  --event-type=pr \
  --status=failed \
  --pipeline-name=pr-validation
```

In Python, this is similar to using `argparse`.

In Go, you commonly use:

```go
flag.String(...)
flag.Parse()
```

---

### Day 2 — Structs and event model

You learned that Go structs are like Python `dataclass` or simple classes.

Python:

```python
@dataclass
class PipelineEvent:
    event_type: str
    status: str
```

Go:

```go
type PipelineEvent struct {
    EventType string
    Status    string
}
```

In your project:

```text
CLI flags -> PipelineEvent struct -> Validate()
```

---

### Day 3 — JSON, HTTP, Slack webhook

You learned:

```text
Go struct -> JSON -> HTTP POST -> Slack webhook
```

Python equivalent:

```python
requests.post(webhook_url, json=payload)
```

Go equivalent:

```go
http.Post(webhookURL, "application/json", body)
```

---

### Day 4 — Router logic

You learned that routing chooses the correct Slack webhook.

Example:

```text
event_type = pr  -> PR webhook
event_type = cd  -> CD webhook
event_type = job -> Job webhook or fallback webhook
```

This keeps `main.go` small.

---

### Day 5 — Error handling and logging

You learned that Go uses explicit error checks:

```go
if err != nil {
    return err
}
```

Python usually uses exceptions:

```python
try:
    do_work()
except Exception as e:
    print(e)
```

Go style is more explicit.

---

### Day 6 — Testing

You learned:

```text
_test.go files
table-driven tests
httptest mock server
```

This helps you test Slack webhook logic without calling real Slack.

---

### Day 7 — Shell scripting

You learned shell scripts for local workflow:

```bash
#!/usr/bin/env bash
set -euo pipefail
go run cmd/slack-notifier/main.go ...
```

This helps wrap long Go commands into reusable scripts.

---

## 3. Explain Kubernetes in very simple language

Think of Kubernetes like a **smart manager for running applications**.

Without Kubernetes:

```text
You manually start app
You manually restart app if it crashes
You manually set env vars
You manually check logs
You manually scale app
```

With Kubernetes:

```text
You describe what you want
Kubernetes tries to keep it running
```

Simple analogy:

```text
You = restaurant owner
Kubernetes = restaurant manager
Pods = workers
Deployment = staffing plan
Secrets = locked drawer with passwords
Namespace = separate room/department
ServiceAccount = worker ID card
```

You do not usually say:

```text
Start container manually now.
```

Instead, you say:

```text
I want 1 copy of this app running.
Use this image.
Use these env vars.
Use this secret.
Put it in this namespace.
```

Then Kubernetes tries to make reality match your desired state.

---

## 4. Explain cluster, node, pod, namespace

### 4.1 Cluster

A **cluster** is the full Kubernetes environment.

Simple analogy:

```text
Cluster = entire office building
```

It contains machines where your applications run.

In Minikube, your local laptop creates a small local cluster for learning and testing.

```bash
minikube start
kubectl get nodes
```

---

### 4.2 Node

A **node** is a machine inside the cluster.

Simple analogy:

```text
Node = one computer/server inside the office building
```

In real production, you may have many nodes:

```text
node-1
node-2
node-3
```

In Minikube, usually you start with one node.

---

### 4.3 Pod

A **pod** is the smallest runnable unit in Kubernetes.

Simple analogy:

```text
Pod = small room where one application container runs
```

For your project, one pod can run your Slack notifier container.

Example:

```text
Pod
 └── Container
      └── slack-notifier binary
```

A pod can contain one or more containers, but as a beginner, think:

```text
1 pod usually runs 1 main app container
```

Kubernetes documentation describes Pods as workload resources used to run containers in the cluster. ([Kubernetes][2])

---

### 4.4 Deployment

A **Deployment** manages pods.

Simple analogy:

```text
Pod = one worker
Deployment = manager who keeps required workers running
```

Suppose you say:

```text
I want 1 Slack notifier pod running.
```

The Deployment makes sure one pod exists.

If the pod crashes:

```text
Deployment creates a new pod
```

Kubernetes documentation describes a Deployment as a controller that manages Pods and ReplicaSets and moves actual state toward the desired state. ([Kubernetes][3])

---

## 5. Pod vs Deployment

| Concept         | Simple meaning                       | Real project meaning                                |
| --------------- | ------------------------------------ | --------------------------------------------------- |
| Pod             | Running app instance                 | One running Slack notifier container                |
| Deployment      | Keeps pods running                   | Ensures Slack notifier pod is recreated if it fails |
| Pod name        | Often changes                        | New pod gets a new generated name                   |
| Deployment name | Stable                               | You usually interact with Deployment name           |
| Beginner rule   | Do not directly manage pods for apps | Use Deployment for long-running apps                |

### Simple example

Bad for real app:

```yaml
kind: Pod
```

Better for app:

```yaml
kind: Deployment
```

Why?

Because a Deployment gives Kubernetes a desired state:

```text
Always keep 1 replica running.
```

---

## 6. Explain namespace clearly

A **namespace** is a logical separation inside a cluster.

Simple analogy:

```text
Cluster = office building
Namespace = separate department
```

Example namespaces:

```text
slack-integration-dev
slack-integration-stage
slack-integration-prod
```

Why namespaces help:

```text
dev resources stay separate from prod resources
different secrets can exist in each namespace
same app name can exist in different namespaces
kubectl commands become safer
```

Example:

```bash
kubectl create namespace slack-integration-dev
kubectl get pods -n slack-integration-dev
```

Without namespace awareness, beginners often look in the wrong place.

Common mistake:

```bash
kubectl get pods
```

But your pod is actually in:

```bash
kubectl get pods -n slack-integration-dev
```

---

## 7. Explain secrets and service accounts clearly

### 7.1 Secret

A **Secret** stores sensitive data.

For your project, the Slack webhook URL should not be hardcoded in code or YAML.

Bad:

```go
webhookURL := "https://hooks.slack.com/services/..."
```

Better:

```text
Store webhook URL in Kubernetes Secret
Read it as environment variable inside the pod
```

Kubernetes documentation describes a Secret as an object for storing a small amount of sensitive data such as a password, token, or key, avoiding putting confidential data directly in code or container images. ([Kubernetes][4])

---

### 7.2 Secret vs ConfigMap

| Concept   | Used for             | Example                                  |
| --------- | -------------------- | ---------------------------------------- |
| Secret    | Sensitive values     | Slack webhook URL, API token             |
| ConfigMap | Non-sensitive config | retry count, log level, environment name |

Simple rule:

```text
Password/token/webhook URL -> Secret
Normal setting -> ConfigMap
```

For your Slack project:

```text
SLACK_WEBHOOK_URL       -> Secret
APP_ENV=dev             -> ConfigMap or plain env
RETRY_COUNT=3           -> ConfigMap or plain env
LOG_LEVEL=debug         -> ConfigMap or plain env
```

---

### 7.3 ServiceAccount

A **ServiceAccount** is an identity for a pod or controller.

Simple analogy:

```text
ServiceAccount = ID card for the app inside Kubernetes
```

For example:

```text
Tekton task pod uses a ServiceAccount
That ServiceAccount decides what the pod is allowed to do
```

Kubernetes documentation describes a ServiceAccount as a non-human account that provides a distinct identity in a Kubernetes cluster, and pods can use ServiceAccount credentials to authenticate to the API server. ([Kubernetes][5])

Beginner mental model:

```text
Human uses kubectl with user identity
Pod uses ServiceAccount identity
```

---

## 8. Explain how Tekton runs on top of Kubernetes

Tekton is not separate from Kubernetes.

Tekton is installed **inside Kubernetes** and adds new Kubernetes-style resources such as:

```text
Task
TaskRun
Pipeline
PipelineRun
TriggerBinding
TriggerTemplate
EventListener
```

Tekton Pipelines is a Kubernetes extension that installs and runs on a Kubernetes cluster, and after installation its resources can be used through `kubectl` and API calls like other Kubernetes resources. ([Tekton][6])

Important beginner point:

```text
Tekton Task does not run magically.
Tekton Task becomes a Kubernetes pod.
```

Tekton documentation says a Task executes as a Pod on the Kubernetes cluster. ([Tekton][7])

So when your Tekton pipeline runs this:

```yaml
- name: notify-slack
  taskRef:
    name: task-slack-notify
```

Under the hood:

```text
Tekton controller creates a pod
pod runs container
container executes script/command
logs are available via kubectl/tkn
```

---

## 9. Show how this project uses Kubernetes resources

Your `slack-integration` style project can use Kubernetes like this:

```text
Namespace
 ├── Secret
 │    └── SLACK_WEBHOOK_URL
 │
 ├── ConfigMap / env config
 │    └── APP_ENV, RETRY_COUNT, LOG_LEVEL
 │
 ├── ServiceAccount
 │    └── slack-notifier-sa
 │
 ├── Deployment
 │    └── slack-notifier app pod
 │
 └── Tekton resources
      ├── Task
      ├── Pipeline
      ├── PipelineRun
      └── EventListener
```

Two possible project styles:

### Style 1 — App runs as a normal Deployment

Use this when Slack notifier is a long-running service.

```text
Deployment -> Pod -> Slack notifier HTTP service
```

### Style 2 — App runs inside Tekton Task

Use this when Slack notifier is a CLI called during CI/CD.

```text
PipelineRun -> TaskRun -> Pod -> go run / binary -> Slack message
```

For your current learning project, Style 2 is very relevant because your notifier behaves like a CLI.

---

## 10. ASCII diagram for Kubernetes resource flow

```text
Developer
   |
   | kubectl apply -f k8s/
   v
Kubernetes API Server
   |
   | stores desired state
   v
Controllers
   |
   | create/update resources
   v
Deployment Controller
   |
   | creates ReplicaSet
   v
ReplicaSet
   |
   | creates Pod
   v
Pod
   |
   | starts container
   v
slack-notifier container
   |
   | reads env vars
   | reads Secret value
   | sends HTTP POST
   v
Slack Webhook
   |
   v
Slack Channel
```

Tekton version:

```text
GitHub event / manual trigger
   |
   v
Tekton EventListener / PipelineRun
   |
   v
Tekton Controller
   |
   v
TaskRun
   |
   v
Kubernetes Pod
   |
   v
Container step executes Slack notifier command
   |
   v
Slack message
```

---

## 11. Pseudocode for `kubectl apply -> controller -> pod -> logs`

```text
START

User runs:
    kubectl apply -f deployment.yaml

kubectl sends YAML to Kubernetes API server

API server stores desired state:
    "I want Deployment slack-notifier with 1 replica"

Deployment controller watches desired state

IF Deployment does not have required pod:
    create ReplicaSet

ReplicaSet checks desired replicas

IF current pods < desired pods:
    create new Pod

Scheduler chooses a node for the Pod

Kubelet on that node starts the container

Container runs slack-notifier app

User checks:
    kubectl get pods
    kubectl logs pod-name

END
```

Python analogy:

```text
YAML is like a config dictionary
Kubernetes controller is like a background worker loop
It constantly checks:
    desired_state == actual_state?
If not:
    fix actual_state
```

---

## 12. Real YAML examples for beginner understanding

### 12.1 Namespace YAML

File:

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: slack-integration-dev
```

Apply:

```bash
kubectl apply -f k8s/namespace.yaml
```

Check:

```bash
kubectl get namespaces
```

Expected:

```text
slack-integration-dev   Active
```

---

### 12.2 Secret YAML for Slack webhook

For learning only, use a placeholder.

```yaml
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: slack-webhook-secret
  namespace: slack-integration-dev
type: Opaque
stringData:
  SLACK_WEBHOOK_URL: "https://hooks.slack.com/services/REPLACE/ME"
```

Important:

```text
stringData lets you write plain text in YAML.
Kubernetes stores it as secret data.
Do not commit real webhook URLs to Git.
```

Apply:

```bash
kubectl apply -f k8s/secret.yaml
```

Check:

```bash
kubectl get secrets -n slack-integration-dev
```

Expected:

```text
slack-webhook-secret   Opaque
```

---

### 12.3 ConfigMap YAML for non-sensitive config

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: slack-notifier-config
  namespace: slack-integration-dev
data:
  APP_ENV: "dev"
  LOG_LEVEL: "debug"
  RETRY_COUNT: "3"
```

Use this for normal settings.

Apply:

```bash
kubectl apply -f k8s/configmap.yaml
```

Check:

```bash
kubectl get configmap -n slack-integration-dev
```

---

### 12.4 ServiceAccount YAML

```yaml
# k8s/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: slack-notifier-sa
  namespace: slack-integration-dev
```

Apply:

```bash
kubectl apply -f k8s/serviceaccount.yaml
```

Check:

```bash
kubectl get serviceaccounts -n slack-integration-dev
```

---

### 12.5 Deployment YAML

This example assumes you have built and pushed an image named:

```text
slack-notifier:dev
```

For Minikube learning, you may later build the image inside Minikube.

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: slack-notifier
  namespace: slack-integration-dev
spec:
  replicas: 1
  selector:
    matchLabels:
      app: slack-notifier
  template:
    metadata:
      labels:
        app: slack-notifier
    spec:
      serviceAccountName: slack-notifier-sa
      containers:
        - name: slack-notifier
          image: slack-notifier:dev
          imagePullPolicy: IfNotPresent
          env:
            - name: APP_ENV
              valueFrom:
                configMapKeyRef:
                  name: slack-notifier-config
                  key: APP_ENV

            - name: LOG_LEVEL
              valueFrom:
                configMapKeyRef:
                  name: slack-notifier-config
                  key: LOG_LEVEL

            - name: RETRY_COUNT
              valueFrom:
                configMapKeyRef:
                  name: slack-notifier-config
                  key: RETRY_COUNT

            - name: SLACK_WEBHOOK_URL
              valueFrom:
                secretKeyRef:
                  name: slack-webhook-secret
                  key: SLACK_WEBHOOK_URL
```

Apply:

```bash
kubectl apply -f k8s/deployment.yaml
```

Check:

```bash
kubectl get deployments -n slack-integration-dev
kubectl get pods -n slack-integration-dev
```

Expected:

```text
deployment.apps/slack-notifier created
```

Then:

```text
slack-notifier-xxxxx   1/1   Running
```

---

## 13. Important YAML line-by-line explanation

### Deployment important parts

```yaml
apiVersion: apps/v1
```

Means:

```text
Use Kubernetes apps API group.
```

---

```yaml
kind: Deployment
```

Means:

```text
This YAML creates a Deployment.
```

---

```yaml
metadata:
  name: slack-notifier
  namespace: slack-integration-dev
```

Means:

```text
Name this resource slack-notifier.
Create it inside slack-integration-dev namespace.
```

---

```yaml
replicas: 1
```

Means:

```text
Keep 1 pod running.
```

---

```yaml
selector:
  matchLabels:
    app: slack-notifier
```

Means:

```text
This Deployment manages pods with label app=slack-notifier.
```

---

```yaml
template:
```

Means:

```text
This is the pod template.
Whenever Kubernetes creates a pod, it uses this section.
```

---

```yaml
serviceAccountName: slack-notifier-sa
```

Means:

```text
Run this pod using slack-notifier-sa identity.
```

---

```yaml
env:
  - name: SLACK_WEBHOOK_URL
    valueFrom:
      secretKeyRef:
```

Means:

```text
Inject secret value as environment variable.
```

Go app sees it like this:

```go
os.Getenv("SLACK_WEBHOOK_URL")
```

Python equivalent:

```python
os.environ.get("SLACK_WEBHOOK_URL")
```

---

## 14. How to inspect resources locally in Minikube

Start Minikube:

```bash
minikube start
```

Check cluster node:

```bash
kubectl get nodes
```

Check namespaces:

```bash
kubectl get namespaces
```

Check all resources in your namespace:

```bash
kubectl get all -n slack-integration-dev
```

Check pods:

```bash
kubectl get pods -n slack-integration-dev
```

Describe pod:

```bash
kubectl describe pod <pod-name> -n slack-integration-dev
```

Check logs:

```bash
kubectl logs <pod-name> -n slack-integration-dev
```

Watch resources:

```bash
kubectl get pods -n slack-integration-dev -w
```

Minikube documentation notes that after `minikube start`, `kubectl` gets configured to access the Minikube Kubernetes control plane; if local `kubectl` is not installed, `minikube kubectl --` can be used. ([minikube][8])

For troubleshooting, beginner-friendly commands include `kubectl get`, `kubectl describe`, `kubectl logs`, and `kubectl exec`. ([minikube][9])

---

## 15. Hands-on tasks

### Task 1 — Start Minikube

```bash
minikube start
kubectl get nodes
```

Expected:

```text
NAME       STATUS   ROLES
minikube   Ready    control-plane
```

---

### Task 2 — Create namespace

```bash
kubectl create namespace slack-integration-dev
kubectl get namespaces
```

Expected:

```text
slack-integration-dev   Active
```

---

### Task 3 — Create Secret manually

```bash
kubectl create secret generic slack-webhook-secret \
  --from-literal=SLACK_WEBHOOK_URL="https://hooks.slack.com/services/REPLACE/ME" \
  -n slack-integration-dev
```

Check:

```bash
kubectl get secrets -n slack-integration-dev
```

---

### Task 4 — Create ConfigMap manually

```bash
kubectl create configmap slack-notifier-config \
  --from-literal=APP_ENV=dev \
  --from-literal=LOG_LEVEL=debug \
  --from-literal=RETRY_COUNT=3 \
  -n slack-integration-dev
```

Check:

```bash
kubectl get configmap -n slack-integration-dev
```

---

### Task 5 — Inspect everything

```bash
kubectl get all -n slack-integration-dev
kubectl get secrets -n slack-integration-dev
kubectl get configmaps -n slack-integration-dev
kubectl get serviceaccounts -n slack-integration-dev
```

---

## 16. Expected output

After creating namespace, secret, configmap, and service account, you should see:

```text
namespace/slack-integration-dev created
secret/slack-webhook-secret created
configmap/slack-notifier-config created
serviceaccount/slack-notifier-sa created
```

When checking resources:

```bash
kubectl get all -n slack-integration-dev
```

You may see no pods yet if you have not applied Deployment.

That is okay.

Expected before Deployment:

```text
No resources found in slack-integration-dev namespace.
```

Expected after Deployment:

```text
NAME                                  READY   STATUS    RESTARTS   AGE
pod/slack-notifier-xxxxx              1/1     Running   0          30s

NAME                             READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/slack-notifier   1/1     1            1           30s
```

---

## 17. Common mistakes

### Mistake 1 — Looking in the wrong namespace

Wrong:

```bash
kubectl get pods
```

Correct:

```bash
kubectl get pods -n slack-integration-dev
```

---

### Mistake 2 — Hardcoding secrets

Bad:

```go
webhookURL := "https://hooks.slack.com/services/..."
```

Good:

```go
webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
```

---

### Mistake 3 — Confusing Pod and Deployment

Wrong thinking:

```text
I created one pod, so Kubernetes will always restart it.
```

Better thinking:

```text
Deployment is responsible for keeping pods running.
```

---

### Mistake 4 — Forgetting labels/selectors

Deployment selector:

```yaml
selector:
  matchLabels:
    app: slack-notifier
```

Pod template label must match:

```yaml
labels:
  app: slack-notifier
```

If labels do not match, Deployment cannot manage the pod correctly.

---

### Mistake 5 — Image not found

You may see:

```text
ImagePullBackOff
```

Meaning:

```text
Kubernetes cannot pull or find your container image.
```

Debug:

```bash
kubectl describe pod <pod-name> -n slack-integration-dev
```

---

### Mistake 6 — Secret key name mismatch

YAML expects:

```yaml
key: SLACK_WEBHOOK_URL
```

But Secret has:

```text
SLACK_URL
```

Then pod may fail to start or env var may be missing.

---

## 18. Debugging tips using `kubectl`

### Check current context

```bash
kubectl config current-context
```

Expected for Minikube:

```text
minikube
```

---

### Check resources

```bash
kubectl get all -n slack-integration-dev
```

---

### Describe deployment

```bash
kubectl describe deployment slack-notifier -n slack-integration-dev
```

Use this when:

```text
pod is not created
deployment is not available
replica count is wrong
```

---

### Describe pod

```bash
kubectl describe pod <pod-name> -n slack-integration-dev
```

Use this when:

```text
pod is Pending
pod is CrashLoopBackOff
pod is ImagePullBackOff
env var is missing
secret is missing
```

---

### Check logs

```bash
kubectl logs <pod-name> -n slack-integration-dev
```

Use this when:

```text
app starts but fails internally
Slack message is not sent
config loader reports missing env
```

---

### Enter pod shell

```bash
kubectl exec -it <pod-name> -n slack-integration-dev -- sh
```

Then check env vars:

```bash
env | grep SLACK
env | grep APP_ENV
```

---

### Delete and recreate

```bash
kubectl delete deployment slack-notifier -n slack-integration-dev
kubectl apply -f k8s/deployment.yaml
```

---

## 19. Why this foundation is needed before Tekton makes sense

Tekton uses Kubernetes resources.

So before understanding this:

```text
TaskRun -> Pod -> Step container -> Logs
```

You must understand:

```text
Pod
Namespace
Secret
ServiceAccount
Logs
kubectl describe
```

In your project:

```text
Tekton PipelineRun
   |
   v
TaskRun
   |
   v
Pod
   |
   v
Container step
   |
   v
Go Slack notifier CLI
   |
   v
Slack webhook
```

When Tekton fails, you debug it using Kubernetes thinking:

```text
Is the TaskRun created?
Is the Pod created?
Is the Pod running?
Are Secrets mounted?
Is ServiceAccount correct?
What do pod logs say?
```

That is why Kubernetes comes before Tekton.

---

# 20. Go + Python comparison for Kubernetes config

Your module-based practice is:

```text
Create a config loader that reads env vars with defaults.
```

In Python, you might write:

```python
import os

app_env = os.environ.get("APP_ENV", "dev")
retry_count = int(os.environ.get("RETRY_COUNT", "3"))
```

In Go, you write:

```go
appEnv := getEnv("APP_ENV", "dev")
retryCount := getEnvAsInt("RETRY_COUNT", 3)
```

Main differences:

| Topic           | Python             | Go                    |
| --------------- | ------------------ | --------------------- |
| Env read        | `os.environ.get()` | `os.Getenv()`         |
| Type conversion | `int(value)`       | `strconv.Atoi(value)` |
| Error handling  | `try/except`       | `value, err := ...`   |
| Data object     | dict/dataclass     | struct                |
| Convention      | flexible           | explicit and typed    |

---

## 21. Module-based practice task: Config loader with defaults

### Goal

Create this package:

```text
pkg/config/config.go
```

It should read:

```text
APP_ENV
LOG_LEVEL
RETRY_COUNT
SLACK_WEBHOOK_URL
```

With defaults:

```text
APP_ENV=dev
LOG_LEVEL=info
RETRY_COUNT=3
SLACK_WEBHOOK_URL=no default because it is required
```

---

## 22. Pseudocode for config loader

```text
START

Create Config struct:
    AppEnv
    LogLevel
    RetryCount
    SlackWebhookURL

Create getEnv function:
    input key and default value
    read env var
    if empty:
        return default value
    return actual value

Create getEnvAsInt function:
    read env var as string
    if empty:
        return default int
    convert string to int
    if conversion fails:
        return error
    return int value

Create Load function:
    read APP_ENV with default dev
    read LOG_LEVEL with default info
    read RETRY_COUNT with default 3
    read SLACK_WEBHOOK_URL

    if SLACK_WEBHOOK_URL is empty:
        return error

    return Config

END
```

---

## 23. Real Go code: `pkg/config/config.go`

```go
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds application configuration loaded from environment variables.
//
// Python comparison:
// In Python, this might be a dataclass.
// In Go, we use a struct for typed configuration.
type Config struct {
	AppEnv          string
	LogLevel        string
	RetryCount      int
	SlackWebhookURL string
}

// Load reads environment variables and returns application config.
func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:          getEnv("APP_ENV", "dev"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		SlackWebhookURL: getEnv("SLACK_WEBHOOK_URL", ""),
	}

	retryCount, err := getEnvAsInt("RETRY_COUNT", 3)
	if err != nil {
		return nil, err
	}

	cfg.RetryCount = retryCount

	if cfg.SlackWebhookURL == "" {
		return nil, fmt.Errorf("SLACK_WEBHOOK_URL is required")
	}

	return cfg, nil
}

// getEnv returns the environment variable value.
// If the value is empty, it returns the default value.
//
// Python equivalent:
// os.environ.get("APP_ENV", "dev")
func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt reads an environment variable and converts it to int.
//
// Python equivalent:
// int(os.environ.get("RETRY_COUNT", "3"))
func getEnvAsInt(key string, defaultValue int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue, nil
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", key, err)
	}

	return intValue, nil
}
```

---

## 24. Example usage in `main.go`

```go
package main

import (
	"fmt"
	"log"

	"slack-integration/pkg/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	fmt.Println("App Env:", cfg.AppEnv)
	fmt.Println("Log Level:", cfg.LogLevel)
	fmt.Println("Retry Count:", cfg.RetryCount)
	fmt.Println("Slack Webhook URL loaded successfully")
}
```

---

## 25. How to run locally

Without webhook:

```bash
go run cmd/slack-notifier/main.go
```

Expected:

```text
failed to load config: SLACK_WEBHOOK_URL is required
```

With env vars:

```bash
export APP_ENV=dev
export LOG_LEVEL=debug
export RETRY_COUNT=3
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/REPLACE/ME"

go run cmd/slack-notifier/main.go
```

Expected:

```text
App Env: dev
Log Level: debug
Retry Count: 3
Slack Webhook URL loaded successfully
```

Invalid retry count:

```bash
export RETRY_COUNT=abc
go run cmd/slack-notifier/main.go
```

Expected:

```text
failed to load config: RETRY_COUNT must be a valid integer
```

---

# 26. DSA topic: Binary Search

## Simple explanation

Binary search is used when data is **sorted**.

Instead of checking every item one by one, binary search checks the middle item.

Analogy:

```text
You are searching for page 700 in a 1000-page book.

You do not start from page 1.
You open near the middle.
If page is too small, search right side.
If page is too large, search left side.
Repeat.
```

Important rule:

```text
Binary search only works correctly on sorted data.
```

---

## Binary search flow

Array:

```text
[2, 4, 6, 8, 10, 12, 14]
```

Target:

```text
10
```

Steps:

```text
left = 0
right = 6

middle = 3
nums[3] = 8

8 < 10
So search right side

left = 4
right = 6

middle = 5
nums[5] = 12

12 > 10
So search left side

left = 4
right = 4

middle = 4
nums[4] = 10

Found
```

---

## Python version

```python
def binary_search(nums, target):
    left = 0
    right = len(nums) - 1

    while left <= right:
        mid = (left + right) // 2

        if nums[mid] == target:
            return mid
        elif nums[mid] < target:
            left = mid + 1
        else:
            right = mid - 1

    return -1
```

---

## Go version

```go
package main

import "fmt"

func binarySearch(nums []int, target int) int {
	left := 0
	right := len(nums) - 1

	for left <= right {
		mid := left + (right-left)/2

		if nums[mid] == target {
			return mid
		}

		if nums[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return -1
}

func main() {
	nums := []int{2, 4, 6, 8, 10, 12, 14}
	target := 10

	index := binarySearch(nums, target)

	fmt.Println("Index:", index)
}
```

Expected output:

```text
Index: 4
```

---

## Go vs Python syntax difference

| Concept          | Python                   | Go                           |
| ---------------- | ------------------------ | ---------------------------- |
| List/array       | `nums = [1, 2, 3]`       | `nums := []int{1, 2, 3}`     |
| Function         | `def binary_search(...)` | `func binarySearch(...) int` |
| While loop       | `while left <= right:`   | `for left <= right {}`       |
| Integer division | `//`                     | `/` for ints                 |
| Return not found | `return -1`              | `return -1`                  |

Go does not have `while`.

Go uses `for` for all loops.

---

# 27. Day 8 Go DSA problem

## Problem: Search Insert Position

Given a sorted array and a target, return the index if the target is found.

If not found, return the index where it should be inserted.

Example:

```text
nums = [1, 3, 5, 6]
target = 5
output = 2
```

Example:

```text
nums = [1, 3, 5, 6]
target = 2
output = 1
```

Example:

```text
nums = [1, 3, 5, 6]
target = 7
output = 4
```

---

## Go solution

```go
package main

import "fmt"

func searchInsert(nums []int, target int) int {
	left := 0
	right := len(nums) - 1

	for left <= right {
		mid := left + (right-left)/2

		if nums[mid] == target {
			return mid
		}

		if nums[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return left
}

func main() {
	fmt.Println(searchInsert([]int{1, 3, 5, 6}, 5)) // 2
	fmt.Println(searchInsert([]int{1, 3, 5, 6}, 2)) // 1
	fmt.Println(searchInsert([]int{1, 3, 5, 6}, 7)) // 4
	fmt.Println(searchInsert([]int{1, 3, 5, 6}, 0)) // 0
}
```

Expected output:

```text
2
1
4
0
```

---

## Why return `left`?

When the loop ends:

```text
left points to the correct insert position
```

Example:

```text
nums = [1, 3, 5, 6]
target = 2
```

At the end:

```text
left = 1
```

So `2` should be inserted at index `1`.

---

# 28. Revision checkpoint

You should now be able to answer these:

1. What is Kubernetes?
2. What is a cluster?
3. What is a node?
4. What is a pod?
5. What is a Deployment?
6. Why is Deployment better than directly creating a pod?
7. What is a namespace?
8. Why should dev/stage/prod use separate namespaces?
9. What should go into a Secret?
10. What should go into a ConfigMap?
11. What is a ServiceAccount?
12. How does Tekton run on top of Kubernetes?
13. What Kubernetes object does a Tekton Task usually create underneath?
14. How do you check pods in a namespace?
15. How do you check pod logs?
16. How do you debug `ImagePullBackOff`?
17. How do you debug missing environment variables?
18. Why does binary search need sorted data?
19. What is the time complexity of binary search?
20. How is `os.Getenv` in Go similar to `os.environ.get` in Python?

---

# 29. Homework

## Part 1 — Kubernetes practice

Create these files:

```text
k8s/namespace.yaml
k8s/secret.yaml
k8s/configmap.yaml
k8s/serviceaccount.yaml
k8s/deployment.yaml
```

Apply them:

```bash
kubectl apply -f k8s/
```

Check:

```bash
kubectl get all -n slack-integration-dev
kubectl get secrets -n slack-integration-dev
kubectl get configmaps -n slack-integration-dev
```

---

## Part 2 — Debugging practice

Run:

```bash
kubectl describe deployment slack-notifier -n slack-integration-dev
kubectl get pods -n slack-integration-dev
kubectl describe pod <pod-name> -n slack-integration-dev
kubectl logs <pod-name> -n slack-integration-dev
```

Write down:

```text
Pod name:
Pod status:
Image name:
Environment variables:
Any error:
```

---

## Part 3 — Go config loader

Create:

```text
pkg/config/config.go
```

Implement:

```text
Load()
getEnv()
getEnvAsInt()
```

Test with:

```bash
export APP_ENV=dev
export LOG_LEVEL=debug
export RETRY_COUNT=3
export SLACK_WEBHOOK_URL="dummy"

go run cmd/slack-notifier/main.go
```

---

## Part 4 — DSA

Solve:

```text
Search Insert Position
```

Then test with:

```text
[1,3,5,6], target 5 -> 2
[1,3,5,6], target 2 -> 1
[1,3,5,6], target 7 -> 4
[1,3,5,6], target 0 -> 0
```

---

# 30. Day 8 final mental model

Remember this:

```text
Kubernetes = runs and manages containers
Pod = smallest running app unit
Deployment = keeps pods running
Namespace = separate project/environment area
Secret = sensitive config
ConfigMap = normal config
ServiceAccount = identity for pod
Tekton = CI/CD system built on Kubernetes
Tekton Task = creates Kubernetes pod
kubectl = your inspection/debugging tool
```

For your project:

```text
Slack notifier code
   |
   v
Container image
   |
   v
Kubernetes pod
   |
   v
Secret + config env vars
   |
   v
Tekton task or Deployment
   |
   v
Slack notification
```

[1]: https://kubernetes.io/docs/home/?utm_source=chatgpt.com "Kubernetes Documentation"
[2]: https://kubernetes.io/docs/concepts/workloads/pods/?utm_source=chatgpt.com "Pods"
[3]: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/?utm_source=chatgpt.com "Deployments"
[4]: https://kubernetes.io/docs/concepts/configuration/secret/?utm_source=chatgpt.com "Secrets"
[5]: https://kubernetes.io/docs/concepts/security/service-accounts/?utm_source=chatgpt.com "Service Accounts"
[6]: https://tekton.dev/docs/pipelines/?utm_source=chatgpt.com "Tasks and Pipelines - Tekton"
[7]: https://tekton.dev/docs/pipelines/tasks/?utm_source=chatgpt.com "Tasks - Tekton"
[8]: https://minikube.sigs.k8s.io/docs/handbook/kubectl/?utm_source=chatgpt.com "Kubectl - Minikube - Kubernetes"
[9]: https://minikube.sigs.k8s.io/docs/tutorials/kubernetes_101/module3/?utm_source=chatgpt.com "Module 3 - Explore your app - Minikube - Kubernetes"
