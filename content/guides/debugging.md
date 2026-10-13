# Debugging with a method

**In one sentence:** debugging is running small experiments: observe what's wrong, guess one cause, predict what you'd see if the guess were right, then check, until only one guess survives.

**Why it matters:** changing things at random until a bug goes away is slow, and at 3 a.m. during an outage it makes things worse. A method gets you from "it's broken" to "line 6 divides by zero" in a predictable number of steps, and the same method works on a failing test, a crashing CLI and a misbehaving production service.

**You'll learn:**
- The debugging loop: observe, hypothesise, predict, test
- Why you reproduce a bug first and then shrink it
- How to read compiler errors, test failures and panics
- Print debugging that actually shows you the problem
- Binary search over code and commits, rubber ducking and a bug journal
- When to stop and take a break, and how incidents use the same loop

## Debugging is the scientific method

A bug is a gap between what you **expected** and what **happened**. You close it the way a scientist would:

```text
   ┌─────────────┐     ┌────────────────┐     ┌─────────────┐
   │ 1. Observe  │ ──▶ │ 2. Hypothesise │ ──▶ │ 3. Predict  │
   └─────────────┘     └────────────────┘     └──────┬──────┘
          ▲                                          │
          │                ┌────────────┐            │
          └──── wrong ──── │  4. Test   │ ◀──────────┘
                           └─────┬──────┘
                                 │ right
                                 ▼
                     fix it, re-run, confirm
```

| Step | Question | Example: disk report says 0% |
|---|---|---|
| 1. Observe | What exactly happened? | `Percent(45, 60)` returned 0, not 75 |
| 2. Hypothesise | What one thing could cause that? | `45 / 60` is done in whole numbers, giving 0 |
| 3. Predict | If that's true, what would I see? | Printing `usedGB / totalGB` shows `0` |
| 4. Test | Look. Was the prediction right? | It prints `0`: hypothesis confirmed |

Two rules make the loop work:

- **A hypothesis must be checkable.** "Something's weird with the maths" can't be tested. "45 / 60 gives 0" can.
- **Change one thing at a time.** If you change three things and the bug goes away, you don't know which one fixed it, or what the other two broke.

A *wrong* prediction isn't wasted. It crosses off a cause, and that's progress.

## Step zero: reproduce it

You can't fix what you can't make happen. Before changing any code, find a way to trigger the bug **on purpose**, every time.

Write down exactly what you ran:

- The command and its flags
- The input: which file, which host, which values
- The version of the code, and anything about the environment that might matter

"Works on my machine" usually means two machines have different *state*: a different config file, a different version, a different environment variable. Reproducing forces you to find that difference.

In ops, reproducing might mean running the same `curl` that failed, against the same endpoint, with the same headers. If a bug only shows up sometimes, that's a clue too: it often depends on timing or on something outside your program (the *Concurrency* module covers timing bugs). In the *Testing* module you'll turn reproductions into tests that fail until the bug is fixed, starting with *Practice: reproduce a bug report*.

## Read the whole error

Most beginners read the last line of an error, panic, and start editing. Read **all** of it first. Go gives you three kinds, and each tells you different things.

| Kind | When it appears | Looks like | Look first at |
|---|---|---|---|
| compiler error | before anything runs | `./main.go:7:14: message` | that line and column |
| test failure | tests ran, an answer was wrong | `--- FAIL: TestName` | the inputs, got and want |
| panic | program crashed while running | `panic:` + stack trace | the first frame in your code |

### Compiler errors

```go
package main

import "fmt"

func main() {
	port := 8080
	fmt.Println("listening on :" + port)
}
```

```text
./main.go:7:14: invalid operation: "listening on :" + port (mismatched types untyped string and int)
```

Split it into parts:

| Part | Meaning |
|---|---|
| `./main.go` | The file |
| `7` | The line |
| `14` | The column: the character where the problem starts (a tab counts as one) |
| `invalid operation` | What Go refused to do |
| `mismatched types untyped string and int` | Why: you're gluing text to a number |

Fix: use a format verb, `fmt.Printf("listening on :%d\n", port)`, so `fmt` turns the number into text for you. When there are several errors, **fix the first one first**: later ones are often caused by it. *Reading compiler errors* practises exactly this.

### Test failures

The code compiled, the tests ran, and an answer was wrong. A function `Percent` should say how full a disk is; its test calls `Percent(45, 60)` and expects 75. Running `go test` in its folder prints:

```text
--- FAIL: TestPercent (0.00s)
    disk_test.go:8: Percent(45, 60) = 0, want 75
FAIL
exit status 1
FAIL	disk	0.001s
```

| Part | Meaning |
|---|---|
| `--- FAIL: TestPercent` | Which test failed |
| `disk_test.go:8` | The line **in the test** that reported it |
| `Percent(45, 60)` | The inputs that trigger the bug |
| `= 0, want 75` | What you got, and what was expected |

That's a complete reproduction, handed to you: inputs, expected and actual. The bug is almost never in the test line. It's in the function the test called, `Percent`, with those inputs.

### Panics

A panic is a crash while the program runs. It prints the message, then the stack: who called whom (the *Functions and decomposition* guide explains the call stack).

```text
web-1: 75% full
panic: runtime error: integer divide by zero

goroutine 1 [running]:
main.diskPercent(...)
	/tmp/diskreport/main.go:6
main.report({0x49c1ac?, 0x24bdeea0cea8?}, 0x2d?, 0x3c?)
	/tmp/diskreport/main.go:10 +0xd0
main.main()
	/tmp/diskreport/main.go:16 +0x45
exit status 2
```

How to read it:

1. **The `panic:` line says what.** Integer divide by zero.
2. **Ignore what you don't need yet.** `goroutine 1 [running]` matters in the *Concurrency* module. The hex numbers in brackets and the `+0xd0` are for tools.
3. **Each frame is two lines:** a function, then `file:line`. The top frame is where it crashed: `main.go:6`.
4. **Walk down to find where the bad value came from.** Line 6 divides correctly; the 0 was passed in from line 10, which got it from line 16.
5. **The program printed `web-1` first.** Output before a panic tells you how far it got.

`exit status 2` is how `go run` tells you the program ended with an error code (more in *CLI tools and the OS*).

## Shrink the problem

A bug in a 500-host run, with a 2,000-line config, has hundreds of possible causes. A bug with one host and three lines of config has a few. **Make the failing case as small as you can while it still fails.**

- Fewer inputs: one host, one log line, one request.
- Less code: comment out or delete parts. If the bug vanishes, the last thing you removed is involved. Put it back and remove something else.
- Simpler values: `Percent(45, 60)` is easier to reason about than `Percent(48213, 64284)`.

The smallest failing case is also what you'd put in a bug report or a test.

## Print debugging: make the state visible

Tracing by hand (the *State and tracing code by hand* guide) works well for ten lines. For more, let the computer fill in the trace table: print the values you care about.

Back to the failing `Percent`. The hypothesis is "the division gives 0". Predict, then print:

```go
package disk

import "fmt"

// Percent returns how full a disk is, from 0 to 100.
func Percent(usedGB int, totalGB int) int {
	fmt.Println("DEBUG usedGB:", usedGB, "totalGB:", totalGB)
	fmt.Println("DEBUG usedGB / totalGB:", usedGB/totalGB)
	return usedGB / totalGB * 100
}
```

Running the test again prints:

```text
DEBUG usedGB: 45 totalGB: 60
DEBUG usedGB / totalGB: 0
--- FAIL: TestPercent (0.00s)
    disk_test.go:8: Percent(45, 60) = 0, want 75
FAIL
exit status 1
FAIL	disk	0.001s
```

The inputs are right; the division is the first wrong value. Whole-number division drops the fraction, so `0 * 100` is still 0. The fix is to multiply first: `return usedGB * 100 / totalGB`. Delete the `DEBUG` lines, re-run, and the test passes.

### Make your prints useful

- **Label every print.** `DEBUG total: 465` beats a bare `465` when there are ten of them.
- **Print the in-between values**, not only the final answer. You're looking for the first wrong one.
- **Use `%q` for text.** It shows quotes around strings, so spaces and empty strings become visible.

That last one catches a lot of real bugs. A host name read from a config file has a trailing space:

```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	host := "db-1 " // as it came out of a config file
	fmt.Println(host, host == "db-1")
	fmt.Printf("%q\n", host)
	host = strings.TrimSpace(host)
	fmt.Printf("%q %v\n", host, host == "db-1")
}
```

```text
db-1  false
"db-1 "
"db-1" true
```

