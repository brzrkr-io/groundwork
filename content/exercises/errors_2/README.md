# Errors are values: fix the replica scaler

A teammate wrote a helper that scales a service to a new number of replicas.
It hides errors and returns bad values next to them. Fix it before someone
scales production to zero by accident.

**New here:** nothing; practice fixing error-handling bugs.

## Learn

### Recap

From *Errors are values*:

- A function that can fail returns an `error` **last**. `nil` means success.
- On failure, return straight away. Other results get zero values (`0`, `""`).
- Messages start lowercase, have no full stop, and show the bad value.
- `fmt.Errorf("outer: %v", err)` puts context in front of another error.

### Three common bugs

| Bug | Looks like | What goes wrong |
|---|---|---|
| Ignoring an error | `size, _ := strconv.Atoi(s)` | `_` throws the error away; bad text silently becomes `0` |
| Returning a value with the error | `return size, err` | the caller may use `size` anyway |
| Not returning early | `if err != nil { err = ... }` and carrying on | the function keeps working with bad data |

> **Why is it written like this?** `_` is the *blank identifier*. Go requires
> you to use every variable you declare, so `_` is how you say "I don't want
> this one". For errors, that is almost always a bug.

### Watch the imports

Go refuses to compile a file with an import it doesn't use. If you stop using
`errors.New`, remove `"errors"` from the import list too.

## Your task

Fix the three bugs in `exercise.go`. Each task matches one `TODO`.

### 1. Stop ignoring the conversion error

**Goal:** `ParseReplicas` returns `0` and an error when the text is not a number.

Right now the error from `strconv.Atoi` goes into `_`. Keep it and check it.
The message shows the input text in quotes:

```go
ParseReplicas("three") // => 0, error `replicas "three" is not a number`
ParseReplicas("")      // => 0, error `replicas "" is not a number`
ParseReplicas("4")     // => 4, nil
```

**Think it through:**
- What name should replace `_` so you can check it on the next line?
- The bad value is the *text* parameter `s`. Which verb adds the quotes?

### 2. Fix the negative-count error

**Goal:** for a negative count, `ParseReplicas` returns `0` and a well-styled error that shows the number.

`0` itself is fine: it means "scale down to nothing".

```go
ParseReplicas("-2")  // => 0, error "replicas -2 must not be negative"
ParseReplicas("-10") // => 0, error "replicas -10 must not be negative"
ParseReplicas("0")   // => 0, nil
```

**Think it through:**
- Which of the two returned values is wrong today, and which value should it be?
- The message now includes the number. Which error tool can do that?
- Compare the old message with the new one: capital letter, `!`, missing value.

### 3. Return early from Scale

**Goal:** when `ParseReplicas` fails, `Scale` returns `""` and the error with `scale: ` in front.

```go
Scale(3, "lots") // => "", error `scale: replicas "lots" is not a number`
Scale(3, "-1")   // => "", error "scale: replicas -1 must not be negative"
```

The success cases already work and must keep working, with a `nil` error:

```go
Scale(3, "5") // => "scaled from 3 to 5 replicas", nil
Scale(4, "0") // => "scaled from 4 to 0 replicas", nil
Scale(3, "3") // => "already at 3 replicas", nil
```

**Think it through:**
- Inside the `if err != nil` block, the code assigns but doesn't stop. What keyword stops a function?
- Look at the two success `return` lines. After your fix, is `err` there always `nil`? What could you write instead?

## Learn more

- [Effective Go: Errors](https://go.dev/doc/effective_go#errors)
- [Go Code Review Comments: Handle errors](https://go.dev/wiki/CodeReviewComments#handle-errors)
- [Kubernetes: Scaling a Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#scaling-a-deployment)
