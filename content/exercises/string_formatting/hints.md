## Add the import

Below `package formatting`, add `import "fmt"`.

## Task 1: one verb per value

```go
return fmt.Sprintf("%s: %d replicas ready", service, replicas)
```

## Task 2: three verbs

`%q` adds the quotes for you, `%.1f` formats the number, and `%%` writes the
percent sign. The format string is `"rollout %q %.1f%% complete"`.

## Task 3: widths

`%-10s` pads a string on the right to 10 characters, and `%5d` pads a number
on the left to 5. The whole format is `"%-10s | %5d | %s"`.