The first line *looks* fine: the extra space hides between `db-1` and `false`. `%q` shows `"db-1 "` with the space inside the quotes. `strings.TrimSpace` removes it, and the comparison becomes `true`.

**In tests,** use `t.Log` or `t.Logf` instead of `fmt.Println`. Their output is attached to the test and shown when it fails (or always, with `go test -v`). You'll write tests in the *Testing* module.

When you're done, search for `DEBUG` and remove every print. Leftover debug output in a CLI breaks the scripts that read it.

## Check your assumptions one at a time

When the code "can't possibly" be wrong, one of your assumptions is. List them, then check each, cheapest first.

| Assumption | Cheap check |
|---|---|
| My code is running at all | Print something on the first line |
| I'm running the version I just edited | Did you save? Rebuild? Print a word you just added |
| The function gets the values I think | Print the parameters with `%q` at the top of the body |
| The config file is the one I edited | Print the path it actually loaded |
| The server runs the new release | Check the version in its startup log or status endpoint |

"Is the pod even running the new image?" is one of the most useful questions in ops. Half of all "my fix didn't work" moments are an old binary, an old container, or an unsaved file.

## Binary search: halve the haystack

If you don't know *where* things go wrong, don't check every line in order. Check the **middle**.

- Print the state halfway through. If it's already wrong, the bug is in the first half; if it's right, the second half.
- Repeat on the half that's wrong.

Each check halves what's left. 1,000 lines take about 10 checks, not 1,000. The *Algorithms and Big-O* guide explains why halving is so powerful.

### Over commits: git bisect

"It worked last week, it's broken today, and there are 64 commits in between." `git bisect` runs the same binary search over your history. You mark one commit bad and one good; git checks out the middle; you test it and say which it was.

```bash
git bisect start
git bisect bad              # the commit you're on is broken
git bisect good v1.4.0      # this release was fine
# git checks out a commit halfway between. Test it, then:
git bisect good             # or: git bisect bad
# repeat until git prints the first bad commit, then:
git bisect reset            # go back to where you started
```

64 commits take at most 6 tests. If a command can decide good or bad on its own, `git bisect run go test ./...` does the whole search for you.

The same trick works on a pipeline: if a report is wrong, check the data in the **middle** stage before checking every stage.

## Explain it to a rubber duck

Put an object on your desk (a rubber duck is traditional) and explain your code to it, **line by line, out loud**: what each line does and what each variable holds.

It sounds silly. It works because it forces you to trace instead of skim. Somewhere you'll hear yourself say "and then this adds the disk to the total..." and notice the line doesn't say that. This is *rubber duck debugging*.

If the duck doesn't help, ask a person, and ask well: what you expected, what happened, the exact error, and what you've already tried. Writing that down often answers the question before you send it.

## Keep a bug journal

Keep one file of short notes, one entry per bug that took you more than a few minutes:

```text
2026-10-13  TestPercent: got 0, want 75
Tried:  printed usedGB/totalGB -> 0
Cause:  whole-number division dropped the fraction before * 100
Fix:    multiply first: usedGB * 100 / totalGB
Lesson: with int maths, multiply before you divide
```

Two reasons. Your bugs repeat: after a month you'll see your own patterns (quotes, off-by-one, unsaved files) and check those first. And it's practice for incident notes: on-call engineers write down every hypothesis and test as they go, so nobody repeats a dead end.

## Know when to stop

Signs that you've stopped debugging and started flailing:

- You're re-running the same thing hoping for a different result.
- You're changing things without a hypothesis.
- You can't say, in one sentence, what you're testing right now.
- 25 minutes have gone by with nothing crossed off.

