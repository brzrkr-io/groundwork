## A fake is just a type with the method

Task 1 is the struct from the Learn section plus one small method. The
method takes `message string`, has the receiver
`(f *fakeNotifier)`, and its body appends the message to `f.messages`, the
same way you append to any slice.

## Run, then look at the fake

Every check in `TestCheckDisk` has the same shape: make a fresh fake with
`f := &fakeNotifier{}`, call `CheckDisk(f, ...)`, then compare `f.messages`
with what you expect. For a later check, reuse `f` with `f = &fakeNotifier{}`.

## Task 2, nearly done

```go
f := &fakeNotifier{}
CheckDisk(f, "web-1", 93)
if len(f.messages) != 1 {
	t.Fatalf("CheckDisk(web-1, 93) sent %d messages, want 1: %q", len(f.messages), f.messages)
}
if want := "disk on web-1 is 93% full"; f.messages[0] != want {
	t.Errorf("CheckDisk(web-1, 93) sent %q, want %q", f.messages[0], want)
}
```

Repeat for `"db-2"` at `90`.

## Task 3, nearly done

```go
for _, used := range []int{50, 89} {
	f = &fakeNotifier{}
	CheckDisk(f, "web-1", used)
	if len(f.messages) != 0 {
		t.Errorf("CheckDisk(web-1, %d) sent %q, want no messages", used, f.messages)
	}
}
```
