# AWS API errors: codes, faults and NoSuchKey

AWS calls fail all the time: a file is missing, a bucket name has a typo,
you hit a rate limit, or AWS has a bad minute. Good automation tells these
apart. It shrugs off a missing file, retries a rate limit later, and stops
with a clear message on everything else.

## Learn

### What a failed SDK call returns

When AWS rejects a call, the SDK returns an error with three layers, each
wrapping the next, just like the chains in *Wrapping errors*:

```text
operation error S3: GetObject, https response error StatusCode: 404, RequestID: 4e7f1c9a-EXAMPLE, NoSuchKey: The specified key does not exist.

*smithy.OperationError           which service and operation failed (S3, GetObject)
  wraps *awshttp.ResponseError   the HTTP status code and AWS request ID
    wraps *types.NoSuchKey       the API error: what AWS actually said
```

Only the innermost layer, the *API error*, tells you what went wrong. The
outermost layer is a `*smithy.OperationError`, so a type assertion such as
`err.(*types.NoSuchKey)` always fails. Use `errors.AsType`, which searches
the whole chain. Matching on `err.Error()` text is fragile, so don't.

### smithy.APIError: the code of any AWS error

Every API error, from every AWS service, satisfies the `smithy.APIError`
interface from the package `github.com/aws/smithy-go`:

```go
type APIError interface {
	error
	ErrorCode() string         // "NoSuchKey", "AccessDenied", "SlowDown", ...
	ErrorMessage() string      // human text from AWS
	ErrorFault() smithy.ErrorFault
}
```

`errors.AsType` works with an interface type too. It finds the first error
in the chain that has these methods:

```go
if apiErr, ok := errors.AsType[smithy.APIError](err); ok {
	fmt.Println(apiErr.ErrorCode()) // "NoSuchKey"
}
```

`ErrorFault()` says whose fault the failure was:

| Fault | Meaning | Retry? |
|---|---|---|
| `smithy.FaultClient` | Your request was wrong: missing file, no permission, bad input. | No, the same request fails again. |
| `smithy.FaultServer` | AWS had a problem on its side. | Yes, later. |
| `smithy.FaultUnknown` | AWS did not say. | Decide by code. |

### Throttling codes

AWS limits how fast you can call each API. Go over the limit and you get a
*throttling* error. Different services use different codes for it. In this
exercise these five codes count as throttling, and the starter has an
`isThrottling(code)` helper that checks for them:

```text
Throttling  ThrottlingException  RequestLimitExceeded  TooManyRequestsException  SlowDown
```

The SDK already retries throttling and server errors for you, up to 3
attempts in total by default. Your own checks decide what happens after
the SDK gives up, for example whether your deploy loop waits and tries the
whole step again, as in *Retries: backoff and jitter*.

### Typed S3 errors

Some services also give their errors their own Go types. S3 has
`*types.NoSuchKey` (the object does not exist) and `*types.NoSuchBucket` (the
bucket does not exist), from `github.com/aws/aws-sdk-go-v2/service/s3/types`.
Ask for the exact type when you need to react to one specific error:

```go
if _, ok := errors.AsType[*types.NoSuchKey](err); ok {
	// the object is not there
}
```

### Imports

The starter only imports what it already uses. As you work, add these to the
import block:

```go
"errors"
"fmt"

"github.com/aws/aws-sdk-go-v2/service/s3/types"
"github.com/aws/smithy-go"
```

The package name of `github.com/aws/smithy-go` is `smithy`.

## Your task

Your deploy tool reads each service's settings file from S3. Give it error
handling that knows what AWS is telling it.

### 1. Read the error code

Write `ErrorCode(err)`. If a `smithy.APIError` is anywhere in `err`'s chain,
return its `ErrorCode()`. Otherwise (including when `err` is `nil`) return
`""`.

```go
ErrorCode(errFromGetObject)             // NoSuchKey from the SDK
// => "NoSuchKey"
ErrorCode(errors.New("dial tcp: i/o timeout"))
// => ""
```

### 2. Decide whether to retry

Write `ShouldRetry(err)`. Return `true` when the chain holds a
`smithy.APIError` and either:

- its code is one of the five throttling codes (call `isThrottling`), or
- its `ErrorFault()` is `smithy.FaultServer`.

The starter already returns `isThrottling(ErrorCode(err))`, so the first
case works as soon as task 1 does. Add the second.

Return `false` for everything else: `nil`, errors that are not AWS API
errors, and API errors with any other code and a client or unknown fault.

```go
ShouldRetry(slowDownErr)     // code "SlowDown"
// => true
ShouldRetry(internalErr)     // code "InternalError", fault smithy.FaultServer
// => true
ShouldRetry(accessDeniedErr) // code "AccessDenied", fault smithy.FaultClient
// => false
```

### 3. Handle errors in LoadSettings

`LoadSettings(ctx, api, bucket, key)` already calls `GetObject` and reads the
body. Finish the `if err != nil` block:

- If the chain holds a `*types.NoSuchKey`, the service has no settings file
  yet. Return `[]byte(DefaultSettings), nil`.
- Any other error: return `nil` and the error wrapped with `%w` as
  `load settings s3://BUCKET/KEY: <err>`.

```go
LoadSettings(ctx, api, "deploy-configs", "payments/settings.json")
// key missing    => []byte(DefaultSettings), nil
// bucket missing => nil, error "load settings s3://deploy-configs/payments/settings.json: operation error S3: GetObject, ..."
```

The tests check that `errors.AsType[*types.NoSuchBucket]` and
`errors.AsType[smithy.APIError]` still find the original error through your
wrap.

## Learn more

- [AWS SDK for Go v2: Handle errors](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/handle-errors.html)
- [AWS SDK for Go v2: Retries and timeouts](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-retries-timeouts.html)
- [Package smithy: APIError](https://pkg.go.dev/github.com/aws/smithy-go#APIError)
- [Amazon S3 error responses](https://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html)
