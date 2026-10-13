## Task 1: remainder

The shard number is the remainder, which is what `%` gives you:
`userID % shards`.

## Task 2: multiply first

`errors / requests` is 0 whenever errors is smaller than requests, and
multiplying 0 by 100 is still 0. Multiply `errors` by 100 **before** you
divide.

## Task 3: add parentheses

Right now `usedGiB / totalGiB * 100` runs first, and the result is subtracted
from `totalGiB`. Put parentheses around the subtraction so it happens first.

## Almost the answer

```go
return errors * 100 / requests

return (totalGiB - usedGiB) / totalGiB * 100
```
