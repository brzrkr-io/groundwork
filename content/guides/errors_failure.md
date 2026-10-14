# Errors and failure

**In one sentence:** every program that touches the real world will meet failure, and good code decides on purpose what to do about each one: handle it, pass it up with context, or stop.

**Why it matters:** at 3am the on-call engineer sees one line from your tool. `error: failed` wakes up three more people; `deploy api: push image api:v42: dial tcp 10.0.3.7:443: i/o timeout` tells them exactly where to look.

**You'll learn:**

- The three kinds of things that go wrong, and why each needs a different response
- How Go treats errors as ordinary values, and how that differs from exceptions
- The three choices at every error: handle it, return it, or wrap it with context
- When to fail fast and when to keep going and collect errors
- What preconditions and invariants are, and why panics are for bugs
- How to clean up with `defer`, survive partial failure, and write error messages humans can act on

## Three kinds of things going wrong

Not all failures are the same, and treating them the same is where most error-handling mistakes start.

| Kind | Examples | What to do |
|---|---|---|
| *Bug* | index out of range, a forgotten `switch` case | fix the code |
| *Expected failure* | config typo, unknown flag, port already in use | tell the user how to fix it |
| *Environment failure* | network timeout, disk full, API returns 503 | report it; maybe retry or skip |

**Bugs** are the programmer's mistake. No amount of handling at runtime fixes them; the code must change.

**Expected failures** are normal. Users mistype things. Files go missing. Your code should expect these and explain them clearly.

**Environment failures** come from the world your program runs in: networks, disks, other services. They are not your fault, but they are your responsibility. In ops tooling, they are the most common kind by far.

## Errors are values

In Go, a function that can fail returns an `error` as its last result. `nil` means "no error". The caller checks it straight away:

```go
replicas, err := strconv.Atoi(text)
if err != nil {
	return fmt.Errorf("parse replicas %q: %w", text, err)
}
```

An `error` is just a value, like an `int` or a `string`. Any type with an `Error() string` method is an error. Because it's a value, you can store it, compare it, put it in a slice, or wrap it in another error with ordinary code. That's the idea behind the Go proverb "errors are values".

### How other languages do it

Python, Java and JavaScript mostly use *exceptions*. A failing function "throws", and the exception jumps up through the callers, skipping their remaining code, until some `try`/`catch` block catches it. If nothing catches it, the program crashes.

| | Go error values | Exceptions |
|---|---|---|
| Can this call fail? | visible in the signature | you have to read docs or code |
| Where is it handled? | right at the call | anywhere up the stack, or nowhere |
| Cost | more typing: `if err != nil` | easy to forget a failure path exists |

Neither is magic. Go chose to make failure visible on the page, at the price of more lines. When you read Go code, every line that can fail says so.

## Handle, return, or wrap

Every time you get an error, you choose one of three things.

**Handle it** when you know what to do. A missing optional config file? Use the defaults. A host that answered 503? Retry later. Handling means the error stops here.

**Return it** when you can't do anything useful at this level. Let the caller, who knows more about the bigger picture, decide.

**Wrap it** when you return it: add what *you* were doing. `fmt.Errorf` with the `%w` verb builds a new error that contains the old one:

```go
return fmt.Errorf("push image %s: %w", image, err)
```

### Error chains read like a story

Each function adds its own piece of context as the error travels up:

```go
package main

import (
	"errors"
	"fmt"
)

// ErrTimeout stands in for the network's "nobody answered" error.
var ErrTimeout = errors.New("i/o timeout")

func dial(addr string) error {
	// Pretend the registry never answers.
	return fmt.Errorf("dial tcp %s: %w", addr, ErrTimeout)
}

func pushImage(image string) error {
	if err := dial("10.0.3.7:443"); err != nil {
		return fmt.Errorf("push image %s: %w", image, err)
	}
	return nil
}

func deploy(service string) error {
	image := service + ":v42"
	if err := pushImage(image); err != nil {
		return fmt.Errorf("deploy %s: %w", service, err)
	}
	return nil
}

func main() {
	err := deploy("api")
	fmt.Println(err)
	fmt.Println("timeout?", errors.Is(err, ErrTimeout))
}
```

It prints:

```text
deploy api: push image api:v42: dial tcp 10.0.3.7:443: i/o timeout
timeout? true
```

Step by step:

