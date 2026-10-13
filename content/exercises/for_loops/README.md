# For loops: backoff and progress bars

Ops work is full of repetition: retry a request, check every host, print a
progress bar. In Go, every kind of repetition uses one keyword: `for`.

## Learn

A *loop* runs the same block of code again and again. Go has only one loop
keyword, `for`, but you can write it in a few shapes.

### The three-part loop

```go
total := 0
for i := 1; i <= 3; i++ {
	total += i
}
// total is 6 (1 + 2 + 3)
```

| Part | Example | Runs |
|---|---|---|
| init | `i := 1` | once, before the loop starts |
| condition | `i <= 3` | before every round; the loop stops when it is `false` |
| post | `i++` | after every round (`i++` means `i = i + 1`) |

The variable `i` only exists inside the loop. `total += i` is short for
`total = total + i`. There is also `*=`: `x *= 2` doubles `x`.

### Counting with range

When you just want "do this `n` times", use `range` over a number (Go 1.22
and newer):

```go
for i := range 3 {
	fmt.Println(i) // prints 0, then 1, then 2
}
```

`i` goes from `0` up to `n-1`. If `n` is `0`, the body never runs. If you
don't need `i`, write `for range 3 { ... }`.

### The condition-only loop

Leave out init and post and you get a loop that runs *while* a condition is
true. Other languages call this a `while` loop:

```go
delay := 1
for delay < 100 {
	delay *= 2
}
// delay is 128
```

Make sure something in the body changes the condition, or the loop never ends.

### Forever, break and continue

`for { ... }` loops forever until you leave it. `break` jumps out of the loop
immediately. `continue` skips the rest of this round and starts the next one.
A `return` inside a loop also ends it, because it leaves the whole function.

```go
for {
	if done() {
		break
	}
}
```

### Building a string in a loop

To grow a string one piece at a time, use a `strings.Builder` from the
`strings` package (`import "strings"`, as in *Strings*). Declare one, call
`WriteString` on it for each piece, and call `String` at the end to get the
result:

```go
var b strings.Builder // an empty builder, ready to use
for range 3 {
	b.WriteString("ab") // add "ab" to the end
}
line := b.String() // "ababab"
```

The dot in `b.WriteString(...)` calls a function that belongs to `b`. You'll
learn how that works in *Methods*; for now, just copy this pattern. You could
also write `line += "ab"` in the loop, but that copies the whole string every
round, so Go tools suggest a builder instead. For the same reason, give each
piece its own `WriteString` call (`b.WriteString(a)` then `b.WriteString(c)`)
rather than `b.WriteString(a + c)`.

## Your task

Your deploy tool retries failed requests with *exponential backoff*: it waits
`base` seconds before the first retry, then twice as long before each next
retry (`base`, `2*base`, `4*base`, ...). It also prints a progress bar.

### 1. Add up the backoff delays

Change `TotalBackoff(attempts, baseSeconds int) int` so it returns the sum of
the first `attempts` delays. Zero attempts means zero seconds.

```go
TotalBackoff(4, 1) // => 15  (1 + 2 + 4 + 8)
TotalBackoff(3, 5) // => 35  (5 + 10 + 20)
TotalBackoff(0, 5) // => 0
```

### 2. Draw a progress bar

Change `ProgressBar(done, total int) string` so it returns `[`, then `done`
`#` characters, then `total-done` `-` characters, then `]`. Build it with
loops (for example `for range done`) and a `strings.Builder`. Add
`import "strings"` below the `package` line. You can assume
`0 <= done <= total`.

```go
ProgressBar(2, 5) // => "[##---]"
ProgressBar(0, 3) // => "[---]"
ProgressBar(0, 0) // => "[]"
```

### 3. Count the retries that fit in a budget

Change `RetriesWithin(budgetSeconds, baseSeconds int) int`. Delays double as
in task 1. Return how many retries you can make before the total waiting time
would go *over* `budgetSeconds`. A total exactly equal to the budget still
fits. Use a condition-only loop. `baseSeconds` is always at least 1.

```go
RetriesWithin(10, 1) // => 3  (1+2+4 = 7; adding 8 would make 15, too much)
RetriesWithin(7, 1)  // => 3  (1+2+4 = 7 fits exactly)
RetriesWithin(4, 5)  // => 0  (even the first 5s delay is too long)
```

## Learn more

- [A Tour of Go: For](https://go.dev/tour/flowcontrol/1)
- [Go by Example: For](https://gobyexample.com/for)
- [Go by Example: Range over integers](https://gobyexample.com/range-over-built-in-types)
