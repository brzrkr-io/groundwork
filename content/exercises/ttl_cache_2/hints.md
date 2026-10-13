## Task 1: After vs Before

`now.After(expiresAt)` is false when the two times are equal, so the entry is
still served at its expiry time. "Expired" means "not before":
`!c.now().Before(e.expiresAt)`.

## Task 2: three steps

Try `c.Get(k)` first and return early on a hit. After the leader's
`v, err := load(k)`, call `c.Set(k, v)` only when `err == nil`.

## Task 3: startCall

`startCall` mirrors a map lookup: add `if cl, ok := c.inflight[k]; ok {
return cl, false }` before the code that builds a new call. The
`defer c.mu.Unlock()` already unlocks on every return.

## Near-solution

```go
func (c *Cache[K, V]) GetOrLoad(k K, load func(K) (V, error)) (V, error) {
	if v, ok := c.Get(k); ok {
		return v, nil
	}
	cl, leader := c.startCall(k)
	if !leader {
		<-cl.done
		return cl.val, cl.err
	}
	v, err := load(k)
	if err == nil {
		c.Set(k, v)
	}
	c.finishCall(k, cl, v, err)
	return v, err
}
```
