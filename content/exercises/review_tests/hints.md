## Task 1: Test the daily rule

Copy the shape of the `"fresh daily"` row. Daily is the default, so you only set `name`, `age` and `want`.

### Closer

Three rows, `age: 6 * day`, `age: 7 * day` and `age: 7*day + time.Hour`, with the `want` from the README's table. The exactly-7-days row fails: the PR uses `>=`.

### Answer

```go
{name: "daily, 6 days", age: 6 * day, want: false},
{name: "daily, exactly 7 days", age: 7 * day, want: false},
{name: "daily, 7 days 1 hour", age: 7*day + time.Hour, want: true},
```

## Task 2: Test the weekly rule

Same three-point pattern around 28 days, with `weekly: true` in each row.

### Closer

Rows for `10 * day`, `28 * day` and `28*day + time.Hour`, each with `weekly: true`. The 10-day row catches code that applies the daily limit to weekly backups.

### Answer

```go
{name: "weekly, 10 days", age: 10 * day, weekly: true, want: false},
{name: "weekly, exactly 28 days", age: 28 * day, weekly: true, want: false},
{name: "weekly, 28 days 1 hour", age: 28*day + time.Hour, weekly: true, want: true},
```

## Task 3: Test the overrides

Both pinned rows are old enough that they'd be deleted without the pin. The future backup has a negative age.

### Closer

Set `pinned: true` on the 30-day and 60-day rows, and `weekly: true` on the 60-day one. The future row is `age: -time.Hour`. The pinned weekly row fails: the PR returns from the weekly branch before it checks the pin.

### Answer

```go
{name: "pinned daily, 30 days", age: 30 * day, pinned: true, want: false},
{name: "pinned weekly, 60 days", age: 60 * day, weekly: true, pinned: true, want: false},
{name: "from the future", age: -time.Hour, want: false},
```

## Task 4: Fix the PR

Bug one: `>=` deletes a daily backup at exactly 7 days. Bug two: the pinned check comes after the weekly `return`, so pinned weekly backups never reach it.

### Closer

Move `if b.Pinned { return false }` to the top of the function, before anything else. Change `>=` to `>` on the daily line.

### Answer

```go
func realExpired(b Backup, now time.Time) bool {
	if b.Pinned {
		return false
	}
	age := now.Sub(b.Taken)
	if b.Weekly {
		return age > weeklyRetention
	}
	return age > dailyRetention
}
```

A future backup has a negative age, which is never more than the retention, so it's kept without a check of its own.
