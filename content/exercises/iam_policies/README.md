# IAM policies: how AWS decides allow or deny

Every request to AWS, from the console, the CLI or your Go code, is checked
against JSON *policies* before anything happens. When a deploy fails with
`AccessDenied`, you need to know how AWS reached that answer. Here you build
that decision in Go.

**New here:** IAM identities and policies, policy evaluation, custom `UnmarshalJSON` (`json.Unmarshaler`), `regexp.QuoteMeta`.

## Learn

### Who asks: identities

*IAM* (Identity and Access Management) decides who may do what in an
account. Every request is made by an *identity*:

| Identity | What it is | Typical use |
|---|---|---|
| user | a person or program with its own long-term keys | a legacy CI job, a break-glass admin |
| group | a set of users that share policies | `developers`, `oncall` |
| role | an identity with no keys of its own; others *borrow* it for an hour or so | servers, CI pipelines, people logging in through SSO |

Roles are the modern default; *Assume role: trust policies and temporary
credentials* shows how borrowing works. The ideas behind all of this are in
*Cloud and cluster access: IAM, RBAC and short-lived credentials*.

### What they may do: policies

A *policy* is a JSON document attached to an identity. It holds a list of
*statements*:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ReadQueue",
      "Effect": "Allow",
      "Action": ["sqs:ReceiveMessage", "sqs:DeleteMessage"],
      "Resource": "arn:aws:sqs:eu-west-1:111122223333:orders"
    }
  ]
}
```

| Field | Meaning |
|---|---|
| `Version` | the policy *language* version. Always `"2012-10-17"`; it's not a date you update |
| `Statement` | the rules. IAM also accepts a single object; in this course it is always a list |
| `Sid` | an optional label for people |
| `Effect` | `"Allow"` or `"Deny"` |
| `Action` | API calls, written `service:Action`: `s3:GetObject`, `ec2:StartInstances` |
| `Resource` | the ARNs the actions may touch (see *ARNs: accounts, regions and resource scope*) |

`Action` and `Resource` may be **one string or a list**. Two more rules:

- **Action names ignore case**: `s3:getobject` is `s3:GetObject`.
- **Resource ARNs don't**: `acme-app-logs` and `ACME-APP-LOGS` are different
  buckets.

### Wildcards

Both fields may use wildcards: `*` stands for any run of characters
(including none), and `?` for exactly one.

| Pattern | Matches | Doesn't match |
|---|---|---|
| `ec2:Describe*` | `ec2:DescribeInstances`, `ec2:DescribeVolumes` | `ec2:StartInstances` |
| `arn:aws:s3:::acme-app-logs/*` | every object in the bucket | the bucket itself, `arn:aws:s3:::acme-app-logs` |
| `*` | anything | |

(AWS adds a few rules about wildcards and colons in ARNs. This course uses
the simple version above.)

### How AWS decides

AWS collects **every statement of every policy** attached to the identity
into one pool. Order doesn't matter. Then:

```text
request: an action on a resource
        │
        ▼
Does any matching statement say "Deny"? ──yes──► ExplicitDeny
        │ no
        ▼
Does any matching statement say "Allow"? ──yes──► Allow
        │ no
        ▼
ImplicitDeny: nothing allowed it, so the default "no" stands
```

A statement *matches* when one of its actions matches the request's action
**and** one of its resources matches the request's resource. Two things
surprise people:

- **An explicit deny always wins**, even over an `Allow` in another policy.
  That's how security teams build *guardrails* nobody can override.
- **Everything starts denied.** A brand-new role can do nothing at all.

Statements can also have *conditions* ("only from the office network"); you
meet them in *IAM policies: write least-privilege policies*.

### One string or a list: a custom UnmarshalJSON

`Action` can be `"s3:GetObject"` or `["s3:GetObject"]`. A plain `[]string`
field fails on the first form. The starter's `StringList` type accepts both:

```go
type StringList []string

func (l *StringList) UnmarshalJSON(data []byte) error {
	var one string
	if err := json.Unmarshal(data, &one); err == nil {
		*l = StringList{one}
		return nil
	}
	var many []string
	if err := json.Unmarshal(data, &many); err != nil {
		return err
	}
	*l = many
	return nil
}
```

As `fmt` calls `String` on *Stringers*, `encoding/json` looks for an
`UnmarshalJSON([]byte) error` method (the `json.Unmarshaler` interface). When
a type has one, json hands it the raw JSON of that value, here
`"s3:GetObject"` with its quotes, and lets it decide:

1. Try to decode a single string. If that works, store a one-item list.
2. Otherwise decode a list of strings, or return the error.

> **Why is it written like this?** The method name and signature are
> required: json only calls a method spelled exactly
> `UnmarshalJSON([]byte) error`. The pointer receiver `*StringList` is
> required too, because the method must change the caller's value;
> `*l = many` writes through the pointer. Trying the string first is the
> author's choice; the list first works just as well.

### Building a regular expression safely: regexp.QuoteMeta

To use a wildcard pattern as a regular expression, you must stop `.`, `(`,
`+` and friends from meaning something. `regexp.QuoteMeta` puts a `\` before
every special character so it matches literally. Here's a different pattern
language, where `#` stands for one digit:

```go
// digits matches host names like "web-#.example.com".
func digits(pattern string) *regexp.Regexp {
	expr := regexp.QuoteMeta(pattern)        // `web-#\.example\.com`
	expr = strings.ReplaceAll(expr, "#", "[0-9]")
	return regexp.MustCompile("^" + expr + "$")
}

digits("web-#.example.com").MatchString("web-7.example.com") // => true
digits("web-#.example.com").MatchString("web-7xexample.com") // => false: "." is literal now
```

`QuoteMeta` escapes `*` and `?` too: `regexp.QuoteMeta("a*b?")` returns the
text `a\*b\?`. `#` isn't special, so it stays as it is.

> **Why is it written like this?** `^` and `$` anchor the expression, so the
> **whole** value must match, not just a part of it. That is a regexp rule.
> The raw string (backticks) is convention for anything with backslashes:
> `` `\.` `` is easier to read than `"\\."`.

### Try it (optional)

Needs an AWS account and credentials; never needed for the tests.

```sh
aws sts get-caller-identity                    # which identity your CLI calls run as
aws iam get-policy --policy-arn arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess   # an AWS-written policy; note its DefaultVersionId
aws iam get-policy-version --policy-arn arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess \
  --version-id <DefaultVersionId>              # the policy's JSON statements
aws iam simulate-principal-policy --policy-source-arn arn:aws:iam::111122223333:user/alice \
  --action-names s3:GetObject --resource-arns arn:aws:s3:::acme-app-assets/logo.png   # AWS's own evaluator
```

For the last one, use a user or role ARN from your own account. Its
`EvalDecision` is `allowed`, `explicitDeny` or `implicitDeny`: the same three
answers you are about to compute.

## Your task

You're writing a small offline policy checker, so you can ask "would this be
allowed?" without calling AWS. The types, `StringList` and the three
`Decision` constants are in `exercise.go`, ready.

### 1. Match a wildcard pattern

**Goal:** `Match(pattern, value)` reports whether `value` matches `pattern`.

- `*` matches any run of characters, including none.
- `?` matches exactly one character.
- Every other character matches only itself, including `.`, `+`, `(` and `)`.
- The **whole** value must match, and case matters (`Applies` deals with
  case for actions).
- Add `"regexp"` and `"strings"` to the imports, or solve it any other way.

```go
Match("s3:Get*", "s3:GetObject")       // => true
Match("s3:Get*", "s3:Get")             // => true  (* can match nothing)
Match("s3:Get?bject", "s3:GetObject")  // => true
Match("?", "")                         // => false (? needs one character)
Match("arn:aws:s3:::acme-app-logs/*", "arn:aws:s3:::acme-app-logs") // => false
Match("arn:aws:s3:::acme.assets/*", "arn:aws:s3:::acmexassets/logo.png") // => false
Match("arn:aws:s3:::reports-(old)/*", "arn:aws:s3:::reports-(old)/2020.csv") // => true
Match("s3:getobject", "s3:GetObject")  // => false (case matters here)
Match("logs", "app-logs-old")          // => false (whole value)
```

**Think it through:**
- After `QuoteMeta`, what do `*` and `?` look like in the string?
- Which regexp pieces mean "anything, any length" and "exactly one
  character"?

### 2. Check whether a statement applies

**Goal:** `Applies(s, action, resource)` reports whether statement `s` matches the request.

- At least one pattern in `s.Action` matches `action`, **ignoring case**;
  and
- at least one pattern in `s.Resource` matches `resource`, with case.
- `Effect` doesn't matter here: a `Deny` statement applies just like an
  `Allow`.
- No actions or no resources: it applies to nothing.

```go
s := Statement{
	Effect:   "Allow",
	Action:   StringList{"s3:GetObject", "s3:ListBucket"},
	Resource: StringList{"arn:aws:s3:::acme-app-logs", "arn:aws:s3:::acme-app-logs/*"},
}
Applies(s, "s3:GetObject", "arn:aws:s3:::acme-app-logs/prod/web.log") // => true
Applies(s, "S3:getobject", "arn:aws:s3:::acme-app-logs/prod/web.log") // => true
Applies(s, "s3:PutObject", "arn:aws:s3:::acme-app-logs/prod/web.log") // => false
Applies(s, "s3:GetObject", "arn:aws:s3:::ACME-APP-LOGS/prod/web.log") // => false
```

**Think it through:**
- How can you use your case-sensitive `Match` for a case-insensitive check?
- Both lists need a hit. Can you stop early once one of them has none?

### 3. Decide a request

**Goal:** `Evaluate(policies, action, resource)` returns `ExplicitDeny`, `Allow` or `ImplicitDeny`, as in the diagram above.

- Look at every statement of every policy that `Applies`.
- If any of them has `Effect` `"Deny"`, the answer is `ExplicitDeny`.
- Otherwise, if any has `Effect` `"Allow"`, the answer is `Allow`.
- Otherwise (or with no policies at all) it is `ImplicitDeny`.

The tests load two policies from `testdata/`; open them. `developer.json`
lets a developer read app assets, describe EC2, and do anything in their
own sandbox bucket except change its bucket policy. `guardrails.json` only
denies: deleting buckets, anything on the payroll bucket, terminating
servers.

```go
Evaluate([developer], "s3:GetObject", "arn:aws:s3:::acme-app-assets/logo.png")  // => Allow
Evaluate([developer], "s3:PutObject", "arn:aws:s3:::acme-app-assets/logo.png")  // => ImplicitDeny
Evaluate([developer], "s3:PutBucketPolicy", "arn:aws:s3:::acme-dev-sandbox")    // => ExplicitDeny
Evaluate([developer], "s3:DeleteBucket", "arn:aws:s3:::acme-dev-sandbox")       // => Allow
Evaluate([developer, guardrails], "s3:DeleteBucket", "arn:aws:s3:::acme-dev-sandbox") // => ExplicitDeny
Evaluate([guardrails], "s3:GetObject", "arn:aws:s3:::acme-app-assets/logo.png") // => ImplicitDeny
Evaluate(nil, "s3:GetObject", "arn:aws:s3:::acme-app-assets/logo.png")          // => ImplicitDeny
```

**Think it through:**
- When you find a matching `Allow`, can you return straight away? What if a
  later policy denies it?
- When you find a matching `Deny`, can you?

## Learn more

- [Policy evaluation logic](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_evaluation-logic.html)
- [IAM JSON policy element reference](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements.html)
- [json.Unmarshaler](https://pkg.go.dev/encoding/json#Unmarshaler)
- [regexp.QuoteMeta](https://pkg.go.dev/regexp#QuoteMeta)
