## Task 1: Ack only after the deploy succeeds

Acking says "done, forget it", so it can only happen once the deploy has
worked. Move the ack below the deploy, and give the failure branch its own
call that sends the message back.

### Closer

Delete the early `c.broker.Ack(m.ID)`. Inside the
`if err := c.deploy(...); err != nil { ... }` block, call
`c.broker.Nack(m.ID)` before returning. After that block, call
`c.broker.Ack(m.ID)`.

### Answer

```go
if err := c.deploy(m.Service, m.Version); err != nil {
	c.broker.Nack(m.ID)
	return true, err
}
c.broker.Ack(m.ID)
return true, nil
```

## Task 2: Skip messages that were already deployed

Before deploying, look the message's ID up in `c.done`. If it's there, the
work is already done: tell the broker so, and stop.

### Closer

`if c.done[m.ID] { ... }` is true only for IDs you recorded. Inside, call
`c.broker.Ack(m.ID)` and `return true, nil`.

### Answer

```go
if c.done[m.ID] {
	c.broker.Ack(m.ID)
	return true, nil
}
```

Without the ack, the duplicate stays in flight and the broker delivers it
again and again.

## Task 3: Record a deploy only after it succeeds

`c.done[m.ID] = true` runs before the deploy, so even a failed deploy counts
as done. Move it to where the deploy is known to have worked.

### Closer

Move the line to just after the `if err := c.deploy(...)` block, before
`c.broker.Ack(m.ID)`.

### Answer

```go
if err := c.deploy(m.Service, m.Version); err != nil {
	c.broker.Nack(m.ID)
	return true, err
}
c.done[m.ID] = true
c.broker.Ack(m.ID)
return true, nil
```

Recording before the ack means a crash between the two lines leads to a
skipped redelivery, not a second deploy.
