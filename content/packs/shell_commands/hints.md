## Task 1: Print the deploy line

Without quotes, bash reads `(` as part of its own syntax and stops with a
syntax error. Give `echo` the whole line as one quoted argument.

### Closer

Either kind of quotes works here, because the line has no `$`:
`echo "..."` or `echo '...'`, with the exact text between them.

### Answer

```bash
echo "Deploying checkout (3 replicas)"
```

## Task 2: Print the reminder

The first line has a `$`, so double quotes would make bash look for a
variable named `IMAGE_TAG`. The second line holds double quotes of its own.
One kind of quotes handles both.

### Closer

Single quotes keep everything inside as plain text, `$` and `"` included:
`echo '...'`.

### Answer

```bash
echo 'Set $IMAGE_TAG before you deploy.'
echo 'Then run "make deploy".'
```

## Task 3: Print the status table

Each row is the same shape: a left column padded to 10 characters, a space,
the right column, a newline. That's one format string.

### Closer

`%-10s` pads to 10 characters, lined up on the left. Put the format in
single quotes: `printf '%-10s %s\n' SERVICE STATUS`, then the same for each
row.

### Answer

```bash
printf '%-10s %s\n' SERVICE STATUS
printf '%-10s %s\n' api running
printf '%-10s %s\n' worker stopped
printf '%-10s %s\n' payments running
```

One call works too, because `printf` repeats the format for leftover
arguments: `printf '%-10s %s\n' SERVICE STATUS api running ...`.
