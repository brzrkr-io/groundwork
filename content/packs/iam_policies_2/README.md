# IAM policies: write least-privilege policies

A leaked key or a buggy script can only do what its policy allows. So every
job gets a policy that allows exactly what it does and nothing more. Here you
write three of them, as JSON files, the way they are reviewed in a real
pull request.

**New here:** least privilege, bucket vs object ARNs, IAM conditions, `aws:SourceIp`.

## Learn

### Recap

As in *IAM policies: how AWS decides allow or deny*: a statement matches a
request when an `Action` pattern matches the action (case ignored) and a
`Resource` pattern matches the ARN. A matching `Deny` wins, then a matching
`Allow`; otherwise the answer is the implicit deny. `*` matches any run of
characters, `?` exactly one.

### Least privilege

*Least privilege* means: start from nothing, and allow only the calls the job
really makes, on only the resources it really touches. Three habits get you
there:

| Habit | Too broad | Least privilege |
|---|---|---|
| Name the actions | `"s3:*"` | `"s3:GetObject"` |
| Name the resources | `"*"` | `"arn:aws:s3:::acme-app-logs/prod/*"` |
| Narrow the callers | anyone holding the key | only from the office network (a condition) |

`"s3:*"` is worse than it looks: it also covers every S3 action AWS adds in
the future.

### Bucket ARN or object ARN?

In S3, a *bucket* holds *objects* (files). Each object has a *key*, its path
in the bucket, like `prod/web/2026-10-13.log`. The bucket and its objects
have **different ARNs**, and each S3 action works on one or the other:

| ARN | Example | Actions that use it |
|---|---|---|
| bucket | `arn:aws:s3:::acme-app-logs` | `s3:ListBucket` (list keys), `s3:PutBucketPolicy`, `s3:DeleteBucket` |
| object | `arn:aws:s3:::acme-app-logs/prod/web/2026-10-13.log` | `s3:GetObject`, `s3:PutObject`, `s3:DeleteObject`, `s3:PutObjectAcl` |

`arn:aws:s3:::acme-app-logs/*` matches every object but **not** the bucket
itself. That's why a policy that lists and reads usually has two statements,
one per kind of ARN. Also, a pattern like `arn:aws:s3:::acme-app-logs*` (no
slash) matches someone else's bucket `acme-app-logs-archive` too.

### Conditions

A statement can have a `Condition` block: extra tests on the request, beyond
action and resource. It is a map from *operator* to *condition key* to
values:

```json
"Condition": {
  "StringEquals": {"aws:RequestedRegion": ["eu-west-1", "eu-central-1"]},
  "Bool": {"aws:MultiFactorAuthPresent": "true"}
}
```

That reads: the call goes to `eu-west-1` **or** `eu-central-1`, **and** the
caller logged in with MFA (a second factor, like a phone code). The rules:

- **Every** key under every operator must hold (AND).
- A key holds when **any** of its values matches (OR).
- A key the request doesn't carry makes its condition fail (`NotIpAddress`
  is the exception: a missing address counts as "not in the range").

The operators these tests support:

| Operator | Holds when the request's value... | Example value |
|---|---|---|
| `StringEquals` | is exactly one of the values | `"eu-west-1"` |
| `StringLike` | matches one of the values, with `*` and `?` | `"repo:acme/*"` |
| `IpAddress` | is an IP address inside one of the ranges | `"192.0.2.0/24"`, `"192.0.2.7"` |
| `NotIpAddress` | is inside **none** of the ranges | `"192.0.2.0/24"` |
| `Bool` | is `true` or `false`, as given | `"true"` |

A *condition key* names something AWS knows about each request. The one you
need here is `aws:SourceIp`: the IP address the call came from. CIDR ranges
work as in *IP addresses: classify, allowlist, overlap*. (Calls that an AWS
service makes for you, or that arrive through a private VPC endpoint, don't
carry your public address; real policies handle those separately.)

> **Why is it written like this?** IAM requires the nesting operator → key →
> values. JSON objects can't repeat a key, so to test two keys with the same
> operator you put both inside one operator object. A single value can be a
> plain string or a one-item list; both mean the same.

### Writing a policy file

