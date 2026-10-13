# AWS API errors: EC2 dry-run permission checks

Before a bot terminates anything, it should check that it is allowed to. EC2
has a built-in way to ask, and the answer comes back as an error code.

## Learn

Recap: an SDK error is a chain (`*smithy.OperationError` wrapping the HTTP
error wrapping the API error), so search it with
`errors.AsType[smithy.APIError](err)` and read `ErrorCode()`, never with a
type assertion or by matching `err.Error()` text.

New: EC2 has no typed errors like S3's `*types.NoSuchKey`. Every EC2 error
is a `smithy.APIError` with just a code and a message, so you always decide
by code:

| Code | Meaning |
|---|---|
| `InvalidInstanceID.NotFound` | No instance with that ID (often: already terminated). |
| `UnauthorizedOperation` | Your IAM permissions don't allow this call. |
| `DryRunOperation` | Dry run only: the call *would* have succeeded. |

Most EC2 mutating calls take `DryRun: aws.Bool(true)`. EC2 then checks your
permissions and the request, changes nothing, and always answers with an
error: `DryRunOperation` if the real call would work, or the real problem
(such as `UnauthorizedOperation`) if it would not. It is the one place where
an error means yes.

## Your task

A cleanup bot terminates stale CI runner instances. Fix its not-found check
and finish its permission check.

### 1. Fix IsNotFound

`IsNotFound(err)` should return `true` when an API error with code
`InvalidInstanceID.NotFound` is anywhere in `err`'s chain. Its type
assertion only sees the outermost error, so it works on a bare API error but
misses every error that really comes back from the SDK. Use `errors.AsType`
instead, and add the `"errors"` import.

`nil`, other codes, and plain errors (even one whose text is
`"InvalidInstanceID.NotFound"`) return `false`.

```go
IsNotFound(errFromTerminateInstances) // code InvalidInstanceID.NotFound
// => true
IsNotFound(errors.New("InvalidInstanceID.NotFound"))
// => false
```

### 2. Finish CanTerminate

`CanTerminate(ctx, api, id)` already sends `TerminateInstances` for just
`id` with `DryRun: aws.Bool(true)`. Decide by the error's code:

| Error | Return |
|---|---|
| code `DryRunOperation` | `true, nil` |
| code `UnauthorizedOperation` | `false, nil` |
| anything else (other codes, non-AWS errors) | `false` and the error wrapped with `%w` as `dry-run terminate ID: <err>` |

```go
CanTerminate(ctx, api, "i-0a1b2c3d4e5f60718")
// DryRunOperation       => true, nil
// UnauthorizedOperation => false, nil
// RequestLimitExceeded  => false, error "dry-run terminate i-0a1b2c3d4e5f60718: operation error EC2: ..."
```

## Learn more

- [EC2 API error codes](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/errors-overview.html)
- [EC2 TerminateInstances](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_TerminateInstances.html)
- [AWS SDK for Go v2: Handle errors](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/handle-errors.html)
