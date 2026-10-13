# Methods: certificate checks

Expired TLS certificates are a classic cause of outages. In this exercise a
teammate wrote methods that check certificates, and two of them have bugs at
the boundaries, exactly where real bugs like to hide.

## Learn

### Recap

A method is a function with a receiver: `func (c Certificate) Expired() bool`
belongs to `Certificate`, and you call it as `cert.Expired()`. Inside the
method the receiver `c` works like a parameter, and it can call other methods
on itself, like `c.Expired()`. These are value receivers, so each method reads
its own copy of the certificate.

### Check the edges

Boundary bugs come from picking the wrong comparison. When a rule says "30
days or fewer", the edge value 30 must be included:

| Rule | Comparison | Is 30 included? |
|---|---|---|
| fewer than 30 | `days < 30` | no |
| 30 or fewer | `days <= 30` | yes |
| more than 30 | `days > 30` | no |

When you fix a boundary, check the value just below, on, and just above the
edge.

## Your task

`Certificate` holds a domain and `DaysLeft`, the number of whole days until
the certificate expires. Fix and finish its methods so the alerting job can
trust them.

### 1. Fix `Expired`

A certificate is expired when `DaysLeft` is **0 or less**. On its last day
(`DaysLeft` is 0) it is already treated as expired. `Expired` gets the 0 case
wrong.

```go
Certificate{Domain: "api.example.com", DaysLeft: 0}.Expired()
// => true

Certificate{Domain: "api.example.com", DaysLeft: 1}.Expired()
// => false
```

### 2. Fix `NeedsRenewal`

A certificate needs renewal when `DaysLeft` is `RenewWithinDays` (30) **or
fewer**. That includes expired certificates: they need renewal most of all.
`NeedsRenewal` currently says expired certificates are fine.

```go
Certificate{Domain: "api.example.com", DaysLeft: 30}.NeedsRenewal()
// => true

Certificate{Domain: "api.example.com", DaysLeft: -4}.NeedsRenewal()
// => true

Certificate{Domain: "api.example.com", DaysLeft: 31}.NeedsRenewal()
// => false
```

### 3. Finish `Summary`

`Summary` returns one line for the alert dashboard. The expired case is
already written. Add the other two, using `NeedsRenewal`. Check them in this
order:

| Situation | Summary |
|---|---|
| Expired | `api.example.com: EXPIRED` |
| Needs renewal | `api.example.com: renew soon (12d left)` |
| Anything else | `api.example.com: ok (85d left)` |

The number is `DaysLeft`, followed directly by `d left`. You'll need
`fmt.Sprintf`, so add `"fmt"` to the imports.

```go
Certificate{Domain: "shop.example.com", DaysLeft: 12}.Summary()
// => "shop.example.com: renew soon (12d left)"

Certificate{Domain: "shop.example.com", DaysLeft: 85}.Summary()
// => "shop.example.com: ok (85d left)"
```

## Learn more

- [A Tour of Go: Methods](https://go.dev/tour/methods/1)
- [Go by Example: Methods](https://gobyexample.com/methods)
- [Let's Encrypt: Why ninety-day lifetimes for certificates?](https://letsencrypt.org/2015/11/09/why-90-days/)
