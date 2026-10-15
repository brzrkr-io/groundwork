# Cloud and cluster access: IAM, RBAC and short-lived credentials

**In one sentence:** AWS IAM and Kubernetes RBAC decide which identity may do which action on which resource, and the safest identities have only the permissions their job needs and credentials that expire on their own.

**Why it matters:** many cloud incidents start with a leaked key or a too-broad permission, not clever hacking. If a deploy role can do `*` on `*`, one leaked CI token costs you the account; if the role is narrow, the leak is a non-event.

**You'll learn:**

- The identities in AWS (root user, IAM users, roles) and in Kubernetes (users, ServiceAccounts).
- How to read and write an IAM policy, and exactly how AWS decides allow or deny.
- Least privilege in the cloud: shrinking the blast radius of every credential.
- How roles hand out temporary credentials, and how CI and pods get them without stored keys.
- How Kubernetes RBAC grants permissions with Roles and bindings.
- How to check a permission before you rely on it.

## The same two questions, at cloud scale

The *Security basics for automation* guide split access into two questions: *authentication* (who are you?) and *authorization* (what may you do?). Clouds and clusters ask the same questions with their own vocabulary:

| | AWS | Kubernetes |
|---|---|---|
| Identity | Root user, IAM user or IAM role | User, group or ServiceAccount |
| Proves it with | Requests signed with an access key or role credentials | Client certificate or token |
| Permissions written as | IAM policies (JSON) | Roles and ClusterRoles (YAML) |
| Attached by | Policy on the user or role | RoleBinding or ClusterRoleBinding |
| Default answer | Deny | Deny |

In both, **no means no**: an identity can only do what something explicitly grants.

## AWS identities

- **The root user** is the email address that created the account. It can do anything, including closing the account, and IAM policies can't restrict it. Give it a strong password and hardware MFA, and then don't use it.
- **IAM users** are long-lived identities with a password and/or *access keys* (`AKIA...` plus a secret). Keys never expire on their own, so they're what leaks. Use them rarely.
- **IAM roles** are identities with no password and no long-term keys. Someone or something *assumes* a role and gets temporary credentials. Programs, CI jobs, EC2 instances and pods should use roles.

Humans usually sign in through *IAM Identity Center* (single sign-on), which also hands out role sessions, so nobody keeps a personal access key on their laptop.

## Reading an IAM policy

