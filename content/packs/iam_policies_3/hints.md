## Task 1: Narrow the CI policy

The evidence table is the whole policy: two actions, one folder of one
bucket. Replace both wildcards; one statement is enough.

### Closer

`"Action"` becomes a list of the two actions. `"Resource"` becomes an object
ARN that stops at the folder: `arn:aws:s3:::` + bucket + `/builds/*`. Then
delete the `"Id"` line.

### Answer

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "CIArtifacts",
      "Effect": "Allow",
      "Action": ["s3:PutObject", "s3:GetObject"],
      "Resource": "arn:aws:s3:::acme-ci-artifacts/builds/*"
    }
  ]
}
```

## Task 2: Fix the report reader

Put the error's resource, `arn:aws:s3:::acme-reports`, next to the `Resource`
of `ListReports`. Then do the same for a report's object ARN and
`ReadReports`.

### Closer

The two resources are swapped. Listing works on the bucket ARN (no slash);
reading works on objects (`/*` after the bucket name). Keep the actions where
they are and swap the two `Resource` values.

### Answer

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ListReports",
      "Effect": "Allow",
      "Action": "s3:ListBucket",
      "Resource": "arn:aws:s3:::acme-reports"
    },
    {
      "Sid": "ReadReports",
      "Effect": "Allow",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::acme-reports/*"
    }
  ]
}
```

## Task 3: Make the guardrail hold

The deleted file's ARN is `arn:aws:s3:::acme-db-backups/daily/2026-10-01.sql.gz`.
Does `NeverDeleteDailyBackups` match it? A delete targets an object, but the
guardrail names something else.

### Closer

Change only the guardrail's `Resource` to an object ARN that covers
everything under `daily/` but nothing under `tmp/`. `arn:aws:s3:::acme-db-backups/*`
would be too much: the job could no longer clean up.

### Answer

```json
    {
      "Sid": "NeverDeleteDailyBackups",
      "Effect": "Deny",
      "Action": "s3:DeleteObject",
      "Resource": "arn:aws:s3:::acme-db-backups/daily/*"
    }
```
