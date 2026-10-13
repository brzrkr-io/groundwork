## Which method fails a test?

Only `t.Errorf` and `t.Fatalf` (and their cousins `t.Error`, `t.Fatal`)
mark a test as failed. `t.Logf` just prints.

## Read the new failure

After task 1, `TestGrader` shows your test's output, something like:

```text
ShortSHA("9f86d081884c7d659a2feaa0c55ad015") = "9f86d08", want "9f86d081"
```

The code returned 7 characters. Count the characters in `want`.

## Task 3, nearly done

```go
if got := ShortSHA("a1b2c"); got != "a1b2c" {
	t.Errorf("ShortSHA(%q) = %q, want %q", "a1b2c", got, "a1b2c")
}
```
