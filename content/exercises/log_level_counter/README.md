# Practice: log level counter

When an incident starts, the first questions are "how many errors?" and
"which error keeps repeating?". In this exercise you build a small tool that
answers both from raw log lines, using everything from this module: slices,
maps, sorting and the `strings` package.

## Learn

This is a practice exercise, so there is nothing brand new. Here is a recap of
the pieces you will reach for.

| You need to… | Use | Taught in |
|---|---|---|
| split a line into words, ignoring extra spaces | `strings.Fields(line)` | *The strings package* |
| glue words back together | `strings.Join(words, " ")` | *The strings package* |
| make text upper case | `strings.ToUpper(s)` | *Strings* |
| count things | `counts[key]++` on a `map[string]int` | *Maps* |
| get a map's keys in order | `slices.Sorted(maps.Keys(m))` | *Sorting* |
| sort with your own rule | `slices.SortFunc` + `cmp.Compare` | *Sorting* |
| build text piece by piece | `strings.Builder` or `strings.Join` | *The strings package* |
| check whether a slice holds a value | `slices.Contains(s, v)` (new: returns `true` if any element `== v`) | – |

A quick reminder of how `strings.Fields` treats a log line:

```go
strings.Fields("2025-01-02T10:00:00Z  ERROR   db timeout")
// => []string{"2025-01-02T10:00:00Z", "ERROR", "db", "timeout"}
```

It splits on any run of spaces and never returns empty strings, so stray
double spaces don't matter.

### The log format

Every line looks like this:

```text
2025-01-02T10:00:00Z ERROR db timeout
└──── timestamp ───┘ └lvl┘ └ message ┘
```

1. A timestamp (you don't need to check it).
2. A *level*: how serious the line is, such as `ERROR`, `WARN`, `INFO` or
   `DEBUG`.
3. A message: all the remaining words.

## Your task

The on-call team pastes log lines into your tool. Build it in four steps.
Each step has its own test, and later steps can call your earlier functions.

### 1. Parse one line

Change `ParseLine` so it returns the level, the message and `ok`.

- Split the line with `strings.Fields`.
- A line needs at least **three** fields (timestamp, level and at least one
  message word). Otherwise return `"", "", false`.
- The level is the second field, converted to upper case, so `warn` becomes
  `WARN`.
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

### 2. Count lines per level

Change `CountLevels` so it returns a map from level to the number of lines
with that level. Skip lines where `ParseLine` returns `ok == false`.

```go
CountLevels([]string{
	"2025-01-02T10:00:00Z ERROR db timeout",
	"2025-01-02T10:00:01Z info request served",
	"2025-01-02T10:00:02Z ERROR db timeout",
	"garbage",
})
// => map[string]int{"ERROR": 2, "INFO": 1}
```

No lines (or only bad lines) gives an empty map. Always return a real map made
with `make`, never `nil`.

### 3. Format a summary

Change `Summary` so it turns the counts into one line of text. The starter
already has `standardLevels`, the four standard levels in order, but for now
it only joins their names. Make it follow these rules:

- Always show the four standard levels, in severity order: `ERROR`, `WARN`,
  `INFO`, `DEBUG`. Show `0` for a level that has no count.
- After them, show any other levels from the map (such as `TRACE`), in
  alphabetical order.
- Each entry is `LEVEL=count`. Entries are separated by a single space.

```go
Summary(map[string]int{"INFO": 5, "ERROR": 2, "WARN": 1})
// => "ERROR=2 WARN=1 INFO=5 DEBUG=0"

Summary(map[string]int{"ERROR": 1, "TRACE": 4, "AUDIT": 2})
// => "ERROR=1 WARN=0 INFO=0 DEBUG=0 AUDIT=2 TRACE=4"

Summary(map[string]int{})
// => "ERROR=0 WARN=0 INFO=0 DEBUG=0"
```

### 4. List the top messages for a level

Change `TopMessages` so it finds the most repeated messages for one level.

- Only look at valid lines whose level (as returned by `ParseLine`, so
  already upper case) equals `level`. The `level` argument
  is always upper case, such as `"ERROR"`.
- Count how many times each message appears.
- Order the messages by count, highest first. When two messages have the
  same count, put them in alphabetical order.
- Return at most `n` entries. If there are fewer messages, return them all.
  If `n` is 0, or nothing matches, return an empty slice: `[]string{}`, not
  `nil`.
- Format each entry as `<count>x <message>`, for example `"3x db timeout"`.

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
the tie because it comes first alphabetically. The `WARN` line is ignored.

## Learn more

- [strings.Fields](https://pkg.go.dev/strings#Fields)
- [slices.SortFunc](https://pkg.go.dev/slices#SortFunc)
- [maps.Keys](https://pkg.go.dev/maps#Keys)
- [Go by Example: Sorting by Functions](https://gobyexample.com/sorting-by-functions)
