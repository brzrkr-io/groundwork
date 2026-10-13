# Exported names: design a small API

Your platform limits how many production deploys each team may run per day.
The `quota` package that enforces it works, but its API lets callers cheat.
You'll tighten it up.

## Learn

### Recap

Names that start with a capital letter are exported and visible to other
packages. Lower-case names are private to their package. Structs with
private fields come with a constructor, and tests in `package x_test` see
only exported names.

### Names read with their package

Callers always write the package name in front of what you export, so
choose names that read well together:

| Stutters | Reads well |
|---|---|
| `quota.NewQuota(5)` | `quota.New(5)` |
| `quota.QuotaError` | `quota.ErrExhausted` |

When a package is built around one type, its constructor is usually just
`New`.

### Getters

To let callers read a private field without changing it, add a method that
returns it. Go getters don't use a `Get` prefix: the getter for `limit` is
`Limit()`, not `GetLimit()`.

```go
func (q *Quota) Limit() int {
	return q.limit
}
```

A struct can't have a field and a method with the same name. To add a
`Limit()` method you must first rename the `Limit` field to `limit`.

### Exported errors

An error that callers need to check with `errors.Is` must be exported, as you
saw in *Errors are values*. That's why `ErrExhausted` starts with a capital
letter.

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

Rename `NewQuota` to `New`. This fixes `undefined: quota.New`.

### 2. Make the fields private and add getters

A caller can set `q.Limit = 1000` and deploy as often as they like. The tests
call getters instead, so the compiler currently reports:

```text
invalid operation: cannot call q.Limit (variable of type int): int is not a function
q.Used undefined (type *quota.Quota has no field or method Used, but does have unexported field used)
```

Rename the field `Limit` to `limit`, then add two methods on `*Quota`:

- `Limit() int` returns the daily limit.
- `Used() int` returns how many deploys have been taken.

```go
q := quota.New(5)
q.Limit() // => 5
q.Used()  // => 0
```

`TestNew` checks the getters. `TestFieldsArePrivate` checks that `Quota` has
no exported fields at all.

### 3. Enforce the limit in Take

`Take` records one deploy. Right now it never says no. Change it so that:

- If `Used()` has already reached `Limit()`, it returns `ErrExhausted` and
  doesn't change the count.
- Otherwise it adds one to the count and returns `nil`.

```go
q := quota.New(2)
q.Take() // => nil
q.Take() // => nil
q.Take() // => ErrExhausted, and q.Used() is still 2
```

A limit of 0 means the very first `Take` returns `ErrExhausted`.

## Learn more

- [Effective Go: Getters](https://go.dev/doc/effective_go#Getters)
- [Effective Go: Package names](https://go.dev/doc/effective_go#package-names)
- [Go blog: Package names](https://go.dev/blog/package-names)
