# Fakes and dependency injection

A disk-space check that pages the on-call engineer is only useful if the page
really goes out, and goes out only when it should. But a test that pages a
real person every time it runs is a disaster. This exercise shows how to test
code like that without sending anything: hand it a *fake*.

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
`Notify` method, because `Notifier` is an interface (as in *Interfaces*). The
caller decides what `n` is. In production that is a type that posts to Slack
or PagerDuty. Passing in the things a function depends on, instead of having
the function create them itself, is called *dependency injection*. It sounds
fancy, but it is just a parameter.

If `CheckDisk` called a Slack client directly, every test run would post to
Slack, need the network and a token, and you still couldn't easily check what
it sent. With an interface parameter, a test can pass in something else.

### A fake that records calls

A *fake* is a small, simple type that implements the interface for tests.
Instead of doing the real work, this one writes down what it was asked to do:

```go
type fakeNotifier struct {
	messages []string
}

func (f *fakeNotifier) Notify(message string) {
	f.messages = append(f.messages, message)
}
```

The test then runs the code with the fake and looks at what was recorded:

```go
f := &fakeNotifier{}
CheckDisk(f, "web-1", 93)
// f.messages is now []string{"disk on web-1 is 93% full"}
```

Nobody gets paged, nothing touches the network, and you can check the exact
messages, and how many there were.

### Why the pointer receiver matters

`Notify` changes the fake, so it needs a pointer receiver, as in *Pointer
receivers*. With `func (f fakeNotifier) Notify(...)`, each call would append
to a copy, and `f.messages` would stay empty in your test.

With a pointer receiver, only `*fakeNotifier` has the `Notify` method, so only
a pointer satisfies `Notifier`. That is why the test creates the fake with
`&fakeNotifier{}`. If you pass a plain `fakeNotifier`, the compiler stops you
with a message like `fakeNotifier does not implement Notifier (method Notify
has pointer receiver)`. Use a fresh fake for each check, so messages from one
check don't leak into the next.

### Grading

As in *Your first test*, `TestGrader` re-runs your `TestCheckDisk` against the
real `CheckDisk`, where it must pass, and against *mutants* with one bug each,
where it must fail. It names any bug your test missed. The `checkDisk`
variable in `exercise.go` is how it swaps them in; leave it alone.

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

In `exercise_test.go`, declare `fakeNotifier` as a struct with one field,
`messages []string`. Then add a `Notify` method with a pointer receiver that
appends `message` to `f.messages`. Now `*fakeNotifier` is a `Notifier`.

### 2. Check that a full disk sends one alert

In `TestCheckDisk`, create a fake with `&fakeNotifier{}`, call
`CheckDisk(f, "web-1", 93)`, and check that:

- `f.messages` has exactly 1 message (fail with `t.Fatalf` if not, because
  there is no message to look at),
- that message is exactly `"disk on web-1 is 93% full"`.

Then do the same with a fresh fake for `CheckDisk(f, "db-2", 90)`, which must
send `"disk on db-2 is 90% full"`. 90 is exactly the threshold, where a bug
like `>` instead of `>=` hides.

### 3. Check that a healthy disk sends nothing

With a fresh fake each time, check that `CheckDisk(f, "web-1", 50)` and
`CheckDisk(f, "web-1", 89)` send no messages at all (`len(f.messages)` is 0).

The exercise is done when `TestCheckDisk` passes and `TestGrader` reports that
every mutant was caught.

## Learn more

- [Go by Example: Interfaces](https://gobyexample.com/interfaces)
- [Effective Go: Interfaces](https://go.dev/doc/effective_go#interfaces)
- [Package testing](https://pkg.go.dev/testing)
