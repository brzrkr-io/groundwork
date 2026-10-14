# Naming things

Code is read many more times than it is written: in review, during an
incident at 3 a.m., by you in six months. A good name answers "what is this?"
before the reader has to work it out. This exercise takes working code with
terrible names and makes it readable without changing what it does.

**New here:** choosing names, renaming safely with your editor, explaining
variables, tests that read your code's structure.

## Learn

### What a good name does

Compare two versions of the same line:

```go
if m[x] >= limit {
```

```go
if counts[service] >= limit {
```

Both compile and both work. Only the second tells you what the map holds and
what the key is. With the first, you scroll up, find where `m` was filled,
and work it out. Every reader pays that cost, every time.

A good name says **what the value is** (or, for a function, **what it
does**), in words the reader already uses. For this alert bot, those words
are *service*, *alert*, *level*, *page*.

### How long should a name be?

Go likes short names, but "short" depends on how far the name travels:

| Where the name lives | Good names |
|---|---|
| a loop of 2–3 lines | `i`, `n`, `s` are fine |
| a whole function | `counts`, `line`, `service` |
| a whole package | `criticalService`, `pageLine` |

The further a reader is from the line that created a value, the more the
name must explain. A one-letter name is fine when its whole life fits on
one screen and its meaning is obvious. In a 20-line function, it isn't.

> **Why is it written like this?** **Go requires** names to start with a
> letter or `_`, and an upper-case first letter exports a name from the
> package. **Convention** (gofmt doesn't check it, but every Go reviewer
> will) is *MixedCaps*: `pageLine`, not `page_line`. Initialisms stay in
> one case: `serviceURL`, not `serviceUrl`. **Your choice:** the words
> themselves. Pick the ones your team says out loud.

### Rename with your editor, not by hand

Renaming by hand means finding every use, and missing one breaks the build.
Your editor knows Go: put the cursor on a name and use **Rename Symbol**
(`F2` in VS Code, `:lua vim.lsp.buf.rename()` or `grn` in Neovim). It
renames the declaration and every use, and nothing else. A plain
find-and-replace of `s` would also change every `s` inside `"strings"`.

Comments are text, so check them yourself. In Go, a function's comment
starts with the function's name, and a stale comment is a small lie.

### Explaining variables

Indexes and positions are hard to read:

```go
if fields[0] == "GET" && fields[2] == "500" {
```

Give each part a name first, and the condition reads like a sentence:

```go
method, status := fields[0], fields[2]
if method == "GET" && status == "500" {
```

A variable whose only job is to name a value is an *explaining variable*. It
costs a line and saves the reader a puzzle.

> **Why is it written like this?** `a, b := x, y` creates two variables in
> one line; **Go requires** the same number of names and values. Splitting it
> into two lines is **your choice**.

### How the tests check names

Your tests can't see names by running the code, so `structure_test.go`
reads `exercise.go` as Go syntax, with the standard library's `go/ast`
package, and looks at the names you declared. Every test also runs `Pages`
on a sample log and fails if the result changed: renaming must never change
behaviour. You don't need to read `structure_test.go`, but it's there if
you're curious.

## Your task

The alert bot pages the on-call engineer for every service with enough
critical alerts. `Pages` works, but someone wrote it in a hurry. Make it
readable. Keep `Pages`' name and what it returns exactly as they are.

### 1. Rename the helper functions

**Goal:** rename `prs` to `criticalService` and `fmtP` to `pageLine`.

Rename each one where it's declared, where `Pages` calls it, and as the
first word of its comment:

```go
// pageLine formats one page line, such as "api: 3 critical".
func pageLine(...
```

**Think it through:**
- Put the cursor on `prs` and use your editor's Rename Symbol. How many places changed?
- Which part did the editor leave for you to change by hand?

### 2. Rename the variables in Pages

**Goal:** in `Pages`, rename `m`, `o`, `l`, `s`, `r` and `x` to names of three
letters or more that say what each one holds.

The test rejects any variable or parameter in `Pages` shorter than three
letters, except `ok` (Go's usual name for "did it work?") and `_`. The words
are yours. One possible set: `counts`, `services`, `line`, `service`,
`pages`.

**Think it through:**
- For each one, find where it's created. What goes into it?
- `s` and `x` hold the same kind of thing in two different loops. Can they share a name?

### 3. Name the parts of an alert line

**Goal:** in `criticalService`, give the level and the service their own
variables named `level` and `service`, and give every other name in
`criticalService` and `pageLine` at least three letters.

An alert line splits into fields like this:

```text
2026-10-13T09:00:00Z  critical  api  disk almost full
fields[0]             fields[1] fields[2] ...
```

`Pages` must keep ignoring lines with fewer than three fields, such as
`"garbage"`, without crashing.

**Think it through:**
- Which field is the level and which is the service?
- The current code checks the length and the level in one `if`. If you name `fields[1]` before checking the length, what happens on the line `"garbage"`?

## Learn more

- [Effective Go: Names](https://go.dev/doc/effective_go#names)
- [Go Code Review Comments: Variable names](https://go.dev/wiki/CodeReviewComments#variable-names)
- [What's in a name? (Andrew Gerrand)](https://go.dev/talks/2014/names.slide)
