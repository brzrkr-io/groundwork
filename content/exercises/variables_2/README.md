# Variables: fix the broken builders

Someone on your team wrote a few string builders for the log shipper and the
alert bot. One doesn't compile and one has a bug. Time to fix them.

## Learn

### Recap

Parameters are a function's named inputs. `name := value` creates a variable,
and `name = value` changes one that already exists. `+` joins strings, and you
add every separator yourself.

### Reading a compiler error

Go checks your code before it runs anything. If something is wrong, it prints
the file, line and column, then a message:

```text
./exercise.go:8:7: no new variables on left side of :=
```

This one is on line 8, column 7 of `exercise.go`. It means `:=` was used to
create a variable that already exists. While the code doesn't compile, **every
test shows ✗**, because no test can run. Fix the error and the tests run again.

### The `+=` shortcut

`x += y` is short for `x = x + y`. It is handy when you build a string up step
by step:

```go
msg := "deploy"
msg += " started" // msg is now "deploy started"
```

## Your task

Fix the log path and alert title helpers.

### 1. Fix the compile error in FQDN

The starter does not compile. Read the error, then fix `FQDN` so it compiles
and returns the host, a dot, then the domain. A host name written out in full
like this is a *fully qualified domain name* (FQDN).

```go
FQDN("web-01", "prod.example.com")
// => "web-01.prod.example.com"

FQDN("db-02", "staging.example.com")
// => "db-02.staging.example.com"
```

### 2. Fix the log path bug

`LogPath` should return `/var/log/`, the service name, a slash, the date, then
`.log`. Right now it returns the wrong thing. Run the tests, compare the
output with the expected value, and fix it.

```go
LogPath("nginx", "2026-10-11")
// => "/var/log/nginx/2026-10-11.log"

LogPath("api", "2026-01-31")
// => "/var/log/api/2026-01-31.log"
```

### 3. Finish the alert title

`AlertTitle` builds a title like `[critical] api: disk almost full`. It
already adds the severity in square brackets. Use `+=` to add a space, the
service, a colon and a space, then the summary.

```go
AlertTitle("critical", "api", "disk almost full")
// => "[critical] api: disk almost full"

AlertTitle("warning", "billing", "slow responses")
// => "[warning] billing: slow responses"
```

## Learn more

- [Go spec: Short variable declarations](https://go.dev/ref/spec#Short_variable_declarations)
- [Go spec: Assignment statements](https://go.dev/ref/spec#Assignment_statements)
- [Go by Example: Variables](https://gobyexample.com/variables)
