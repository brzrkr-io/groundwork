# The strings package: which pods are broken?

`kubectl get pods` prints a table. Scripts and chat bots often read that text
to answer one question: which pods need attention? In this exercise you write
that check.

## Learn

### Recap

Loop over the lines of some text with
`for line := range strings.SplitSeq(output, "\n")`, and break a line into
columns with `strings.Fields(line)`, which ignores however many spaces line
the columns up. `strings.Cut(s, "/")` splits at the first `"/"`, and
`strings.Join(names, ", ")` glues a slice back into one string. Sort your
result with `slices.Sort`, as in *Sorting*.

### The input

Here is a typical `kubectl get pods` output. It always ends with a newline.

```text
NAME                     READY   STATUS             RESTARTS      AGE
api-7d4b9c6f5-x2kqz      1/1     Running            0             5d
web-5f6b8d7c9-k8s2m      0/1     CrashLoopBackOff   12 (3m ago)   2h
migrate-db-4xk2p         0/1     Completed          0             1d
cache-0                  1/2     Running            0             3h
```

`READY` reads "ready containers / total containers". A pod is ready when the
two numbers are equal. A pod whose `STATUS` is `Completed` was a one-off job
that finished, so it is fine even though it shows `0/1`.

## Your task

Your team's chat bot runs `kubectl get pods` and posts a one-line health
summary. You write the parsing. Add `"slices"` and `"strings"` to the imports
in `exercise.go`.

### 1. Read the READY column

Change `IsReady` to report whether a READY value like `"2/3"` means all
containers are ready: there is a `"/"`, and the text before it equals the
text after it.

```go
IsReady("1/1")  // => true
IsReady("3/3")  // => true
IsReady("0/1")  // => false
IsReady("1/2")  // => false
IsReady("")     // => false
```

### 2. Find the pods that are not ready

Change `NotReadyPods` to take the whole `kubectl get pods` output and return
the names of the pods that need attention, in alphabetical order. For each
line:

- Split it into columns with `strings.Fields`.
- Skip lines with fewer than 3 columns (such as the empty string after the
  last newline).
- Skip the header line: its first column is `"NAME"`.
- The pod needs attention when its READY column (the 2nd) is not ready **and**
  its STATUS column (the 3rd) is not `"Completed"`.

For the example output above:

```go
NotReadyPods(output)
// => ["cache-0" "web-5f6b8d7c9-k8s2m"]
```

When every pod is fine, or there are no pods, return an empty slice.

### 3. Write the summary

Change `Report` to return:

- `"all pods ready"` when `NotReadyPods` finds nothing.
- Otherwise `"not ready: "` followed by the pod names, separated by `", "`.

```go
Report(output)
// => "not ready: cache-0, web-5f6b8d7c9-k8s2m"
```

## Learn more

- [kubectl get](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_get/)
- [Kubernetes: Pod lifecycle](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/)
- [Package strings](https://pkg.go.dev/strings)
