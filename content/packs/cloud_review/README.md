# Practice: review an AWS account

A new team hands you their AWS account: "can you take a look before we go
live?" You write a small review tool that reads an exported snapshot of the
account and lists what's dangerous, most urgent first. Real tools (AWS
Security Hub, Prowler, ScoutSuite) run hundreds of checks like these.

**New here:** `time.Time.IsZero`; the rest is practice of everything in *Cloud fundamentals*.

## Learn

### Recap

| You need | Where you saw it |
|---|---|
| Security group rules, ports and CIDRs | *Security groups and network ACLs: will this packet get through?* |
| Policy statements, `StringList`, `"*"` and `service:*` | *IAM policies: how AWS decides allow or deny* |
| Why `"Action": "*"` on `"Resource": "*"` is dangerous | *IAM policies: fix the over-broad policy* |
| Bucket policies, `Principal`, Block Public Access | *S3 bucket policies: is this bucket public?* |
| Long-lived `AKIA` keys versus temporary credentials | *Assume role: trust policies and temporary credentials* |
| `Sub`, `Hours` on times and durations | *Time and durations: cert expiry, uptime, maintenance windows* |
| `slices.SortFunc`, `cmp.Compare`, `slices.Clone` | *Sorting*, *append and copy* |

### The zero time

A `time.Time` field that JSON leaves out keeps its *zero value*: January 1,
year 1. Don't compare it with a date you make up; ask it:

```go
var deployed time.Time // never set
deployed.IsZero()      // => true
time.Now().IsZero()    // => false
```

### The snapshot

`testdata/account.json` is what an export script might save: the account's
security groups, IAM policies, S3 buckets and access keys. Open it and read
it before you start: you'll find most of the problems by eye. `Load` (already
written) decodes it into an `Account`, and `Review` (already written) runs
your four checks and passes everything they find to `Report`.

```text
Review(path, now)
  ├─ Load(path)                 → Account
  ├─ OpenAdminPorts(groups)     ┐
  ├─ BroadPolicies(policies)    │ each returns []Finding
  ├─ PublicBuckets(buckets)     │
  ├─ StaleKeys(keys, now)       ┘
  └─ Report(all findings)       → the text below
```

Each check returns `Finding` values: a `Severity` (`"high"`, `"medium"` or
`"low"`), the `Resource` it's about, and the `Problem` in words. Return `nil`
when a check finds nothing. Keep findings in input order inside each check;
`Report` sorts them.

| File | What it holds |
|---|---|
| `exercise.go` | The types, `Load`, `Review`, and the five functions you write. |
| `exercise_test.go` | One test per task; the last one reviews `testdata/account.json`. |

### Try it (optional)

With an AWS account and credentials, these read-only commands show the same
settings for real (never needed for the tests):

```bash
aws ec2 describe-security-groups --query 'SecurityGroups[].{id:GroupId,name:GroupName}'
aws iam list-policies --scope Local
aws s3api get-public-access-block --bucket acme-website
aws iam generate-credential-report && aws iam get-credential-report
```

The credential report is a CSV of every IAM user with their key ages and last
use, which is what task 4 checks.

## Your task

### 1. Find admin ports open to the internet

**Goal:** `OpenAdminPorts(groups)` returns a `"high"` finding for each admin
port that an inbound rule opens to the whole internet.

- The admin ports are `22` (SSH) and `3389` (RDP), checked in that order.
- "The whole internet" is a rule whose `CIDR` is `0.0.0.0/0` or `::/0`.
- A rule opens a port when its `Protocol` is `"tcp"` and the port is between
  `FromPort` and `ToPort` (both included), or when its `Protocol` is `"-1"`
  (all traffic, whatever the ports say). `"udp"` rules never count.
- `Resource` is the group's ID, a space and its name in brackets;
  `Problem` is `port <port> open to <cidr>`.

```go
OpenAdminPorts([]SecurityGroup{{ID: "sg-02", Name: "win", Inbound: []Rule{
	{Protocol: "tcp", FromPort: 22, ToPort: 22, CIDR: "::/0"},
	{Protocol: "tcp", FromPort: 3000, ToPort: 4000, CIDR: "0.0.0.0/0"},
}}})
// => {"high", "sg-02 (win)", "port 22 open to ::/0"},
//    {"high", "sg-02 (win)", "port 3389 open to 0.0.0.0/0"}
```

**Think it through:**
- You need three loops: groups, their rules, and the two ports. Which check can skip a rule before the port loop?
- How do you build `"port 22 open to ::/0"` from a number and a string?

### 2. Find over-broad policies

**Goal:** `BroadPolicies(policies)` returns a finding for each wildcard
action in an `Allow` statement whose resources include `"*"`.

