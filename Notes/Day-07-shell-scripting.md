# Day 7 — Shell Scripting for `slack-integration` Project

## 1. Day 7 learning goals

By the end of Day 7, you should understand:

1. What shell scripting is.
2. Why DevOps/backend teams use shell scripts.
3. How scripts help run Go CLI commands repeatedly.
4. How to use variables, arguments, environment variables, and exit codes.
5. How `set -euo pipefail` makes scripts safer.
6. How local helper scripts connect to Tekton-style automation.
7. How to write small scripts like:

   * `local-run.sh`
   * `test-all.sh`
   * `collect-failure-trace.sh`
   * `log-filter.sh`
8. Basic sorting concepts in DSA.
9. One easy sorting problem in Go.

---

# 2. Quick revision of Days 1 to 6

## Day 1 — Go CLI basics

You learned:

```text
CLI flags → Go program → process input → print or send result
```

Example:

```bash
go run cmd/slack-notifier/main.go \
  --event-type pr \
  --status failed
```

In Python, this is similar to:

```python
python app.py --event-type pr --status failed
```

---

## Day 2 — Structs and event model

You learned that Go uses structs to group related data.

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

Main idea:

```text
Loose CLI values → structured PipelineEvent object
```

---

## Day 3 — JSON, HTTP, Slack webhook

You learned how Go converts structs into JSON and sends HTTP POST requests.

Flow:

```text
PipelineEvent → SlackMessage struct → JSON → HTTP POST → Slack webhook
```

Python equivalent:

```python
requests.post(webhook_url, json=payload)
```

---

## Day 4 — Router logic

You learned how routing chooses the correct Slack webhook.

Example:

```text
event-type=pr  → PR webhook
event-type=cd  → CD webhook
event-type=job → job webhook or fallback CD webhook
```

Router means:

```text
Look at event type → choose destination
```

---

## Day 5 — Error handling and logging

You learned:

```go
if err != nil {
    return err
}
```

In Go, errors are explicit.

Python equivalent:

```python
try:
    do_something()
except Exception as e:
    print(e)
```

Go convention:

```text
Check error immediately.
Return meaningful error.
Log useful context.
```

---

## Day 6 — Testing

You learned Go testing basics:

```bash
go test ./...
```

Go test file naming:

```text
event.go       → production code
event_test.go  → test code
```

You also learned:

```text
table-driven tests
mock HTTP server
router testing
payload testing
```

---

# 3. Explain shell scripting in very simple language

A shell script is a file that contains terminal commands.

Instead of typing many commands again and again, you save them in one file and run that file.

Example: instead of typing this every time:

```bash
go test ./...
go run cmd/slack-notifier/main.go --event-type pr --status failed
```

You can create a script:

```bash
./scripts/test-all.sh
./scripts/local-run.sh
```

Simple meaning:

```text
Shell script = saved terminal commands
```

Python comparison:

```text
Shell script is like a small automation script.
Python automates using Python code.
Shell automates using terminal commands.
```

Example Python automation:

```python
import os

os.system("go test ./...")
```

Shell version:

```bash
go test ./...
```

For DevOps work, shell scripts are very common because they directly control terminal commands, tools, files, Docker, Kubernetes, Git, Go, and Tekton.

---

# 4. What is a script and why teams use it?

A script is a reusable command file.

Teams use scripts because they make work:

```text
repeatable
simple
consistent
less error-prone
easy to share
easy to run in CI/CD
```

Without script:

```bash
go run cmd/slack-notifier/main.go --event-type pr --stage failure --status failed --pipeline-name pr-check
```

Every developer may type it differently.

With script:

```bash
./scripts/local-run.sh pr failure failed
```

Now everyone runs the same command.

---

## Real project usage

In your `slack-integration` project, scripts can help with:

```text
local-run.sh              → run Go notifier locally
test-all.sh               → run all Go tests
collect-failure-trace.sh  → collect logs after failure
log-filter.sh             → filter important logs
```

In Tekton, scripts are useful because Tekton tasks often run shell commands inside containers.

Example Tekton step:

```yaml
script: |
  #!/bin/sh
  go test ./...
```

So learning shell helps you understand both:

