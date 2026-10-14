# Fakes and dependency injection

A disk-space check that pages the on-call engineer is only useful if the page
really goes out, and only when it should. But a test that pages a real person
on every run is a disaster. This exercise shows how to test such code without
sending anything: you hand it a *fake*.

**New here:** dependency injection, writing a fake that records calls.

## Learn

### Dependency injection

Here is the code you will test:

```go
type Notifier interface {
	Notify(message string)
}

func CheckDisk(n Notifier, host string, usedPercent int) {
	if usedPercent >= 90 {
		n.Notify(fmt.Sprintf("disk on %s is %d%% full", host, usedPercent))
	}
}
```

`CheckDisk` doesn't know *how* alerts are sent. It only knows that `n` has a
`Notify` method, because `Notifier` is an interface (as in *Interfaces*).
The caller decides what `n` really is: in production, a type that posts to
Slack or PagerDuty.

Passing in the things a function depends on, instead of letting it create
them, is called *dependency injection*. It sounds fancy, but it is just a
parameter. And because it's a parameter, a test can pass in something else.

### A fake that records calls

A *fake* is a small type that implements the interface for tests. Instead of
doing the real work, it writes down what it was asked to do. Here is one for
a different interface, a `Logger` with a `Log(line string)` method:

```go
type fakeLogger struct {
	lines []string
}

func (l *fakeLogger) Log(line string) {
	l.lines = append(l.lines, line)
}
```

A test runs the code with the fake, then looks inside it:

```go
l := &fakeLogger{}
Deploy(l, "api")
// l.lines now holds every line Deploy logged
```

Nothing touches the network, and you can check the exact text and how many
calls there were.

> **Why is it written like this?** Go requires the method's name, parameter
> types and result types to match the interface exactly, or the type doesn't
> implement it. The struct name, the field name and the receiver name (`l`)
> are your choice. Starting fake types with `fake` is convention.

### Why the pointer receiver matters

The method changes the fake, so it needs a pointer receiver (`*fakeLogger`),
as in *Pointer receivers*. With a plain receiver, each call appends to a
*copy*, and the slice in your test stays empty.

With a pointer receiver, only the pointer type has the method. So you create
the fake with `&` in front. Forget it, and the compiler says something like
`fakeLogger does not implement Logger (method Log has pointer receiver)`.
That error points at the call, but the fix is the `&` where you created the
fake.

Use a fresh fake for each check, so messages from one check don't leak into
the next.

### Grading

As in *Your first test*, `TestGrader` re-runs your `TestCheckDisk` against
the real `CheckDisk`, where it must pass, and against *mutants* with one bug
each, where it must fail. The `checkDisk` variable in `exercise.go` is how it
swaps them in; leave it alone.

## Your task

The monitoring agent calls `CheckDisk` for every host. The code in
`exercise.go` is correct. Write a fake notifier and use it in `TestCheckDisk`
in `exercise_test.go`.

`CheckDisk` sends exactly one message when the disk is 90% full or more, and
nothing below that:

```go
CheckDisk(n, "web-1", 93) // n.Notify("disk on web-1 is 93% full")
CheckDisk(n, "web-1", 50) // no call to n.Notify
```

### 1. Write the fake

**Goal:** declare a `fakeNotifier` struct with one field, `messages
[]string`, and a `Notify` method that appends each message to it.

Use a pointer receiver. When you're done, `*fakeNotifier` is a `Notifier`.

**Think it through:**
- What exact method line does `Notifier` ask for? Copy its name and
  parameter.
- Which field should each message go into?

### 2. Check that a full disk sends one alert

**Goal:** check that a full disk produces exactly one message with the right
text.

Use a fresh fake (made with `&`) for each call:

| Call | Expected messages |
| --- | --- |
| `CheckDisk(f, "web-1", 93)` | exactly `"disk on web-1 is 93% full"` |
| `CheckDisk(f, "db-2", 90)` | exactly `"disk on db-2 is 90% full"` |

First check there is exactly 1 message, and stop with `t.Fatalf` if not,
because there is no message to look at. Then check its text. 90 is exactly
the threshold, where a bug like `>` instead of `>=` hides.

**Think it through:**
- After the call, where does the fake keep what it received?
- Why must you check the count before reading the first message?

### 3. Check that a healthy disk sends nothing

**Goal:** check that `CheckDisk(f, "web-1", 50)` and `CheckDisk(f, "web-1",
89)` send no messages at all.

Use a fresh fake for each call, and fail if its `messages` isn't empty.

**Think it through:**
- What length does an empty slice have?
- Could a loop over the two percentages save you repeating code?

The exercise is done when `TestCheckDisk` passes and `TestGrader` reports that
every mutant was caught.

## Learn more

- [Go by Example: Interfaces](https://gobyexample.com/interfaces)
- [Effective Go: Interfaces](https://go.dev/doc/effective_go#interfaces)
- [Package testing](https://pkg.go.dev/testing)
