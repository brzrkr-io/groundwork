## Task 1: Write the port rule once

The three port blocks differ in only two things: the setting's name in the message and the field being checked. Those are `checkPort`'s two parameters.

### Closer

Inside `checkPort`, use `port` and `name` where the copies used `c.HTTPPort` and `"http_port"`, and return `fmt.Errorf(...)` instead of appending. In `Validate`, write `if err := checkPort("http_port", c.HTTPPort); err != nil { problems = append(problems, err.Error()) }` for each port.

### Answer

```go
// checkPort returns the problem with the port setting called name, or nil.
func checkPort(name string, port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("%s %d is out of range 1-65535", name, port)
	}
	if port < 1024 {
		return fmt.Errorf("%s %d needs root; use 1024 or above", name, port)
	}
	return nil
}
```

And in `Validate`, once per port:

```go
if err := checkPort("admin_port", c.AdminPort); err != nil {
	problems = append(problems, err.Error())
}
```

## Task 2: Write the environments once

One list, three uses: the check in `Validate`, the message in `Validate`, and the `switch` in `AlertChannel`. Each reads the list instead of spelling out the names.

### Closer

`var environments = []string{"development", "staging", "production"}`. The check is `!slices.Contains(environments, c.Environment)`; the message uses `strings.Join(environments, ", ")`. Add `"slices"` and `"strings"` to the imports.

### Answer

```go
// environments lists every environment a service can run in.
var environments = []string{"development", "staging", "production"}
```

In `Validate`:

```go
if !slices.Contains(environments, c.Environment) {
	problems = append(problems, fmt.Sprintf("environment %q is not one of %s", c.Environment, strings.Join(environments, ", ")))
}
```

And `AlertChannel`:

```go
if !slices.Contains(environments, env) {
	return "#alerts-unknown"
}
return "#alerts-" + env
```

## Task 3: Check the gRPC port

The rule already exists in one place. You only need to call it for one more field.

### Closer

Copy the three-line `if err := checkPort(...)` block for the admin port, and change the name and the field.

### Answer

```go
if err := checkPort("grpc_port", c.GRPCPort); err != nil {
	problems = append(problems, err.Error())
}
```
