# Debugging pods: describe, events and exit codes

`kubectl get pods` says *that* something is wrong. `kubectl describe pod`
and `kubectl logs` say *why*. Here you learn to read them, then write Go
that reads them for you and names the cause.

**New here:** `kubectl describe pod` (State, Last State, Events), events and their counts, reading exit codes and signals in a Pod (128 + n, from *Signals: graceful stop for a rolling drain*), `kubectl logs --previous`, `kubectl get events`, a triage order.

## Learn

### Recap

In *Pod status: reading kubectl get* you listed the Pods that need a look.
`testdata/` holds the `describe` output for each of them, captured from a
real cluster. Open `testdata/memory.txt` alongside this section.

### The parts of describe that matter

```text
Name:             report-58c6d7f9b-lm4vq
Node:             worker-1/172.18.0.3          <- empty while Pending
Containers:
  report:
    Image:          ghcr.io/stefanprodan/podinfo:6.9.2
    State:          Terminated                  <- what it is doing now
      Reason:       OOMKilled
      Exit Code:    137
    Last State:     Terminated                  <- how the previous run ended
      Reason:       OOMKilled
      Exit Code:    137
    Ready:          False
    Restart Count:  4
    Limits:
      memory:  128Mi
...
Events:                                         <- the last hour, oldest first
  Type     Reason     Age                 From      Message
  Warning  BackOff    10s (x5 over 106s)  kubelet   ... Back-off restarting failed container report ...
```

- **State** is now: `Running`, `Waiting` (with a Reason such as
  `CrashLoopBackOff`) or `Terminated`.
- **Last State** is the run before. It only appears after a restart, and
  it usually holds the real reason.
- **Events** are short notes from the scheduler and the kubelet. `(x5 over
  106s)` means the same event five times in 106 seconds. `Warning` events
  are the ones to read first.

Every line under `Last State:` that belongs to it is indented deeper than
`Last State:` itself. The block ends at the first line that isn't.

### Exit codes

The *exit code* is the number a process ends with: 0 for success, anything
else for failure. A process killed by a signal has no exit code of its
own, so the convention is 128 + the signal's number.

| Exit code | Meaning | Usual cause in Kubernetes |
|---|---|---|
| 0 | Success | A job finished; or an app that stops cleanly on `SIGTERM` |
| 1, 2, … | The app reported an error | A bad config, a missing variable: read the logs |
| 137 | 128 + 9: `SIGKILL` | Over the memory limit (`OOMKilled`), or killed after ignoring `SIGTERM` |
| 143 | 128 + 15: `SIGTERM` | Asked to stop: a rollout, a scale-down, a failing liveness probe |

Signals are explained in *The operating system for programmers* guide.

### Logs of the run that crashed

`kubectl logs POD` shows the *current* container, which may have only just
started. After a crash you want the previous one:

```bash
kubectl logs worker-5d6f7c8b9-n7vbx --previous
# starting worker
# error: DATABASE_URL is not set
```

### A triage order

1. `kubectl get pods`: which Pods, what STATUS, how many restarts?
2. `kubectl describe pod NAME`: read **Events** from the bottom, then
   **Last State**.
3. `kubectl logs NAME --previous` if it crashed; `kubectl logs NAME` if it
   runs.
4. `kubectl get events --sort-by=.lastTimestamp` for the whole namespace.

