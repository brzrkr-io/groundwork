# Arrays

Some data comes in a fixed number of pieces: seven days in a week, three
availability zones in a cloud region, 24 hours in a day. An *array* holds
exactly that many values of one type, side by side.

## Learn

### Declaring an array

The type `[3]string` means "an array of exactly 3 strings". The length is part
of the type, so `[3]string` and `[4]string` are different types.

```go
var zones [3]string // ["" "" ""]
var errors [7]int   // [0 0 0 0 0 0 0]
var healthy [3]bool // [false false false]
```

A freshly declared array is filled with the *zero value* of its element type,
just like a plain variable: `""` for strings, `0` for numbers, `false` for
bools.

### Array literals

To fill an array as you create it, write an *array literal*: the type followed
by the values in braces.

```go
zones := [3]string{"eu-west-1a", "eu-west-1b", "eu-west-1c"}
```

If you write `[...]` instead of a number, Go counts the values for you:

```go
days := [...]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
// days has type [7]string
```

### Reading and writing elements

Each value has an *index*, its position, counting from `0`. Use `a[i]` to read
or change one element. `len(a)` gives the number of elements.

```go
zones[0]            // "eu-west-1a" (the first element)
zones[2]            // "eu-west-1c" (the last element is at len-1)
zones[1] = "eu-west-1z"
len(zones)          // 3
```

Index `3` does not exist in a `[3]string`. If you write `zones[3]`, the
compiler stops you. If the index is in a variable and is out of range when the
program runs, the program crashes with an "index out of range" error.

### Looping over an array

Combine `len` with the loops from *For loops*:

```go
counts := [3]int{4, 0, 2}
for i := range len(counts) {
	fmt.Println(i, counts[i]) // 0 4, then 1 0, then 2 2
}
```

`for i := 0; i < len(counts); i++ { ... }` does the same thing.

### Arrays are values

When you assign an array to another variable, or pass it to a function, Go
copies *every element*. The two arrays are independent afterwards:

```go
a := [3]int{1, 2, 3}
b := a    // b is a copy
b[0] = 99
// a is still [1 2 3], b is [99 2 3]
```

The same happens with function parameters. A function that changes its array
parameter only changes its own copy. To hand the change back, it must
`return` the array.

### Comparing arrays

Two arrays of the same type can be compared with `==` and `!=`. They are equal
when every element is equal:

```go
[3]int{1, 2, 3} == [3]int{1, 2, 3} // true
[3]int{1, 2, 3} == [3]int{1, 2, 4} // false
```

The tests in this exercise use `==` to check your arrays.

## Your task

Your team runs a service in three availability zones of one cloud region and
keeps a week of error counts. Write three small helpers for the on-call
dashboard.

### 1. Add up a week of errors

Change `WeeklyTotal(daily [7]int) int` so it returns the sum of all seven daily
error counts. Use a loop over the indexes.

```go
WeeklyTotal([7]int{3, 0, 1, 4, 0, 0, 2})
// => 10
```

### 2. Name the zones of a region

Change `ZoneNames(region string) [3]string` so it returns the region name with
`"a"`, `"b"` and `"c"` added, in that order.

```go
ZoneNames("eu-west-1")
// => [3]string{"eu-west-1a", "eu-west-1b", "eu-west-1c"}
```

### 3. Mark a zone as down

`WithZoneDown(status [3]bool, zone int) [3]bool` takes the health of the three
zones (`true` means healthy) and the index of a zone that just went down. Set
that zone to `false` and return the updated array. `zone` is always `0`, `1`
or `2`.

```go
WithZoneDown([3]bool{true, true, true}, 1)
// => [3]bool{true, false, true}
```

Because arrays are copied, the caller's array stays exactly as it was. The test
checks that too. You get this for free, as long as you `return` the changed
array.

## Learn more

- [A Tour of Go: Arrays](https://go.dev/tour/moretypes/6)
- [Go by Example: Arrays](https://gobyexample.com/arrays)
- [Go spec: Array types](https://go.dev/ref/spec#Array_types)
