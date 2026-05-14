# Day 12 — Go Concurrency Basics for a `slack-integration` Project

Today we will keep concurrency **simple and practical**.

Think of your future `slack-integration` project:

You may later need to:

* process many Slack notifications
* parse many log lines
* send messages to different users
* retry failed jobs
* process events from a queue

Concurrency can help there.

But beginner rule:

> Do not add concurrency just because Go makes it easy.
> Add it only when there are many independent tasks that can safely run at the same time.

---

# 1. Day 12 Learning Goals

By the end of Day 12, you should understand:

* what a goroutine is
* how it differs from a normal function call
* what channels are
* how `sync.WaitGroup` helps wait for goroutines
* how a simple worker pool works
* where concurrency may fit in a Slack-style project
* where concurrency is unnecessary
* basic race condition awareness
* sliding window DSA basics
* how to build a small notification worker pool

Python comparison for today:

| Go concept     | Rough Python equivalent              |
| -------------- | ------------------------------------ |
| goroutine      | thread / asyncio task                |
| channel        | `queue.Queue` / `asyncio.Queue`      |
| WaitGroup      | `thread.join()` / `asyncio.gather()` |
| worker pool    | fixed group of worker threads/tasks  |
| race condition | shared mutable data bug              |

---

# 2. Quick Revision of Days 1 to 11

Assuming your first 11 days covered Go basics, here is the quick revision.

You have likely learned:

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello")
}
```

Important Go basics:

```go
var name string = "Slack Bot"
age := 12
```

Python comparison:

```python
name = "Slack Bot"
age = 12
```

In Go, types matter more. Go is statically typed.

Functions:

```go
func add(a int, b int) int {
    return a + b
}
```

Python:

```python
def add(a, b):
    return a + b
```

Structs:

```go
type Notification struct {
    UserID  string
    Message string
}
```

Python equivalent:

```python
@dataclass
class Notification:
    user_id: str
    message: str
```

Slices:

```go
messages := []string{"hello", "hi", "welcome"}
```

Python:

```python
messages = ["hello", "hi", "welcome"]
```

Loops:

```go
for _, msg := range messages {
    fmt.Println(msg)
}
```

Python:

```python
for msg in messages:
    print(msg)
```

Error handling:

```go
if err != nil {
    return err
}
```

Python:

```python
try:
    ...
except Exception as err:
    ...
```

Go convention shift:

> Go prefers explicit error checks instead of exceptions.

Today we add concurrency on top of these basics.

---

# 3. Explain Concurrency in Very Simple Language

Concurrency means:

> Doing multiple tasks by letting them make progress independently.

Not always literally at the exact same time.

Simple example:

You are making tea.

Sequential way:

1. boil water
2. wait
3. add tea leaves
4. wait
5. get cup
6. pour tea

Concurrent way:

1. start boiling water
2. while water boils, get cup
3. while tea brews, clean counter

You are managing multiple tasks efficiently.

In a Slack-style project, imagine receiving 100 notification jobs.

Without concurrency:

```text
process job 1
process job 2
process job 3
...
process job 100
```

With concurrency:

```text
worker 1 processes job 1
worker 2 processes job 2
worker 3 processes job 3
...
```

This can be useful when jobs are independent.

---

# 4. Goroutine vs Normal Function

A normal function runs immediately and blocks until finished.

```go
sendMessage()
fmt.Println("Done")
```

This means:

```text
sendMessage finishes first
then Done prints
```

A goroutine runs independently.

```go
go sendMessage()
fmt.Println("Done")
```

This means:

```text
sendMessage starts in the background
Done may print before sendMessage finishes
```

The keyword is:

```go
go
```

That is the syntax change.

Example:

```go
go myFunction()
```

Python comparison:

A goroutine is somewhat like:

```python
threading.Thread(target=my_function).start()
```

or an asyncio task:

```python
asyncio.create_task(my_function())
```

But Go goroutines are much lighter and built deeply into the language.

Important beginner point:

> Starting a goroutine does not mean your program will automatically wait for it.

This code may not print from the goroutine:

```go
package main

import "fmt"

func sayHello() {
    fmt.Println("Hello from goroutine")
}

