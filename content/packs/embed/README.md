# Embedding files

A Go program compiles to a single binary. With `//go:embed`, files such as a
default config or message templates go *inside* that binary. You copy one
file into a container image, and nothing can go missing at 3am.

**New here:** the `//go:embed` directive, `embed.FS`, `template.ParseFS`.
(Blank imports are from *database/sql: query deploys from Go*.)

## Learn

### The go:embed directive

A *directive* is a special comment that gives an instruction to the Go
tools. `//go:embed` says: "when you build this program, read this file and
store its contents in the variable on the next line".

```go
import _ "embed"

//go:embed motd.txt
var motd string
```

Once the program is built, it no longer needs `motd.txt` on disk. The text is
inside the binary, in `motd`.

> **Why is it written like this?**
> - **Go requires:** `//go:embed` with no space after `//`. It must sit
>   directly above a `var` at package level (outside every function). Only
>   ordinary `//` comment lines may come between them, never a blank line or
>   other code.
> - **Go requires:** the path is relative to the folder of the `.go` file. It
>   can't use `..`, so you can only embed files from this package's folder and
>   the folders inside it.
> - **Author's choice:** the variable name (`motd`) and the file name.

### What type to use

The variable's type decides what you get:

| Type | You get | Use it for |
|---|---|---|
| `string` | the text of one file | a banner, a SQL query |
| `[]byte` | the raw bytes of one file | JSON to decode, a binary blob |
| `embed.FS` | a read-only file system of files and folders | templates, many files |

### The import

The directive only works in a file that imports the `embed` package.

- If you use `embed.FS`, you already write `embed.` in your code, so a normal
  `import "embed"` covers it.
- If you only embed into a `string` or `[]byte`, nothing in your code says
  `embed.`. Go rejects imports you don't use, so you write a *blank import*:
  `import _ "embed"`. The `_` means "load the package, but I won't call it
  by name".

> **Why is it written like this?** The import is **required** by Go. The
> `_` is how Go lets you keep an import you never name in your code.

### The silent mistake

```go
// go:embed motd.txt
var motd string
```

With a space after `//`, this is an ordinary comment. Go doesn't complain,
and `motd` is just empty. If an embedded value is mysteriously empty, look
for that space first.

### Embedding a folder

Name a folder in the directive, and use `embed.FS`:

```go
//go:embed static
var static embed.FS
```

Files whose names start with `.` or `_` are skipped. Paths inside an
`embed.FS` always start with the folder name and use `/`, on every
operating system:

```go
data, err := static.ReadFile("static/index.html")
entries, err := static.ReadDir("static") // sorted by file name
for _, entry := range entries {
	fmt.Println(entry.Name()) // "index.html", ...
}
```

`entry.Name()` is only the file name, without the folder in front.

### Cutting an ending off a name

`strings.TrimSuffix(s, suffix)` returns `s` without `suffix` at the end. If
`s` doesn't end with it, you get `s` back unchanged:

```go
strings.TrimSuffix("backup.tar.gz", ".gz") // => "backup.tar"
```

### Templates from an embed.FS

`text/template` can load a template straight from any file system,
including an `embed.FS`. `Execute` fills it in and writes the result to a
writer. A `strings.Builder` is a writer that collects everything written to
it, and `b.String()` gives it back as one string:

```go
tmpl, err := template.ParseFS(static, "static/welcome.txt")
if err != nil {
	return "", err // for example, no file matched that path
}
var b strings.Builder
if err := tmpl.Execute(&b, user); err != nil {
	return "", err
}
return b.String(), nil
```

> **Why is it written like this?** `&b` is **required**: `Execute` must
> change `b`, so it needs a pointer to it, as in *Pointers*. Returning
> `"", err` the moment something fails is a Go **convention**.

## Your task

The deploy notifier ships its default settings and message templates inside
the binary. These files sit next to `exercise.go`:

```text
.
├── exercise.go
├── exercise_test.go
├── defaults.json
└── templates/
    ├── deploy.txt
    ├── failed.txt
    └── rollback.txt
```

The tests first move to an empty folder (`t.Chdir`), so reading these files
from disk won't work. Only embedded files will.

### 1. Embed the default config

**Goal:** fill the variable `defaultConfig` with the contents of
`defaults.json`, so `DefaultConfig()` returns the settings below.

`DefaultConfig` is already written: it decodes `defaultConfig` as JSON.
`defaults.json` contains:

```json
{
  "channel": "#deploys",
  "username": "deploy-bot",
  "retries": 3
}
```

```go
DefaultConfig()
// => Config{Channel: "#deploys", Username: "deploy-bot", Retries: 3}, nil
```

**Think it through:**
- Which line does the directive go above? Look for `TODO(1)`.
- The file already imports `embed`. Do you still need a blank import?

### 2. Embed the templates folder

**Goal:** fill `templates` with the `templates` folder, and make `Kinds`
return the template names without their `.txt` ending.

```go
Kinds()
// => []string{"deploy", "failed", "rollback"}, nil
```

The names come back sorted, because `ReadDir` sorts them. If reading the
folder fails, `Kinds` returns the error (the starter already does this).
Until the directive is in place, `templates` is empty, so `ReadDir` fails and
the tests show that error.

**Think it through:**
- This time the directive names a folder, not a file. Which variable has the
  type for that?
- The starter loop already appends `entry.Name()`. What should happen to
  each name before it's appended?

### 3. Render a notification

**Goal:** make `Render(kind, d)` load `templates/<kind>.txt` from the
embedded files, fill it in with `d`, and return the text.

`templates/deploy.txt` is:

```text
[deploy] {{.Service}} {{.Version}} is live in {{.Env}}.
Deployed by {{.User}}.
```

```go
Render("deploy", Deploy{Service: "payments", Version: "v1.4.2", Env: "production", User: "alice"})
// => "[deploy] payments v1.4.2 is live in production.\nDeployed by alice.\n", nil
```

- The result ends with a newline, because the template file does.
- For an unknown kind such as `"party"`, return `""` and a non-nil error.
  `template.ParseFS` already returns an error when no file matches the path.
- Add `"strings"` and `"text/template"` to the imports when you need them.

**Think it through:**
- The path has three parts: the fixed text `templates/`, the parameter
  `kind`, and the fixed text `.txt`. Which ones go in quotes?
- There are two steps that can fail (load, then execute). What do you return
  after each one fails?

## Learn more

- [Package embed](https://pkg.go.dev/embed)
- [template.ParseFS](https://pkg.go.dev/text/template#ParseFS)
- [Go by Example: Embed Directive](https://gobyexample.com/embed-directive)
