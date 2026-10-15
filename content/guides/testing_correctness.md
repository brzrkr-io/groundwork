# Testing and correctness

**In one sentence:** a test is a small program that runs your code on inputs you chose and checks the answer, so a machine can tell you in seconds, every time, whether your code still keeps its promises.

**Why it matters:** a bug caught by `go test` costs you a minute. The same bug caught by a failed deploy at 2 a.m. costs an incident. Ops code deletes old backups, scales clusters and rotates certificates, so a wrong answer can destroy things, not just look ugly.

**You'll learn:**
- what a test is, and how to read a failing one (got vs want)
- what a passing test proves, and what it never can
- how to choose test cases: the happy path, the edges and the errors
- the table-test idea and the unit / integration / end-to-end pyramid
- why tests must give the same result every run, and how fakes help
- the red-green-refactor loop and why every bug gets a test

## A test is an executable example

You've been running tests since *Hello, Go*: every Groundwork exercise is checked by one. Now you write them.

Say your alerting job needs the percentage of failed requests:

```go
package rate

// ErrorRate returns the percentage of requests that failed.
func ErrorRate(failed, total int) float64 {
	return float64(failed/total) * 100
}
```

A test is ordinary Go code in a file ending in `_test.go`, next to the code it tests:

```go
package rate

import "testing"

func TestErrorRate(t *testing.T) {
	got := ErrorRate(3, 200)
	want := 1.5
	if got != want {
		t.Errorf("ErrorRate(3, 200) = %v, want %v", got, want)
	}
}
```

Step by step:

1. The function name starts with `Test`, and it takes `t *testing.T`. That's how `go test` finds it.
2. It calls the code with an input you chose: 3 failures out of 200 requests.
3. It compares the result, `got`, with the answer you worked out by hand, `want`.
4. If they differ, `t.Errorf` marks the test as failed and records a message.

Read the test again and notice it is also documentation. It says, in code that can't go out of date: "3 failures out of 200 is 1.5 percent." That's why people call a test an *executable example*.

## Reading a failing test

Run `go test` in that folder and you get:

```text
--- FAIL: TestErrorRate (0.00s)
    rate_test.go:9: ErrorRate(3, 200) = 0, want 1.5
FAIL
exit status 1
FAIL	example.com/rate	0.001s
```

Read it top to bottom:

| Piece | What it tells you |
|---|---|
| `--- FAIL: TestErrorRate` | which test failed |
| `rate_test.go:9` | the file and line of the `t.Errorf` call |
| `ErrorRate(3, 200) = 0` | the call and what it actually returned: *got* |
| `want 1.5` | what the test expected: *want* |

The message format `Call(args) = got, want want` is a Go convention. It answers the three questions you always ask: what was called, what came back, what should have come back.

Now debug. `got` is 0, so trace the code by hand with `failed = 3` and `total = 200`. `failed/total` divides two `int`s, and integer division throws away the remainder: 3 / 200 is 0. The conversion to `float64` comes too late; the fraction is already gone.

The fix converts first, then divides:

```go
return float64(failed) / float64(total) * 100
```

Run again and you see `ok  example.com/rate  0.001s`. The test went from failing to passing, and it will now guard this line forever. The *Debugging with a method* guide covers this tracing step in depth.

## What a passing test proves

A passing test proves one thing: **for these inputs, today, the code returned these outputs.** Nothing more.

The computer scientist Edsger Dijkstra put it in one line in 1970: "Program testing can be used to show the presence of bugs, but never to show their absence!"

Think about why. A function that takes a string has endless possible inputs. Your test tries a handful. If the bug hides in an input you didn't try, every test passes and the bug ships.

So testing is *sampling*: you can't check everything, so you choose a few inputs that are most likely to reveal a bug. The next section is about choosing well.

Two more limits worth knowing:

- **A test can be wrong too.** If you worked out `want` wrongly, the test happily approves wrong code. You'll fix exactly this in *Your first test: fix a broken test*.
- **A test that has never failed might test nothing.** Break the code on purpose once and watch the test go red. If it stays green, it isn't checking what you think.

## Choosing test cases

Take a function that turns text like `"8080"` into a port number. Valid ports are 1 to 65535. Which inputs should you try?

Use three buckets:

