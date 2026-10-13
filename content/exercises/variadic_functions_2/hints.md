## Task 1: treat the first host differently

Return early when `len(hosts) == 0`. Otherwise start the message with the
first host, `hosts[0]`, and put `", "` in front of every host after it. The
slice `hosts[1:]` holds "every host after the first".

## Task 2: spread the slice

`Alert` accepts any number of hosts, and `down` is already a slice of them.
Put `...` after `down` in the call. An empty or `nil` slice spreads into zero
hosts, so you don't need the `if` any more.

## Task 3: one slice, then spread it

Build a new slice that starts with `canary` and continues with `others`, then
spread it into `Alert`:

```go
hosts := append([]string{canary}, others...)
return Alert("warning", hosts...)
```
