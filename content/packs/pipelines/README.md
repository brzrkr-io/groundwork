# Pipelines

A *pipeline* is a chain of *stages* connected by channels. Each stage does
one small job and passes its results on, like `cat app.log | grep ERROR` in
a shell. All the stages run at the same time, each in its own goroutine.

**New here:** pipeline stages. (Receive-only channel types `<-chan T` are
from *Channels: deadlocks and directions*.)

## Learn

### A stage

Every stage has the same shape: take an input, return an output channel, and
do the work in a goroutine.

```go
// Upper sends every string from in, upper-cased.
func Upper(in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out) // runs when the goroutine ends
		for s := range in {
			out <- strings.ToUpper(s)
		}
	}()
	return out // returns straight away; the goroutine keeps working
}
```

Read it in three steps:

1. Make the output channel.
2. Start a goroutine that loops, sends, and closes the channel at the end.
3. Return the channel **immediately**. Don't wait for the goroutine.

> **Why is it written like this?**
> - **Needed, not style:** a `range` over a channel only ends when the
>   channel is closed, so `out` must be closed (here by `defer close(out)`).
>   The compiler doesn't check this; a missing close makes receivers hang.
> - **Convention:** the stage that makes `out` is the only one that sends on
>   it and closes it. The return type `<-chan string` (receive only) lets the
>   compiler stop callers from sending or closing.
> - **Author's choice:** the names `in` and `out`.

### Connecting stages

Each stage's output is the next stage's input:

```go
for line := range Upper(Generate(lines)) {
	fmt.Println(line)
}
```

```text
Generate ──chan string──▶ Parse ──chan Entry──▶ OnlyLevel ──chan Entry──▶ range
```

Closing flows down the chain: `Generate` closes, so `Parse`'s `range` ends and
`Parse` closes, and so on.

### Looping over a slice vs a channel

| Looping over | Syntax | Gives you |
|---|---|---|
| a slice | `for i, v := range items` | index and value (use `_` to skip the index) |
| a channel | `for v := range ch` | one value per receive, until closed |

### Splitting with strings.Cut

`strings.Cut(s, sep)` splits `s` around the **first** `sep` and returns
three values:

```go
before, after, found := strings.Cut("key=value=x", "=")
// before = "key", after = "value=x", found = true

before, after, found = strings.Cut("novalue", "=")
// before = "novalue", after = "", found = false
```

> **Why is it written like this?** Go requires you to accept all three
> results or skip one with `_`. To use `strings`, add `"strings"` to the
> `import` list at the top of the file.

## Your task

Build a three-stage pipeline that finds error lines in an application log.
Each stage returns its channel straight away, works in a goroutine, and
closes its channel when finished. The tests fail if a stage returns `nil` or
never closes its channel.

### 1. Generate

**Goal:** `Generate(lines)` returns a channel that carries every line, in order, then is closed.

The values come from the `lines` slice, not from a channel. An empty slice
gives a channel that is just closed.

```go
Generate([]string{"INFO started", "ERROR disk full"})
// carries "INFO started", "ERROR disk full", then is closed
```

**Think it through:**
- Follow the three steps of `Upper`. What type of value does the channel carry?
- What do you `range` over here: a slice or a channel?

### 2. Parse

**Goal:** `Parse(in)` turns each line from `in` into an `Entry`.

`Level` is the text before the first space; `Message` is everything after
it. A line with no space is all `Level`, with `Message` `""`.

```go
// in carries "ERROR disk full", "HEARTBEAT"
Parse(in)
// carries Entry{"ERROR", "disk full"}, Entry{"HEARTBEAT", ""}, then is closed
```

**Think it through:**
- This channel carries `Entry`, not `string`. Which type goes in `make`?
- What separator do you cut on? Which of the three results do you need?

### 3. Filter by level

**Goal:** `OnlyLevel(in, level)` passes on only the entries whose `Level` equals `level`, in order.

Entries that don't match are not sent at all. If none match, the channel
is just closed.

```go
lines := []string{"INFO started", "ERROR disk full", "WARN slow", "ERROR timeout"}
OnlyLevel(Parse(Generate(lines)), "ERROR")
// carries Entry{"ERROR", "disk full"}, Entry{"ERROR", "timeout"}, then is closed
```

**Think it through:**
- You compare a field of each entry with the `level` parameter. Which field?
- The send only happens sometimes. What wraps it?

## Learn more

- [Go blog: Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines)
- [strings.Cut](https://pkg.go.dev/strings#Cut)
- [Go by Example: Channel directions](https://gobyexample.com/channel-directions)
