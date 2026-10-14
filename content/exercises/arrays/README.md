# Arrays

*Practice: HTTP status triage* closed *Control flow and functions*: you can
make decisions and loop over a range of numbers. This module adds
collections, so one variable can hold many values.

Some data comes in a fixed number of pieces: seven days in a week, three
availability zones in a cloud region, 24 hours in a day. An *array* holds
exactly that many values of one type, side by side.

**New here:** array types like `[3]string`, array literals, indexing with
`a[i]`, `len`, arrays being copied.

## Learn

### Declaring an array

`[3]string` means "an array of exactly 3 strings". The length is part of the
type, so `[3]string` and `[4]string` are different types.

```go
var racks [4]string // ["" "" "" ""]
var retries [5]int  // [0 0 0 0 0]
var online [2]bool  // [false false]
```

A new array is filled with the *zero value* of its element type: `""` for
strings, `0` for numbers, `false` for bools.

> **Why is it written like this?** Go requires the brackets *before* the
> element type: `[3]string`, never `string[3]`. The number must be a constant
> the compiler knows. The variable name is your choice.

### Array literals

To fill an array as you create it, write the type, then the values in braces:

```go
disks := [3]string{"sda", "sdb", "sdc"}
```

Write `[...]` instead of a number and Go counts the values for you:

```go
months := [...]string{"Jan", "Feb", "Mar"} // type [3]string
```

> **Why is it written like this?** The braces and the commas between values
> are required. Spaces after commas are convention (gofmt adds them).

### Reading and writing one element

Each element has an *index*, its position, counting from `0`. `a[i]` reads
or changes one element. `len(a)` gives the number of elements.

```go
disks[0]          // "sda", the first element
disks[2]          // "sdc", the last one is at len-1
disks[1] = "nvme0" // change one element
len(disks)        // 3
```

`disks[3]` does not exist. With a plain number the compiler stops you; with
an index from a variable, the running program crashes with "index out of
range".

> **Why is it written like this?** Counting from `0` is required by Go (and
> most languages). `a[i] = value` changes the element in place; no `:=`,
> because the array already exists.

### Looping over an array

Combine `len` with the loops from *For loops*. `i` takes every index in turn:

```go
temps := [3]int{21, 25, 19}
for i := range len(temps) {
	fmt.Println(i, temps[i]) // 0 21, then 1 25, then 2 19
}
```

`for i := 0; i < len(temps); i++ { ... }` does the same thing.

### Arrays are copied

Assigning an array, or passing it to a function, copies *every element*:

```go
a := [3]int{1, 2, 3}
b := a    // b is a copy
b[0] = 99 // a is still [1 2 3], b is [99 2 3]
```

So a function that changes its array parameter only changes its own copy. To
hand the change back to the caller, it must `return` the array.

### Comparing arrays

Two arrays of the same type can be compared with `==`. They are equal when
every element is equal. The tests use this to check your results.

```go
[2]int{1, 2} == [2]int{1, 2} // true
[2]int{1, 2} == [2]int{1, 3} // false
```

## Your task

Your service runs in three availability zones of one cloud region, and you
keep a week of error counts. Write three helpers for the on-call dashboard.

### 1. Add up a week of errors

**Goal:** make `WeeklyTotal` return the sum of all seven numbers in `daily`.

Use a loop over the indexes. An all-zero week returns `0`.

```go
WeeklyTotal([7]int{3, 0, 1, 4, 0, 0, 2})
// => 10
```

**Think it through:**
- The starter already has `total` and returns it. What must happen to
  `total` once for each day?
- How many times should the loop run, and which index does it start at?

### 2. Name the zones of a region

**Goal:** make `ZoneNames` return an array of the three zone names for
`region`.

Each name is the `region` parameter followed by the fixed letter `"a"`, `"b"`
or `"c"`, in that order. No space or dash in between.

```go
ZoneNames("eu-west-1")
// => [3]string{"eu-west-1a", "eu-west-1b", "eu-west-1c"}
```

**Think it through:**
- Which piece comes from the parameter, and which piece is fixed text in
  quotes?
- You can fill `zones` one index at a time, or build the whole array with a
  literal. Either is fine.

### 3. Mark a zone as down

**Goal:** make `WithZoneDown` return `status` with the element at index
`zone` set to `false`.

`status` holds the health of the three zones (`true` = healthy). `zone` is
always `0`, `1` or `2`. Every other zone keeps its value, and a zone that is
already `false` stays `false`.

```go
WithZoneDown([3]bool{true, true, true}, 1)
// => [3]bool{true, false, true}
```

The test also checks that the caller's array did not change. Because arrays
are copied, you get this for free, as long as you return the changed array.

**Think it through:**
- Which parameter says *where* to change, and which one is the array?
- The starter already returns `status`. You only need one line before it.

## Learn more

- [A Tour of Go: Arrays](https://go.dev/tour/moretypes/6)
- [Go by Example: Arrays](https://gobyexample.com/arrays)
- [Go spec: Array types](https://go.dev/ref/spec#Array_types)
