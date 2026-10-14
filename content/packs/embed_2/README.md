# Embedding files: database migrations

Database tools usually ship their *migrations*: numbered SQL files that
build the schema step by step. Embedding them means the tool can never run
against a missing or half-copied migrations folder. You'll fix two classic
embed mistakes, then use the files.

**New here:** glob patterns in `//go:embed`; finding a broken directive.

## Learn

### Recap

As in *Embedding files*: `//go:embed <path>` directly above a package-level
`var` stores files in the binary when it's built. One file goes in a
`string` or `[]byte`; files and folders go in an `embed.FS`. Paths inside an
`embed.FS` start with the folder name: `"static/index.html"`.

### Patterns

A directive can name files, folders or *glob patterns*. It can also list
several, separated by spaces:

```go
//go:embed static/*.html
var pages embed.FS

//go:embed banner.txt static
var assets embed.FS
```

| Pattern | Embeds |
|---|---|
| `static` | every file in the folder, except names starting with `.` or `_` |
| `static/*.html` | only the files in that folder ending in `.html` |
| `all:static` | every file, including names starting with `.` or `_` |

`*` matches any run of characters inside one name, as in a shell.

> **Why is it written like this?**
> - **Go requires:** paths use `/`, are relative to the `.go` file's folder,
>   and can't contain `..`. Every pattern must match at least one file, or
>   the build fails.
> - **Go requires:** the directive is a comment that starts exactly with
>   `//go:embed`. One space after `//` turns it into a plain comment.
> - **Author's choice:** which pattern you use. A narrow pattern keeps stray
>   files (notes, backups) out of the binary.

### The silent mistake

```go
// go:embed banner.txt
var banner string
```

This looks like a directive, but the space after `//` makes it an ordinary
comment. Go builds without any error, and `banner` is just empty. When an
embedded value is mysteriously empty, look for that space first.

### Skipping the first few items in a loop

`for i, name := range names` gives you each position `i` (starting at 0)
and each value. `continue` jumps straight to the next round:

```go
for i, line := range lines {
	if i < 2 {
		continue // skip the two header lines
	}
	fmt.Println(line)
}
```

`ReadFile` returns `[]byte`. `string(data)` turns it into a `string`.

## Your task

The `migrate` package embeds the inventory database's migrations. These files
sit next to `exercise.go`:

```text
.
├── version.txt                     2.3.0
└── migrations/
    ├── 001_create_hosts.sql
    ├── 002_create_services.sql
    ├── 003_add_owner.sql
    └── NOTES.txt                   notes for humans, not SQL
```

As in *Embedding files*, the tests move to an empty folder first, so only
embedded files count.

### 1. Fix the version directive

**Goal:** make `SchemaVersion()` return `"2.3.0"` by fixing the broken
directive above `schemaVersion`.

```go
SchemaVersion() // => "2.3.0"
```

Right now it returns `""`. The trimming of the trailing newline is already
written. Only the directive is wrong.

**Think it through:**
- Compare the comment above `schemaVersion` with the one above `migrations`,
  character by character.
- Leave the `TODO` comment line alone or delete it; only the directive line
  matters.

### 2. Embed only the SQL files

**Goal:** change the directive above `migrations` so it embeds only the
`.sql` files, not `NOTES.txt`.

`Names` is already written: it returns every file name in the embedded
`migrations` folder. Today that includes `NOTES.txt`, which the migrator
would try to run as SQL.

```go
Names()
// => []string{"001_create_hosts.sql", "002_create_services.sql", "003_add_owner.sql"}, nil
```

**Think it through:**
- Which row of the patterns table picks files by their ending?
- You change one line: the directive. `Names` stays as it is.

### 3. List the pending SQL

**Goal:** finish `Pending(applied)` so it returns the SQL text of every
migration after the first `applied` ones, in file-name order.

The database records how many migrations it has *applied* so far.

```go
Pending(1)
// => []string{
//   "CREATE TABLE services (name TEXT PRIMARY KEY, port INTEGER NOT NULL);\n",
//   "ALTER TABLE services ADD COLUMN owner TEXT;\n",
// }, nil
```

- Each string is the whole file, including its final newline.
- `Pending(0)` returns all three.
- When `applied` is 3 or more, nothing is pending: return an empty result
  (`nil` is fine) and a `nil` error.
- If reading the folder or a file fails, return the error.

**Think it through:**
- Which existing function already gives you the names in the right order?
- `Names` returns bare file names. What must you put in front of a name to
  get its path inside the `embed.FS`?

## Learn more

- [Package embed: Directives](https://pkg.go.dev/embed#hdr-Directives)
- [path.Match: glob syntax](https://pkg.go.dev/path#Match)
- [Go by Example: Embed Directive](https://gobyexample.com/embed-directive)
