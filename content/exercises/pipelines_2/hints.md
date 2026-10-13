## Task 1: a select around the send

```go
select {
case out <- line:
case <-ctx.Done():
	return
}
```

`return` ends the goroutine, and the deferred `close(out)` still runs.

## Task 2: forwarders, a WaitGroup and a closer

Make `out`. For each input, `wg.Go` a forwarder that ranges over the input
and sends each value with the same `select` as task 1. After the loop, start
a goroutine that does `wg.Wait()` and then `close(out)`. Return `out`.

## Task 3: a slice of outputs

```go
outputs := make([]<-chan string, workers)
for i := range outputs {
	outputs[i] = checkStage(ctx, hosts, check)
}
return Merge(ctx, outputs...)
```

`outputs...` passes the slice to the variadic parameter, as in *Variadic
functions*.
