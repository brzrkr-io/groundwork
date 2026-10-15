## Task 1: Make the counter count

The bodies of `Inc` and `Reset` are fine. Look at the receiver `(c Counter)`:
it hands each method a copy of the counter.

### Closer

Add a `*` in front of the type in the receiver, `(c *Counter)`, for both
methods. `c.Count++` still works through the pointer.

### Answer

```go
func (c *Counter) Inc() {
	c.Count++
}

func (c *Counter) Reset() {
	c.Count = 0
}
```

## Task 2: Accept a connection

First change the receiver to `(s *Server)` so the new connection sticks. Then
handle the draining case before counting anything.

### Closer

Check `if s.Draining { return false }` first. After that `if`, increment
with `s.Connections++` and return `true`.

### Answer

```go
func (s *Server) Accept() bool {
	if s.Draining {
		return false
	}
	s.Connections++
	return true
}
```

## Task 3: Drain the server

Once you set `s.Connections = 0`, the old number is gone. You need to
remember it before you change anything.

### Closer

Save it first with `closed := s.Connections`, then set `s.Draining` and
`s.Connections`, then return `closed`.

### Answer

```go
closed := s.Connections
s.Draining = true
s.Connections = 0
return closed
```
