# Extract function

Long functions are where bugs hide: you can't hold 50 lines in your head,
so you skim, and skimming misses things. *Extract function* is the most
used refactoring there is: move a block of code into its own function with
a name that says what it does. This exercise breaks a long deploy planner
into steps you can read at a glance.

**New here:** refactoring, extract function. (`append(a, b...)` is from
*append and copy*; `if err := f(); err != nil` from *if and else*.)

## Learn

### Refactoring

*Refactoring* means changing how code is structured **without changing
what it does**. The tests are your safety net: they pass before, and they
must still pass after every small step. If a test goes red, undo the last
step instead of debugging a big change.

### Comments that name a block

Look for a comment that names the next few lines:

```go
// Parse the timestamp.
ts, err := time.Parse(time.RFC3339, fields[0])
if err != nil {
	return Event{}, err
}
```

That comment is a function waiting to be born. Move the lines into a
function, and let its name replace the comment:

```go
ts, err := parseTimestamp(fields[0])
```

### The steps, every time

1. Create the new function, empty, below the old one.
2. **Copy** (don't cut yet) the block into it.
3. Work out its inputs: every variable the block reads but doesn't create
   becomes a parameter.
4. Work out its output: what the rest of the old function needs from the
   block becomes the return value.
5. Replace the block in the old function with a call. Run the tests.

Many editors can do this for you: select the lines, then **Extract
function** in the code actions menu (the light bulb in VS Code). For
task 1, do it by hand: that block has `return`s in it, and the editor's
version turns them into extra flag variables that are harder to read.

### Returning an error from a helper

Inside the old function, a failed check did `return nil, err` to leave
`Plan` immediately. Inside a helper, `return err` only leaves the helper,
so the caller must pass the error on:

```go
if err := checkDisk(host); err != nil {
	return nil, err
}
```

This `if` creates `err`, runs the call, and checks the result in one line.
`err` exists only inside the `if`. When a helper finds no problem, it
returns `nil`.

> **Why is it written like this?** **Go requires** a function that returns
> `error` to return something on every path, so a check helper ends with
> `return nil`. The short `if err := ...; err != nil` form is
> **convention** when you only need `err` inside the `if`.

### Adding one slice to another

`append(steps, more...)` adds every element of the slice `more` to
`steps`. The `...` spreads the slice out into separate arguments, as in
*Variadic functions*:

```go
lines := []string{"start"}
lines = append(lines, []string{"a", "b"}...)
// lines is ["start", "a", "b"]
```

### How the tests check the shape

`structure_test.go` reads `exercise.go` with Go's `go/ast` package. It
checks that each helper exists, that `Plan` calls it, and finally that
`Plan` has at most 12 lines of code (blank lines and comments don't
count). The behaviour checks run first: if a refactor changes the output,
the test says so before it looks at the shape.

## Your task

`Plan` turns a deploy request into the steps of a rolling deploy. It works,
but it's almost 50 lines long and does four jobs. For a request with `Service`
`"api"`, no registry, `Tag` `"v2"`, 5 replicas and a max surge of 2:

```go
Plan(Request{Service: "api", Tag: "v2", Replicas: 5, MaxSurge: 2})
// => "deploy docker.io/api:v2 to 5 replicas"
//    "batch 1: replace 2 pods (2/5)"
//    "batch 2: replace 2 pods (4/5)"
//    "batch 3: replace 1 pod (5/5)"
//    "verify api"
```

Every task keeps this output exactly the same. Extract the three blocks the
comments name, then `Plan` becomes a short list of calls.

### 1. Extract validate

**Goal:** move the five checks into `validate(r Request) error`, and call it
from `Plan`.

`validate` returns the first problem it finds (the same messages as now),
or `nil`. `Plan` returns `nil` and that error when `validate` fails:

```go
Plan(Request{Service: "api", Tag: "latest", Replicas: 1, MaxSurge: 1})
// => nil, error `tag "latest" is not allowed; pin a version`
```

`Plan` itself must no longer call `errors.New` or `fmt.Errorf`.

**Think it through:**
- Inside `validate`, what replaces each `return nil, errors.New(...)`?
- What does `validate` return at the end, when every check passed?

### 2. Extract imageRef

**Goal:** move the registry and image lines into `imageRef(r Request) string`,
and call it from `Plan`.

`imageRef` returns the registry, `/`, the service, `:`, then the tag, with
`"docker.io"` when `Registry` is empty:

```go
imageRef(Request{Service: "web", Registry: "registry.example.com", Tag: "2026.10"})
// => "registry.example.com/web:2026.10"
```

`Plan` must no longer contain the text `"docker.io"`.

**Think it through:**
- Which variable does the rest of `Plan` need from this block?
- Do you still need the variable `image` in `Plan`, or can the call go straight into `fmt.Sprintf`?

### 3. Extract batchSteps

**Goal:** move the batch loop into `batchSteps(r Request) []string`, call it
from `Plan`, and get `Plan` down to 12 lines of code or fewer.

`batchSteps` returns only the batch lines; `Plan` adds the first and last
steps around them. `Plan` must have no `for` loop left.

**Think it through:**
- The loop appends to `steps`. In the new function, which slice does it append to, and where does that slice start?
- How do you add every line `batchSteps` returns to `steps` in one `append`?

## Learn more

- [Refactoring catalog: Extract Function (Martin Fowler)](https://refactoring.com/catalog/extractFunction.html)
- [Go Code Review Comments: Indent error flow](https://go.dev/wiki/CodeReviewComments#indent-error-flow)
- [Go spec: Passing arguments to ... parameters](https://go.dev/ref/spec#Passing_arguments_to_..._parameters)
