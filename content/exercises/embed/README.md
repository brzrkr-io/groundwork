# Embedding files

A Go program compiles to a single binary. With `//go:embed`, files such as a
default config or message templates go *inside* that binary. You copy one
file into a container image and nothing can go missing at 3am.

## Learn

### The go:embed directive

Put a `//go:embed` line directly above a package-level variable. When the
program is built, Go reads the file and stores its contents in the variable:

```go
import _ "embed"

//go:embed motd.txt
var motd string
```

The path is relative to the folder of the `.go` file. Once the program is
built, it no longer needs `motd.txt` on disk.

The variable's type decides what you get:

| Type | You get | Use it for |
|---|---|---|
| `string` | the text of one file | a banner, a SQL query |
| `[]byte` | the raw bytes of one file | JSON to decode, a binary blob |
| `embed.FS` | a read-only file system of files and folders | templates, many files |

### The import

The directive only works in files that import the `embed` package. If you
use `embed.FS`, a normal `import "embed"` covers it. If you only embed into
a `string` or `[]byte`, nothing names the package, so you write a *blank
import*: `import _ "embed"`. The `_` means "load the package, but I won't
call it by name".

### The rules

- No space between `//` and `go:embed`. `// go:embed` is just a comment, and
  the variable stays empty without any error.
- The directive goes directly above the `var`, with nothing in between.
- Paths can't use `..` and can't leave the module.
- When you embed a folder, files starting with `.` or `_` are skipped.

### Embedding a folder

```go
//go:embed templates
var templates embed.FS
```

Paths inside an `embed.FS` always start with the folder name and use `/`:

```go
data, err := templates.ReadFile("templates/deploy.txt")
entries, err := templates.ReadDir("templates") // sorted by file name
for _, entry := range entries {
	fmt.Println(entry.Name()) // "deploy.txt", ...
}
```

`entry.Name()` is the file name only, without the folder.

### Templates from an embed.FS

`text/template` can load templates straight from any file system, including
an `embed.FS`:

```go
tmpl, err := template.ParseFS(templates, "templates/deploy.txt")
if err != nil {
	return "", err // for example, the file doesn't exist
}
var b strings.Builder
if err := tmpl.Execute(&b, data); err != nil {
	return "", err
}
return b.String(), nil
```

`strings.Builder` collects everything `Execute` writes, and `b.String()`
returns it as one string.

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

The tests first change to an empty folder (`t.Chdir`), so reading these
files from disk won't work. Only embedded files will.

### 1. Embed the default config

Add a `//go:embed defaults.json` directive above `var defaultConfig []byte`.
`DefaultConfig` already decodes it. `defaults.json` contains:

```json
{
  "channel": "#deploys",
  "username": "deploy-bot",
  "retries": 3
}
```

so `DefaultConfig()` should return
`Config{Channel: "#deploys", Username: "deploy-bot", Retries: 3}` and a `nil`
error. The file already imports `embed`, so you don't need a blank import.

### 2. Embed the templates folder

Add a `//go:embed templates` directive above `var templates embed.FS`. Then
finish `Kinds`. It lists the files in the `templates` folder and returns
their names without the `.txt` ending, sorted. The starter already lists the
folder with `templates.ReadDir`, but keeps the `.txt` ending:

```go
Kinds()
// => []string{"deploy", "failed", "rollback"}, nil
```

If reading the folder fails, it returns the error. Until the directive is in
place, `templates` is empty, so that is what you'll see.

### 3. Render a notification

Finish `Render(kind string, d Deploy) (string, error)`. It loads
`templates/<kind>.txt` from the embedded files, executes it with `d`, and
returns the text. `templates/deploy.txt` is:

```text
[deploy] {{.Service}} {{.Version}} is live in {{.Env}}.
Deployed by {{.User}}.
```

```go
Render("deploy", Deploy{Service: "payments", Version: "v1.4.2", Env: "production", User: "alice"})
// => "[deploy] payments v1.4.2 is live in production.\nDeployed by alice.\n"
```

The output ends with a newline, because the template file does. For an
unknown kind such as `"party"`, `Render` returns `""` and a non-nil error.
`template.ParseFS` already returns an error when no file matches.

Add `"strings"` and `"text/template"` to the imports when you need them.

## Learn more

- [Package embed](https://pkg.go.dev/embed)
- [template.ParseFS](https://pkg.go.dev/text/template#ParseFS)
- [Go by Example: Embed Directive](https://gobyexample.com/embed-directive)
