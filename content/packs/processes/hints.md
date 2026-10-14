## Task 1: Find the runaway

Two steps: turn the exact name in `$1` into a PID, then ask `ps` for that
PID's PID and parent PID columns, without the header line.

### Closer

`pgrep -x "$1"` prints the PID. `ps -o pid=,ppid= -p PID` prints the two
columns for one PID. Keep the PID in a variable with `$(...)` in between.

### Answer

```bash
pid="$(pgrep -x "$1")"
ps -o pid=,ppid= -p "$pid"
```

`ps -o pid=,ppid= -C "$1"` does it in one go: `-C` selects by exact name.

## Task 2: Stop it politely, then by force

Send TERM first. Then loop 10 times: wait 0.1 seconds and check whether the
process is gone. If it is, print the TERM line and exit right away. If the
loop runs out, the process ignored TERM: send KILL and print the KILL line.

### Closer

`kill -0 "$pid"` succeeds while the process exists; `2> /dev/null` hides
its complaint once it's gone. Loop with `for _ in $(seq 10); do ... done`,
and inside it test with `if ! kill -0 "$pid" 2> /dev/null; then ... fi`.

### Answer

```bash
pid="$1"
kill -TERM "$pid"
for _ in $(seq 10); do
	sleep 0.1
	if ! kill -0 "$pid" 2> /dev/null; then
		echo "stopped by TERM (exit code 143)"
		exit 0
	fi
done
kill -KILL "$pid"
echo "killed by KILL (exit code 137)"
```

## Task 3: Look inside it

Everything is under `/proc/$1/`: `cwd` is a link to the working directory,
and `fd/` holds one link per open file. Print where each link points.

### Closer

`readlink "/proc/$pid/cwd"` prints the directory. Loop over the fd links
with a glob, `for fd in "/proc/$pid/fd/"*; do ... done`, and use `basename`
for the number and `readlink` for the target.

### Answer

```bash
pid="$1"
echo "cwd $(readlink "/proc/$pid/cwd")"
for fd in "/proc/$pid/fd/"*; do
	echo "$(basename "$fd") $(readlink "$fd")"
done
```

## Task 4: Lock down the secret

`chmod` with three octal digits sets an exact mode, so it doesn't matter
what the files had before. For the token, change the umask before you write
the file, not afterwards.

### Closer

`chmod 600 "$1/secret.env"` and the same with `755` for `deploy.sh`. Then
`umask 077`, and write the token with `echo "$2" > ...`.

### Answer

```bash
chmod 600 "$1/secret.env"
chmod 755 "$1/deploy.sh"
umask 077
echo "$2" > "$1/token"
```

## Task 5: Fix the cron job

Cron's working directory and `PATH` aren't yours. Fix both at the top of
the script: move to the folder the script is in, then put that folder's
`bin` on `PATH`, using an absolute path.

### Closer

`cd "$(dirname "$0")"` moves to the script's folder. After it, `$PWD` is
that folder as an absolute path: `PATH="$PWD/bin:$PATH"`. Then
`echo "using $(command -v healthcheck)" >&2`.

### Answer

```bash
cd "$(dirname "$0")"
PATH="$PWD/bin:$PATH"
echo "using $(command -v healthcheck)" >&2

healthcheck checks.conf
```
