## Task 1: Name the day thresholds

`7` is the critical threshold and appears in `CertStatus` and `NextCheck`. `30` is the warning threshold and appears in `CertStatus` and `RenewalList`. Declare each once, then replace all four numbers.

### Closer

Put a `const ( ... )` block where `TODO(1)` is, with `criticalDays = 7` and `warningDays = 30`. Then `daysLeft <= 7` becomes `daysLeft <= criticalDays`, and so on.

### Answer

```go
// Alert levels, in days before a certificate expires.
const (
	criticalDays = 7
	warningDays  = 30
)
```

Then use `criticalDays` in `CertStatus` and `NextCheck`, and `warningDays` in `CertStatus` and `RenewalList`.

## Task 2: Name the check intervals

`3600 * time.Second` is one hour and `86400 * time.Second` is one day. The `time` package has `time.Hour`, and a day is 24 of those.

### Closer

`urgentInterval = time.Hour` and `normalInterval = 24 * time.Hour`, in a `const` block. `NextCheck` returns the constants directly.

### Answer

```go
// How often to check a certificate.
const (
	urgentInterval = time.Hour      // critical or expired
	normalInterval = 24 * time.Hour // everything else
)

func NextCheck(daysLeft int) time.Duration {
	if daysLeft <= criticalDays {
		return urgentInterval
	}
	return normalInterval
}
```

## Task 3: Roll out the new warning policy

Every place that used the warning threshold now reads it from one constant. Change the constant, and the whole package follows.

### Closer

Change the value of `warningDays`, and nothing else.

### Answer

```go
warningDays = 21
```
