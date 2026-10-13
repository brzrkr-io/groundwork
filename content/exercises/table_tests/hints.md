## One row per case

Each row is one test case: `{name, used, want}`. You don't touch the loop at
all. Add rows inside the `{ ... }` of `tests`, each ending with a comma, just
like `{"half full", 50, "ok"}`.

## Read TestGrader's message

If `TestGrader` says "your tests didn't notice when DiskAlert(80) said "ok"",
your table has no row with `used` set to `80`. Add one with the answer the
README gives.

## Tasks 1 and 2, nearly done

```go
{"warning", 85, "warning"},
{"critical", 95, "critical"},
{"just below warning", 79, "ok"},
{"warning starts", 80, "warning"},
```

Add `89` and `90` the same way.

## Task 3, nearly done

```go
{"empty", 0, "ok"},
{"negative", -1, "invalid"},
```

Then add `100` (a full disk is `"critical"`) and `101` (`"invalid"`).
