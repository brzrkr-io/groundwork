# Arrays: weekly on-call summary

Every Monday your on-call team reviews last week's *pages* (alerts that woke
someone up). You will write the helper that turns two weeks of page counts
into a one-line summary for the team chat.

**New here:** the `%+d` verb; otherwise practice.

## Learn

### Recap from *Arrays*

- `[7]int` holds exactly seven ints. Loop with `for i := range len(a)` and use
  `a[i]`.
- Two arrays of the same type can be compared with `==`: equal means every
  element is equal.
- A function can take and return whole arrays.

### The `dayNames` array

The starter gives you `dayNames`, a `[7]string`: `dayNames[0]` is `"Mon"`,
`dayNames[6]` is `"Sun"`. Index `i` in a week of pages is the same day as
`dayNames[i]`.

### Printing a number with its sign

In `fmt.Sprintf`, the verb `%+d` prints an int *with* its sign, even when it
is positive or zero:

```go
fmt.Sprintf("disk %+d GB", 3)  // "disk +3 GB"
fmt.Sprintf("disk %+d GB", -2) // "disk -2 GB"
fmt.Sprintf("disk %+d GB", 0)  // "disk +0 GB"
```

Plain `%d` prints `3`, `-2`, `0`. `%s` prints a string.

> **Why is it written like this?** The verbs (`%d`, `%+d`, `%s`) are fixed by
> the `fmt` package. The values after the format string fill the verbs in
> order, one value per verb. Everything else in the quotes is your own text,
> copied exactly, spaces included.

You need `import "fmt"` at the top of `exercise.go` for task 3. Go requires
imports right after the `package` line.

### Helper functions

You may write small functions of your own (for example, one that adds up a
week), and later tasks may call earlier ones. Lowercase names stay private to
the package; that's a Go rule.

## Your task

Each function takes one or two weeks of page counts, Monday first. When there
are two, the **first argument is last week** and the second is this week:
`DailyChange(last, this)`, `WeekSummary(last, this)`.

### 1. Day-by-day change

**Goal:** make `DailyChange` return an array where element `i` is
`this[i] - last[i]`.

```go
DailyChange([7]int{2, 0, 1, 0, 0, 0, 0}, [7]int{1, 3, 1, 0, 0, 0, 2})
// => [7]int{-1, 3, 0, 0, 0, 0, 2}
```

**Think it through:**
- The starter has an empty `change` array and returns it. What goes into each
  of its seven slots?
- Order matters: this week minus last week.

### 2. Busiest day

**Goal:** make `BusiestDay` return the name of the day with the most pages.

If several days tie, return the earliest. A week with no pages returns
`"Mon"`. The starter already returns `dayNames[best]`; your job is to find the
right `best`.

```go
BusiestDay([7]int{1, 4, 0, 4, 2, 0, 0}) // => "Tue"
BusiestDay([7]int{0, 0, 0, 0, 0, 0, 5}) // => "Sun"
```

**Think it through:**
- `best` is an index. How do you get the page count for the day at `best`?
- When should a later day replace `best`? Only when it is strictly bigger.
- This is the same idea as the peak hour in *Arrays: fix the hourly histogram*.

### 3. The weekly summary

**Goal:** make `WeekSummary` return a one-line summary of `this` week
compared with `last` week.

Let `total` be the sum of this week's pages, and `change` be `total` minus
the sum of last week's pages.

- If `this` equals `last` (every day equal), return
  `same as last week: <total> pages`.
- Otherwise return `<total> pages (<change> vs last week), busiest <day>`.
  `<change>` is printed with `%+d`. `<day>` is this week's busiest day from
  task 2.

```go
WeekSummary([7]int{1, 2, 0, 0, 0, 0, 0}, [7]int{1, 2, 0, 0, 0, 0, 0})
// => "same as last week: 3 pages"

WeekSummary([7]int{2, 0, 1, 0, 0, 0, 0}, [7]int{1, 3, 1, 0, 0, 0, 2})
// => "7 pages (+4 vs last week), busiest Tue"

WeekSummary([7]int{4, 4, 4, 0, 0, 0, 0}, [7]int{0, 1, 0, 0, 2, 0, 0})
// => "3 pages (-9 vs last week), busiest Fri"

WeekSummary([7]int{3, 0, 0, 0, 0, 0, 0}, [7]int{0, 0, 0, 0, 3, 0, 0})
// => "3 pages (+0 vs last week), busiest Fri"
```

The last example has the same total both weeks, but the days differ, so it
is not "same as last week". Two empty weeks give
`"same as last week: 0 pages"`.

**Think it through:**
- Which three values fill the slots: one total, one difference, one day name?
  Which earlier function gives you the day?
- Which text is fixed (copy it exactly, spaces and brackets included), and
  which parts come from your values?
- How do you check "every day equal" in one comparison?

## Learn more

- [Go by Example: Arrays](https://gobyexample.com/arrays)
- [Go spec: Comparison operators](https://go.dev/ref/spec#Comparison_operators)
- [fmt package: printing verbs](https://pkg.go.dev/fmt#hdr-Printing)