```text
local development + CI/CD automation
```

---

# 5. Shebang, variables, args, env vars, and exit codes

## 5.1 Shebang

Shebang tells the system which shell should run the script.

```bash
#!/usr/bin/env bash
```

Meaning:

```text
Run this script using bash.
```

Example file:

```bash
#!/usr/bin/env bash

echo "Hello from script"
```

Python comparison:

```python
#!/usr/bin/env python3

print("Hello from Python")
```

Same idea:

```text
Shebang tells OS which interpreter to use.
```

---

## 5.2 Shell variables

Shell variable:

```bash
name="Radhe"
echo "$name"
```

Output:

```text
Radhe
```

Important rule:

```bash
name="Radhe"   # correct
name = "Radhe" # wrong in shell
```

There should be no spaces around `=`.

Python equivalent:

```python
name = "Radhe"
print(name)
```

Shell convention:

```bash
APP_NAME="slack-notifier"
```

Many shell variables use uppercase.

---

## 5.3 Arguments

Arguments are values passed to a script from the command line.

Script:

```bash
#!/usr/bin/env bash

echo "First argument: $1"
echo "Second argument: $2"
```

Run:

```bash
./demo.sh pr failed
```

Output:

```text
First argument: pr
Second argument: failed
```

Meaning:

```text
$1 = first argument
$2 = second argument
$3 = third argument
```

Python equivalent:

```python
import sys

print(sys.argv[1])
print(sys.argv[2])
```

---

## 5.4 Environment variables

Environment variables are values available to the running process.

Example:

```bash
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/xxx"
```

Then your script or Go program can read it.

Shell:

```bash
echo "$SLACK_WEBHOOK_URL"
```

Go:

```go
webhook := os.Getenv("SLACK_WEBHOOK_URL")
```

Python:

```python
import os

webhook = os.getenv("SLACK_WEBHOOK_URL")
```

Simple meaning:

```text
Environment variable = external configuration passed to program
```

Real backend convention:

```text
Do not hardcode secrets in code.
Read them from environment variables.
```

Good:

```go
webhook := os.Getenv("SLACK_WEBHOOK_URL")
```

Bad:

```go
webhook := "https://hooks.slack.com/services/secret"
```

---

## 5.5 Exit codes

Every command returns an exit code.

```text
0     = success
non-0 = failure
```

Example:

```bash
echo "hello"
echo $?
```

`$?` means previous command exit code.

Example:

```bash
ls existing-file.txt
echo $?
```

If file exists:

```text
0
```

If file does not exist:

```text
1 or non-zero
```

Python equivalent:

```python
import sys

sys.exit(0)  # success
sys.exit(1)  # failure
```

Go equivalent:

```go
os.Exit(0) // success
os.Exit(1) // failure
```

In CI/CD and Tekton:

```text
exit code 0     → task passed
exit code non-0 → task failed
```

This is very important.

---

# 6. Explain `set -euo pipefail` clearly

At the top of many professional shell scripts, you see:

```bash
set -euo pipefail
```

This makes the script safer.

Think of it like strict mode.

---

## `set -e`

Stop script immediately if any command fails.

Example without `set -e`:

```bash
mkdir /root/test
echo "Still running"
```

Even if `mkdir` fails, script may continue.

With:

```bash
set -e
```

Script stops immediately on failure.

Python comparison:

```python
raise Exception("stop here")
```

---

## `set -u`

Fail if you use an undefined variable.

Example:

```bash
echo "$WEBHOOK_URL"
```

If `WEBHOOK_URL` is not set, script fails.

Without `set -u`, it may silently use an empty value.

This prevents dangerous bugs.

Python comparison:

```python
print(webhook_url)
```

If variable is not defined, Python raises:

```text
NameError
```

---

## `pipefail`

Normally shell pipelines can hide failures.

Example:

```bash
cat missing-file.txt | grep ERROR
```

Without `pipefail`, the full command may not correctly show the first failure.

With:

```bash
set -o pipefail
```

If any command in the pipe fails, the whole pipeline fails.

---

## Final meaning

```bash
set -euo pipefail
```

Means:

