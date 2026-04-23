# Repository Structure Runbook

This runbook defines a practical directory structure for this repository as you use it for:

- daily Go learning
- daily DSA practice
- gradual growth into a real Go project

The goal is not just to "store files." The goal is to keep learning notes, problem solving, and project code separate so the repository stays clean after weeks and months of practice.

## Why this runbook exists

If you practice one topic every day, the repository grows quickly.

A flat structure becomes hard to manage:

- notes mix with code
- DSA solutions mix with project files
- it becomes hard to revise by topic
- it becomes hard to test or reuse code

This runbook gives one stable structure that supports both learning and production-style habits.

## Recommended directory tree

```text
golang-fundamental/
├── README.md
├── LICENSE
├── go.mod
│
├── docs/
│   └── repo-structure-runbook.md
│
├── notes/
│   ├── day-001-cli-basics.md
│   ├── day-002-structs-methods.md
│   └── day-003-json-http.md
│
├── dsa/
│   ├── arrays/
│   │   ├── notes.md
│   │   ├── two-sum/
│   │   │   ├── main.go
│   │   │   ├── main_test.go
│   │   │   └── README.md
│   │   └── best-time-to-buy-sell-stock/
│   ├── strings/
│   ├── maps/
│   ├── stack/
│   ├── queue/
│   ├── linked-list/
│   ├── recursion/
│   ├── tree/
│   ├── graph/
│   └── dynamic-programming/
│
├── cmd/
│   └── app/
│       └── main.go
│
├── internal/
│   ├── cli/
│   ├── config/
│   ├── model/
│   ├── parser/
│   ├── router/
│   └── output/
│
├── pkg/
│   └── utils/
│
└── scripts/
```

## Thought process behind this structure

This structure is based on one simple idea:

```text
Learning notes != DSA practice != application code
```

These are related, but they are not the same kind of content.

### 1. `notes/` is for day-wise learning

Use `notes/` for your daily study journal:

- what you learned
- examples you want to remember
- diagrams
- mistakes
- revision points

This is where your current `Day 1`, `Day 2`, `Day 3` style material belongs.

Why keep this separate:

- notes are chronological
- notes are meant for revision
- notes are not meant to be imported, tested, or run

That means day-based naming works well here.

Example:

```text
notes/day-001-cli-basics.md
notes/day-002-structs-methods.md
notes/day-003-json-http.md
```

### 2. `dsa/` is for topic-wise practice

DSA should be grouped by topic, not by day.

Reason:

- interview revision usually happens by topic
- patterns become easier to see inside one topic
- you can compare multiple solutions to similar problems
- daily folders become noisy after enough practice

This is why `dsa/arrays/`, `dsa/strings/`, `dsa/tree/`, and similar folders are better than `day-07`, `day-08`, `day-09` for problem code.

Inside each topic, each problem should have its own folder.

Example:

```text
dsa/arrays/two-sum/
├── main.go
├── main_test.go
└── README.md
```

This gives you:

- one place for the solution
- one place for tests
- one place for explanation and complexity notes

### 3. `cmd/` and `internal/` are for real project code

Your repository is not only a notebook. It is also a Go project.

As your learning grows, you will likely build:

- CLI tools
- parsers
- config loaders
- notification formatters
- integration experiments

These should not live inside `dsa/`.

Use:

- `cmd/` for application entry points
- `internal/` for project-specific packages

This follows a common Go layout and builds good habits early.

### 4. `pkg/` should stay small and intentional

Use `pkg/` only for code that is generic enough to be reused.

Do not put everything there.

For a learning repository, `internal/` will usually be more useful than `pkg/`.

### 5. `docs/` stores durable guidance

Some files are not daily notes and not code.

Examples:

- runbooks
- architecture notes
- study plans
- setup instructions

Those belong in `docs/`.

This runbook belongs there because it explains how the repository should evolve over time.

## Why this structure is good

### 1. It scales cleanly

A repository with 5 files can survive with weak structure.
A repository with 100 practice problems cannot.

This layout stays understandable as the repository grows.

### 2. It supports both beginner learning and real engineering habits

You are learning syntax, DSA, and project design at the same time.
This structure supports all three without mixing them into one folder.

### 3. Revision becomes easier

If you want to revise:

- by day: go to `notes/`
- by topic: go to `dsa/`
- by project architecture: go to `cmd/`, `internal/`, and `docs/`

That reduces friction during revision.

### 4. Testing becomes easier

Problem-by-problem folders make it natural to add:

- `main_test.go`
- edge cases
- alternate solutions later

That is much cleaner than putting all solutions into one file.

### 5. It avoids future refactoring pain

If everything begins in the root directory, cleanup becomes harder later.
It is better to choose a clean layout early than move dozens of files later.

### 6. It makes your Git history clearer

Commits become easier to read:

- `notes: add day 4 learning notes`
- `dsa: add queue using slice`
- `internal/parser: add input parser`

This makes your repository easier to maintain.

## Naming guidance

Use lowercase directory names for consistency:

- `notes/` instead of `Notes/`
- `docs/` instead of mixed naming styles

Use zero-padded day files so sorting stays stable:

- `day-001-...`
- `day-002-...`
- `day-010-...`

Use problem-specific folder names in `dsa/`:

- `valid-anagram`
- `binary-search`
- `level-order-traversal`

## Recommended file pattern for each DSA problem

For each problem, use:

```text
problem-name/
├── main.go
├── main_test.go
└── README.md
```

Suggested responsibility of each file:

- `main.go`: solution implementation
- `main_test.go`: test cases
- `README.md`: problem statement, approach, complexity, and learning notes

If you later want multiple approaches:

```text
problem-name/
├── brute_force.go
├── optimized.go
├── main_test.go
└── README.md
```

## Daily workflow recommendation

When you practice each day:

1. Write the lesson summary in `notes/`.
2. Add the DSA problem under the correct topic in `dsa/`.
3. Add tests for the problem.
4. If the day includes project work, place that code in `cmd/` or `internal/`.
5. Commit with a message that reflects the actual area changed.

Example:

```text
notes/day-004-clean-architecture.md
dsa/queue/implement-queue-using-slices/
internal/router/
```

## Suggested next cleanup for this repository

The current repository has:

- `Notes/Day-1.md`
- `main.go`
- `go.mod`

Recommended future cleanup:

```text
Notes/Day-1.md           -> notes/day-001-cli-basics.md
main.go                  -> cmd/app/main.go
```

You do not need to do all cleanup immediately.
You can migrate gradually as the repository grows.

## Final rule

Use this simple decision rule whenever you create a new file:

```text
Is this a learning note?
-> put it in notes/

Is this a DSA problem or topic revision artifact?
-> put it in dsa/

Is this real Go application code?
-> put it in cmd/ or internal/

Is this long-term project guidance?
-> put it in docs/
```

That rule is enough to keep the repository clean.