| Bucket | Question | Port examples |
|---|---|---|
| happy path | does the normal case work? | `"8080"` |
| edges | what happens at the limits? | `"1"`, `"65535"`, `"0"`, `"65536"` |
| errors | is bad input rejected? | `""`, `"http"` |

### Edges are where bugs live

An *edge case* is an input at the boundary of what's allowed. Most bugs sit there, because the boundary is where you wrote `<` and meant `<=`. You saw this in *For loops: fix the off-by-one bugs*.

A checklist for edges:

- **empty**: an empty string, an empty list, zero hosts
- **one**: a single item, where "the first" and "the last" are the same thing
- **many**: more than one, to exercise the loop
- **limits**: the smallest and largest allowed values, and one past each

Try the checklist on `ErrorRate`. What about `ErrorRate(0, 0)`, a minute with no traffic at all? With floats, 0 divided by 0 gives NaN ("not a number"), and NaN compared with anything is false, so an alert like `rate > 5` would silently never fire. Is that what you want? The test forces you to decide what the code *should* do. Writing that decision down is a *specification*.

### Errors are part of the contract

If a function returns an error, the error is part of its promise. Test that bad input gets an error, not just that good input gets a value. A config loader that silently accepts port `80800` is a bug even if every valid config works.

Later in the *Testing* module, *Fuzzing* lets the computer invent thousands of inputs for you, which finds the edges you didn't think of.

## Table tests

Seven inputs means seven nearly identical blocks of code. A *table test* writes the cases as data, one row each, and runs the same check on every row.

Here is the function:

```go
package port

import (
	"fmt"
	"strconv"
)

// ParsePort turns text like "8080" into a port number.
func ParsePort(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("port %q is not a number", s)
	}
	if n < 1 || n > 65535 {
		return 0, fmt.Errorf("port %d is out of range 1-65535", n)
	}
	return n, nil
}
```

And its table test:

```go
package port

import "testing"

func TestParsePort(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{name: "typical", input: "8080", want: 8080},
		{name: "lowest", input: "1", want: 1},
		{name: "highest", input: "65535", want: 65535},
		{name: "zero", input: "0", wantErr: true},
		{name: "too big", input: "65536", wantErr: true},
		{name: "empty", input: "", wantErr: true},
		{name: "not a number", input: "http", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePort(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParsePort(%q) = %d, want an error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePort(%q) returned error %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParsePort(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
```

How it runs:

1. `tests` is a slice of structs. Each row is one case: a name, an input, and the expected result.
2. The loop takes each row in turn.
3. `t.Run` starts a *subtest* with the row's name, so a failure tells you which row broke.
4. Rows with `wantErr` must get an error. The others must get no error and the right number.

`go test -v` lists every row: `--- PASS: TestParsePort/too_big`, and so on. Adding a case is now one line, which means you actually add them.

Now break the code on purpose: change `n > 65535` to `n >= 65535`. The output names the guilty row:

```text
--- FAIL: TestParsePort (0.00s)
    --- FAIL: TestParsePort/highest (0.00s)
        port_test.go:30: ParsePort("65535") returned error port 65535 is out of range 1-65535
```

Only the edge case caught it. "typical" kept passing. That's the edges rule in action.

You practise this in *Table-driven tests* and *Subtests and helpers*.

## Unit, integration and end-to-end

Tests come in sizes. The bigger the test, the more it proves and the more it costs.

```text
            ┌──────────┐
            │   E2E    │      few · slow · realistic · can be flaky
          ┌─┴──────────┴─┐
          │ integration  │    some · your code + a real neighbour
        ┌─┴──────────────┴─┐
        │       unit       │  many · fast (milliseconds) · reliable
        └──────────────────┘
```

This shape is called the *test pyramid*: lots of small fast tests at the bottom, a few big slow ones at the top.

| Kind | Checks | Ops example |
|---|---|---|
| unit | one function, no network or disk | `ParsePort`, a log-line parser |
| integration | your code with a real dependency | your tool against a real Postgres in a container |
| end-to-end | the whole system from outside | deploy to staging, place a test order |
| smoke test | "is it alive at all?" right after a deploy | `curl` the `/healthz` endpoint |

