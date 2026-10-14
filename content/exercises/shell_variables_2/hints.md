## Task 1: Fix the region assignment

With spaces around `=`, bash reads the line as a command named `region`
with two arguments. An assignment is one word.

### Closer

Remove the spaces on both sides of `=`: `name=value`, all touching.

### Answer

```bash
region="${1:-eu-west-1}"
```

## Task 2: Fix the restart message

The line is in single quotes, and single quotes keep everything as plain
text, `$` included.

### Closer

Swap the single quotes for double quotes, so bash swaps in `$service`.

### Answer

```bash
echo "Restarting $service on all nodes"
```

## Task 3: Fix the file name with a space

After `$1` becomes `testdata/release notes.txt`, bash splits it at the space
(step 2), so `<` gets two words and gives up. Quotes stop the split.

### Closer

Put double quotes around the variable after `<`: `< "$1"`.

### Answer

```bash
lines=$(wc -l < "$1")
```

## Task 4: Fix the backup name

Bash reads a variable name for as long as it sees letters, digits and `_`.
In `$service_$date` it reads `service_`, a variable that doesn't exist.

### Closer

Braces mark where the name ends: `${service}` followed by `_`.

### Answer

```bash
echo "${service}_$date.tar.gz"
```
