## Task 1: Load the login banner

`os.ReadFile(path)` gives you the whole file as bytes plus an error. A missing file is not a real error here, so check for that case before the general error check.

### Closer

Use `errors.Is(err, fs.ErrNotExist)` to spot a missing file and return `DefaultBanner, nil`. For any other error return `"", err`. Convert the bytes with `string(data)`.

### Answer

```go
data, err := os.ReadFile(path)
if errors.Is(err, fs.ErrNotExist) {
	return DefaultBanner, nil
}
if err != nil {
	return "", err
}
return string(data), nil
```

## Task 2: Read the host list

Open the file, make sure it gets closed, and read it line by line. For each line, trim it, then decide whether to keep it.

### Closer

Use `os.Open`, then `defer f.Close()`, then a `bufio.NewScanner(f)` loop. Skip lines where `line == ""` or `strings.HasPrefix(line, "#")` with `continue`. Return `sc.Err()` at the end.

### Answer

```go
f, err := os.Open(path)
if err != nil {
	return nil, err
}
defer f.Close()

var hosts []string
sc := bufio.NewScanner(f)
for sc.Scan() {
	line := strings.TrimSpace(sc.Text())
	if line == "" || strings.HasPrefix(line, "#") {
		continue
	}
	hosts = append(hosts, line)
}
return hosts, sc.Err()
```

## Task 3: Write the host list

Create the file, write one host per line, then close it. `Close` can report a failed write, so its error matters too.

### Closer

`fmt.Fprintln(f, h)` writes `h` plus `"\n"`. If a write fails, close the file and return the write error. At the end, `return f.Close()` instead of ignoring it.

### Answer

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
