# Pipelines

A *pipeline* is a chain of *stages* connected by channels. Each stage
receives values, does one small job, and sends the results to the next stage,
just like `cat app.log | grep ERROR | cut -d' ' -f2-` in a shell. All the
stages run at the same time, each in its own goroutine.

## Learn

### A stage

Every stage has the same shape: take an input channel, return an output
channel, and do the work in a goroutine:

```go
// Upper sends every string from in, upper-cased.
func Upper(in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out) // when in is closed and drained, close out
		for s := range in {
			out <- strings.ToUpper(s)
		}
	}()
	return out // return straight away; the goroutine keeps working
}
```

- The stage **owns** `out`: it makes it, it is the only one that sends on it,
  and it closes it when its input runs out.
- It returns `<-chan string` (receive only), so callers can't send on it or
  close it by mistake.
- `defer close(out)` at the top of the goroutine means `out` is closed
  however the goroutine ends.

### Connecting stages

Each stage's output is the next stage's input:

```go
for line := range Upper(Generate(lines)) {
	fmt.Println(line)
}
```

```text
Generate ──chan string──▶ Parse ──chan Entry──▶ OnlyLevel ──chan Entry──▶ range
 goroutine                goroutine             goroutine
```

Values flow through one at a time. While `Parse` works on line 2,
`OnlyLevel` can already be filtering line 1. Closing also flows down the
chain: `Generate` closes its channel, so `Parse`'s `range` ends and `Parse`
closes its channel, and so on, until the final `range` loop ends.

### Splitting a line with strings.Cut

`strings.Cut(s, sep)` splits `s` around the **first** `sep`:

```go
before, after, found := strings.Cut("ERROR disk full", " ")
// before = "ERROR", after = "disk full", found = true

before, after, found = strings.Cut("HEARTBEAT", " ")
// before = "HEARTBEAT", after = "", found = false
```

## Your task

Build a three-stage pipeline that finds error lines in an application log.
Each stage returns its output channel straight away, does its work in a
goroutine, and closes its output channel when it has finished.

### 1. Generate

`Generate(lines)` returns a channel that carries every line, in order, and
is then closed.

```go
Generate([]string{"INFO started", "ERROR disk full"})
// carries "INFO started", "ERROR disk full", then is closed
```

### 2. Parse

`Parse(in)` turns each line from `in` into an `Entry`. `Level` is the text
before the first space, and `Message` is everything after it. A line with no
space is all `Level`, with an empty `Message`.

```go
// in carries "ERROR disk full", "HEARTBEAT"
Parse(in)
// carries Entry{"ERROR", "disk full"}, Entry{"HEARTBEAT", ""}, then is closed
```

### 3. Filter by level

`OnlyLevel(in, level)` passes on only the entries whose `Level` equals
`level`, in order.

```go
lines := []string{"INFO started", "ERROR disk full", "WARN slow", "ERROR timeout"}
OnlyLevel(Parse(Generate(lines)), "ERROR")
// carries Entry{"ERROR", "disk full"}, Entry{"ERROR", "timeout"}, then is closed
```

## Learn more

- [Go blog: Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines)
- [strings.Cut](https://pkg.go.dev/strings#Cut)
- [Go by Example: Channel directions](https://gobyexample.com/channel-directions)
