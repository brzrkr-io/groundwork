## Task 1: Announce the deploy

Two lines copy the arguments into named variables; one more line prints the
message. Remember: no spaces around `=`.

### Closer

`service="$1"` stores the first argument. Do the same for `env`, then
`echo "Deploying $service to ..."`.

### Answer

```bash
service="$1"
env="$2"
echo "Deploying $service to $env"
```

## Task 2: Build the image reference

The name is always `$1`. The tag needs a default: bash has a form of `${...}`
that falls back to a value when the argument is missing.

### Closer

`tag="${2:-latest}"` gives `$2`, or `latest` when there's no second argument.
Then print `registry.example.com/`, the name, `:` and the tag in one
double-quoted string.

### Answer

```bash
name="$1"
tag="${2:-latest}"
echo "registry.example.com/$name:$tag"
```

## Task 3: Count the hosts

`wc -l` counts lines. Feed it the file with `<` so it prints just the number,
and capture that number with `$(...)`.

### Closer

`count=$(wc -l < "$1")` stores the number. Then
`echo "$1: ... hosts"`.

### Answer

```bash
count=$(wc -l < "$1")
echo "$1: $count hosts"
```
