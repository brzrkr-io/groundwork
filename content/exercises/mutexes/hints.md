## Task 1: lock, defer unlock, change

```go
c.mu.Lock()
defer c.mu.Unlock()
c.hits[path]++
```

## Task 2: the read lock

Same shape as task 1, but with `c.mu.RLock()` and `defer c.mu.RUnlock()`.
Reading a missing key from a map already gives 0, so `return c.hits[path]` is
enough.

## Task 3: copy under the read lock

Take the read lock, then `return maps.Clone(c.hits)`. Add `"maps"` to the
import block:

```go
import (
	"maps"
	"sync"
)
```

If the test says "race detected", check that every method that touches
`c.hits` takes the lock first.
