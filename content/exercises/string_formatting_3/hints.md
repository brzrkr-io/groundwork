## Task 1: two floats

Use `%.2f%%` twice: once for the uptime and once for the target.

## Task 2: mix the verbs

`%q` quotes host and mount, `%.1f%%` formats the used percentage and `%d%%`
the threshold.

## Task 3: build the format piece by piece

Name: `%-16s`. CPU: `%5.1f%%`. Memory: `%5d MiB`. Put single spaces between
them:

```go
fmt.Sprintf("%-16s %5.1f%% %5d MiB", name, cpuPercent, memMiB)
```
