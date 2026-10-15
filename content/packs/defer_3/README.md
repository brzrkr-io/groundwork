# defer: safe database migrations

A schema migration that fails halfway can leave a database broken. Databases
solve this with transactions, and Go code makes them safe with one deferred
call.

**New here:** transactions, a cleanup method that is safe to call twice.

## Learn

### Recap from *defer* and *defer: fix the log rotation job*

- `defer` runs a call when the function returns, on every return path.
- Defer the cleanup right after you acquire something.
- Put work that needs its own cleanup into its own function.

### What a transaction is

A *transaction* groups changes so they happen all together or not at all.

| Step | Meaning |
|---|---|
| `Begin` | start a group of changes |
| `Exec` | make one change inside the group |
| `Commit` | keep every change; the transaction is finished |
| `Rollback` | throw every change away; the transaction is finished |

Real Go database code, like the `database/sql` package, works this way.

### "Always defer the undo, then finish on success"

The trick: defer the *undo* straight after starting, and make the undo do
nothing once the job is finished. Here is the same idea with a temporary file:

```go
tmp := NewTempFile("report.tmp")
defer tmp.Discard() // deletes the file, unless it was already saved

if err := tmp.Write(data); err != nil {
	return err // Discard runs: the half-written file is deleted
}
return tmp.Save() // Save marks it finished, so Discard does nothing
```

For that to work, the undo method checks a "finished" flag first:

```go
func (f *TempFile) Discard() {
	if f.saved {
		return // already finished: nothing to undo
	}
	// ... delete the file ...
}
```

> **Why is it written like this?** In a function with no results, a bare
> `return` ends the function early. Checking the flag at the top (an *early
> return*) is convention: it keeps the main work unindented.

## Your task

Your migration tool applies schema changes such as `001_create_users`. The
`DB` and `Tx` types are fakes that append a line to `db.Events` for every
call:

| Call | Appends to `db.Events` | Fails when |
|---|---|---|
| `db.Begin()` | `begin` | never |
| `tx.Exec("002_add_email")` | `exec 002_add_email` | change is `db.Reject`: `exec 002_add_email: rejected` (wraps `ErrRejected`) |
| `tx.Commit()` | `commit` | `db.Conflict` is true: `commit: write conflict` (wraps `ErrConflict`) |
| `tx.Rollback()` | `rollback` | never (you write this one) |

`Exec` and `Commit` append their event even when they fail. `Commit` marks
the transaction finished by setting `tx.done = true`, even when it fails.

### 1. Make Rollback safe to call twice

**Goal:** finish `Rollback` so it does nothing on a finished transaction.

If `tx.done` is already true, it does nothing. Otherwise it appends
`rollback` (the line already there does this) and sets `tx.done = true`.

```go
tx := db.Begin()
tx.Commit()
tx.Rollback()
// db.Events: begin, commit

tx = db.Begin()
tx.Rollback()
tx.Rollback()
// db.Events: begin, rollback
```

**Think it through:**
- Which field says "finished"? Look at what `Commit` does to it.
- What must happen *before* the append line, and what *after* it?

### 2. Apply changes in one transaction

**Goal:** write `ApplyChanges(db, changes)` so all changes are committed together, or rolled back on the first failure.

1. Begin a transaction and defer its rollback.
2. `Exec` each change in order.
3. At the first failing change, return its error wrapped with `%w` and prefixed with `change <n>: `, where `n` counts from 1.
4. If every change works, return the result of `Commit` unchanged.

An empty list still begins and commits.

```go
db := &DB{Reject: "002_add_email"}
ApplyChanges(db, []string{"001_create_users", "002_add_email", "003_add_index"})
// => error "change 2: exec 002_add_email: rejected"
// db.Events: begin, exec 001_create_users, exec 002_add_email, rollback

db = &DB{Conflict: true}
ApplyChanges(db, []string{"001_create_users"})
// => error "commit: write conflict"
// db.Events: begin, exec 001_create_users, commit
```

On success the events end with `commit` and no `rollback` appears. That
relies on your task 1 `Rollback`.

**Think it through:**
- `Begin` is a method on `db` and gives you a `*Tx`. `Exec`, `Commit` and `Rollback` are methods on that `Tx`.
- `range` gives an index starting at 0. The message counts from 1: what do you add?
- Do you need to call `Rollback` by hand anywhere, or does the `defer` cover it?

### 3. Apply every batch in its own transaction

**Goal:** write `ApplyBatches(db, batches)` that applies every batch, keeps going after failures, and returns all failures joined.

- Apply each batch with `ApplyChanges`, so each batch gets its own transaction.
- Wrap each failure with `%w`, prefixed with `batch <n>: ` (counting from 1).
- Combine them with `errors.Join`, as in *Reporting many errors*.
- Return `nil` when every batch works.

```go
db := &DB{Reject: "002_add_email"}
ApplyBatches(db, [][]string{
	{"001_create_users"},
	{"002_add_email", "003_add_index"},
	{"004_add_role"},
})
// => error "batch 2: change 1: exec 002_add_email: rejected"
// db.Events: begin, exec 001_create_users, commit,
//            begin, exec 002_add_email, rollback,
//            begin, exec 004_add_role, commit
```

When several batches fail, the joined message has one line per failed batch:

```text
batch 1: commit: write conflict
batch 2: commit: write conflict
```

**Think it through:**
- Where do you collect the errors while the loop runs?
- What does `errors.Join` return when it gets no errors?

## Learn more

- [Go docs: Executing transactions](https://go.dev/doc/database/execute-transactions)
- [database/sql: Tx.Rollback](https://pkg.go.dev/database/sql#Tx.Rollback)
- [Effective Go: Defer](https://go.dev/doc/effective_go#defer)
