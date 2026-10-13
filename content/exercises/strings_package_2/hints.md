## Look at the inputs that fail

Each failure message shows the input with `%q`, so you can see the spaces,
`\t` tabs and `\n` newlines. Ask which `strings` function would handle that
input.

## Task 1: two functions

Trim the whitespace with `strings.TrimSpace`, then compare with
`strings.EqualFold(..., "production")` instead of `==`.

## Task 2: first "=" only

`strings.Split` cuts at every `"="`, so `parts[1]` is only the piece up to
the second one. `strings.Cut(line, "=")` returns the key, everything after the
first `"="`, and whether it found one.

## Task 3: one word to change

Replace `strings.Split(line, " ")` with `strings.Fields(line)`.
