# Testing HTTP: a regression test for a bug report

When a bug report comes in, the professional move is to write a test that
shows the bug *before* you fix it. The test fails, you fix the code, the test
passes, and it guards against the bug coming back. That's a *regression test*.

**New here:** request bodies in tests, checking JSON responses, test-first bug fixing.

## Learn

### Recap

- `httptest.NewRequest` builds a fake request; `httptest.NewRecorder` records
  what a handler writes.
- Call `h.ServeHTTP(rec, req)`, then check `rec.Code`,
  `rec.Header().Get(...)` and `rec.Body.String()`.
- `t.Errorf` reports and carries on; `t.Fatalf` stops when carrying on makes no sense.

### Requests with a body

`httptest.NewRequest` takes the body as an `io.Reader`. `strings.NewReader`
turns a string into one:

```go
req := httptest.NewRequest("PUT", "/users/7", strings.NewReader(`{"name": "ana"}`))
```

The test file already has a `postDeploy(body)` helper that does this for
`POST /deploys` and returns the recorder. You only pass the JSON text.

> **Why is it written like this?** The backticks make a *raw string*: Go
> keeps everything inside as-is, so the JSON's own `"` quotes need no escaping.
> Using raw strings for JSON is convention.

### Checking a JSON response

Comparing JSON as text is fragile: spacing and field order can change without
anything being wrong. Decode it into a struct instead. `rec.Body.Bytes()`
gives the body as `[]byte`, which is what `json.Unmarshal` wants:

```go
var got User
if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
	t.Fatalf("body %q is not valid JSON: %v", rec.Body.String(), err)
}
want := User{Name: "ana", Role: "admin"}
if got != want {
	t.Errorf("user = %+v, want %+v", got, want)
}
```

- Structs whose fields are all strings, numbers or bools compare with `==` in one go.
- `%+v` prints a struct with field names: `{Name:ana Role:admin}`.
- To check one field, compare it directly: `got.Name != "ana"`.

> **Why is it written like this?** Go requires `&got` (a pointer) so
> `Unmarshal` can fill the variable in. `t.Fatalf` here is the author's choice:
> if the body isn't JSON, the next checks are meaningless.

### How Groundwork grades this exercise

As in *Testing HTTP: recorders and fake servers*, `TestGrader` runs your
`TestDeployAccepted` and `TestDeployRejectsEmptyService` against the real
code in `exercise.go` (must pass) and against mutants (must fail). This time
the "real code" is the handler **you** fix in task 3.

| Mutant | The bug |
|---|---|
| `ok_instead_of_accepted` | an accepted deploy gets status 200 instead of 202 |
| `text_content_type` | responses have `Content-Type: text/plain; charset=utf-8` |
| `pending_status` | an accepted deploy has `"status": "pending"` |
| `drops_version` | an accepted deploy has an empty `"version"` |
| `accepts_empty_service` | the bug from the report below |
| `vague_error` | an empty service gets the error `"bad request"` |

## Your task

The deploy API in `exercise.go` takes `POST /deploys` with a JSON body. A
good deploy is answered like this:

```text
POST /deploys  {"service": "api", "version": "v1.2.0"}
=> 202 Accepted, Content-Type: application/json
   {"service": "api", "version": "v1.2.0", "status": "queued"}
```

Then this bug report arrives:

> Someone's script sent `{"service": "", "version": "v1.2.0"}` and the API
> answered 202. A deploy of nothing got queued. It should answer
> 400 Bad Request with `{"error": "service is required"}`. The same goes for
> a body with no `"service"` field at all.

### 1. Finish the happy-path test

**Goal:** in `TestDeployAccepted`, check the `Content-Type` header and the decoded JSON body.

Below the `TODO(1)`:

- `Content-Type` header must be `application/json`
- decode the body into a `DeployResponse` with `json.Unmarshal`
- it must equal `DeployResponse{Service: "api", Version: "v1.2.0", Status: "queued"}`
- add the `"encoding/json"` import

**Think it through:**
- Which struct type do you decode into, and what is your `want` value?
- If decoding fails, should the test carry on?

### 2. Write the regression test

**Goal:** fill in `TestDeployRejectsEmptyService` so it posts both bad bodies and checks for a 400 with the right error.

Post each body with `postDeploy`:

- `{"service": "", "version": "v1.2.0"}`
- `{"version": "v1.2.0"}`

For each one:

- status must be 400
- decode the body into an `ErrorResponse`
- its `Error` field must be `service is required`

Your new test **fails** now, and `TestGrader` says your tests must pass on the
real code first. That's the point: the test has caught the bug.

**Think it through:**
- Two inputs, same checks: a small table and a loop keep it short.
- Which struct matches the error JSON, and which field do you compare?

### 3. Fix the bug

**Goal:** in `handleDeploy`, reject a deploy whose service is empty, before the version check.

- answer with `writeJSON`, status 400, and an `ErrorResponse` whose `Error` is
  `service is required`
- then `return`

Done when both your tests pass and `TestGrader` reports every mutant caught.

**Think it through:**
- The version check right below does almost the same thing. What differs?
- Which field of `req` holds the service?

## Learn more

- [Package net/http/httptest](https://pkg.go.dev/net/http/httptest)
- [json.Unmarshal](https://pkg.go.dev/encoding/json#Unmarshal)
- [Go by Example: JSON](https://gobyexample.com/json)
