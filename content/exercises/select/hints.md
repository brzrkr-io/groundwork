## Task 1: one case per channel

A `select` with two receive cases runs whichever is ready first:

```go
select {
case answer := <-primary:
	return answer
case answer := <-replica:
	return answer
}
```

## Task 2: race the timer

Use two cases: `<-ready` (return `nil`) and `<-time.After(timeout)` (return
`ErrNotReady`). A closed channel is always ready to receive from, so `<-ready`
fires as soon as `ready` is closed.

## Task 3: a send case plus default

```go
select {
case alerts <- msg:
	return true
default:
	return false
}
```