A *smoke test* is a tiny end-to-end check run straight after a deploy. If it fails, the pipeline rolls back before users notice. You'll write health endpoints in the *Production services* module.

Why not only end-to-end tests? Because when one fails, it could be any of a thousand moving parts, and it takes minutes to run. A failing unit test points at one function in one second.

## Determinism: same input, same result

A test must give the same result every time it runs. A test that sometimes passes and sometimes fails, with no code change, is called *flaky*. Flaky tests are worse than none: people learn to ignore red, and then they ignore a real failure.

Flakiness comes from anything outside your inputs:

| Source | Example |
|---|---|
| the clock | a cert check that passes today and fails next month |
| randomness | random backoff delays, random IDs |
| the network | a test that calls the real AWS API |
| map order | Go deliberately varies the order of `range` over a map |
| timing | "sleep 1 second and hope the goroutine finished" |

The cure is to make those things **inputs**. Instead of reading the clock deep inside, take the current time as a parameter. Here both times are *Unix timestamps*: whole seconds since 1 January 1970, a common way to store a moment as a plain number:

```go
package cert

// ExpiresSoon reports whether a certificate expires within 14 days of now.
// expiry and now are Unix timestamps (seconds since 1 January 1970).
func ExpiresSoon(expiry, now int64) bool {
	const fourteenDays = 14 * 24 * 60 * 60 // in seconds
	return expiry-now < fourteenDays
}
```

The test then pins the clock to one fixed moment:

```go
package cert

import "testing"

func TestExpiresSoon(t *testing.T) {
	const day = 24 * 60 * 60     // one day in seconds
	var now int64 = 1_791_882_000 // 13 October 2026, 09:00 UTC

	if !ExpiresSoon(now+3*day, now) {
		t.Error("cert expiring in 3 days: ExpiresSoon = false, want true")
	}
	if ExpiresSoon(now+90*day, now) {
		t.Error("cert expiring in 90 days: ExpiresSoon = true, want false")
	}
}
```

