## Task 1: Read a setting with a default

Call `lookupEnv(key)` to get the value and whether it exists. Trim the
value, then decide: is there anything left to return?

### Closer

`strings.TrimSpace` removes spaces at both ends. After trimming, an unset
variable and a blank one both give `""`, so one check covers them:

```go
value, ok := lookupEnv(key)
value = strings.TrimSpace(value)
if !ok || value == "" {
	// ...
}
```

### Answer

```go
func Lookup(lookupEnv func(string) (string, bool), key, fallback string) string {
	value, ok := lookupEnv(key)
	value = strings.TrimSpace(value)
	if !ok || value == "" {
		return fallback
	}
	return value
}
```

## Task 2: Parse the port

Let `Lookup` supply the text, with `"8080"` as its fallback. Then two checks
in order: is it a number, and is the number in range?

### Closer

`strconv.Atoi(value)` returns the number and an error. Build each message
with `fmt.Errorf`: `%q` for the text (it adds the double quotes), `%d` for
the number. On any error, return `0` first.

```go
if port < 1 || port > 65535 {
	return 0, fmt.Errorf(/* ... */)
}
```

### Answer

```go
func Port(lookupEnv func(string) (string, bool)) (int, error) {
	value := Lookup(lookupEnv, "PORT", "8080")
	port, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("PORT %q is not a number", value)
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("PORT %d is out of range 1-65535", port)
	}
	return port, nil
}
```

## Task 3: Load the whole config

Check the settings one after another, appending each problem to a
`[]error` instead of returning early. After the last check,
`errors.Join` decides whether anything went wrong.

### Closer

`Port` gives you the port and its error; `Lookup` gives you `APP_ENV` and
`LOG_LEVEL` with their defaults. `DATABASE_URL` needs `lookupEnv` itself,
because only its `ok` result tells unset from blank. To check `APP_ENV`
against three names, use `slices.Contains([]string{...}, env)` or a
`switch` with one `case` listing all three.

```go
var errs []error
port, err := Port(lookupEnv)
if err != nil {
	errs = append(errs, err)
}
// ... APP_ENV, DATABASE_URL ...
if err := errors.Join(errs...); err != nil {
	return Config{}, err
}
```

### Answer

```go
func Load(lookupEnv func(string) (string, bool)) (Config, error) {
	var errs []error

	port, err := Port(lookupEnv)
	if err != nil {
		errs = append(errs, err)
	}

	env := Lookup(lookupEnv, "APP_ENV", "dev")
	if !slices.Contains([]string{"dev", "staging", "production"}, env) {
		errs = append(errs, fmt.Errorf("APP_ENV %q must be dev, staging or production", env))
	}

	databaseURL, ok := lookupEnv("DATABASE_URL")
	databaseURL = strings.TrimSpace(databaseURL)
	if !ok {
		errs = append(errs, errors.New("DATABASE_URL is not set"))
	} else if databaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is set but empty"))
	}

	if err := errors.Join(errs...); err != nil {
		return Config{}, err
	}
	return Config{
		Port:        port,
		Env:         env,
		DatabaseURL: databaseURL,
		LogLevel:    Lookup(lookupEnv, "LOG_LEVEL", "info"),
	}, nil
}
```

## Task 4: Redact the password in logs

Work out the `db=` part first: parse `c.DatabaseURL`, and use `Redacted` if
that worked or `"invalid"` if it didn't. Then put the four parts into one
line.

### Closer

A method on `Config` starts with `func (c Config) String() string`. Start
with `db := "invalid"` and replace it only when `url.Parse` returns no
error. `fmt.Sprintf("port=%d env=%s ...", ...)` builds the line.

### Answer

```go
func (c Config) String() string {
	db := "invalid"
	if u, err := url.Parse(c.DatabaseURL); err == nil {
		db = u.Redacted()
	}
	return fmt.Sprintf("port=%d env=%s log=%s db=%s", c.Port, c.Env, c.LogLevel, db)
}
```