```text
-e        stop on command failure
-u        fail on undefined variable
pipefail  fail if any command in pipeline fails
```

Beginner-friendly meaning:

```text
Do not silently continue when something is wrong.
```

This is useful in:

```text
local scripts
CI/CD scripts
Tekton task scripts
deployment scripts
debugging scripts
```

---

# 7. How shell scripts can wrap Go CLI commands

Your Go CLI may require many flags.

Example:

```bash
go run cmd/slack-notifier/main.go \
  --event-type pr \
  --stage failure \
  --status failed \
  --repository cloud-resource-onboarding \
  --branch feature/slack-alert \
  --pipeline-name pr-validation \
  --failed-step unit-tests \
  --error-message "unit tests failed"
```

This is long.

A shell script can wrap it.

Then you run:

```bash
./scripts/local-run.sh
```

The script internally runs the long Go command.

Simple flow:

```text
Developer runs script
        ↓
Script prepares values
        ↓
Script calls Go CLI
        ↓
Go CLI builds event
        ↓
Go sends Slack message
```

ASCII diagram:

```text
+------------------+
| Developer        |
| ./local-run.sh   |
+--------+---------+
         |
         v
+------------------+
| Shell Script     |
| env + args       |
+--------+---------+
         |
         v
+------------------+
| Go CLI           |
| main.go          |
+--------+---------+
         |
         v
+------------------+
| Slack Client     |
| webhook POST     |
+------------------+
```

---

# 8. Pseudocode first for a helper script

Goal:

```text
Create a local helper script to test failed PR notification.
```

Pseudocode:

```text
START

Enable strict script mode

Read EVENT_TYPE from first argument
Read STATUS from second argument

If EVENT_TYPE is empty:
    print usage message
    exit with failure

If STATUS is empty:
    print usage message
    exit with failure

Read SLACK_WEBHOOK_URL from environment variable

If SLACK_WEBHOOK_URL is empty:
    print error
    exit with failure

Print what script is going to run

Run Go CLI with:
    event type
    status
    repository
    branch
    pipeline name
    failed step
    error message

If Go command succeeds:
    print success
Else:
    script fails automatically due to set -e

END
```

---

# 9. Real shell script examples

Recommended folder:

```text
slack-integration/
├── cmd/
│   └── slack-notifier/
│       └── main.go
├── pkg/
│   └── notify/
├── scripts/
│   ├── local-run.sh
│   ├── test-all.sh
│   ├── collect-failure-trace.sh
│   └── log-filter.sh
└── go.mod
```

---

## Example 1: Very small beginner script

File:

```text
scripts/hello.sh
```

Code:

```bash
#!/usr/bin/env bash

echo "Hello from shell script"
echo "Today we are learning DevOps helper scripts"
```

Run:

```bash
chmod +x scripts/hello.sh
./scripts/hello.sh
```

Expected output:

```text
Hello from shell script
Today we are learning DevOps helper scripts
```

Important command:

```bash
chmod +x scripts/hello.sh
```

Meaning:

```text
Make this file executable.
```

---

## Example 2: Script with arguments

File:

```text
scripts/print-event.sh
```

Code:

```bash
#!/usr/bin/env bash
set -euo pipefail

EVENT_TYPE="${1:-}"
STATUS="${2:-}"

if [[ -z "$EVENT_TYPE" ]]; then
  echo "Error: event type is required"
  echo "Usage: ./scripts/print-event.sh <event-type> <status>"
  exit 1
fi

if [[ -z "$STATUS" ]]; then
  echo "Error: status is required"
  echo "Usage: ./scripts/print-event.sh <event-type> <status>"
  exit 1
fi

echo "Event type: $EVENT_TYPE"
echo "Status: $STATUS"
```

Run:

```bash
chmod +x scripts/print-event.sh
./scripts/print-event.sh pr failed
```

Expected output:

```text
Event type: pr
Status: failed
```

Python equivalent:

```python
import sys

event_type = sys.argv[1]
status = sys.argv[2]

print(f"Event type: {event_type}")
print(f"Status: {status}")
```

Key difference:

```text
Shell uses $1, $2
Python uses sys.argv[1], sys.argv[2]
```

---