A *policy* is a JSON document of *statements*. Each statement says: allow or deny these actions on these resources, optionally under these conditions.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "WriteNightlyBackups",
      "Effect": "Allow",
      "Action": "s3:PutObject",
      "Resource": "arn:aws:s3:::acme-db-backups/nightly/*"
    }
  ]
}
```

| Element | Means | Example |
|---|---|---|
| `Version` | The policy language version; always `"2012-10-17"` | |
| `Sid` | An optional label for the statement | `WriteNightlyBackups` |
| `Effect` | `Allow` or `Deny` | `Allow` |
| `Action` | One or more API actions, `service:Name` | `s3:PutObject`, `ec2:Describe*` |
| `Resource` | One or more ARNs the actions apply to | `arn:aws:s3:::acme-db-backups/nightly/*` |
| `Condition` | Extra tests on the request | Only from `203.0.113.0/24` |

`*` matches any run of characters, and `?` exactly one: `ec2:Describe*` covers every EC2 read action, and `.../nightly/*` covers every object under `nightly/`. Action names ignore case; ARNs don't.

**Where policies live.** An *identity-based policy* is attached to a user, group or role and says what *it* may do. A *resource-based policy* is attached to a resource, such as an S3 bucket policy or a role's trust policy, and has an extra `Principal` element saying *who* may use it. You'll meet both in *S3 bucket policies: is this bucket public?* and *Assume role: trust policies and temporary credentials*.

## How AWS decides

For every request, AWS gathers every policy that applies and walks this order:

```text
request: alice wants s3:DeleteObject on acme-db-backups/nightly/x
   │
   ├─ does any statement DENY it?          yes → DENIED (explicit deny)
   │                                        no ↓
   ├─ does any statement ALLOW it?         yes → ALLOWED
   │                                        no ↓
   └─ nothing matched                          → DENIED (implicit deny)
```

Three rules fall out of this:

1. **Default deny.** If nothing allows it, it's denied. You never need a Deny to block something nobody granted.
2. **Deny beats allow.** One matching Deny wins against any number of Allows, in any policy. That's why Deny statements make good *guardrails*: "nobody may delete backups", whatever else they're granted.
3. **Order doesn't matter.** Statements aren't read top to bottom; all of them are checked.

Big organisations add more gates on top: *service control policies* (set for whole accounts by AWS Organizations) and *permissions boundaries* (a ceiling on one role). They can only take permissions away, never add them. You'll write this evaluator in Go in *IAM policies: how AWS decides allow or deny*.

## Least privilege in the cloud

*Least privilege* means each identity gets only the permissions its job needs. In the cloud the damage a credential can do, its *blast radius*, can be enormous, so this matters more than anywhere.

Here's a statement that shows up in real accounts, attached to a nightly backup script "just to get it working":

```json
{
  "Effect": "Allow",
  "Action": "*",
  "Resource": "*"
}
```

That says: this script may do anything, to anything. If its key leaks, the attacker can read every bucket, delete every database and create new admin users for later. The script only needed the `s3:PutObject` statement from *Reading an IAM policy*. With that, a leaked key can only write files into one backup folder.

How to get to least privilege in practice:

- **Start from what the job does**, not from what might be handy: list the API calls it makes.
- **Scope resources too.** `s3:GetObject` on `arn:aws:s3:::acme-reports/*` is much smaller than on `*`.
- **Watch the bucket-versus-object trap.** `s3:ListBucket` applies to the bucket ARN (`arn:aws:s3:::acme-reports`); `s3:GetObject` and `s3:PutObject` apply to object ARNs (`arn:aws:s3:::acme-reports/*`). Mixing them up is the most common "access denied" bug.
- **Some actions can't be scoped.** Most `Describe*` and `List*` calls only accept `"Resource": "*"`. That's fine for reads.
- **Remove what isn't used.** IAM shows when a role last used each service, and *IAM Access Analyzer* can suggest a policy from what CloudTrail saw the role do.
- **Add guardrails.** An explicit Deny for the few actions that must never happen (deleting backups, turning off CloudTrail).

You'll write narrow policies in *IAM policies: write least-privilege policies* and fix broken ones in *IAM policies: fix the over-broad policy*.

## Roles and temporary credentials

A role has two policies:

| Policy | Answers | Example |
|---|---|---|
| *Trust policy* | Who may assume this role? | "the CI provider, for repo `acme/deploy` on `main`" |
| *Permissions policy* | What may the role do once assumed? | "`ecs:UpdateService` on the `web` service" |

Assuming a role is an API call such as `sts:AssumeRole` to the *Security Token Service* (STS). If the trust policy allows the caller, STS returns *temporary credentials*:

| Part | Looks like | Note |
|---|---|---|
| Access key ID | `ASIA...` | `ASIA` means temporary; `AKIA` means a long-term user key |
| Secret access key | `wJal...` | Never log it |
| Session token | A long string | Must be sent with every request |
| Expiration | A timestamp | Usually 1 hour; 15 minutes to 12 hours |

A leaked session stops working when it expires, and there's no long-lived secret to rotate. That's why roles beat users for everything that isn't a human.

### Where the credentials come from

You rarely call STS yourself. The platform does it, and the SDK's *default credential chain* finds the result:

- **EC2 instances** get a role through an *instance profile*; the instance's metadata service hands out fresh credentials.
- **Pods on EKS** get a role per ServiceAccount, through *EKS Pod Identity* or the older *IRSA* (IAM roles for service accounts).
- **CI jobs** swap an identity token for a role session (next section).
- **Your laptop** gets a session from `aws sso login`.

The *Cloud automation with the AWS SDK* module loads these from Go, starting with *AWS SDK configuration: regions, profiles and credentials*.

## CI without stored keys: OIDC

The *Continuous integration and delivery* guide showed why a CI job should use *OIDC* (OpenID Connect) instead of a stored cloud key. Here is the AWS side:

```text
GitHub Actions job                      AWS
   │  1. "give me an ID token"   ─▶ GitHub's OIDC provider
   │  ◀─ signed token: repo=acme/deploy, ref=refs/heads/main
   │  2. AssumeRoleWithWebIdentity(token) ─────────▶ STS
   │                          checks the role's trust policy
   │  ◀─ temporary credentials, 1 hour ────────────── STS
   │  3. deploy with them
```

The role's trust policy decides which workflows may assume it:

```json
{
  "Effect": "Allow",
  "Principal": {
    "Federated": "arn:aws:iam::111122223333:oidc-provider/token.actions.githubusercontent.com"
  },
  "Action": "sts:AssumeRoleWithWebIdentity",
  "Condition": {
    "StringEquals": {
      "token.actions.githubusercontent.com:aud": "sts.amazonaws.com",
      "token.actions.githubusercontent.com:sub": "repo:acme/deploy:ref:refs/heads/main"
    }
  }
}
```

The `sub` condition is the important line. Without it, **any** GitHub repository in the world could assume your role. With `StringLike` and `repo:acme/deploy:*`, every branch and pull request of that repo can, which is usually too wide for a production role. You'll check conditions like these in Go in *Assume role: trust policies and temporary credentials*.

## Kubernetes RBAC

Inside a cluster, *RBAC* (role-based access control) decides what each identity may do through the API server. Workloads use a *ServiceAccount*; every Pod runs as one (the namespace's `default` unless you pick another).

A *Role* lists rules: which *verbs* on which *resources* in which API groups, inside one namespace. A *RoleBinding* grants a Role to users, groups or ServiceAccounts.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: pod-reader
  namespace: shop
rules:
  - apiGroups: [""]            # "" is the core group: pods, services, secrets
    resources: ["pods", "pods/log"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: log-shipper-reads-pods
  namespace: shop
subjects:
  - kind: ServiceAccount
    name: log-shipper
    namespace: shop
roleRef:
  kind: Role
  name: pod-reader
  apiGroup: rbac.authorization.k8s.io
```

Step by step: the Role allows reading Pods and their logs in `shop`; the RoleBinding gives that to the `log-shipper` ServiceAccount; a Pod running as `log-shipper` can now list Pods in `shop`, and nothing else.

| Object | Scope |
|---|---|
| Role + RoleBinding | One namespace |
| ClusterRole + ClusterRoleBinding | The whole cluster, including cluster-wide resources like nodes |
| ClusterRole + RoleBinding | A reusable rule set, granted in one namespace |

Differences from IAM worth knowing:

- **RBAC only adds.** There's no Deny. To take something away, remove the binding.
- **Some verbs are stronger than they look.** `get secrets` reads every password in the namespace. `create pods` is nearly the same, since a Pod can mount any Secret in its namespace. `escalate`, `bind` and `impersonate` let an identity give itself more.
- **`cluster-admin` is root.** Bind it to as few people as possible, and never to a workload.

A controller like the one you built in *Kubernetes with client-go* needs exactly this in a real cluster: a ServiceAccount with a Role allowing `get`, `list`, `watch` and `update` on the objects it reconciles.

## Checking before you act

Guessing about permissions wastes time and causes outages. Ask the system:

| Question | Command |
|---|---|
| Who am I right now? | `aws sts get-caller-identity` |
| Would this role be allowed? | `aws iam simulate-principal-policy --policy-source-arn <role ARN> --action-names s3:PutObject --resource-arns <object ARN>` |
| Would this EC2 call work? | Any EC2 call with `--dry-run` (or `DryRun` in the SDK) |
| Can I do this in the cluster? | `kubectl auth can-i delete pods -n shop` |
| Can that ServiceAccount? | `kubectl auth can-i list secrets -n shop --as system:serviceaccount:shop:log-shipper` |

All of these are read-only. You'll handle the dry-run answer in Go in *AWS API errors: EC2 dry-run permission checks*.

## Common mistakes

- **`"Action": "*"` to get something working.** Grant the exact actions on the exact resources from the start.
- **Using the root user day to day.** Lock it away with MFA; use roles.
- **Long-lived access keys for programs.** Use roles: instance profiles, Pod Identity, OIDC for CI.
- **An OIDC trust policy without a `sub` condition.** Any repository could assume the role.
- **Granting `s3:GetObject` on the bucket ARN.** Objects need `bucket/*`; `ListBucket` needs the bucket itself.
- **Binding `cluster-admin` to a workload's ServiceAccount.** Give it a Role with the verbs it really uses.
- **Expecting a Deny in Kubernetes RBAC.** There isn't one; remove the binding.

## Check yourself

1. In AWS, one attached policy allows `s3:*` on a bucket and another denies `s3:DeleteObject` on the same objects. A user tries to delete an object. What happens, and why?
2. A policy allows `s3:GetObject` on `arn:aws:s3:::acme-reports`, but every read of `acme-reports/2026/q3.csv` is denied. What's wrong?
3. Why are IAM roles safer than IAM users with access keys for programs?
4. A role's trust policy lets GitHub's OIDC provider assume it with only an `aud` condition and no `sub` condition. What's the risk?
5. Which Kubernetes objects do you create so a Pod running as ServiceAccount `log-shipper` can list Pods in its own namespace, and nothing else?
6. Why is granting `create pods` in a namespace almost as powerful as granting `get secrets` there?
7. An access key ID starts with `ASIA`. What does that tell you, and what else must accompany it on each request?

## Answers

1. The delete is denied. AWS checks for any matching explicit Deny first, and an explicit Deny beats any Allow, whichever policy it's in.
2. `s3:GetObject` acts on objects, whose ARNs are `arn:aws:s3:::acme-reports/<key>`. The resource must be `arn:aws:s3:::acme-reports/*` (or a narrower prefix); the bucket ARN alone matches no object.
3. A role has no long-term secret. Whoever assumes it gets credentials that expire within hours, so a leak is short-lived and there's nothing to rotate; user access keys stay valid until someone deletes them.
4. Any GitHub Actions workflow, in any repository, can get a token with that audience and assume the role. The `sub` condition is what restricts it to your repository and branch.
5. A Role in that namespace allowing `list` (and usually `get` and `watch`) on `pods`, and a RoleBinding granting that Role to the `log-shipper` ServiceAccount.
6. A Pod can mount any Secret in its namespace, so someone who can create Pods can create one that mounts a Secret and prints it.
7. It's a temporary credential from STS (a role session). It only works together with its session token, and it stops working at its expiration time.

## Key terms

| Term | Meaning |
|---|---|
| IAM | AWS Identity and Access Management: identities and their permissions |
| Root user | The account's original identity, which can do anything; locked away after setup |
| IAM user | A long-lived AWS identity that can have a password and access keys |
| IAM role | An AWS identity with no long-term credentials, assumed for temporary sessions |
| Policy | A JSON document of statements that allow or deny actions on resources |
| Statement | One rule in a policy: Effect, Action, Resource and optional Condition |
| Explicit deny | A matching Deny statement; it beats every Allow |
| Implicit deny | The default answer when no statement allows a request |
| Identity-based policy | A policy attached to a user, group or role |
| Resource-based policy | A policy attached to a resource, with a Principal saying who it applies to |
| Trust policy | A role's resource-based policy saying who may assume it |
| STS | The Security Token Service, which issues temporary credentials |
| Session token | The extra value that must accompany temporary credentials |
| Default credential chain | The order in which an AWS SDK looks for credentials |
| OIDC federation | Exchanging a signed identity token from another system for a role session |
| Guardrail | A Deny or policy that blocks dangerous actions whatever else is allowed |
| RBAC | Kubernetes role-based access control |
| ServiceAccount | The Kubernetes identity a Pod runs as |
| Role / ClusterRole | Kubernetes rules: verbs on resources, in one namespace or cluster-wide |
| RoleBinding / ClusterRoleBinding | Grants a Role or ClusterRole to users, groups or ServiceAccounts |

## Watch

- [IAM Policy Evaluation Series: AWS IAM policy language explained | Amazon Web Services](https://www.youtube.com/watch?v=qsF6Kauh2J4) · Amazon Web Services · 39 min: the policy elements, the dos and don'ts, and how IAM matches statements against a request.
- [AWS re:Invent 2021 - A least privilege journey: AWS IAM policies and Access Analyzer](https://www.youtube.com/watch?v=pKPiPplJNak) · AWS Events · 56 min: setting, verifying and refining permissions over their lifetime with IAM Access Analyzer, with demos of conditions and cross-account access.
- [Kubernetes RBAC Explained Visually (Service Accounts, Roles & RoleBindings)](https://www.youtube.com/watch?v=bhVMSSWJL6s) · arconsis · 7 min: how a request from a ServiceAccount flows through Roles and RoleBindings.
- [GitHub Actions: Configure OIDC Authentication for AWS and Terraform](https://www.youtube.com/watch?v=USIVWqXVv_U) · More than Certified · 10 min: setting up the OIDC provider, role and trust policy for a real workflow.

## Go deeper

- [Policy evaluation logic](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_evaluation-logic.html) (AWS): the full decision flow, including service control policies and permissions boundaries.
- [Security best practices in IAM](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html) (AWS): least privilege and temporary credentials, from the source.
- [Using RBAC Authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) (kubernetes.io): Roles, bindings, default roles and real examples.
- [Configuring OpenID Connect in Amazon Web Services](https://docs.github.com/en/actions/how-tos/secure-your-work/security-harden-deployments/oidc-in-aws) (GitHub): the provider, the trust policy conditions and a workflow.
