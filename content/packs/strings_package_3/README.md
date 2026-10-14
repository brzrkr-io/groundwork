# The strings package: which pods are broken?

`kubectl get pods` prints a table. Chat bots and scripts read that text to
answer one question: which pods need attention? Here you write that check.

**New here:** nothing; practice with `strings` and `slices`.

## Learn

### Recap

| Tool | Does |
|---|---|
| `strings.SplitSeq(text, "\n")` | loop over lines |
| `strings.Fields(line)` | split into columns, any amount of spacing |
| `strings.Cut(s, "/")` | before, after, found |
| `strings.Join(names, ", ")` | glue a slice into one string |
| `slices.Sort(s)` | sort in place, as in *Sorting* |

### Picking columns by position

`Fields` returns a slice; index it from `0`. Check the length first, because
indexing past the end panics.

```go
cols := strings.Fields("eth0  up  1500")
if len(cols) >= 2 {
	fmt.Println(cols[1]) // "up": the 2nd column is index 1
}
```

> **Why is it written like this?** Go requires indexes to start at `0`. The
> length check is not required by the compiler, but without it a short line
> crashes the program.

### The input

A typical `kubectl get pods` output. It always ends with a newline.

```text
NAME                     READY   STATUS             RESTARTS      AGE
api-7d4b9c6f5-x2kqz      1/1     Running            0             5d
web-5f6b8d7c9-k8s2m      0/1     CrashLoopBackOff   12 (3m ago)   2h
migrate-db-4xk2p         0/1     Completed          0             1d
cache-0                  1/2     Running            0             3h
```

`READY` means "ready containers / total containers"; a pod is ready when the
two numbers are equal. `Completed` means a one-off job that finished, so it's
fine even at `0/1`.

## Your task

Your team's chat bot runs `kubectl get pods` and posts a one-line health
summary. Add `"slices"` and `"strings"` to the imports in `exercise.go`.

### 1. Read the READY column

**Goal:** make `IsReady` return `true` when a value like `"2/3"` has a `"/"`
and the text before it equals the text after it.

No `"/"` at all (like `"1"` or `""`) means `false`.

```go
IsReady("1/1")  // => true
IsReady("3/3")  // => true
IsReady("0/1")  // => false
IsReady("1/2")  // => false
IsReady("")     // => false
```

**Think it through:**
- Which tool gives you the two sides *and* whether `"/"` was there?
- Both conditions must be true. Which operator joins them?

### 2. Find the pods that are not ready

**Goal:** make `NotReadyPods` return the names of pods that need attention,
sorted alphabetically.

For each line of `output`:

- Split it into columns with `strings.Fields`.
- Skip lines with fewer than 3 columns (like the empty piece after the last
  newline).
- Skip the header: its first column is `"NAME"`.
- The pod needs attention when its READY column (2nd) is not ready **and**
  its STATUS column (3rd) is not `"Completed"`.

```go
NotReadyPods(output) // the example above
// => ["cache-0" "web-5f6b8d7c9-k8s2m"]
```

No pods, or all fine: return an empty slice.

**Think it through:**
- The pod name is the 1st column. What index is that?
- You already wrote a function that checks READY. Reuse it.
- Sort once, after the loop.

### 3. Write the summary

**Goal:** make `Report` return a one-line summary built from `NotReadyPods`.

- Nothing needs attention: `"all pods ready"`.
- Otherwise: the fixed text `"not ready: "` (note the space after the
  colon), then the pod names separated by `", "`.

```go
Report(output)
// => "not ready: cache-0, web-5f6b8d7c9-k8s2m"
```

**Think it through:**
- How do you check that a slice is empty?
- Which part is fixed text, and which part comes from the slice?

## Learn more

- [kubectl get](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_get/)
- [Kubernetes: Pod lifecycle](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/)
- [Package strings](https://pkg.go.dev/strings)
