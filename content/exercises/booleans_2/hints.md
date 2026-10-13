## Read each line out loud

Say the code in words and compare it with the rule in the README. Where do
they differ: an "and" that reads as "or", or a "not" that is missing?

## Task 1: two problems in one rule

The rule says "and", so use `&&`. It also says "not acknowledged", which is
`!acknowledged`.

## Task 2 and 3: one operator each

"At most 500" is `<= 500`. For the window, an hour can never be both 22 or
more and below 2, so swap `&&` for `||`.

## Almost the answer

```go
return severity == "critical" && !acknowledged
return errorRate < 0.01 && latencyMs <= 500
return hour >= 22 || hour < 2
```
