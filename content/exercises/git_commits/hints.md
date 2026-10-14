## Task 1: Create the repository

Three commands: one creates the repo, two set config values. `git init` takes
an option that names the first branch.

### Closer

`git init -b <branch>` names the first branch. `git config <key> "<value>"`
sets a value for this repo; the keys are `user.name` and `user.email`.

### Answer

```bash
git init -b main
git config user.name "Sam Rivera"
git config user.email "sam@example.com"
```

## Task 2: Make the first commit

Two steps: stage the files, then commit what's staged with a message.

### Closer

`git add .` stages everything under the current folder (the script runs in
the sandbox). `git commit -m "..."` commits with a message.

### Answer

```bash
git add .
git commit -m "Add deploy script and app config"
```

## Task 3: Commit a change

First change the file with bash, then stage and commit it, as in task 2.
`>>` appends to a file; `>` would replace it.

### Closer

`echo "some line" >> path/to/file` adds one line at the end. Then
`git add config/app.yaml` and `git commit -m "..."`.

### Answer

```bash
echo "healthcheck: /healthz" >> config/app.yaml
git add config/app.yaml
git commit -m "Add health check path"
```
