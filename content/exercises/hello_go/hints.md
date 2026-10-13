## Where to look

Open `exercise.go`. Each function has a `return` line with a string in double
quotes. Those strings are the only thing you need to change.

## Read the failure message

The test output reads like this:

```text
Greeting() = "hello world", want "Hello, Gopher!"
```

The value after `=` is what your code returned. The value after `want` is what
the test expected.

## Almost the answer

```go
func Greeting() string {
	return "Hello, Gopher!"
}
```

Do the same for `Status`, using `"ready to deploy"`.
