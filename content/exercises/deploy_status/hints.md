## Task 1: Write the design

Answer the questions in the README's table, one section at a time, in plain
sentences. Start with Estimates: pick numbers, multiply, and write the
results down. Then delete every line containing `TODO`.

### Closer

A quick estimate: 300 services × 10 deploys × 2 events = 6,000 events a
day, about 0.07 per second. For Trade-offs, use the shape "- X instead of
Y: we gain ..., we give up ..., which is fine because ...". The queue,
the duplicate check and invalidate-on-write are three decisions worth a
bullet each.

### Answer

```markdown
## Trade-offs

- Queue instead of writing during the request: CI gets an answer at once
  and bursts are smoothed out, but a status can lag an event by a moment,
  and queued events are lost if the process crashes.
- Invalidate instead of updating the cache on write: the next read pays
  one database read, but the cache can never hold a value the database
  doesn't.
```

The reference `DESIGN.md` you'll see when you finish has all four sections
in full.

## Task 2: Accept events with backpressure

Two steps: reject bad events first, then try to send without waiting. A
`select` with a `default` case is the non-blocking send.

### Closer

Check the three strings with `== ""` joined by `||`, and the state with a
`switch` that has an empty case for the three good values and a `default`
that returns `fmt.Errorf("%w: ...", ErrInvalidEvent, ...)`. Then:

```go
select {
case s.queue <- ev:
	// queued
default:
	// full
}
```

### Answer

```go
if ev.ID == "" || ev.Service == "" || ev.Version == "" {
	return fmt.Errorf("%w: id, service and version are required", ErrInvalidEvent)
}
switch ev.State {
case "started", "succeeded", "failed":
default:
	return fmt.Errorf("%w: unknown state %q", ErrInvalidEvent, ev.State)
}
select {
case s.queue <- ev:
	return nil
default:
	return ErrQueueFull
}
```

## Task 3: Apply each event once

First look `ev.ID` up in `s.processed` and return early if it's there. Then
three writes: the status, the processed ID, and removing the cache entry.

### Closer

`s.processed[ev.ID]` is `false` for a missing key, so it works directly in
an `if`. Build the status with
`Status{Service: ev.Service, Version: ..., State: ...}` and finish with
`s.cache.Remove(...)`.

### Answer

```go
if s.processed[ev.ID] {
	return
}
s.statuses[ev.Service] = Status{Service: ev.Service, Version: ev.Version, State: ev.State}
s.processed[ev.ID] = true
s.cache.Remove(ev.Service)
```

## Task 4: Read through the cache

Try `s.cache.Get` first and return on a hit. On a miss, count the read,
look in `s.statuses`, and only `Put` into the cache when the service was
found.

### Closer

```go
if st, ok := s.cache.Get(service); ok {
	return st, true
}
s.dbReads++
st, ok := s.statuses[service]
// fill the cache only when ok
```

### Answer

```go
if st, ok := s.cache.Get(service); ok {
	return st, true
}
s.dbReads++
st, ok := s.statuses[service]
if ok {
	s.cache.Put(service, st)
}
return st, ok
```

A missing map key gives the zero `Status` and `false`, exactly what a miss
returns.

## Task 5: Serve it over HTTP

`handleEvent` decodes with `json.NewDecoder(r.Body).Decode(&ev)`, calls
`Submit`, and picks a status code from the error. `handleStatus` gets the
name with `r.PathValue("service")` and calls `Lookup`.

### Closer

Use a `switch` on the error: `case errors.Is(err, ErrQueueFull):` first
(set the header, then `http.Error(w, ..., http.StatusServiceUnavailable)`),
then `case err != nil:` for 400, and `default:` for
`w.WriteHeader(http.StatusAccepted)`. For the status reply, set the header
before `json.NewEncoder(w).Encode(st)`. Add `"encoding/json"` to the
imports.

### Answer

```go
// handleEvent
var ev Event
if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
	http.Error(w, "bad JSON: "+err.Error(), http.StatusBadRequest)
	return
}
err := s.Submit(ev)
switch {
case errors.Is(err, ErrQueueFull):
	w.Header().Set("Retry-After", "1")
	http.Error(w, err.Error(), http.StatusServiceUnavailable)
case err != nil:
	http.Error(w, err.Error(), http.StatusBadRequest)
default:
	w.WriteHeader(http.StatusAccepted)
}

// handleStatus
service := r.PathValue("service")
st, ok := s.Lookup(service)
if !ok {
	http.Error(w, "no deploys recorded for "+service, http.StatusNotFound)
	return
}
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(st)
```
