# AWS API errors: EC2 dry-run permission checks

Before a bot terminates anything, it should check that it is allowed to. EC2
has a built-in way to ask, and the answer comes back as an error code.

**New here:** EC2 error codes, `DryRun` requests.

## Learn

### Recap: search the chain

An SDK error is a chain: `*smithy.OperationError` wraps the HTTP error, which
wraps the API error. Find the API error with
`errors.AsType[smithy.APIError](err)` and read `ErrorCode()`. Never use a
type assertion or match `err.Error()` text.

| Check | Looks at |
|---|---|
| `err.(smithy.APIError)` (type assertion) | the outermost error only |
| `errors.AsType[smithy.APIError](err)` | every error in the chain |

### EC2 errors are codes only

EC2 has no typed errors like S3's `*types.NoSuchKey`. Every EC2 error is a
`smithy.APIError` with a code and a message, so you decide by code:

| Code | Meaning |
|---|---|
| `InvalidInstanceID.NotFound` | No instance with that ID (often: already terminated). |
| `UnauthorizedOperation` | Your IAM permissions don't allow this call. |
| `DryRunOperation` | Dry run only: the call *would* have succeeded. |

### Dry runs: the error that means yes

Most EC2 calls that change things take `DryRun: aws.Bool(true)`. EC2 then
checks permissions, changes nothing, and *always* answers with an error:
`DryRunOperation` if the real call would work, or the real problem (such as
`UnauthorizedOperation`) if not.

### Choosing by code with switch

A `switch` compares one value against several cases. Cases you don't list
fall through to the code after the `switch`:

```go
switch region {
case "us-east-1":
	return "Virginia"
case "eu-west-1":
	return "Ireland"
}
return "somewhere else"
```

> **Why is it written like this?** `case` lines with a colon are required
> syntax. Go needs no `break`: a case never runs into the next one. Putting
> the "anything else" `return` after the `switch` instead of a `default:` is
> the author's choice; both work.

## Your task

A cleanup bot terminates stale CI runner instances. Fix its not-found check
and finish its permission check.

### 1. Fix IsNotFound

**Goal:** `IsNotFound(err)` returns `true` when an API error with code `InvalidInstanceID.NotFound` is anywhere in `err`'s chain.

The starter's type assertion sees only the outermost error, so it works on a
bare API error but misses every real SDK error. Change how it finds the API
error, and add the `"errors"` import. The `return` line can stay.

`nil`, other codes, and plain errors (even one whose text is
`"InvalidInstanceID.NotFound"`) return `false`.

```go
IsNotFound(errFromTerminateInstances) // code InvalidInstanceID.NotFound
// => true
IsNotFound(fmt.Errorf("stop ci-runner-7: %w", errFromTerminateInstances))
// => true
IsNotFound(errors.New("InvalidInstanceID.NotFound"))
// => false
```

**Think it through:**
- Which line finds `apiErr`? That is the only line to change.
- Look at the recap table: which check sees the whole chain?

### 2. Finish CanTerminate

**Goal:** `CanTerminate(ctx, api, id)` turns the dry-run error into an answer: allowed, not allowed, or a real error.

It already sends `TerminateInstances` for just `id` with
`DryRun: aws.Bool(true)`. Decide by the error's code:

| Error | Return |
|---|---|
| code `DryRunOperation` | `true, nil` |
| code `UnauthorizedOperation` | `false, nil` |
| anything else (other codes, non-AWS errors) | `false` and the error wrapped with `%w` as `dry-run terminate ID: <err>` |

`ID` is the `id` parameter; `dry-run terminate ` and `: ` are fixed text.

```go
CanTerminate(ctx, api, "i-0a1b2c3d4e5f60718")
// DryRunOperation       => true, nil
// UnauthorizedOperation => false, nil
// RequestLimitExceeded  => false, error "dry-run terminate i-0a1b2c3d4e5f60718: operation error EC2: ..."
```

**Think it through:**
- Find the API error as in task 1. What should happen if there is none?
- Two codes get special answers. Where do all other cases end up?

## Learn more

- [EC2 API error codes](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/errors-overview.html)
- [EC2 TerminateInstances](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_TerminateInstances.html)
- [AWS SDK for Go v2: Handle errors](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/handle-errors.html)
- [Go by Example: Switch](https://gobyexample.com/switch)
