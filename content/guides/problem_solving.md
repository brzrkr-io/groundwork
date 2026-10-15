# Solving problems you haven't seen before

**In one sentence:** when a task looks new, you don't need a flash of genius;
you need a method you repeat every time, which breaks the task into small
steps you already know how to do.

**Why it matters:** ops work is mostly problems nobody handed you before: a
new log format, a command your script must build, a pipeline that fails in a
new way. The people who stay calm on call aren't smarter than you. They have
a method, and they use it even when they're tired.

**You'll learn:**

- A seven-step method that works on any exercise or ticket.
- How to work examples by hand and plan in pseudocode.
- How to work backwards from the expected output (the `// =>` lines).
- What to do when you're stuck.
- How to use hints and solutions so they actually teach you.

## The method on one page

```text
 1. Understand       say the task in your own words
 2. Examples         work two or three cases by hand
 3. Inputs, output   what goes in, what comes out, and their types
 4. Steps            break it into small steps, in plain words
 5. Smallest piece   write just enough code for the first step
 6. Test             run it and compare with your hand-made examples
 7. Improve          add the next piece, tidy up; repeat 5 to 7
```

It's a loop, not a ladder. Steps 5 to 7 repeat until the task is done. If a
test surprises you, go back to step 1: the surprise usually means you
misunderstood something.

The idea is old. The mathematician George Pólya described four phases in
his 1945 book *How to Solve It*: understand the problem, make a plan, carry
it out, look back. The seven steps above are those four, cut smaller for
code.

## A worked example: a kubectl scale command

Here's a ticket you could get in your first ops job:

> Our deploy bot should show the `kubectl` command for scaling a deployment,
> so a human can check it before running it. Given a deployment name, a
> replica count and a namespace, return the command.

You haven't seen this exact problem. Let's walk the method.

### Step 1: Understand the problem

Say it back in your own words: "I get three things and must give back one
line of text, the kubectl command."

Then find what you don't know. Here: what does a scale command look like?
The kubectl docs show it:

```bash
kubectl scale deployment/api --replicas=3 -n prod
```

Good questions at this step:

- What exactly comes out? A string that is returned, not printed.
- What goes in? A name, a number, a namespace.
- Is anything unclear? If so, ask now, not after an hour of code.

### Step 2: Work examples by hand

Write two or three cases out in full, on paper or in a notes file:

| deployment | replicas | namespace | expected result |
|---|---|---|---|
| `api` | `3` | `prod` | `kubectl scale deployment/api --replicas=3 -n prod` |
| `worker` | `0` | `staging` | `kubectl scale deployment/worker --replicas=0 -n staging` |

Doing this by hand does two things. It proves you understand the task: if
you can't write the answer yourself, no code will help yet. And it gives you
the test cases for step 6.

Pick at least one unusual case, an *edge case*. Zero replicas is valid: it
scales a deployment down to nothing.

### Step 3: Name the inputs and the output

Now turn that into a function *signature*, the first line of the function:

```text
func ScaleCommand(deployment string, replicas int, namespace string) string
     └─────┬────┘ └───────────────────────┬───────────────────────┘  └──┬─┘
           │                              │                             │
       the name        parameters: each is a name, then a type       result
```

- `ScaleCommand` is the function's name. You chose it.
- Inside `( )` are the three parameters. Each one is a name you choose,
  then its type.
- `replicas` is an `int` because it's a count. `deployment` and `namespace`
  are text, so they're `string`.
- The final `string` is the type of the result.

### Step 4: Break it into steps, working backwards

Take one expected result and label every piece: is it fixed text, or does
it come from a parameter?

| Piece of the result | Where it comes from |
|---|---|
| `kubectl scale deployment/` | Fixed text, same every time |
| `api` | The `deployment` parameter |
| ` --replicas=` (starts with a space) | Fixed text |
| `3` | The `replicas` parameter, an `int`: needs turning into text |
| ` -n ` (spaces both sides) | Fixed text |
| `prod` | The `namespace` parameter |

