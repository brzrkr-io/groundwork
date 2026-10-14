# S3 bucket policies: is this bucket public?

"Company leaves customer data in a public S3 bucket" is a headline every year.
The cause is nearly always one line of JSON. Here you write the check a
security team runs on every bucket: is it public, is that blocked, and does
it insist on HTTPS?

**New here:** S3 bucket policies, principals in resource policies, S3 ACLs, Block Public Access, `aws:SecureTransport`.

## Learn

### Recap: policies in Go

As in *IAM policies: how AWS decides allow or deny*, a policy is a JSON
`Statement` list (always a list in this course), and `StringList` decodes
`"x"` or `["x", "y"]`. A `Condition` is a map of maps: operator → condition
key → values. The starter has all of this ready.

### Buckets and objects

*S3* stores files called *objects* in *buckets*. A bucket name is unique
across all of AWS. Each has two kinds of ARN (as in
*ARNs: accounts, regions and resource scope*):

| ARN | Means |
|---|---|
| `arn:aws:s3:::acme-www` | the bucket itself (list it, read its settings) |
| `arn:aws:s3:::acme-www/*` | every object in it (read or write files) |

### Bucket policies: the bucket says who may come in

The policies you wrote before were attached to a user or role and said what
**it** may do. A *bucket policy* is attached to the bucket and says **who**
may do what to it. So each statement has a `Principal`:

```json
{
  "Sid": "PublicRead",
  "Effect": "Allow",
  "Principal": "*",
  "Action": "s3:GetObject",
  "Resource": "arn:aws:s3:::acme-www/*"
}
```

| Principal | Who |
|---|---|
| `"*"` or `{"AWS": "*"}` | **anyone**, signed in or not: the whole internet |
| `{"AWS": "arn:aws:iam::444455556666:root"}` | account `444455556666` |
| `{"AWS": "arn:aws:iam::111122223333:role/log-reader"}` | one role |
| `{"Service": "cloudfront.amazonaws.com"}` | an AWS service |

The starter's `Principal` type decodes both shapes; for `"*"` it stores
`AWS: ["*"]`, so you only check one field.

### When `"*"` isn't public

An Allow for `"*"` with a condition on one of these keys is limited to known
networks or accounts, so AWS doesn't call it public:

| Condition key | Limits requests to |
|---|---|
| `aws:SourceIp` | these IP ranges (your office) |
| `aws:SourceVpce` | your private network's *VPC endpoint* |
| `aws:PrincipalOrgID` | accounts in your AWS Organization |
| `aws:PrincipalAccount` | these accounts |

```json
"Condition": { "IpAddress": { "aws:SourceIp": "203.0.113.0/24" } }
```

AWS compares condition key names ignoring case. A condition on anything
else, such as `aws:SecureTransport` or `s3:prefix`, still lets the whole
internet in. (Real AWS is stricter still: `aws:SourceIp` of `0.0.0.0/0`
counts as public. This exercise skips that.)

### ACLs: the older way

Before bucket policies, S3 had *ACLs* (access control lists): grants to
*grantees*. Two grantees are groups that mean "public":

| Grantee | Means |
|---|---|
| `http://acs.amazonaws.com/groups/global/AllUsers` | anyone on the internet |
| `http://acs.amazonaws.com/groups/global/AuthenticatedUsers` | anyone with **any** AWS account (not just yours!) |

Others, such as a canonical user ID or
`http://acs.amazonaws.com/groups/s3/LogDelivery`, are not public.

### Block Public Access

Four on/off settings sit on top of policies and ACLs:

| Setting | What it does | Fixes an existing public… |
|---|---|---|
| `BlockPublicAcls` | rejects **new** public ACLs | nothing |
| `IgnorePublicAcls` | S3 ignores public ACLs | **ACL** |
| `BlockPublicPolicy` | rejects **new** public bucket policies | nothing |
| `RestrictPublicBuckets` | only your account and AWS services get through a public policy | **policy** |

The trap: the two `Block…` settings stop someone from *adding* a public
policy or ACL, but one that is already there keeps working.

### Requiring HTTPS

S3 accepts plain HTTP unless you forbid it. `aws:SecureTransport` is
`"true"` for HTTPS requests and `"false"` for HTTP. The standard fix is a
**Deny**:

```json
{
  "Effect": "Deny",
  "Principal": "*",
  "Action": "s3:*",
  "Resource": ["arn:aws:s3:::acme-app-logs", "arn:aws:s3:::acme-app-logs/*"],
  "Condition": { "Bool": { "aws:SecureTransport": "false" } }
}
```

An *Allow* with `"aws:SecureTransport": "true"` doesn't do it: any other Allow
still works over HTTP. A Deny always wins.

### Reading nested maps safely

Reading a missing key gives the zero value, even from a `nil` map. For a map
of maps the zero value is a `nil` inner map, so you can chain lookups:

```go
limits := map[string]map[string]int{"prod": {"cpu": 4}}
limits["prod"]["cpu"]      // => 4
limits["staging"]["cpu"]   // => 0, no panic
```

> **Why is it written like this?** Reading from a `nil` map is allowed by
> the Go spec; only *writing* to one panics. Chaining like this is the
> author's choice to keep a check on one line.

### Try it (optional)

Needs an AWS account and credentials; never needed for the tests.

```bash
aws s3api get-bucket-policy-status --bucket BUCKET_NAME
aws s3api get-public-access-block --bucket BUCKET_NAME
aws s3api get-bucket-policy --bucket BUCKET_NAME --query Policy --output text
aws s3api get-bucket-acl --bucket BUCKET_NAME
```

- `get-bucket-policy-status` shows AWS's own verdict: `"IsPublic": true`.
- `get-public-access-block` shows the four settings.
- `get-bucket-policy` prints the policy JSON.
- `get-bucket-acl` lists the grantees.

