# Practice: incident board

When things break, on-call needs one shared list: what's wrong, how bad, and
who's on it. You'll build that board from your own types, with no new syntax:
just the pieces from this module working together.

**New here:** nothing; practice of *Types and methods*.

## Learn

### How the board fits together

```text
Board ──holds──▶ []Incident ──each──▶ Incident{ID, Title, Severity, State, Owner}
  │                                        │
  ├─ Open(title, sev) → new ID             ├─ String()   one line for chat
  ├─ Find(id)  → *Incident or nil          ├─ Ack(owner) open → acknowledged
  └─ Worst()   → *Incident or nil          └─ Resolve()  → resolved
```

- `Severity` is an `int` enum (`Minor`, `Major`, `Critical`); a bigger value
  is worse.
- `State` is a `string` type with three constants: `StateOpen` (`"open"`),
  `StateAcked` (`"acknowledged"`) and `StateResolved` (`"resolved"`).
- The zero `Board` is ready to use: `var b Board` gives an empty board.

### File map

| File | What's in it | Do you edit it? |
|---|---|---|
| `exercise.go` | The types and every method signature, with `TODO(n)` markers | Yes: the method bodies and an `import` block |
| `exercise_test.go` | The tests | No, read it if you're curious |

### Pieces you'll reuse

| Piece | From | Here |
|---|---|---|
| Struct | *Structs* | `Incident` and `Board` |
| `iota` and a `string`-based type | *Custom types and enums* | `Severity` and `State` |
| `String()` method | *Stringers* | `Severity` and `Incident` print themselves |
| Stringers inside Stringers | *Stringers: a kubectl-style deployment report* | `Incident.String` prints its `Severity` with `%v` |
| Pointer receiver | *Pointer receivers* | `Ack`, `Resolve` and the `Board` methods change the real value |
| `&s[i]`, the address of a slice element | *Pointers: fix the silent no-ops* | `Find` and `Worst` point into the board |
| `nil` for "no value" | *Pointers: optional config fields* | `Find` and `Worst` when nothing matches |

> **Why is it written like this?** Value receivers on the `String` methods
> are convention: values and pointers both print. Pointer receivers on the
> rest are required, or the changes land on a copy. You'll need `"fmt"`
> for tasks 1 and 2; Go rejects unused imports, so add it when you first
> use it.

### Pointers into a slice

`for _, inc := range b.incidents` gives you a **copy** of each incident, so
`&inc` points at the copy, and changes through it never reach the board.
Loop with the index instead and take `&b.incidents[i]`.

Use the pointer straight away. The next `Open` may make `append` move the
slice to new memory (as in *Slice gotchas: shared memory*), so call `Find`
again after opening.

## Your task

Your team tracks incidents in a chat thread and loses them every week. Fill
in the method bodies in `exercise.go`. The test for task 5 fills the board
directly, so any order works, but 1 to 5 is easiest.

### 1. Name the severities

**Goal:** make `Severity.String` return the severity as a lower-case word.

| Value | `String()` |
|---|---|
| `Minor` | `"minor"` |
| `Major` | `"major"` |
| `Critical` | `"critical"` |
| anything else | `"Severity(N)"`, with the number in place of `N` |

```go
fmt.Println(Critical)    // critical
fmt.Println(Severity(7)) // Severity(7)
```

**Think it through:**
- A `switch` and a lookup slice both work (*Stringers*,
  *Stringers: pod phases and uptime*). Which one handles "anything else"
  more easily?
- In the fallback, what must you pass so `String` isn't called again?

### 2. Print an incident

**Goal:** give `Incident.String` the format
`#<ID> [<Severity>] <Title> (<State>)`, adding ` by <Owner>` after the
state when `Owner` isn't empty.

- `ID`, `Severity`, `Title`, `State` and `Owner` come from the fields.
- Fixed text: `#` before the ID, `[` and `]` around the severity, single
  spaces between the parts, `(` and `)` around the state, and ` by ` before
  the owner.
- The severity prints as its word from task 1, so finish that first.