func main() {
    go sayHello()
    fmt.Println("Main finished")
}
```

Possible output:

```text
Main finished
```

Why?

Because `main` ended before the goroutine got time to run.

That is why we need `WaitGroup`.

---

# 5. Channels in Beginner-Friendly Terms

A channel is a safe pipe between goroutines.

One goroutine can send data.

Another goroutine can receive data.

Visual:

```text
goroutine A ---> channel ---> goroutine B
```

Create a channel:

```go
ch := make(chan string)
```

Send data:

```go
ch <- "hello"
```

Receive data:

```go
msg := <-ch
```

Python comparison:

Go channel is similar to:

```python
queue.Queue()
```

Example in Python:

```python
q.put("hello")
msg = q.get()
```

In Go:

```go
ch <- "hello"
msg := <-ch
```

Important syntax:

| Action         | Go                  |
| -------------- | ------------------- |
| create channel | `make(chan string)` |
| send           | `ch <- value`       |
| receive        | `value := <-ch`     |
| close          | `close(ch)`         |

Simple example:

```go
package main

import "fmt"

func main() {
    ch := make(chan string)

    go func() {
        ch <- "Hello from goroutine"
    }()

    msg := <-ch
    fmt.Println(msg)
}
```

Output:

```text
Hello from goroutine
```

What happened?

1. `main` created a channel
2. goroutine sent a message into the channel
3. `main` waited to receive it
4. message printed

Channels are useful when goroutines need to communicate.

Beginner rule:

> Use channels for passing data between goroutines, not for showing off.

---

# 6. WaitGroup Simply

A `WaitGroup` helps `main` wait for goroutines to finish.

Think of it like a counter.

You say:

```text
I am starting 3 goroutines.
Please wait until all 3 say they are done.
```

Go syntax:

```go
var wg sync.WaitGroup
```

Before starting a goroutine:

```go
wg.Add(1)
```

Inside the goroutine when finished:

```go
defer wg.Done()
```

Wait for all:

```go
wg.Wait()
```

Full example:

```go
package main

import (
    "fmt"
    "sync"
)

func printMessage(message string, wg *sync.WaitGroup) {
    defer wg.Done()

    fmt.Println(message)
}

func main() {
    var wg sync.WaitGroup

    wg.Add(1)
    go printMessage("Hello from goroutine", &wg)

    wg.Wait()

    fmt.Println("Main finished")
}
```

Possible output:

```text
Hello from goroutine
Main finished
```

Python comparison:

This is similar to:

```python
thread = threading.Thread(target=print_message)
thread.start()
thread.join()
```

For multiple goroutines, `WaitGroup` is like calling `.join()` on many threads.

Important Go convention:

> Pass `*sync.WaitGroup` as a pointer.

Because the goroutine needs to update the same WaitGroup.

---

# 7. Tiny Worker Pool Example

A worker pool means:

> Start a fixed number of workers and give them jobs through a channel.

Why fixed?

Because starting unlimited goroutines can overload your program.

Imagine 1,000 Slack notifications.

Bad beginner idea:

```go
for _, job := range jobs {
    go process(job)
}
```

This may be okay for tiny demos, but risky in real systems.

Better idea:

```text
Create 3 workers.
Send 100 jobs.
Workers process jobs one by one.
```

Visual:

```text
jobs channel:
[job1, job2, job3, job4, job5]

worker 1 -> job1
worker 2 -> job2
worker 3 -> job3
worker 1 -> job4
worker 2 -> job5
```

---

# 8. Where Concurrency May Fit in This Project and Where It May Not

In a `slack-integration` style project, concurrency may help with:

```text
Receiving many events
Processing notification jobs
Parsing large log files
Sending messages to multiple users
Retrying independent failed jobs
Calling external APIs carefully
```

Example:

```text
Slack sends 500 events.
Your app puts them into a job queue.
Workers process the events.
```

Concurrency may not help with:

```text
Simple config loading
Small validation functions
Printing one report
Tiny scripts
Code that must happen in exact order
Beginner code that is still changing often
```

Beginner rule:

> First make it correct and simple.
> Then add concurrency only where it clearly helps.

Race condition, high-level only:

A race condition happens when two goroutines touch the same data at the same time and at least one changes it.

Example problem:

```go
counter++
```

If many goroutines do this together, the result may be wrong.

Beginner-safe approach:

* avoid shared mutable data
* pass data through channels
* keep worker jobs independent
* use `sync.Mutex` later when needed
* use `go run -race main.go` to detect many race issues

---

# 9. Pseudocode First

## Toy goroutine pseudocode

```text
function sayHello:
    print hello

