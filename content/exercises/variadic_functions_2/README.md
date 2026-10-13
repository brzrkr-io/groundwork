# Variadic functions: fix the pager bot

Your pager bot builds alert messages from any number of hosts. Its helpers
were written in a hurry, and every one of them has a bug.

**New here:** you can't mix single values and a spread slice in one call.

## Learn

### Recap

- A `...T` parameter accepts zero or more values. Inside the function it is
  a plain `[]T`. It must be the last parameter.
- To pass a slice you already have, put `...` after it: `Total(counts...)`.
- `Alert` builds text with a `strings.Builder`, as in *For loops: backoff and
  progress bars*: `b.WriteString(...)` adds a piece, `b.String()` returns the
  whole text.

### Separators go *between* items

Writing `", "` after every item leaves one hanging at the end. A common fix:
write the first item on its own, then write the separator *before* each of
the remaining items. Here with ports and `" | "`:

```go
var b strings.Builder
b.WriteString(ports[0])
for _, p := range ports[1:] {
	b.WriteString(" | ")
	b.WriteString(p)
}
// ports {"80", "443"} => "80 | 443"
```

This uses `ports[0]`, so it needs at least one item. Check the empty case
first.

### You can't mix single values and a spread slice

This looks reasonable, but it does not compile:

```go
Total(5, more...) // compile error
```

After the fixed parameters, a call passes *either* separate values *or* one
slice with `...`, never both. To send one extra value plus a slice, build a
single slice first with `append`, as in *Variadic functions*, then spread
that.

> **Why is it written like this?** Go requires this. With `slice...`, the
> function receives that very slice, not a copy, so there's no room to slip
> another value in front of it.

## Your task

Fix the pager bot's three helpers in `exercise.go`.

### 1. Fix the alert format

**Goal:** make `Alert` return `severity`, then `": "`, then the hosts separated by `", "`, with nothing left over at the end.

```go
Alert("critical", "web-1", "web-2") // => "critical: web-1, web-2"
Alert("warning", "db-1")            // => "warning: db-1"
Alert("critical")                   // => "critical: no hosts"
```

- `severity` and the host names come from the parameters.
- `": "`, `", "` and `"no hosts"` are fixed text.
- With no hosts at all, return `severity` followed by `": no hosts"`.

**Think it through:**
- Where does the stray `", "` at the end come from in the current loop?
- What should happen before the loop when `hosts` is empty?

### 2. Page about every host

**Goal:** make `PageOnCall` return a `"critical"` alert that lists **every** host in `down`.

```go
PageOnCall([]string{"web-1", "web-2", "web-3"})
// => "critical: web-1, web-2, web-3"

PageOnCall(nil)
// => "critical: no hosts"
```

Right now only `down[0]` reaches `Alert`. Let `Alert` do the work: it already
handles any number of hosts, including none.

**Think it through:**
- `down` is a slice, but `Alert` wants separate hosts. How do you spread a
  slice into a variadic call?
- Once that works, do you still need the `if`?

### 3. Put the canary first

**Goal:** make `WarnCanary` return a `"warning"` alert listing `canary` first, then every host in `others`.

```go
WarnCanary("canary-1", []string{"web-1", "web-2"})
// => "warning: canary-1, web-1, web-2"

WarnCanary("canary-1", nil)
// => "warning: canary-1"
```

Right now it forgets `others`.

**Think it through:**
- Why can't you just add `others...` after `canary` in the call?
- How can you make *one* slice that holds `canary` followed by `others`?

## Learn more

- [Go by Example: Variadic Functions](https://gobyexample.com/variadic-functions)
- [Go spec: Passing arguments to ... parameters](https://go.dev/ref/spec#Passing_arguments_to_..._parameters)
