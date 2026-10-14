## Task 1: Dispatch the commands

The usage and help branches are done. What's left is the end of the `Run`
from *Subcommands: a command table and generated help*: look `name` up in
`commands`, with two more parameters passed along.

### Closer

`cmd, ok := commands[name]` gives you the command and whether it exists.
`args[1:]` is everything after the command name. Call it with
`cmd.run(args[1:], stdin, stdout, stderr, lookup)`.

### Answer

```go
cmd, ok := commands[name]
if !ok {
	fmt.Fprintf(stderr, "preflight: unknown command %q\n", name)
	writeUsage(stderr)
	return 2
}
return cmd.run(args[1:], stdin, stdout, stderr, lookup)
```

## Task 2: Validate the settings

Three checks, one per setting, in the order of the table. Return as soon as
one fails; reach the end and return `nil`.

### Closer

A `switch` with an empty `case` accepts the good words, and `default`
reports the rest:

```go
switch s.Env {
case "dev", "staging", "prod":
default:
	return fmt.Errorf("env must be ..., got %q", s.Env)
}
```

The empty-file message never changes, so `errors.New` is enough for it.

### Answer

```go
switch s.Env {
case "dev", "staging", "prod":
default:
	return fmt.Errorf("env must be dev, staging or prod, got %q", s.Env)
}
if s.HostsFile == "" {
	return errors.New("hosts_file must not be empty")
}
if s.MaxHosts < 1 {
	return fmt.Errorf("max_hosts must be at least 1, got %d", s.MaxHosts)
}
return nil
```

## Task 3: Write the init command

Copy the flag handling from `runVersion` and add two flags. Then check
whether the file exists, build the JSON and hand it to `WriteFileAtomic`.

### Closer

```go
path := fs.String("config", "preflight.json", "config file to write")
force := fs.Bool("force", false, "replace an existing config file")
```

`os.Stat(*path)` returns a `nil` error when the file exists. The struct tags
on `Settings` already name the JSON keys, so
`json.MarshalIndent(Defaults(), "", "  ")` gives the right text; add the
newline with `append(data, '\n')`.

### Answer

```go
fs := flag.NewFlagSet("init", flag.ContinueOnError)
fs.SetOutput(stderr)
path := fs.String("config", "preflight.json", "config file to write")
force := fs.Bool("force", false, "replace an existing config file")
if err := fs.Parse(args); err != nil {
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	return 2
}
if fs.NArg() != 0 {
	fmt.Fprintln(stderr, "preflight: init takes no arguments")
	return 2
}

if _, err := os.Stat(*path); err == nil && !*force {
	fmt.Fprintf(stderr, "preflight: %s already exists; use -force to replace it\n", *path)
	return 1
}
data, err := json.MarshalIndent(Defaults(), "", "  ")
if err != nil {
	fmt.Fprintf(stderr, "preflight: %v\n", err)
	return 1
}
if err := WriteFileAtomic(*path, append(data, '\n')); err != nil {
	fmt.Fprintf(stderr, "preflight: %v\n", err)
	return 1
}
fmt.Fprintf(stdout, "wrote %s\n", *path)
return 0
```

## Task 4: Check the host list

Scan `r` with a `bufio.Scanner`, counting every line. Skip blanks and
comments, count the host, then check it against the three rules in order.
The empty and too-many checks come after the loop.

### Closer

A `switch` with no value runs the first `case` that is true, so each line
gets one message at most:

```go
switch {
case strings.ContainsAny(host, " \t"):
	problems = append(problems, fmt.Sprintf("line %d: invalid host %q", lineNo, host))
case !strings.HasSuffix(host, "."+env):
	// ...
case seen[host]:
	// ...
default:
	seen[host] = true
}
```

`seen` is a `map[string]bool`. Check `sc.Err()` before the final messages.

### Answer

```go
var problems []string
seen := map[string]bool{}
count := 0
lineNo := 0
sc := bufio.NewScanner(r)
for sc.Scan() {
	lineNo++
	host := strings.TrimSpace(sc.Text())
	if host == "" || strings.HasPrefix(host, "#") {
		continue
	}
	count++
	switch {
	case strings.ContainsAny(host, " \t"):
		problems = append(problems, fmt.Sprintf("line %d: invalid host %q", lineNo, host))
	case !strings.HasSuffix(host, "."+env):
		problems = append(problems, fmt.Sprintf("line %d: host %q is not in %s", lineNo, host, env))
	case seen[host]:
		problems = append(problems, fmt.Sprintf("line %d: duplicate host %q", lineNo, host))
	default:
		seen[host] = true
	}
}
if err := sc.Err(); err != nil {
	return nil, err
}
if count == 0 {
	problems = append(problems, "no hosts")
} else if count > maxHosts {
	problems = append(problems, fmt.Sprintf("%d hosts, more than max_hosts %d", count, maxHosts))
}
return problems, nil
```

