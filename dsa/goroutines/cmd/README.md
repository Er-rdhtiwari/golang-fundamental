# Goroutines Worker Pool Example

This example shows how to process notification jobs concurrently using goroutines, channels, and a `sync.WaitGroup`.

If you are coming from Python, you can think of it like this:

```text
Go goroutine        ~= lightweight thread
Go channel          ~= typed queue
sync.WaitGroup      ~= wait until all worker threads finish
```

## Files

```text
main.go
```

The program creates multiple workers. Each worker waits for jobs from a shared channel. The main function sends notification jobs into the channel, closes the channel, and waits until all workers finish.

## Main Concepts

### NotificationJob

```go
type NotificationJob struct {
	ID      int
	UserID  string
	Message string
}
```

This struct represents one notification task.

Python comparison:

```python
job = {
    "id": 1,
    "user_id": "U001",
    "message": "Welcome to the workspace",
}
```

### Channel

```go
jobs := make(chan NotificationJob)
```

This creates a channel that can send and receive only `NotificationJob` values.

You can think of it as a typed queue:

```text
main goroutine -> jobs channel -> worker goroutines
```

### Worker

```go
func Worker(id int, jobs <-chan NotificationJob, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		ProcessNotification(id, job)
	}
}
```

Each worker receives jobs from the channel.

This line keeps reading jobs until the channel is closed:

```go
for job := range jobs
```

The parameter:

```go
jobs <-chan NotificationJob
```

means the worker can only receive from the channel. It cannot send jobs into it.

### WaitGroup

```go
var wg sync.WaitGroup
```

The WaitGroup keeps the `main` function alive until all workers finish.

Before starting each worker:

```go
wg.Add(1)
```

When the worker exits:

```go
defer wg.Done()
```

At the end, `main` waits:

```go
wg.Wait()
```

## Execution Flow

### 1. Start the program

Execution begins in:

```go
func main()
```

### 2. Create the jobs channel

```go
jobs := make(chan NotificationJob)
```

This is where notification jobs will be sent.

### 3. Start workers

```go
workerCount := 3

for i := 1; i <= workerCount; i++ {
	wg.Add(1)
	go Worker(i, jobs, &wg)
}
```

This starts 3 background goroutines:

```text
Worker 1
Worker 2
Worker 3
```

Each worker waits for jobs from the same `jobs` channel.

### 4. Create notification jobs

```go
notificationJobs := []NotificationJob{
	{ID: 1, UserID: "U001", Message: "Welcome to the workspace"},
	{ID: 2, UserID: "U002", Message: "Your report is ready"},
	{ID: 3, UserID: "U003", Message: "You have a new mention"},
	{ID: 4, UserID: "U004", Message: "Daily summary is available"},
	{ID: 5, UserID: "U005", Message: "Password changed successfully"},
}
```

This is a slice of notification jobs.

Python comparison:

```python
notification_jobs = [
    {"id": 1, "user_id": "U001", "message": "Welcome to the workspace"},
    {"id": 2, "user_id": "U002", "message": "Your report is ready"},
]
```

### 5. Send jobs to workers

```go
for _, job := range notificationJobs {
	jobs <- job
}
```

This sends each job into the channel.

Any available worker can receive the next job. The worker order is not guaranteed.

Example:

```text
Worker 1 may process job 1
Worker 3 may process job 2
Worker 2 may process job 3
```

Another run may produce a different order. That is normal in concurrent programs.

### 6. Close the channel

```go
close(jobs)
```

This tells workers:

```text
No more jobs will be sent.
```

After the channel is closed and all jobs are consumed, this loop ends inside every worker:

```go
for job := range jobs
```

### 7. Wait for workers to finish

```go
wg.Wait()
```

The main goroutine blocks here until all workers call:

```go
wg.Done()
```

### 8. Finish

```go
fmt.Println("All notifications processed")
```

This prints only after all workers have exited.

## Full Flow Diagram

```text
main()
  |
  |-- create jobs channel
  |
  |-- start Worker 1
  |-- start Worker 2
  |-- start Worker 3
  |
  |-- create notification jobs
  |
  |-- send jobs into channel
  |
  |-- close channel
  |
  |-- wait for workers
  |
  |-- print "All notifications processed"


Worker 1 / Worker 2 / Worker 3
  |
  |-- wait for job from channel
  |-- process job
  |-- wait for next job
  |-- exit when channel is closed
```

## Run The Example

From the repository root:

```bash
go run ./dsa/goroutines/cmd
```

Example output:

```text
Worker 3 semding notifcation 1 to user U001: Welcome to the workspace
Worker 1 semding notifcation 2 to user U002: Your report is ready
Worker 2 semding notifcation 3 to user U003: You have a new mention
Worker 3 semding notifcation 4 to user U004: Daily summary is available
Worker 1 semding notifcation 5 to user U005: Password changed successfully
All notifications processed
```

The worker numbers and job order may be different each time.

## Important Notes

Do not forget this:

```go
close(jobs)
```

Without closing the channel, workers keep waiting for more jobs forever.

Do not forget this:

```go
wg.Wait()
```

Without waiting, the main function may exit before workers finish.

Use `wg.Add(1)` before starting each goroutine:

```go
wg.Add(1)
go Worker(i, jobs, &wg)
```

Use `defer wg.Done()` inside the worker:

```go
defer wg.Done()
```

This guarantees that the worker marks itself as done when it exits.
