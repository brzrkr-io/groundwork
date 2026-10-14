# Interfaces: swap in a fake for tests

A service reads its settings (database host, region, log level) from
somewhere. If your code asks for "a `Store`" instead of "this machine's
environment variables", you can test it with a fake store and stack
overrides on top of defaults.

**New here:** fakes for testing, empty structs, writing a method from scratch
to satisfy an interface; what an environment variable is.

## Learn

### Recap

As in *Interfaces*, an interface lists methods, and any type with those
methods satisfies it. A function with an interface parameter works with
every type that satisfies it, even types written later.

### Environment variables, in one minute

Every running program gets a small set of *environment variables* from
whoever started it: name/value pairs of text such as `HOME=/home/sam` or
`CACHE_URL=redis://cache:6379`. Servers and containers often get their
settings this way. Go reads one with `os.Getenv("CACHE_URL")`, which returns
`""` if it isn't set. The *Linux and the shell* module shows how to set them;
here you only need to know that their values depend on the machine.

### Interfaces make code testable

This function is hard to test, because its result depends on the machine:

```go
func CacheURL() string {
	return os.Getenv("CACHE_URL") // reads the real environment
}
```

This version asks for "anything with a `Lookup` method" instead:

```go
type Source interface {
	Lookup(name string) (string, bool)
}

func CacheURL(src Source) string {
	url, _ := src.Lookup("cache_url")
	return url
}
```

In production you pass a source that reads the environment. In a test you
pass a *fake*: a tiny in-memory version that returns exactly what the test
wants. Go programmers call this "accept interfaces": ask only for the methods
you need.

> **Why is it written like this?** Go requires nothing here; any function
> may take a concrete type. Taking an interface is a community convention
> that makes code easy to test.

### Reading the `Store` interface

```go
type Store interface {
	Get(key string) (value string, ok bool)
}
```

- `Get` is the method name; `key string` is its one parameter.
- `(value string, ok bool)` means it returns **two** results.

> **Why is it written like this?** The names `value` and `ok` are the
> author's choice, for documentation. A method satisfies `Store` with or
> without them: `Get(key string) (string, bool)` matches too. Only the
> name, parameter types and result types must match (Go requires that).
> Returning `ok bool` as the last result is a Go convention, as with maps.

### The pieces in this exercise

| Type | Where values come from | Used by |
|---|---|---|
| `EnvStore` | environment variables, via `os.LookupEnv` | the real service |
| `MemoryStore` | a `map[string]string` | the tests |
| `Layered` | other `Store`s, asked in order | both |

`EnvStore` is already written. `os.LookupEnv(key)` already returns a value
and whether it was set. `EnvStore` is an *empty struct*, `struct{}`: it holds
no data and exists only to carry its method.

### Writing a method from scratch

A method line has fixed parts (Go requires them) and parts you choose:

```go
func (r Router) Ping(host string) bool {
```

| Piece | Meaning | Required? |
|---|---|---|
| `func` | starts a function or method | Go requires |
| `(r Router)` | the *receiver*: which type gets the method, and its name inside | the type is required; the name `r` is your choice (convention: one short letter) |
| `Ping` | method name | must match the interface exactly |
| `(host string)` | parameters | types must match the interface; names are your choice |
| `bool` | result type(s) | must match the interface exactly |

## Your task

Finish the in-memory store, the function that loads a service's settings,
and a store that stacks overrides on top of defaults.

### 1. Write MemoryStore.Get

**Goal:** `MemoryStore.Get(key)` returns the value stored under `key` in
`m.Values`, and whether the key was there.

- Key present: return its value and `true`, even when the value is `""`.
- Key missing: return `""` and `false`.
- `MemoryStore{}` (no map at all) finds nothing.

```go
m := MemoryStore{Values: map[string]string{"region": "eu-west-1", "proxy": ""}}
m.Get("region")  // => "eu-west-1", true
m.Get("proxy")   // => "", true
m.Get("db_host") // => "", false
```

**Think it through:**
- From *Maps*: which form of map lookup gives you both the value and
  "was it there"?
- Does that already return exactly the two things `Get` needs?

### 2. Write Resolve

**Goal:** `Resolve(s, keys)` looks up each key in the store `s` and returns
a map from key to value, plus a `nil` error.

- With no keys, return an empty map (not `nil`) and `nil`.
- If a key is missing, stop right away and return `nil` and the error
  `fmt.Errorf("missing config key %q", key)`. The `%q` puts quotes around
  the key.
- Add `"fmt"` to the imports.

```go
store := MemoryStore{Values: map[string]string{"db_host": "db-1.internal", "db_port": "5432"}}

Resolve(store, []string{"db_host", "db_port"})
// => map[db_host:db-1.internal db_port:5432], nil

Resolve(store, []string{"db_host", "db_password"})
// => nil, error `missing config key "db_password"`
```

The test uses `MemoryStore`, so finish task 1 first.

**Think it through:**
- You need a new empty map before the loop. How did *Maps* create one?
- `s` is a `Store`. Which method can you call on it, and what two things
  does it give back?

### 3. Make Layered a Store

**Goal:** write a `Get` method for `Layered`, from scratch, so `Layered`
satisfies `Store`.

- Ask each store in `l.Layers` in order. Return the first value found, with
  `true`.
- If no layer has the key, or there are no layers, return `""` and `false`.

```go
overrides := MemoryStore{Values: map[string]string{"log_level": "debug"}}
defaults := MemoryStore{Values: map[string]string{"log_level": "info", "region": "eu-west-1"}}
l := Layered{Layers: []Store{overrides, defaults}}

l.Get("log_level") // => "debug", true     (overrides win)
l.Get("region")    // => "eu-west-1", true (falls back to defaults)
l.Get("db_host")   // => "", false
```

The test checks `any(Layered{...}).(Store)` as in *Interfaces*, so the file
compiles before you write the method.

**Think it through:**
- Copy the method line shape from `MemoryStore.Get`. Which part changes for
  `Layered`?
- Each layer is a `Store`. When should the loop return early, and what
  happens after the loop ends?

Because `Layered` is itself a `Store`, you can pass it to `Resolve`. In the
real service the layers might be `[]Store{EnvStore{}, defaults}`:
environment variables override the defaults.

## Learn more

- [Go Code Review Comments: Interfaces](https://go.dev/wiki/CodeReviewComments#interfaces)
- [Package os: LookupEnv](https://pkg.go.dev/os#LookupEnv)
- [Effective Go: Generality](https://go.dev/doc/effective_go#generality)