JSON has no comments, so each starter keeps its task in a top-level `"Id"`
field, a label IAM allows only in *resource-based* policies (the kind you
attach to a bucket, coming in *S3 bucket policies: is this bucket public?*).
Your policies attach to users and roles, so **delete the `Id` line** once the
task is done. The tests check that it's gone.

In a real repository these files live next to the Terraform or
CloudFormation that creates the role. You can also paste them into the
console or pass them to `aws iam create-policy --policy-document file://...`.

### How the tests read your policy

`policy_test.go` holds a small evaluator with exactly the rules above; read it
if a result surprises you. For every file it checks:

- The file is valid JSON. A syntax error is reported with the file name and
  line.
- It only uses the fields `Version`, `Id`, `Statement`, and in each
  statement `Sid`, `Effect`, `Action`, `Resource` and `Condition`. Anything
  else, often a typo like `"Actions"`, is an error.
- `Version` is `"2012-10-17"`.
- There is at least one statement, and no `Id` is left.
- Every statement has an `Effect` of `"Allow"` or `"Deny"`, an `Action` and
  a `Resource`.
- No `Allow` statement has the action `"*"` or `"<service>:*"`, like
  `"s3:*"`.
- Conditions only use the five operators in the table, and `IpAddress` /
  `NotIpAddress` values are addresses or CIDR ranges.
- Every request in the task's lists gets the right answer. "Denied" means
  either kind of deny.

### Try it (optional)

Needs an AWS account and credentials; never needed for the tests.

```sh
aws accessanalyzer validate-policy --policy-type IDENTITY_POLICY \
  --policy-document file://1_backup_writer.json        # AWS's own review: errors, warnings, suggestions
aws iam simulate-custom-policy --policy-input-list file://3_deploy_bot.json \
  --action-names ec2:StopInstances \
  --resource-arns arn:aws:ec2:eu-west-1:111122223333:instance/i-0abc1234def567890 \
  --context-entries ContextKeyName=aws:SourceIp,ContextKeyValues=203.0.113.25,ContextKeyType=ip
                                                       # AWS's evaluator, for a policy that isn't attached anywhere
```

Both commands only read. Remove the `Id` line first, or `validate-policy`
reports it too.

## Your task

Three jobs need policies. For each, write the statements into its file and
delete the `Id` line. Every request listed must get the answer shown.

### 1. Let the backup job upload

**Goal:** in `1_backup_writer.json`, allow `s3:PutObject` on `arn:aws:s3:::acme-db-backups/nightly/*` and nothing else.

The nightly database backup uploads one file per night under `nightly/`, from
the backup server `192.0.2.10`.

Allowed:

| Action | Resource |
|---|---|
| `s3:PutObject` | `arn:aws:s3:::acme-db-backups/nightly/2026-10-13.sql.gz` |
| `s3:PutObject` | `arn:aws:s3:::acme-db-backups/nightly/orders/2026-10-13.sql.gz` |

Denied:

| Action | Resource | Why |
|---|---|---|
| `s3:GetObject` | `arn:aws:s3:::acme-db-backups/nightly/2026-10-13.sql.gz` | the job only writes |
| `s3:DeleteObject` | `arn:aws:s3:::acme-db-backups/nightly/2026-10-12.sql.gz` | a stolen key must not wipe backups |
| `s3:PutObject` | `arn:aws:s3:::acme-db-backups/weekly/2026-10-11.sql.gz` | not its folder |
| `s3:PutObject` | `arn:aws:s3:::acme-db-backups-old/nightly/2026-10-13.sql.gz` | a different bucket |
| `s3:PutObjectAcl` | `arn:aws:s3:::acme-db-backups/nightly/2026-10-13.sql.gz` | could make a backup public |
| `s3:PutBucketPolicy` | `arn:aws:s3:::acme-db-backups` | could open the whole bucket |

**Think it through:**
- Does `s3:Put*` pass this list? Which row stops it?
- Which part of the resource pattern keeps `weekly/` and `acme-db-backups-old` out?

### 2. Let the log reader read prod logs

**Goal:** in `2_log_reader.json`, allow `s3:ListBucket` on the bucket `acme-app-logs` and `s3:GetObject` on its objects under `prod/`.

