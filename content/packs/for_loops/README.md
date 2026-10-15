# For loops: backoff and progress bars

Ops work is full of repetition: retry a request, check every host, print a
progress bar. In Go, every kind of repetition uses one keyword: `for`.

**New here:** `for` (three-part, `range` over a number, condition-only, forever), `break`, `continue`, `+=` and `*=`, `strings.Builder`.

## Learn

A *loop* runs the same block of code again and again. Each run is called a
*round* (or *iteration*). Go has only one loop keyword, `for`, written in a
few shapes.

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

`i` only exists inside the loop. `total += i` is short for
`total = total + i`. `x *= 2` is short for `x = x * 2`.

> **Why is it written like this?** Go requires the semicolons between the
> three parts, and the `{` on the same line as `for`. There are no `( )`
> around the parts; gofmt removes them. The name `i` is convention for a
> counter; you may pick any name.

### Counting with range

To "do this `n` times", use `range` over a number (Go 1.22 and newer):

```go
for i := range 3 {
	fmt.Println(i) // prints 0, then 1, then 2
}
```

`i` goes from `0` up to `n-1`. If `n` is `0`, the body never runs. If you
don't need `i`, leave it out: `for range 3 { ... }`.

> **Why is it written like this?** Go rejects a variable you declare but
> never use. That's why `for range n` without `i` exists.

### The condition-only loop

Leave out init and post, and the loop runs *while* a condition is true.
Other languages call this a `while` loop:

```go
size := 1
for size < 100 {
	size *= 3
}
// size is 243
```

Something in the body must change the condition, or the loop never ends.

### Forever, break and continue

| Word | What it does |
|---|---|
| `for { ... }` | loops forever, until something leaves it |
| `break` | jumps out of the loop right now |
| `continue` | skips the rest of this round, starts the next |
| `return` | leaves the whole function, so the loop ends too |

```go
for {
	if done() {
		break
	}
}
```

### Building a string in a loop

To grow a string piece by piece, use a `strings.Builder` from the `strings`
package (`import "strings"`, as in *Strings*):

```go
var b strings.Builder // an empty builder, ready to use
for range 3 {
	b.WriteString("ab") // add "ab" to the end
}
line := b.String() // "ababab"
```

The dot in `b.WriteString(...)` calls a function that belongs to `b`. You'll
learn how that works in *Methods*; for now, copy this pattern.

> **Why is it written like this?** `line += "ab"` in a loop also works, but it
> copies the whole string every round, so Go tools suggest a builder. That's
> convention, not a rule. For the same reason, give each piece its own
> `WriteString` call instead of joining pieces with `+` first.

## Your task

Your deploy tool retries failed requests with *exponential backoff*: it waits
`base` seconds before the first retry, then twice as long before each later
one (`base`, `2*base`, `4*base`, ...). It also prints a progress bar. Each
starter function returns a placeholder; replace it.

### 1. Add up the backoff delays

**Goal:** make `TotalBackoff(attempts, baseSeconds)` return the sum of the first `attempts` delays.

The first delay is `baseSeconds`; each later delay is double the one before.
Zero attempts means `0`.

```go
TotalBackoff(4, 1) // => 15  (1 + 2 + 4 + 8)
TotalBackoff(3, 5) // => 35  (5 + 10 + 20)
TotalBackoff(1, 2) // => 2
TotalBackoff(0, 5) // => 0
```

**Think it through:**
- Which two numbers change every round? Each needs its own variable.
- What does each start at, and how does each change?

### 2. Draw a progress bar

**Goal:** make `ProgressBar(done, total)` return a bar like `"[##---]"`.

The bar is four pieces, in order:

1. the fixed text `[`
2. `done` copies of `#`
3. `total-done` copies of `-`
4. the fixed text `]`

Build it with loops and a `strings.Builder`. Add `import "strings"` below the
`package` line. You can assume `0 <= done <= total`.

```go
ProgressBar(2, 5) // => "[##---]"
ProgressBar(4, 4) // => "[####]"
ProgressBar(0, 3) // => "[---]"
ProgressBar(0, 0) // => "[]"
```

**Think it through:**
- Which pieces repeat, and how many times does each repeat?
- Which pieces are written only once?

### 3. Count the retries that fit in a budget

**Goal:** make `RetriesWithin(budgetSeconds, baseSeconds)` return how many retries fit in the time budget.

Delays double as in task 1. Keep adding retries while the total wait stays at
or under `budgetSeconds`; stop before the total would go *over* it. A total
exactly equal to the budget still fits. Use a condition-only loop.
`baseSeconds` is always at least 1.

```go
RetriesWithin(10, 1)  // => 3  (1+2+4 = 7; adding 8 would make 15, too much)
RetriesWithin(7, 1)   // => 3  (1+2+4 = 7 fits exactly)
RetriesWithin(100, 10) // => 3  (10+20+40 = 70; adding 80 is too much)
RetriesWithin(5, 5)   // => 1
RetriesWithin(4, 5)   // => 0  (even the first 5s delay is too long)
```

**Think it through:**
- Which numbers do you need to track (time used, next delay, count)?
- What question decides whether one more retry fits? That's your loop condition.

## Learn more

- [A Tour of Go: For](https://go.dev/tour/flowcontrol/1)
- [Go by Example: For](https://gobyexample.com/for)
- [Go by Example: Range over integers](https://gobyexample.com/range-over-built-in-types)
- [strings.Builder](https://pkg.go.dev/strings#Builder)