- `dial` fails and wraps the base error `ErrTimeout` with the address.
- `pushImage` gets that error and wraps it with the image name.
- `deploy` wraps it again with the service name and returns it to `main`.
- Read left to right, the message goes from *what you were trying to do* to *what actually broke*.
- `errors.Is` looks through every layer of wrapping and finds `ErrTimeout` at the bottom. Code can make decisions (such as "retry on timeout") without parsing the text.

Go style keeps each piece short and lowercase, with no "failed to" or "error:" prefixes. Otherwise chains read like `failed to deploy: failed to push: failed to dial`, which is noise.

`%w` keeps the original error inspectable; `%v` would turn it into plain text, and `errors.Is` would no longer find it. You practise this in *Wrapping errors* and *Custom error types*.

## Fail fast or keep going?

### Fail fast: validate at startup

Your backup tool reads a config with a typo: `retension: 7d`. If it only notices at 2am when the first backup runs, nobody is awake to see it. If it checks every field at startup and refuses to run, the person who just deployed it sees the problem immediately.

*Failing fast* means stopping as soon as you know something is wrong, before you've done any damage. A program that refuses to start is far easier to fix than one that starts and does the wrong thing.

Kubernetes rewards this. A Pod that exits at startup shows up as `CrashLoopBackOff` during the rollout, and the rollout stops making progress while most of the old Pods keep serving. A Pod that starts with a bad config and quietly misbehaves sails through.

### Keep going: the other 999 hosts

Now you're patching 1,000 servers, and host number 17 doesn't answer. If you stop there, 983 servers stay unpatched because of one dead machine.

When the items are independent, keep going, collect the errors, and report them all at the end:

```go
package main

import (
	"errors"
	"fmt"
)

// patch pretends to patch a host; two of them are unreachable.
func patch(host string) error {
	switch host {
	case "web-2":
		return errors.New("ssh: connection refused")
	case "web-4":
		return errors.New("disk full")
	}
	return nil
}

func main() {
	hosts := []string{"web-1", "web-2", "web-3", "web-4", "web-5"}
	var errs []error
	patched := 0
	for _, h := range hosts {
		if err := patch(h); err != nil {
			errs = append(errs, fmt.Errorf("patch %s: %w", h, err))
			continue // one bad host must not stop the others
		}
		patched++
	}
	fmt.Printf("patched %d of %d hosts\n", patched, len(hosts))
	if err := errors.Join(errs...); err != nil {
		fmt.Println(err)
	}
}
```

It prints:

```text
patched 3 of 5 hosts
patch web-2: ssh: connection refused
patch web-4: disk full
```

Each failure is wrapped with its host name and saved, the loop carries on, and `errors.Join` combines the saved errors into one, printed one per line. If nothing failed, `errors.Join` returns `nil`. The *Reporting many errors* exercises build on this.

### How to choose

| Fail fast when... | Keep going when... |
|---|---|
| the input or config is broken | the items are independent |
| continuing could cause damage | partial progress is valuable |
| nothing useful has happened yet | you can report every failure at the end |

One refinement: if 300 of the 1,000 hosts fail, something systemic is wrong (a bad patch, a dead network). Good fleet tools stop when failures pass a threshold.

## Preconditions and invariants

A *precondition* is something that must be true before a function can do its job: "replicas must be between 0 and 100", "the image tag must not be empty". Check preconditions at the edges of your program, where input arrives, and return an error that says what was wrong.

An *invariant* is something that must *always* be true about your data: "ready replicas never exceed desired replicas", "a finished deploy has a finish time". Your own code is responsible for keeping invariants true. If one is broken, that is a bug, not bad input.

| Something is wrong with... | It is a... | Respond with... |
|---|---|---|
| input from users, files, networks | expected failure | an error |
| your program's own logic | bug | a panic is acceptable |

### Panics are for bugs

A *panic* stops the normal flow, runs any deferred calls, and crashes the program with a stack trace. Use it for "this cannot happen unless the code is wrong": an enum value your `switch` doesn't know, or a regular expression written in your source that fails to compile at startup.

Never panic for network failures, missing files or bad user input. Those are expected; return an error.

A server can use `recover` to stop one buggy request from killing the whole process:

```go
package main

import "fmt"

// lastLine has a bug: it never checks for an empty slice.
func lastLine(lines []string) string {
	return lines[len(lines)-1]
}

// handle is one request. A panic here must not kill the whole server.
func handle(lines []string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("request crashed:", r)
		}
	}()
	fmt.Println("last line:", lastLine(lines))
}

func main() {
	handle([]string{"booting", "ready"})
	handle(nil)
	fmt.Println("server still running")
}
```

It prints:

```text
last line: ready
request crashed: runtime error: index out of range [-1]
server still running
```

The second call panics inside `lastLine`. The deferred function runs, `recover` catches the panic, and `main` carries on. Go's own HTTP server does something similar for each request. But `recover` is a safety net, not a fix: the bug in `lastLine` is still there. See *panic and recover*.

## Cleanup with defer

Your migration tool turns on maintenance mode, runs some steps, and turns it off. If a step fails halfway and you return early, maintenance mode stays on and the site stays down.

`defer` schedules a call to run when the function returns, by *any* path: a normal return, an early `return err`, or a panic.

```go
package main

import "fmt"

func migrate(steps []string) error {
	fmt.Println("maintenance mode: on")
	defer fmt.Println("maintenance mode: off")

	for _, s := range steps {
		if s == "drop table users" {
			return fmt.Errorf("step %q: refused, needs approval", s)
		}
		fmt.Println("ran:", s)
	}
	return nil
}

func main() {
	err := migrate([]string{"add column", "drop table users", "backfill"})
	fmt.Println("result:", err)
}
```

It prints:

```text
maintenance mode: on
ran: add column
maintenance mode: off
result: step "drop table users": refused, needs approval
```

The early `return` still ran the deferred call. Put the cleanup right next to the setup, and you can't forget it on some rarely used error path. If you defer several calls, they run in reverse order: last deferred, first run. Practise in *defer* and *defer: safe database migrations*.

## Partial failure

The hardest failures leave the world half-changed. Your deploy updated 3 of 5 regions, then the network dropped. You created the DNS record but not the load balancer.

Worse: when a request times out, you don't know whether it worked. The server may have done the work and the reply got lost.

Strategies that help:

- **Make steps safe to repeat.** "Create the bucket if it's missing" can be run again after a failure; "create the bucket" fails the second time. A step you can safely repeat is *idempotent*.
- **Order steps carefully.** Write the new thing before deleting the old one. Writing a temp file and renaming it, as in *Atomic writes: temp file and rename*, means readers see the old file or the new one, never half of each.
- **Say exactly what got done.** "Updated eu-west-1, eu-central-1, us-east-1; failed at us-west-2" lets a human finish the job.
- **Decide: roll back or roll forward.** Undo the finished steps, or fix the problem and continue.

The *Distributed systems basics* guide and the *Declarative systems and control loops* guide go much further.

## Error messages for humans

Picture the reader: someone half-awake, who has never seen your code, holding a pager. A good message answers four questions:

| Question | Example |
|---|---|
| What were you doing? | `load config` |
| Where? | `/etc/backup.yaml: line 12` |
| Why did it fail? | `unknown field "retension"` |
| What next? | `did you mean "retention"?` |

Bad and better, side by side:

```text
bad:    error
better: load config /etc/backup.yaml: line 12: unknown field
        "retension" (did you mean "retention"?)

bad:    failed to connect
better: connect to postgres db-1.internal:5432: connection refused
        (is the database running?)

bad:    invalid input
better: --replicas must be between 1 and 50, got 500
```

Two more habits:

- **Include the value you got.** "got 500" turns a guessing game into a glance.
- **Never put secrets in errors.** Error messages end up in logs, CI output and chat. A URL with a token in it is a leaked token. See the *Security basics for automation* guide.

When a command-line tool fails, it should print the message to standard error and exit with a non-zero code, so scripts and pipelines notice. You'll do this in the *CLI tools and the OS* module.

## Common mistakes

- **Ignoring errors.** `n, _ := strconv.Atoi(s)` turns `"ten"` into `0` without a word. Every dropped error is a future mystery.
- **Logging and returning.** If every layer logs the error and returns it too, one failure appears five times in the logs. Do one or the other.
- **Wrapping without adding anything.** `fmt.Errorf("error: %w", err)` adds noise. Add the *what* and *where*.
- **Comparing error text.** `err.Error() == "i/o timeout"` breaks when the wording changes. Use `errors.Is` or `errors.As`.
- **Panicking on expected failures.** A missing file is not a bug. Return an error.
- **Stopping a fleet run at the first error,** or the opposite: carrying on after the config itself turned out to be broken.
- **Swallowing errors to "keep things running".** A tool that hides failures reports success while nothing worked.

