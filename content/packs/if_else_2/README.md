# if and else: fix the maintenance script

Your team's maintenance script labels backups, warns about TLS certificates
and picks a log level. A colleague started it, but it doesn't build yet and
one check gives the wrong answers.

**New here:** *scope* (where a variable can be used); practice with `if`.

This starter does not compile yet. Your first job is to read the compiler
error.

## Learn

### Recap

As in *if and else*, Go checks `if` / `else if` conditions from top to bottom
and runs only the first block that matches. A `return` inside a branch ends
the function early.

### Scope: where a name exists

A variable exists only inside the block where it was created. That area is
its *scope*. A variable made in an `if` short statement lives only in that
`if` / `else` chain:

```go
if n := len(hosts); n > 0 {
	fmt.Println(n) // fine: inside the if
}
fmt.Println(n) // compile error: undefined: n
```

To use a value both in the `if` and after it, create it on its own line
**before** the `if`. Then its scope is the whole rest of the function.

> **Why is it written like this?** Scope rules are what Go requires; the
> compiler enforces them. Keeping a helper inside the `if` is good style when
> nobody needs it afterwards, but it's your choice.

### Reading a compiler error

When code doesn't compile, Go prints the file, line, column and a message:

```text
./exercise.go:12:32: undefined: days
```

Line 12, column 32 uses a name `days` that Go can't see there. The line in
the error is where the name is **used**. The fix often belongs where the name
is **declared**. See *Reading compiler errors* for more.

## Your task

### 1. Fix the compile error in BackupAgeLabel

**Goal:** make `BackupAgeLabel` compile by declaring `days` on its own line
before the `if`, so the whole function can use it.

The starter creates `days` in the `if` short statement, then uses it after
the `if` has ended. The labels are already right; don't change them:

- under 24 hours: `"fresh"`
- otherwise: the number of whole days (`ageHours / 24`), then `"d old"`

```go
BackupAgeLabel(0)   // => "fresh"
BackupAgeLabel(23)  // => "fresh"
BackupAgeLabel(24)  // => "1d old"
BackupAgeLabel(50)  // => "2d old"
BackupAgeLabel(168) // => "7d old"
```

**Think it through:**
- The error points at the last line, but which line creates `days`?
- After you move the declaration, what's left for the `if` to hold?

### 2. Fix the order of the certificate checks

**Goal:** reorder the checks in `CertStatus` so the most urgent one comes
first.

`daysLeft` is the number of days until the certificate expires:

| `daysLeft` | Return |
|---|---|
| below 0 | `"expired"` |
| 0 to 6 | `"renew now"` |
| 7 to 29 | `"renew soon"` |
| 30 and above | `"valid"` |

The starter checks `daysLeft < 30` first. Any value below 7 is also below 30,
so that branch wins every time and the others never run.

```go
CertStatus(-30) // => "expired"
CertStatus(-1)  // => "expired"
CertStatus(0)   // => "renew now"
CertStatus(6)   // => "renew now"
CertStatus(7)   // => "renew soon"
CertStatus(29)  // => "renew soon"
CertStatus(30)  // => "valid"
```

**Think it through:**
- The three conditions and labels are correct. Only their order is wrong.
- Which condition matches the fewest values? That one goes first.

### 3. Finish LogLevelFor

**Goal:** make `LogLevelFor` return `"warn"` when `env` is exactly
`"production"` and `debug` is `false`.

The rules, in order:

1. If `debug` is `true`, return `"debug"`, whatever the environment.
2. Otherwise, if `env` is exactly `"production"`, return `"warn"`.
3. Otherwise, return `"info"`.

Rules 1 and 3 are already written. Add rule 2 between them.

```go
LogLevelFor("production", true)  // => "debug"
LogLevelFor("staging", true)     // => "debug"
LogLevelFor("production", false) // => "warn"
LogLevelFor("staging", false)    // => "info"
LogLevelFor("dev", false)        // => "info"
```

**Think it through:**
- `env` is a parameter, so it has no quotes. `production` is fixed text, so
  it does.
- Which operator checks "exactly equal"? (Not `=`.)

## Learn more

- [A Tour of Go: If with a short statement](https://go.dev/tour/flowcontrol/6)
- [Go spec: Declarations and scope](https://go.dev/ref/spec#Declarations_and_scope)
- [Go by Example: If/Else](https://gobyexample.com/if-else)
