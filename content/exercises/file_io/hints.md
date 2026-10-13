## Task 1: check for "not exist" first

`os.ReadFile(path)` gives you `data` and `err`. Check
`errors.Is(err, fs.ErrNotExist)` before the general `err != nil` check, so a
missing file becomes `DefaultBanner, nil` and everything else is returned.

## Task 2: the open-defer-scan pattern

```go
f, err := os.Open(path)
if err != nil {
	return nil, err
}
defer f.Close()

sc := bufio.NewScanner(f)
for sc.Scan() {
	line := strings.TrimSpace(sc.Text())
	// skip "" and lines starting with "#", append the rest
}
return hosts, sc.Err()
```

## Task 3: write, then return Close

```go
f, err := os.Create(path)
if err != nil {
	return err
}
for _, h := range hosts {
	if _, err := fmt.Fprintln(f, h); err != nil {
		f.Close()
		return err
	}
}
return f.Close()
```
