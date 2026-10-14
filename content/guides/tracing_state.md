# State and tracing code by hand

**In one sentence:** a program's *state* is the current value of every variable plus the line that runs next, and *tracing* means stepping through the code yourself, writing that state down after every line.

**Why it matters:** almost every bug is a variable holding a value you didn't expect. A rollout script that skips the last server, a retry loop that tries once too often, a disk report that says 0 GB: a trace shows you the exact line where the numbers go wrong, instead of guessing and re-running.

**You'll learn:**
- What state is, and why "which line runs next" is part of it
- How to build a trace table for assignments, `if`/`else` and loops
- How to spot off-by-one errors before they reach production
- How to describe a loop with one sentence that stays true (a loop invariant)
- Why tracing beats guessing, every time

## State is a snapshot

Imagine you could pause a running program, like freezing a video. What would you see?

- Every variable, each holding one value right now.
- A finger pointing at the next line to run.

That snapshot is the program's *state*. Running a program is just moving from one state to the next, one line at a time.

| What | Value right now |
|---|---|
| `replicas` | `4` |
| `ready` | `2` |
| next line | 10 |

Every line of code does one of three things to the state:

1. **Reads** it: `fmt.Println(ready)` looks at `ready` and changes nothing.
2. **Changes** it: `ready = 3` puts a new value in `ready`.
3. **Chooses the next line**: `if`, `switch` and `for` decide where the finger goes.

The *How a program runs* guide explains why lines run in order. This guide is about keeping track of the values while they do.

### `=` means "replace", not "is equal"

In maths, `x = x + 1` is nonsense. In Go it's an instruction with three steps:

1. Read the **current** value of `x`.
2. Add 1.
3. Store the result back in `x`, throwing the old value away.

A variable holds exactly one value at a time. Comparing two values is `==`, a different operator (see *Booleans*).

## Tracing assignments

Here is a small program about pods in a deployment:

```go
package main

import "fmt"

func main() {
	replicas := 3
	ready := 0
	ready = ready + 2
	replicas = replicas + 1
	ready = replicas - 1
	fmt.Println(replicas, ready)
}
```

To trace it, make a *trace table*: one column per variable, one row per line that runs. Fill each row with the state **after** that line.

| Line | Code | `replicas` | `ready` |
|---|---|---|---|
| 6 | `replicas := 3` | 3 | (none yet) |
| 7 | `ready := 0` | 3 | 0 |
| 8 | `ready = ready + 2` | 3 | 2 |
| 9 | `replicas = replicas + 1` | 4 | 2 |
| 10 | `ready = replicas - 1` | 4 | 3 |

Line 11 prints the last row:

```text
4 3
```

**The trap is line 10.** It uses `replicas` as it is *now* (4), not as it was when the program started (3). Code doesn't happen all at once. Each line sees the state left by the line before it.

> **Tip:** write the full state on every row, even the values that didn't change. It costs a few seconds and stops you mixing up rows.

## Tracing if and else

An `if` doesn't change any variable by itself. It moves the finger. Add a column for the condition, so you can see which way it went.

```go
package main

import "fmt"

func main() {
	usedPct := 91
	status := "ok"
	if usedPct >= 90 {
		status = "critical"
	} else if usedPct >= 75 {
		status = "warn"
	}
	fmt.Println(status)
}
```

| Line | Code | Condition | `status` |
|---|---|---|---|
| 6 | `usedPct := 91` | | (none yet) |
| 7 | `status := "ok"` | | `"ok"` |
| 8 | `if usedPct >= 90` | 91 >= 90 is **true** | `"ok"` |
| 9 | `status = "critical"` | | `"critical"` |
| 13 | `fmt.Println(status)` | | prints `critical` |

The table keeps the quotes on `"ok"` to remind you that `status` holds text. The quotes aren't printed: the output is just `critical`.

**Notice that lines 10 and 11 never ran.** Once one branch of an `if`/`else if` chain matches, Go skips the rest and jumps past the closing `}`.

### One trace per input

