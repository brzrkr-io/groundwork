# Variadic functions: fix the pager bot

Your pager bot builds alert messages from any number of hosts. Its helpers
were written in a hurry, and every one of them has a bug.

## Learn

### Recap

A `...T` parameter accepts zero or more values, and inside the function it is
a plain `[]T`. It must be the last parameter. To pass a slice you already
have, put `...` after it: `Alert("critical", hosts...)`. `Alert` builds its
message with a `strings.Builder`, as in *For loops: backoff and progress
bars*: `b.WriteString(...)` adds a piece and `b.String()` returns the whole
text.

### You can't mix single values and a spread slice

This looks reasonable, but it does not compile:

```go
Alert("warning", canary, others...) // compile error
```

After the fixed parameters, a call passes *either* separate values *or* one
slice with `...`, never both. To send one extra value plus a slice, build a
single slice first, as in *Variadic functions*:

```go
hosts := append([]string{canary}, others...)
Alert("warning", hosts...)
```

## Your task

Fix the pager bot's three helpers in `exercise.go`.

### 1. Fix the alert format

`Alert` should return the severity, a colon and a space, then the hosts
separated by `", "`. Right now it leaves a stray `", "` at the end. When there
are no hosts at all, it should return the severity followed by `": no hosts"`.

```go
Alert("critical", "web-1", "web-2") // => "critical: web-1, web-2"
Alert("warning", "db-1")            // => "warning: db-1"
Alert("critical")                   // => "critical: no hosts"
```

### 2. Page about every host

`PageOnCall` gets a slice of hosts that are down and should build a
`"critical"` alert listing **all** of them. Right now only the first host
makes it into the message. Pass the whole slice to `Alert`.

```go
PageOnCall([]string{"web-1", "web-2", "web-3"})
// => "critical: web-1, web-2, web-3"

PageOnCall(nil)
// => "critical: no hosts"
```

### 3. Put the canary first

`WarnCanary` should build a `"warning"` alert that lists the `canary` host
first, followed by every host in `others`. Right now it forgets `others`.

```go
WarnCanary("canary-1", []string{"web-1", "web-2"})
// => "warning: canary-1, web-1, web-2"

WarnCanary("canary-1", nil)
// => "warning: canary-1"
```

## Learn more

- [Go by Example: Variadic Functions](https://gobyexample.com/variadic-functions)
- [Go spec: Passing arguments to ... parameters](https://go.dev/ref/spec#Passing_arguments_to_..._parameters)
