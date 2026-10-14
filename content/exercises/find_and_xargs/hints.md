## Task 1: List the log files

Start find at the folder in `$1`, keep only regular files whose name ends in
`.log`, and pipe the result into `sort`.

### Closer

`-type f` keeps files and `-name '*.log'` keeps names ending in `.log`.
Quote `"$1"` and the pattern: `find "$1" -type f -name ... | sort`.

### Answer

```bash
find "$1" -type f -name '*.log' | sort
```

## Task 2: List the old log files

It's task 1 with one more test, about age. Ages are whole days, and a `+`
in front of the number means "more than".

### Closer

`-mtime +7` passes for files modified more than 7 days ago. Add it after
the `-name` test.

### Answer

```bash
find "$1" -type f -name '*.log' -mtime +7 | sort
```

## Task 3: Count the config lines

Three steps in one pipeline: find lists the `.conf` files, xargs hands them
to `cat`, and `wc -l` counts the lines that come out. Watch the names with a
space.

### Closer

Make find end each path with a null byte (`-print0`) and tell xargs to split
there (`xargs -0`): `find "$1" -type f -name '*.conf' -print0 | xargs -0 ... | wc -l`.

### Answer

```bash
find "$1" -type f -name '*.conf' -print0 | xargs -0 cat | wc -l
```

## Task 4: Delete the old log files

Use the find command from task 2 without the `sort`, and let find delete
what it finds. Run it without the delete part first to check the list.

### Closer

`-delete` goes last, after all the tests, so it only runs on files that
passed them.

### Answer

```bash
find "$1" -type f -name '*.log' -mtime +7 -delete
```
