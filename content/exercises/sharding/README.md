# Sharding: route keys and plan a rebalance

When one database can't hold all your data, you split it across several,
called *shards*. You'll write the rule that decides which shard each key
lives on, and find out why adding one more shard is harder than it sounds.

**New here:** sharding, shard keys, hashing with `hash/fnv`, `uint32`
conversions, slices of slices, rebalancing.

## Learn

### What sharding is

Your deploy history has one row per deploy of every service, and one
Postgres server can no longer keep up. *Sharding* splits the rows across `n`
databases, numbered `0` to `n-1`. Each one holds only part of the data, so
each one does only part of the work.

Every row needs a rule for which shard it lives on. The rule looks at one
field, the *shard key*. Here it's the service name, so all deploys of `api`
sit together on one shard, and "show me api's deploy history" asks just that
one database.

### Why not the first letter?

The obvious rule: names starting `a` to `m` go to shard 0, `n` to `z` to
shard 1. But names aren't spread evenly:

| Service | First letter | Shard |
|---|---|---|
| `payments-api` | p | 1 |
| `payments-worker` | p | 1 |
| `payments-cron` | p | 1 |
| `platform-ui` | p | 1 |
| `auth` | a | 0 |

Shard 1 gets four services, shard 0 gets one. An overloaded shard is a *hot
shard*: it's slow while the others sit idle, and adding shards doesn't help
it.

### Hash, then modulo

A *hash function* turns any string into a number that looks random, but is
always the same for the same string. Similar names get completely different
numbers. Go's `hash/fnv` package has *FNV-1a*, a small, fast hash.
`fnv.New32a` makes a 32-bit one:

```go
h := fnv.New32a()
h.Write([]byte("eu-west-1"))
h.Sum32() // => 1346204939 (a uint32)
```

The *modulo* (`%`) of that number by the shard count gives a shard from `0`
to `n-1`: `1346204939 % 4` is `3`. Spread 1,000 keys (`"service-0"` to
`"service-999"`) this way over 4 shards and you get 249, 251, 251 and 249
keys per shard.

> **Why is it written like this?** `Write` takes a `[]byte`, so Go requires
> the conversion from `string`. `Write` also returns a count and an error,
> but the docs promise a hash's `Write` never fails, so calling it without
> using its results is the convention. `Sum32` returns a `uint32`, and Go
> never mixes `uint32` and `int` in one expression: convert with `uint32(x)`
> and back with `int(x)`. Picking FNV-1a is the author's choice; what
> matters is that every program that routes keys uses the **same** hash.

### Slices of slices

A `[][]string` is a slice whose elements are `[]string`. `make` creates
the outer one with `n` empty (`nil`) inner slices, and `append` works on any
of them by index:

```go
byHour := make([][]string, 24) // 24 empty slices, one per hour
byHour[9] = append(byHour[9], "deploy api")
byHour[9] = append(byHour[9], "deploy web")
len(byHour)    // => 24
len(byHour[9]) // => 2
```

### Rebalancing, and why modulo makes it hard

Your shards fill up, so you go from 4 shards to 5. A key stays put only if
`hash % 4` equals `hash % 5`. For most hashes it doesn't: a hash of `7` was
on shard `3` and is now on shard `2`. Each key that moves means copying its
rows to another database while the service keeps running.

| From | To | Share of keys that move |
|---|---|---|
| 4 | 5 | about 80% |
| 10 | 11 | about 91% |
| 4 | 8 | about 50% |

Going from `n` to `n+1` shards moves about `n/(n+1)` of the keys: almost
all of them. That's why the next exercise,
*Consistent hashing: a ring with virtual nodes*, exists: it adds a shard
while moving only about `1/(n+1)` of the keys.

### Dividing to get a fraction

Dividing two `int`s throws the remainder away: `787 / 1000` is `0`. Convert
both to `float64` first:

```go
failed, total := 3, 12
float64(failed) / float64(total) // => 0.25
```

## Your task

Your deploy-history service is about to split its Postgres database into
shards by service name. Write the routing rule, then work out what a
rebalance would cost.

### 1. Pick a shard for a key

**Goal:** make `ShardFor(key, n)` return the 32-bit FNV-1a hash of `key`,
modulo `n`, as an `int`.

Hash `key` with `fnv.New32a()`, take `Sum32()`, and use `%` with `n`. Add
`"hash/fnv"` to the imports. `n` is always at least 1.

```go
ShardFor("api", 4)     // => 3
ShardFor("billing", 4) // => 0
ShardFor("api", 8)     // => 7
ShardFor("api", 1)     // => 0
```

**Think it through:**
- `Write` wants bytes, and you have a `string`. Which conversion fixes that?
- `Sum32()` is a `uint32` and `n` is an `int`. Which one do you convert
  before `%`, and what do you convert the result to?

### 2. Group keys by shard

**Goal:** make `Group(keys, n)` return `n` slices, where slice `i` holds the
keys that `ShardFor` puts on shard `i`.

The result always has exactly `n` slices, even for shards with no keys (an
empty or `nil` slice is fine). Inside each slice, keys keep the order they
have in `keys`.

```go
Group([]string{"api", "billing", "checkout", "payments", "web"}, 4)
// => [["billing"] ["checkout" "web"] ["payments"] ["api"]]
```

**Think it through:**
- How do you make a slice of `n` slices before the loop?
- `ShardFor` already tells you the index. Which slice do you `append` to?

### 3. Plan a rebalance

**Goal:** make `Plan(keys, from, to)` return a `Move` for every key whose
shard changes when the shard count goes from `from` to `to`.

Each `Move` holds the key, its shard with `from` shards (`From`) and its
shard with `to` shards (`To`). Keys that stay on the same shard are left
out. Moves come in the order of `keys`. When nothing moves, return an empty
or `nil` slice.

```go
Plan([]string{"api", "checkout", "payments"}, 4, 5)
// => [{Key:api From:3 To:2} {Key:payments From:2 To:3}]
//    (checkout is on shard 1 both times, so it stays)
```

**Think it through:**
- For each key, you need two calls to `ShardFor`. What changes between
  them?
- When do you `append`, and what goes in the `Move`?

### 4. Measure how much moves

**Goal:** make `MovedFraction(keys, from, to)` return the share of keys that
move, from `0` to `1`.

That's the number of moves from `Plan` divided by the number of keys. For no
keys at all, return `0` (dividing by zero gives `NaN`, which is no answer).

```go
MovedFraction([]string{"api", "billing", "checkout", "payments", "search",
	"auth", "web", "worker", "deploy-bot", "ledger"}, 4, 5) // => 0.8
// 1,000 keys "service-0" ... "service-999":
MovedFraction(keys, 4, 5) // => 0.787
MovedFraction(keys, 4, 8) // => 0.501
MovedFraction(keys, 4, 4) // => 0
```

**Think it through:**
- Which function already lists the moves? What does `len` give you?
- Which type do both numbers need before you divide?

## Learn more

- [pkg.go.dev: hash/fnv](https://pkg.go.dev/hash/fnv)
- [Wikipedia: Fowler–Noll–Vo hash function](https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function)
- [Wikipedia: Shard (database architecture)](https://en.wikipedia.org/wiki/Shard_(database_architecture))
- [AWS: What is database sharding?](https://aws.amazon.com/what-is/database-sharding/)
