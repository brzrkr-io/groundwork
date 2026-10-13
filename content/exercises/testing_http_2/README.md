# Testing HTTP: a regression test for a bug report

When a bug report comes in, the professional move is to write a test that
shows the bug *before* you fix it. That test fails, you fix the code, the
test passes, and it guards against the bug coming back forever. Such a test
is a *regression test*.

## Learn

### Recap

`httptest.NewRequest` builds a fake request and `httptest.NewRecorder`
records what a handler writes. Call `h.ServeHTTP(rec, req)`, then check
`rec.Code`, `rec.Header().Get(...)` and `rec.Body.String()`. Report problems
with `t.Errorf`, and use `t.Fatalf` when carrying on makes no sense.

### Requests with a body

`httptest.NewRequest` takes the body as an `io.Reader`.
`strings.NewReader` turns a string into one:

```go
req := httptest.NewRequest("POST", "/deploys", strings.NewReader(`{"service": "api"}`))
```

The test file already has a `postDeploy(body)` helper that does this and
returns the recorder.

### Checking a JSON response

Comparing JSON as a string is fragile: spacing and field order can change
without anything being wrong. Decode it into a struct instead, as you did in
*Decoding JSON: load a service config*. `rec.Body.Bytes()` gives the body as
a `[]byte`, which is exactly what `json.Unmarshal` wants:

```go
var got DeployResponse
if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
	t.Fatalf("body %q is not valid JSON: %v", rec.Body.String(), err)
}
want := DeployResponse{Service: "api", Version: "v1.2.0", Status: "queued"}
if got != want {
	t.Errorf("response = %+v, want %+v", got, want)
}
```

Structs whose fields are all comparable (strings, numbers, bools) can be
compared with `==` in one go. `%+v` prints a struct with its field names,
such as `{Service:api Version:v1.2.0 Status:queued}`.

### How Gopherlings grades this exercise

As in *Testing HTTP: recorders and fake servers*, `TestGrader` re-runs your
`TestDeployAccepted` and `TestDeployRejectsEmptyService` against the real
code in `exercise.go`, where they must pass, and against mutants, where they
must fail. This time the "real code" is the handler **you** fix in task 3.
The mutants:

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

`TestDeployAccepted` only checks the status code. Below the TODO, check that
the `Content-Type` header is `application/json`. Then decode the body into a
`DeployResponse` with `json.Unmarshal` and check that it equals
`DeployResponse{Service: "api", Version: "v1.2.0", Status: "queued"}`. Add the
`"encoding/json"` import.

### 2. Write the regression test

Fill in `TestDeployRejectsEmptyService`. Post both bodies from the bug report
with `postDeploy`:

- `{"service": "", "version": "v1.2.0"}`
- `{"version": "v1.2.0"}`

For each one, check that the status is 400, decode the body into an
`ErrorResponse`, and check that its `Error` field is `service is required`.

Your new test **fails** now, and `TestGrader` says your tests must pass on
the real code first. That's the point: the test has caught the bug.

### 3. Fix the bug

In `handleDeploy`, reject a deploy whose service is empty, before the
version check. Answer with `writeJSON`, status 400, and
`ErrorResponse{Error: "service is required"}`, then `return`.

The exercise is done when both of your tests pass and `TestGrader` reports
that every mutant was caught.

## Learn more

- [Package net/http/httptest](https://pkg.go.dev/net/http/httptest)
- [json.Unmarshal](https://pkg.go.dev/encoding/json#Unmarshal)
- [Go by Example: JSON](https://gobyexample.com/json)
