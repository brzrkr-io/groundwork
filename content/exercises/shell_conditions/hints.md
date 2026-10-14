## Task 1: Check the config file

You need an `if` with a file test in `[[ ]]`, and an `else` that prints an
error and stops. Remember the spaces inside the brackets.

### Closer

`if [[ -f "$1" ]]; then ... else ... fi`. In the `else` part:
`echo "missing $1" >&2` and then `exit 1`.

### Answer

```bash
if [[ -f "$1" ]]; then
	echo "found $1"
else
	echo "missing $1" >&2
	exit 1
fi
```

## Task 2: Grade disk usage

Three outcomes means `if`, `elif` and `else`. Numbers need the letter
operators, and the highest threshold goes first.

### Closer

`[[ "$percent" -ge 90 ]]` is true for 90 and above. Use it in the `if`, the
same with 75 in the `elif`, and print `ok: $percent%` in the `else`.

### Answer

```bash
if [[ "$percent" -ge 90 ]]; then
	echo "critical: $percent%"
elif [[ "$percent" -ge 75 ]]; then
	echo "warning: $percent%"
else
	echo "ok: $percent%"
fi
```

## Task 3: Validate the environment

Two separate checks. First: is `env` empty? If so, print the usage line to
stderr and exit. Second: is it one of the two good values?

### Closer

`[[ -z "$env" ]]` is true for an empty string. For the second check,
`[[ "$env" == "production" || "$env" == "staging" ]]`, with an `else` for the
unknown case.

### Answer

```bash
if [[ -z "$env" ]]; then
	echo "usage: 3_check_env.sh ENV" >&2
	exit 2
fi

if [[ "$env" == "production" || "$env" == "staging" ]]; then
	echo "$env is a valid environment"
else
	echo "unknown environment: $env" >&2
	exit 2
fi
```

## Task 4: Look for a rollback

`if` takes any command as its condition. A quiet `grep` succeeds when it
finds the text, which is exactly the question.

### Closer

`if grep -q 'ROLLBACK' "$1"; then ... else ... fi`, without brackets.

### Answer

```bash
if grep -q 'ROLLBACK' "$1"; then
	echo "$1: rolled back"
else
	echo "$1: clean"
fi
```