## Your task

You're auditing your company's buckets. Their policies are in `testdata/`
(`public_website.json`, `office_only.json`, `https_only.json`, …): open them.

### 1. Spot a public statement

**Goal:** `PublicStatement(s)` returns `true` when `s` is `"Effect": "Allow"`, its `Principal.AWS` lists `"*"`, and no condition key restricts it.

- A condition key restricts when it equals one of `RestrictingKeys`
  (`aws:SourceIp`, `aws:SourceVpce`, `aws:PrincipalOrgID`,
  `aws:PrincipalAccount`), ignoring case, under **any** operator.
- Keys from other conditions don't help (`aws:SecureTransport`, `s3:prefix`).
- `Deny` statements, and Allows for an account, role or service, are never
  public.

```go
PublicStatement(/* public_website.json: Allow "*" s3:GetObject */)             // => true
PublicStatement(/* allow_secure.json: same, with aws:SecureTransport "true" */) // => true
PublicStatement(/* office_only.json: same, with aws:SourceIp */)               // => false
PublicStatement(/* org_only.json: {"AWS": "*"} with aws:PrincipalOrgID */)     // => false
PublicStatement(/* Allow "*" with IpAddress {"aws:sourceip": ...} */)          // => false (case doesn't matter)
PublicStatement(/* Allow "*" with StringLike {"s3:prefix": "public/*"} */)     // => true
PublicStatement(/* cloudfront.json: Allow {"Service": "cloudfront..."} */)     // => false
PublicStatement(/* https_only.json statement 0: Deny "*" */)                   // => false
```

**Think it through:**
- What can make you return `false` before looking at conditions?
- You only need the condition **keys**, not the operators or values. Which
  loop variables do you keep, and which become `_`?
- Which `strings` function compares ignoring case?

### 2. Decide a bucket's exposure

**Goal:** `Exposure(b)` returns `"public"`, `"blocked"` or `"private"`.

Work out two facts first:

- *public policy*: any statement in `b.Policy.Statement` is a
  `PublicStatement`;
- *public ACL*: `b.ACLGrantees` contains `AllUsers` or `AuthenticatedUsers`
  (constants in the starter).

Then:

| Result | When |
|---|---|
| `"public"` | public policy **and not** `RestrictPublicBuckets`, **or** public ACL **and not** `IgnorePublicAcls` |
| `"blocked"` | otherwise, if there is a public policy or a public ACL |
| `"private"` | neither |

`BlockPublicAcls` and `BlockPublicPolicy` don't change the result.

```go
Exposure(Bucket{Policy: publicWebsite}) // => "public"
Exposure(Bucket{Policy: publicWebsite, Settings: Settings{RestrictPublicBuckets: true}})
// => "blocked"
Exposure(Bucket{Policy: publicWebsite, Settings: Settings{BlockPublicAcls: true, BlockPublicPolicy: true}})
// => "public" (those two only stop new public policies and ACLs)
Exposure(Bucket{ACLGrantees: []string{AuthenticatedUsers}}) // => "public"
Exposure(Bucket{ACLGrantees: []string{AllUsers}, Settings: Settings{IgnorePublicAcls: true}})
// => "blocked"
Exposure(Bucket{ACLGrantees: []string{AllUsers}, Settings: Settings{RestrictPublicBuckets: true}})
// => "public" (RestrictPublicBuckets is about policies, not ACLs)
Exposure(Bucket{Policy: officeOnly, ACLGrantees: []string{ownerID}}) // => "private"
Exposure(Bucket{}) // => "private"
```

`publicWebsite` and `officeOnly` are the policies in `public_website.json` and
`office_only.json`; `ownerID` is the bucket owner's canonical user ID.

**Think it through:**
- Can you reuse your task 1 function for the policy fact?
- Write the `"public"` row as one boolean expression. A `switch` with no
  value after it checks `case` conditions in order.

### 3. Check that HTTPS is required

**Goal:** `RequiresTLS(p)` returns `true` when **some** statement in `p` denies insecure requests to everyone.

That statement must have all of:

- `Effect` `"Deny"`;
- `Principal.AWS` listing `"*"`;
- `Action` listing `"s3:*"` or `"*"` (one action such as `s3:GetObject`
  leaves uploads open over HTTP);
- `Condition["Bool"]["aws:SecureTransport"]` listing `"false"`.

This exercise doesn't check `Resource`. No such statement gives `false`.

```go
RequiresTLS(/* https_only.json: Deny "*" s3:* when SecureTransport "false" */) // => true
RequiresTLS(/* vpc_endpoint.json: same with Action "*" */)                     // => true
RequiresTLS(/* tls_get_only.json: Deny only s3:GetObject */)                   // => false
RequiresTLS(/* allow_secure.json: Allow when SecureTransport "true" */)        // => false
RequiresTLS(/* Deny s3:* for one role only */)                                 // => false
```

**Think it through:**
- Which helper from `slices` tells whether a `StringList` holds a value?
  (Add `"slices"` to the imports; you may have already for task 1.)
- How does the nested-map section let you read the condition value in one
  expression, even when a statement has no `Condition`?

## Learn more

- [Bucket policies for Amazon S3](https://docs.aws.amazon.com/AmazonS3/latest/userguide/bucket-policies.html)
- [Blocking public access to your Amazon S3 storage](https://docs.aws.amazon.com/AmazonS3/latest/userguide/access-control-block-public-access.html)
- [Examples of Amazon S3 bucket policies](https://docs.aws.amazon.com/AmazonS3/latest/userguide/example-bucket-policies.html)
- [Access control list (ACL) overview](https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl-overview.html)
