# Stringers: a kubectl-style deployment report

`kubectl get deployments` turns a pile of structs into a few readable lines.
Here you build the same kind of output for your own rollout tool: small
`String` methods that combine into one report.

**New here:** Stringers inside Stringers, `fmt.Fprintf` and `fmt.Fprintln`. (`strings.Builder` is from *For loops: backoff and progress bars*.)

## Learn

### Recap

A type with a `String() string` method is a `fmt.Stringer`. `fmt` calls it
for `%v`, `%s`, `Print`, `Println` and `Sprint`, as in *Stringers*.

Never format the receiver itself with `%v` inside its own `String`: that
calls `String` again, forever.

### Stringers inside Stringers

If a struct's *fields* have their own `String` methods, the struct's `String`
can print those fields with `%v`. Each field formats itself:

```go
type Version struct{ Major, Minor int }

func (v Version) String() string {
	return fmt.Sprintf("v%d.%d", v.Major, v.Minor)
}

type Release struct {
	App     string
	Version Version
}

func (r Release) String() string {
	return fmt.Sprintf("%s@%v", r.App, r.Version) // r.Version prints itself
}

fmt.Sprint(Release{App: "billing", Version: Version{2, 4}})
// => "billing@v2.4"
```

`%v` on a field calls the *field's* `String`. Only `%v` on `r` itself would loop.

> **Why is it written like this?**
> - **Go requires** the exact signature `String() string` for `fmt` to find it.
> - **Convention:** a value receiver (`(v Version)`), so values and pointers both print nicely.
> - **Author's choice:** the receiver name (`v`, `r`) and the text layout.

### Building text with strings.Builder

A `strings.Builder` collects text piece by piece. `b.String()` returns
everything you wrote.

`fmt.Fprintf` and `fmt.Fprintln` work like `Sprintf` and `Println`, but they
*write into* the thing you pass first instead of returning a string:

```go
var b strings.Builder
fmt.Fprintf(&b, "hosts: %d\n", 2)
fmt.Fprintln(&b, "web-1") // Fprintln adds the "\n" for you
fmt.Fprintln(&b, "web-2")
b.String()
// => "hosts: 2\nweb-1\nweb-2\n"
```

`Fprintln` calls `String` on your values, just like `Println`.

> **Why is it written like this?**
> - **Go requires** `&b`, not `b`: the builder must be changed in place, so you pass a pointer. (`Fprintf` writes into anything with the right `Write` method, and `*strings.Builder` has one; *Interfaces* shows how that works.)
> - **Go requires** `"strings"` in the imports to use `strings.Builder`.
> - **Convention:** `var b strings.Builder`. The zero value is ready to use; no constructor needed.
> - **Author's choice:** the name `b`.

## Your task

`Status` already prints itself as `Progressing`, `Available` or `Degraded`.
Add the rest of the output for your rollout tool.

### 1. Print replica counts

**Goal:** add a `String` method to `Replicas` that returns `Ready`, a slash, then `Desired`.

```go
fmt.Sprint(Replicas{Ready: 3, Desired: 5})
// => "3/5"
```

**Think it through:**
- Where does a method go? Look for the `TODO(1)` comment under the `Replicas` type.
- Both numbers are `int`. Which `Sprintf` verb prints an `int`?
- The `/` is fixed text; the numbers come from the receiver's fields.

### 2. Print a deployment

**Goal:** add a `String` method to `Deployment` that returns `namespace/name replicas status`.

Namespace, a slash, the name, a space, the replicas, a space, the status:

```go
d := Deployment{
	Name:      "api",
	Namespace: "prod",
	Replicas:  Replicas{Ready: 1, Desired: 3},
	Status:    Degraded,
}
fmt.Sprint(d)
// => "prod/api 1/3 Degraded"
```

**Think it through:**
- Which two fields already have their own `String` method? Let them print themselves.
- Namespace comes *before* name, even though `Name` is declared first.
- Fixed text: one `/` and two single spaces.

### 3. Build the report

**Goal:** finish `Report` so it returns a summary line, then one line per deployment, each ending in `"\n"`.

```text
deployments: 3, degraded: 1
prod/api 3/3 Available
prod/web 1/2 Progressing
staging/worker 0/1 Degraded
```

- Summary: `deployments: N, degraded: M`. `N` is how many deployments there
  are; `M` is how many have status `Degraded`.
- Each deployment line is exactly what its `String` method returns, in the
  order given.
- With no deployments, return only the summary:
  `"deployments: 0, degraded: 0\n"`.

**Think it through:**
- The summary needs the degraded count *before* you write any lines. Count first, then write.
- Which `fmt` function writes a value plus a newline into a builder?
- Remember to add `"strings"` to the imports.

## Learn more

- [fmt.Stringer](https://pkg.go.dev/fmt#Stringer)
- [strings.Builder](https://pkg.go.dev/strings#Builder)
- [fmt.Fprintf](https://pkg.go.dev/fmt#Fprintf)
- [Kubernetes: Deployment status](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#deployment-status)
