# Arrays: weekly on-call summary

Every Monday your on-call team reviews last week's pages (alerts that woke
someone up). You will write the helper that turns two weeks of page counts
into a one-line summary for the team chat.

## Learn

Recap from *Arrays*: `[7]int` holds exactly seven ints. Loop with
`for i := range len(a)` and read or write `a[i]`. Arrays are copied when you
assign or pass them, and two arrays of the same type can be compared with
`==`. A function can return a whole array.

The starter gives you `dayNames`, a `[7]string` that maps an index to a day:
`dayNames[0]` is `"Mon"` and `dayNames[6]` is `"Sun"`. Index `i` in a week of
pages always means the same day as `dayNames[i]`.

### New: printing the sign of a number

In `fmt.Sprintf`, the verb `%+d` prints an int *with* its sign, even when it
is positive or zero:

```go
fmt.Sprintf("%+d", 3)  // "+3"
fmt.Sprintf("%+d", -2) // "-2"
fmt.Sprintf("%+d", 0)  // "+0"
```

You need `import "fmt"` for task 3. Add it at the top of `exercise.go`.

## Your task

Each function takes one or two weeks of page counts, Monday first. Small
helper functions of your own are welcome, and later tasks may call earlier
ones.

### 1. Day-by-day change

Change `DailyChange(last, this [7]int) [7]int` so element `i` of the result is
`this[i] - last[i]`.

```go
DailyChange([7]int{2, 0, 1, 0, 0, 0, 0}, [7]int{1, 3, 1, 0, 0, 0, 2})
// => [7]int{-1, 3, 0, 0, 0, 0, 2}
```

### 2. Busiest day

Change `BusiestDay(pages [7]int) string` so it returns the name of the day
with the most pages, taken from `dayNames`. If several days tie, return the
earliest. A week with no pages at all returns `"Mon"`. The starter already
returns `dayNames[best]`; your job is to find the right `best`.

```go
BusiestDay([7]int{1, 4, 0, 4, 2, 0, 0}) // => "Tue"
BusiestDay([7]int{0, 0, 0, 0, 0, 0, 5}) // => "Sun"
```

### 3. The weekly summary

Change `WeekSummary(last, this [7]int) string`. Let `total` be the sum of
this week's pages and `change` be `total` minus the sum of last week's pages.

- If `this` is exactly the same as `last` (every day equal), return
  `"same as last week: <total> pages"`.
- Otherwise return `"<total> pages (<change> vs last week), busiest <day>"`,
  where `<change>` is printed with `%+d` and `<day>` is this week's busiest
  day from task 2.

```go
WeekSummary([7]int{1, 2, 0, 0, 0, 0, 0}, [7]int{1, 2, 0, 0, 0, 0, 0})
// => "same as last week: 3 pages"

WeekSummary([7]int{2, 0, 1, 0, 0, 0, 0}, [7]int{1, 3, 1, 0, 0, 0, 2})
// => "7 pages (+4 vs last week), busiest Tue"

WeekSummary([7]int{3, 0, 0, 0, 0, 0, 0}, [7]int{0, 0, 0, 0, 3, 0, 0})
// => "3 pages (+0 vs last week), busiest Fri"
```

The last example has the same total both weeks, but the days differ, so it is
not "same as last week".

## Learn more

- [Go by Example: Arrays](https://gobyexample.com/arrays)
- [Go spec: Comparison operators](https://go.dev/ref/spec#Comparison_operators)
- [fmt package: printing verbs](https://pkg.go.dev/fmt#hdr-Printing)
