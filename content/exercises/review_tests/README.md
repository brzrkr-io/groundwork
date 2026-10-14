# Code review: catch the bugs in a PR

A teammate has opened a pull request (PR) for the backup cleanup job, and
you're the reviewer. Reading the code finds some bugs; writing tests finds
the rest, and proves them. The best review comment is often "here's a
failing test". This exercise is that review.

**New here:** reviewing against a spec, tests as review evidence.

## Learn

### What a reviewer checks first

A PR description says what the change **should** do. Your first job is to
check the code does exactly that, in every case the description promises.
For each rule, ask:

- **The boundary.** "Kept for 7 days": is a backup *exactly* 7 days old
  kept or deleted? Test just under, exactly on, and just over every limit.
- **Overrides.** "Pinned backups are never deleted": does *every* path
  through the code respect that, or only some?
- **The odd input.** What if a timestamp is in the future, a list is empty,
  a name is blank?

Bugs live where rules meet: one rule's code path skips another rule's
check.

### Recap: tests that catch bugs

As in *Table-driven tests*, each row is one case with the answer you want.
Rows can name only the fields they need: a missing `bool` field is
`false`.

```go
{name: "pinned, old", age: 90 * day, pinned: true, want: false},
```

`7*day + time.Hour` is 7 days and 1 hour, a `time.Duration`. A negative
age, like `-time.Hour`, means the backup was taken one hour **after**
`now`: in the future.

### Red, then green

When a test fails on the PR's code, you've found a bug: that's red. In a
real review you'd post the failing case as a comment. Here you also fix it
(task 4), and the test turns green, as in *Practice: reproduce a bug
report*.

### How Groundwork grades your test

`TestGrader` runs your test against the real code, where it must pass,
and against *mutants*, copies of the code with one bug each, where it must
fail. A mutant that slips through is named in the output, such as "your
tests didn't notice when a weekly backup 10 days old was deleted". While
the PR's bugs are still in the code, your new rows fail on the real code
and `TestGrader` says so first. That's expected until task 4.

## Your task

> **PR #482: Clean up old backups automatically**
>
> Adds `Expired(b Backup, now time.Time) bool`. The cleanup job deletes
> every backup for which it returns `true`. The rules:
>
> 1. **Pinned** backups are never deleted. Engineers pin backups during
>    an incident.
> 2. A backup **taken in the future** (clocks on different servers can
>    disagree) is never deleted.
> 3. **Daily** backups are kept for 7 days: deleted once they are more
>    than 7 days old. A backup exactly 7 days old is kept.
> 4. **Weekly** backups are kept for 28 days, the same way.

Read `realExpired` in `exercise.go`: that's the PR's code. Then write the
tests the PR should have had, in `TestExpired` in `exercise_test.go`. Each
row is name, `age`, `weekly`, `pinned` and `want` (true means "delete
it"). Pick any names.

### 1. Test the daily rule

**Goal:** add rows for daily backups 6 days old, exactly 7 days old, and
7 days and 1 hour old.

| `age` | `want` |
|---|---|
| `6 * day` | `false` |
| `7 * day` | `false` |
| `7*day + time.Hour` | `true` |

One of these fails on the PR's code. That's your first bug.

**Think it through:**
- Which rule in the PR description says what happens at exactly 7 days?
- Which line of `realExpired` handles daily backups? Which operator does it use?

### 2. Test the weekly rule

**Goal:** add rows for weekly backups 10 days old, exactly 28 days old, and
28 days and 1 hour old.

| `age` | `weekly` | `want` |
|---|---|---|
| `10 * day` | `true` | `false` |
| `28 * day` | `true` | `false` |
| `28*day + time.Hour` | `true` | `true` |

**Think it through:**
- Why test 10 days? Which bug would delete a weekly backup that young?

### 3. Test the overrides

**Goal:** add rows for a pinned daily backup 30 days old, a pinned weekly
backup 60 days old, and a daily backup taken 1 hour in the future. None of
them may be deleted.

| `age` | `weekly` | `pinned` | `want` |
|---|---|---|---|
| `30 * day` | | `true` | `false` |
| `60 * day` | `true` | `true` | `false` |
| `-time.Hour` | | | `false` |

One of these finds the second bug.

**Think it through:**
- Follow a pinned weekly backup through `realExpired`, line by line. Does it ever reach the pinned check?

### 4. Fix the PR

**Goal:** fix both bugs in `realExpired`, so every row passes and
`TestGrader` reports every mutant caught.

**Think it through:**
- Rule 1 overrides everything. Where must its check go so that every path passes it?
- Which operator keeps a backup that is exactly 7 days old?

## Learn more

- [Google Engineering Practices: What to look for in a code review](https://google.github.io/eng-practices/review/reviewer/looking-for.html)
- [Go wiki: Table-driven tests](https://go.dev/wiki/TableDrivenTests)
- [time.Duration](https://pkg.go.dev/time#Duration)
