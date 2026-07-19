# simple-worker-in-golang

Companion code for the blog post [_Un worker en Go : la brique de base de la concurrence_](https://www.clevertechware.fr/blog/2026/un-simple-worker-en-go).

A **worker** is the simplest concurrency pattern in Go: a goroutine that consumes tasks from a channel until it's closed. This repo walks through four increasingly production-ready versions of that idea.

## Structure

| Folder | Concept | Run |
|---|---|---|
| [`01-basic`](01-basic) | Minimal worker: a goroutine + a channel + `range`. Ends with a `time.Sleep` to let the output flush. | `make run-basic` |
| [`02-waitgroup`](02-waitgroup) | Same worker, but shutdown is synchronized with a `sync.WaitGroup` instead of a sleep. | `make run-waitgroup` |
| [`03-context-aware`](04-advanced) | A small restaurant simulation (`orders` package): a `Kitchen` that cooks orders concurrently, a `Waiter` that takes and serves them, an in-memory `OrderRepository`, and a generic `fp.Result[T]` type to carry success/failure through channels. Graceful shutdown on `SIGINT`/`SIGTERM`. | `make run-advanced` |
| [`04-email-use-case`](04-email-use-case) | HTTP server with a `POST /signup` endpoint: the handler enqueues a welcome email on a buffered, `context`-aware `EmailWorker` and responds immediately while the (mocked) email sends in the background. Graceful shutdown on `SIGINT`/`SIGTERM`. | `make run-context-aware` |

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
make run-context-aware  # go run 03-email-use-case/main.go
make run-email-use-case # go run 04-email-use-case/main.go
make clean              # removes bin/
```

## Simulating overload (03-context-aware)

`EmailWorker.Enqueue` never blocks: the `jobs` channel has a 100-slot buffer, and once
it's full `Enqueue` returns `ErrQueueFull` immediately instead of stalling the caller.
The `/signup` handler turns that into `503 Service Unavailable`.

To see it in practice, start the server in one terminal:

```bash
make run-email-use-case
```

Then, in another terminal, fire more concurrent signups than the buffer can hold
(the mock sender takes 4s per email, so the buffer drains slowly):

```bash
for i in $(seq 1 110); do
  (
    code=$(curl -s -o /dev/null -w "%{http_code}" \
      -X POST localhost:8080/signup \
      -d "{\"email\":\"user${i}@example.com\"}")
    if [ "$code" = "201" ]; then
      echo "user $i created"
    else
      echo "user $i not created: service unavailable (HTTP $code)"
    fi
  )
done
```

Expect ~101 lines of `user N created` (1 in-flight job + 100 buffered) followed by
a handful of `user N not created: service unavailable` once the buffer fills up.