- Only statements with `Effect` `"Allow"` and `"*"` among their `Resource`
  values count.
- An action of exactly `"*"` is `"high"`, with the problem
  `allows every action on every resource`.
- An action ending in `:*` (like `s3:*`) is `"medium"`, with the problem
  `allows s3:* on every resource`.
- Other actions, including `ec2:Describe*`, are fine: they don't end in `:*`.
- `Resource` is `policy <name>`. One finding per matching action, in order.

```go
// policy "data-team": Allow ["s3:*", "athena:*"] on ["arn:aws:s3:::acme-data-lake", "*"]
// => {"medium", "policy data-team", "allows s3:* on every resource"},
//    {"medium", "policy data-team", "allows athena:* on every resource"}
```

**Think it through:**
- Which `slices` function tells you whether `"*"` is among a statement's resources?
- Which `strings` function checks the end of an action? Add the imports you need.

### 3. Find public buckets

**Goal:** `PublicBuckets(buckets)` returns a `"high"` finding for each
statement that makes a bucket public.

This review uses a simpler rule than *S3 bucket policies: is this bucket
public?*. A statement makes the bucket public when all of these hold:

- the bucket's `BlockPublicAccess` is `false` (when it's `true`, skip the
  whole bucket);
- `Effect` is `"Allow"`;
- `Principal.AWS` contains `"*"` (both `"Principal": "*"` and
  `"Principal": {"AWS": "*"}` decode to that);
- the statement has no `Condition` at all.

`Resource` is `bucket <name>`; `Problem` is `anyone can ` followed by the
statement's actions joined with `", "`.

```go
// acme-old-assets: Allow, Principal {"AWS": "*"}, ["s3:GetObject", "s3:ListBucket"]
// => {"high", "bucket acme-old-assets", "anyone can s3:GetObject, s3:ListBucket"}
```

**Think it through:**
- What does `len` return for a nil map, and how does that tell you there's no condition?
- Which `strings` function joins a slice with a separator?

### 4. Find stale access keys

**Goal:** `StaleKeys(keys, now)` returns findings for active keys that are
too old or never used.

- Skip keys whose `Status` isn't `"Active"`.
- The key's age is the time from `Created` to `now`, in whole days, rounded
  down (23 hours is 0 days). Over 90 days is `"medium"`:
  `access key <id> is <days> days old`. Exactly 90 is fine.
- A key whose `LastUsed` is the zero time (`IsZero` is true) has never been used: `"low"`,
  `access key <id> has never been used`.
- A key can get both findings, the age one first. `Resource` is
  `user <name>`.

```go
now := time.Date(2026, 10, 13, 9, 0, 0, 0, time.UTC)
// old-jenkins, created 2024-11-05T15:45:00Z, never used
// => {"medium", "user old-jenkins", "access key AKIAOLDJENKINSEXAMPL is 706 days old"},
//    {"low", "user old-jenkins", "access key AKIAOLDJENKINSEXAMPL has never been used"}
```

**Think it through:**
- `now.Sub(k.Created)` is a `time.Duration`. How do you turn it into whole days?
- Converting a `float64` to `int` drops the fraction. Is that the rounding you want?

### 5. Write the report

**Goal:** `Report(findings)` returns one line per finding, most urgent
first, then a line of counts.

- Sort a copy of `findings` (don't reorder the caller's slice): `high`, then
  `medium`, then `low`; within a severity by `Resource`, then by `Problem`,
  both alphabetically.
- Each line is `[<severity>] <resource>: <problem>`.
- The last line is `<n> high, <n> medium, <n> low`.
- Every line, the last one too, ends with `\n`.

```go
Report([]Finding{
	{"low", "user b", "z"},
	{"high", "bucket a", "anyone can s3:GetObject"},
})
// => "[high] bucket a: anyone can s3:GetObject\n[low] user b: z\n1 high, 0 medium, 1 low\n"
Report(nil) // => "0 high, 0 medium, 0 low\n"
```

The last test runs `Review` on `testdata/account.json` and expects 16 lines,
ending with `8 high, 5 medium, 2 low`.

**Think it through:**
- Severity words don't sort alphabetically in the order you want. What can you compare instead?
- How do you compare by a second field only when the first is equal?

## Learn more

- [Security best practices in IAM](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html)
- [Control traffic to your AWS resources using security groups](https://docs.aws.amazon.com/vpc/latest/userguide/vpc-security-groups.html)
- [Blocking public access to your Amazon S3 storage](https://docs.aws.amazon.com/AmazonS3/latest/userguide/access-control-block-public-access.html)
- [Generate credential reports for your AWS account](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_getting-report.html)