It passes today, and it will still pass in 2030, because nothing in it depends on the real date. Only the caller in production reads the real clock and passes it in. (Go's `time` package has proper types for moments and durations; *Time and durations: cert expiry, uptime, maintenance windows* teaches them, and the idea stays the same.)

For bigger dependencies, like an AWS client or an HTTP API, you pass in a *fake*: a small stand-in that behaves the way you script it, with no network. Interfaces make the swap possible, as you saw in *Interfaces: swap in a fake for tests*. The *Abstraction* guide explains why this works, and you build fakes in *Fakes and dependency injection* and *Fakes: a fake clock for retries*.

## Every bug gets a test

When a bug reaches production, the fix has three steps, in this order:

1. **Reproduce it in a test.** Write a test that feeds in the input from the bug report. Run it. It must fail. Now you've proved you understand the bug.
2. **Fix the code.** The test turns green.
3. **Keep the test forever.**

That kept test is a *regression test*: it stops the bug from coming back (regressing) when someone refactors the code next year.

In ops terms, it's the action item from the postmortem that actually gets done, because the build enforces it. You practise this in *Practice: reproduce a bug report*.

## Red, green, refactor

You already know this loop. It's how every Groundwork exercise works.

```text
   ┌─────────┐  write code   ┌─────────┐   tidy up    ┌────────────┐
   │   RED   │ ────────────► │  GREEN  │ ───────────► │  REFACTOR  │
   └─────────┘  until it     └─────────┘  tests stay  └────────────┘
        ▲       passes                    green             │
        │                                                   │
        └──────────────── next small behaviour ─────────────┘
```

- **Red:** write a test for the next small piece of behaviour. It fails, because the code doesn't do it yet. In Groundwork, the exercise starts here.
- **Green:** write the simplest code that makes it pass.
- **Refactor:** clean the code up: better names, less repetition. The tests prove you didn't change behaviour.

Working this way is called *test-driven development* (TDD). You don't have to use it for everything. But the refactor step is the real prize of any test suite: you can improve old code without fear, because the tests tell you the moment you break something.

## Common mistakes

- **Only testing the happy path.** Bugs live at the edges and in error handling.
- **Never seeing the test fail.** Break the code once and confirm it goes red.
- **Copying the code's logic into `want`.** If you compute the expected value the same (wrong) way the code does, both agree and the bug hides. Work `want` out by hand.
- **Depending on the real clock, network or random numbers.** Pass them in, or use a fake.
- **Vague failure messages.** `t.Error("wrong")` makes you dig. Write `Call(args) = got, want want`.
- **Deleting or skipping a failing test to get a green build.** The test is the messenger.
- **Only end-to-end tests.** Slow, flaky and hard to diagnose. Most tests should be small unit tests.

## Check yourself

1. All 40 tests pass. Can you say the code has no bugs? Why or why not?
2. A test fails with `Retries(3) = 4, want 3`. Which value came from the code, and which from the test?
3. List the edge cases you'd test for a function that returns the largest number in a slice of ints.
4. A cert-expiry test passed for months, then failed on 1 March with no code change. What is the likely cause, and how do you fix it?
5. A customer reports that a hostname with a trailing dot breaks your inventory tool. What do you do before touching the code?
6. Why should most of your tests be unit tests rather than end-to-end tests?

## Answers

1. No. Passing tests prove the code works for the inputs you tried. Dijkstra's point: tests show the presence of bugs, never their absence. A bug can hide in an input no test tried.
2. `4` is *got*, from the code. `3` is *want*, from the test. The convention is `Call(args) = got, want want`.
3. An empty slice (what should it return, or should it error?), a single element, several elements with the largest first, last and in the middle, all negative numbers, and duplicates of the largest value.
4. The test or the code uses the real current date, so the result changed when the calendar did. Pass `now` in as a parameter (or use a fake clock) and pin it to a fixed date in the test.
5. Write a test using that exact hostname, run it, and see it fail. That proves you've reproduced the bug. Then fix the code, and keep the test as a regression test.
6. Unit tests run in milliseconds, give the same result every time, and point at one function when they fail. End-to-end tests are slow, can be flaky, and a failure could come from anywhere. Keep a few for confidence, but rely on many small ones.

## Key terms

| Term | Meaning |
|---|---|
| test | code that runs other code on chosen inputs and checks the result |
| got / want | the actual result and the expected result in a test |
| edge case | an input at the boundary of what's allowed: empty, max, one past |
| specification | a precise statement of what code should do for each input |
| table test | test cases written as rows of data, checked by one loop |
| subtest | a named test inside a test, started with `t.Run` |
| unit test | a fast test of one function or type in isolation |
| integration test | a test of your code together with a real dependency |
| end-to-end test | a test of the whole system from the outside |
| smoke test | a quick "is it alive?" check, often right after a deploy |
| flaky test | a test that sometimes passes and sometimes fails with no code change |
| fake | a simple stand-in for a real dependency, used in tests |
| regression test | a test that stops a fixed bug from coming back |
| TDD | test-driven development: the red, green, refactor loop |

## Watch

- [Software Testing Explained in 100 Seconds](https://www.youtube.com/watch?v=u6QfIXgjwGQ) · Fireship · 2 min: a fast overview of unit, integration and end-to-end tests before you dive in.
- [Go Class: 38 Testing](https://www.youtube.com/watch?v=PIPfNIWVbc8) · Matt KØDVB · 36 min: a patient lecture on Go's `testing` package, with table tests and subtests written live.
- [GopherCon 2017: Advanced Testing with Go - Mitchell Hashimoto](https://www.youtube.com/watch?v=8hQG7QlcLBk) · Gopher Academy · 45 min: how HashiCorp tests real ops tools like Terraform; the first half on table tests and test helpers fits this guide best.

## Go deeper

- [Add a test](https://go.dev/doc/tutorial/add-a-test): the official Go tutorial's short, hands-on first test.
- [Go Wiki: TableDrivenTests](https://go.dev/wiki/TableDrivenTests): the table-test pattern with real examples from the Go project.
- [Using Subtests and Sub-benchmarks](https://go.dev/blog/subtests): how `t.Run` names and runs each row, and how to run just one.
- [The Practical Test Pyramid](https://martinfowler.com/articles/practical-test-pyramid.html): a long but clear tour of unit, integration and end-to-end tests in real projects.
- [Notes on Structured Programming (EWD249)](https://www.cs.utexas.edu/~EWD/ewd02xx/EWD249.PDF): Dijkstra's 1970 notes, the source of the "presence of bugs" quote.
