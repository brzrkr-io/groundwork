# Assume role: trust policies and temporary credentials

Long-lived access keys leak: into Git, CI logs, laptops. *Roles* are how AWS
avoids them. A program borrows a role for an hour and gets keys that expire on
their own. Here you check who may borrow a role, and when borrowed keys need
renewing.

**New here:** IAM roles, trust policies, `sts:AssumeRole`, temporary credentials, GitHub Actions OIDC (`sts:AssumeRoleWithWebIdentity`).

## Learn

### Recap: policies in Go

As in *IAM policies: how AWS decides allow or deny*, a policy is JSON with a
`Statement` list (always a list in this course), and `StringList` decodes
either `"x"` or `["x", "y"]`. The starter has them ready, plus `Match(pattern,
value)`: `*` matches any run of characters, `?` exactly one, case-sensitive.

```go
Match("repo:acme/*", "repo:acme/api") // => true
Match("s3:Get*", "s3:PutObject")      // => false
```

### Roles: an identity without a password

An IAM *user* has a password or *long-term access keys*: they work until
someone deletes them. A *role* has neither. Someone (a person, a program, an
AWS service) *assumes* the role and gets short-lived keys for it.

Every role has two policies:

| Policy | Answers | Example |
|---|---|---|
| *trust policy* | **who** may assume the role | "EC2 instances may", "account 444455556666 may" |
| *permissions policy* | **what** the role may do once assumed | "read `acme-app-logs`" |

The permissions policy is the kind you wrote before. This exercise is about the
trust policy.

### Trust policies and principals

A trust policy statement has a *Principal*: who it is about. Its `Action` is
`sts:AssumeRole` (STS, the *Security Token Service*, hands out the keys).

```json
{
  "Effect": "Allow",
  "Principal": { "Service": "ec2.amazonaws.com" },
  "Action": "sts:AssumeRole"
}
```

| Principal | Means |
|---|---|
| `{"AWS": "arn:aws:iam::111122223333:role/ci-runner"}` | exactly that role (or user) |
| `{"AWS": "arn:aws:iam::444455556666:root"}` | **any** identity in account `444455556666` (that account decides which of its users may, with its own policies) |
| `{"Service": "ec2.amazonaws.com"}` | the EC2 service, on behalf of your instances |
| `{"Federated": "arn:aws:iam::111122223333:oidc-provider/token.actions.githubusercontent.com"}` | someone holding a token from that identity provider (GitHub Actions) |
| `{"AWS": "*"}` | **anyone with an AWS account.** Never, unless a condition narrows it |

`:root` does not mean the root user here: it means "the whole account". The
account ID is the fifth `:`-separated part of an ARN, as in
*ARNs: accounts, regions and resource scope*.

In Go, the starter decodes `Principal` into three `StringList` fields: `AWS`,
`Service` and `Federated`. A field the JSON doesn't mention stays empty.

### Temporary credentials

`sts:AssumeRole` returns *temporary credentials*:

```json
{
  "Credentials": {
    "AccessKeyId": "ASIAIOSFODNN7EXAMPLE",
    "SecretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYzEXAMPLEKEY",
    "SessionToken": "IQoJb3JpZ2luX2VjEXAMPLETOKEN",
    "Expiration": "2026-10-13T13:00:00Z"
  }
}
```

| | Long-term key (IAM user) | Temporary key (role) |
|---|---|---|
| Key ID starts with | `AKIA` | `ASIA` |
| Session token | none | required with every call |
| Stops working | when someone deletes it | at `Expiration`, 15 minutes to 12 hours later |

Tools refresh temporary credentials **before** they expire, with a safety
margin, so a call that starts at 12:59:59 doesn't fail half-way. Durations are
as in *Time and durations: cert expiry, uptime, maintenance windows*:
`b.Sub(a)` is how long from `a` to `b`, negative if `b` is earlier.

### GitHub Actions: no stored keys at all

A deploy workflow used to keep an `AKIA…` key in GitHub secrets. With *OIDC*
(OpenID Connect) it stores nothing:

```text
GitHub Actions ──"I am repo acme/deploy, branch main" (signed token)──▶ AWS STS
AWS STS: is the token signed by GitHub? does the role's trust policy accept it?
AWS STS ──temporary credentials (1 hour)──▶ the workflow
```

The token carries *claims*. The two that matter:

| Condition key | Example value |
|---|---|
| `token.actions.githubusercontent.com:aud` | `sts.amazonaws.com` (who the token is for) |
| `token.actions.githubusercontent.com:sub` | `repo:acme/deploy:ref:refs/heads/main` (which repo and branch) |

Other `sub` values: `repo:acme/deploy:pull_request`,
`repo:acme/deploy:ref:refs/tags/v1.2.0`. The trust policy uses the action
`sts:AssumeRoleWithWebIdentity` and a `Condition`:

```json
"Condition": {
  "StringEquals": { "token.actions.githubusercontent.com:aud": "sts.amazonaws.com" },
  "StringLike":   { "token.actions.githubusercontent.com:sub": "repo:acme/*" }
}
```

`StringEquals` compares exactly; `StringLike` allows `*` and `?`. Every key
must hold, and a key holds when **any** of its values matches. Leave out the
`sub` condition and **every GitHub repository in the world** can assume your
role.

### Nested maps

`Condition` is a map of maps: operator → condition key → values.

```go
limits := map[string]map[string]int{
	"prod":    {"cpu": 4, "memory": 8},
	"staging": {"cpu": 1},
}
for env, perResource := range limits {
	for resource, n := range perResource {
		fmt.Println(env, resource, n)
	}
}
```

> **Why is it written like this?** `map[string]map[string]int` is Go's
> syntax for "a map whose values are maps"; the compiler needs the full type.
> Inside the literal you can drop the inner type (`{"cpu": 4}`): Go allows
> that and gofmt prefers it. Map order is random, so only write loops whose
> result doesn't depend on it ("do *all* keys hold?" is fine).

### Try it (optional)

Needs an AWS account and credentials; never needed for the tests.

```bash
aws sts get-caller-identity
aws iam list-roles --query 'Roles[].RoleName'
aws iam get-role --role-name ROLE_NAME --query 'Role.AssumeRolePolicyDocument'
```

- `get-caller-identity` shows who you are now; after assuming a role, the ARN
  says `assumed-role/…`.
- `list-roles` lists the roles in your account.
- `get-role` with `--query` prints just a role's trust policy.

## Your task

Your team is moving CI to GitHub Actions and reviewing the roles in account
`111122223333`. The trust policies are in `testdata/` (`ec2_instance.json`,
`cross_account.json`, `github_main.json`, …): open them.

### 1. Spot temporary credentials

**Goal:** `IsTemporary(c)` returns `true` when `c.AccessKeyID` starts with `"ASIA"` **and** `c.SessionToken` isn't empty.

The prefix is case-sensitive.

```go
IsTemporary(Credentials{AccessKeyID: "ASIAIOSFODNN7EXAMPLE", SessionToken: "IQoJb3JpZ2lu…"}) // => true
IsTemporary(Credentials{AccessKeyID: "AKIAIOSFODNN7EXAMPLE"})                              // => false
IsTemporary(Credentials{AccessKeyID: "ASIAIOSFODNN7EXAMPLE"})                              // => false (no token)
```

**Think it through:**
- Which `strings` function checks how a string starts?
- Both conditions must be true: which operator joins them?

### 2. Refresh before expiry

**Goal:** `NeedsRefresh(c, now)` returns `true` when less than five minutes are left until `c.Expiration`, or it has passed.

Exactly five minutes left is still fine (`false`).

```go
// now = 12:00:00
NeedsRefresh(Credentials{Expiration: /* 13:00:00 */}, now) // => false
NeedsRefresh(Credentials{Expiration: /* 12:05:00 */}, now) // => false
NeedsRefresh(Credentials{Expiration: /* 12:04:59 */}, now) // => true
NeedsRefresh(Credentials{Expiration: /* 11:59:00 */}, now) // => true (expired)
```

**Think it through:**
- Which method gives the time left as a `time.Duration`? What is it when the
  credentials already expired?
- How do you write "five minutes" as a `time.Duration`?

### 3. Check who may assume a role

**Goal:** `TrustsCaller(s, caller)` returns `true` when statement `s` lets `caller` call `sts:AssumeRole`.

`caller` is an IAM ARN (`arn:aws:iam::444455556666:role/auditor`) or a service
name (`ec2.amazonaws.com`). All three must hold:

- `s.Effect` is `"Allow"`.
- Some entry of `s.Action` matches `sts:AssumeRole`, ignoring case: use
  `Match` on both lowercased. So `sts:AssumeRole`, `STS:assumerole`,
  `sts:*` and `*` count; `sts:AssumeRoleWithWebIdentity` and
  `sts:TagSession` don't.
- The principal trusts the caller. That's the case when **any** of these holds:
  - `caller` is listed in `s.Principal.AWS` or `s.Principal.Service`
    (exact, case-sensitive);
  - `s.Principal.AWS` lists `"*"`;
  - `s.Principal.AWS` lists `arn:aws:iam::ACCOUNT:root`, where ACCOUNT is
    the fifth `:`-separated part of `caller`. A service name has no account.

```go
TrustsCaller(ec2Instance, "ec2.amazonaws.com")                         // => true
TrustsCaller(crossAccount, "arn:aws:iam::444455556666:user/bob")       // => true (account root)
TrustsCaller(crossAccount, "arn:aws:iam::123456789012:role/auditor")   // => false (other account)
TrustsCaller(ciRunner, "arn:aws:iam::111122223333:user/bob")           // => false (same account, not listed)
TrustsCaller(anyone, "arn:aws:iam::123456789012:user/mallory")         // => true (that's the danger)
```

`ec2Instance`, `crossAccount`, `ciRunner` and `anyone` are the statements in
`testdata/ec2_instance.json`, `cross_account.json`, `ci_runner.json` and
`anyone.json`.

A statement with `"Effect": "Deny"`, or whose actions are only
`sts:AssumeRoleWithWebIdentity`, gives `false` for every caller.

**Think it through:**
- Which function from the `slices` package checks whether a list contains a
  value? (Add `"slices"` to the imports.)
- How do you get the fifth part of `"arn:aws:iam::444455556666:user/bob"`, and
  build the root ARN from it? What happens with `"ec2.amazonaws.com"`?

### 4. Check which workflows may deploy

**Goal:** `TrustsWorkflow(s, provider, claims)` returns `true` when statement `s` lets a token from `provider` with these `claims` call `sts:AssumeRoleWithWebIdentity`.

`claims` is keyed by the full condition key. All of these must hold:

- `s.Effect` is `"Allow"`.
- Some entry of `s.Action` matches `sts:AssumeRoleWithWebIdentity`, ignoring
  case (as in task 3).
- `s.Principal.Federated` lists `provider` exactly.
- Every condition holds: for each operator and each key under it, the claim
  with that key exists, and at least one of the key's values matches it:
  - `StringEquals`: equal, case-sensitive;
  - `StringLike`: `Match(value, claim)`;
  - any other operator never matches, so it gives `false`.
- No `Condition` at all means nothing to check: `true`.

```go
provider := "arn:aws:iam::111122223333:oidc-provider/token.actions.githubusercontent.com"
claims := map[string]string{
	"token.actions.githubusercontent.com:aud": "sts.amazonaws.com",
	"token.actions.githubusercontent.com:sub": "repo:acme/deploy:ref:refs/heads/main",
}
TrustsWorkflow(githubMain, provider, claims) // => true
// same claims, but sub = "repo:acme/deploy:pull_request":     false (a PR can't deploy)
// same claims, but sub = "repo:evil/deploy:ref:refs/heads/main": false (someone else's repo)
// same claims, but aud = "https://github.com/acme":           false
// claims without the aud key:                                 false
TrustsWorkflow(githubAnyRef, provider, claims) // => true; with sub "repo:acme/deploy:pull_request" also true
// githubAnyRef with sub "repo:acme/deploy-tools:ref:refs/heads/main": false
TrustsWorkflow(githubNoSub, provider, /* claims from repo:evil/deploy */) // => true (the danger)
```

`githubMain`, `githubAnyRef` and `githubNoSub` are the statements in
`testdata/github_main.json`, `github_any_ref.json` and `github_no_sub.json`.

**Think it through:**
- Return `false` as soon as one check fails; what's left to return after the
  loops?
- For each key, look up the claim with the two-value form (`claim, ok :=`).
  Where does the operator decide how to compare?
- A small helper function for "does any value match?" keeps the loops short.

## Learn more

- [IAM roles](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html)
- [Policy elements: Principal](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html)
- [Temporary security credentials in IAM](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp.html)
- [Create a role for OpenID Connect federation](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-idp_oidc.html)