main:
    start sayHello as goroutine
    wait for it to finish
    print done
```

## Channel pseudocode

```text
create message channel

start goroutine:
    send "hello" into channel

main:
    receive message from channel
    print message
```

## Worker pool pseudocode

```text
create jobs channel
create WaitGroup

start 3 workers:
    each worker keeps reading from jobs channel
    process each job

send jobs into jobs channel
close jobs channel

wait until all workers finish
print all done
```

## Notification worker pool pseudocode

```text
NotificationJob:
    ID
    UserID
    Message

worker:
    read notification job from channel
    process notification
    print result

main:
    create jobs channel
    start workers
    send notification jobs
    close channel
    wait for workers
```

---

# 10. Real Go Code Examples

## Example 1: Normal function

```go
package main

import "fmt"

func sayHello() {
    fmt.Println("Hello")
}

func main() {
    sayHello()
    fmt.Println("Main finished")
}
```

Expected output:

```text
Hello
Main finished
```

This is sequential.

---

## Example 2: Goroutine with WaitGroup

```go
package main

import (
    "fmt"
    "sync"
)

func sayHello(wg *sync.WaitGroup) {
    defer wg.Done()

    fmt.Println("Hello from goroutine")
}

func main() {
    var wg sync.WaitGroup

    wg.Add(1)
    go sayHello(&wg)

    wg.Wait()

    fmt.Println("Main finished")
}
```

Expected output:

```text
Hello from goroutine
Main finished
```

Important syntax:

```go
go sayHello(&wg)
```

means:

> Start `sayHello` as a goroutine.

```go
defer wg.Done()
```

means:

> Mark this goroutine done when the function exits.

Python comparison:

Go:

```go
wg.Add(1)
go sayHello(&wg)
wg.Wait()
```

Python threading:

```python
thread = threading.Thread(target=say_hello)
thread.start()
thread.join()
```

---

## Example 3: Channel basics

```go
package main

import "fmt"

func main() {
    messages := make(chan string)

    go func() {
        messages <- "New Slack message received"
    }()

    msg := <-messages

    fmt.Println(msg)
}
```

Expected output:

```text
New Slack message received
```

Python comparison:

Go:

```go
messages := make(chan string)
messages <- "hello"
msg := <-messages
```

Python:

```python
from queue import Queue

messages = Queue()
messages.put("hello")
msg = messages.get()
```

---

## Example 4: Tiny worker pool

```go
package main

import (
    "fmt"
    "sync"
)

func worker(id int, jobs <-chan int, wg *sync.WaitGroup) {
    defer wg.Done()

    for job := range jobs {
        fmt.Printf("Worker %d processing job %d\n", id, job)
    }
}

func main() {
    jobs := make(chan int)

    var wg sync.WaitGroup

    workerCount := 3

    for i := 1; i <= workerCount; i++ {
        wg.Add(1)
        go worker(i, jobs, &wg)
    }

    for job := 1; job <= 5; job++ {
        jobs <- job
    }

    close(jobs)

    wg.Wait()

    fmt.Println("All jobs processed")
}
```

Possible output:

```text
Worker 1 processing job 1
Worker 2 processing job 2
Worker 3 processing job 3
Worker 1 processing job 4
Worker 2 processing job 5
All jobs processed
```

The exact worker order may change.

That is normal.

Concurrency does not guarantee order.

Important syntax:

```go
jobs <-chan int
```

This means the worker can only receive from the channel.

It cannot send into it.

This is a nice Go convention for clarity.

---

## Example 5: Notification worker pool

```go
package main

import (
    "fmt"
    "sync"
)

type NotificationJob struct {
    ID      int
    UserID  string
    Message string
}

func processNotification(workerID int, job NotificationJob) {
    fmt.Printf(
        "Worker %d sending notification %d to user %s: %s\n",
        workerID,
        job.ID,
        job.UserID,
        job.Message,
    )
}

