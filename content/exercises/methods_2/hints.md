## Task 1: is 0 included?

The rule says "0 or less". Look at the comparison in `Expired`: does
`c.DaysLeft < 0` include 0? Check the table in the README.

## Task 2: drop the extra condition

`NeedsRenewal` has two conditions joined by `&&`. One of them throws away
every certificate with 0 or fewer days left. "30 or fewer" already covers
expired certificates on its own.

## Task 3: one more if, then Sprintf

After the expired check, add `if c.NeedsRenewal() { ... }`. Build both
remaining lines with `fmt.Sprintf`, using `%s` for the domain and `%d` for the
days. Write `%dd` to print the number followed by the letter `d`.

## Almost the answer

```go
if c.NeedsRenewal() {
	return fmt.Sprintf("%s: renew soon (%dd left)", c.Domain, c.DaysLeft)
}
return fmt.Sprintf("%s: ok (%dd left)", c.Domain, c.DaysLeft)
```

Fix 1 is `c.DaysLeft <= 0`, and fix 2 is `c.DaysLeft <= RenewWithinDays`.
