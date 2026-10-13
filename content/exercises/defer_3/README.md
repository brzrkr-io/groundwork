# defer: safe database migrations

A schema migration that fails halfway can leave a database broken. Databases
solve this with transactions, and Go code makes them safe with one line:
`defer tx.Rollback()`.

## Learn

Recap from the previous exercises: `defer` runs a call when the function
returns, on every return path. Defer the cleanup right after you acquire
something, and put work that needs its own cleanup into its own function.

A *transaction* groups changes so they happen all together or not at all.
You `Begin` it, run changes, then either `Commit` (keep everything) or
`Rollback` (throw everything away). Real Go database code, like the
`database/sql` package, uses this pattern:

```go
tx := db.Begin()
defer tx.Rollback() // undo everything if we return early

if err := tx.Exec("..."); err != nil {
	return err // the deferred Rollback throws the change away
}
return tx.Commit() // after Commit, the deferred Rollback does nothing
```

This works because `Rollback` does nothing on a transaction that is already
finished. So you can always defer it, and it only acts when you never reached
`Commit`. Even when `Commit` itself fails, the transaction counts as finished.

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

`Exec` and `Commit` append their event even when they fail. `Commit` already
marks the transaction finished by setting `tx.done = true`.

### 1. Make Rollback safe to call twice

Finish the `Rollback` method. If `tx.done` is already true, it does nothing.
Otherwise it appends `rollback` and sets `tx.done = true`.

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

### 2. Apply changes in one transaction

Write `ApplyChanges(db, changes)`. Begin a transaction, defer its rollback,
then `Exec` each change in order. At the first failing change, return its
error wrapped with `%w` and prefixed by `change <n>: `, where `n` counts from
1. If every change works, return the result of `Commit` unchanged. An empty
list still begins and commits.

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

With no problems, the events end with `commit`, and no `rollback` appears.
That relies on your `Rollback` from task 1, so finish that first.

### 3. Apply every batch in its own transaction

Write `ApplyBatches(db, batches)`. Apply each batch with `ApplyChanges`, so
each batch gets its own transaction. Keep going after a failed batch. Wrap
each failure with `%w` and prefix it with `batch <n>: ` (counting from 1),
then combine them with `errors.Join`, as in *Reporting many errors*. Return
`nil` when every batch works.

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

## Learn more

- [Go docs: Executing transactions](https://go.dev/doc/database/execute-transactions)
- [database/sql: Tx.Rollback](https://pkg.go.dev/database/sql#Tx.Rollback)
- [Effective Go: Defer](https://go.dev/doc/effective_go#defer)
