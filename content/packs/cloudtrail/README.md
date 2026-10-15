# CloudTrail: trace a leaked access key

A CI user's access key ended up in a public Git repository, and someone used
it. Before you can clean up, you need to know what they did. AWS writes down
every API call in *CloudTrail*; here you load that log and reconstruct the
attack.

**New here:** CloudTrail events, `time.Time` in JSON, `time.Time.Compare`.

## Learn

### What CloudTrail records

Every call to an AWS API (from the console, the CLI, an SDK, Terraform)
becomes a CloudTrail *event*: who, what, when, from where, and whether it
worked.

| | |
|---|---|
| *Management events* | calls that read or change resources: `RunInstances`, `CreateUser`, `ListBuckets`. Recorded by default. |
| *Event history* | the last **90 days** of management events, free, in every account |
| *Trail* | a copy of events delivered to an S3 bucket, kept as long as you like. Set one up **before** you need it. |
| *Data events* | object-level calls like `s3:GetObject`. Off by default (there are a lot of them). |

### One event

A trail file is `{"Records": [ ... ]}`. One record, shortened:

```json
{
  "eventTime": "2026-10-02T13:50:08Z",
  "eventSource": "iam.amazonaws.com",
  "eventName": "CreateUser",
  "awsRegion": "us-east-1",
  "sourceIPAddress": "198.51.100.77",
  "userAgent": "Boto3/1.34.0 Botocore/1.34.0 ...",
  "userIdentity": {
    "type": "IAMUser",
    "arn": "arn:aws:iam::111122223333:user/ci-deployer",
    "accountId": "111122223333",
    "accessKeyId": "AKIAIOSFODNN7EXAMPLE",
    "userName": "ci-deployer"
  },
  "errorCode": "AccessDenied",
  "errorMessage": "User: arn:aws:iam::111122223333:user/ci-deployer is not authorized to perform: iam:CreateUser ...",
  "readOnly": false
}
```

| Field | Tells you |
|---|---|
| `eventTime` | when, in UTC |
| `eventSource` + `eventName` | which API: the service's endpoint and the call (`iam.amazonaws.com`, `CreateUser`) |
| `awsRegion` | where. Global services such as IAM log as `us-east-1` |
| `sourceIPAddress` | the caller's IP address |
| `userAgent` | the tool: `aws-cli/2…`, `Boto3/…` (Python), `AWS Internal` (the console) |
| `userIdentity` | who: `type` (`IAMUser`, `AssumedRole`, …), `arn`, and the `accessKeyId` used |
| `errorCode` | empty when the call worked; `AccessDenied`, `AccessDeniedException`, `Client.UnauthorizedOperation`… when it didn't |
| `readOnly` | `true` for calls that only look (`Describe…`, `List…`, `Get…`) |

Events arrive in batches, so a file is **not** sorted by time.

### Responding to a leaked key

1. **Deactivate** the key (`UpdateAccessKey` to `Inactive`). Don't delete
   it: you can reactivate it if you picked the wrong one.
2. **Trace** it: every event with that `accessKeyId`, oldest first. Which
   IPs? Which regions? (Attackers like regions you never use.)
3. **Look for persistence:** `CreateUser`, `CreateAccessKey`,
   `AttachUserPolicy`. A key the attacker created keeps working after you
   deactivate the leaked one.
4. **Clean up** what they made, then fix how the key leaked. (The fix that
   lasts: no long-term keys in CI at all, as in
   *Assume role: trust policies and temporary credentials*.)

### `time.Time` in JSON

You decoded JSON into structs in *Decoding JSON: load a service config*. A
`time.Time` field needs nothing extra: it decodes an RFC 3339 string
(`"2026-10-02T13:50:08Z"`) by itself, and a malformed time is a decode error.

```go
var deploy struct {
	Finished time.Time `json:"finished"`
}
err := json.Unmarshal([]byte(`{"finished": "2026-10-02T09:30:00Z"}`), &deploy)
// deploy.Finished.Hour() => 9
```

> **Why is it written like this?** `time.Time` has its own `UnmarshalJSON`
> method, the same hook the `StringList` type used in
> *IAM policies: how AWS decides allow or deny*. `encoding/json` calls it for
> you. RFC 3339 is the format it expects; that's the standard library's
> choice, and AWS uses the same format.

To decode a file whose list sits under one key, decode into a struct with just
that field. An *anonymous struct*, as in *Table-driven tests*, saves naming a
type you need once:

```go
var page struct {
	Items []string `json:"items"`
}
err := json.Unmarshal([]byte(`{"items": ["a", "b"], "next": "x"}`), &page)
// page.Items => ["a" "b"]; "next" is skipped
```

### Sorting by time

`a.Compare(b)` on two `time.Time` values returns `-1` if `a` is earlier, `0`
if equal, `+1` if later: exactly what `slices.SortFunc` wants (as in
*Sorting*).

```go
slices.SortFunc(backups, func(a, b Backup) int {
	return a.Taken.Compare(b.Taken) // oldest first
})
```

> **Why is it written like this?** `Compare` is a method on `time.Time`, so
> it's `a.Taken.Compare(...)`, not `cmp.Compare` (which only works on numbers
> and strings). Swapping `a` and `b` would sort newest first.

### Try it (optional)

Needs an AWS account and credentials; never needed for the tests.

