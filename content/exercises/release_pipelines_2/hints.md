## Task 1: Stop leaking the deploy token

Two problems: the token's value is written in the file, and one step
prints it. The value should come from the repository's secrets, and the
printing step has no job to do.

### Closer

Replace the value after `DEPLOY_TOKEN:` with an expression that reads the
secret of the same name: `${{ secrets.... }}`. Delete the whole
`Show which token we use` step.

### Answer

```yaml
    env:
      DEPLOY_TOKEN: ${{ secrets.DEPLOY_TOKEN }}
    steps:
      - uses: actions/checkout@v7
      - run: ./deploy.sh production
```

## Task 2: Grant only what each job needs

Start every job from "read the code". Then look for the one job that
writes something, and give only that job more.

### Closer

Replace `write-all` with a block holding `contents: read`. Under the
`image` job, beside `runs-on`, add its own `permissions:` block with two
entries; it replaces the top one for that job.

### Answer

```yaml
permissions:
  contents: read
```

```yaml
  image:
    needs: test
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
```

## Task 3: Don't deploy pull requests

The `deploy` job runs whenever `test` passes, whatever the event. Which two
facts are true only for a push to `main`? Look at the first row of the
task's table.

### Closer

`if:` goes on the job, beside `needs` and `runs-on`. It joins two
comparisons with `&&`: `github.event_name == '...'` and
`github.ref == '...'`. Check for `push` itself, not "not a pull request":
`pull_request_target` isn't `pull_request` either.

### Answer

```yaml
  deploy:
    needs: test
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
```

## Task 4: Deploy exactly what you tested

`:latest` appears twice: where the image is pushed and where it is
deployed. Both need a tag that only this run's commit can have.

### Closer

Replace `latest` with `${{ github.sha }}` in the build step's `tags`, and
in the deploy command.

### Answer

```yaml
          tags: ghcr.io/acme/statusd:${{ github.sha }}
```

```yaml
      - run: ./deploy.sh production ghcr.io/acme/statusd:${{ github.sha }}
```
