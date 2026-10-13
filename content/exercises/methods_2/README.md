# Methods: certificate checks

Expired TLS certificates are a classic cause of outages. A teammate wrote
methods that check certificates, and two have bugs at the boundaries, exactly
where real bugs like to hide.

**New here:** nothing new; practice methods, boundary comparisons and `fmt.Sprintf`.

## Learn

### Recap

As in *Methods*, a method is a function with a receiver:

```go
func (b Backup) Stale() bool {
	return b.AgeHours > 24
}
```

- `(b Backup)` says the method belongs to `Backup`; inside, it's called `b`.
- Call it as `nightly.Stale()`. Inside another `Backup` method, call it as `b.Stale()`.
- These are value receivers: each method reads its own copy.

> **Why is it written like this?** A method that returns a `bool` can return
> a comparison directly; `b.AgeHours > 24` already *is* `true` or `false`.
> Go doesn't require that, but it's convention over `if … { return true }`.

### Check the edges

Boundary bugs come from picking the wrong comparison. When a rule says "30
or fewer", the edge value 30 must be included:

| Rule | Comparison | Is 30 included? |
|---|---|---|
| fewer than 30 | `days < 30` | no |
| 30 or fewer | `days <= 30` | yes |
| more than 30 | `days > 30` | no |
| 30 or more | `days >= 30` | yes |

When you fix a boundary, check the value just below, on, and just above it.

### Putting a number straight before a letter

In `fmt.Sprintf`, `%d` is replaced by an int. Text right after it is
printed as-is, so `%dh` gives a number followed by `h`:

```go
fmt.Sprintf("%s is %dh old", "nightly", 30)
// => "nightly is 30h old"
```

## Your task

`Certificate` holds a `Domain` and `DaysLeft`, the number of whole days until
the certificate expires. Fix and finish its methods so the alerting job can
trust them.

### 1. Fix `Expired`

**Goal:** make `Expired()` return `true` when `DaysLeft` is **0 or less**.

On its last day (`DaysLeft` is 0) a certificate already counts as expired.
The current comparison gets the 0 case wrong.

```go
Certificate{Domain: "api.example.com", DaysLeft: 0}.Expired()
// => true

Certificate{Domain: "api.example.com", DaysLeft: 1}.Expired()
// => false
```

**Think it through:**
- Does the current operator include 0? Check the table above.
- You only need to change the operator, not the whole line.

### 2. Fix `NeedsRenewal`

**Goal:** make `NeedsRenewal()` return `true` when `DaysLeft` is `RenewWithinDays` (30) **or fewer**, including expired certificates.

Expired certificates (0 or negative days) need renewal most of all, but the
current code says they're fine. 31 days or more → `false`.

```go
Certificate{Domain: "api.example.com", DaysLeft: 30}.NeedsRenewal()
// => true

Certificate{Domain: "api.example.com", DaysLeft: -4}.NeedsRenewal()
// => true

Certificate{Domain: "api.example.com", DaysLeft: 31}.NeedsRenewal()
// => false
```

**Think it through:**
- The condition has two parts joined by `&&`. Try `DaysLeft: -4` in each part: which one says `false`?
- Does "30 or fewer" on its own already include negative numbers?

### 3. Finish `Summary`

**Goal:** make `Summary()` return the right dashboard line for all three situations.

The expired case is already written. Add the other two, checked in this order:

| Situation | Summary |
|---|---|
| `Expired()` is true | `api.example.com: EXPIRED` |
| `NeedsRenewal()` is true | `api.example.com: renew soon (12d left)` |
| anything else | `api.example.com: ok (85d left)` |

The pieces: the `Domain` field, the fixed text (`: renew soon (` or
`: ok (`), the `DaysLeft` field, then the fixed text `d left)`. No space
between the number and `d`. You'll need `fmt.Sprintf`, so add
`import "fmt"` under the `package` line.

```go
Certificate{Domain: "shop.example.com", DaysLeft: 12}.Summary()
// => "shop.example.com: renew soon (12d left)"

Certificate{Domain: "shop.example.com", DaysLeft: 85}.Summary()
// => "shop.example.com: ok (85d left)"
```

**Think it through:**
- The existing `if c.Expired()` block is your pattern. Which method does the next `if` call?
- The last `return` line is the "ok" case. What's missing from it compared with the table?

## Learn more

- [A Tour of Go: Methods](https://go.dev/tour/methods/1)
- [Go by Example: Methods](https://gobyexample.com/methods)
- [Let's Encrypt: Why ninety-day lifetimes for certificates?](https://letsencrypt.org/2015/11/09/why-90-days/)