## Example 3: `local-run.sh` wrapping Go CLI

File:

```text
scripts/local-run.sh
```

Code:

```bash
#!/usr/bin/env bash
set -euo pipefail

EVENT_TYPE="${1:-pr}"
STAGE="${2:-failure}"
STATUS="${3:-failed}"

REPOSITORY="${REPOSITORY:-cloud-resource-onboarding}"
BRANCH="${BRANCH:-feature/slack-integration}"
PIPELINE_NAME="${PIPELINE_NAME:-pr-validation-pipeline}"
PIPELINE_RUN_NAME="${PIPELINE_RUN_NAME:-local-pr-run-001}"
FAILED_STEP="${FAILED_STEP:-unit-tests}"
ERROR_MESSAGE="${ERROR_MESSAGE:-unit tests failed in local run}"

if [[ -z "${SLACK_WEBHOOK_URL:-}" ]]; then
  echo "Error: SLACK_WEBHOOK_URL environment variable is not set"
  echo "Example:"
  echo "  export SLACK_WEBHOOK_URL='https://hooks.slack.com/services/xxx'"
  exit 1
fi

echo "Running local Slack notifier..."
echo "Event Type : $EVENT_TYPE"
echo "Stage      : $STAGE"
echo "Status     : $STATUS"
echo "Repository : $REPOSITORY"
echo "Branch     : $BRANCH"

go run cmd/slack-notifier/main.go \
  --event-type "$EVENT_TYPE" \
  --stage "$STAGE" \
  --status "$STATUS" \
  --repository "$REPOSITORY" \
  --branch "$BRANCH" \
  --pipeline-name "$PIPELINE_NAME" \
  --pipeline-run-name "$PIPELINE_RUN_NAME" \
  --failed-step "$FAILED_STEP" \
  --error-message "$ERROR_MESSAGE"

echo "Local Slack notifier completed successfully"
```

Run:

```bash
chmod +x scripts/local-run.sh
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/xxx"
./scripts/local-run.sh pr failure failed
```

Expected output:

```text
Running local Slack notifier...
Event Type : pr
Stage      : failure
Status     : failed
Repository : cloud-resource-onboarding
Branch     : feature/slack-integration
Local Slack notifier completed successfully
```

---

# 10. Explain every important line

## Line 1

```bash
#!/usr/bin/env bash
```

This says:

```text
Use bash to run this script.
```

---

## Line 2

```bash
set -euo pipefail
```

This says:

```text
Run safely. Stop when something goes wrong.
```

---

## Arguments with default values

```bash
EVENT_TYPE="${1:-pr}"
STAGE="${2:-failure}"
STATUS="${3:-failed}"
```

Meaning:

```text
Use first argument as EVENT_TYPE.
If first argument is missing, use pr.
```

Example:

```bash
./scripts/local-run.sh cd success succeeded
```

Then:

```text
EVENT_TYPE=cd
STAGE=success
STATUS=succeeded
```

If you run:

```bash
./scripts/local-run.sh
```

Then defaults are used:

```text
EVENT_TYPE=pr
STAGE=failure
STATUS=failed
```

Python equivalent:

```python
event_type = sys.argv[1] if len(sys.argv) > 1 else "pr"
```

---

## Environment variable fallback

```bash
REPOSITORY="${REPOSITORY:-cloud-resource-onboarding}"
```

Meaning:

```text
If REPOSITORY env var exists, use it.
Otherwise use cloud-resource-onboarding.
```

Run with custom value:

```bash
export REPOSITORY="my-service"
./scripts/local-run.sh
```

---

## Required env var check

```bash
if [[ -z "${SLACK_WEBHOOK_URL:-}" ]]; then
```

Meaning:

```text
If SLACK_WEBHOOK_URL is empty, show error and stop.
```

`-z` means:

```text
string is empty
```

Python equivalent:

```python
if not os.getenv("SLACK_WEBHOOK_URL"):
    print("missing webhook")
    sys.exit(1)
```

---

## Running Go command

```bash
go run cmd/slack-notifier/main.go \
  --event-type "$EVENT_TYPE" \
  --stage "$STAGE"
```