A trace follows **one** set of inputs. To understand the code, trace a few:

| `usedPct` | Line 8 check | Line 10 check | Prints |
|---|---|---|---|
| 91 | true | (skipped) | `critical` |
| 80 | false | true | `warn` |
| 50 | false | false | `ok` |

Which inputs should you pick? The interesting ones sit on the edges: 89 and 90, 74 and 75. These are *boundary values*, where `>` versus `>=` changes the answer. Bugs love boundaries.

## Tracing loops

A `for` loop moves the finger backwards, so the same lines run many times. First, know the order its parts run in:

```text
for disk := 1; disk <= 3; disk++ {
    ────┬────  ─────┬────  ──┬───
        │           │        └─ 4. after each pass: add 1, go to 2
        │           └─ 2. before each pass: keep going?
        └─ 1. runs once, at the very start
    3. the body: runs only when step 2 said true
}
```

One run of the body is a *pass* (also called an *iteration*). The check in step 2 runs **one more time than the body**: the last check is the one that says "stop".

### Worked trace: total disk usage

A script asks a server how full each of its three disks is and adds them up. The function `diskUsedGB` stands in for the real question. It uses a `switch`, which picks the `case` that matches `disk` and runs only that one (you'll write your own in *Switch: size, method, latency*).

```go
package main

import "fmt"

// diskUsedGB pretends to ask a server how full each disk is.
func diskUsedGB(disk int) int {
	switch disk {
	case 1:
		return 120
	case 2:
		return 45
	case 3:
		return 300
	}
	return 0
}

func main() {
	total := 0
	for disk := 1; disk <= 3; disk++ {
		total = total + diskUsedGB(disk)
	}
	fmt.Println("total GB:", total)
}
```

For loops, use one row per **check** of the condition:

| Step | `disk` | `disk <= 3`? | `total` after the body |
|---|---|---|---|
| before the loop | (none yet) | | 0 |
| pass 1 | 1 | true | 0 + 120 = 120 |
| pass 2 | 2 | true | 120 + 45 = 165 |
| pass 3 | 3 | true | 165 + 300 = 465 |
| final check | 4 | **false** | 465, loop ends |

The program prints:

```text
total GB: 465
```

Two things the trace makes obvious:

- **`total` starts at 0 outside the loop.** If you wrote `total := 0` *inside* the body, it would reset every pass and you'd print only the last disk.
- **`disk` reaches 4.** The body ran 3 times, the check ran 4 times. That fourth check is where off-by-one bugs live.

## Off-by-one errors

An *off-by-one error* is a loop that runs one time too many or too few. It's one of the most common bugs in all of programming. Here is a rollout that should deploy to 3 servers:

```go
package main

import "fmt"

func main() {
	servers := 3
	for i := 0; i <= servers; i++ {
		fmt.Println("deploying to server", i)
	}
}
```

```text
deploying to server 0
deploying to server 1
deploying to server 2
deploying to server 3
```

Four deploys for three servers. Trace the checks and you'll see why: `i` takes 0, 1, 2 **and** 3 before `3 <= 3` finally turns into `4 <= 3`. Starting at 0 *and* using `<=` adds one pass.

Two correct versions:

| Loop header | Values of `i` | Passes |
|---|---|---|
| `i := 0; i < n; i++` | 0, 1, …, n-1 | n |
| `i := 1; i <= n; i++` | 1, 2, …, n | n |
| `i := 0; i <= n; i++` | 0, 1, …, n | **n+1**, usually a bug |

Pick one style and use it everywhere. Go code mostly starts at 0 and uses `<`, because positions in lists start at 0 (see the *Collections* module).

### Fenceposts

A fence 3 sections long needs 4 posts. The same shape hides in ops work:

- **Retries:** "try, then retry 3 times" is 4 attempts, not 3.
- **Rolling deploys:** 5 servers updated one by one have 4 pauses *between* them.
- **Time ranges:** from 09:00 to 17:00 inclusive, hourly, is 9 checks, not 8.

This is called a *fencepost error*. When you count things, ask: am I counting the posts or the gaps?

**The habit that catches them:** trace only the first pass, the last pass and the final check. Those three rows find almost every off-by-one bug. You'll hunt real ones in *For loops: fix the off-by-one bugs*.

## Loop invariants: one sentence that stays true

Tracing every pass gets slow when a loop runs 10,000 times. A better tool for big loops is a sentence that's true at the top of **every** pass. That sentence is a *loop invariant*.

For the disk loop:

> At the top of each pass, `total` is the sum of every disk **before** `disk`.

Check it against the trace:

| When | Disks before `disk` | `total` | True? |
|---|---|---|---|
| pass 1 (`disk` is 1) | none | 0 | yes |
| pass 2 (`disk` is 2) | disk 1 | 120 | yes |
| pass 3 (`disk` is 3) | disks 1–2 | 165 | yes |
| final check (`disk` is 4) | disks 1–3 | 465 | yes |

At the final check, the sentence says "total is the sum of disks 1 to 3". That's exactly the answer you wanted. **If the invariant holds at the start, each pass keeps it true, and the loop stops at the right place, the loop is correct.**

You don't need to write invariants down formally. Saying one out loud is a quick check that you know what the loop is for. If you can't finish the sentence, you don't understand the loop yet, and that's a good moment to trace it.

Ops engineers use the same idea without the name. During a rolling deploy: "every server before the current one is on the new version, every server after it is on the old one". If that's ever false, something has gone wrong.

## Why tracing beats guessing

When a result is wrong, the tempting move is to change something and re-run. Sometimes the output looks right afterwards, and you never learn why. That fix often hides the bug instead of removing it.

Tracing works differently:

1. Add an **expected** column to your trace table: what each value *should* be.
2. Trace with the smallest input that shows the bug.
3. Find the **first row** where actual and expected differ.

That first wrong row is where the bug is. Every row before it was fine, so the cause is on that line or in a value it used.

| Guessing | Tracing |
|---|---|
| Change, re-run, hope | Find the line, then change it |
| Can "fix" by accident | Each fix has a reason |
| You learn nothing | You learn how the code really behaves |
| Fast for 1 minute, slow for an hour | Slow for 5 minutes, then done |

Tracing on paper is the skill underneath print debugging. A line like `fmt.Println("DEBUG total:", total)` is just a trace table the computer fills in for you. The *Debugging with a method* guide builds on that.

## Practice trace

Your turn. A monitor checks a service's `/healthz` endpoint four times and counts failures. Fill in the table before reading the answer.

```go
package main

import "fmt"

// checkPassed pretends to probe /healthz. Attempts 2 and 3 fail.
func checkPassed(attempt int) bool {
	return attempt != 2 && attempt != 3
}

func main() {
	failed := 0
	for attempt := 1; attempt <= 4; attempt++ {
		if !checkPassed(attempt) {
			failed++
		}
	}
	fmt.Println("failed checks:", failed)
}
```

Two reminders before you start: `!` flips `true` and `false`, and `failed++` is short for `failed = failed + 1`.

| Step | `attempt` | `checkPassed(attempt)` | `failed` after the body |
|---|---|---|---|
| before the loop | | | ? |
| pass 1 | ? | ? | ? |
| pass 2 | ? | ? | ? |
| pass 3 | ? | ? | ? |
| pass 4 | ? | ? | ? |
| final check | ? | (not called) | ? |

### Answer to the practice trace

| Step | `attempt` | `checkPassed(attempt)` | `failed` after the body |
|---|---|---|---|
| before the loop | | | 0 |
| pass 1 | 1 | true | 0 |
| pass 2 | 2 | false | 1 |
| pass 3 | 3 | false | 2 |
| pass 4 | 4 | true | 2 |
| final check | 5 | (not called) | 2 |

It prints:

```text
failed checks: 2
```

So `failed++` runs only when a check fails. An invariant for this loop: "at the top of each pass, `failed` counts the failed checks before `attempt`".

## Common mistakes

- **Reading the code all at once.** Each line sees the state the line before it left. Use the new value, not the starting one.
- **Forgetting the final check.** The loop condition runs one more time than the body. Trace that row too.
- **Mixing up `=` and `==`.** `=` stores a value; `==` compares two values and gives `true` or `false`.
- **Tracing what you meant instead of what you wrote.** Read the actual characters: `<` and `<=` look alike and behave differently.
- **Skipping rows because they're "obvious".** The obvious row is usually where the bug hides.
- **Resetting inside the loop.** A counter declared inside the body starts from zero on every pass.

## Check yourself

1. After `a := 5`, then `b := a`, then `a = a + 1`, what are `a` and `b`?
2. How many times does the body of `for i := 2; i <= 6; i++` run, and what is `i` when the loop stops?
3. Trace this backoff loop. What does it print? `wait := 1`, `total := 0`, then `for attempt := 1; attempt <= 4; attempt++ { total = total + wait; wait = wait * 2 }`, then `fmt.Println(wait, total)`.
4. Code prints `critical` if `usedPct >= 90`, else `warn` if `usedPct >= 75`, else `ok`. Which values of `usedPct` would you choose to test it, and why?
5. A colleague says "the job retries 2 times". How many attempts is that in total?

## Answers

1. `a` is 6 and `b` is 5. `b := a` copies the value 5 at that moment. Changing `a` later doesn't reach back into `b`.
2. 5 times (i = 2, 3, 4, 5, 6). When the loop stops, `i` is 7: that's the check where `7 <= 6` is false.
3. It prints `16 15`. Passes: total 1, wait 2; total 3, wait 4; total 7, wait 8; total 15, wait 16. You'll write loops like this in *For loops: backoff and progress bars*.
4. The boundaries: 89 and 90, 74 and 75, plus one value well inside each range (like 50). Those are where a `>` written instead of `>=` gives a different answer.
5. 3 attempts: the first try plus 2 retries. That's a fencepost: retries are the gaps, attempts are the posts.

## Key terms

| Term | Meaning |
|---|---|
| state | The current value of every variable, plus which line runs next |
| trace | Stepping through code by hand, recording the state after each line |
| trace table | A table with one column per variable and one row per step |
| assignment | `=`: replace a variable's value with a new one |
| pass (iteration) | One run of a loop's body |
| boundary value | An input right at the edge where a condition flips, like 89 and 90 |
| off-by-one error | A loop or count that's one too many or one too few |
| fencepost error | An off-by-one from mixing up counting items and counting gaps |
| loop invariant | A sentence about the variables that's true at the top of every pass |

## Watch

- [86. CAMBRIDGE IGCSE (0478-0984) 7 Trace tables](https://www.youtube.com/watch?v=-7_kyY88Y9Y) · Craig'n'Dave · 6 min: a short walk-through of filling in a trace table row by row, the same technique as this guide.
- [Loops - CS50 Shorts](https://www.youtube.com/watch?v=WgX8e_O7eG8) · CS50 · 9 min: Doug Lloyd shows the order a `for` loop's parts run in; the code is C, but its `for` loop works like Go's.
- [How to Read Complex Code? • Felienne Hermans • GOTO 2021](https://www.youtube.com/watch?v=az-MX_M11lg) · GOTO Conferences · 33 min: why reading code is hard work for your short-term memory, and techniques that make it easier; good for when a trace feels like too much to hold in your head.

## Go deeper

- [Off-by-one error (Wikipedia)](https://en.wikipedia.org/wiki/Off-by-one_error): the fencepost picture and the `<` vs `<=` loop, explained with diagrams.
- [Loop invariant (Wikipedia)](https://en.wikipedia.org/wiki/Loop_invariant): the informal example at the top shows invariants as plain comments in a loop.
- [State (Wikipedia)](https://en.wikipedia.org/wiki/State_(computer_science)): how "state" is used across computing, from programs to servers.
- [A Tour of Go: Flow control](https://go.dev/tour/flowcontrol/1): run and change `for` and `if` examples in the browser, then trace them.
