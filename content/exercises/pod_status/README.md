# Pod status: reading kubectl get

`kubectl get pods` is the first command anyone runs when something is
wrong. Here you learn to read every column, then write Go that reads it for
you and lists the Pods that need a look.

**New here:** the columns of `kubectl get pods`, the common Pod statuses (`CrashLoopBackOff`, `ImagePullBackOff`, `OOMKilled`, …) and where to look next.

## Learn

### Recap

A Pod runs one or more containers on a node (*Pods: your first manifest*).
`kubectl get pods` shows one row per Pod. Here is part of
`testdata/get_pods.txt`, a real-looking listing from a busy namespace:

```text
NAME                            READY   STATUS                       RESTARTS        AGE
checkout-7b9f6d8c5d-2mvzr       1/1     Running                      0               3h12m
gateway-5c4d8b7f9-hl2jd         1/2     Running                      0               26m
report-58c6d7f9b-lm4vq          0/1     OOMKilled                    4 (71s ago)     8m
worker-5d6f7c8b9-n7vbx          0/1     CrashLoopBackOff             6 (2m14s ago)   12m
```

### The columns

| Column | Example | Meaning |
|---|---|---|
| NAME | `checkout-7b9f6d8c5d-2mvzr` | The Pod's name. A Deployment adds the two random parts. |
| READY | `1/2` | Containers passing their readiness check / containers in the Pod |
| STATUS | `Running` | A one-word summary that kubectl works out from the Pod's state |
| RESTARTS | `4 (71s ago)` | How often a container was restarted, and when it last was |
| AGE | `3h12m` | Time since the Pod was created |

`gateway` above is `Running` but only `1/2`: one of its containers isn't
ready, so it gets no traffic. **Running is not the same as healthy**: read
READY too.

### The statuses you'll meet

| STATUS | What happened | Look next |
|---|---|---|
| `Running` | The containers started. | READY and RESTARTS |
| `Completed` | Every container exited with code 0, as a one-off job should. | nothing |
| `Pending` | No node has been picked yet, usually because none has room. | `kubectl describe pod`: the Events |
| `ContainerCreating` | A node was picked; it is pulling the image or mounting volumes. | wait, then `describe` |
| `ErrImagePull`, `ImagePullBackOff` | The image can't be pulled: wrong name or tag, or no access. | `describe`: "Failed to pull image" |
| `Error` | The container exited with a non-zero code. | `kubectl logs` |
| `CrashLoopBackOff` | It keeps exiting, so Kubernetes waits longer before each restart (10s, 20s, 40s, … up to 5 minutes). | `kubectl logs --previous` |
| `OOMKilled` | The kernel killed it for using more memory than its limit. | `describe`: Last State and Limits |
| `CreateContainerConfigError` | Something the container needs is missing, such as a ConfigMap key. | `describe`: the Events |

*BackOff* means "waiting before retrying". A crashing Pod switches between
`Error` and `CrashLoopBackOff` as it is restarted and waits again, so two
runs of `kubectl get pods` can show different words for the same problem.

A few useful variations:

```bash
kubectl get pods -o wide          # adds the Pod's IP and node
kubectl get pods --show-labels    # adds the labels
kubectl get pods -w               # keeps printing changes; Ctrl+C to stop
```

### Reading the RESTARTS column in Go

`strings.Fields` (from *The strings package*) splits on any run of spaces, so
the padding doesn't matter. But `4 (71s ago)` holds spaces too:

```go
strings.Fields("report   0/1   OOMKilled   4 (71s ago)   8m")
// => [report 0/1 OOMKilled 4 (71s ago) 8m]   (seven fields, not five)
```

Columns 0 to 3 are always the name, READY, STATUS and the restart count, so
read those by position and ignore the rest.

## Your task

The on-call engineer wants a quick "what's broken?" list. Build it in three
steps in `exercise.go`.

### 1. Split the READY column

**Goal:** make `ParseReady(s)` return the two numbers in a READY value.

```go
ParseReady("1/2")   // => 1, 2, nil
ParseReady("0/1")   // => 0, 1, nil
ParseReady("11")    // => 0, 0, error mentioning "11" (with the quotes)
```

Return an error when there is no `/` or either half isn't a whole number.
The error message must include `s` in double quotes: use `%q`.

**Think it through:**
- Which `strings` function splits a string at the first `/` and tells you
  whether it found one?
- Which function turns `"12"` into `12`, with an error for `"two"`?

### 2. Parse one row

**Goal:** make `ParseLine(line)` return a `Pod` with the name, both READY
numbers, the status and the restart count.

```go
ParseLine("worker-5d6f7c8b9-n7vbx   0/1   CrashLoopBackOff   6 (2m14s ago)   12m")
// => Pod{Name: "worker-5d6f7c8b9-n7vbx", Ready: 0, Total: 1,
//        Status: "CrashLoopBackOff", Restarts: 6}, nil
```

Return an error when the row has fewer than five columns, or when READY or
the restart count doesn't parse. Use `ParseReady` for READY.

**Think it through:**
- After splitting, which positions hold the four values you need?
- Why check the number of columns before reading position 3?

### 3. List the Pods that need attention

**Goal:** make `NeedsAttention(output)` return the names of the Pods that
aren't fine, in the order they appear.

A Pod is **fine** when:

- its status is `Completed`, or
- its status is `Running`, every container is ready (READY like `2/2`),
  and it has fewer than 3 restarts.

Everything else needs attention. The first line is the header; skip it,
and skip blank lines. Return the first parse error you hit.

```go
NeedsAttention(<the contents of testdata/get_pods.txt>)
// => [cart-6f8d9c7b54-9xkqp gateway-5c4d8b7f9-hl2jd indexer-85f7c9d6b-t4wcx
//     payments-6d5f8b9c47-7zkxn report-58c6d7f9b-lm4vq search-7c8d9f6b5-w2xpr
//     settings-sync-6b7c8d9f5-d9kwm worker-5d6f7c8b9-n7vbx]
```

**Think it through:**
- `payments` is `Running` and `1/1`. Why is it on the list?
- Is it easier to write "fine" as one condition and negate it, or to list
  every bad case?

## Run it for real

Optional. Start a cluster and break some Pods on purpose (see *Pods: your
first manifest* if kind isn't set up):

```bash
nix shell nixpkgs#kind
kind create cluster --name gw-pod-status
kubectl run crashy --image=busybox:1.37 -- sh -c 'echo starting; exit 1'
kubectl run typo --image=ghcr.io/stefanprodan/podinfo:6.9.9
kubectl run fine --image=ghcr.io/stefanprodan/podinfo:6.9.2
kubectl get pods -w        # watch crashy go Error -> CrashLoopBackOff; Ctrl+C
kubectl get pods > pods.txt
```

`kubectl run NAME --image=IMAGE` creates a single Pod without a manifest:
handy for experiments, never for real apps. Feed `pods.txt` to your
function from a small test or `main`, then clean up:

```bash
kind delete cluster --name gw-pod-status
```

## Learn more

- [Kubernetes: Pod lifecycle](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/)
- [Kubernetes: Debug Pods](https://kubernetes.io/docs/tasks/debug/debug-application/debug-pods/)
- [kubectl get reference](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_get/)
