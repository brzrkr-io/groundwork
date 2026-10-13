## Task 1: Fix `Expired`

The rule says "0 or less". Look at the comparison in `Expired`: does
`c.DaysLeft < 0` include 0? Check the table in the README.

### Closer

`<` leaves the edge value out; `<=` includes it.

### Answer

```go
return c.DaysLeft <= 0
```

## Task 2: Fix `NeedsRenewal`

`NeedsRenewal` has two conditions joined by `&&`. One of them throws away
every certificate with 0 or fewer days left.

### Closer

"30 or fewer" already covers expired certificates on its own, so drop the
`c.DaysLeft > 0 &&` part and keep only the `<= RenewWithinDays` check.

### Answer

```go
return c.DaysLeft <= RenewWithinDays
```

## Task 3: Finish `Summary`

After the expired check, you need one more `if` that calls `c.NeedsRenewal()`,
and then the "ok" line for everything else. Both lines include the number of
days, so you'll need `fmt.Sprintf` (add `"fmt"` to the imports).

### Closer

Use `%s` for the domain and `%d` for the days. Write `%dd` to print the number
followed by the letter `d`: `fmt.Sprintf("%s: renew soon (%dd left)", ...)`.

### Answer

```go
if c.NeedsRenewal() {
	return fmt.Sprintf("%s: renew soon (%dd left)", c.Domain, c.DaysLeft)
}
return fmt.Sprintf("%s: ok (%dd left)", c.Domain, c.DaysLeft)
```

The expired check runs first, so `NeedsRenewal` only catches the
not-yet-expired ones.
