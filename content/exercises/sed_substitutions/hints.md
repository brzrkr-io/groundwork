## Task 1: Bump the api image tag

The tag `1.4.1` appears twice, so replacing just `1.4.1` changes the
log-shipper too. Match something longer that only the api line has, like the
image name with its tag. The new tag comes from `$1`.

### Closer

Match `registry.example.com/api:1.4.1` and replace it with the same name plus
`$1`. Use `|` as the delimiter because of the slash, and double quotes so
bash swaps in `$1`: `sed "s|...|...$1|" testdata/deploy.yaml`.

### Answer

```bash
sed "s|registry.example.com/api:1.4.1|registry.example.com/api:$1|" testdata/deploy.yaml
```

## Task 2: Redact the secrets

In a secret line, the key ends with `_PASSWORD` or `_TOKEN`, so that text is
followed directly by `=`. Replace from there to the end of the line, and put
back the part you need to keep.

### Closer

With `-E`, `(_PASSWORD|_TOKEN)=.*` matches the end of the key, the `=` and
the whole value. The group remembers which ending it was, so the new text can
start with `\1`. The file name is `"$1"`.

### Answer

```bash
sed -E 's/(_PASSWORD|_TOKEN)=.*/\1=REDACTED/' "$1"
```

## Task 3: Strip comments and empty lines

Use the `d` command with a pattern in front of it, once for comments and once
for empty lines. Pin both patterns to the start of the line with `^`.

### Closer

`/^#/d` deletes lines that start with `#`. An empty line is a start followed
directly by an end: `^$`. Give each command its own `-e`.

### Answer

```bash
sed -e '/^#/d' -e '/^$/d' "$1"
```

`sed -n '/^[A-Z]/p' "$1"` works too: it prints only the lines that start
with a capital letter.

## Task 4: Change the log level in place

This is a substitute command with `-i`. Match the `LOG_LEVEL=` line from its
start, and replace the whole line with `LOG_LEVEL=` and the value in `$2`.
The file to edit is `$1`.

### Closer

`^LOG_LEVEL=.*` matches the whole setting line and nothing in the comment.
The script needs `$2`, so use double quotes:
`sed -i "s/^LOG_LEVEL=.*/LOG_LEVEL=.../" "$1"`.

### Answer

```bash
sed -i "s/^LOG_LEVEL=.*/LOG_LEVEL=$2/" "$1"
```
