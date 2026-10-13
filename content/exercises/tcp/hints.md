## Task 1: Answer one command

Follow the shape of the `PING` check: one `if` per row of the table. Anything
that matched nothing falls through to the last `return`.

### Closer

For `ECHO`, use `strings.CutPrefix(line, "ECHO ")` (note the space) and return
the text after it. `QUIT` is an exact match that returns `true` as the second
value.

### Answer

```go
if text, ok := strings.CutPrefix(line, "ECHO "); ok {
	return text, false
}
if line == "QUIT" {
	return "BYE", true
}
return "ERR unknown command", false
```

## Task 2: Serve one connection

Close the connection however the function returns, so put that first. Inside
the loop, ask `Respond` for the reply instead of echoing the line.

### Closer

`defer conn.Close()` at the top. Then `reply, quit := Respond(scanner.Text())`,
write `reply` with `"\n"`, and `return` if `quit` is true.

### Answer

```go
defer conn.Close()
scanner := bufio.NewScanner(conn)
for scanner.Scan() {
	reply, quit := Respond(scanner.Text())
	fmt.Fprintf(conn, "%s\n", reply)
	if quit {
		return
	}
}
```

## Task 3: Send a command as a client

The Learn section has every line you need, in order: dial, close, set a
deadline, write, read one line, trim. Check the error after each step that
returns one.

### Closer

`net.DialTimeout("tcp", addr, 2*time.Second)`, `defer conn.Close()`,
`conn.SetDeadline(time.Now().Add(2 * time.Second))`,
`fmt.Fprintf(conn, "%s\n", cmd)`, then
`bufio.NewReader(conn).ReadString('\n')` and `strings.TrimSuffix`. If the test
says "EOF", check that you sent the `"\n"`.

### Answer

```go
conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
if err != nil {
	return "", err
}
defer conn.Close()

if err := conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
	return "", err
}
if _, err := fmt.Fprintf(conn, "%s\n", cmd); err != nil {
	return "", err
}
line, err := bufio.NewReader(conn).ReadString('\n')
if err != nil {
	return "", err
}
return strings.TrimSuffix(line, "\n"), nil
```
