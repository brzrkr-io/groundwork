# Functions and decomposition

**In one sentence:** a function is a named, reusable step that takes inputs (its parameters), does one job, and hands back an output (its return value); *decomposition* is splitting a big task into steps like that.

**Why it matters:** a 300-line deploy script that does everything in one block is hard to read, impossible to test piece by piece, and miserable to fix during an incident. The same script split into `validate`, `build`, `push` and `rollout` tells you which step broke from the error message alone.

**You'll learn:**
- What a function is: a name, inputs, an output and a body
- The difference between defining a function and calling it
- How to tell the name, the parameters and the arguments apart
- Scope: which names you can use where
- The call stack, and how to read a stack trace
- Pure functions vs side effects, and how to split a task top-down

## A function is a named step

Think of a runbook. One page says "Check disk usage: divide used space by total space, times 100". Other pages just say "check disk usage on web-1". Nobody rewrites the steps each time; they refer to them by name.

A function is that page. You write the steps once, give them a name, and then use the name wherever you need the result.

```text
              ┌──────────────────────────┐
  45, 60 ───▶ │       diskPercent        │ ───▶ 75
  (inputs)    │  usedGB * 100 / totalGB  │      (output)
              └──────────────────────────┘
```

Every function has four parts:

| Part | Runbook version | Go version |
|---|---|---|
| name | "Check disk usage" | `diskPercent` |
| inputs | which disk, how big | *parameters* `usedGB`, `totalGB` |
| steps | the instructions | the *body* between `{` and `}` |
| output | "it's 75% full" | the *return value* |

## Defining vs calling

**Defining** a function writes the page of the runbook. Nothing happens yet. **Calling** a function runs it, with real values, and gives you its answer.

```go
package main

import "fmt"

func diskPercent(usedGB int, totalGB int) int {
	return usedGB * 100 / totalGB
}

func main() {
	fmt.Println(diskPercent(45, 60))
	fmt.Println(diskPercent(450, 500))
	root := 30
	size := 120
	fmt.Println(diskPercent(root, size))
}
```

```text
75
90
25
```

Step by step:

1. Lines 5–7 **define** `diskPercent`. Go reads them but runs nothing.
2. The program starts at `main`. Line 10 **calls** `diskPercent(45, 60)`.
3. Go jumps into the body. Inside it, `usedGB` is 45 and `totalGB` is 60.
4. `return` works out `45 * 100 / 60`, which is `75`, and jumps back to line 10.
5. The call `diskPercent(45, 60)` is **replaced by** its result, so line 10 becomes `fmt.Println(75)`.
6. Lines 11 and 14 do the same with other values. One definition, three calls.

The multiplication comes before the division on purpose. With whole numbers, `45 / 60` is `0`, so `45 / 60 * 100` would be `0` too (see the *Values and types* guide).

## Name, parameters, arguments

These three words get mixed up all the time. Here they are, labelled:

```text
DEFINITION: written once; runs nothing by itself

func diskPercent(usedGB int, totalGB int) int {
     └────┬────┘└───────────┬───────────┘ └┬┘
          │                 │              └─ RESULT TYPE: what comes back
          │                 └─ PARAMETERS: names for the inputs, each with
          │                    a type. You choose these names.
          └─ NAME: what callers write to run it
    return usedGB * 100 / totalGB
}

CALL: written wherever you need the answer

pct := diskPercent(45, 60)
       └────┬────┘└──┬───┘
            │        └─ ARGUMENTS: the values for this one call
            └─ NAME again: "run diskPercent now"
```

| Word | Where it appears | Example | What it is |
|---|---|---|---|
| name | after `func`, and in every call | `diskPercent` | How you refer to the function |
| parameter | inside `( )` of the **definition** | `usedGB int` | A variable that exists only inside the body |
| argument | inside `( )` of the **call** | `45`, `root` | The actual value handed in this time |
| return value | after `return` | `usedGB * 100 / totalGB` | What the call turns into |

A memory hook: **p**arameters are **p**laceholders, **a**rguments are **a**ctual values.

### Arguments match by position, not by name

Go fills the first parameter with the first argument, the second with the second, and so on. The names don't have to match: in `diskPercent(root, size)`, `usedGB` gets the value of `root` (30) and `totalGB` gets the value of `size` (120).

That also means Go can't save you from swapping them. `diskPercent(size, root)` compiles fine, because both are `int`, and returns `400`. You'll fix bugs like this in *Functions: swapped arguments and helpers*.

### Quotes make text, names make variables

Inside a function, use a parameter by writing its name, **without** quotes:

```go
func greeting(host string) string {
	return "deploying to " + host
}
```

`greeting("web-1")` returns `deploying to web-1`. If you wrote `"deploying to host"`, it would always say the word `host`: inside quotes, `host` is just four letters, not the variable. Rule: quotes around a value you type in (`"web-1"`), no quotes around a name (`host`).

### Parameters get a copy

When you call a function, each parameter gets a **copy** of its argument's value. If the body changes `usedGB`, the caller's `root` stays 30. The *Memory, pointers and sharing* guide covers what to do when you really want a function to change something that belongs to its caller.

## Scope: which names you can use where

A name exists from the line that creates it to the `}` that closes the block it was created in. That region is the name's *scope*. Outside it, Go has never heard of the name.

```go
package main

import "fmt"

func sumDisks() int {
	total := 0
	for disk := 1; disk <= 3; disk++ {
		total = total + 100
	}
	fmt.Println(disk)
	return total
}

func main() {
	fmt.Println(sumDisks())
	fmt.Println(total)
}
```

This doesn't compile:

```text
./main.go:10:14: undefined: disk
./main.go:16:14: undefined: total
```

Picture the blocks as nested boxes. Code can see names in its own box and in every box around it, never inside other boxes:

```text
┌ package main ─────────────────────────────────────────┐
│ sumDisks, main: visible everywhere in the package     │
│                                                       │
│  ┌ func sumDisks() ─────────────────────────────────┐ │
│  │ total (exists only inside sumDisks)              │ │
│  │  ┌ for disk := 1; disk <= 3; disk++ ───────────┐ │ │
│  │  │ disk  (exists only inside the loop)         │ │ │
│  │  └─────────────────────────────────────────────┘ │ │
│  │ fmt.Println(disk)   <- error: no disk here       │ │
│  └──────────────────────────────────────────────────┘ │
│                                                       │
│  ┌ func main() ─────────────────────────────────────┐ │
│  │ fmt.Println(total)  <- error: no total here      │ │
│  └──────────────────────────────────────────────────┘ │
└───────────────────────────────────────────────────────┘
```

**Scope is a feature.** Two functions can both have a variable called `total` and never clash. To move a value from one function to another, don't reach into it: **return** it from one and **pass** it as an argument to the other. `main` already does that right: `fmt.Println(sumDisks())`.

## The call stack

When `main` calls `report`, `main` doesn't vanish. It pauses on that line and waits for the answer. If `report` then calls `diskPercent`, `report` pauses too. Go keeps a pile of "who's waiting, and where": the *call stack*.

```go
package main

import "fmt"

func diskPercent(usedGB int, totalGB int) int {
	return usedGB * 100 / totalGB
}

func report(host string, usedGB int, totalGB int) {
	pct := diskPercent(usedGB, totalGB)
	fmt.Printf("%s: %d%% full\n", host, pct)
}

func main() {
	report("web-1", 45, 60)
	report("web-2", 10, 0)
}
```

While `diskPercent` runs during the second call, the stack looks like this:

```text
        ┌────────────────────────────────────────┐
top ──▶ │ diskPercent   running line 6           │
        ├────────────────────────────────────────┤
        │ report        paused at line 10        │
        ├────────────────────────────────────────┤
        │ main          paused at line 16        │
        └────────────────────────────────────────┘
```

Each box is a *stack frame*: one call in progress, with its own parameters and variables. A call adds a frame on top; a `return` removes it and the frame below carries on. That's why two calls to `report` don't mix up their `host` values: each call gets a fresh frame.

### Reading a stack trace

The second call divides by zero, which crashes the program. When Go crashes it prints the stack, so you can see who called whom:

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

Your folder and the hex numbers will differ. Read it like the diagram above, top to bottom:

| Lines | Meaning |
|---|---|
| `panic: ... divide by zero` | **What** went wrong |
| `main.diskPercent` + `main.go:6` | **Where**: the top frame, the line that crashed |
| `main.report` + `main.go:10` | Who called it, and from which line |
| `main.main` + `main.go:16` | Who called *that*, and from which line |

Line 6 is correct maths. The real question is "where did the zero come from?", and the stack answers it: follow the frames down to line 16, `report("web-2", 10, 0)`. **The line that crashes and the line that's wrong are often different frames.** The *Debugging with a method* guide shows how to use this, and the *Errors* module covers panics properly.

## Pure functions and side effects

`diskPercent` only computes an answer from its inputs. Give it 45 and 60 and you get 75, today, tomorrow, on any server. A function like that, whose result depends only on its arguments and which changes nothing else, is a *pure function*.

