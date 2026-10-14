## Task 1: Reject jobs when the queue is full

The queue *is* the buffered channel `q.jobs`. Sending on it when it has room
queues the job; you need a send that gives up instead of waiting when it's
full. `Len` is one built-in call.

### Closer

Use a `select` with two cases: `case q.jobs <- j:` returns `nil`, and
`default:` returns `ErrFull`. For `Len`, `len` works on channels too.

### Answer

```go
// TryEnqueue
select {
case q.jobs <- j:
	return nil
default:
	return ErrFull
}

// Len
return len(q.jobs)
```

`default` runs only when the send can't happen right now, so a full queue
answers at once.

## Task 2: Wait for room until the deadline

Same `select` as in task 1, but this time it should wait. Replace the
`default` with a case that fires when the context is done.

### Closer

`ctx.Done()` returns a channel that closes when the context is cancelled or
times out. Receive from it in a case (`case <-ctx.Done():`) and return
`ctx.Err()`.

### Answer

```go
select {
case q.jobs <- j:
	return nil
case <-ctx.Done():
	return ctx.Err()
}
```

## Task 3: Run the workers and drain on Close

`Start` loops `workers` times and starts one goroutine each time with
`q.wg.Go`. Each goroutine ranges over `q.jobs`. `Close` needs one line before
the `Wait`.

### Closer

`for range workers { q.wg.Go(func() { ... }) }` starts the workers. Inside,
`for j := range q.jobs { handle(j) }` ends once the channel is closed and
empty. In `Close`, call `close(q.jobs)` before `q.wg.Wait()`.

### Answer

```go
// Start
for range workers {
	q.wg.Go(func() {
		for j := range q.jobs {
			handle(j)
		}
	})
}

// Close
close(q.jobs)
q.wg.Wait()
```

Without the `close`, the workers' loops never end and `Wait` blocks forever.

## Task 4: Answer HTTP with 202 or 503

Work through the four steps in order, and `return` after each error
response so nothing else gets written. Decode into a `Job` variable first,
then check its fields, then try to queue it.

### Closer

`var j Job` then `json.NewDecoder(r.Body).Decode(&j)`; on error,
`http.Error(w, "...", http.StatusBadRequest)` and `return`. When
`q.TryEnqueue(j)` fails, call `w.Header().Set("Retry-After", "1")` before
`http.Error(..., http.StatusServiceUnavailable)`. Success is
`w.WriteHeader(http.StatusAccepted)`.

### Answer

```go
var j Job
if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
	http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
	return
}
if j.ID == "" || j.Service == "" {
	http.Error(w, "id and service are required", http.StatusBadRequest)
	return
}
if err := q.TryEnqueue(j); err != nil {
	w.Header().Set("Retry-After", "1")
	http.Error(w, "build queue is full, try again later", http.StatusServiceUnavailable)
	return
}
w.WriteHeader(http.StatusAccepted)
```
