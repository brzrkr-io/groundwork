## Task 1: Use whichever replica answers first

The starter only waits on `primary`. You need a statement that waits on both channels and runs whichever is ready first.

### Closer

Use `select` with one receive case per channel, for example `case answer := <-primary:`, and return the answer from each case.

### Answer

```go
select {
case answer := <-primary:
	return answer
case answer := <-replica:
	return answer
}
```

`select` runs only the case that is ready first.

## Task 2: Wait for a service with a timeout

Race two things in a `select`: the `ready` channel closing, and a timer running out. A closed channel is always ready to receive from.

### Closer

One case is `case <-ready:`, the other is `case <-time.After(timeout):`. Return `nil` from the first and `ErrNotReady` from the second.

### Answer

```go
select {
case <-ready:
	return nil
case <-time.After(timeout):
	return ErrNotReady
}
```

Whichever happens first wins, so you return as soon as `ready` closes or as soon as the timeout passes.

## Task 3: Drop alerts instead of blocking

You want to try a send without waiting. A `select` with a `default` case never waits.

### Closer

Put the send in a case, `case alerts <- msg:`, and add a `default:` that runs if the send can't happen right now.

### Answer

```go
select {
case alerts <- msg:
	return true
default:
	return false
}
```

If the buffer is full, the send case isn't ready, so `default` runs straight away.