Anything else a function does is a *side effect*: printing, writing a file, sending a request, restarting a pod, changing a variable outside itself. `report` has a side effect: it prints.

| Function | Pure? | Why |
|---|---|---|
| `diskPercent(45, 60)` | yes | Same inputs, always 75, nothing else touched |
| `report("web-1", 45, 60)` | no | Prints to the screen |
| a function that deletes old log files | no | Changes the disk |
| a function that lists EC2 instances | no | Talks to the network; the answer changes over time |

Side effects aren't bad: they're the whole point of an ops tool. But they're where the danger is. Calling `diskPercent` twice by accident is harmless; calling "drain this node" twice might not be.

**A useful split:** do the deciding in pure functions and the doing in a few small functions with side effects. "Which pods are unhealthy?" is pure and easy to test with made-up numbers. "Restart these pods" is the thin, careful outer layer. You'll see this pay off in the *Testing* module, and the danger of repeated side effects returns in *Idempotency keys: replay retried requests*. Functions that change variables outside themselves appear in *Functions as values: state that sticks*.

## Decomposing a task, top-down

*Decomposition* means breaking a big job into smaller named steps until every step is easy. The simplest way is *top-down*: start from the whole task and work downward.

1. **Say the task in one sentence.** "Deploy a version of a service to some replicas."
2. **List 3–5 steps**, each a verb. Validate the input, build the image name, push it, roll it out.
3. **Write the top function first**, calling the steps as if they already existed.
4. **Fill in each step.** If one is still big, split it again.

```text
deploy(service, version, replicas)       → error
├── validate(service, version)           → error
├── build(service, version)              → image name
├── push(image)                          → error
└── rollout(image, replicas)
```

Here it is as a program. `push` and `rollout` only print, standing in for the real registry and cluster calls you'll write in later modules.

```go
package main

import (
	"errors"
	"fmt"
	"strings"
)

func validate(service string, version string) error {
	if service == "" {
		return errors.New("service name is empty")
	}
	if !strings.HasPrefix(version, "v") {
		return fmt.Errorf("version %q must start with v", version)
	}
	return nil
}

func build(service string, version string) string {
	return "registry.internal/" + service + ":" + version
}

func push(image string) error {
	fmt.Println("pushing", image)
	return nil
}

func rollout(image string, replicas int) {
	for i := 1; i <= replicas; i++ {
		fmt.Printf("replica %d/%d runs %s\n", i, replicas, image)
	}
}

func deploy(service string, version string, replicas int) error {
	err := validate(service, version)
	if err != nil {
		return err
	}
	image := build(service, version)
	err = push(image)
	if err != nil {
		return err
	}
	rollout(image, replicas)
	return nil
}

func main() {
	err := deploy("api", "v1.4.2", 2)
	fmt.Println("first deploy error:", err)
	err = deploy("api", "1.4.3", 2)
	fmt.Println("second deploy error:", err)
}
```

```text
pushing registry.internal/api:v1.4.2
replica 1/2 runs registry.internal/api:v1.4.2
replica 2/2 runs registry.internal/api:v1.4.2
first deploy error: <nil>
second deploy error: version "1.4.3" must start with v
```

What happened:

1. The first `deploy` runs all four steps. Nothing failed, so `err` is `nil`, Go's word for "no error".
2. The second `deploy` stops at `validate`: `1.4.3` has no `v`. `deploy` returns that error straight away, so `build`, `push` and `rollout` never run. Nothing half-deployed.
3. The error names the problem in words. You know which step failed without reading any other code.

Read `deploy` on its own: it's a table of contents. You can understand the whole process in five seconds, and dig into one step only when you need to. Returning `(value, error)` and checking it every time is the habit from *Multiple return values: don't drop the error*.

### Naming functions

A good name lets a reader skip the body.

| Kind of function | Name style | Examples |
|---|---|---|
| Does something | verb | `push`, `rotateLogs`, `drainNode` |
| Computes a value | what it returns | `diskPercent`, `imageRef` |
| Answers yes/no | `is`/`has`/`should` | `isHealthy`, `hasQuota` |

Two quick tests. If the honest name needs "and" (`validateAndPush`), it's two functions. If you can't describe the job in one sentence, it isn't one job yet.

## Functions are your first abstraction

Once `diskPercent` works, the code that calls it stops caring *how* it works. Callers rely only on its *signature*: the name, the parameters and the result type. You could change the body completely, and no caller would notice.

