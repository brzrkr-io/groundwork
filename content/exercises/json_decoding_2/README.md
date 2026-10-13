# Decoding JSON: fix the kubectl pod reader

A teammate wrote a tool that reads `kubectl get pods -o json`, but it reports
nothing useful. Fix three classic decoding bugs, then read a stream of JSON
objects one at a time.

## Learn

### Recap

`json.Unmarshal(data, &v)` fills `v` from JSON. Struct tags like
`json:"name"` say which key fills which field. Keys that match nothing are
ignored, and fields whose key is missing keep their zero value. A
`json.NewDecoder(r)` reads JSON from an `io.Reader` with `dec.Decode(&v)`.

### Three bugs that fail quietly

| Bug | What happens |
|---|---|
| `json.Unmarshal(data, v)` without `&` | error: `json: Unmarshal(non-pointer pods.PodList)` |
| lowercase field `namespace string` | never filled, stays `""`, **no error** |
| tag `json:"state"` but the key is `"phase"` | never filled, stays `""`, **no error** |

The last two are sneaky: decoding "succeeds" and you just get empty strings.
The json package can only set exported (capitalised) fields, and it only
matches a tag to the exact key name (ignoring case). `go vet` warns about a
json tag on an unexported field, which is one more reason to run it.

### Streams of JSON values

Some tools print many JSON objects back to back instead of one big array.
`kubectl get pods -o json --watch` does this, and so do many log shippers
(*JSON Lines*: one object per line). A `json.Decoder` handles this naturally:
each `Decode` call reads exactly **one** value and leaves the rest for the
next call. Whitespace and newlines between values don't matter.

When the input runs out, `Decode` returns `io.EOF`. That's the normal way to
finish, not a failure:

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
	// use ev
}
```

Declare `ev` **inside** the loop so each value starts from zero. Otherwise a
field missing in one object would keep its value from the previous one.

If the input stops in the middle of an object, you get `io.ErrUnexpectedEOF`
instead. That is a different error, so it is correctly treated as a failure.

## Your task

Read the starter top to bottom first. The `TODO`s mark every bug.

### 1. Fix DecodePods

`DecodePods(data)` decodes a pod list like this and returns its `Items`:

```json
{"kind": "List", "items": [{"metadata": {"name": "api-7d9f"}}, {"metadata": {"name": "worker-1"}}]}
```

Right now it fails on every valid input. Find the bug and fix it. Errors keep
the existing `decode pods: ` prefix.

```go
pods, err := DecodePods(data) // pods[0].Metadata.Name == "api-7d9f", err == nil
```

### 2. Fix the Pod struct

Two fields are never filled in:

- `Metadata.namespace` is unexported. Rename it to `Namespace` and update
  `Ref` to use the new name.
- `Status.Phase` has the wrong tag. The JSON key is `"phase"`.

```go
// {"metadata":{"name":"api-7d9f","namespace":"shop"},"status":{"phase":"Running"}}
p.Ref()         // => "shop/api-7d9f"
p.Status.Phase  // => "Running"
```

### 3. Count phases in a stream

`CountPhases(r)` reads pod objects one after another from `r` (one per line,
or spread over several lines) and counts how many are in each phase. Do task 2
first, otherwise every phase is `""`.

- Use a `json.NewDecoder` and loop until `Decode` returns `io.EOF`. Then
  return the counts and `nil`. An empty stream gives an empty map.
- If any other error happens, return `nil` and the error wrapped as
  `pod N: %w`, where `N` counts the objects from 1 (the one that failed).
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

## Learn more

- [json.Decoder example: streams](https://pkg.go.dev/encoding/json#example-Decoder)
- [JSON Lines format](https://jsonlines.org/)
- [kubectl output formats](https://kubernetes.io/docs/reference/kubectl/#output-options)
