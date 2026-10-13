## Task 1: Add a row for each level

You don't touch the loop at all. Each row is one case, `{name, used, want}`, and you add rows inside the `{ ... }` of `tests`.

### Closer

Copy the shape of `{"half full", 50, "ok"},`, including the comma at the end. Change the name, the percentage and the answer.

### Answer

```go
{"warning", 85, "warning"},
{"critical", 95, "critical"},
```

## Task 2: Add the boundary rows

Bugs hide at the edges. For each limit (80 and 90), add the last value before it and the first value at it.

### Closer

You need rows for `79`, `80`, `89` and `90`. Check the README's table for the level each one should give.

### Answer

```go
{"just below warning", 79, "ok"},
{"warning starts", 80, "warning"},
{"just below critical", 89, "warning"},
{"critical starts", 90, "critical"},
```

## Task 3: Add the edges of the valid range

The valid range is 0 to 100. Test the two ends that are inside it and the two values just outside it.

### Closer

`0` is an empty disk and `100` a full one. `-1` and `101` are both `"invalid"`.

### Answer

```go
{"empty", 0, "ok"},
{"full", 100, "critical"},
{"negative", -1, "invalid"},
{"over 100", 101, "invalid"},
```
