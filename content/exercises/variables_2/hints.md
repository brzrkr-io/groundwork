## Task 1: create once, change after

`name` is created on the first line with `:=`. On the second line it already
exists, so that line should change it with `=` instead.

## Task 2: compare the output

The test prints what you returned and what it wanted. Look for the character
that is missing between the service name and the date.

## Task 3: one more +=

Add everything in a single statement:

```go
title += " " + service + ": " + summary
```

## Almost the answer

```go
// FQDN
name := host
name = name + "." + domain

// LogPath
return "/var/log/" + service + "/" + date + ".log"
```
