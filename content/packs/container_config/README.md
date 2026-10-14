# Container config: settings from the environment

One image runs in dev, staging and production. What changes between them
(the port, the database, how much to log) comes from *environment
variables* set by `docker run -e` or a Kubernetes manifest. Here you read
them safely, check them at startup, and log them without leaking a password.

**New here:** twelve-factor config, `os.Getenv` next to the `os.LookupEnv`
you know from *Config precedence: defaults, file, env and flags*, passing a
lookup function in, `net/url`'s `url.Parse` and `Redacted`.

## Learn

### Config lives in the environment

The *twelve-factor* rule says: keep settings out of the code and out of the
image, and read them from the environment. Then the image you tested in
staging is byte for byte the one you ship.

```bash
docker run --rm -e REGION=eu-west-1 -e WORKERS=4 reportd:1.2.0
```

In Kubernetes, the same settings sit in the Deployment's container spec:

```yaml
env:
  - name: REGION
    value: eu-west-1
  - name: WORKERS
    value: "4"
```

Every value arrives as a string, even `"4"`. Turning text into numbers and
checking it is the program's job.

### Unset or empty?

Go has two ways to read a variable:

| Call | `REGION` unset | `REGION=` (set, empty) | `REGION=eu-west-1` |
|---|---|---|---|
| `os.Getenv("REGION")` | `""` | `""` | `"eu-west-1"` |
| `os.LookupEnv("REGION")` | `"", false` | `"", true` | `"eu-west-1", true` |

`os.Getenv` can't tell "nobody set it" from "someone set it to nothing".
`os.LookupEnv` can: its second result, usually called `ok`, says whether the
variable exists at all. That difference matters when you report a mistake:
"missing from the manifest" and "set, but the secret behind it is empty" are
fixed in different places.

### Passing the lookup in

A test can't safely change the real environment of the whole program. So the
functions you write don't call `os.LookupEnv` themselves. They take a
parameter that *works like it*:

```go
func Greeting(lookupEnv func(string) (string, bool)) string {
	name, ok := lookupEnv("USER")
	if !ok {
		return "hello, stranger"
	}
	return "hello, " + name
}
```

The type `func(string) (string, bool)` is exactly the type of
`os.LookupEnv`, so the real program passes it in by name:
`Greeting(os.LookupEnv)`. A test passes a *fake* that reads a map, as in
*Fakes and dependency injection*:

```go
users := map[string]string{"USER": "ana"}
fake := func(key string) (string, bool) {
	value, ok := users[key]
	return value, ok
}
Greeting(fake) // => "hello, ana"
```

> **Why is it written like this?**
> - **Go requires:** the parameter's type to match exactly. A function
>   returning only `string`, like `os.Getenv`, can't be passed where
>   `func(string) (string, bool)` is expected.
> - **Convention:** a function value is passed without `()`.
>   `os.LookupEnv` is the function itself; `os.LookupEnv("USER")` would call
>   it and pass its results.
> - **Author's choice:** the name `lookupEnv`, and taking a function rather
>   than an interface. One method is all we need, so a function is simpler.

### Fail fast, and say everything

A program with a bad setting should stop at startup with a clear message,
not run for an hour and fail on the first request. In Kubernetes, a
container that exits right away is restarted again and again, with a growing
wait between tries: the pod shows `CrashLoopBackOff`.

Each restart costs time. If the program reports only the first mistake, you
fix it, wait for the restart, and find the next one. Report every problem at
once with `errors.Join`, as in *Reporting many errors*:

```go
var errs []error
if workers < 1 {
	errs = append(errs, fmt.Errorf("WORKERS %d must be at least 1", workers))
}
// ... more checks ...
return errors.Join(errs...) // nil when errs is empty
```

Numbers come from `strconv.Atoi`, as in *Errors are values*: it returns an
error for anything that isn't a whole number, such as `"8O80"` (letter O)
or `"80.5"`.

### Keeping passwords out of logs

Services often log their config at startup, which is handy until the
database password lands in your log system. A URL can carry a password:

```text
https://deploy:s3cret@registry.example.com/v2/
        user   password  host
```

`url.Parse` from the `net/url` package splits a URL into its parts and
returns a `*url.URL`. Its `Redacted` method gives the URL back as a string
with the password replaced by `xxxxx`:

```go
u, err := url.Parse("https://deploy:s3cret@registry.example.com/v2/")
if err != nil {
	// not a valid URL
}
fmt.Println(u.Redacted()) // https://deploy:xxxxx@registry.example.com/v2/
```

A URL without a password comes back as it was. `url.Parse` fails on a
broken URL, such as a password with a bare `%` in it, and its error message
contains the *whole* URL, password included, so never print that error.

> **Why is it written like this?** `Redacted` only hides the password in the
> `user:password@` part. A `?password=...` query parameter stays visible.
> That's how `net/url` defines it (**Go's choice**), so keep passwords in the
> user part, or in a separate variable.

## Your task

The `inventory` service is crash-looping in production: someone typed
`APP_ENV=prod` in its manifest, and the logs say nothing useful. `main.go`
calls your functions with the real `os.LookupEnv` and is already written.
Your code goes in `config.go`.

Every function gets the environment as `lookupEnv`. Read settings only
through it: the tests set the real environment to junk, so a call to
`os.Getenv` or `os.LookupEnv` in your code shows up as a failure.

### 1. Read a setting with a default

**Goal:** finish `Lookup(lookupEnv, key, fallback)` so it returns the value
of `key`, or `fallback` when there is no real value.

