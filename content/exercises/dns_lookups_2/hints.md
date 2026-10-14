## Task 1: Fix the expiry time

`time.Duration(r.TTL)` turns 300 into 300 *nanoseconds*. The TTL is in
seconds, so the duration needs scaling to the right unit.

### Closer

Multiply the converted number by a unit constant from the `time` package:
`time.Duration(n) * time.Minute` is `n` minutes.

### Answer

```go
return fetched.Add(time.Duration(r.TTL) * time.Second)
```

## Task 2: Count down the TTL in `Get`

Read the clock once, into a variable. If it isn't before `e.expires`, the copy
is gone. Otherwise the seconds left are the gap between `now` and `e.expires`.

### Closer

`!now.Before(e.expires)` means "expired". The gap is `e.expires.Sub(now)`, a
`time.Duration`; divide it by `time.Second` and wrap it in `int(...)` to get
whole seconds.

### Answer

```go
now := c.now()
if !now.Before(e.expires) {
	return Record{}, false
}
r := e.record
r.TTL = int(e.expires.Sub(now) / time.Second)
return r, true
```

## Task 3: Check a migration

First loop over `answers` and collect the resolvers that still return
`m.Old.Addr`. If there are none, you're done. Otherwise compare `now` with the
expiry from `ExpiresAt` to choose between "waiting" and "stuck".

### Closer

`var stale []string` and `stale = append(stale, a.Resolver)` in the loop.
`strings.Join(stale, ", ")` builds the list. The expiry is
`ExpiresAt(m.Old, m.Changed)`, printed with `expires.Format(time.RFC3339)`;
the time left is `expires.Sub(now)` with `%v`.

### Answer

```go
var stale []string
for _, a := range answers {
	if a.Addr == m.Old.Addr {
		stale = append(stale, a.Resolver)
	}
}
if len(stale) == 0 {
	return fmt.Sprintf("done: all %d resolvers return %s", len(answers), m.NewAddr)
}
expires := ExpiresAt(m.Old, m.Changed)
list := strings.Join(stale, ", ")
if now.Before(expires) {
	return fmt.Sprintf("waiting: %s still cached on %s until %s (%v left)",
		m.Old.Addr, list, expires.Format(time.RFC3339), expires.Sub(now))
}
return fmt.Sprintf("stuck: %s still served by %s after its TTL ran out at %s",
	m.Old.Addr, list, expires.Format(time.RFC3339))
```