| Clue in describe | Cause |
|---|---|
| `ErrImagePull` or `ImagePullBackOff` | wrong image name or tag, or no access |
| `FailedScheduling` | no node has room (or matches the Pod's rules) |
| `CreateContainerConfigError` | a ConfigMap, Secret or key is missing |
| Last State reason `OOMKilled` | over the memory limit |
| `failed liveness probe` | the liveness probe kills it |
| Last State with a non-zero exit code | the app itself fails: read the logs |
| `Readiness probe failed` | running, but taken out of its Services |

## Your task

Build a tiny `kubectl doctor` in `exercise.go`. The fixtures in
`testdata/` are real `describe` outputs.

### 1. Decode an exit code

**Goal:** make `Signal(code)` return the signal number and `true` for an
exit code above 128, and `0, false` otherwise.

```go
Signal(137) // => 9, true
Signal(143) // => 15, true
Signal(1)   // => 0, false
Signal(128) // => 0, false
```

**Think it through:**
- What do you subtract from 137 to get 9?
- Is 128 itself a signal?

### 2. Read the last termination

**Goal:** make `LastTermination(describe)` return the `Reason` and
`Exit Code` from the `Last State:     Terminated` block.

```go
LastTermination(<testdata/memory.txt>)  // => Termination{Reason: "OOMKilled", ExitCode: 137}, nil
LastTermination(<testdata/crash.txt>)   // => Termination{Reason: "Error", ExitCode: 1}, nil
LastTermination(<testdata/healthy.txt>) // => Termination{}, ErrNoTermination
```

- Find the line whose trimmed text is `Last State:` followed by
  `Terminated`.
- Read the lines after it while they are indented **deeper** than it.
  Take `Reason:` and `Exit Code:` from them. Ignore the `State:` block
  above: it can say something different.
- No such line: return `ErrNoTermination` (it's declared for you).
- An exit code that's missing or not a number: return a different error.

**Think it through:**
- `strings.Cut(trimmed, ":")` splits `Exit Code:    137` into a key and a
  value. What do you still need to do to the value?
- How do you count the spaces at the start of a line? (Compare its length
  before and after `strings.TrimLeft(line, " ")`.)

### 3. Name the cause

**Goal:** make `Diagnose(describe)` return the cause, checking in this
order and returning the first that matches:

| # | If describe… | Return |
|---|---|---|
| 1 | contains `ErrImagePull` or `ImagePullBackOff` | `CauseImage` |
| 2 | contains `FailedScheduling` | `CauseUnschedulable` |
| 3 | contains `CreateContainerConfigError` | `CauseConfig` |
| 4 | has a last termination with reason `OOMKilled` | `CauseMemory` |
| 5 | contains `failed liveness probe` | `CauseLiveness` |
| 6 | has a last termination with a non-zero exit code | `CauseCrash` |
| 7 | contains `Readiness probe failed` | `CauseReadiness` |
| 8 | none of these | `CauseNone` |

```go
Diagnose(<testdata/config.txt>)   // => "config"
Diagnose(<testdata/liveness.txt>) // => "liveness"
```

**Think it through:**
- Which rows can use `strings.Contains`, and which need
  `LastTermination`?
- A Pod killed by its liveness probe often exits with 137. Why does the
  order of rows 5 and 6 matter?

## Run it for real

Optional. Set up kind as in *Pods: your first manifest*, then break Pods
on purpose and point your functions at real output:

```bash
nix shell nixpkgs#kind
kind create cluster --name gw-pod-debugging
kubectl run crashy --image=busybox:1.37 -- sh -c 'echo "error: DATABASE_URL is not set"; exit 3'
kubectl run hungry --image=ghcr.io/stefanprodan/podinfo:6.9.2 \
  --overrides='{"spec":{"containers":[{"name":"hungry","image":"ghcr.io/stefanprodan/podinfo:6.9.2","command":["./podinfo","--stress-memory=300"],"resources":{"limits":{"memory":"128Mi"}}}]}}'
sleep 30
kubectl describe pod crashy > crashy.txt     # Last State: Error, Exit Code 3
kubectl describe pod hungry > hungry.txt     # Last State: OOMKilled, 137
kubectl logs crashy --previous
kubectl get events --sort-by=.lastTimestamp
kind delete cluster --name gw-pod-debugging
```

## Learn more

- [Kubernetes: Debug running Pods](https://kubernetes.io/docs/tasks/debug/debug-application/debug-running-pod/)
- [Kubernetes: Determine the reason for Pod failure](https://kubernetes.io/docs/tasks/debug/debug-application/determine-reason-pod-failure/)
- [kubectl describe reference](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_describe/)
