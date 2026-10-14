# IAM policies: fix the over-broad policy

Policies go wrong in three ways: they allow too much, they allow too little,
or a guardrail that should stop something doesn't. Each file here has one of
those bugs, found in a review or an incident. Read the evidence, find the
statement, fix only that.

**New here:** IAM last accessed information; otherwise practice of diagnosing policy bugs.

## Learn

### Recap

As in *IAM policies: write least-privilege policies*: a statement matches when
an action pattern matches (case ignored), a resource pattern matches (case
counts) and every condition holds. A matching `Deny` wins, then a matching
`Allow`, else the implicit deny. Bucket actions such as `s3:ListBucket` take
the bucket ARN `arn:aws:s3:::name`; object actions such as `s3:GetObject`
take object ARNs `arn:aws:s3:::name/key`.

### Diagnosing a policy bug

1. **Read the symptom.** An access-denied message names the action and the
   resource. A "should have been denied" report names the call that got
   through.
2. **Find the statement that should decide it.** Which `Sid` was written for
   this call?
3. **Compare, character by character.** Put the request's action and ARN next
   to the statement's patterns. Most bugs are one of these:

| Bug | Looks like | Effect |
|---|---|---|
| bucket ARN where an object ARN belongs | `arn:aws:s3:::reports` with `s3:GetObject` | reads denied |
| object ARN where the bucket belongs | `arn:aws:s3:::reports/*` with `s3:ListBucket` | listing denied |
| missing slash | `arn:aws:s3:::reports*` | also matches the bucket `reports-archive` |
| wrong region or account | `arn:aws:ec2:us-east-1:...` | denied in the region you use |
| wildcards everywhere | `"s3:*"` on `"*"` | allows everything, forever |

4. **Fix only that, then ask what else your change allows.** Widening a
   pattern to make one call work often lets in calls nobody wanted.

### Reading the error messages

AWS error messages name the IAM action, not always the API call you made:

```text
$ aws s3 ls s3://acme-photos/
An error occurred (AccessDenied) when calling the ListObjectsV2 operation:
User: arn:aws:sts::111122223333:assumed-role/photo-viewer/alice is not
authorized to perform: s3:ListBucket on resource: "arn:aws:s3:::acme-photos"
because no identity-based policy allows the s3:ListBucket action
```

The API call is `ListObjectsV2`; the permission it needs is `s3:ListBucket`,
on the **bucket** ARN. Downloads are terser: `aws s3 cp` first asks for the
object's details (`HeadObject`), and S3 answers with a bare `403 Forbidden`.
`HeadObject` needs `s3:GetObject` on the **object** ARN.

### Finding what a role really uses

To narrow an over-broad policy, you need the calls the role really makes.
Two places record them:

- *IAM last accessed information*: which services the role used, and when,
  over at least 400 days. For S3, EC2, IAM and Lambda it also lists the
  *management actions* (changes to settings, such as `s3:PutBucketPolicy`).
- *CloudTrail*: one record per API call, with the resource. Object reads and
  writes (`s3:GetObject`, `s3:PutObject`) are *data events*, recorded only
  when the trail is set up for them. *CloudTrail: trace a leaked access key*
  reads those records.

Allow what the evidence shows, nothing else. If a rare job breaks later, its
access-denied message tells you exactly what to add.

### Guardrails must be explicit denies

A *guardrail* is a `Deny` that must hold no matter what. Removing an action
from an `Allow` gives an implicit deny, and that is weak: tomorrow someone
attaches a broad policy like AWS's `AmazonS3FullAccess` to the same role, and
the implicit deny is gone. An explicit `Deny` beats every `Allow`, in any
policy. So when a guardrail leaks, fix the guardrail; don't just trim the
`Allow`.

### Try it (optional)

Needs an AWS account and credentials; never needed for the tests.

```sh
aws iam generate-service-last-accessed-details --granularity ACTION_LEVEL \
  --arn arn:aws:iam::111122223333:role/ci-artifacts    # starts a report; prints a JobId
aws iam get-service-last-accessed-details --job-id <JobId>   # which services and actions the role used, and when
aws iam simulate-custom-policy --policy-input-list file://2_report_reader.json \
  --action-names s3:ListBucket --resource-arns arn:aws:s3:::acme-reports   # check your fix the way AWS does
```

None of them changes anything. Use a role from your own account.

## Your task

Three policies, three bugs. Each starter's `"Id"` holds the symptom. Fix the
statements and delete the `Id` line. The tests check the same rules as in
*IAM policies: write least-privilege policies*: valid JSON with only the
known fields, `Version` `"2012-10-17"`, no `Id` left, every statement with
`Effect` `"Allow"` or `"Deny"` plus an `Action` and a `Resource`, no `Allow`
of `"*"` or `"<service>:*"`, and every request below decided as shown.
"Denied" means either kind of deny. `policy_test.go` holds the evaluator.

### 1. Narrow the CI policy

**Goal:** in `1_ci_artifacts.json`, replace `"s3:*"` on `"*"` with exactly what CI uses.

The CI role's policy allows every S3 action on every bucket. The evidence
from the last 90 days: IAM last accessed information shows S3 is the only
service the role used, and no S3 management action at all. CloudTrail's data
events for the role:

| Action | Calls | Resources |
|---|---|---|
| `s3:GetObject` | 9,377 | `arn:aws:s3:::acme-ci-artifacts/builds/...` |
| `s3:PutObject` | 4,812 | `arn:aws:s3:::acme-ci-artifacts/builds/...` |
| anything else | 0 | |

