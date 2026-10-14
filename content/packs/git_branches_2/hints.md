## Task 1: Merge a branch with a merge commit

You're on `main` already. Merge the branch, and use the option that skips
the editor for the merge commit's message.

### Closer

`git merge --no-edit <branch>`.

### Answer

```bash
git merge --no-edit add-metrics
```

## Task 2: Start the conflicting merge

The merge will fail with exit code 1, and `set -e` would stop the script
there. You want it to stop the merge, not the script.

### Closer

`git merge raise-limits || true`: when the merge fails, `true` runs and the
line as a whole succeeds.

### Answer

```bash
git merge raise-limits || true
```

## Task 3: Resolve the conflict

Overwrite the conflicted file with the final version (no markers), then tell
Git it's resolved and commit the merge.

### Closer

`cat > config/app.yaml <<'EOF'` ... `EOF` writes the file. Then
`git add config/app.yaml` and `git commit --no-edit`.

### Answer

```bash
cat > config/app.yaml <<'EOF'
name: web
replicas: 2
timeout: 30s
max_connections: 200
log_level: info
EOF
git add config/app.yaml
git commit --no-edit
```