- Remove spaces around the value (`strings.TrimSpace`).
- Return `fallback` when the variable is unset, empty, or only spaces.
- Otherwise return the trimmed value, even if it looks like nothing, such as
  `"0"`.

```go
// environment: REGION=eu-west-1, LOG_LEVEL= debug , WORKERS=0, ZONE=
Lookup(lookupEnv, "REGION", "us-east-1")  // => "eu-west-1"
Lookup(lookupEnv, "LOG_LEVEL", "info")    // => "debug"
Lookup(lookupEnv, "WORKERS", "4")         // => "0"
Lookup(lookupEnv, "ZONE", "a")            // => "a"
Lookup(lookupEnv, "TEAM", "platform")     // => "platform"
```

**Think it through:**
- Call `lookupEnv` the way the `Greeting` example does. Which of its two
  results do you need, now that empty and unset both mean "use the default"?
- Trim first, then decide.

### 2. Parse the port

**Goal:** finish `Port(lookupEnv)` so it returns the `PORT` setting as an
`int`.

- `PORT` unset or blank: `8080`.
- Not a whole number: return `0` and the error
  `PORT "<value>" is not a number`, with the trimmed value in double quotes.
- Below 1 or above 65535: return `0` and the error
  `PORT <number> is out of range 1-65535`.

```go
// PORT=9090      => 9090, nil
// (PORT unset)   => 8080, nil
// PORT=:8080     => 0, error: PORT ":8080" is not a number
// PORT=70000     => 0, error: PORT 70000 is out of range 1-65535
// PORT=0         => 0, error: PORT 0 is out of range 1-65535
```

**Think it through:**
- Your `Lookup` already handles the default and the spaces. What should it
  return for `PORT`, so `strconv.Atoi` turns it into `8080`?
- Which `fmt.Errorf` verb puts double quotes around a string?

### 3. Load the whole config

**Goal:** finish `Load(lookupEnv)` so it reads every setting into a
`Config`, or returns one error listing every problem.

| Variable | Field | Rule | Problem message |
|---|---|---|---|
| `PORT` | `Port` | as in task 2 | the error from `Port` |
| `APP_ENV` | `Env` | as in task 1, default `dev`; must then be exactly `dev`, `staging` or `production` | `APP_ENV "<value>" must be dev, staging or production` |
| `DATABASE_URL` | `DatabaseURL` | required, spaces trimmed | unset: `DATABASE_URL is not set`; empty or only spaces: `DATABASE_URL is set but empty` |
| `LOG_LEVEL` | `LogLevel` | as in task 1, default `info` | none: any value is fine |

- Check all of them, and join the problems with `errors.Join`, in the
  table's order.
- If there is any problem, return `Config{}` and the joined error.
- Otherwise return the filled `Config` and `nil`.

```go
// environment: PORT=8O80, APP_ENV=prod, DATABASE_URL=
cfg, err := Load(lookupEnv)
// cfg == Config{}, err prints:
// PORT "8O80" is not a number
// APP_ENV "prod" must be dev, staging or production
// DATABASE_URL is set but empty

// environment: DATABASE_URL=postgres://app:hunter2@db:5432/inventory
cfg, err = Load(lookupEnv)
// cfg == Config{Port: 8080, Env: "dev", DatabaseURL: "postgres://app:hunter2@db:5432/inventory", LogLevel: "info"}, err == nil
```

**Think it through:**
- Which two functions you wrote already do most of the work?
- `DATABASE_URL` has two different messages. Which one can tell "unset"
  from "set but empty": `Lookup`, or `lookupEnv` itself?
- Where do you keep problems until every setting has been checked?

### 4. Redact the password in logs

**Goal:** give `Config` a `String` method that returns every setting on one
line, with the database password hidden.

- The format is `port=<Port> env=<Env> log=<LogLevel> db=<DatabaseURL>`,
  one space between each pair.
- Show `DatabaseURL` through `url.Parse` and `Redacted`, so a password
  becomes `xxxxx` and a URL without one stays as it was.
- If `url.Parse` fails, write `db=invalid` instead of the URL.

Write the method under the ready-made `// String ...` comment at the bottom
of `config.go`, and add the `net/url` import. Because `Config` then has a
`String` method, `fmt.Println(cfg)` in `main.go` uses it, as in *Stringers*.

```go
Config{Port: 8080, Env: "production", LogLevel: "info",
	DatabaseURL: "postgres://app:hunter2@db:5432/inventory"}
// prints: port=8080 env=production log=info db=postgres://app:xxxxx@db:5432/inventory

Config{Port: 8080, Env: "dev", LogLevel: "info",
	DatabaseURL: "postgres://app:50%off@db:5432/inventory"}
// prints: port=8080 env=dev log=info db=invalid
```

**Think it through:**
- Which parts of the line are fixed text, and which come from fields?
- Why is printing the raw `DatabaseURL` when parsing fails not an option?

Then try it from the exercise folder:

```bash
APP_ENV=prod PORT=8O80 go run .
DATABASE_URL=postgres://app:hunter2@db:5432/inventory go run .
```

The first prints all three problems and exits with status 1, as the
crash-looping pod would. The second prints the config with `xxxxx` in place
of the password.

## Learn more

- [The Twelve-Factor App: Config](https://12factor.net/config)
- [Define Environment Variables for a Container (Kubernetes)](https://kubernetes.io/docs/tasks/inject-data-application/define-environment-variable-container/)
- [os.LookupEnv](https://pkg.go.dev/os#LookupEnv)
- [net/url: URL.Redacted](https://pkg.go.dev/net/url#URL.Redacted)
