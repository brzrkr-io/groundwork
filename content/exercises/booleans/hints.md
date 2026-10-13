## One line each

Every function needs just one line: `return` followed by a boolean
expression. Delete `false` and write the check in its place.

## Task 1: a single comparison

"Below 1024" means less than 1024. Use the `<` operator: `port < 1024`.

## Task 2: all three must hold

Join the three conditions with `&&`. "No freeze active" is `!freezeActive`.

## Almost the answer

```go
return testsPassed && approved && !freezeActive
```

For `IsWorkHours`, join three parts the same way:
`hour >= 9`, `hour < 17` and `!weekend`.
