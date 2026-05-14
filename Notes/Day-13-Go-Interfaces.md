# Day 13 — Interfaces, Dependency Injection, Config, and Cleaner Architecture

Today’s goal is not to make the project “fancy.”

Today’s goal is to make your Slack notifier:

* easier to change
* easier to test
* easier to understand
* less dependent on hardcoded details
* safer to refactor step by step

We will keep everything beginner-friendly.

---

# 1. Day 13 learning goals

By the end of Day 13, you should understand:

1. What an interface is in Go.
2. How an interface is different from a struct.
3. Why interfaces help with testing.
4. What dependency injection means in simple terms.
5. How to introduce a `Sender` interface for Slack messages.
6. How to move environment/config logic into a `config` package.
7. How to keep `main.go` small and readable.
8. How to refactor without breaking the project.
9. How prefix sum works in DSA.
10. How to build one small config-driven module.

Python comparison:

In Python, you often rely on “duck typing.”

```python
def notify(sender):
    sender.send("hello")
```

You do not always define the interface formally. In Go, we often make that expectation explicit:

```go
type Sender interface {
    Send(message string) error
}
```

Go wants the shape of the dependency to be clear.

---

# 2. Quick revision of Days 1 to 12

You have already been moving through core backend project ideas.

A typical journey so far:

## Day 1–3: Go basics

You learned:

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello")
}
```

Python comparison:

```python
print("Hello")
```

Go requires:

* package declaration
* explicit imports
* `func main()` as entry point
* braces `{}`

---

## Day 4–5: Variables, functions, structs

Go struct:

```go
type User struct {
    Name string
    Age  int
}
```

Python equivalent:

```python
class User:
    def __init__(self, name, age):
        self.name = name
        self.age = age
```

Or closer:

```python
from dataclasses import dataclass

@dataclass
class User:
    name: str
    age: int
```

Go convention:

* struct names use `PascalCase`
* field names starting with capital letters are exported
* field names starting with lowercase letters are package-private

---

## Day 6–8: HTTP, JSON, Slack webhook style

You probably worked with things like:

```go
http.Post(url, "application/json", body)
```

Python equivalent:

```python
requests.post(url, json=payload)
```

Go is more explicit. You usually handle:

* JSON encoding
* errors
* status codes
* request context
* HTTP client

---

## Day 9–10: Packages

Instead of putting everything in one file, you split code:

```text
slack-integration/
  main.go
  notifier/
  slack/
  config/
```

Python equivalent:

```text
slack_integration/
  main.py
  notifier/
  slack/
  config/
```

Go packages are simpler than Python modules in some ways, but stricter about names and visibility.

---

## Day 11–12: Testing and error handling

Go error handling:

```go
if err != nil {
    return err
}
```

Python equivalent:

```python
try:
    do_something()
except Exception as e:
    raise e
```

Go tests usually live beside the code:

```text
notifier/
  service.go
  service_test.go
```

Today we make testing easier by using interfaces.

---

# 3. Explain interfaces in very simple language

An interface in Go says:

> “I do not care what exact type you are. I only care that you have these methods.”

Example:

```go
type Sender interface {
    Send(message string) error
}
```

This means:

> Anything that has a method called `Send` which accepts a `string` and returns an `error` can be used as a `Sender`.

For example:

```go
type SlackSender struct{}

func (s SlackSender) Send(message string) error {
    fmt.Println("Sending to Slack:", message)
    return nil
}
```

`SlackSender` automatically satisfies the `Sender` interface.

There is no need to write:

```go
implements Sender
```

Go does not use `implements`.

This is different from languages like Java.

Python comparison:

Python often works like this:

```python
class SlackSender:
    def send(self, message):
        print("Sending to Slack:", message)

def notify(sender):
    sender.send("Build failed")
```

Python says:

> “If it has a `send` method, I will try to use it.”

Go says:

> “Let us clearly describe the expected method set using an interface.”

So Go interfaces are like a formal version of duck typing.

---

# 4. Interface vs struct clearly

## Struct

A struct is real data.

```go
type SlackClient struct {
    WebhookURL string
}
```

This creates a concrete thing with fields.

Python equivalent:

```python
class SlackClient:
    def __init__(self, webhook_url):
        self.webhook_url = webhook_url
