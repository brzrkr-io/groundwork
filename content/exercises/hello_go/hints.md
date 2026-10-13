## Task 1: Fix the greeting

Open `exercise.go` and find `Greeting`. It has a `return` line with some text
in double quotes. That text is the only thing you need to change.

### Closer

The test output shows what you returned and what it wanted:
`Greeting() = "hello world", want "Hello, Gopher!"`. Replace the text between
the quotes with the wanted text, keeping the quotes. Capitals, the comma, the
space and the `!` all count.

### Answer

```go
func Greeting() string {
	return "Hello, Gopher!"
}
```

## Task 2: Fix the status message

Find `Status` in `exercise.go`. Just like `Greeting`, it returns some text in
double quotes that needs to change.

### Closer

Keep the `return` and the double quotes, and put `ready to deploy` between
them. It is all lowercase, with single spaces.

### Answer

```go
func Status() string {
	return "ready to deploy"
}
```
