## Task 1: Check every host

A `for` loop over the script's arguments prints one line per host. The
total comes after `done`, from the variable that counts arguments.

### Closer

`for host in "$@"; do ... done`, then `echo "hosts checked: $#"`.

### Answer

```bash
for host in "$@"; do
	echo "checking $host"
done
echo "hosts checked: $#"
```

## Task 2: Read the inventory

Read the file with `while IFS= read -r line`. Inside the loop, an `if`
decides whether to skip the line; everything else gets printed.

### Closer

`[[ -z "$line" || "$line" == "#"* ]]` is true for the lines to skip, and
`continue` skips them. Feed the file in with `done < "$1"`.

### Answer

```bash
while IFS= read -r line; do
	if [[ -z "$line" || "$line" == "#"* ]]; then
		continue
	fi
	echo "ssh deploy@$line uptime"
done < "$1"
```

## Task 3: Total the requests

`read` with two names splits each line: the service into the first, the
count into the second. Then add the count to `total` with arithmetic.

### Closer

`while read -r service count; do ... done < "$1"`, and inside:
`total=$((total + count))`.

### Answer

```bash
while read -r service count; do
	total=$((total + count))
done < "$1"
```

## Task 4: Count lines in each log

Loop over a glob of the `.log` files in `$1`. For each one, count its lines
as in *Shell variables* and strip the directory with `basename`.

### Closer

`for file in "$1"/*.log; do ... done`. Inside: `lines=$(wc -l < "$file")`,
and `$(basename "$file")` for the name.

### Answer

```bash
for file in "$1"/*.log; do
	lines=$(wc -l < "$file")
	echo "$(basename "$file"): $lines lines"
done
```
