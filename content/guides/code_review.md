# Code review and working on a team

**In one sentence:** code review is a teammate reading your change before it ships, and together with small pull requests, clear commits, written designs and a blameless culture, it is how a team builds software that no single person could keep safe alone.

**Why it matters:** in platform work one merged line can restart every pod in a cluster, open a firewall, or delete a bucket. Review is the last moment a human can say "wait, what happens at 3 a.m. when this fails?". It's also how a team shares what it knows, so the person on call isn't the only one who understands the code that just broke.

**You'll learn:**

- What code review is for, beyond catching bugs
- A checklist of what to look for, from correctness to operational impact
- How to give feedback that helps, and how to receive it without taking it personally
- Why small pull requests and tidy commits get better reviews
- What a design doc is, and when to write one before writing code
- How pair programming and blameless culture make a team safer

## Why review code at all

### More than catching bugs

Reviews do catch bugs, but studies of large teams find that's not their main value. A review also:

- **Spreads knowledge.** After review, at least two people understand the change. When it breaks at night, the on-call engineer may well be the reviewer.
- **Keeps the code consistent.** Reviewers notice when a change invents a second way to do something the codebase already does.
- **Teaches, both ways.** Juniors learn from seniors' comments; seniors learn from juniors' questions ("why is this retry here?" sometimes has no good answer).
- **Checks the operational story.** Will this change roll out safely? Can it be rolled back? Will we know if it's broken?

### What review is not

Review is not a gate to prove you're clever, or a place to argue about taste. Formatting, import order and simple mistakes belong to tools: `gofmt`, `go vet`, linters and CI. Humans should spend their attention on what tools can't see: whether the change is **right**, **clear** and **safe**.

## What to look for

Read the description first: what is this change **supposed** to do? Then read the code against that promise, biggest questions first. There's no point polishing names in a change that solves the wrong problem.

| Area | Questions to ask |
|---|---|
| Correctness | Does it do what the description says, in every case it promises? What about the boundaries: exactly 7 days, an empty list, a zero? |
| Tests | Is there a test for each new behaviour and each fixed bug? Would the tests fail if the code were wrong? |
| Design | Is this the simplest approach? Does it fit how the rest of the codebase does things? |
| Naming and readability | Could you explain each function from its name? Would the 3 a.m. on-call engineer follow it? |
| Error handling | Are errors checked, wrapped with context and returned, not swallowed? What happens on a timeout or a partial failure? |
| Security | Are secrets kept out of code and logs? Is user input validated? Does anything gain more permissions than it needs? |
| Operational impact | Is it observable (logs, metrics)? Can it roll out gradually and roll back? Does it change config, a schema or a dependency? What does it cost? |

**Bugs live where rules meet.** One rule's code path often skips another rule's check: an early `return` for weekly backups that never reaches the "pinned backups are never deleted" check. Follow each special case through the code, line by line.

### Tests are review evidence

The strongest review comment is often a failing test: "this case should keep the backup, and here's a test showing it doesn't". It replaces an argument with a fact, and once the bug is fixed the test stays and guards against it forever. You do exactly this in *Code review: catch the bugs in a PR*.

## Giving feedback that helps

### Talk about the code, not the person

Compare two comments on the same line:

| Unhelpful | Helpful |
|---|---|
| "This is wrong." | "A backup exactly 7 days old gets deleted here, but the description says it's kept. `>` instead of `>=`?" |
| "Why would you do it like this?" | "I'd have expected a map lookup here; was the switch on purpose?" |
| "Use better names." | "`filterHealthy` returns the unhealthy nodes. Maybe `unhealthyNodes`?" |

The helpful versions are **specific** (which line, which case), **explain why**, and often **ask** rather than tell. The author may know something you don't.

### Label how much each comment matters

An author with twenty comments can't tell which ones block the merge. Say so:

- **blocking:** must change before merge (a bug, a security hole, no rollback plan).
- **suggestion:** worth considering; author decides.
- **nit:** tiny polish; fine to ignore.
- **question:** you want to understand, not necessarily change anything.

Some teams use a fixed format for this, *conventional comments* (linked below).

### Say what's good, and approve when it's good enough

Point out the clever test or the clear name: it tells the author what to keep doing. And approve a change once it **makes the code better than before**, even if it isn't how you'd have written it. Google's review guide sets exactly that standard. Perfect is the enemy of shipped.

## Receiving feedback

It's hard not to feel review comments as personal. They aren't: they're about a diff, and the reviewer is spending their time to make your work better.

- **Assume good intent.** A short comment is usually a busy reviewer, not an angry one.
- **Reply to every comment.** "Done", "Good catch, fixed in the next commit", or a reason you disagree. Silence leaves the reviewer guessing.
- **Push back with reasons.** "I kept the switch because each case calls a different function" is a fine answer. Reviewers are sometimes wrong too.
- **Take big disagreements off the page.** If a thread goes back and forth three times, talk for five minutes, then write the outcome in the PR.
- **Say thanks.** Someone caught your bug before production did.

## Small pull requests

### Why size matters