The `\` means:

```text
Continue command on next line.
```

This keeps long commands readable.

---

## Why quotes are important

Good:

```bash
--error-message "$ERROR_MESSAGE"
```

Bad:

```bash
--error-message $ERROR_MESSAGE
```

If error message has spaces:

```text
unit tests failed
```

Without quotes, shell may split it into multiple words.

Always quote variables:

```bash
"$VARIABLE"
```

---

# 11. Hands-on tasks

## Task 1: Create hello script

Create:

```text
scripts/hello.sh
```

Code:

```bash
#!/usr/bin/env bash

echo "Hello, shell scripting"
```

Run:

```bash
chmod +x scripts/hello.sh
./scripts/hello.sh
```

---

## Task 2: Create argument script

Create:

```text
scripts/print-event.sh
```

Run:

```bash
./scripts/print-event.sh pr failed
```

Expected:

```text
Event type: pr
Status: failed
```

---

## Task 3: Create local Go wrapper

Create:

```text
scripts/local-run.sh
```

Run:

```bash
export SLACK_WEBHOOK_URL="dummy-or-real-webhook"
./scripts/local-run.sh pr failure failed
```

---

## Task 4: Create test script

Create:

```text
scripts/test-all.sh
```

Code:

```bash
#!/usr/bin/env bash
set -euo pipefail

echo "Running gofmt check..."
gofmt -w .

echo "Running Go tests..."
go test ./...

echo "All checks completed successfully"
```

Run:

```bash
chmod +x scripts/test-all.sh
./scripts/test-all.sh
```

---

## Task 5: Create failure trace collector

Create:

```text
scripts/collect-failure-trace.sh
```

Code:

```bash
#!/usr/bin/env bash
set -euo pipefail

OUTPUT_DIR="${1:-failure-traces}"
LOG_FILE="$OUTPUT_DIR/failure.log"

mkdir -p "$OUTPUT_DIR"

echo "Collecting failure trace..."
echo "Timestamp: $(date)" > "$LOG_FILE"
echo "Project: slack-integration" >> "$LOG_FILE"
echo "Status: failed" >> "$LOG_FILE"
echo "Failed Step: unit-tests" >> "$LOG_FILE"
echo "Error: sample failure for local debugging" >> "$LOG_FILE"

echo "Failure trace saved at: $LOG_FILE"
```

Run:

```bash
chmod +x scripts/collect-failure-trace.sh
./scripts/collect-failure-trace.sh
```

Expected:

```text
Failure trace saved at: failure-traces/failure.log
```

Check file:

```bash
cat failure-traces/failure.log
```

---

# 12. Expected output

After today, your project should have:

```text
scripts/
├── hello.sh
├── print-event.sh
├── local-run.sh
├── test-all.sh
├── collect-failure-trace.sh
└── log-filter.sh
```

You should be able to run:

```bash
./scripts/test-all.sh
```

Output:

```text
Running gofmt check...
Running Go tests...
All checks completed successfully
```

You should also be able to run:

```bash
./scripts/local-run.sh pr failure failed
```

Output:

```text
Running local Slack notifier...
Event Type : pr
Stage      : failure
Status     : failed
...
Local Slack notifier completed successfully
```

---

# 13. Common mistakes

## Mistake 1: Forgetting executable permission

Wrong:

```bash
./scripts/local-run.sh
```

Error:

```text
Permission denied
```

Fix:

```bash
chmod +x scripts/local-run.sh
```

---

## Mistake 2: Spaces around `=`

Wrong:

```bash
EVENT_TYPE = "pr"
```

Correct:

```bash
EVENT_TYPE="pr"
```

Python allows:

```python
event_type = "pr"
```

Shell does not allow spaces around `=`.

---

## Mistake 3: Not quoting variables

Wrong:

```bash
echo $ERROR_MESSAGE
```

Better:

```bash
echo "$ERROR_MESSAGE"
```

Always prefer quotes.

---

## Mistake 4: Missing env var

If script needs:

```bash
SLACK_WEBHOOK_URL
```

But you forgot:

```bash
export SLACK_WEBHOOK_URL="..."
```

Then script should fail clearly.

---

## Mistake 5: Using bash syntax with `sh`

This script uses bash:

```bash
[[ -z "$VAR" ]]
```

So shebang should be:

```bash
#!/usr/bin/env bash
```

Not:

```bash
#!/bin/sh
```

---

# 14. Debugging tips

## Tip 1: Print command before running

```bash
echo "Running tests..."
go test ./...
```

---

## Tip 2: Check previous command exit code

```bash
go test ./...
echo $?
```

If output is:

```text
0
```

Tests passed.

If non-zero:

```text
tests failed
```

---

## Tip 3: Run script in debug mode

```bash
bash -x scripts/local-run.sh
```

This prints each command before execution.

Very useful for debugging.

---

## Tip 4: Check env vars

```bash
echo "$SLACK_WEBHOOK_URL"
```

Or safer:

```bash
printenv SLACK_WEBHOOK_URL
```

---

## Tip 5: Check current folder

Many script errors happen because you run from the wrong folder.

Check:

```bash
pwd
ls
```

Expected:

```text
You should be inside slack-integration project root.
```

---

# 15. One DSA topic — Sorting basics

## Simple explanation

Sorting means arranging values in order.

Example:

```text
Before:  [5, 2, 9, 1]
After:   [1, 2, 5, 9]
```

Sorting can be:

```text
ascending  → small to large
descending → large to small
```

Real backend examples:

```text
sort logs by timestamp
sort pipeline runs by duration
sort failed jobs by severity
sort users by creation time
sort API responses by name
```

Python sorting:

```python
numbers = [5, 2, 9, 1]
numbers.sort()
print(numbers)
```

Go sorting:

```go
sort.Ints(numbers)
```

Go requires importing the `sort` package.

---

## Bubble sort idea

Bubble sort compares nearby values and swaps them if they are in the wrong order.

Example:

```text
[5, 2, 9, 1]

