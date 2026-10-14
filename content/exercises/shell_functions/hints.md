## Task 1: Write the log helpers

Two small functions, each with one `echo`. The message is the function's
`$1`. One of them sends its line to stderr.

### Closer

`log_info() { echo "[INFO] $1"; }` is the pattern. `log_error` is the same
shape with `[ERROR]` and `>&2` at the end of its `echo`.

### Answer

```bash
log_info() {
	echo "[INFO] $1"
}

log_error() {
	echo "[ERROR] $1" >&2
}
```

## Task 2: Check the ports

The function's last command decides its exit status, so make the last line
a test that's true only for 1 to 65535.

### Closer

`[[ "$port" -ge 1 && "$port" -le 65535 ]]`, after `local port="$1"`.

### Answer

```bash
valid_port() {
	local port="$1"
	[[ "$port" -ge 1 && "$port" -le 65535 ]]
}
```

## Task 3: Build the image references

The function prints the reference, and `$(image_ref ...)` captures it. Copy
both arguments into `local` variables first, so the loop's `name` and `tag`
stay untouched.

### Closer

`local name="$1"` and `local tag="$2"`, then
`echo "registry.example.com/$name:$tag"`.

### Answer

```bash
image_ref() {
	local name="$1"
	local tag="$2"
	echo "registry.example.com/$name:$tag"
}
```

## Task 4: Require the tools

`command -v "$1"` fails when the tool is missing. Throw its output away and
turn the test around with `!`. `exit`, not `return`, stops the whole script.

### Closer

`if ! command -v "$1" > /dev/null; then ... fi`, with an `echo ... >&2` and
`exit 1` inside.

### Answer

```bash
require() {
	if ! command -v "$1" > /dev/null; then
		echo "missing tool: $1" >&2
		exit 1
	fi
}
```