```go
Incident{ID: 12, Title: "api 5xx spike", Severity: Critical, State: StateOpen}
// => "#12 [critical] api 5xx spike (open)"

Incident{ID: 7, Title: "disk full on db-1", Severity: Major, State: StateAcked, Owner: "ana"}
// => "#7 [major] disk full on db-1 (acknowledged by ana)"

Incident{ID: 3, Title: "slow dashboard", Severity: Minor, State: StateResolved, Owner: "sam"}
// => "#3 [minor] slow dashboard (resolved by sam)"
```

The tests print incidents with `fmt.Sprint`, so they pass only once `fmt`
finds your method.

**Think it through:**
- Build the part inside the parentheses first, then put the whole line
  together with one `fmt.Sprintf`.
- Which verb makes `i.Severity` print itself?

### 3. Acknowledge and resolve

**Goal:** make `Ack(owner)` and `Resolve()` move an incident along, and
report whether they changed it.

`Ack(owner string) bool`:

- An incident in `StateOpen` → set `State` to `StateAcked`, set `Owner` to
  `owner`, return `true`.
- Any other state → change nothing, return `false`. A second `Ack` keeps
  the first owner.

`Resolve() bool`:

- Already `StateResolved` → change nothing, return `false`.
- Otherwise → set `State` to `StateResolved`, return `true`. `Owner` stays
  as it is.

```go
inc := Incident{ID: 1, Title: "api down", State: StateOpen}
inc.Ack("ana")   // => true; inc.State is StateAcked, inc.Owner is "ana"
inc.Ack("sam")   // => false; inc.Owner is still "ana"
inc.Resolve()    // => true; inc.State is StateResolved
inc.Resolve()    // => false
```

**Think it through:**
- Check the state first, and return early when there's nothing to do.
- The receivers are already `*Incident`. Do you need `*` to set a field?

### 4. Open and find incidents

**Goal:** `Open` adds a new incident and returns its ID; `Find` returns a
pointer to the incident with that ID, or `nil`.

`Open(title string, sev Severity) int`:

- IDs start at `1` and go up by one with each `Open`. `lastID` holds the ID
  given most recently (`0` on an empty board).
- The new incident has the next ID, the `title` and `sev` passed in, and
  `State` `StateOpen`. Append it to `b.incidents` and return its ID.

`Find(id int) *Incident`:

- Return the address of the matching incident **inside** `b.incidents`, so
  changes through the pointer stick. The test changes a field through your
  pointer and checks the board.
- No match → return `nil`.

```go
var b Board
b.Open("api 5xx spike", Critical)   // => 1
b.Open("disk full on db-1", Major)  // => 2
*b.Find(2) // => Incident{ID: 2, Title: "disk full on db-1", Severity: Major, State: StateOpen}
b.Find(99) // => nil

b.Find(1).Ack("ana") // changes incident #1 on the board itself
```

**Think it through:**
- In `Open`, which field do you change before you build the incident?
- In `Find`, which loop form lets you write `&b.incidents[i]`? (See
  *Pointers into a slice*.)

### 5. Pick the worst open incident

**Goal:** `Worst()` returns a pointer to the unresolved incident with the
highest `Severity`, or `nil` if there is none.

- Skip incidents in `StateResolved`. Open and acknowledged ones both count.
- On a tie, return the one that comes first in `b.incidents` (it was opened
  first).
- Like `Find`, point into `b.incidents`, not at a copy.
- No unresolved incidents, or an empty board → `nil`.

```go
// #1 minor open, #2 critical acknowledged, #3 major open
b.Worst().ID // => 2

// #4 minor open, #5 major open, #6 major open
b.Worst().ID // => 5

// #1 critical resolved, #2 major resolved
b.Worst() // => nil
```

**Think it through:**
- Keep a `*Incident` for "the worst so far", starting at `nil`. When does a
  new incident replace it?
- `Severity` is an `int` underneath, so `>` compares two severities. Should
  a tie replace the current worst?

## Learn more

- [A Tour of Go: Pointer receivers](https://go.dev/tour/methods/4)
- [Effective Go: Pointers vs. values](https://go.dev/doc/effective_go#pointers_vs_values)
- [Go by Example: Enums](https://gobyexample.com/enums)
- [Go FAQ: Should I define methods on values or pointers?](https://go.dev/doc/faq#methods_on_values_or_pointers)