```

Use a struct when you want to store actual data.

---

## Interface

An interface is a behavior contract.

```go
type Sender interface {
    Send(message string) error
}
```

It does not store data.

It only says:

> “This is what something must be able to do.”

Python equivalent:

```python
from typing import Protocol

class Sender(Protocol):
    def send(self, message: str) -> None:
        ...
```

---

## Comparison

| Concept             | Go struct          | Go interface             |
| ------------------- | ------------------ | ------------------------ |
| Meaning             | Concrete data type | Behavior contract        |
| Stores fields?      | Yes                | No                       |
| Has implementation? | Yes                | No                       |
| Used for            | Real objects       | Flexible dependencies    |
| Example             | `SlackClient`      | `Sender`                 |
| Python equivalent   | class/dataclass    | Protocol/ABC/duck typing |

---

## Simple rule

Use a **struct** when you know what the thing is.

Use an **interface** when you care what the thing can do.

Example:

```go
type SlackClient struct {
    webhookURL string
}
```

This is a real Slack client.

```go
type Sender interface {
    Send(message string) error
}
```

This means anything that can send messages.

---

## When interfaces are useful

Interfaces are useful when:

1. You want to test code without calling real Slack.
2. You want to swap implementations later.
3. You want your business logic to depend on behavior, not concrete details.
4. You want cleaner architecture.

Example:

* Real app uses `SlackClient`
* Test uses `FakeSender`

Both can satisfy:

```go
type Sender interface {
    Send(message string) error
}
```

---

## When interfaces are unnecessary

Do not create interfaces everywhere.

Avoid this:

```go
type ConfigInterface interface {
    GetSlackURL() string
}
```

When a simple struct is enough:

```go
type Config struct {
    SlackWebhookURL string
}
```

Beginner rule:

> Start with structs. Introduce interfaces only when they solve a real problem.

Real problem examples:

* hard-to-test code
* external service dependency
* need to swap implementations
* package boundary becoming messy

---

# 5. Explain dependency injection simply

Dependency injection means:

> Instead of creating a dependency inside a function, pass it from outside.

Bad for testing:

```go
func Notify(message string) error {
    client := slack.NewClient("real-webhook-url")
    return client.Send(message)
}
```

Problem:

* `Notify` always uses real Slack.
* You cannot easily test without sending real messages.
* The webhook URL may be hardcoded.
* The function controls too much.

Better:

```go
func Notify(sender Sender, message string) error {
    return sender.Send(message)
}
```

Now the caller decides what sender to use.

In production:

```go
sender := slack.NewClient(webhookURL)
Notify(sender, "Build failed")
```

In tests:

```go
sender := FakeSender{}
Notify(sender, "Build failed")
```

Python comparison:

Bad:

```python
def notify(message):
    client = SlackClient("real-url")
    client.send(message)
```

Better:

```python
def notify(sender, message):
    sender.send(message)
```

This is dependency injection.

It is not a complicated framework idea. It simply means:

> Give the function or struct what it needs, instead of making it create everything itself.

---

# 6. Explain how to refactor the project gradually

Refactoring means:

> Change the structure of code without changing its behavior.

A beginner mistake is trying to rewrite everything at once.

Do not do that.

Instead, refactor in tiny safe steps.

---

## Current possible structure

You may currently have something like:

```text
slack-integration/
  main.go
```

With everything inside `main.go`.

Example:

```go
func main() {
    webhookURL := os.Getenv("SLACK_WEBHOOK_URL")

    message := "Build failed"

    payload := map[string]string{
        "text": message,
    }

    // send HTTP request to Slack
}
```

This works, but over time it becomes hard to test.

---

## Better gradual structure

Move step by step:

```text
slack-integration/
  go.mod
  main.go
  internal/
    config/
      config.go
    notifier/
      service.go
      service_test.go
    slack/
      client.go
