## Task 1: Stop ignoring the conversion error

The `_` in `n, _ := strconv.Atoi(s)` throws the error away. Give it a name and check it before using `n`.

### Closer

Replace `_` with `err`, then add `if err != nil { ... }` that returns `0` and an error built with `fmt.Errorf` and `%q`.

### Answer

```go
n, err := strconv.Atoi(s)
if err != nil {
	return 0, fmt.Errorf("replicas %q is not a number", s)
}
```

Now text like `"three"` fails loudly instead of quietly becoming `0`.

## Task 2: Fix the negative-count error

Two things are wrong in that `if`: it returns the bad number `n`, and the message breaks the style rules (capital letter, `!`, no value shown).

### Closer

Return `0` instead of `n`, and use `fmt.Errorf` with `%d` so the message shows the number. Once nothing uses `errors.New`, remove the `"errors"` import.

### Answer

```go
if n < 0 {
	return 0, fmt.Errorf("replicas %d must not be negative", n)
}
```

## Task 3: Return early from Scale

Inside `if err != nil`, the code only changes `err` and then keeps going. Make it leave the function right there.

### Closer

Put a `return` inside the `if`, with `""` for the string. After that, the two status `return` lines only run on success, so they can return `nil` as the error.

### Answer

```go
n, err := ParseReplicas(input)
if err != nil {
	return "", fmt.Errorf("scale: %v", err)
}
if n == current {
	return fmt.Sprintf("already at %d replicas", n), nil
}
return fmt.Sprintf("scaled from %d to %d replicas", current, n), nil
```

Returning early means no status message is ever built from bad data.
