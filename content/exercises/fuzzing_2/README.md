# Fuzzing: properties of a truncated alert

"Never panics" is the simplest thing a fuzz target can check. Most bugs
don't crash, they return the wrong answer. This exercise checks *properties*
of the answer, on inputs you never wrote down.

## Learn

### Recap

As in *Fuzzing*, a fuzz target calls `f.Add` with seeds and `f.Fuzz` with
the function to run. A plain `go test` replays the seeds and the files in
`testdata/fuzz/FuzzTruncate/`, and `go test -run='^$' -fuzz=FuzzTruncate
-fuzztime=10s` lets Go invent new inputs and saves any failure there.
Gopherlings grades with plain `go test`, using mutants as before.

### Properties

You can't write a want for an input you haven't seen, but you can often say
what must be true of every answer. For `Truncate(msg, limit)`:

| Property | In Go |
|---|---|
| The result fits the limit | `len(got) <= limit` (when `limit >= 0`) |
| The result is the start of the message | `strings.HasPrefix(msg, got)` |
| No character is cut in half | `utf8.ValidString(got)` |

Check each one with `t.Errorf`, so one failing input reports every broken
property. Another common property, for pairs like encode and decode, is the
*round trip*: `Decode(Encode(x))` must give back `x`.

`Truncate` has a fourth hidden "property": it must not panic, even for a
negative `limit`. Any panic fails the fuzz target, so you get that check for
free as long as a seed or the fuzzer reaches it.

### Bytes, characters and invalid UTF-8

`len` counts bytes, and in UTF-8 a character can take more than one byte:
`"é"` is 2 bytes and `"🔥"` is 4. Cut `"héllo"` after 2 bytes and you get
`"h"` plus half of `"é"`, which is not valid UTF-8. A pager shows it as `�`,
or rejects the message.

`utf8.ValidString(s)` from `unicode/utf8` reports whether `s` is valid
UTF-8. The fuzzer doesn't care about UTF-8, though: it changes raw bytes, and
a Go string can hold any bytes. Many generated `msg` values are already
invalid, and then the result can't be expected to be valid either. So only
check that property when the input is valid:

```go
if utf8.ValidString(msg) && !utf8.ValidString(got) {
	t.Errorf(...)
}
```

You can also skip an input entirely with `t.Skip()`, but that would skip
the length and prefix checks too, which still hold for invalid input.

## Your task

The alerting service posts messages to a pager API that accepts at most a
certain number of bytes. `Truncate` shortens messages to fit. A fuzzer found
that it breaks characters, and saved the input as
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

In `FuzzTruncate` in `exercise_test.go`, add the seeds
`("disk 92% full on db-1", 10)` and `("disk full", -1)`. Inside `f.Fuzz`,
call `Truncate(msg, limit)` and report with `t.Errorf` when:

- `limit >= 0` and the result is longer than `limit` bytes,
- the result is not a prefix of `msg`.

The negative seed catches a `Truncate` that panics on negative limits. You
need to add `"strings"` to the imports.

### 2. Check that characters stay whole

Also report with `t.Errorf` when `msg` is valid UTF-8 but the result is not.
Add `"unicode/utf8"` to the imports. Your fuzz target now fails on
`FuzzTruncate/cut_in_character`: it found the bug.

### 3. Fix Truncate

Fix `realTruncate` in `exercise.go`. When `msg` is too long, cut at `limit`
bytes, but if that would land inside a character, move the cut back to where
the character starts. `utf8.RuneStart(b)` reports whether byte `b` is the
first byte of a character, so `msg[cut]` being a start means `msg[:cut]`
ends on a whole character.

`TestTruncate` in `grader_test.go` checks the examples above plus
`("disk full", 20)` → `"disk full"`, `("disk full", 9)` → `"disk full"`,
`("disk full", 0)` → `""`, `("héllo", 3)` → `"hé"`,
`("🔥 db-1 down", 3)` → `""` and `("🔥 db-1 down", 5)` → `"🔥 "`.

The exercise is done when `FuzzTruncate` and `TestTruncate` pass and
`TestGrader` reports that every mutant was caught.

## Learn more

- [Package unicode/utf8](https://pkg.go.dev/unicode/utf8)
- [Strings, bytes, runes and characters in Go](https://go.dev/blog/strings)
- [Go Fuzzing](https://go.dev/doc/security/fuzz/)
