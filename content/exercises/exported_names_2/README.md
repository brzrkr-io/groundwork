# Exported names: design a small API

Your platform limits how many production deploys each team may run per day.
The `quota` package that enforces it works, but its API lets callers cheat.
You'll tighten it up.

**New here:** naming exported things well, getters.

## Learn

### Recap

- Capitalised names are exported. Lower-case names are private to their package.
- Structs with private fields come with a constructor.
- Tests in `package x_test` see only exported names.

### Names read with their package

Callers always write the package name in front of what you export. So
choose names that read well *together with* the package name:

| Stutters | Reads well |
|---|---|
| `cache.NewCache(10)` | `cache.New(10)` |
| `cache.CacheError` | `cache.ErrMiss` |

> **Why is it written like this?**
> - **Convention:** when a package is built around one type, its constructor
>   is just `New`. The standard library does this too: `errors.New`,
>   `list.New`.
> - **Go requires** nothing here. `NewCache` compiles fine. It just repeats
>   itself every time someone calls it.

### Getters

To let callers *read* a private field without being able to *change* it, add
a method that returns it:

```go
type Server struct {
	port int
}

func (s *Server) Port() int {
	return s.port
}
```

Callers write `srv.Port()` to read it. They have no way to set it.

> **Why is it written like this?**
> - **Convention:** Go getters don't use a `Get` prefix. The getter for
>   `port` is `Port()`, not `GetPort()`.
> - **Go requires** different names for a struct's fields and methods. A
>   `Port` field and a `Port()` method can't both exist. That is why the
>   field becomes lower case first.
> - **Go requires** the `(s *Server)` part, the *receiver*, as in *Methods*.
>   The name `s` is the author's choice; convention is one or two letters
>   from the type name.

### Exported errors

An error that callers check with `errors.Is` must be exported, as you saw in
*Errors are values*. That's why `ErrExhausted` starts with a capital letter.

## Your task

```text
.
├── go.mod
└── quota/
    ├── quota.go          package quota (you edit this)
    └── quota_test.go     package quota_test
```

The test file only compiles once tasks 1 and 2 are both done. Until then
you'll see compiler errors. Each error points at one of these tasks.

### 1. Rename the constructor

**Goal:** rename `NewQuota` to `New`, so callers write `quota.New(5)`.

This fixes `undefined: quota.New`. Update the doc comment so it starts with
the new name.

**Think it through:**
- Only the function's name changes. Its parameter, return type and body stay.

### 2. Make the fields private and add getters

**Goal:** `Quota` has no exported fields, and two methods on `*Quota` let callers read them: `Limit() int` and `Used() int`.

Right now a caller can set `q.Limit = 1000` and deploy as often as they
like. The tests call getters instead, so the compiler reports:

```text
invalid operation: cannot call q.Limit (variable of type int): int is not a function
q.Used undefined (type *quota.Quota has no field or method Used, but does have unexported field used)
```

- `Limit()` returns the daily limit.
- `Used()` returns how many deploys have been taken.

```go
q := quota.New(5)
q.Limit() // => 5
q.Used()  // => 0
```

`TestNew` checks the getters. `TestFieldsArePrivate` checks that `Quota` has
no exported fields at all.

**Think it through:**
- Rename the `Limit` field first. Which other line in `quota.go` uses that field?
- Each getter returns one field. Which field does each one return?
- Give each getter a doc comment that starts with its name.

### 3. Enforce the limit in Take

**Goal:** `Take` refuses with `ErrExhausted` once the limit is reached.

- If the used count has already reached the limit, return `ErrExhausted`
  and don't change the count.
- Otherwise add one to the count and return `nil`.

```go
q := quota.New(2)
q.Take() // => nil
q.Take() // => nil
q.Take() // => ErrExhausted, and q.Used() is still 2
```

A limit of 0 means the very first `Take` returns `ErrExhausted`.

**Think it through:**
- Check *before* you count. Where does the `if` go relative to `q.used++`?
- Which comparison is true both when `used` equals the limit and when the limit is 0?

## Learn more

- [Effective Go: Getters](https://go.dev/doc/effective_go#Getters)
- [Effective Go: Package names](https://go.dev/doc/effective_go#package-names)
- [Go blog: Package names](https://go.dev/blog/package-names)