Compare 5 and 2 → swap
[2, 5, 9, 1]

Compare 5 and 9 → okay
[2, 5, 9, 1]

Compare 9 and 1 → swap
[2, 5, 1, 9]
```

After multiple rounds:

```text
[1, 2, 5, 9]
```

Simple mental model:

```text
Large values slowly move to the end.
```

---

# 16. One Go DSA practice problem

## Problem: Sort pipeline durations

You are given pipeline execution times in seconds:

```text
[45, 12, 90, 30, 10]
```

Sort them in ascending order.

Expected output:

```text
[10 12 30 45 90]
```

---

## Go solution using built-in sort

```go
package main

import (
	"fmt"
	"sort"
)

func main() {
	durations := []int{45, 12, 90, 30, 10}

	sort.Ints(durations)

	fmt.Println(durations)
}
```

Output:

```text
[10 12 30 45 90]
```

Python equivalent:

```python
durations = [45, 12, 90, 30, 10]
durations.sort()
print(durations)
```

---

## Manual bubble sort in Go

```go
package main

import "fmt"

func bubbleSort(numbers []int) {
	n := len(numbers)

	for i := 0; i < n; i++ {
		for j := 0; j < n-i-1; j++ {
			if numbers[j] > numbers[j+1] {
				numbers[j], numbers[j+1] = numbers[j+1], numbers[j]
			}
		}
	}
}

func main() {
	durations := []int{45, 12, 90, 30, 10}

	bubbleSort(durations)

	fmt.Println(durations)
}
```

Important Go syntax:

```go
numbers[j], numbers[j+1] = numbers[j+1], numbers[j]
```

This swaps two values.

Python equivalent:

```python
numbers[j], numbers[j + 1] = numbers[j + 1], numbers[j]
```

This part is very similar.

---

# 17. One module-based practice task — File processor / log filter

## Goal

Create a small log filter script.

Input file:

```text
logs/app.log
```

Example content:

```text
INFO pipeline started
INFO running unit tests
ERROR unit tests failed
INFO collecting trace
ERROR slack webhook failed
INFO process completed
```

Your script should print only error lines.

---

## Create sample log file

Run:

```bash
mkdir -p logs
cat > logs/app.log <<EOF
INFO pipeline started
INFO running unit tests
ERROR unit tests failed
INFO collecting trace
ERROR slack webhook failed
INFO process completed
EOF
```

Beginner meaning:

```text
mkdir -p logs       → create logs folder if missing
cat > file <<EOF    → write multiple lines into a file
EOF                 → end of file content
```

---

## Create script

File:

```text
scripts/log-filter.sh
```

Code:

```bash
#!/usr/bin/env bash
set -euo pipefail

