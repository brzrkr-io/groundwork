## Task 1: Uptime against target

First add `import "fmt"`. Both numbers are floats with 2 decimals, each followed by a literal percent sign.

### Closer

`%.2f` gives 2 decimals and `%%` prints `%`. Use `%.2f%%` twice; the rest of the sentence is fixed text in the format string.

### Answer

```go
return fmt.Sprintf("%s: %.2f%% uptime (target %.2f%%)", service, uptime, target)
```

## Task 2: Disk alert

Pick a verb for each value: quotes for host and mount, one decimal for the used percentage, a whole number for the threshold.

### Closer

`%q` quotes host and mount, `%.1f%%` formats the used percentage and `%d%%` the threshold. Start with `"ALERT host=%q mount=%q ..."`.

### Answer

```go
return fmt.Sprintf("ALERT host=%q mount=%q used=%.1f%% threshold=%d%%", host, mount, usedPercent, thresholdPercent)
```

## Task 3: Pod table row

Build the format piece by piece, following the numbered list in the README. Width and precision can go together, as in *Width and precision together*.

### Closer

Name: `%-16s`. CPU: `%5.1f%%`. Memory: `%5d MiB`. Put single spaces between them.

### Answer

```go
return fmt.Sprintf("%-16s %5.1f%% %5d MiB", name, cpuPercent, memMiB)
```
