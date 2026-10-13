## Task 1: one check per row

Follow the shape of the `PING` check. For `ECHO`, use
`strings.CutPrefix(line, "ECHO ")` (note the space) and return the text after
it. Anything that matched nothing falls through to the last `return`.

```go
if text, ok := strings.CutPrefix(line, "ECHO "); ok {
	return text, false
}
```

## Task 2: reply, then maybe stop

Put `defer conn.Close()` as the first line of `HandleConn`. Inside the loop,
call `Respond`, write its reply with `fmt.Fprintf`, and `return` if it said
to quit.

```go
reply, quit := Respond(scanner.Text())
fmt.Fprintf(conn, "%s\n", reply)
if quit {
	return
}
```

## Task 3: dial, deadline, write, read

The Learn section has every line you need, in order: `net.DialTimeout`,
`defer conn.Close()`, `conn.SetDeadline`, `fmt.Fprintf(conn, "%s\n", cmd)`,
`bufio.NewReader(conn).ReadString('\n')`, `strings.TrimSuffix`. Check the
error after each step that returns one.

## Task 3: almost the answer

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

If the test says "EOF", the server never got a full line: did you send the
`"\n"`?
