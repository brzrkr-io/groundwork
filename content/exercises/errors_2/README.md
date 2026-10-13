# Errors are values: fix the replica scaler

A teammate wrote a small helper that scales a service to a new number of
replicas. It looks fine, but it hides errors and returns bad values next to
them. Fix it before someone scales production to zero.

## Learn

Recap from the previous exercise: a function that can fail returns an `error`
last, and `nil` means success. When something fails, return early with zero
values for the other results (`0`, `""`) and an error message that starts
lowercase, has no trailing punctuation and shows the bad value.

Three mistakes show up again and again in real code:

| Mistake | Looks like | Problem |
|---|---|---|
| Ignoring an error | `n, _ := strconv.Atoi(s)` | `_` throws the error away; bad input silently becomes `0` |
| Returning a value with the error | `return n, err` | the caller may use `n` anyway |
| Not returning early | `if err != nil { err = ... }` then carrying on | the function keeps working with bad data |

## Your task

Fix the three bugs in `exercise.go`. Each task matches one `TODO`.

### 1. Stop ignoring the conversion error

`ParseReplicas` throws away the error from `strconv.Atoi`. Check it, and when
it fails return `0` and this error:

```go
ParseReplicas("three") // => 0, error "replicas \"three\" is not a number"
ParseReplicas("4")     // => 4, nil
```

### 2. Fix the negative-count error

For a negative count, `ParseReplicas` returns the negative number together
with a badly styled message. Return `0` and this error instead (`0` itself is
fine: it means "scale down to nothing"):

```go
ParseReplicas("-2") // => 0, error "replicas -2 must not be negative"
ParseReplicas("0")  // => 0, nil
```

### 3. Return early from Scale

`Scale` adds context to the error from `ParseReplicas`, but then keeps going
and returns a status message along with it. When `ParseReplicas` fails, return
`""` and the error with context:

```go
Scale(3, "lots") // => "", error "scale: replicas \"lots\" is not a number"
Scale(3, "-1")   // => "", error "scale: replicas -1 must not be negative"
```

The successful cases already work and must keep working:

```go
Scale(3, "5") // => "scaled from 3 to 5 replicas", nil
Scale(3, "3") // => "already at 3 replicas", nil
```

## Learn more

- [Effective Go: Errors](https://go.dev/doc/effective_go#errors)
- [Go Code Review Comments: Handle errors](https://go.dev/wiki/CodeReviewComments#handle-errors)
- [Kubernetes: Scaling a Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#scaling-a-deployment)
