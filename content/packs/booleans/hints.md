## Task 1: Spot privileged ports

Every function here needs just one line: `return` followed by a check.
"Below 1024" means "less than 1024". Which operator from the table means
"less than"?

### Closer

Replace `false` with a comparison using `<`, like `port < ...`.

### Answer

```go
return port < 1024
```

The comparison is already `true` or `false`, so you can return it directly.

## Task 2: Decide whether a deploy may go out

Three things must all be true at once. Which operator means "and"? And how
do you say "no freeze active"?

### Closer

Join the conditions with `&&`. Flip a boolean with `!`: `!freezeActive` is
`true` when there is no freeze.

### Answer

```go
return testsPassed && approved && !freezeActive
```

## Task 3: Check work hours

This is three small checks joined together, like task 2: the hour is late
enough, the hour is early enough, and it's not the weekend.

### Closer

The three parts are `hour >= 9`, `hour < 17` and `!weekend`. Join them with
`&&`.

### Answer

```go
return hour >= 9 && hour < 17 && !weekend
```

`< 17` (not `<= 17`) leaves 17:00 out.