A reviewer facing 1,000 changed lines skims; a reviewer facing 100 lines reads. Small pull requests get **faster** reviews and **better** ones. They're also safer to ship: if a small change breaks production, the cause is obvious and the revert is clean.

Ways to keep changes small:

- **Separate refactoring from behaviour changes.** First a PR that renames and extracts (tests stay green, nothing changes), then a PR with the fix, now easy to see. The *Code for humans: naming, simplicity and refactoring* guide explains why.
- **Split by layer or step.** Add the new config option in one PR, use it in the next, delete the old one in a third.
- **Ship dark.** Merge code behind a flag that's off, then turn it on in a separate, tiny change.

### A description reviewers can use

```text
Delete expired backups automatically

Why: the backup bucket grew 40% last quarter and nobody cleans it.
What: Expired() decides per backup; the nightly job deletes those.
How tested: table test for every rule, including the 7- and 28-day
boundaries; dry run against a copy of the production bucket.
Risk and rollback: deletion is permanent. The job runs with --dry-run
for a week first; to roll back, revert this PR.
```

What and why, how you know it works, and what happens if it doesn't. A reviewer can check the code against every line of that.

## Commit hygiene

Review usually happens on a pull request, but the commits inside it outlive the review. `git log`, `git blame` and `git bisect` will show them to someone years later.

- **One logical change per commit.** "Rename helpers" and "Fix pinned weekly backups" are two commits, even in one PR.
- **Messages that explain why.** A short subject in the imperative ("Keep pinned backups"), a blank line, then the reason. The diff already shows what changed.
- **Every commit builds and passes tests,** so `git bisect` can land on any of them.

The *How Git works* guide covers commits, messages and history in depth.

## Design docs: review before code

Some changes are too big to review as a diff: a new service, a storage migration, a change to how every team deploys. By the time the code exists, nobody wants to hear "this was the wrong approach". So the team reviews the **plan** first, as a short written *design doc*. Platform teams often call these *RFCs* (requests for comments); Kubernetes calls them *KEPs*.

A design doc is one or two pages:

| Section | Answers |
|---|---|
| Problem | What's wrong, who it hurts, and the evidence (with numbers) |
| Constraints | What every solution must respect: budget, rules, deadlines, things that mustn't change |
| Options | Two or three real choices, each with its costs as well as its benefits |
| Decision | Which option, why, and how to roll back if it goes wrong |

**Always more than one option**: a doc with one option is an announcement, not a design. **Plan the way back**: an ops change without a rollback plan isn't finished. Writing a doc is worth it when a change is hard to undo, affects other teams, or will take more than a week or two. You write one in *Design docs: propose a log retention fix*.

## Pair programming

In *pair programming*, two people work on one problem at one screen (or one shared session). The *driver* types; the *navigator* reviews each line as it's written, thinks ahead and spots problems. Swap roles often, every 15 to 30 minutes.

It feels like half the speed, but pairs often finish sooner: fewer wrong turns, fewer bugs, and the review happens continuously instead of days later. It's especially good for:

- **Onboarding.** A newcomer learns the codebase and the team's habits in days instead of weeks.
- **Hard problems.** Two people talking through a design catch more than one person thinking alone.
- **Incidents.** One person drives the terminal, the other reads logs, checks dashboards and keeps notes. Many on-call teams pair by default during an outage.

Pairing is tiring, so take breaks, and don't pair on everything. Small, well-understood changes are fine solo.

## Blameless culture

### Systems, not people

After an incident, a team writes a *postmortem*: what happened, why, and what will stop it happening again. In a *blameless* postmortem, the question is never "who messed up?" but **"how did our system make this mistake easy?"**.

Say an engineer ran a cleanup script against production instead of staging. Blaming them changes nothing: the next tired engineer will do the same. Asking why the system allowed it finds real fixes: the script should show which cluster it's about to touch, ask for confirmation in production, or not have production credentials by default.

### Why blame makes things worse

When people are punished for mistakes, they hide them. Incidents get reported late, details get left out, and the team stops learning. When people can say "I ran the wrong command" safely, the postmortem gets the facts it needs, and the fixes go into the system where they help everyone.

The same spirit applies to code review. "Who wrote this bug?" is the wrong question; "why didn't our tests or review catch it, and what would?" is the right one. `git blame` is a tool for finding context, not culprits.

## Common mistakes

- **Rubber-stamping.** Approving a big PR after a skim because the author is senior or the reviewer is busy. If you can't review it properly, say so, or ask for it to be split.
- **Nitpicking first.** Twenty comments about naming on a change that solves the wrong problem. Read for correctness and design before style.
- **Leaving formatting to humans.** Arguing about spacing that `gofmt` and CI should enforce automatically.
- **Unlabelled comments.** The author can't tell what blocks the merge, so they either change everything or argue about everything.
- **Giant pull requests.** Mixing a refactor, a feature and a dependency upgrade in one 1,500-line PR. Nobody can review it well, and nobody can revert just one part.
- **"Fix stuff" commits.** Messages that don't say why leave the next engineer guessing during an incident.
- **Skipping the rollback plan.** In a PR description or a design doc, "how do we undo this?" is part of the change.
- **Blaming people in postmortems.** It hides the facts and leaves the system just as easy to break.