A support tool lists the log bucket and downloads production logs, from
`198.51.100.7`. Staging logs may contain test customer data, so they stay out.

Allowed:

| Action | Resource |
|---|---|
| `s3:ListBucket` | `arn:aws:s3:::acme-app-logs` |
| `s3:GetObject` | `arn:aws:s3:::acme-app-logs/prod/web/2026-10-13.log` |
| `s3:GetObject` | `arn:aws:s3:::acme-app-logs/prod/api/2026-10-12.log` |

Denied:

| Action | Resource |
|---|---|
| `s3:GetObject` | `arn:aws:s3:::acme-app-logs/staging/web/2026-10-13.log` |
| `s3:PutObject` | `arn:aws:s3:::acme-app-logs/prod/web/2026-10-13.log` |
| `s3:DeleteObject` | `arn:aws:s3:::acme-app-logs/prod/web/2026-10-13.log` |
| `s3:ListBucket` | `arn:aws:s3:::acme-db-backups` |
| `s3:GetObject` | `arn:aws:s3:::acme-app-logs-archive/prod/web/2026-01-01.log` |

**Think it through:**
- Which of the two actions takes the bucket ARN, and which an object ARN?
- One statement or two? What would one statement with both actions and
  both resources allow that you didn't mean?

### 3. Let the deploy bot restart servers from the office

**Goal:** in `3_deploy_bot.json`, allow starting, stopping and rebooting instances in `eu-west-1` of account `111122223333`, and describing instances, all only from `203.0.113.0/24`.

The deploy bot runs on a machine in the office network `203.0.113.0/24`.

- The actions are `ec2:StartInstances`, `ec2:StopInstances` and
  `ec2:RebootInstances` on `arn:aws:ec2:eu-west-1:111122223333:instance/*`.
- It also needs `ec2:DescribeInstances`. Describe actions can't be limited to
  particular resources, so AWS checks them against `"*"`: their `Resource`
  must be `"*"`.
- **Every** allow only holds when `aws:SourceIp` is in `203.0.113.0/24`.

Allowed (`web` = `arn:aws:ec2:eu-west-1:111122223333:instance/i-0abc1234def567890`):

| Action | Resource | From |
|---|---|---|
| `ec2:StartInstances` | `web` | `203.0.113.25` |
| `ec2:StopInstances` | `web` | `203.0.113.25` |
| `ec2:RebootInstances` | `web` | `203.0.113.25` |
| `ec2:RebootInstances` | `arn:aws:ec2:eu-west-1:111122223333:instance/i-0fed9876cba543210` | `203.0.113.200` |
| `ec2:DescribeInstances` | `*` | `203.0.113.200` |

Denied:

| Action | Resource | From |
|---|---|---|
| `ec2:TerminateInstances` | `web` | `203.0.113.25` |
| `ec2:RunInstances` | `arn:aws:ec2:eu-west-1:111122223333:instance/*` | `203.0.113.25` |
| `ec2:StartInstances` | `arn:aws:ec2:us-east-1:111122223333:instance/i-0abc1234def567890` | `203.0.113.25` |
| `ec2:StartInstances` | `arn:aws:ec2:eu-west-1:444455556666:instance/i-0abc1234def567890` | `203.0.113.25` |
| `ec2:DeleteVolume` | `arn:aws:ec2:eu-west-1:111122223333:volume/vol-0123456789abcdef0` | `203.0.113.25` |
| `ec2:StopInstances` | `web` | `198.51.100.7` |
| `ec2:DescribeInstances` | `*` | `198.51.100.7` |
| `ec2:StartInstances` | `web` | `192.0.2.44` |

**Think it through:**
- Which parts of the instance ARN pin the region and the account?
- Which operator and key say "only from this network", and does the
  describe statement need the condition too?

## Learn more

- [IAM JSON policy elements: Condition](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_condition.html)
- [IAM JSON policy elements: Condition operators](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_condition_operators.html)
- [Identity-based policy examples for Amazon S3](https://docs.aws.amazon.com/AmazonS3/latest/userguide/example-policies-s3.html)
- [Security best practices in IAM](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html): apply least-privilege permissions
