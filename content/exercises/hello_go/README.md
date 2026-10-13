# Hello, Go

Welcome to Gopherlings. Every exercise is a small Go program that is nearly
finished. You fix it, save, and Gopherlings re-runs the tests straight away.

## Learn

### How an exercise works

Each exercise folder holds a few files:

| File | What it is |
|---|---|
| `exercise.go` | The code you edit. |
| `exercise_test.go` | Tests that check your code. Read them, but don't change them. |
| `README.md` | These instructions. |

Gopherlings shows each test as ✓ (passing) or ✗ (failing), along with the
failure message. A failure message reads "you returned X, the test wanted Y".

### Your first Go code

```go
package hello

// Greeting returns a friendly message.
func Greeting() string {
	return "hello"
}
```

Line by line:

- `package hello` names the package. Every `.go` file starts by saying which
  package it belongs to. Files in the same folder share one package.
- `// Greeting returns...` is a comment. Go ignores everything after `//` on a
  line. A comment directly above a function documents it.
- `func Greeting() string` declares a function called `Greeting`. It takes no
  inputs (the empty `()`), and it gives back a `string`, which is a piece of text.
- `{ ... }` is the function body: the code that runs when the function is called.
- `return "hello"` hands the value `"hello"` back to whoever called the
  function. Text in double quotes is a *string literal*.

## Your task

Your team's deploy bot greets people with a couple of fixed messages. Both
messages are wrong.

### 1. Fix the greeting

Change `Greeting` so it returns exactly:

```text
Hello, Gopher!
```

Capital letters, the comma, the space and the exclamation mark all count.

### 2. Fix the status message

Change `Status` so it returns exactly:

```text
ready to deploy
```

When both tests pass, Gopherlings marks the exercise done. Press `n` to move
on.

## Learn more

- [A Tour of Go: Packages](https://go.dev/tour/basics/1)
- [A Tour of Go: Functions](https://go.dev/tour/basics/4)
- [Go spec: String literals](https://go.dev/ref/spec#String_literals)