func worker(id int, jobs <-chan NotificationJob, wg *sync.WaitGroup) {
    defer wg.Done()

    for job := range jobs {
        processNotification(id, job)
    }
}

func main() {
    jobs := make(chan NotificationJob)

    var wg sync.WaitGroup

    workerCount := 3

    for i := 1; i <= workerCount; i++ {
        wg.Add(1)
        go worker(i, jobs, &wg)
    }

    notificationJobs := []NotificationJob{
        {ID: 1, UserID: "U001", Message: "Welcome to the workspace"},
        {ID: 2, UserID: "U002", Message: "Your report is ready"},
        {ID: 3, UserID: "U003", Message: "You have a new mention"},
        {ID: 4, UserID: "U004", Message: "Daily summary is available"},
        {ID: 5, UserID: "U005", Message: "Password changed successfully"},
    }

    for _, job := range notificationJobs {
        jobs <- job
    }

    close(jobs)

    wg.Wait()

    fmt.Println("All notifications processed")
}
```

Possible output:

```text
Worker 1 sending notification 1 to user U001: Welcome to the workspace
Worker 2 sending notification 2 to user U002: Your report is ready
Worker 3 sending notification 3 to user U003: You have a new mention
Worker 1 sending notification 4 to user U004: Daily summary is available
Worker 2 sending notification 5 to user U005: Password changed successfully
All notifications processed
```

Again, order may differ.

That is normal.

---

# 11. Hands-On Tasks

## Task 1: Run a normal function

Create:

```go
func printTask(name string)
```

Call it normally:

```go
printTask("parse logs")
printTask("send notification")
```

Observe that output order is predictable.

---

## Task 2: Run the same function as goroutines

Use:

```go
go printTask("parse logs")
go printTask("send notification")
```

Then add `WaitGroup`.

Observe that output order may change.

---

## Task 3: Create a channel

Create a channel:

```go
messages := make(chan string)
```

Start a goroutine that sends:

```go
messages <- "notification ready"
```

Receive in `main`.

---

## Task 4: Build a small worker pool

Use:

```go
workerCount := 2
jobs := make(chan string)
```

Send jobs:

```go
"send welcome message"
"send report alert"
"send mention alert"
```

Have workers print the job.

---

# 12. Expected Output

For the notification worker pool, output may look like this:

```text
Worker 1 sending notification 1 to user U001: Welcome to the workspace
Worker 2 sending notification 2 to user U002: Your report is ready
Worker 3 sending notification 3 to user U003: You have a new mention
Worker 1 sending notification 4 to user U004: Daily summary is available
Worker 2 sending notification 5 to user U005: Password changed successfully
All notifications processed
```

But this is also valid:

```text
Worker 3 sending notification 1 to user U001: Welcome to the workspace
Worker 1 sending notification 2 to user U002: Your report is ready
Worker 2 sending notification 3 to user U003: You have a new mention
Worker 3 sending notification 4 to user U004: Daily summary is available
Worker 1 sending notification 5 to user U005: Password changed successfully
All notifications processed
```

Important lesson:

> Concurrent output order is often not guaranteed.

---

# 13. Common Mistakes

## Mistake 1: Forgetting `wg.Wait()`

Wrong:

```go
go doWork()
fmt.Println("Done")
```

The program may exit too early.

Better:

```go
wg.Add(1)
go doWork(&wg)
wg.Wait()
```

---

## Mistake 2: Forgetting `wg.Done()`

Wrong:

```go
func worker(wg *sync.WaitGroup) {
    fmt.Println("working")
}
```

This can make your program wait forever.

Better:

```go
func worker(wg *sync.WaitGroup) {
    defer wg.Done()
    fmt.Println("working")
}
```

---

## Mistake 3: Closing a channel too early

Wrong:

```go
close(jobs)
jobs <- "new job"
```

This causes panic.

You cannot send into a closed channel.

---

## Mistake 4: Never closing a job channel

If workers use:

```go
for job := range jobs {
    ...
}
```

Then the channel must eventually be closed.

Otherwise workers keep waiting.

---

## Mistake 5: Assuming goroutines run in order

This is wrong thinking:

```go
go task1()
go task2()
go task3()
```

You should not assume `task1` finishes before `task2`.

---

## Mistake 6: Sharing data casually

Risky:

```go
count := 0