```bash
aws cloudtrail lookup-events --lookup-attributes AttributeKey=AccessKeyId,AttributeValue=AKIAIOSFODNN7EXAMPLE --max-results 20
aws cloudtrail lookup-events --lookup-attributes AttributeKey=EventName,AttributeValue=CreateAccessKey --start-time 2026-10-01T00:00:00Z
aws cloudtrail describe-trails
aws iam list-access-keys --user-name ci-deployer
```

- The first lists recent events made with one key (from event history).
- The second finds every new access key since a date.
- `describe-trails` shows whether a trail keeps events beyond 90 days.
- `list-access-keys` shows a user's keys and whether each is `Active`.

## Your task

The leaked key is `AKIAIOSFODNN7EXAMPLE`, belonging to `ci-deployer`, whose
CI runner calls from `192.0.2.10`. `testdata/events.json` holds 27 events
from that account: CI runs, two colleagues, the attacker and the clean-up.
Open it and skim a few records. The tests for tasks 2 and 3 read it with your
`Load`, so do task 1 first.

### 1. Load the trail

**Goal:** `Load(path)` returns the events in the file's `"Records"` list, in file order.

- The `Event` and `UserIdentity` types and their JSON tags are ready.
- If the file can't be read, or isn't valid JSON, return the error (any
  message, `nil` events). `testdata/broken.json` is a cut-off file.
- Add `"encoding/json"` and `"os"` to the imports (and `"fmt"` if you wrap
  errors).

```go
events, err := Load("testdata/events.json")
// len(events) => 27, err => nil
// events[0].EventName => "ListBuckets", events[0].SourceIPAddress => "198.51.100.77"
// events[6].ErrorCode => "AccessDenied", events[6].UserIdentity.UserName => "ci-deployer"
Load("testdata/missing.json") // => nil, error
Load("testdata/broken.json")  // => nil, error
```

**Think it through:**
- Which function reads a whole file into `[]byte`?
- What struct do you decode into, so that `"Records"` lands in a `[]Event`?

### 2. Follow one access key

**Goal:** `ByAccessKey(events, keyID)` returns the events whose `UserIdentity.AccessKeyID` is `keyID`, sorted by `EventTime`, oldest first.

- Don't reorder the caller's `events`: build a new slice and sort that.
- No matching events gives an empty result (`nil` is fine).

```go
ByAccessKey(events, "AKIAIOSFODNN7EXAMPLE")
// => 17 events:
//    2026-10-01T09:12:03Z GetCallerIdentity   (CI, 192.0.2.10)
//    2026-10-01T09:12:05Z GetAuthorizationToken
//    ...
//    2026-10-02T13:47:12Z GetCallerIdentity   (first call from 198.51.100.77)
//    ...
//    2026-10-02T14:05:47Z DescribeInstances
ByAccessKey(events, "AKIAUNKNOWNEXAMPLE00") // => []
```

**Think it through:**
- Filtering into a new slice with `append` already leaves `events` alone.
  Which order do you filter and sort in, then?
- What does the compare function for `slices.SortFunc` return here? (Add
  `"slices"` to the imports.)

### 3. Summarize the incident

**Goal:** `Summarize(events)` fills in every `Summary` field for the incident report.

| Field | Value |
|---|---|
| `First` / `Last` | the earliest and latest `EventTime`. The input may be in any order. |
| `SourceIPs` | every `SourceIPAddress`, sorted, each once |
| `Regions` | every `AWSRegion`, sorted, each once |
| `Actions` | `"service:EventName"` for every event, sorted, each once. `service` is `EventSource` without the `.amazonaws.com` ending: `"ec2.amazonaws.com"` + `"RunInstances"` gives `"ec2:RunInstances"`. |
| `Denied` | how many events have a non-empty `ErrorCode` (any error counts) |

No events gives the zero `Summary{}`: zero times, empty lists, `Denied` 0.

```go
Summarize([]Event{
	{EventTime: /* 09:00 */, EventSource: "s3.amazonaws.com", EventName: "GetObject", AWSRegion: "eu-west-1", SourceIPAddress: "192.0.2.10"},
	{EventTime: /* 07:30 */, EventSource: "ec2.amazonaws.com", EventName: "RunInstances", AWSRegion: "eu-west-1", SourceIPAddress: "198.51.100.77", ErrorCode: "Client.UnauthorizedOperation"},
	{EventTime: /* 08:15 */, EventSource: "s3.amazonaws.com", EventName: "GetObject", AWSRegion: "us-east-1", SourceIPAddress: "192.0.2.10", ErrorCode: "AccessDenied"},
})
// => First 07:30, Last 09:00
//    SourceIPs ["192.0.2.10" "198.51.100.77"]
//    Regions   ["eu-west-1" "us-east-1"]
//    Actions   ["ec2:RunInstances" "s3:GetObject"]
//    Denied    2
```

**Think it through:**
- "Sorted, each once": which data structure drops duplicates for you, and
  how did you get its keys out sorted in earlier exercises? (Add `"maps"`
  and `"strings"` to the imports.)
- What should `First` and `Last` start as, so the first event always
  replaces them? Starting `First` at the zero time fails: no event is earlier.

When all tests pass, run `Summarize` on the leaked key's events in your head
or in a scratch test: which regions did CI never use, and what did
`CreateAccessKey` give the attacker? (Look at that record's
`responseElements` in `events.json`, then trace that key too.)

## Learn more

- [CloudTrail record contents](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-event-reference-record-contents.html)
- [Working with CloudTrail event history](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/view-cloudtrail-events.html)
- [Manage access keys for IAM users](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html)
- [time.Time.UnmarshalJSON](https://pkg.go.dev/time#Time.UnmarshalJSON)
