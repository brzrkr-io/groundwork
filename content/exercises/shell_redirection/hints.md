## Task 1: Start the deploy log

Two `echo` lines, each redirected into `deploy.log`. The first must start
the file over; the second must keep what's there.

### Closer

`>` replaces the file, `>>` appends to it: `echo "..." > deploy.log`, then
`echo "..." >> deploy.log`.

### Answer

```bash
echo "deploy started" > deploy.log
echo "build ok" >> deploy.log
```

## Task 2: Split results and warnings

One command line, two redirections: one for stream 1, one for stream 2.

### Closer

`bash testdata/deploy.sh > FILE 2> OTHER_FILE`.

### Answer

```bash
bash testdata/deploy.sh > result.txt 2> warnings.txt
```

## Task 3: Put everything in one log

Point stdout at the file first. Then `2>&1` copies stdout's target to
stderr. The other order leaves stderr on the terminal.

### Closer

`bash testdata/deploy.sh > all.log 2>&1`, in exactly that order.

### Answer

```bash
bash testdata/deploy.sh > all.log 2>&1
```

## Task 4: Write the config with a here-doc

`cat` copies its input to its output. Redirect its output into `app.yaml`
and give it the three lines as a here-doc. Leave the marker unquoted so
`$service` and `$replicas` are filled in.

### Closer

```bash
cat > app.yaml <<EOF
service: $service
...
EOF
```

### Answer

```bash
cat > app.yaml <<EOF
service: $service
replicas: $replicas
image: registry.example.com/$service:stable
EOF
```
