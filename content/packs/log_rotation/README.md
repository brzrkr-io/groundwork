# Practice: refactor the log rotator

The log rotator runs every night on every node: it rotates big logs,
compresses old ones and deletes the oldest. It works, and its tests pass,
but nobody wants to touch it: one 27-line function, five levels deep,
with names like `f` and numbers like `604800`. Your job is to make it
code the next on-call engineer can change safely, without changing what
it does.

**New here:** nothing; practice. This combines *Naming things*, *Named
constants*, *Extract function* and *Guard clauses*.

## Learn

### Recap

- **Refactor in small steps.** Change one thing, run the tests, repeat.
  If they go red, undo the last step instead of debugging a big change.
- **Names** say what a value holds; a unit or a role beats a letter.
- **Named constants** replace numbers that come from a decision.
- **Extract function** moves a block into a function whose name replaces
  a comment.
- **Guard clauses** handle a case and leave, so there is no `else`.

Every test first runs `Plan` on example files like the ones below (the
full list is `checkPlan` in `exercise_test.go`). If your refactor changes
the output, the test stops there and tells you to undo the last step.

### 1 MiB

Sizes in ops are usually powers of two: 1 KiB is 1024 bytes and 1 MiB is
1024 × 1024 = 1,048,576 bytes. `100 << 20` is a short way to write 100 MiB:
`<<` shifts the bits left, and shifting by 20 multiplies by 2²⁰, which is
1 MiB. `100 * 1024 * 1024` is the same number; pick the one you find
clearer.

## Your task

The rules `Plan` follows today, which must not change:

| File | Rule | Action |
|---|---|---|
| the live log (no `.log.` in the name, like `app.log`) | 100 MiB (104,857,600 bytes) or more | `rotate` |
| the live log | smaller | `keep` |
| a rotated log (`.log.` in the name, like `app.log.3`) | older than 7 days | `delete` |
| a rotated log | not ending `.gz`, older than 1 day | `compress` |
| a rotated log | anything else | `keep` |

Each line of the result is the action, a space, and the file name, in the
order of the files. With `now` at 03:00:

```text
app.log       150 MiB   now               => "rotate app.log"
cron.log      100 MiB-1 now               => "keep cron.log"
app.log.2     80 MiB    24 hours ago      => "keep app.log.2"
app.log.3     80 MiB    25 hours ago      => "compress app.log.3"
app.log.5.gz  9 MiB     7 days ago        => "keep app.log.5.gz"
app.log.6.gz  9 MiB     7 days 1 min ago  => "delete app.log.6.gz"
```

The tasks can be done in any order, but this order is the easiest.

### 1. Rename the cryptic names

**Goal:** give every variable and parameter in `exercise.go` a name of
three letters or more (`ok` and `_` excepted).

`fs`, `t`, `r` and `f` all need real names. Keep the name `Plan`.

**Think it through:**
- What does `t` hold? Is "t" for "time" enough when there's also a `Modified` time?
- What is `r` once the function returns it?

### 2. Name the numbers

**Goal:** declare constants `rotateSize`, `compressAfter` and `deleteAfter`,
and use them so no function has a number in it.

`compressAfter` and `deleteAfter` are `time.Duration`s built from `time`
units; `rotateSize` is 100 MiB in bytes.

**Think it through:**
- How many seconds is `86400`? And `604800`? What's the clearer way to write them with `time.Hour`?
- Where do constants go: inside `Plan` or outside every function?

### 3. Extract the decisions

**Goal:** move the decisions into `liveAction(file File) string` and
`rotatedAction(file File, now time.Time) string`, each returning the action
word (`"rotate"`, `"keep"`, ...), and get `Plan` to 12 lines of code or
fewer (blank lines and comments don't count).

`Plan` keeps the loop: it decides which helper a file needs and builds
each line from the action and the name.

**Think it through:**
- Which `if` decides between the two kinds of file? That one stays in `Plan`.
- Once the helpers return just the word, how many places add `" "` and the file name?

### 4. Flatten into guard clauses

**Goal:** remove every `else`, and nest no deeper than 2 (a loop with one
`if` inside) in any function.

In the helpers, return as soon as you know the answer.

**Think it through:**
- In `rotatedAction`, which check settles the answer first, no matter what else is true?
- "Not compressed and older than a day: compress". When is the answer `"keep"` instead? See *Guard clauses: fix the flipped conditions*.

## Learn more

- [Refactoring catalog (Martin Fowler)](https://refactoring.com/catalog/)
- [logrotate(8) manual page](https://man7.org/linux/man-pages/man8/logrotate.8.html)
- [Go spec: Arithmetic operators (shifts)](https://go.dev/ref/spec#Arithmetic_operators)