Fixed text goes in quotes. Parameters don't. Now write the plan in plain
words. This is *pseudocode*:

```text
start with the fixed text "kubectl scale deployment/"
add the deployment name
add " --replicas=" and the replica count, turned into text
add " -n " and the namespace
return the whole thing
```

### Step 5: Write the smallest piece

Don't write it all at once. Write the first two lines of the plan, and
check that the result so far is right:

```go
// ScaleCommand returns the kubectl command that scales a deployment.
func ScaleCommand(deployment string, replicas int, namespace string) string {
	return "kubectl scale deployment/" + deployment
}
```

This compiles even though `replicas` and `namespace` aren't used yet. Go
lets parameters go unused; only unused variables you create inside the
function are errors.

### Step 6: Test it

In an exercise, Groundwork runs the tests for you every time you save. Here
you can test it yourself with a small program that prints what you got next
to what you wanted:

```go
func main() {
	got := ScaleCommand("api", 3, "prod")
	want := "kubectl scale deployment/api --replicas=3 -n prod"
	fmt.Printf("got:  %q\nwant: %q\n", got, want)
}
```

```text
got:  "kubectl scale deployment/api"
want: "kubectl scale deployment/api --replicas=3 -n prod"
```

They don't match yet, and that's fine: you've only written the first piece.
What matters is that `got` is exactly the start of `want`, so the first
steps of the plan work. The `%q` verb puts quotes around each value, so a
stray or missing space at either end would show.

Now add the next piece. Your first try might be `+ " --replicas=" +
replicas`, and the compiler stops you with a long message that ends in
`(mismatched types string and int)`.

Step 4 already warned you: `replicas` is an `int` and needs turning into
text. `strconv.Itoa` does that. Here's the finished program:

```go
package main

import (
	"fmt"
	"strconv"
)

// ScaleCommand returns the kubectl command that scales a deployment.
func ScaleCommand(deployment string, replicas int, namespace string) string {
	return "kubectl scale deployment/" + deployment +
		" --replicas=" + strconv.Itoa(replicas) +
		" -n " + namespace
}

func main() {
	got := ScaleCommand("api", 3, "prod")
	want := "kubectl scale deployment/api --replicas=3 -n prod"
	fmt.Printf("got:  %q\nwant: %q\n", got, want)

	got = ScaleCommand("worker", 0, "staging")
	want = "kubectl scale deployment/worker --replicas=0 -n staging"
	fmt.Printf("got:  %q\nwant: %q\n", got, want)
}
```

It prints:

```text
got:  "kubectl scale deployment/api --replicas=3 -n prod"
want: "kubectl scale deployment/api --replicas=3 -n prod"
got:  "kubectl scale deployment/worker --replicas=0 -n staging"
want: "kubectl scale deployment/worker --replicas=0 -n staging"
```

Both examples from step 2 match, including the edge case.

### Step 7: Improve, and look back

It works. Now read it as if someone else wrote it. Could it be clearer?
`fmt.Sprintf` lets you write the whole command as one template:

```go
return fmt.Sprintf("kubectl scale deployment/%s --replicas=%d -n %s",
	deployment, replicas, namespace)
```

`%d` formats the `int`, so there's no `strconv` to remember. Both versions
are correct; pick the one you find easier to read, and run the tests again
after the change.

Last, look back: what will you reuse? Here, **"a command is fixed text plus
values; values that aren't strings need converting."** That pattern comes
back in *Variables: build ops commands* and in nearly every CLI tool you'll
write.

## Pseudocode: plan in plain words

*Pseudocode* is a plan written in plain language, shaped like code but with
no syntax to get wrong. It lets you think about **what** to do before you
think about **how** Go spells it.

A few rules keep it useful:

- One small step per line.
- Use the words from the problem: "deployment", "used", "total".
- If a line feels big, split it into smaller lines.
- Stop when every line looks like something you could write in one or two
  lines of Go.

Here's a plan for a disk summary line, `/var: 45% used`:

```text
work out the percentage: used times 100, divided by total
turn the percentage into text
join: the mount, ": ", the percentage, "% used"
return it
```

Notice the first line multiplies before dividing. A hand example (90 used,
200 total) shows why. See the *Values and types* guide for what goes wrong
otherwise.

## Working backwards from the expected output

Every task in an exercise README shows what the code should return:

```go
Banner() // => "groundwork v1.0.0"
```

Read the `// =>` line as the *specification*: the exact result the tests
will demand. Work backwards from it, the way you did in step 4:

1. **Copy the expected result** somewhere you can mark it up.
2. **Underline what changes** with the inputs. Each underlined piece should
   match a parameter, a constant or a call.
3. **Everything left is fixed text.** It goes in quotes, character for
   character, spaces included.
4. **Check the types.** A result in quotes is a `string`. A bare `45` is an
   `int`. `45.5` is a `float64`.

When there are two `// =>` examples, compare them. Whatever differs between
them comes from the inputs; whatever is the same is probably fixed text.

The same trick works on the job. Before writing a log parser, paste a real
log line and the output you want under it, and mark which parts map where.

## When you're stuck

You're stuck when you're changing code at random, you've seen the same error
three times, or 15 minutes have passed with no progress. That's normal, not
a sign you're bad at this. Try these, roughly in order:

- **Read the error again, slowly.** Fix only the first one. *Reading
  compiler errors* teaches a method for these.
- **Print what you have.** `fmt.Printf("%q\n", x)` shows what a value really
  is, not what you think it is.
- **Do it on paper.** If you can't get the answer by hand, the problem is
  understanding the task, not Go. Go back to step 1.
- **Shrink the problem.** Drop parts until something works, then add them
  back one at a time. Return only the fixed text first, then add one
  parameter.
- **Explain it to a rubber duck.** Say out loud, line by line, what each line
  does. Programmers really do this, with a toy duck or a patient pet.
  Explaining it often makes the gap obvious.
- **Take a break.** Walk away for ten minutes. Answers often turn up once you
  stop staring. A timer helps you actually come back.
- **Write the question down.** Say what you tried and what happened. Half
  the time, writing it answers it. The other half, you have a good question
  to ask someone.

The *Debugging with a method* guide turns this into a full process, for when
code compiles but does the wrong thing.

## Hints and solutions without cheating yourself

Groundwork gives you two kinds of help:

| Key | What you get |
|---|---|
| `h` | A hint for the task. Press again for a stronger one: a nudge, then a closer look, then the answer. |
| `s` | The reference solution, once you pass, or early once you've seen a task's Answer hint (you confirm first). |

Hints are there to be used. The only way to misuse them is to get the green
tick without being able to do it again. A tick you can't repeat teaches
nothing. Some habits that help:

- **Try first.** Give yourself a set time, say ten minutes, before the first
  hint.
- **One level at a time.** After each hint, go back and try again before you
  press `h` again.
- **Type it, don't paste it.** If you use the answer hint, type it in
  yourself, then explain each piece to yourself out loud.
- **Always compare with the solution.** After you pass, press `s`. Yours
  might be different and still fine. Find one thing you'd do their way.
- **Redo it tomorrow.** Delete your function body and write it again from
  memory. Pulling something back out of your memory is called *retrieval
  practice*, and it's one of the best-studied ways to make learning stick.
- **Quiz yourself on the guides.** Every guide has *Check yourself*
  questions and *Key terms*. `groundwork cards` turns them into flashcards
  you can import into Anki on your phone. Answer them a few days apart, not
  all at once.
- **Welcome the reviews.** A few days later, and again weeks later,
  Groundwork brings back practice exercises and any exercise whose solution
  you opened early. Spacing practice out like this makes it stick.

Reading a solution and thinking "yes, that makes sense" feels like learning,
but it's recognition. Being able to write it from a blank page is the skill you
need on call.

## Common mistakes

- **Typing before understanding.** Skipping steps 1 and 2 feels faster and
  usually costs more time.