## Task 5: Load the config layers

First declare `fileSettings`, then fill in `Load`: defaults, then the file
if it exists, then the environment. It's the `fileSettings`, `ApplyFile` and
`ApplyEnv` from *Config precedence: defaults, file, env and flags*, in one
function.

### Closer

Each `fileSettings` field is a pointer with the same JSON tag as in
`Settings`, such as ``Env *string `json:"env"` ``. Then:

```go
data, err := os.ReadFile(path)
if err != nil && !errors.Is(err, fs.ErrNotExist) {
	return Settings{}, err
}
if err == nil {
	var f fileSettings
	// json.Unmarshal, then one "if f.X != nil" per field
}
```

Wrap the JSON error with `fmt.Errorf("%s: %w", path, err)`. For the
environment, write one `if v, ok := lookup("PREFLIGHT_..."); ok` per variable.

### Answer

```go
// fileSettings mirrors the JSON config file. A field is nil when its key is
// missing from the file.
type fileSettings struct {
	Env       *string `json:"env"`
	HostsFile *string `json:"hosts_file"`
	MaxHosts  *int    `json:"max_hosts"`
}
```

```go
// in Load
s := Defaults()

data, err := os.ReadFile(path)
if err != nil && !errors.Is(err, fs.ErrNotExist) {
	return Settings{}, err
}
if err == nil {
	var f fileSettings
	if err := json.Unmarshal(data, &f); err != nil {
		return Settings{}, fmt.Errorf("%s: %w", path, err)
	}
	if f.Env != nil {
		s.Env = *f.Env
	}
	if f.HostsFile != nil {
		s.HostsFile = *f.HostsFile
	}
	if f.MaxHosts != nil {
		s.MaxHosts = *f.MaxHosts
	}
}

if v, ok := lookup("PREFLIGHT_ENV"); ok {
	s.Env = v
}
if v, ok := lookup("PREFLIGHT_HOSTS_FILE"); ok {
	s.HostsFile = v
}
if v, ok := lookup("PREFLIGHT_MAX_HOSTS"); ok {
	n, err := strconv.Atoi(v)
	if err != nil {
		return Settings{}, fmt.Errorf("invalid PREFLIGHT_MAX_HOSTS %q: %w", v, err)
	}
	s.MaxHosts = n
}
return s, nil
```

## Task 6: Write the check command

Follow the six steps in the README, one block each. Every error path prints
`preflight: ` and the error to stderr and returns `2`; only "problems found"
returns `1`.

### Closer

`fs.Visit(func(f *flag.Flag) { switch f.Name { case "env": s.Env = cli.env ... } })`
applies only the typed flags. For the host list, start with
`in := stdin` and replace it with the opened file unless
`s.HostsFile == "-"`, as in the Learn section.

### Answer

```go
s, err := Load(cli.config, lookup)
if err != nil {
	fmt.Fprintf(stderr, "preflight: %v\n", err)
	return 2
}
fs.Visit(func(f *flag.Flag) {
	switch f.Name {
	case "env":
		s.Env = cli.env
	case "hosts":
		s.HostsFile = cli.hosts
	case "max-hosts":
		s.MaxHosts = cli.maxHosts
	}
})
if err := s.Validate(); err != nil {
	fmt.Fprintf(stderr, "preflight: %v\n", err)
	return 2
}

in := stdin
if s.HostsFile != "-" {
	f, err := os.Open(s.HostsFile)
	if err != nil {
		fmt.Fprintf(stderr, "preflight: %v\n", err)
		return 2
	}
	defer f.Close()
	in = f
}
problems, err := CheckHosts(in, s.Env, s.MaxHosts)
if err != nil {
	fmt.Fprintf(stderr, "preflight: reading hosts: %v\n", err)
	return 2
}

for _, p := range problems {
	fmt.Fprintln(stdout, p)
}
if len(problems) > 0 {
	fmt.Fprintf(stderr, "preflight: %d problem(s) found\n", len(problems))
	return 1
}
fmt.Fprintf(stderr, "preflight: ok for %s\n", s.Env)
return 0
```

Remove the `not implemented yet` lines when you're done.