## Check yourself

1. Classify each as a bug, an expected failure or an environment failure: (a) the S3 API returns 503, (b) a user passes `--replicas=-3`, (c) your code reads `hosts[len(hosts)]`.
2. A function returns `fmt.Errorf("fetch secret %s: %v", name, err)`. Later, the caller's `errors.Is(err, ErrNotFound)` is always false. Why?
3. Your tool rotates certificates on 200 load balancers. Should it stop at the first failure? What should it print at the end?
4. Your service reads its config at startup and finds `port: "eighty"`. What should it do, and why not wait until the port is needed?
5. A teammate writes `panic("database unreachable")` when the database connection fails. What would you suggest?
6. Rewrite this message so on-call can act on it: `Error: operation failed`.

## Answers

1. (a) Environment failure: the world, not your code; report it, maybe retry. (b) Expected failure: bad user input; return an error saying the valid range. (c) Bug: valid indexes stop at `len(hosts)-1`; fix the code.
2. `%v` turns the inner error into plain text, so the chain is broken and `errors.Is` can't find `ErrNotFound`. Use `%w` to wrap.
3. No: the load balancers are independent, so one failure shouldn't leave the other 199 untouched. Collect errors as it goes, then print how many succeeded and each failure with its load balancer name, and exit non-zero. Consider stopping if failures pass a threshold.
4. Fail fast: exit at startup with a message like `config: port: want a number, got "eighty"`. The person deploying sees it right away, and Kubernetes halts the rollout instead of shipping a broken version.
5. An unreachable database is an environment failure, not a bug, so return an error (wrapped with the database address). Let the caller decide whether to retry, wait, or exit with a clear message.
6. Something like `rotate cert for lb-prod-3: upload to ACM: AccessDenied (does the deploy role have acm:ImportCertificate?)`: what, where, why and what next.

## Key terms

| Term | Meaning |
|---|---|
| bug | a mistake in the code; fixed by changing the code |
| expected failure | a normal problem, such as bad input, the program should explain |
| environment failure | a failure from outside: network, disk, other services |
| error value | in Go, an ordinary value describing a failure; `nil` means success |
| exception | a failure that jumps up the call stack until caught (other languages) |
| wrapping | adding context to an error while keeping the original inside (`%w`) |
| error chain | an error wrapped in layers, read from outer action to root cause |
| fail fast | stop as soon as something is known to be wrong, before doing damage |
| precondition | what must be true before a function can do its job |
| invariant | what must always be true about your data; broken means a bug |
| panic | an immediate stop with a stack trace; for bugs, not expected failures |
| recover | catches a panic inside a deferred function |
| defer | schedules a call to run when the function returns, by any path |
| partial failure | some steps succeeded and some didn't, leaving a half-changed world |
| idempotent | safe to repeat: doing it twice has the same effect as once |

## Watch

- [GopherCon 2016: Don't Just Check Errors Handle Them Gracefully - Dave Cheney](https://www.youtube.com/watch?v=lsBF58Q-DnY) · Gopher Academy · 22 min: the classic Go talk on why errors are values, and on handling an error once while adding context as it travels up.
- [Go Class: 32 Error Handling](https://www.youtube.com/watch?v=oIxXp0OgK_0) · Matt KØDVB · 34 min: a calm lecture on wrapping with `%w`, `errors.Is` and `errors.As`, and when a panic is the right call.
- [Debugging Under Fire: Keep your Head when Systems have Lost their Mind • Bryan Cantrill • GOTO 2017](https://www.youtube.com/watch?v=30jNsCVLpAE) · GOTO Conferences · 53 min: what failure looks like from the pager side, and why clear errors and calm thinking matter during an outage.

## Go deeper

- [Errors are values](https://go.dev/blog/errors-are-values) (Rob Pike): why Go treats errors as ordinary values, and patterns that cut down on repetition.
- [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors): wrapping with `%w`, `errors.Is` and `errors.As`, and when to wrap at all.
- [Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover): exactly how the three work together, with examples.
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments#error-strings): the style rules for error strings and "don't panic", used in real code reviews.