```

Why `internal/`?

In Go, `internal` means:

> This code is private to this project.

Other projects cannot import it directly.

Python comparison:

Python does not have the exact same enforced `internal/` rule. You may use `_private_module.py` by convention, but Go can enforce this at compile time.

---

## Safe refactoring process

Step 1: Run tests or run the app before changing anything.

```bash
go test ./...
go run .
```

Step 2: Move config loading into `config` package.

Step 3: Keep behavior same.

Step 4: Move Slack HTTP sending into `slack` package.

Step 5: Add `Sender` interface in the package that needs it.

Step 6: Change notifier to depend on `Sender`.

Step 7: Add fake sender test.

Step 8: Keep `main.go` as wiring code only.

Important mindset:

> Refactoring is not about adding new features. It is about making current behavior cleaner and safer.

---

# 7. Introduce a `Sender` interface for Slack sending

Let us say your notifier currently directly uses Slack:

```go
func NotifyBuildFailed(webhookURL string, projectName string) error {
    client := slack.NewClient(webhookURL)

    message := "Build failed for project: " + projectName

    return client.Send(message)
}
```

This works, but it is tightly connected to Slack.

Better:

```go
type Sender interface {
    Send(message string) error
}
```

Now notifier only needs “something that can send.”

```go
type Service struct {
    sender Sender
}

func NewService(sender Sender) *Service {
    return &Service{sender: sender}
}

func (s *Service) NotifyBuildFailed(projectName string) error {
    message := "Build failed for project: " + projectName
    return s.sender.Send(message)
}
```

Now the notifier does not know Slack exists.

That is powerful.

Production:

```go
slackClient := slack.NewClient(webhookURL)
notifierService := notifier.NewService(slackClient)
```

Testing:

```go
fakeSender := &FakeSender{}
notifierService := notifier.NewService(fakeSender)
```

This is how interfaces connect directly to unit testing and maintainability.

---

# 8. Explain config package design

A config package should answer one question:

> What settings does my app need to run?

Example settings:

* Slack webhook URL
* app environment
* request timeout
* rate limit

Config should not send Slack messages.

Config should not contain business logic.

Config should only load and validate settings.

Good:

```go
type Config struct {
    SlackWebhookURL    string
    AppEnv             string
    HTTPTimeoutSeconds int
    RateLimitPerMinute int
}
```

Python equivalent:

```python
@dataclass
class Config:
    slack_webhook_url: str
    app_env: str
    http_timeout_seconds: int
    rate_limit_per_minute: int
```

Go config loader:

```go
func Load() (Config, error) {
    // read env vars
    // validate required values
    // return Config
}
```

Python equivalent:

```python
def load_config() -> Config:
    # read os.environ
    # validate
    # return Config(...)
```

---

## Config package responsibilities

Good responsibilities:

```text
config/
  - read environment variables
  - provide defaults
  - validate required values
  - return a Config struct
```

Bad responsibilities:

```text
config/
  - send Slack messages
  - start HTTP server
  - call business logic
  - know about notifier internals
```

Keep config boring.

Boring code is good architecture.

---

# 9. Show how to keep `main.go` small

`main.go` should mainly do wiring.

Think of `main.go` as the place where you connect the parts.

It should not contain all business logic.

Good `main.go` flow:

```text
load config
create Slack client
create notifier service
call notifier
handle error
```

That is it.

Python comparison:

Python `main.py` often does the same:

```python
def main():
    config = load_config()
    sender = SlackSender(config.slack_webhook_url)
    service = NotifierService(sender)
    service.notify_build_failed("billing-service")
```

Go version:

```go
func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    sender := slack.NewClient(cfg.SlackWebhookURL)
    service := notifier.NewService(sender)

    if err := service.NotifyBuildFailed("billing-service"); err != nil {
        log.Fatal(err)
    }
}
```

Small `main.go` is easier to read, test, and maintain.

---

# 10. Pseudocode first

Before writing Go code, think like this:

```text
Config:
    read SLACK_WEBHOOK_URL
    read APP_ENV
    read HTTP_TIMEOUT_SECONDS
    validate Slack webhook URL exists
    return config object

SlackClient:
    store webhook URL
    implement Send(message)
    convert message to JSON
    send HTTP request to Slack

Notifier:
    accept a Sender
    create a useful message
    ask Sender to send it

Main:
    load config
    create Slack client
    create notifier service
    send one notification
```

Testing pseudocode:

```text
FakeSender:
    store sent messages in a slice
    when Send is called:
        save message
        return nil

Test:
    create FakeSender
    inject it into Notifier
    call NotifyBuildFailed
    check that one message was sent
    check message contains project name
