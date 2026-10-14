## Task 1: List names and IPs

Two steps: drop the header line, then keep the first two comma-separated
fields. Join them with a pipe.

### Closer

`tail -n +2 "$1"` prints everything from line 2. `cut -d, -f...` keeps the
fields you list, separated by commas: `-f1,2` is fields 1 and 2.

### Answer

```bash
tail -n +2 "$1" | cut -d, -f1,2
```

## Task 2: List the production hosts

awk can choose the lines and the column at once. The fields are split by
commas, the env is the third field, and the name the first.

### Closer

`awk -F, 'CONDITION {print $1}' "$1"`. The condition compares field 3 with
the text `"prod"` using `==`, which only matches the whole field.

### Answer

```bash
awk -F, '$3 == "prod" {print $1}' "$1"
```

## Task 3: Find the server errors

awk's default splitting (on spaces) works here. Count the fields of a log
line to find the path and the status, then pick the lines with a number
comparison.

### Closer

The path is `$7` and the status `$9`. The pattern is `$9 >= 500`, and
`print a, b` puts one space between the two.

### Answer

```bash
awk '$9 >= 500 {print $9, $7}' "$1"
```

## Task 4: Add up the bytes

One block runs on every line and adds the size to a variable; an `END`
block runs once and prints it with the fixed text in front.

### Closer

The size is `$10`. `{total += $10}` adds it up. In `END`, print the text in
double quotes followed by the variable: `END {print "..." total}`.

### Answer

```bash
awk '{total += $10} END {print "total bytes: " total}' "$1"
```