go func() {
    count++
}()

go func() {
    count++
}()
```

This can cause a race condition.

Beginner-safe rule:

> Avoid multiple goroutines changing the same variable.

---

# 14. Debugging Tips

Use simple prints first:

```go
fmt.Println("worker started")
fmt.Println("job received:", job)
fmt.Println("worker finished")
```

Use worker IDs:

```go
fmt.Printf("Worker %d processing job %d\n", id, job.ID)
```

Run with race detector:

```bash
go run -race main.go
```

This helps catch shared data problems.

Check these questions:

```text
Did I call wg.Add(1) before starting the goroutine?
Did I call defer wg.Done() inside the goroutine?
Did I close the jobs channel after sending all jobs?
Am I sending to a closed channel?
Am I waiting forever because nobody sends data?
Am I reading forever because the channel was never closed?
```

Beginner debugging rule:

> Add concurrency slowly. Test after each small step.

---

# 15. One DSA Topic: Sliding Window

Sliding window is a technique for arrays and strings.

It is useful when you need to look at a continuous part of data.

Example:

```text
nums = [2, 1, 5, 1, 3, 2]
k = 3
```

Find maximum sum of any 3 continuous numbers.

Windows of size 3:

```text
[2, 1, 5] = 8
[1, 5, 1] = 7
[5, 1, 3] = 9
[1, 3, 2] = 6
```

Answer:

```text
9
```

Naive way:

```text
Calculate every window from scratch.
```

Sliding window way:

```text
Take first window sum.
Move right:
    subtract number leaving window
    add number entering window
```

Visual:

```text
[2, 1, 5] 1 3 2
sum = 8

2 [1, 5, 1] 3 2
new sum = 8 - 2 + 1 = 7

2 1 [5, 1, 3] 2
new sum = 7 - 1 + 3 = 9
```

Python comparison:

Python:

```python
window_sum = sum(nums[:k])
```

Go:

```go
windowSum := 0
for i := 0; i < k; i++ {
    windowSum += nums[i]
}
```

Go does not have built-in `sum()` for slices.

---

# 16. One Go DSA Problem

## Problem: Maximum Sum of Subarray of Size K

Given an integer slice `nums` and an integer `k`, find the maximum sum of any continuous subarray of size `k`.

Example:

```text
nums = [2, 1, 5, 1, 3, 2]
k = 3
```

Answer:

```text
9
```

Because:

```text
[5, 1, 3] = 9
```

## Go Solution

```go
package main

import "fmt"

func maxSumSubarray(nums []int, k int) int {
    if len(nums) < k || k <= 0 {
        return 0
    }

    windowSum := 0

    for i := 0; i < k; i++ {
        windowSum += nums[i]
    }

    maxSum := windowSum

    for right := k; right < len(nums); right++ {
        left := right - k

        windowSum = windowSum - nums[left] + nums[right]

        if windowSum > maxSum {
            maxSum = windowSum
        }
    }

    return maxSum
}

func main() {
    nums := []int{2, 1, 5, 1, 3, 2}
    k := 3

    result := maxSumSubarray(nums, k)

    fmt.Println(result)
}
```

Expected output:

```text
9
```

Key syntax comparison:

Python:

```python
for right in range(k, len(nums)):
    left = right - k
```

Go:

```go
for right := k; right < len(nums); right++ {
    left := right - k
}
```

Python:

```python
if window_sum > max_sum:
    max_sum = window_sum
```

Go:

```go
if windowSum > maxSum {
    maxSum = windowSum
}
```

Go style convention:

* `windowSum`, not `window_sum`
* `maxSum`, not `max_sum`
* braces `{}` are required
* no parentheses needed around `if` condition

---

# 17. Module-Based Practice Task

## Task: Build a Notification Worker Pool

Create this file:

```text
notification_worker_pool/main.go
```

Goal:

Process notification jobs using 3 workers.

Each notification should have:

```go
type NotificationJob struct {
    ID      int
    UserID  string
    Message string
}
```

Your program should:

1. create a jobs channel
2. start 3 workers
3. send 6 notification jobs
4. close the jobs channel
5. wait for all workers to finish
6. print `"All notification jobs completed"`

Starter code:

```go
package main