```

Python comparison:

```python
class FakeSender:
    def __init__(self):
        self.messages = []

    def send(self, message):
        self.messages.append(message)
```

Go fake:

```go
type FakeSender struct {
    messages []string
}

func (f *FakeSender) Send(message string) error {
    f.messages = append(f.messages, message)
    return nil
}
```

---

# 11. Real Go code examples

Recommended project structure:

```text
slack-integration/
  go.mod
  main.go
  internal/
    config/
      config.go
    notifier/
      service.go
      service_test.go
    slack/
      client.go
```

---

## `internal/config/config.go`

```go
package config

import (
    "errors"
    "os"
    "strconv"
)

type Config struct {
    SlackWebhookURL    string
    AppEnv             string
    HTTPTimeoutSeconds int
    RateLimitPerMinute int
}

func Load() (Config, error) {
    cfg := Config{
        SlackWebhookURL:    os.Getenv("SLACK_WEBHOOK_URL"),
        AppEnv:             getEnv("APP_ENV", "local"),
        HTTPTimeoutSeconds: getEnvAsInt("HTTP_TIMEOUT_SECONDS", 10),
        RateLimitPerMinute: getEnvAsInt("RATE_LIMIT_PER_MINUTE", 60),
    }

    if cfg.SlackWebhookURL == "" {
        return Config{}, errors.New("SLACK_WEBHOOK_URL is required")
    }

    return cfg, nil
}

func getEnv(key string, fallback string) string {
    value := os.Getenv(key)
    if value == "" {
        return fallback
    }
    return value
}

func getEnvAsInt(key string, fallback int) int {
    value := os.Getenv(key)
    if value == "" {
        return fallback
    }

    parsedValue, err := strconv.Atoi(value)
    if err != nil {
        return fallback
    }

    return parsedValue
}
```

Python comparison:

```python
import os

def get_env(key, fallback):
    return os.environ.get(key, fallback)
```

Important Go syntax notes:

```go
func Load() (Config, error)
```

This means the function returns two values:

1. `Config`
2. `error`

Python usually raises exceptions. Go usually returns errors explicitly.

---

## `internal/notifier/service.go`

```go
package notifier

import (
    "fmt"
)

type Sender interface {
    Send(message string) error
}

type Service struct {
    sender Sender
}

func NewService(sender Sender) *Service {
    return &Service{
        sender: sender,
    }
}

func (s *Service) NotifyBuildFailed(projectName string) error {
    message := fmt.Sprintf("Build failed for project: %s", projectName)

    return s.sender.Send(message)
}

func (s *Service) NotifyDeploymentSucceeded(projectName string, environment string) error {
    message := fmt.Sprintf(
        "Deployment succeeded for project: %s in environment: %s",
        projectName,
        environment,
    )

    return s.sender.Send(message)
}
```

Important Go syntax:

```go
func (s *Service) NotifyBuildFailed(projectName string) error
```

This is a method.

Python equivalent:

```python
class Service:
    def notify_build_failed(self, project_name: str):
        ...
```

In Go, this part is the receiver:

```go
(s *Service)
```

It is similar to Python’s `self`, but written before the method name.

---

## `internal/slack/client.go`

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
    webhookURL string
    httpClient *http.Client
}

func NewClient(webhookURL string, timeoutSeconds int) *Client {
    return &Client{
        webhookURL: webhookURL,
        httpClient: &http.Client{
            Timeout: time.Duration(timeoutSeconds) * time.Second,
        },
    }
}

func (c *Client) Send(message string) error {
    payload := map[string]string{
        "text": message,
    }

    body, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("marshal slack payload: %w", err)
    }

    response, err := c.httpClient.Post(
        c.webhookURL,
        "application/json",
        bytes.NewBuffer(body),
    )
    if err != nil {
        return fmt.Errorf("send slack message: %w", err)
    }
    defer response.Body.Close()

    if response.StatusCode < 200 || response.StatusCode >= 300 {
        return fmt.Errorf("slack returned status code: %d", response.StatusCode)
    }

    return nil
}
```

Why does this satisfy the `Sender` interface?

Because it has this method:

```go
func (c *Client) Send(message string) error
```

And the interface requires this:

```go
type Sender interface {
    Send(message string) error
}
```

No extra declaration is needed.

---

## `main.go`

Replace this import path with your actual module name from `go.mod`.

Example `go.mod`:

```go
module slack-integration
```

Then `main.go`:

```go
package main

import (
    "log"

    "slack-integration/internal/config"
    "slack-integration/internal/notifier"
    "slack-integration/internal/slack"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    slackClient := slack.NewClient(
        cfg.SlackWebhookURL,
        cfg.HTTPTimeoutSeconds,
    )

    notifierService := notifier.NewService(slackClient)

    err = notifierService.NotifyBuildFailed("billing-service")
    if err != nil {
        log.Fatal(err)
    }
}
```

Now `main.go` is small.

It only wires things together.

It does not know how to:

* parse JSON
* build Slack HTTP requests
* create notification messages
* validate environment variables internally

Each package has one job.

---

## `internal/notifier/service_test.go`

```go
package notifier

import (
    "errors"
    "strings"
    "testing"
)

type fakeSender struct {
    messages []string
    err      error
}

func (f *fakeSender) Send(message string) error {
    if f.err != nil {
        return f.err
    }

    f.messages = append(f.messages, message)
    return nil
}

func TestNotifyBuildFailedSendsMessage(t *testing.T) {
    sender := &fakeSender{}
    service := NewService(sender)

    err := service.NotifyBuildFailed("billing-service")
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    if len(sender.messages) != 1 {
        t.Fatalf("expected 1 message, got %d", len(sender.messages))
    }

    got := sender.messages[0]

    if !strings.Contains(got, "billing-service") {
        t.Fatalf("expected message to contain project name, got %q", got)
    }
}

func TestNotifyBuildFailedReturnsSenderError(t *testing.T) {
    sender := &fakeSender{
        err: errors.New("slack failed"),
    }

    service := NewService(sender)

    err := service.NotifyBuildFailed("billing-service")
    if err == nil {
        t.Fatal("expected error, got nil")
    }
}
```

This test does not call real Slack.

That is the big win.

The interface made the notifier testable.

Python equivalent:

```python
class FakeSender:
    def __init__(self):
        self.messages = []

    def send(self, message):
        self.messages.append(message)

def test_notify_build_failed():
    sender = FakeSender()
    service = NotifierService(sender)

    service.notify_build_failed("billing-service")

    assert len(sender.messages) == 1
    assert "billing-service" in sender.messages[0]
```

---

# 12. Hands-on tasks

## Task 1: Create the `Sender` interface

Inside:

```text
internal/notifier/service.go
```

Create:

```go
type Sender interface {
    Send(message string) error
}
```

Then make `Service` depend on `Sender`.

---

## Task 2: Move Slack sending into `internal/slack`

Create:

```text
internal/slack/client.go
```

Add:

```go
type Client struct {
    webhookURL string
}
```

Then implement:

```go
func (c *Client) Send(message string) error
```

---

## Task 3: Add config loading

Create:

```text
internal/config/config.go
```

Support these env vars:

```text
SLACK_WEBHOOK_URL
APP_ENV
HTTP_TIMEOUT_SECONDS
RATE_LIMIT_PER_MINUTE
```

Make `SLACK_WEBHOOK_URL` required.

---

## Task 4: Keep `main.go` small

Your final `main.go` should mostly do:

```text
load config
create slack client
create notifier service
send notification
```

---

## Task 5: Add tests using fake sender

Test:

1. successful notification
2. sender error
3. message contains project name

---

# 13. Expected output

When config is missing:

```bash
go run .
```

Expected output:

```text
SLACK_WEBHOOK_URL is required
```

When config is present:

```bash
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/your/webhook/url"
export APP_ENV="local"
export HTTP_TIMEOUT_SECONDS="10"
export RATE_LIMIT_PER_MINUTE="60"

go run .
```

Expected behavior:

```text
A Slack message is sent.
```

The program may not print anything if success is silent.

That is okay.

You can temporarily add:

```go
log.Println("notification sent successfully")
```

After:

```go
err = notifierService.NotifyBuildFailed("billing-service")
```

Example:

```go
if err != nil {
    log.Fatal(err)
}

log.Println("notification sent successfully")
```