That's *abstraction*: hiding the details behind a simple promise, so you can think about one level at a time. `deploy` is an abstraction over four steps; `kubectl rollout` is an abstraction over hundreds. Functions are the smallest kind. The *Abstraction* guide builds from here to interfaces and packages.

## Common mistakes

- **Defining but never calling.** A definition alone runs nothing. If "nothing happens", check that something calls your function.
- **Using the name instead of calling it.** `Version` is the function itself; `Version()` runs it and gives its result. The compiler says `mismatched types string and func() string`, as in *Reading compiler errors*.
- **Quoting a parameter.** `"host"` is the text h-o-s-t. `host` is the variable.
- **Swapping arguments of the same type.** The compiler can't catch it. Check the order against the definition.
- **Expecting a parameter change to reach the caller.** Parameters are copies. Return the new value instead.
- **Ignoring a returned error.** If a function returns an `error`, check it before using the other results.
- **One giant `main`.** If you scroll to understand a function, split it.

## Check yourself

1. In `func restart(pod string, graceSeconds int) error`, what is the function's name, what are its parameters, and what's its result type? In the call `restart("web-7", 30)`, what are the arguments?
2. With `root := 30` and `size := 120`, what does `diskPercent(size, root)` return, and why?
3. In the scope example, why can't `main` use `total`? How should it get the value?
4. In the stack trace above, which line crashed, and which line holds the real mistake?
5. Pure or not? `func nextBackoff(seconds int) int { return seconds * 2 }` and `func logLine(msg string) { fmt.Println(msg) }`.
6. Decompose "rotate the logs on a server" into three or four functions. Give each a name and say what it returns.

## Answers

1. Name: `restart`. Parameters: `pod` (a `string`) and `graceSeconds` (an `int`). Result type: `error`. Arguments in the call: `"web-7"` and `30`. `pod` becomes `"web-7"`, `graceSeconds` becomes `30`, matched by position.
2. `400`. Arguments match by position, so `usedGB` gets 120 and `totalGB` gets 30: `120 * 100 / 30` is 400. Both are `int`, so Go can't tell they're swapped.
3. `total` was created inside `sumDisks`, so its scope ends at `sumDisks`'s closing `}`. `main` should use the value `sumDisks` returns: `t := sumDisks()`.
4. Line 6 crashed (`usedGB * 100 / totalGB` divided by zero). The real mistake is line 16, `report("web-2", 10, 0)`, which passed a total of 0. You find it by following the frames down.
5. `nextBackoff` is pure: the same input always gives the same output, and it touches nothing else. `logLine` has a side effect: it prints.
6. One good answer: `findOldLogs(dir string) (int, error)` counts logs older than a cutoff; `compress(path string) error` gzips one; `deleteOld(dir string, keep int) error` removes the oldest beyond `keep`; and `rotate(dir string) error` calls the others in order and stops at the first error. Any split works if each function has one job and a name that says it.

## Key terms

| Term | Meaning |
|---|---|
| function | A named, reusable step with inputs and an output |
| definition | The `func ...{ }` code that describes a function; runs nothing |
| call | Running a function with real values: `name(arguments)` |
| parameter | A named input in the definition, like `usedGB int` |
| argument | The actual value passed in a call, like `45` |
| return value | The result a call hands back; the call is replaced by it |
| signature | A function's name, parameter types and result types |
| scope | The part of the code where a name exists and can be used |
| call stack | The pile of calls in progress, each waiting for the one above |
| stack frame | One call in progress, with its own parameters and variables |
| stack trace | A printout of the call stack, usually when a program crashes |
| pure function | A function whose result depends only on its inputs, with no side effects |
| side effect | Anything a function does besides returning a value: printing, files, network |
| decomposition | Splitting a task into smaller named steps |
| abstraction | Hiding details behind a simple promise, like a function's signature |

## Go deeper

- [Go by Example: Functions](https://gobyexample.com/functions): short, runnable examples of defining and calling functions in Go.
- [Go by Example: Multiple Return Values](https://gobyexample.com/multiple-return-values): the `(value, error)` shape used by `deploy` above.
- [Call stack (Wikipedia)](https://en.wikipedia.org/wiki/Call_stack): frames, return addresses and why a stack is the right shape.
- [Side effect (Wikipedia)](https://en.wikipedia.org/wiki/Side_effect_(computer_science)): a precise definition, and how it relates to pure functions.
- [Top-down and bottom-up design (Wikipedia)](https://en.wikipedia.org/wiki/Top-down_and_bottom-up_design): decomposition as a design approach, beyond programming.
