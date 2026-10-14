## Task 1: Fix the undefined name

`URL`'s parameter is listed in its first line: `func URL(host string)`. The
body has to use exactly that name.

### Closer

Line 12 uses `hostname`, but the parameter is called `host`. Use the
parameter's name.

### Answer

```go
return "https://" + host
```

## Task 2: Fix the mismatched types

`func Version() string` means `Version` is a function. A function's name on
its own is the function itself, and you can't join that to text.

### Closer

To use what a function returns, call it: write its name followed by brackets
with its inputs. `Version` takes no inputs, so the brackets are empty.

### Answer

```go
return service + ":" + Version()
```

## Task 3: Fix the text-instead-of-number

`func Port() int` promises a whole number. `"8080"` has quotes, so it's text,
even though it contains digits.

### Closer

Numbers in Go have no quotes: `42`, `8080`. Remove the quotes from the
returned value.

### Answer

```go
return 8080
```

## Task 4: Fix the unused variable

`full` is created on line 30 but never used. The `return` gives back `"Hello"`
instead.

### Closer

Return the variable instead of the fixed text. Using `full` fixes the error
and returns the whole greeting.

### Answer

```go
full := "Hello, " + name
return full
```
