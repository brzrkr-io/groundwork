# Embedding files: database migrations

Database tools usually ship their *migrations*: numbered SQL files that
build the schema step by step. Embedding them means the tool can never run
against a missing or half-copied migrations folder. This exercise fixes two
classic embed mistakes and then uses the files.

## Learn

### Recap

`//go:embed <path>` directly above a package-level variable stores files in
the binary when it is built: one file in a `string` or `[]byte`, or files and
folders in an `embed.FS`. Paths inside an `embed.FS` start with the folder
name, for example `"migrations/001_create_hosts.sql"`.

### Patterns

A directive can name files, folders or *glob patterns*, and can list several,
separated by spaces:

```go
//go:embed migrations/*.sql
var migrations embed.FS

//go:embed banner.txt templates
var assets embed.FS
```

| Pattern | Embeds |
|---|---|
| `migrations` | every file in the folder (except names starting with `.` or `_`) |
| `migrations/*.sql` | only the `.sql` files in that folder |
| `all:migrations` | every file, including `.` and `_` files |

`*` matches any run of characters within one name, as in a shell.

### The silent mistake

```go
// go:embed version.txt
var schemaVersion string
```

With a space after `//`, this is an ordinary comment. Go doesn't complain:
`schemaVersion` is just empty. If an embedded value is mysteriously empty,
check for that space first.

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

As in *Embedding files*, the tests switch to an empty folder first, so only
embedded files count.

### 1. Fix the version directive

`SchemaVersion` should return `"2.3.0"`, the contents of `version.txt` without
the trailing newline. The trimming is already written, but `SchemaVersion`
returns `""`. Find out why the file isn't embedded, and fix it.

### 2. Embed only the SQL files

`Names` is already written. It returns every file name in the embedded
`migrations` folder, and right now that includes `NOTES.txt`, which the
migrator would try to run as SQL. Change the directive above `migrations` so
it embeds only `.sql` files:

```go
Names()
// => []string{"001_create_hosts.sql", "002_create_services.sql", "003_add_owner.sql"}, nil
```

### 3. List the pending SQL

The database records how many migrations it has *applied* so far. Finish
`Pending(applied int) ([]string, error)`. It returns the SQL text of every
migration after the first `applied` ones, in file-name order:

```go
Pending(1)
// => []string{
//   "CREATE TABLE services (name TEXT PRIMARY KEY, port INTEGER NOT NULL);\n",
//   "ALTER TABLE services ADD COLUMN owner TEXT;\n",
// }, nil
```

- `Pending(0)` returns all three.
- When `applied` is 3 or more, nothing is pending: return an empty result
  (`nil` is fine) and a `nil` error.
- If reading the folder or a file fails, return the error.

Use `Names` for the order, then `migrations.ReadFile("migrations/" + name)`
for each file. `ReadFile` returns `[]byte`, so convert it with `string(...)`.

## Learn more

- [Package embed: Directives](https://pkg.go.dev/embed#hdr-Directives)
- [path.Match: glob syntax](https://pkg.go.dev/path#Match)
- [Go by Example: Embed Directive](https://gobyexample.com/embed-directive)
