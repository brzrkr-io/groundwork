## Task 1: just capital letters

Change the first letter of `newCredential` and `masked` to upper case. Every
name that a test in `package creds_test` uses must start with a capital
letter.

## Task 2: rename the field everywhere

The field appears in three places in `creds.go`: the struct, the
constructor's struct literal (`Token: token`), and `Header`. All three must
become `token`.

## Task 3: slice the first two bytes

`len(token)` gives the length, and `token[:2]` gives the first two
characters:

```go
func mask(token string) string {
	if len(token) <= 4 {
		return "****"
	}
	return token[:2] + "****"
}
```
