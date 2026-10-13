## Task 1: Fix the collection window

Each call to `time.After` makes a brand-new timer. Look at where `time.After(window)` is called: it runs again on every pass through the loop, so the timeout restarts after each event.

### Closer

Create the timer once, before the `for` loop, and keep its channel in a variable. Then the `case` receives from that variable: `deadline := time.After(window)` … `case <-deadline:`.

### Answer

```go
deadline := time.After(window)
for {
	select {
	case event, ok := <-events:
		if !ok {
			return collected
		}
		collected = append(collected, event)
	case <-deadline:
		return collected
	}
}
```

The single timer started when `CollectFor` was called, so `window` is measured from the start.

## Task 2: Send heartbeats

You need a ticker that fires every `interval`, and a loop that waits for either a tick or `stop` being closed. Remember to stop the ticker when the function returns.

### Closer

Make it with `ticker := time.NewTicker(interval)` and follow it with `defer ticker.Stop()`. Inside `for { select { ... } }`, use two cases: `case <-ticker.C:` and `case <-stop:`.

### Answer

```go
ticker := time.NewTicker(interval)
defer ticker.Stop()
for {
	select {
	case <-ticker.C:
		beat()
		beats++
	case <-stop:
		return beats
	}
}
```

A closed `stop` channel is always ready, so the loop returns as soon as it closes.
