# Hello, Go

Welcome to Groundwork. Every exercise is a small Go program that is nearly
finished. You fix it, save, and Groundwork re-runs the tests straight away.

**New here:** packages, comments, functions, `return`, string literals.

## Learn

### How an exercise works

Each exercise folder holds a few files:

| File | What it is |
|---|---|
| `exercise.go` | The code you edit. |
| `exercise_test.go` | Tests that check your code. Read them, but don't change them. |
| `README.md` | These instructions. |

Groundwork shows each test as ✓ (passing) or ✗ (failing). A failure message
reads "you returned X, the test wanted Y".

Look for `// TODO(1)`, `// TODO(2)` and so on in `exercise.go`. Each one marks
the spot for the task with that number.

### Your first Go code

```go
package hello

// Farewell returns a goodbye message.
func Farewell() string {
	return "bye"
}
```

Line by line:

- `package hello` names the package. Every `.go` file starts with it.
- `// Farewell returns...` is a *comment*. Go ignores everything after `//`.
- `func Farewell() string {` starts a *function* called `Farewell`.
- `return "bye"` hands the text `bye` back to whoever called the function.
- `}` ends the function.

### The parts of a function line

```text
func   Farewell   ()        string    {
 │        │        │          │       │
keyword  name    inputs    result   body starts
```

- `func` says "a function starts here".
- `Farewell` is its name. You use the name to *call* the function.
- `()` holds the inputs. Empty means it takes none.
- `string` is the type of value it gives back. A `string` is text.
- `{` opens the *body*, the code that runs. `}` closes it.

> **Why is it written like this?**
> - **Go requires** `func`, the `()` (even when empty), and the `{` on the
>   same line as `func`. Move the `{` to the next line and it won't compile.
> - **Go requires** a `return` with a value of the result type. A function
>   that says `string` must return a string.
> - **Convention:** a comment above a function starts with the function's
>   name. Tools show it as documentation.
> - **Your choice:** the name `Farewell` and the text inside the quotes.

### Text goes in double quotes

A *string literal* is text written straight into the code, between double
quotes:

```go
return "bye"   // returns the text: bye
```

The quotes are not part of the text. They tell Go where the text starts and
ends. Everything between them counts exactly: capitals, spaces, punctuation.

> **Why is it written like this?** **Go requires** the quotes. Without them,
> Go reads `bye` as the name of something in your code, and fails with
> `undefined: bye`.

## Your task

Your team's deploy bot greets people with a couple of fixed messages. Both
messages are wrong.

### 1. Fix the greeting

**Goal:** make `Greeting` return the text `Hello, Gopher!`.

```go
Greeting()
// => "Hello, Gopher!"
```

Capital letters, the comma, the space and the exclamation mark all count.

**Think it through:**
- Find the `return` line inside `Greeting`. Which part of it is the text?
- Which characters must stay, so Go still knows where the text starts and ends?

### 2. Fix the status message

**Goal:** make `Status` return the text `ready to deploy`.

```go
Status()
// => "ready to deploy"
```

It is all lowercase, with one space between words.

**Think it through:**
- This is the same kind of change as task 1, in a different function.

When both tests pass, Groundwork marks the exercise done. Press `n` to move
on.

## Learn more

- [A Tour of Go: Packages](https://go.dev/tour/basics/1)
- [A Tour of Go: Functions](https://go.dev/tour/basics/4)
- [Go spec: String literals](https://go.dev/ref/spec#String_literals)
