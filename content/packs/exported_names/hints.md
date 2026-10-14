## Task 1: Export what callers need

The tests live in `package creds_test`, which is a different package. Code in another package can only use names that start with a capital letter.

### Closer

Change only the first letter: `newCredential` becomes `NewCredential` and `masked` becomes `Masked`. Each doc comment starts with the name, so update those too.

### Answer

```go
// NewCredential returns a Credential for user with the given secret token.
func NewCredential(user, token string) Credential {
	return Credential{User: user, Token: token}
}

// Masked returns the credential in a form that is safe to print in logs,
// such as "deploy-bot:s3****".
func (c Credential) Masked() string {
	return c.User + ":" + mask(c.Token)
}
```

## Task 2: Make the token private

A lower-case first letter makes the field private to the `creds` package. Renaming a field means every use of it has to change too.

### Closer

`Token` appears in three places in `creds.go`: the struct, the struct literal in `NewCredential` (`Token: token`), and `c.Token` in `Header` and `Masked`. All of them become `token`.

### Answer

```go
type Credential struct {
	User  string
	token string
}

func NewCredential(user, token string) Credential {
	return Credential{User: user, token: token}
}

func (c Credential) Header() string {
	return "Bearer " + c.token
}

func (c Credential) Masked() string {
	return c.User + ":" + mask(c.token)
}
```

## Task 3: Mask the token

There are two cases: a short token (4 characters or fewer) and a longer one. Use an `if` on the length.

### Closer

`len(token)` gives the length, and `token[:2]` gives the first two characters. Add `"****"` with `+`.

### Answer

```go
func mask(token string) string {
	if len(token) <= 4 {
		return "****"
	}
	return token[:2] + "****"
}
```
