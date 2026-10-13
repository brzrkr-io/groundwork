# Interfaces: swap in a fake for tests

A service reads its settings (database host, region, log level) from
somewhere. If your code asks for "a `Store`" instead of "the environment
variables of this machine", you can test it with a fake store and layer
overrides on top of defaults.

## Learn

### Recap

As in *Interfaces*, an interface lists methods, and any type with those
methods satisfies it. A function that takes an interface parameter works with
every type that satisfies it, including types written long after the
function.

### Interfaces make code testable

Here is a function that is hard to test:

```go
func DatabaseHost() string {
	return os.Getenv("DB_HOST") // reads the real environment
}
```

Its result depends on the machine it runs on. Now look at this version:

```go
type Store interface {
	Get(key string) (value string, ok bool)
}

func DatabaseHost(s Store) string {
	host, _ := s.Get("db_host")
	return host
}
```

`DatabaseHost` no longer knows where settings come from. In production you
pass a store that reads the environment. In a test you pass a *fake*: a tiny
in-memory version that returns exactly the values the test wants. The same
trick lets you test code that talks to a cloud API, a database or the
Kubernetes API without touching any of them. Go programmers sum this up as
"accept interfaces": a function should ask for the smallest set of methods it
needs.

### The pieces in this exercise

| Type | Where values come from | Used by |
|---|---|---|
| `EnvStore` | environment variables, through `os.LookupEnv` | the real service |
| `MemoryStore` | a `map[string]string` | the tests |
| `Layered` | other `Store`s, checked in order | both |

`EnvStore` is already written. `os.LookupEnv(key)` returns the variable's
value and whether it is set, which already matches `Get`. `EnvStore` is an
*empty struct*, `struct{}`: it holds no data and exists only to carry the
method.

## Your task

Finish the in-memory store, the function that loads a service's settings,
and a store that layers overrides on top of defaults.

### 1. Write MemoryStore.Get

`Get` looks `key` up in `m.Values`. It returns the value and `true` if the key
is there (even when the value is an empty string), and `""` and `false` if
not:

```go
m := MemoryStore{Values: map[string]string{"region": "eu-west-1", "proxy": ""}}
m.Get("region")  // => "eu-west-1", true
m.Get("proxy")   // => "", true
m.Get("db_host") // => "", false
```

A `MemoryStore` with no map at all (`MemoryStore{}`) finds nothing.

### 2. Write Resolve

`Resolve(s Store, keys []string) (map[string]string, error)` looks up each
key in `s` and returns a map from key to value. With no keys it returns an
empty map.

If a key is missing, it stops and returns `nil` and an error built with
`fmt.Errorf("missing config key %q", key)`. Add `"fmt"` to the imports.

```go
store := MemoryStore{Values: map[string]string{"db_host": "db-1.internal", "db_port": "5432"}}

Resolve(store, []string{"db_host", "db_port"})
// => map[db_host:db-1.internal db_port:5432], nil

Resolve(store, []string{"db_host", "db_password"})
// => nil, error `missing config key "db_password"`
```

The test uses `MemoryStore`, so finish task 1 first.

### 3. Make Layered a Store

Write a `Get` method for `Layered`, from scratch, so it satisfies `Store`. It
asks each store in `l.Layers` in order and returns the first value found,
with `true`. If no layer has the key, or there are no layers, it returns `""`
and `false`.

```go
overrides := MemoryStore{Values: map[string]string{"log_level": "debug"}}
defaults := MemoryStore{Values: map[string]string{"log_level": "info", "region": "eu-west-1"}}
l := Layered{Layers: []Store{overrides, defaults}}

l.Get("log_level") // => "debug", true    (overrides win)
l.Get("region")    // => "eu-west-1", true (falls back to defaults)
l.Get("db_host")   // => "", false
```

As in *Interfaces*, the test checks this with
`any(Layered{...}).(Store)`, so the file compiles before you write the method.

Because `Layered` is itself a `Store`, you can pass it to `Resolve`, or even
put one `Layered` inside another. In the real service the layers might be
`[]Store{EnvStore{}, defaults}`: environment variables override the defaults.

## Learn more

- [Go Code Review Comments: Interfaces](https://go.dev/wiki/CodeReviewComments#interfaces)
- [Package os: LookupEnv](https://pkg.go.dev/os#LookupEnv)
- [Effective Go: Generality](https://go.dev/doc/effective_go#generality)
