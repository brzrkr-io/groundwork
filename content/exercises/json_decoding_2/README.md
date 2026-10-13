# Decoding JSON: fix the kubectl pod reader

A teammate wrote a tool that reads `kubectl get pods -o json`, but it reports
nothing useful. Fix three classic decoding bugs, then read a stream of JSON
objects one at a time.

**New here:** decoding a stream of values, `io.EOF` as the normal end.

## Learn

### Recap

As in *Decoding JSON: load a service config*: `json.Unmarshal(data, &v)`
fills `v`. Tags like `json:"name"` say which key fills which field. Unmatched
keys are ignored; fields with a missing key keep their zero value.
`json.NewDecoder(r)` reads from an `io.Reader` with `dec.Decode(&v)`.

### Three bugs that fail quietly

| Bug | What happens |
|---|---|
| `json.Unmarshal(data, v)` without `&` | error: `json: Unmarshal(non-pointer …)` |
| lowercase field `region string` | never filled, stays `""`, **no error** |
| tag `json:"host"` but the key is `"hostname"` | never filled, stays `""`, **no error** |

The last two are sneaky: decoding "succeeds" and you just get empty strings.
`go vet` warns about a json tag on an unexported field, one more reason to
run it.

> **Why is it written like this?** Go requires a capital first letter for a
> field to be visible to another package like `encoding/json`. When you
> rename a field, fix the declaration first; then the compiler points at
> every line still using the old name (see *Reading compiler errors*).

### Streams of JSON values

Some tools print many JSON objects back to back instead of one array:
`kubectl get pods -o json --watch` does, and so do log shippers using *JSON
Lines*. Each `Decode` call reads exactly **one** value and leaves the rest
for the next call. Whitespace and newlines between values don't matter.

When the input runs out, `Decode` returns `io.EOF`. That's the normal finish,
not a failure:

```go
dec := json.NewDecoder(r)
for {
	var ev Event
	err := dec.Decode(&ev)
	if errors.Is(err, io.EOF) {
		break // no more values
	}
	if err != nil {
		return err // real problem: broken JSON
	}
	fmt.Println(ev.Kind)
}
```

> **Why is it written like this?** `for { … }` with no condition is Go's
> loop-forever; you leave with `break` or `return`. Checking `io.EOF`
> *before* the general `err != nil` is required by the logic: otherwise the
> normal end would count as an error. Declaring `ev` **inside** the loop is
> deliberate: each value starts from zero, so a field missing in one object
> can't keep the previous object's value.

If the input stops in the middle of an object, you get `io.ErrUnexpectedEOF`,
a different error, so it is treated as a failure.

## Your task

Read the starter top to bottom first. The `TODO`s mark every bug.

### 1. Fix DecodePods

**Goal:** make `DecodePods(data)` decode a pod list and return its `Items`.

```json
{"kind": "List", "items": [{"metadata": {"name": "api-7d9f"}}, {"metadata": {"name": "worker-1"}}]}
```

Right now it fails on every valid input. Errors keep the existing
`decode pods: ` prefix.

```go
pods, err := DecodePods(data) // pods[0].Metadata.Name == "api-7d9f", err == nil
```

**Think it through:**
- Compare the `Unmarshal` call with the first row of the bug table.

### 2. Fix the Pod struct

**Goal:** make `Ref()` return `"namespace/name"` and `Status.Phase` hold the
pod's phase.

Two fields are never filled in:

- `Metadata.namespace` is unexported. Export it (name it `Namespace`), then
  update `Ref` to use the new name.
- `Status.Phase` has the wrong tag. The JSON key is `"phase"`.

```go
// {"metadata":{"name":"api-7d9f","namespace":"shop"},"status":{"phase":"Running"}}
p.Ref()        // => "shop/api-7d9f"
p.Status.Phase // => "Running"
```

**Think it through:**
- After renaming the field, which method still uses the old name?
- Compare the tag text with the key in the JSON example.

### 3. Count phases in a stream

**Goal:** `CountPhases(r)` reads pod objects one after another from `r` and
returns how many are in each phase.

Do task 2 first, otherwise every phase is `""`.

- Pods may be one per line or spread over several lines.
- Stop when `Decode` returns `io.EOF`, then return the counts and `nil`. An
  empty stream gives an empty map.
- On any other error, return `nil` and the error wrapped as `pod N: ` +
  the original error, where `N` is the position (counting from 1) of the
  object that failed.
- Add `"errors"` to the imports for `errors.Is`.

```go
// input:
// {"status":{"phase":"Running"}}
// {"status":{"phase":"Running"}}
// {"status":{"phase":"Failed"}}
CountPhases(r) // => map[Failed:1 Running:2], nil

// input:
// {"status":{"phase":"Running"}}
// {"status":{"phase":"Running"}}
// {"status":{"phase": Running}}
CountPhases(r) // => nil, error "pod 3: invalid character 'R' looking for beginning of value"
```

**Think it through:**
- Where does `N` come from? Something has to count loop turns, starting at 1.
- Which field of each decoded pod is the map key?

## Learn more

- [json.Decoder example: streams](https://pkg.go.dev/encoding/json#example-Decoder)
- [JSON Lines format](https://jsonlines.org/)
- [kubectl output formats](https://kubernetes.io/docs/reference/kubectl/#output-options)
