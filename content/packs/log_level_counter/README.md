# Practice: log level counter

When an incident starts, the first questions are "how many errors?" and
"which error keeps repeating?". You'll build a small tool that answers both
from raw log lines.

**New here:** nothing; practice.

## Learn

### Recap table

| You need to… | Use | Taught in |
|---|---|---|
| split a line into words | `strings.Fields(line)` | *The strings package* |
| glue words back together | `strings.Join(words, " ")` | *The strings package* |
| make text upper case | `strings.ToUpper(s)` | *Strings* |
| count things | `counts[key]++` on a `map[string]int` | *Maps* |
| a map's keys in order | `slices.Sorted(maps.Keys(m))` | *Sorting* |
| sort with your own rule | `slices.SortFunc` + `cmp.Compare` | *Sorting* |
| format a number into text | `fmt.Sprintf("%d", n)` | *Formatting with fmt* |
| check if a slice holds a value | `slices.Contains(s, v)` | *Slice gotchas: drain and promote without side effects* |

Add each package (`"fmt"`, `"maps"`, `"slices"`, `"cmp"`) to the imports when
you first use it.

### strings.Fields ignores extra spaces

```go
strings.Fields("  red   green blue ")
// => []string{"red", "green", "blue"}
```

It splits on any run of spaces and never returns empty strings.

### The log format

```text
2025-01-02T10:00:00Z ERROR db timeout
└──── timestamp ───┘ └lvl┘ └ message ┘
```

1. A timestamp (you don't need to check it).
2. A *level*: how serious the line is (`ERROR`, `WARN`, `INFO`, `DEBUG`, …).
3. A message: all the remaining words.

> **Why is it written like this?** `ParseLine` returns three values with an
> `ok bool` last. Go allows any number of return values; putting `ok` last is
> a **convention** that callers recognise (`x, ok := ...`).

## Your task

The on-call team pastes log lines into your tool. Each step has its own test,
and later steps can call your earlier functions.

### 1. Parse one line

**Goal:** make `ParseLine(line)` return the level, the message, and `true`,
or `"", "", false` for a bad line.

- A line needs at least **three** fields (timestamp, level, one message word).
  Fewer → `"", "", false`.
- The level is the second field, in upper case (`warn` → `WARN`).
- The message is every field after the level, joined with single spaces.

```go
ParseLine("2025-01-02T10:00:00Z ERROR db timeout")
// => "ERROR", "db timeout", true
ParseLine("2025-01-02T10:00:05Z   warn   disk  80% full")
// => "WARN", "disk 80% full", true
ParseLine("2025-01-02T10:00:09Z INFO")
// => "", "", false   (no message)
ParseLine("")
// => "", "", false
```

**Think it through:**
- Check the number of fields *before* you index into them.
- Counting from 0, which index is the level? Which sub-slice is the message?

### 2. Count lines per level

**Goal:** make `CountLevels(lines)` return a map from level to how many valid
lines have that level.

Skip lines where `ParseLine` says `ok` is `false`. With no valid lines, return
an empty map made with `make`, never `nil`.

```go
CountLevels([]string{
	"2025-01-02T10:00:00Z ERROR db timeout",
	"2025-01-02T10:00:01Z info request served",
	"2025-01-02T10:00:02Z ERROR db timeout",
	"garbage",
})
// => map[string]int{"ERROR": 2, "INFO": 1}
```

**Think it through:**
- Which of `ParseLine`'s three results do you need here? (Use `_` for the
  one you don't.)
- What happens when you `++` a key that isn't in the map yet?

### 3. Format a summary

**Goal:** make `Summary(counts)` return one line like
`"ERROR=2 WARN=1 INFO=5 DEBUG=0"`.

- First the four standard levels from `standardLevels`, in that order
  (`ERROR`, `WARN`, `INFO`, `DEBUG`), with `0` for a missing level.
- Then any other levels in the map (such as `TRACE`), in alphabetical order.
- Each entry is `LEVEL=count`; entries are separated by one space.

The starter only joins the level names; replace that.

```go
Summary(map[string]int{"INFO": 5, "ERROR": 2, "WARN": 1})
// => "ERROR=2 WARN=1 INFO=5 DEBUG=0"
Summary(map[string]int{"ERROR": 1, "TRACE": 4, "AUDIT": 2})
// => "ERROR=1 WARN=0 INFO=0 DEBUG=0 AUDIT=2 TRACE=4"
Summary(map[string]int{})
// => "ERROR=0 WARN=0 INFO=0 DEBUG=0"
```

**Think it through:**
- In `LEVEL=count`, the `=` is fixed text; the level and count come from your
  loop and the map.
- Two loops: one over `standardLevels`, one over the sorted map keys. How do
  you avoid writing a standard level twice?

### 4. List the top messages for a level

**Goal:** make `TopMessages(lines, level, n)` return up to `n` entries like
`"3x db timeout"` for the most repeated messages at `level`.

- Only count valid lines whose level (from `ParseLine`, so upper case) equals
  `level`. `level` is always upper case, like `"ERROR"`.
- Order by count, highest first. Same count → alphabetical by message.
- Return at most `n` entries. If `n` is 0 or nothing matches, return
  `[]string{}`, not `nil`.
- Each entry is `<count>x <message>`: the number, a lower-case `x`, one space,
  the message.

```go
lines := []string{
	"2025-01-02T10:00:00Z ERROR db timeout",
	"2025-01-02T10:00:01Z ERROR cache miss storm",
	"2025-01-02T10:00:02Z ERROR db timeout",
	"2025-01-02T10:00:03Z WARN db timeout",
	"2025-01-02T10:00:04Z ERROR auth failed",
}
TopMessages(lines, "ERROR", 2)
// => []string{"2x db timeout", "1x auth failed"}
```

`auth failed` and `cache miss storm` both appear once, so `auth failed` wins
the tie alphabetically. The `WARN` line is ignored.

**Think it through:**
- Four steps: count messages in a map, sort them, cut to `n`, format each.
- In your sort rule, which comes first when counts differ, and when they're
  equal?
- How do you make an empty slice that isn't `nil`?

## Learn more

- [strings.Fields](https://pkg.go.dev/strings#Fields)
- [slices.SortFunc](https://pkg.go.dev/slices#SortFunc)
- [maps.Keys](https://pkg.go.dev/maps#Keys)
- [Go by Example: Sorting by Functions](https://gobyexample.com/sorting-by-functions)
