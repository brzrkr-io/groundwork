# Fuzzing: properties of a truncated alert

"Never panics" is the simplest thing a fuzz target can check. But most bugs
don't crash; they return the wrong answer. This exercise checks *properties*
of the answer, on inputs you never wrote down.

**New here:** checking properties in a fuzz target, two-parameter fuzz inputs, bytes vs. characters, `unicode/utf8`.

## Learn

### Recap

As in *Fuzzing*: a fuzz target calls `f.Add` with seeds and `f.Fuzz` with the
function to run. A plain `go test` replays the seeds and the files in
`testdata/fuzz/FuzzTruncate/`. `go test -run='^$' -fuzz=FuzzTruncate
-fuzztime=10s` lets Go invent new inputs and saves any failure there.
Groundwork grades with plain `go test`, using *mutants* as before.

### Two inputs

A fuzz function can take more than one input after `t`. Each `f.Add` call
then passes one value per input, in the same order and types:

```go
f.Add("web-1", 3)
f.Fuzz(func(t *testing.T, host string, count int) {
	// ...
})
```

> **Why is it written like this?** Go requires `f.Add`'s values to match the
> fuzz function's parameters after `t`, one for one; a mismatch fails the
> fuzz target when it runs. The parameter names are your choice.

### Properties

You can't write a "want" for an input you haven't seen, but you can say what
must be true of *every* answer. For `Truncate(msg, limit)`:

| Property | In words |
|---|---|
| Fits the limit | when `limit >= 0`, the result is at most `limit` bytes |
| Is the start of the message | `msg` begins with the result |
| No character cut in half | the result is valid UTF-8 |

Check each with its own `if` and `t.Errorf`, so one bad input reports every
broken property. `strings.HasPrefix(s, start)` reports whether `s` begins
with `start`.

`Truncate` must also never panic, even for a negative `limit`. Any panic
fails the fuzz target, so you get that check for free, as long as a seed or
the fuzzer reaches it.

### Bytes, characters and invalid UTF-8

`len` counts *bytes*, and in UTF-8 one character can take several: `"é"` is 2
bytes and `"🔥"` is 4. Cut `"héllo"` after 2 bytes and you get `"h"` plus half
of `"é"`: not valid UTF-8. A pager shows it as `�`, or rejects the message.

`utf8.ValidString(s)` from `unicode/utf8` reports whether `s` is valid UTF-8.
The fuzzer changes raw bytes, so many generated messages are already invalid,
and then you can't expect a valid result. Check this property only when the
input was valid. Combine two conditions with `&&`:

```go
if isReachable(host) && !isHealthy(host) {
	t.Errorf("%s is up but unhealthy", host)
}
```

You could skip invalid input with `t.Skip()`, but that would also skip the
length and prefix checks, which still hold for invalid input.

### Finding where a character starts

`utf8.RuneStart(b)` reports whether byte `b` is the *first* byte of a
character. In `"héllo"`, byte 2 is the second half of `é`, so
`utf8.RuneStart("héllo"[2])` is false, while byte 3 (`l`) gives true.

## Your task

The alerting service posts messages to a pager API that accepts a limited
number of bytes. `Truncate` shortens messages to fit. A fuzzer found that it
breaks characters, and saved the input as
`testdata/fuzz/FuzzTruncate/cut_in_character`:

```text
go test fuzz v1
string("héllo")
int(2)
```

```go
Truncate("disk 92% full on db-1", 10) // => "disk 92% f"
Truncate("héllo", 2)                  // => "h" (é doesn't fit, so drop it)
Truncate("disk full", -1)             // => ""
```

### 1. Check length and prefix

**Goal:** in `FuzzTruncate` in `exercise_test.go`, add two seeds and report
with `t.Errorf` when the result is too long or isn't the start of `msg`.

- Seeds: `("disk 92% full on db-1", 10)` and `("disk full", -1)`.
- Inside `f.Fuzz`, call `Truncate(msg, limit)` once.
- Report when `limit >= 0` and the result is longer than `limit` bytes.
- Report when the result is not a prefix of `msg`.

The negative seed catches a `Truncate` that panics on negative limits. Add
`"strings"` to the imports.

**Think it through:**
- Which property only makes sense for a non-negative limit?
- For the prefix check, which string goes first in `strings.HasPrefix`?

### 2. Check that characters stay whole

**Goal:** also report with `t.Errorf` when `msg` is valid UTF-8 but the
result is not.

Add `"unicode/utf8"` to the imports. Your fuzz target now fails on
`FuzzTruncate/cut_in_character`: it found the bug.

**Think it through:**
- Which two strings do you check, and which one must be valid for the check
  to apply?

### 3. Fix Truncate

**Goal:** fix `realTruncate` in `exercise.go` so it never cuts a character
in half.

When `msg` is too long, cut at `limit` bytes. If that lands inside a
character, move the cut back to where that character starts. `msg[:cut]`
ends on a whole character when `msg[cut]` is the first byte of one. Add
`"unicode/utf8"` to the imports.

`TestTruncate` in `grader_test.go` checks the examples above plus:

| Call | Result |
|---|---|
| `Truncate("disk full", 20)` | `"disk full"` |
| `Truncate("disk full", 9)` | `"disk full"` |
| `Truncate("disk full", 0)` | `""` |
| `Truncate("héllo", 3)` | `"hé"` |
| `Truncate("🔥 db-1 down", 3)` | `""` |
| `Truncate("🔥 db-1 down", 5)` | `"🔥 "` |

**Think it through:**
- Which line in `realTruncate` does the cut right now?
- How many steps back might you need? (A character is up to 4 bytes.) What
  loop keeps stepping while a condition holds?

The exercise is done when `FuzzTruncate` and `TestTruncate` pass and
`TestGrader` reports that every mutant was caught.

## Learn more

- [Package unicode/utf8](https://pkg.go.dev/unicode/utf8)
- [Strings, bytes, runes and characters in Go](https://go.dev/blog/strings)
- [Go Fuzzing](https://go.dev/doc/security/fuzz/)
