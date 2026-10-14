# append and copy: rolling deploy plan

A rolling deploy never touches every host at once. It deploys the canaries
first, then the rest in small batches, so a bad release only hurts a few hosts.

**New here:** nothing; practice with `append` and `copy`. (Slices of slices are from *Slices: rolling-deploy batches*, stepping a loop from *For loops: plan a rollout*.)

## Learn

### Recap

- Grow a result in a loop: start with `var out []T`, then `out = append(out, v)`.
- `slices.Clone(s)` (with `import "slices"`) makes a copy that does not change
  when `s` changes.
- `s[a:b]` cuts a piece of a slice, as in *Slices*. That piece **shares**
  elements with `s`.

> **Why is it written like this?** `out = append(out, v)` is required: `append`
> returns the new slice, and Go rejects an `append` whose result you ignore.
> The name `out` is your choice.

### A slice of slices

A slice's elements can themselves be slices. `[][]string` is a slice whose
elements are `[]string` values:

```go
groups := [][]string{
	{"eu-1", "eu-2"}, // groups[0]
	{"us-1"},         // groups[1]
}
len(groups)  // 2
groups[0][1] // "eu-2": first pick the group, then the name in it

var all [][]string
all = append(all, []string{"ap-1", "ap-2"}) // adds one whole group
```

> **Why is it written like this?** Go requires the type to say both levels:
> `[][]string` reads "slice of (slice of string)". Inside the `{ }` you may
> drop the inner `[]string`, because Go already knows it.

### A loop that jumps

A three-part `for` loop can step by any amount, not only 1:

```go
for n := 0; n < 10; n += 4 {
	fmt.Println(n) // 0, 4, 8
}
```

> **Why is it written like this?** The three parts (start; keep going while;
> step) and the semicolons are required syntax. The variable name is your
> choice.

### Don't run off the end

Slicing past `len(s)` crashes the program. The built-in `min`, as in *For
loops: plan a rollout*, picks the smaller number:

```go
min(7, 5) // 5
```

## Your task

Your deploy tool reads two host lists from the inventory: canaries and regular
hosts. The inventory export sometimes contains blank names (`""`).

### 1. Build the deploy plan

**Goal:** make `Plan(canaries, hosts)` return all canaries, then all hosts, with
blank names left out.

- Keep the order: canaries in their order first, then hosts in their order.
- Skip every `""`, in either list.
- If nothing is left, return a slice with no elements (`nil` is fine).

```go
Plan([]string{"web-0"}, []string{"web-1", "", "web-2"})
// => ["web-0" "web-1" "web-2"]

Plan([]string{"", ""}, nil)
// => []
```

**Think it through:**
- Which list do you walk first, so canaries end up at the front?
- What check decides whether a name gets appended?

### 2. Split it into batches

**Goal:** make `Batches(hosts, size)` split `hosts` into groups of `size`
hosts, in order.

- The last batch holds whatever is left, so it may be smaller.
- If `hosts` is empty, return no batches (`nil`).
- If `size` is less than 1, return no batches (`nil`).
- Each batch must be its own copy. The test changes `hosts` after the call and
  checks that your batches did not change.

```go
Batches([]string{"web-1", "web-2", "web-3", "web-4", "web-5"}, 2)
// => [["web-1" "web-2"] ["web-3" "web-4"] ["web-5"]]

Batches([]string{"web-1", "web-2"}, 5)
// => [["web-1" "web-2"]]

Batches([]string{"web-1"}, 0)
// => []
```

**Think it through:**
- Where does each batch start, and by how much does that start move each time?
- For the last batch, which is smaller: start + `size`, or `len(hosts)`?
- A piece cut with `[a:b]` shares elements with `hosts`. What makes it a copy?

## Learn more

- [Go by Example: Slices](https://gobyexample.com/slices)
- [Package slices](https://pkg.go.dev/slices)
- [Kubernetes: Performing a Rolling Update](https://kubernetes.io/docs/tutorials/kubernetes-basics/update/update-intro/)