Then expected output:

```text
notification sent successfully
```

For tests:

```bash
go test ./...
```

Expected output:

```text
ok   slack-integration/internal/notifier
```

Or similar package paths depending on your module name.

---

# 14. Common mistakes

## Mistake 1: Putting the interface in the wrong package too early

Beginner-friendly rule:

> Define the interface where it is used, not where it is implemented.

The notifier needs a sender, so this is good:

```go
package notifier

type Sender interface {
    Send(message string) error
}
```

Do not rush to create:

```text
internal/interfaces/
```

That often becomes unnecessary overengineering.

---

## Mistake 2: Creating too many interfaces

Bad:

```go
type ConfigLoader interface {
    Load() Config
}

type MessageFormatter interface {
    Format() string
}

type ErrorHandler interface {
    Handle(error)
}
```

This is too much for a beginner project.

Good:

```go
type Sender interface {
    Send(message string) error
}
```

This interface solves a real problem: testing Slack sending.

---

## Mistake 3: Making `main.go` do everything

Bad:

```go
func main() {
    // read env
    // build message
    // encode JSON
    // send HTTP request
    // handle Slack response
    // retry
    // log
}
```

Better:

```go
func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    sender := slack.NewClient(cfg.SlackWebhookURL, cfg.HTTPTimeoutSeconds)
    service := notifier.NewService(sender)

    if err := service.NotifyBuildFailed("billing-service"); err != nil {
        log.Fatal(err)
    }
}
```

---

## Mistake 4: Hardcoding config

Bad:

```go
webhookURL := "https://hooks.slack.com/services/real-url"
```

Better:

```go
webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
```

Even better:

```go
cfg, err := config.Load()
```

Never commit real webhook URLs.

---

## Mistake 5: Ignoring errors

Bad:

```go
service.NotifyBuildFailed("billing-service")
```

Good:

```go
if err := service.NotifyBuildFailed("billing-service"); err != nil {
    log.Fatal(err)
}
```

Python comparison:

In Python, ignoring errors may look like:

```python
service.notify_build_failed("billing-service")
```

But exceptions will crash unless caught.

In Go, errors are returned as values. You must check them.

---

# 15. Debugging tips

## Tip 1: Confirm env vars are loaded

Add temporary logging:

```go
log.Println("app env:", cfg.AppEnv)
log.Println("timeout:", cfg.HTTPTimeoutSeconds)
```

Do not print secrets like full Slack webhook URLs.

Safer:

```go
if cfg.SlackWebhookURL != "" {
    log.Println("slack webhook url is configured")
}
```

---

## Tip 2: Check interface method signature exactly

This interface:

```go
type Sender interface {
    Send(message string) error
}
```

Requires exactly:

```go
Send(string) error
```

This will satisfy it:

```go
func (c *Client) Send(message string) error
```

This will not:

```go
func (c *Client) Send(message string)
```

Because it does not return `error`.

This will not:

```go
func (c *Client) Send(text string, channel string) error
```

Because it has an extra parameter.

---

## Tip 3: Use compile errors as guidance

If you see:

```text
*slack.Client does not implement notifier.Sender
```

It usually means the method signature does not match.

Check:

* method name
* parameters
* return values
* pointer receiver vs value receiver

---

## Tip 4: Test without Slack first

Before testing real Slack, run:

```bash
go test ./...
```

Your fake sender tests should pass without internet and without a webhook.

That is the benefit of dependency injection.

---

## Tip 5: Print request status code during Slack debugging

Inside Slack client:

```go
log.Println("slack status code:", response.StatusCode)
```

A non-2xx status means Slack did not accept the request.

---

# 16. One DSA topic: Prefix sum

Prefix sum is a simple technique for quickly finding the sum of a range.

Suppose you have:

```text
nums = [2, 4, 1, 3]
```

Normal range sum from index `1` to `3`:

```text
4 + 1 + 3 = 8
```

If you do this repeatedly, it can become slow.

Prefix sum stores running totals.

```text
nums:    [2, 4, 1, 3]
prefix:  [0, 2, 6, 7, 10]
```

Why start with `0`?

Because it makes range calculation easier.

Formula:

```text
sum from left to right = prefix[right + 1] - prefix[left]
```

Example:

```text
sum index 1 to 3 = prefix[4] - prefix[1]
                 = 10 - 2
                 = 8
```

Python version:

```python
nums = [2, 4, 1, 3]

prefix = [0]

for num in nums:
    prefix.append(prefix[-1] + num)
```

Go version:

```go
nums := []int{2, 4, 1, 3}

prefix := make([]int, len(nums)+1)

for i := 0; i < len(nums); i++ {
    prefix[i+1] = prefix[i] + nums[i]
}
```

Key syntax difference:

Python:

```python
prefix.append(...)
```

Go:

```go
prefix[i+1] = ...
```

In Go, slices can grow with `append`, but here we already know the size, so `make` is clean.

---

# 17. One Go DSA problem: Range Sum Query

## Problem

Given an integer slice `nums`, build a prefix sum array and answer range sum queries.

Example:

```text
nums = [2, 4, 1, 3]

Query left=1, right=3
Answer = 8
```

Because:

```text
nums[1] + nums[2] + nums[3] = 4 + 1 + 3 = 8
```

---

## Go solution

```go
package main

import "fmt"

type NumArray struct {
    prefix []int
}

func Constructor(nums []int) NumArray {
    prefix := make([]int, len(nums)+1)

    for i := 0; i < len(nums); i++ {
        prefix[i+1] = prefix[i] + nums[i]
    }

    return NumArray{
        prefix: prefix,
    }
}

func (n NumArray) SumRange(left int, right int) int {
    return n.prefix[right+1] - n.prefix[left]
}

func main() {
    nums := []int{2, 4, 1, 3}

    numArray := Constructor(nums)

    fmt.Println(numArray.SumRange(1, 3))
    fmt.Println(numArray.SumRange(0, 2))
    fmt.Println(numArray.SumRange(2, 2))
}
```

Expected output:

```text
8
7
1
```

Python equivalent:

```python
class NumArray:
    def __init__(self, nums):
        self.prefix = [0]

        for num in nums:
            self.prefix.append(self.prefix[-1] + num)

    def sum_range(self, left, right):
        return self.prefix[right + 1] - self.prefix[left]
```

---

## Why this matters for backend work

Prefix sum teaches a useful backend idea:

> Precompute once, answer quickly many times.

Backend examples:

* daily usage totals
* API request counts
* billing summaries
* analytics dashboards
* rate limit windows

This connects to today’s module-based task.

---

# 18. One module-based practice task: Config-driven rate limiter

Today’s practice module:

> Build a simple config-driven rate limiter.

The goal is not production-level rate limiting.

The goal is to practice:

* structs
* config
* small packages
* method receivers
* testable design

---

## Simple behavior

Allow only `N` requests.

Example:

```text
RATE_LIMIT_PER_MINUTE=3
```

Then:

```text
request 1 -> allowed
request 2 -> allowed
request 3 -> allowed
request 4 -> blocked
```

---

## Suggested structure

```text
internal/
  ratelimiter/
    limiter.go
    limiter_test.go
```

---

## `internal/ratelimiter/limiter.go`

```go
package ratelimiter

type Limiter struct {
    limit int
    count int
}

func NewLimiter(limit int) *Limiter {
    return &Limiter{
        limit: limit,
        count: 0,
    }
}

func (l *Limiter) Allow() bool {
    if l.count >= l.limit {
        return false
    }

    l.count++
    return true
}
```

Python equivalent:

```python
class Limiter:
    def __init__(self, limit):
        self.limit = limit
        self.count = 0

    def allow(self):
        if self.count >= self.limit:
            return False

        self.count += 1
        return True
```

---

## Use it from `main.go`

```go
limiter := ratelimiter.NewLimiter(cfg.RateLimitPerMinute)

if !limiter.Allow() {
    log.Fatal("rate limit exceeded")
}
```

Then send Slack notification only when allowed:

```go
if limiter.Allow() {
    err := notifierService.NotifyBuildFailed("billing-service")
    if err != nil {
        log.Fatal(err)
    }
}
```

This is config-driven because the limit comes from:

```go
cfg.RateLimitPerMinute
```

Not from hardcoded code.

---

## Simple test