import (
    "fmt"
    "sync"
)

type NotificationJob struct {
    ID      int
    UserID  string
    Message string
}

func worker(id int, jobs <-chan NotificationJob, wg *sync.WaitGroup) {
    defer wg.Done()

    for job := range jobs {
        fmt.Printf("Worker %d processed notification %d for user %s: %s\n",
            id,
            job.ID,
            job.UserID,
            job.Message,
        )
    }
}

func main() {
    jobs := make(chan NotificationJob)

    var wg sync.WaitGroup

    workerCount := 3

    for i := 1; i <= workerCount; i++ {
        wg.Add(1)
        go worker(i, jobs, &wg)
    }

    notifications := []NotificationJob{
        {ID: 1, UserID: "U001", Message: "Welcome!"},
        {ID: 2, UserID: "U002", Message: "Your daily report is ready."},
        {ID: 3, UserID: "U003", Message: "You were mentioned in a channel."},
        {ID: 4, UserID: "U004", Message: "Reminder: team meeting at 5 PM."},
        {ID: 5, UserID: "U005", Message: "Security alert: new login detected."},
        {ID: 6, UserID: "U006", Message: "Your export has completed."},
    }

    for _, notification := range notifications {
        jobs <- notification
    }

    close(jobs)

    wg.Wait()

    fmt.Println("All notification jobs completed")
}
```

Run:

```bash
go run main.go
```

Possible output:

```text
Worker 1 processed notification 1 for user U001: Welcome!
Worker 2 processed notification 2 for user U002: Your daily report is ready.
Worker 3 processed notification 3 for user U003: You were mentioned in a channel.
Worker 1 processed notification 4 for user U004: Reminder: team meeting at 5 PM.
Worker 2 processed notification 5 for user U005: Security alert: new login detected.
Worker 3 processed notification 6 for user U006: Your export has completed.
All notification jobs completed
```

Order may differ.

That is okay.

---

# 18. Revision Checkpoint

You should now be able to answer these:

1. What does `go someFunc()` do?
2. Why can a goroutine finish after `main` exits?
3. What does `sync.WaitGroup` solve?
4. What does `wg.Add(1)` mean?
5. Why do we use `defer wg.Done()`?
6. What is a channel?
7. What does `jobs <- job` mean?
8. What does `job := <-jobs` mean?
9. What does `for job := range jobs` do?
10. Why do we close the jobs channel?
11. Why is worker output order unpredictable?
12. When should we avoid concurrency?
13. What is a race condition at a high level?
14. What is a sliding window?
15. Why is sliding window better than recalculating every window?

Good beginner answers:

```text
A goroutine runs a function independently.
A WaitGroup waits for goroutines to finish.
A channel passes data between goroutines.
A worker pool limits how many workers run at once.
Sliding window reuses previous work instead of recalculating everything.
```

---

# 19. Homework

## Part A: Goroutine Practice

Write a program with three functions:

```go
parseLogs()
sendNotification()
saveAuditLog()
```

Run them as goroutines.

Use `WaitGroup` so `main` waits for all three.

Expected style:

```text
Parsing logs
Sending notification
Saving audit log
All tasks finished
```

Output order may change.

---

## Part B: Channel Practice

Create a channel of strings.

One goroutine should send:

```text
"event received"
"event validated"
"event processed"
```

Main should receive and print all messages.

Hint:

Use a loop.

---

## Part C: Worker Pool Practice

Create 2 workers.

Send these jobs:

```text
"send welcome notification"
"send reminder notification"
"send alert notification"
"send summary notification"
```

Each worker should print:

```text
Worker 1 processing send welcome notification
```

or similar.

---

## Part D: DSA Homework

Solve this:

```text
nums = [1, 4, 2, 10, 23, 3, 1, 0, 20]
k = 4
```

Find the maximum sum of any subarray of size `4`.

Expected answer:

```text
39
```

Because:

```text
[4, 2, 10, 23] = 39
```

---

## Final Day 12 Mental Model

Keep this in your head:

```text
goroutine = lightweight concurrent task
channel = pipe for passing data
WaitGroup = wait until tasks finish
worker pool = fixed number of workers processing many jobs
```

And the most important beginner rule:

> Make it correct first.
> Make it concurrent only when there is a real reason.