When that happens, **stop on purpose.** Write down where you are in your journal (what you know, what you've ruled out), then step away: walk, drink some water, work on something else. Coming back fresh often makes the answer obvious, because you'll read the code instead of what you remember it says.

A timer helps. Set it before you start, and when it rings, either write down a new hypothesis or take the break. Breaks are part of the method, not a failure of it.

## Incidents use the same loop

Debugging a production incident is the same loop with higher stakes and different tools.

| Debugging step | During an incident |
|---|---|
| Reproduce | Which requests, hosts or regions fail? Can you trigger it with `curl`? |
| Read the whole error | The full log line, alert text and status codes, not just "500" |
| Observe | Dashboards: error rate, latency, recent deploys and config changes |
| Hypothesise and predict | "If it's the new release, only pods on v1.5 return errors" |
| Test | Compare v1.5 pods with v1.4 pods in the logs or metrics |
| Binary search | Which deploy, which config change, which hop in the chain |
| Journal | The incident timeline, later the postmortem |

Two ops-specific habits:

- **Ask what changed last.** Working systems tend to keep working until something changes them: a deploy, a config edit, a traffic shift. Recent changes are the first suspects.
- **Stop the bleeding first.** If users are hurting, roll back or fail over *before* you find the root cause. Keep the logs, then debug calmly.

## Common mistakes

- **Changing several things at once.** You won't know which change mattered.
- **Reading only part of the error.** The file, line, inputs and stack are all clues.
- **Fixing the symptom.** Adding `if usedGB == 45 { return 75 }` makes the test pass and the bug stays. Fix the cause.
- **Not confirming the fix.** Re-run the reproduction, and the other tests, after every fix.
- **Blaming Go, the library or the computer.** It's possible, but it's almost always your code. Check yours first.
- **Unlabelled prints.** Ten bare numbers in the output tell you nothing.
- **Running old code.** Unsaved file, old binary, old container image. Check before you debug.

## Check yourself

1. A health check says `db-1` is down, but `curl` from your laptop to `db-1` works. Give one hypothesis, a prediction, and a test.
2. In `./main.go:7:14: invalid operation: ...`, what do `7` and `14` mean, and which part of the message tells you the cause?
3. In the panic above, which frame do you look at first, and where did the zero come from?
4. Why didn't `fmt.Println(host, host == "db-1")` reveal the bug, and what did?
5. There are 128 commits between a good release and a broken one. At most how many tests does `git bisect` need?
6. You've tried six things in 40 minutes and none changed the result. What should you do next?

## Answers

1. One option: hypothesis "the check reads the host name from config with a trailing space". Prediction: printing the host with `%q` shows `"db-1 "`. Test: add the print and run the check. Another: hypothesis "the checker runs somewhere that can't reach `db-1`, unlike my laptop". Prediction: `curl` from the checker's machine fails. Test: run that `curl` there. Any answer with a checkable guess, a prediction and a way to look is right.
2. `7` is the line, `14` the column (character) where the problem starts. `mismatched types untyped string and int` tells you the cause: `+` can't join text and a number.
3. Start with the top frame, `main.diskPercent` at `main.go:6`, where it crashed. Then walk down: `report` at line 10 passed `totalGB` along, and `main` at line 16 called `report("web-2", 10, 0)`. The zero came from line 16.
4. `Println` prints the space, but next to the space before `false` it's invisible. `fmt.Printf("%q\n", host)` showed `"db-1 "`, with the space inside the quotes.
5. 7. Each test halves the range: 128, 64, 32, 16, 8, 4, 2, 1.
6. Stop. Write down what you know and what you've ruled out, then take a short break. When you come back, write one new hypothesis before you change anything, or explain the code to a duck or a colleague.

## Key terms

| Term | Meaning |
|---|---|
| bug | A gap between what you expected the program to do and what it did |
| reproduce | Make a bug happen on purpose, reliably, with known inputs |
| hypothesis | One checkable guess about the cause |
| prediction | What you'd see if the hypothesis were true |
| compiler error | A problem Go finds before running, reported as `file:line:column` |
| test failure | A test ran and an answer didn't match what it wanted |
| panic | A crash while running, printed with a stack trace |
| print debugging | Printing labelled values to see the program's state |
| binary search | Finding something by checking the middle and halving each time |
| git bisect | Git's binary search over commits to find the one that broke things |
| rubber duck debugging | Explaining code out loud, line by line, to find the gap |
| bug journal | Short notes on each bug: symptom, tries, cause, fix, lesson |

## Go deeper

- [Effective Troubleshooting (Google SRE book)](https://sre.google/sre-book/effective-troubleshooting/): the hypothesis loop applied to production systems, with a real case study.
- [A debugging manifesto (Julia Evans)](https://jvns.ca/blog/2022/12/08/a-debugging-manifesto/): eight short, encouraging rules, like "being stuck is temporary".
- [git-bisect documentation](https://git-scm.com/docs/git-bisect): every bisect command, including `git bisect run`.
- [Rubber duck debugging (Wikipedia)](https://en.wikipedia.org/wiki/Rubber_duck_debugging): where the name comes from and why explaining works.
