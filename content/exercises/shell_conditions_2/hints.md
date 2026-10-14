## Task 1: Fix the backup file check

Bash splits the line at spaces, and `[[-f` has none, so bash looks for a
command with that whole name. The test itself is right.

### Closer

Add a space after `[[` and another before `]]`.

### Answer

```bash
if [[ -f "$1" ]]; then
```

## Task 2: Fix the number comparison

Run `bash -x 2_failure_alert.sh 10` and look at the `[[ ... ]]` line. Is
bash comparing 10 and 5 as numbers, or as text?

### Closer

Inside `[[ ]]`, `>` compares text in dictionary order, where `10` sorts
before `5`. Numbers need a letter operator: `-gt` means "greater than".

### Answer

```bash
if [[ "$count" -gt 5 ]]; then
```

## Task 3: Fix the log check

`if` can run `grep -q` directly and use its exit status. The `[[ ]]` around
it turns grep's words into a broken test.

### Closer

Drop the `[[` and `]]`, and keep the `grep -q ...` and the `; then`.

### Answer

```bash
if grep -q 'backup complete' "$1"; then
```