LOG_FILE="${1:-logs/app.log}"
KEYWORD="${2:-ERROR}"

if [[ ! -f "$LOG_FILE" ]]; then
  echo "Error: log file not found: $LOG_FILE"
  exit 1
fi

echo "Filtering log file: $LOG_FILE"
echo "Keyword: $KEYWORD"
echo "--------------------------------"

grep "$KEYWORD" "$LOG_FILE" || true
```

Run:

```bash
chmod +x scripts/log-filter.sh
./scripts/log-filter.sh logs/app.log ERROR
```

Expected output:

```text
Filtering log file: logs/app.log
Keyword: ERROR
--------------------------------
ERROR unit tests failed
ERROR slack webhook failed
```

---

## Why `|| true` is used here

This line:

```bash
grep "$KEYWORD" "$LOG_FILE" || true
```

Means:

```text
If grep finds nothing, do not fail the whole script.
```

Because with:

```bash
set -e
```

`grep` returns non-zero when no match is found.

Sometimes no match is acceptable.

Example:

```bash
./scripts/log-filter.sh logs/app.log WARNING
```

No warning logs may exist. That should not always be a script failure.

---

## Python equivalent

```python
log_file = "logs/app.log"
keyword = "ERROR"

with open(log_file) as f:
    for line in f:
        if keyword in line:
            print(line.strip())
```

Shell is shorter for file and text filtering.

Python is better when logic becomes complex.

---

# 18. Revision checkpoint

You should now be able to answer these:

1. What is a shell script?
2. Why do backend/DevOps teams use scripts?
3. What does this mean?

```bash
#!/usr/bin/env bash
```

4. What does this mean?

```bash
set -euo pipefail
```

5. What is the difference between argument and environment variable?
6. What is `$1`?
7. What is `$?`?
8. What does exit code `0` mean?
9. Why should secrets like Slack webhook URL come from env vars?
10. How can shell scripts help in Tekton?
11. Why should you quote shell variables?
12. How can a script wrap a long Go CLI command?
13. How do you run all Go tests?
14. What is sorting?
15. How does Go sorting compare with Python sorting?

---

# 19. Homework

## Homework 1: Improve `local-run.sh`

Add support for:

```bash
SENDER
USER_EMAIL
COMMIT_ID
COMMIT_MESSAGE
```

Example:

```bash
export SENDER="radhe"
export USER_EMAIL="radhe@example.com"
export COMMIT_ID="abc123"
export COMMIT_MESSAGE="fix slack notification failure"
./scripts/local-run.sh pr failure failed
```

---

## Homework 2: Create `run-pr-failure.sh`

Create a script specifically for PR failure simulation.

Expected usage:

```bash
./scripts/run-pr-failure.sh
```

Internally it should call:

```bash
./scripts/local-run.sh pr failure failed
```

---

## Homework 3: Extend `log-filter.sh`

Add support for:

```bash
INFO
ERROR
FAILED
SUCCESS
```

Example:

```bash
./scripts/log-filter.sh logs/app.log ERROR
./scripts/log-filter.sh logs/app.log INFO
```

---

## Homework 4: Go DSA practice

Write a Go program that:

1. Takes this slice:

```go
scores := []int{80, 20, 50, 90, 10}
```

2. Sorts it.
3. Prints the smallest value.
4. Prints the largest value.

Expected output:

```text
Sorted: [10 20 50 80 90]
Smallest: 10
Largest: 90
```

---

## Final simple mental model

```text
Go code builds the application.
Shell scripts make it easy to run, test, debug, and automate the application.
Tekton uses the same idea at CI/CD level.
```

For your project:

```text
Day 1-6: Build and test Go Slack notifier
Day 7: Automate local workflow with shell scripts
Next level: Connect these scripts into Tekton tasks
```