Allowed (`build` = `arn:aws:s3:::acme-ci-artifacts/builds/1234/app.tar.gz`):

| Action | Resource |
|---|---|
| `s3:PutObject` | `build` |
| `s3:GetObject` | `arn:aws:s3:::acme-ci-artifacts/builds/1233/app.tar.gz` |

Denied:

| Action | Resource |
|---|---|
| `s3:DeleteObject` | `arn:aws:s3:::acme-ci-artifacts/builds/1200/app.tar.gz` |
| `s3:PutObjectAcl` | `build` |
| `s3:PutObject` | `arn:aws:s3:::acme-ci-artifacts/releases/v1.4.0/app.tar.gz` |
| `s3:PutObject` | `arn:aws:s3:::acme-ci-artifacts-public/builds/1234/app.tar.gz` |
| `s3:GetObject` | `arn:aws:s3:::acme-db-backups/nightly/2026-10-13.sql.gz` |
| `s3:PutBucketPolicy` | `arn:aws:s3:::acme-ci-artifacts` |
| `s3:DeleteBucket` | `arn:aws:s3:::acme-ci-artifacts` |
| `s3:ListAllMyBuckets` | `*` |

**Think it through:**
- Which two actions does the evidence show, and on which part of which
  bucket?
- Does your resource pattern let in `releases/` or `acme-ci-artifacts-public`?

### 2. Fix the report reader

**Goal:** in `2_report_reader.json`, make listing and downloading reports from `acme-reports` work.

The reporting team's tool can't do anything:

```text
$ aws s3 ls s3://acme-reports/2026/
An error occurred (AccessDenied) when calling the ListObjectsV2 operation:
User: arn:aws:sts::111122223333:assumed-role/report-reader/report-tool is not
authorized to perform: s3:ListBucket on resource: "arn:aws:s3:::acme-reports"
because no identity-based policy allows the s3:ListBucket action

$ aws s3 cp s3://acme-reports/2026/09/summary.pdf .
fatal error: An error occurred (403) when calling the HeadObject operation: Forbidden
```

Keep the two statements, `ListReports` and `ReadReports`, and keep their
actions. Allowed:

| Action | Resource |
|---|---|
| `s3:ListBucket` | `arn:aws:s3:::acme-reports` |
| `s3:GetObject` | `arn:aws:s3:::acme-reports/2026/09/summary.pdf` |
| `s3:GetObject` | `arn:aws:s3:::acme-reports/2026/10/costs.csv` |

Denied:

| Action | Resource |
|---|---|
| `s3:PutObject` | `arn:aws:s3:::acme-reports/2026/10/costs.csv` |
| `s3:DeleteObject` | `arn:aws:s3:::acme-reports/2026/09/summary.pdf` |
| `s3:GetObject` | `arn:aws:s3:::acme-reports-archive/2020/summary.pdf` |
| `s3:ListBucket` | `arn:aws:s3:::acme-reports-archive` |

**Think it through:**
- The first error names an action and a resource. Which statement should
  match them, and what does its `Resource` say instead?
- `HeadObject` needs `s3:GetObject` on an object ARN. Does `ReadReports`
  name an object?

### 3. Make the guardrail hold

**Goal:** in `3_backup_guardrail.json`, make every delete under `daily/` an explicit deny, while the job keeps working.

The backup job uploads to `tmp/`, copies the file into `daily/`, and deletes
the temp file. The team's rule: **nobody deletes anything under `daily/`**,
which `NeverDeleteDailyBackups` was written to enforce. Last night this ran
with the job's role, and succeeded:

```text
$ aws s3 rm s3://acme-db-backups/daily/2026-10-01.sql.gz
delete: s3://acme-db-backups/daily/2026-10-01.sql.gz
```

Fix the guardrail; keep `WriteBackups` and `ListBackups` as they are.

Allowed (`tmp` = `arn:aws:s3:::acme-db-backups/tmp/2026-10-13.sql.gz`):

| Action | Resource |
|---|---|
| `s3:PutObject` | `tmp` |
| `s3:PutObject` | `arn:aws:s3:::acme-db-backups/daily/2026-10-13.sql.gz` |
| `s3:GetObject` | `arn:aws:s3:::acme-db-backups/daily/2026-10-01.sql.gz` |
| `s3:DeleteObject` | `tmp` |
| `s3:ListBucket` | `arn:aws:s3:::acme-db-backups` |

**Explicit deny** (`ExplicitDeny`, not just denied):

| Action | Resource |
|---|---|
| `s3:DeleteObject` | `arn:aws:s3:::acme-db-backups/daily/2026-10-01.sql.gz` |
| `s3:DeleteObject` | `arn:aws:s3:::acme-db-backups/daily/orders/2026-10-01.sql.gz` |

The test then adds a second policy to the role, `"Action": "s3:*"` on
`"Resource": "*"` (like `AmazonS3FullAccess`), and checks again: deleting
`daily/2026-10-01.sql.gz` is still `ExplicitDeny`, and deleting `tmp` is
still allowed.

**Think it through:**
- `s3:DeleteObject` works on objects. What kind of ARN does the guardrail
  name?
- Why would taking `s3:DeleteObject` out of `WriteBackups` not pass, even
  though deletes would then fail?

## Learn more

- [Refine permissions in AWS using last accessed information](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_last-accessed.html)
- [Policy evaluation logic](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_evaluation-logic.html)
- [IAM policy testing with the IAM policy simulator](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_testing-policies.html)
