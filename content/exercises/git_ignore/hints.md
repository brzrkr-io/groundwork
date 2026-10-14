## Task 1: Ignore build output and logs

Two patterns, one per line: a folder pattern ends with `/`, and `*` matches
any characters in a file name. Write the file, then stage and commit it.

### Closer

`printf '%s\n' 'first' 'second' > .gitignore` writes two lines. The patterns
are `bin/` and `*.log`. Then `git add .gitignore` and `git commit -m "..."`.

### Answer

```bash
printf '%s\n' 'bin/' '*.log' > .gitignore
git add .gitignore
git commit -m "Ignore build output and logs"
```

## Task 2: Stop tracking the committed secret

Ignoring isn't enough for a file Git already tracks. Remove it from the
staging area only, add the ignore line, and commit.

### Closer

`git rm --cached .env` untracks the file. `echo '.env' >> .gitignore` adds
the rule. Stage `.gitignore` too before committing.

### Answer

```bash
git rm --cached .env
echo '.env' >> .gitignore
git add .gitignore
git commit -m "Stop tracking .env"
```

## Task 3: Ignore every env file except the example

A pattern for files ending in `.env`, then an exception for `example.env`.
The exception must come after the pattern.

### Closer

The lines are `*.env` and `!example.env`. Append them with `>>`, then stage
and commit `.gitignore`.

### Answer

```bash
printf '%s\n' '*.env' '!example.env' >> .gitignore
git add .gitignore
git commit -m "Ignore env files except the example"
```