```go
package ratelimiter

import "testing"

func TestLimiterAllowsUntilLimit(t *testing.T) {
    limiter := NewLimiter(2)

    if !limiter.Allow() {
        t.Fatal("expected first request to be allowed")
    }

    if !limiter.Allow() {
        t.Fatal("expected second request to be allowed")
    }

    if limiter.Allow() {
        t.Fatal("expected third request to be blocked")
    }
}
```

This is intentionally simple.

A real rate limiter would need time windows, concurrency safety, mutexes, and cleanup.

Not today.

Today we want clean module thinking.

---

# 19. Revision checkpoint

You should now be able to answer these:

## Interfaces

What does this mean?

```go
type Sender interface {
    Send(message string) error
}
```

Answer:

> Any type with a `Send(string) error` method can be used as a `Sender`.

---

## Struct vs interface

What is this?

```go
type Client struct {
    webhookURL string
}
```

Answer:

> A concrete struct that stores data.

What is this?

```go
type Sender interface {
    Send(message string) error
}
```

Answer:

> A behavior contract.

---

## Dependency injection

What is dependency injection?

Answer:

> Passing required dependencies from outside instead of creating them inside.

Bad:

```go
func Notify() {
    sender := slack.NewClient(...)
}
```

Good:

```go
func NewService(sender Sender) *Service {
    return &Service{sender: sender}
}
```

---

## Testing

Why did `Sender` help testing?

Answer:

> Because tests can inject a fake sender instead of using real Slack.

---

## Config

What should config package do?

Answer:

> Load, default, and validate settings.

---

## Refactoring

What is the beginner-safe refactoring mindset?

Answer:

> Change one small thing at a time, keep behavior the same, and run tests often.

---

# 20. Homework

## Homework part 1: Refactor notifier

Create this structure:

```text
internal/
  config/
    config.go
  notifier/
    service.go
    service_test.go
  slack/
    client.go
```

Your goal:

* `main.go` should not build Slack JSON manually.
* `notifier` should not know about webhook URLs.
* `slack.Client` should implement `Send(message string) error`.
* `notifier.Service` should depend on `Sender`.

---

## Homework part 2: Add one more notification method

Add:

```go
func (s *Service) NotifyHighErrorRate(serviceName string, errorCount int) error
```

Expected message:

```text
High error rate detected for service: payment-service, errors: 42
```

Test it using `fakeSender`.

---

## Homework part 3: Add config validation

Update `config.Load()` so:

* `SLACK_WEBHOOK_URL` is required
* `HTTP_TIMEOUT_SECONDS` defaults to `10`
* `RATE_LIMIT_PER_MINUTE` defaults to `60`
* invalid numbers fall back to defaults

---

## Homework part 4: Complete prefix sum problem

Write this function:

```go
func BuildPrefix(nums []int) []int {
    // your code here
}
```

Then write:

```go
func RangeSum(prefix []int, left int, right int) int {
    // your code here
}
```

Expected:

```go
nums := []int{5, 2, 7, 1}

prefix := BuildPrefix(nums)

fmt.Println(RangeSum(prefix, 0, 1)) // 7
fmt.Println(RangeSum(prefix, 1, 3)) // 10
fmt.Println(RangeSum(prefix, 2, 2)) // 7
```

---

## Homework part 5: Build simple rate limiter module

Create:

```text
internal/ratelimiter/
  limiter.go
  limiter_test.go
```

Implement:

```go
type Limiter struct {
    limit int
    count int
}
```

With:

```go
func NewLimiter(limit int) *Limiter
func (l *Limiter) Allow() bool
```

Then connect it to config:

```go
limiter := ratelimiter.NewLimiter(cfg.RateLimitPerMinute)
```

---

# Final mental model for Day 13

A beginner-friendly architecture for your Slack integration should feel like this:

```text
main.go
  connects everything

config
  loads settings

notifier
  decides what message to send

slack
  knows how to send to Slack

ratelimiter
  decides whether sending is allowed
```

The most important lesson:

> Interfaces are not for making code look advanced. Interfaces are for making code easier to replace, test, and maintain.

For your project, this is the key interface:

```go
type Sender interface {
    Send(message string) error
}
```

That small interface gives you a big benefit:

* real Slack in production
* fake sender in tests
* cleaner notifier logic
* smaller `main.go`
* safer future refactoring