- **Writing everything at once.** Thirty new lines and six errors are much
  harder than three new lines and one error.
- **Changing code at random.** If you can't say why a change should work,
  stop and go back to your examples.
- **Not reading the test failure.** "You returned X, the test wanted Y" tells
  you exactly what to compare.
- **Forgetting fixed spaces.** Working backwards, the spaces between pieces
  belong to the fixed text, and they're easy to drop.
- **Staying stuck for an hour.** Take the break, use the hint. Struggle helps
  only while you're still making progress.

## Check yourself

1. Why work out examples by hand before you write any code?
2. A README says `Endpoint("api", 8080) // => "http://api:8080/healthz"`.
   Which pieces are fixed text, and which come from the parameters?
3. You wrote a whole 25-line function at once and got six compile errors.
   What would the method have had you do instead?
4. Write pseudocode for `DiskLine(mount string, used int, total int) string`
   that returns `/var: 45% used` when called as `DiskLine("/var", 90, 200)`.
5. You pressed `h` three times, used the answer and passed. What should you
   do next so you actually learn it?

## Answers

1. Because if you can't produce the answer by hand, you don't understand the
   task yet, and code won't fix that. The hand examples also become your
   tests: you know exactly what the code must return.
2. Fixed text: `http://`, `:` and `/healthz`. From the parameters: `api` and
   `8080`. The `8080` is an `int`, so it needs turning into text with
   `strconv.Itoa`, or formatting with `%d`.
3. Write the smallest piece, such as the signature and a first `return`, and
   run it. Then add one step at a time and run again after each. Each error
   would then come from the few lines you just added.
4. For example: "work out the percentage: used times 100, divided by total.
   Turn it into text. Join the mount, `": "`, the percentage and `"% used"`.
   Return it." Multiplying first matters: `90 / 200` is 0 with ints.
5. Type the answer in yourself and explain each piece out loud. After you
   pass, press `s` and compare. Then redo it from a blank function later,
   ideally the next day, without hints.

## Key terms

| Term | Meaning |
|---|---|
| Edge case | An unusual but valid input, like zero replicas or an empty name. |
| Signature | A function's first line: its name, parameters and result type. |
| Pseudocode | A plan in plain words, one small step per line, with no syntax. |
| Specification | The exact behaviour required; in exercises, the `// =>` lines. |
| Working backwards | Starting from the expected result and labelling where each piece comes from. |
| Incremental development | Writing a little code, running it, then adding the next piece. |
| Rubber duck debugging | Explaining code line by line, out loud, to find the mistake. |
| Retrieval practice | Learning by recalling something from memory, not re-reading it. |

## Watch

- [Rubber Duck Debugging - How to Solve a Problem](https://www.youtube.com/watch?v=NBgIHOrjSxs) · Coder Foundry · 6 min: a short, practical look at explaining your code out loud, line by line, until the gap shows itself.
- [CS50x 2024 - Lecture 0 - Scratch](https://www.youtube.com/watch?v=3LPJfIKxwWc) · CS50 · 125 min: watch only "Algorithms" and "Pseudocode" (51:30 to 1:05:10), where David Malan solves a phone-book search step by step and writes the plan in plain words first.
- [Hammock Driven Development - Rich Hickey](https://www.youtube.com/watch?v=f84n5oFoZBc) · ClojureTV · 40 min: a classic talk on understanding a problem before coding, and why stepping away from the keyboard helps you solve it.

## Go deeper

- [Wikipedia: How to Solve It](https://en.wikipedia.org/wiki/How_to_Solve_It)
  — Pólya's four phases and his questions for getting unstuck, in a short
  summary.
- [How to Design Programs](https://htdp.org/) — a free textbook built around
  a step-by-step "design recipe" for writing functions, close to the method
  here.
- [Wikipedia: Rubber duck debugging](https://en.wikipedia.org/wiki/Rubber_duck_debugging)
  — where the duck comes from, and why explaining out loud works.