## Check yourself

1. Besides catching bugs, name two things code review does for a team.
2. A reviewer leaves 25 comments about variable names on a PR that solves the wrong problem. What went wrong in how they reviewed it?
3. Rewrite the review comment "This is wrong" on a line that deletes backups exactly 7 days old, when the spec says they are kept.
4. What does prefixing a review comment with "nit:" tell the author?
5. Why does a 100-line pull request usually get a better review than a 1,000-line one, and what else makes it safer to ship?
6. A PR renames six functions and also fixes a bug. Why should those be two separate commits or PRs?
7. What four sections does a short design doc have, and why should it list more than one option?
8. An engineer runs a cleanup script against production instead of staging. What question does a blameless postmortem ask, and what kind of fix does it lead to?

## Answers

1. Any two of: it spreads knowledge, so more than one person understands each change; it keeps the codebase consistent; it teaches in both directions; it checks the operational story, such as rollout, rollback and observability.
2. They reviewed in the wrong order. Correctness and design come first: does the change do the right thing, the right way? Naming and polish only matter once that's settled, and 25 small comments bury the one that matters.
3. Something like: "A backup exactly 7 days old is deleted here, but the PR description says it should be kept. Should this be `>` instead of `>=`? A test for exactly 7 days would pin it down." It's specific, explains why, and offers a fix and evidence.
4. That the comment is minor polish and doesn't block the merge; the author may take it or leave it.
5. Reviewers read 100 lines carefully but skim 1,000, so problems in a big PR slip through. A small change is also safer to ship: if it breaks something, the cause is easy to find and the revert undoes only that one change.
6. Renames should change no behaviour; a bug fix does. Separated, each is easy to review on its own terms, and if something breaks you know which change caused it. Mixed, the one important line hides among the renames.
7. Problem, constraints, options and decision (including how to roll back). With only one option it's an announcement rather than a design; listing alternatives with their costs shows why the choice is right and lets reviewers challenge it.
8. "How did our system make this mistake easy?" rather than "who did it?". It leads to fixes in the system: the script shows which cluster it targets, asks for confirmation in production, or doesn't get production credentials by default.

## Key terms

| Term | Meaning |
|---|---|
| code review | a teammate reading a change before it is merged, to check it is right, clear and safe |
| pull request | a request to merge a branch, where the diff is reviewed, tested by CI and discussed (GitLab calls it a merge request) |
| nit | a review comment about minor polish that doesn't block the merge |
| blocking comment | a review comment that must be resolved before the change can be merged, such as a bug or a security hole |
| conventional comments | a format that labels each review comment with its kind and weight, such as "suggestion:" or "nit:" |
| commit hygiene | keeping commits small, one logical change each, with messages that explain why |
| design doc | a short document reviewed before building something big: problem, constraints, options and decision |
| RFC | "request for comments": the name many teams use for a design doc that is open for review |
| rollback plan | the written steps for undoing a change if it goes wrong |
| pair programming | two people working on one problem together, one typing (driver) and one reviewing and thinking ahead (navigator) |
| postmortem | a written review after an incident: what happened, why, and what will prevent it next time |
| blameless culture | treating mistakes as signs of weak systems rather than bad people, so facts come out and fixes go into the system |

## Watch

- [RailsConf 2015 - Implementing a Strong Code-Review Culture](https://www.youtube.com/watch?v=PJjmw9TRB7s) · Confreaks · 38 min: Derek Prior on why reviews are about learning and teaching more than catching bugs, and how to write comments people welcome.
- [Postmortem Culture at Google | Ramon Medrano Llamas | Conf42 SRE 2022](https://www.youtube.com/watch?v=qgHWzQ2zcqQ) · Conf42 · 23 min: how a Google SRE writes and uses blameless postmortems, with real examples of system fixes over people fixes.
- [You Must Be Crazy To Do Pair Programming • Dave Farley • GOTO 2022](https://www.youtube.com/watch?v=t92iupKHo8M) · GOTO Conferences · 25 min: why pairing is often faster than working alone, and how to make it work.

## Go deeper

- [Google's Engineering Practices: code review](https://google.github.io/eng-practices/review/): Google's public guide for reviewers and authors, including the "better than before" standard and how to write comments.
- [Conventional Comments](https://conventionalcomments.org/): a simple format for labelling review comments so authors know what matters.
- [How to Write a Git Commit Message](https://cbea.ms/git-commit/) (Chris Beams): the seven rules behind good commit messages, with examples.
- [Design Docs at Google](https://www.industrialempathy.com/posts/design-docs-at-google/) (Malte Ubl): what goes in a design doc, when one is worth writing, and how they get reviewed.
- [Postmortem Culture: Learning from Failure](https://sre.google/sre-book/postmortem-culture/) (Google SRE book): the chapter that made blameless postmortems standard practice in operations teams.
