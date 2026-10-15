# Design docs: propose a log retention fix

Before a team builds something big, someone writes a short *design doc*:
the problem, the limits, the options, and the choice. It costs an hour and
can save weeks, because reviewers argue with a page instead of with a
half-built system. In platform teams these are often called *RFCs*
(requests for comments). This exercise has no Go code: you write one.

**New here:** design docs, constraints, options with trade-offs, a
decision with a rollback plan.

## Learn

### Why write it down?

Talking through a change in a meeting feels faster, but:

- people who weren't there never hear the reasons;
- in six months nobody remembers why the team chose this approach;
- writing exposes gaps. "We'll ship logs somewhere" becomes "which
  bucket, who pays, what if the shipper dies?".

A design doc is short (one or two pages), written **before** the work,
and reviewed like code.

### The four sections

| Section | Answers | A weak version | A strong version |
|---|---|---|---|
| Problem | What's wrong, and how do we know? | "Disks fill up." | "3 incidents last month; logs use up to 84 GB of a 100 GB disk." |
| Constraints | What must any solution respect? | "Cheap." | "Node costs may grow at most 10%." |
| Options | What could we do? | one option, all upside | two or three options, each with a real cost |
| Decision | What do we do, why, and how do we undo it? | "Let's do B." | "Option B, because it's the only one that meets the audit rule; to roll back, revert the config." |

**Problem first.** If reviewers don't agree on the problem, they will
never agree on the solution. Numbers make the problem real.

**Always more than one option.** A doc with one option is an
announcement, not a design. "Do nothing" and "the cheap quick fix" are fine
options, and they make the trade-off visible.

**Every option has a cost.** If you can't name a downside, you haven't
looked hard enough. Reviewers trust a doc more when it is honest about
the chosen option's costs.

**Plan the way back.** Ops changes go wrong. Saying how you would roll
back is part of the decision, not an afterthought.

### Markdown in two lines

`design.md` is Markdown, the same format as this README. `## ` starts a
section, `### ` a smaller heading inside it, and a line starting `- ` is a
bullet point. Text between `<!--` and `-->` is a comment: the reader (and
the tests) never see it.

### How the tests check

The tests in `exercise_test.go` read `design.md` and check that each
section is there and has real content: enough words, the bullets and
headings it needs, no leftover `TODO` or `<placeholder>`. They can't judge
whether your design is good; that's what review is for. Compare yours
with the reference design when you're done.

## Your task

Open `design.md` and fill it in, one section per task. Replace each
`<!-- TODO -->` comment with your writing; keep the `## ` headings exactly
as they are.

The facts, from last month's incident reviews:

- The cluster has **40 nodes**, each with a **100 GB** disk.
- Each node writes about **6 GB of application logs a day**. logrotate
  keeps **14 days** on the node.
- When a disk passes **85% full**, the kubelet evicts pods to free space.
- Disk pressure caused **3 incidents last month**. The longest took **2
  hours** to recover; checkout was down for 25 minutes of it.
- Audit rules say logs must be kept for **90 days** and be searchable.
- The budget allows node costs to grow by about **10%**, not more.
- **Fluent Bit**, a log shipper, already runs on every node but only ships
  error logs. There is an **S3-compatible bucket** the team can use, and
  buckets can delete old objects automatically (*lifecycle rules*).

### 1. Write the problem

**Goal:** fill in `## Problem` with at least 40 words, including at least
one number.

Say what goes wrong, who it hurts, and the evidence. Don't propose a
solution here.

**Think it through:**
- Which facts above are symptoms (what people see), and which are causes?
- Someone from another team reads this. What do they need to know to care?

### 2. List the constraints

**Goal:** fill in `## Constraints` with at least three bullet points
(`- ` at the start), each at least four words.

A constraint is something **every** option must respect: a rule, a
budget, a deadline, something that must not change.

**Think it through:**
- Which facts above limit what you're allowed to do?
- Is "should be cheap" checkable? What would make it checkable?

### 3. Compare the options

**Goal:** fill in `## Options` with at least two options. Each has a
`### Option <letter>: <name>` heading, a description of at least 10
words, and at least one `- Pro:` line and one `- Con:` line (each at least
3 words).

```markdown
### Option A: Restart the pods nightly

Restart every pod at 03:00 so memory leaks never build up for long.

- Pro: Quick to set up with a CronJob.
- Con: Hides the leak instead of fixing it.
```

(That example is about a different problem. Write your own options.)

**Think it through:**
- What's the quickest, cheapest fix? What's the most complete one? Both are worth listing.
- Check each option against your constraints. Which one breaks a constraint? That's a Con.

### 4. Make the decision

**Goal:** fill in `## Decision` with at least 40 words that name the chosen
option (`Option B`, for example), say why with the word `because`, and
say how to undo the change (with the words roll back, revert or undo).

**Think it through:**
- Which option meets every constraint? If none does, which constraint would you ask to relax?
- If the change makes things worse at 2 a.m., what exactly does on-call do?

## Learn more

- [Design Docs at Google (Industrial Empathy)](https://www.industrialempathy.com/posts/design-docs-at-google/)
- [Kubernetes Enhancement Proposals (KEPs)](https://github.com/kubernetes/enhancements/blob/master/keps/README.md)
- [Kubernetes: node-pressure eviction](https://kubernetes.io/docs/concepts/scheduling-eviction/node-pressure-eviction/)
