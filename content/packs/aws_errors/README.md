# AWS API errors: codes, faults and NoSuchKey

AWS calls fail all the time: a missing file, a typo in a bucket name, a rate
limit, or a bad minute at AWS. Good automation tells these apart: it shrugs
off a missing file, retries a rate limit, and stops clearly on everything else.

**New here:** `smithy.APIError`, `ErrorCode()`, `ErrorFault()`, typed S3 errors like `*types.NoSuchKey`.

## Learn

### What a failed SDK call returns

A failed AWS call gives you an error *chain*: three layers, each wrapping the
next, like the chains in *Wrapping errors*.

```text
operation error S3: GetObject, https response error StatusCode: 404, RequestID: 4e7f1c9a-EXAMPLE, NoSuchKey: The specified key does not exist.

*smithy.OperationError           which service and operation failed (S3, GetObject)
  wraps *awshttp.ResponseError   the HTTP status code and AWS request ID
    wraps *types.NoSuchKey       the API error: what AWS actually said
```

Only the innermost layer, the *API error*, says what went wrong. A type
assertion like `err.(*types.NoSuchKey)` only looks at the outer layer, so it
always fails. `errors.AsType` searches the whole chain.

> **Why is it written like this?** Searching the chain is required in
> practice: the SDK always wraps. Not matching on `err.Error()` text is
> convention: text can change, codes and types don't.

### smithy.APIError: the code of any AWS error

Every API error, from every AWS service, satisfies the *interface*
`smithy.APIError` (package `github.com/aws/smithy-go`):

```go
type APIError interface {
	error
	ErrorCode() string         // "NoSuchKey", "AccessDenied", "SlowDown", ...
	ErrorMessage() string      // human text from AWS
	ErrorFault() smithy.ErrorFault
}
```

`errors.AsType` works with an interface type too. It returns the first error
in the chain that has these methods, plus an `ok` bool:

```go
if apiErr, ok := errors.AsType[smithy.APIError](err); ok {
	log.Println("AWS says:", apiErr.ErrorMessage())
}
```

> **Why is it written like this?** The type in square brackets is required:
> it tells `AsType` what to look for. The names `apiErr` and `ok` are
> convention; you may pick others.

`ErrorFault()` says whose fault the failure was:

| Fault | Meaning | Retry? |
|---|---|---|
| `smithy.FaultClient` | Your request was wrong: missing file, no permission, bad input. | No, it fails again. |
| `smithy.FaultServer` | AWS had a problem on its side. | Yes, later. |
| `smithy.FaultUnknown` | AWS did not say. | Decide by code. |

You compare a fault with `==`, like any other value.

### Throttling codes

AWS limits how fast you call each API. Go over and you get a *throttling*
error. In this exercise these five codes count as throttling, and the
starter's helper `isThrottling(code)` checks for them:

```text
Throttling  ThrottlingException  RequestLimitExceeded  TooManyRequestsException  SlowDown
```

The SDK already retries these (3 attempts by default). Your checks decide
what happens after the SDK gives up, as in *Retries: backoff and jitter*.

### Typed S3 errors

S3 also gives some errors their own Go types, in
`github.com/aws/aws-sdk-go-v2/service/s3/types`: `*types.NoSuchKey` (object
missing) and `*types.NoSuchBucket` (bucket missing). Ask for the exact type
to react to one specific error:

```go
if _, ok := errors.AsType[*types.NoSuchBucket](err); ok {
	log.Println("check the bucket name")
}
```

> **Why is it written like this?** The `*` is required: these errors are
> pointers. `_` is required when you don't use a value, because Go rejects
> unused variables.

### Imports

The starter imports only what it already uses. Add these as you go:

```go
"errors"
"fmt"

"github.com/aws/aws-sdk-go-v2/service/s3/types"
"github.com/aws/smithy-go"
```

The package at `github.com/aws/smithy-go` is named `smithy`, so you write
`smithy.APIError`.

## Your task

Your deploy tool reads each service's settings file from S3. Give it error
handling that knows what AWS is telling it.

### 1. Read the error code

**Goal:** `ErrorCode(err)` returns the code of the AWS API error in `err`'s chain, or `""`.

- A `smithy.APIError` anywhere in the chain: return its `ErrorCode()`.
- No API error, or `err` is `nil`: return `""`.

```go
ErrorCode(errFromGetObject)             // NoSuchKey from the SDK
// => "NoSuchKey"
ErrorCode(fmt.Errorf("stop web-1: %w", errFromStopInstances)) // wrapped again
// => "InvalidInstanceID.NotFound"
ErrorCode(errors.New("dial tcp: i/o timeout"))
// => ""
```

**Think it through:**
- Which type do you ask `errors.AsType` for: a specific error, or the interface every AWS error has?
- What does `ok` tell you, and what do you return when it is `false`?

### 2. Decide whether to retry

**Goal:** `ShouldRetry(err)` returns `true` for throttling codes and server faults, `false` otherwise.

Return `true` when the chain holds a `smithy.APIError` and either:

- its code is one of the five throttling codes (use `isThrottling`), or
- its `ErrorFault()` is `smithy.FaultServer`.

The starter already returns `isThrottling(ErrorCode(err))`, so the first case
works once task 1 does. Return `false` for everything else: `nil`, non-AWS
errors, and other codes with a client or unknown fault.

```go
ShouldRetry(slowDownErr)     // code "SlowDown"
// => true
ShouldRetry(internalErr)     // code "InternalError", fault smithy.FaultServer
// => true
ShouldRetry(accessDeniedErr) // code "AccessDenied", fault smithy.FaultClient
// => false
```

**Think it through:**
- To read the fault you need the API error itself, not just its code. How did you get it in task 1?
- Two reasons to retry, either is enough: which operator means "or"?

### 3. Handle errors in LoadSettings

**Goal:** in `LoadSettings`, a missing key gives the default settings; any other error is wrapped and returned.

`LoadSettings(ctx, api, bucket, key)` already calls `GetObject` and reads the
body. Finish its `if err != nil` block:

- The chain holds a `*types.NoSuchKey` (no settings file yet): return
  `[]byte(DefaultSettings), nil`.
- Any other error: return `nil` and the error wrapped with `%w` as
  `load settings s3://BUCKET/KEY: <err>`. `BUCKET` and `KEY` are the
  parameters `bucket` and `key`; the rest is fixed text.

```go
LoadSettings(ctx, api, "deploy-configs", "payments/settings.json")
// key missing    => []byte(DefaultSettings), nil
// bucket missing => nil, error "load settings s3://deploy-configs/payments/settings.json: operation error S3: GetObject, ..."
```

The tests check that `errors.AsType[*types.NoSuchBucket]` and
`errors.AsType[smithy.APIError]` still find the original error through your
wrap.

**Think it through:**
- Which check comes first: the "file missing" case or the wrap?
- In the message, which parts are fixed text and which are `%s` verbs filled from parameters?

## Learn more

- [AWS SDK for Go v2: Handle errors](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/handle-errors.html)
- [AWS SDK for Go v2: Retries and timeouts](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-retries-timeouts.html)
- [Package smithy: APIError](https://pkg.go.dev/github.com/aws/smithy-go#APIError)
- [Amazon S3 error responses](https://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html)
