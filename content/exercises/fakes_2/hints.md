## The fake looks like the notifier fake

`fakeSleeper` is built just like `fakeNotifier` in *Fakes: a recording
notifier*: a struct with a slice field, and a pointer-receiver method that
appends to it.

```go
type fakeSleeper struct {
	slept []time.Duration
}
```

## Count calls with a closure

```go
calls := 0
fn := func() error {
	calls++
	if calls < 3 {
		return errors.New("connection refused")
	}
	return nil
}
```

After `Retry` returns, `calls` tells you how many times `fn` ran.

## Task 2, nearly done

```go
s := &fakeSleeper{}
// ... fn as above ...
if err := Retry(s, 5, fn); err != nil {
	t.Fatalf("Retry returned error %v, want nil", err)
}
if calls != 3 {
	t.Errorf("fn was called %d times, want 3", calls)
}
if want := []time.Duration{time.Second, 2 * time.Second}; !slices.Equal(s.slept, want) {
	t.Errorf("Retry slept %v, want %v", s.slept, want)
}
```

Task 3 is the same with attempts 3, an `fn` that always fails, and
`err == nil` as the failure.
