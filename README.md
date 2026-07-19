# simple-worker-in-golang

Companion code for the blog post [_Un worker en Go : la brique de base de la concurrence_](https://www.clevertechware.fr/blog/2026/un-simple-worker-en-go).

A **worker** is the simplest concurrency pattern in Go: a goroutine that consumes tasks from a channel until it's closed. This repo walks through four increasingly production-ready versions of that idea.

## Structure

| Folder | Concept | Run |
|---|---|---|
| [`01-basic`](01-basic) | Minimal worker: a goroutine + a channel + `range`. Ends with a `time.Sleep` to let the output flush. | `make run-basic` |
| [`02-waitgroup`](02-waitgroup) | Same worker, but shutdown is synchronized with a `sync.WaitGroup` instead of a sleep. | `make run-waitgroup` |
| [`03-context-aware`](03-context-aware) | Two demos: a buffered `EmailWorker` (enqueue/shutdown API) and a `context`-aware worker that stops on cancellation or channel close. | `make run-context-aware` |
| [`04-advanced`](04-advanced) | A small restaurant simulation (`orders` package): a `Kitchen` that cooks orders concurrently, a `Waiter` that takes and serves them, an in-memory `OrderRepository`, and a generic `fp.Result[T]` type to carry success/failure through channels. Graceful shutdown on `SIGINT`/`SIGTERM`. | `make run-advanced` |

## Why a single worker

A single worker is enough when:
- **Throughput is low** — a few tasks per second
- **Order matters** — tasks must be processed sequentially
- **Simplicity wins** — no need for extra complexity
- **Resources are constrained** — a single DB connection, a single file

## Requirements

Go 1.26 (see [`go.mod`](go.mod)).

## Usage

```bash
make build              # builds bin/basic, bin/waitgroup, bin/context-aware, bin/advanced
make run-basic          # go run 01-basic/main.go
make run-waitgroup      # go run 02-waitgroup/main.go
make run-context-aware  # go run 03-context-aware/main.go
make run-advanced       # go run 04-advanced/main.go
make clean              # removes bin/
```

