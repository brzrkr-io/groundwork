## Task 1: Fix the compile error in FQDN

Read the compiler error: `no new variables on left side of :=`. `name` is
created on the first line. What should the second line do instead of creating
it again?

### Closer

Use `:=` to create a variable and `=` to change one that already exists. Swap
the `:=` on the second line for `=`.

### Answer

```go
name := host
name = name + "." + domain
return name
```

## Task 2: Fix the log path bug

Run the tests and compare what you returned with what the test wanted. One
character is missing between the service name and the date.

### Closer

A slash is fixed text, so it goes inside quotes, `"/"`, joined with `+` between
`service` and `date`.

### Answer

```go
return "/var/log/" + service + "/" + date + ".log"
```

## Task 3: Finish the alert title

`title` already holds `[critical]`. You need to add a space, the service, a
colon and a space, then the summary, using `+=`.

### Closer

Fixed text goes inside quotes (`" "` and `": "`); parameters (`service`,
`summary`) don't. One statement is enough: `title += " " + service + ...`.

### Answer

```go
title += " " + service + ": " + summary
```
