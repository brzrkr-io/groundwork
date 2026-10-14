## Task 1: Write a fake sleeper

`fakeSleeper` is built just like `fakeNotifier` in *Fakes and dependency injection*: a struct with a slice field, and a method that appends to it instead of really sleeping.

### Closer

The field is `slept []time.Duration`. The method is `func (f *fakeSleeper) Sleep(d time.Duration)`, with a pointer receiver so the append sticks.

### Answer

```go
type fakeSleeper struct {
	slept []time.Duration
}

func (f *fakeSleeper) Sleep(d time.Duration) {
	f.slept = append(f.slept, d)
}
```

## Task 2: Test a call that succeeds on the third try

You need an `fn` that fails twice and then works, and you need to count how often `Retry` called it. A variable outside `fn` can do the counting.

### Closer

A closure can change a variable declared outside it:

```go
calls := 0
fn := func() error {
	calls++
	// return an error while calls < 3, then nil
}
```

Compare durations with `slices.Equal(s.slept, []time.Duration{time.Second, 2 * time.Second})`.

### Answer

```go
func TestRetrySucceedsOnThirdTry(t *testing.T) {
	s := &fakeSleeper{}
	calls := 0
	fn := func() error {
		calls++
		if calls < 3 {
			return errors.New("connection refused")
		}
		return nil
	}

	if err := Retry(s, 5, fn); err != nil {
		t.Fatalf("Retry returned error %v, want nil", err)
	}
	if calls != 3 {
		t.Errorf("fn was called %d times, want 3", calls)
	}
	if want := []time.Duration{time.Second, 2 * time.Second}; !slices.Equal(s.slept, want) {
		t.Errorf("Retry slept %v, want %v", s.slept, want)
	}
}
```

## Task 3: Test giving up

This test has the same shape as Task 2. The differences are that `fn` always fails, attempts is `3`, and now an error is what you want.

### Closer

`fn` still counts calls but always returns `errors.New("connection refused")`. Fail if `err == nil`. The sleeps are still `[1s 2s]`, with no sleep after the last attempt.

### Answer

```go
func TestRetryGivesUp(t *testing.T) {
	s := &fakeSleeper{}
	calls := 0
	fn := func() error {
		calls++
		return errors.New("connection refused")
	}

	if err := Retry(s, 3, fn); err == nil {
		t.Errorf("Retry returned nil, want an error")
	}
	if calls != 3 {
		t.Errorf("fn was called %d times, want 3", calls)
	}
	if want := []time.Duration{time.Second, 2 * time.Second}; !slices.Equal(s.slept, want) {
		t.Errorf("Retry slept %v, want %v", s.slept, want)
	}
}
```
