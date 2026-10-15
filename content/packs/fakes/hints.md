## Task 1: Write the fake

A fake is just a type that has the method the interface asks for. `Notifier` needs `Notify(message string)`, so give `fakeNotifier` that method and make it remember each message.

### Closer

Use a pointer receiver, `func (f *fakeNotifier) Notify(message string)`, so the method can change the struct. In the body, append to `f.messages` the same way you append to any slice.

### Answer

```go
type fakeNotifier struct {
	messages []string
}

func (f *fakeNotifier) Notify(message string) {
	f.messages = append(f.messages, message)
}
```

## Task 2: Check that a full disk sends one alert

Make a fake, pass it to `CheckDisk`, then look inside the fake to see what was sent. First check how many messages there are, then check the text.

### Closer

`f := &fakeNotifier{}`, then `CheckDisk(f, "web-1", 93)`. Use `t.Fatalf` if `len(f.messages) != 1`, because `f.messages[0]` would crash otherwise. Then compare `f.messages[0]` with the expected text. For the second check, start again with `f = &fakeNotifier{}`.

### Answer

```go
f := &fakeNotifier{}
CheckDisk(f, "web-1", 93)
if len(f.messages) != 1 {
	t.Fatalf("CheckDisk(web-1, 93) sent %d messages, want 1: %q", len(f.messages), f.messages)
}
if want := "disk on web-1 is 93% full"; f.messages[0] != want {
	t.Errorf("CheckDisk(web-1, 93) sent %q, want %q", f.messages[0], want)
}

f = &fakeNotifier{}
CheckDisk(f, "db-2", 90)
if len(f.messages) != 1 {
	t.Fatalf("CheckDisk(db-2, 90) sent %d messages, want 1: %q", len(f.messages), f.messages)
}
if want := "disk on db-2 is 90% full"; f.messages[0] != want {
	t.Errorf("CheckDisk(db-2, 90) sent %q, want %q", f.messages[0], want)
}
```

## Task 3: Check that a healthy disk sends nothing

Same idea, but now the fake must stay empty. Use a fresh fake for each value, so leftovers from an earlier call don't count.

### Closer

Loop over `[]int{50, 89}`. In each pass, make a new fake, call `CheckDisk`, and fail if `len(f.messages) != 0`.

### Answer

```go
for _, used := range []int{50, 89} {
	f := &fakeNotifier{}
	CheckDisk(f, "web-1", used)
	if len(f.messages) != 0 {
		t.Errorf("CheckDisk(web-1, %d) sent %q, want no messages", used, f.messages)
	}
}
```
