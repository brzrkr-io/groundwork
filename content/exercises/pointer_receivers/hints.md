## Task 1: one character per method

The bodies of `Inc` and `Reset` are fine. The receiver `(c Counter)` hands
them a copy. Add a `*` so it reads `(c *Counter)`, and do it for both
methods. Nothing else changes: `c.Count++` still works through the pointer.

## Task 2: refuse first, then count

Change the receiver to `(s *Server)`. Then check `s.Draining` first: if it's
`true`, return `false` straight away. After that `if`, add one to
`s.Connections` and return `true`.

## Task 3: remember before you reset

Once you set `s.Connections = 0`, the old number is gone. Save it in a
variable first, then change the fields, then return the saved value.

## Almost the answer

```go
func (s *Server) Accept() bool {
	if s.Draining {
		return false
	}
	s.Connections++
	return true
}

func (s *Server) Drain() int {
	closed := s.Connections
	s.Draining = true
	s.Connections = 0
	return closed
}
```
