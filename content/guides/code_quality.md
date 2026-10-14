# Code for humans: naming, simplicity and refactoring

**In one sentence:** good code is code the next person can read, trust and change safely, and you get there by choosing clear names, keeping things small and flat, and improving the shape of working code in small steps under tests.

**Why it matters:** the deploy script you write today will be read during an incident at 3 a.m. by someone tired and stressed, possibly you. Every cryptic name, hidden number and five-deep `if` costs them minutes they don't have. Platform code also lives for years and gets changed by many hands; code that's easy to change is code that stays correct.

**You'll learn:**

- Why reading cost, not writing cost, is what to optimise
- How to choose names that say what things are, and that never lie
- Why small functions with one job, flat guard clauses and named constants make code easier to follow
- When duplication is a problem, and when removing it makes things worse
- What refactoring is, and how to do it safely in small steps under tests
- How to recognise common code smells, and the rules of thumb KISS and YAGNI

## Code is read far more than it is written

You write a function once. Then it's read in review, when someone adds a feature next to it, when a test fails, when an alert fires, and when a new teammate learns the system. So the useful question is not "how fast can I write this?" but **"how fast can someone else understand this?"**. Saving ten seconds of typing at the cost of every reader's minute is a bad trade.

The reader is not stupid; they're **busy**, with their own problem in their head. Good code gets out of their way: it says what it does, plainly, where they're already looking.

> **The 3 a.m. test.** Imagine you're paged, half awake, and this function is in the stack trace. Can you tell what it does without scrolling?

## Names are the first documentation

### What a good name does

Read this line from an alert bot:

```go
if m[x] >= n {
```

Now this one:

```go
if counts[service] >= pageThreshold {
```

Both run the same way. The second tells you what the map holds, what the key is, and what the limit means; the first sends you hunting up the file. A good name answers "what is this?" **before** the reader has to ask.

Names for values say **what the value is** (`service`, `retryDelay`). Names for functions say **what they do or return** (`parseHostPort`, `isExpired`). Use the words your team says out loud: if on-call talks about "pages", call it a page, not a "notification event".

### Length follows scope

Go likes short names, but how short depends on how far the name travels:

| Where the name lives | Good names |
|---|---|
| a 3-line loop | `i`, `n`, `s` |
| a 20-line function | `line`, `counts`, `service` |
| a whole package | `criticalService`, `maxRetries` |

The further the reader is from the line that created a value, the more its name must explain. In a long function, one-letter names force the reader to keep a private dictionary in their head.

### Put units in names

`timeout := 30` compiles whether 30 is seconds or milliseconds, and mixing them up is a classic production bug: a retry meant to wait half a second waits eight minutes. Say the unit: `timeoutSeconds`, `sizeBytes`, `delayMillis`. Better still, use a type that carries it, like `time.Duration`: `30 * time.Second` can't be mistaken for milliseconds.

### Names that lie are worse than bad names

A bad name slows the reader down. A *misleading name* makes them write a bug. If `filterHealthy` actually returns the **un**healthy nodes, every caller who trusted the name and didn't read the body is now wrong.

A sneakier lie is a hidden side effect. `ownerOf(owners, service)` sounds like a harmless lookup; if it also writes a default into the map, the name hides that. Rename it so the effect is visible (`claimOwner`), or remove the effect so the name is true.

### Go's own conventions

| Convention | Example |
|---|---|
| *MixedCaps*, never underscores | `pageLine`, not `page_line` |
| Initialisms keep one case | `serviceURL`, `podID` |
| Exported names start upper case | `Plan` is public, `plan` is private |
| Don't repeat the package name | `rollout.Plan`, not `rollout.RolloutPlan` |
| `err` for errors, `ok` for "did it work?" | `value, ok := m[key]` |

Rename with your editor's **Rename Symbol**, not find-and-replace: it changes the declaration and every use, and nothing else. Recent editors also fix the first word of the function's own comment, but other comments are just text, so search for the old name afterwards. *Naming things* and *Naming things: fix the misleading names* practise all of this.

## Small functions with one job

### How to spot a function doing too much

A function that validates a request, builds an image reference, plans the batches and formats the output is doing four jobs. You notice it by:

- **Length.** If it doesn't fit on one screen, you can't see it all at once.
- **"And".** If describing it needs "and" ("checks the request **and** builds the plan **and**..."), each "and" is a candidate function.
- **Comments that name a block.** `// Build the image reference.` above five lines is a function name waiting to happen.

### Extract function

*Extract function* moves a block into its own function whose name replaces the comment. The original reads like a table of contents:

```go
func Plan(r Request) ([]string, error) {
	if err := validate(r); err != nil {
		return nil, err
	}
	steps := []string{"deploy " + imageRef(r)}
	steps = append(steps, batchSteps(r)...)
	return append(steps, "verify "+r.Service), nil
}
```

You can understand `Plan` in five seconds and dive into `batchSteps` only if batching is your problem. Each helper is also small enough to test and review on its own. This is the decomposition idea from the *Functions and decomposition* guide, applied to code that already exists. *Extract function* walks through it step by step.

## Keep the happy path on the left

Nested checks drift right like an arrow, and each `else` ends up far from the `if` it belongs to:

```go
if c.Service != "" {
	if !c.Frozen {
		if c.TestsPassed {
			return "ok"
		} else {
			return "tests failed"
		}
	} else {
		return "change freeze in effect"
	}
} else {
	return "missing service name"
}
```

A *guard clause* handles a special case first and leaves (with `return`, or `continue` in a loop):

```go
if c.Service == "" {
	return "missing service name"
}
if c.Frozen {
	return "change freeze in effect"
}
if !c.TestsPassed {
	return "tests failed"
}
return "ok"
```

Each problem now sits next to its check, and the last line is the *happy path*: what happens when everything is fine. This is the shape Go's own style guide asks for ("indent error flow"), and it's why Go code checks `if err != nil { return ... }` and carries on down the left edge.

Turning a check around is where refactors go wrong. The opposite of `x < 80` is `x >= 80`, not `x > 80`. The opposite of `a && b` is `!a || !b` (*De Morgan's laws*). And a guard must leave: one that records a reason but forgets `continue` falls straight into the normal path. *Guard clauses* and *Guard clauses: fix the flipped conditions* practise both sides.

## Name your numbers

A *magic number* is a bare number whose meaning the reader has to guess:

```go
if daysLeft <= 30 {
```

Thirty what, and says who? A named constant answers both, and keeps the rule in one place:

```go
const warningDays = 30 // security policy SEC-12

if daysLeft <= warningDays {
```

The second benefit is the big one. When the policy changes to 21 days, you change one line, instead of hunting for every `30` and hoping none of them meant something else (like 30 seconds). Not every number is magic: `0` in `count == 0` or `1` in `i + 1` means exactly what it says. Name the numbers that come from a **decision**. *Named constants* shows the payoff.

## Duplication, and the wrong abstraction

### Copies drift

When the same rule lives in three copies, a fix lands in one and the other two keep the bug. Worse, copies drift apart quietly: someone pastes a port check for `admin_port` and forgets to change the message from `metrics_port`. The principle *DRY*, "Don't Repeat Yourself", says every piece of knowledge should have **one** home.

Duplication isn't only copied lines. A fact written twice (the list of environments in a `switch` here and in an error message there) is duplicated knowledge, and it drifts just the same. *Removing duplication* fixes both kinds.

### Similar is not the same

Here is the trap. Two blocks that **look** alike today may encode two different rules that will change for different reasons. Merge them, and the first time they need to differ, someone adds a `bool` parameter. Then another. Soon the shared function is a maze of flags.

That is *the wrong abstraction*, and as the *Abstraction* guide explains, it costs more than the duplication it replaced. Rules of thumb:

- Merge code that must **change together** because it is the same rule.
- Leave code that only looks alike. Some teams wait for the third copy ("rule of three") before abstracting.
- If a shared function grows flags, it's fine to copy it back out into its callers and start again.

## Comments that say why

Code already says **what** it does. A comment that repeats it is noise, and noise that rots: the code changes, the comment doesn't, and now it lies.

```go
// add one to retries
retries++
```

The comments worth writing say **why**: the reason a reader can't see in the code.

```go
// The registry rate-limits us to 10 pulls a minute, so back off for longer
// than the API's own Retry-After suggests.
delay *= 2
```

Comments also earn their place when they link to the incident or ticket behind a strange rule, or say what callers may rely on (Go's doc comments on exported names). If a block seems to need a "what" comment, try a better name or an extracted function first. And delete commented-out code: version control remembers it.

## Refactoring: changing shape, not behaviour

### What refactoring is

*Refactoring* is changing the structure of code **without changing what it does**. Renaming, extracting a function, flattening an `if`: all refactorings. Fixing a bug or adding a feature is not, and mixing the two in one change makes both harder to review.

### Tests first, then small steps

You can only be sure behaviour didn't change if something checks it. So:

1. Make sure tests cover the code you're about to touch. If they don't, write them first (the *Testing and correctness* guide covers how).
2. Make **one** small change: one rename, one extraction.
3. Run the tests. Green: commit or carry on. Red: **undo** the last step rather than debugging it. It was small, so you lose almost nothing.

Small steps feel slow and are fast. A big rewrite that breaks leaves you searching 300 changed lines; a small step that breaks tells you exactly where.

### Four refactorings you'll use every week

| Refactoring | When | What you do |
|---|---|---|
| *Rename* | a name is unclear or lies | Rename Symbol, then check comments for the old name |
| *Extract function* | a block has a job and a comment naming it | move it into a function named after the job |
| *Introduce parameter object* | a function takes many arguments, or several bools | group them into a struct with named fields |
| *Replace conditional with table* | a long `switch` maps keys to values | put the data in a map and look it up |

Introduce parameter object fixes calls like this one, where nobody can tell which `true` means what:

```go
Deploy("api", "api:v2", 3, 1, true, false)
```

With an options struct, every argument has a label, and the ones you leave out get their zero value:

```go
Deploy(DeployOptions{Service: "api", Image: "api:v2", Replicas: 3, MaxSurge: 1, DryRun: true})
```

Replace conditional with table turns a twenty-case `switch` into a map you can read at a glance, add a row to in one line, and even loop over. *Lookup tables* practises it.

## Code smells

A *code smell* is a surface sign that **might** point to a deeper problem. It's a reason to look closer, not an automatic verdict.

| Smell | What you notice | Usual refactoring |
|---|---|---|
| Long function | doesn't fit on a screen; needs "and" to describe | extract function |
| Deep nesting | the arrow shape, `else` far from its `if` | guard clauses |
| Magic number | a bare `86400` or `30` | named constant |
| Duplicated code | the same block, or the same fact, in several places | extract function, one shared list |
| Long parameter list | five or more arguments, several bools | introduce parameter object |
| Cryptic or lying name | `tmp`, `data2`, `filterHealthy` that does the opposite | rename |
| Long switch on a value | every case returns data | replace conditional with table |
| Comment explaining "what" | the comment repeats the code | rename or extract, then delete the comment |

## Simple beats clever

*KISS*, "Keep It Simple", means choosing the plainest code that does the job. A clever one-liner that needs a paragraph to explain is worse than four obvious lines. Go was designed around this idea, as Rob Pike's talk below explains: few features, one obvious way to do most things.

*YAGNI*, "You Aren't Gonna Need It", means not building for a future that may never come. The plugin system for a tool with one plugin, the config option nobody asked for: each costs reading and testing time today, for a benefit that rarely arrives. When the need is real, you'll know its shape, and refactoring under tests makes adding it cheap.

Both rules have the same root: every line of code is a liability someone has to read, test and keep working. Write the ones that earn their place.

## Common mistakes

- **Refactoring without tests.** Without a safety net, "it's just a rename" is how a typo ships. Write the tests first.
- **Refactoring and changing behaviour in the same step.** When something breaks, you can't tell which change did it, and reviewers can't either. Separate commits, ideally separate PRs.
- **Taking big steps.** Rewriting a whole function at once, then debugging it for an hour. Take steps small enough that undo is always cheap.
- **Deduplicating by appearance.** Merging two blocks that look alike but change for different reasons, then adding flags to tell them apart.
- **Abbreviating to save typing.** `cfgMgrSvc` saves four keystrokes and costs every reader a decode. Editors autocomplete; spell it out.
- **Flipping a condition carelessly.** Inverting `< 80` into `> 80` or `a && b` into `!a && !b` while flattening. Test the boundary values in both versions.
- **Comments that describe the code.** They rot. Explain why, or improve the name.

## Check yourself

1. Why is "how fast can someone else understand this code?" a better goal than "how fast can I write it?"
2. A variable in a 30-line function is named `t`. Why is that worse than an `i` in a 3-line loop, and what would you rename it to if it holds the current time?
3. A function is called `getConfig`, and besides reading the config it creates a default config file when none exists. What's wrong with the name, and what are the two honest fixes?
4. A disk check reads `if usedPercent < 80 { return "ok" }`, and you flip it into a guard that returns `"warning"` first. What is the guard's condition, and what mistake is easy to make?
5. Two functions both check that a port is between 1 and 65535, with the same message format. Another two functions both happen to compare a number with 30: one for days until a certificate expires, one for seconds of request timeout. Which pair should share code, and why?
6. What makes a change a refactoring, and why should a refactoring and a bug fix go in separate commits?
7. A function call reads `Deploy("api", "v2", 3, true, false, true)`. Which smell is this, and which refactoring fixes it?
8. What does YAGNI say about adding a plugin system to a log rotator that has exactly one way to rotate logs?

## Answers

1. Code is read many more times than it is written: in review, when adding features, during incidents, by new teammates. Saving the writer a few seconds while costing every reader minutes is a bad trade.
2. The further a name is from where its value was created, the more the name must explain. In a 3-line loop the reader sees everything at once; in 30 lines they lose track of what `t` meant (and in Go, `t` usually means a `*testing.T`). `now` says exactly what it holds.
3. The name promises a read with no side effects, but it writes a file, so callers will use it where creating files is wrong. Rename it so the effect is visible (`loadOrCreateConfig`), or remove the side effect so the name is true.
4. The inverted condition is `usedPercent >= 80`. The easy mistake is writing `usedPercent > 80`, which sends exactly 80 to the wrong branch. The opposite of `<` is `>=`; check the boundary value in both versions.
5. The port checks should share one function: they are the same rule and must change together. The two 30s only look alike; they mean different things and will change for different reasons, so merging them would be the wrong abstraction. Each gets its own named constant.
6. A refactoring changes the structure of code without changing its behaviour, so the tests stay green throughout. Kept apart from a bug fix, each change can be reviewed on its own terms, and if something breaks, you know which change did it.
7. A long parameter list, made worse by bare booleans that nobody can tell apart. Introduce parameter object: pass a struct like `DeployOptions{Service: "api", Tag: "v2", Replicas: 3, DryRun: true}` so every value has a label.
8. Don't build it until a second way to rotate actually exists; by then you'll know what a plugin needs to look like. Until then it's extra code to read, test and maintain for no benefit.

## Key terms

| Term | Meaning |
|---|---|
| refactoring | changing the structure of code without changing what it does, ideally in small steps under tests |
| misleading name | a name that suggests something the code doesn't do, so readers who trust it write bugs |
| side effect | a change a function makes beyond returning a value, such as writing to a map, a file or a counter |
| extract function | moving a block of code into its own function whose name says what the block does |
| guard clause | an early check that handles a special case and leaves with return or continue, keeping the main path flat |
| happy path | the code that runs when everything is fine, ideally down the left edge of a function |
| magic number | a bare number in code whose meaning the reader has to guess |
| DRY | "Don't Repeat Yourself": each rule or fact should live in exactly one place |
| the wrong abstraction | merging code that only looks alike, so the shared function fills up with flags for each caller |
| introduce parameter object | grouping many function arguments into a struct with named fields |
| replace conditional with table | turning a switch that maps keys to values into a map lookup |
| code smell | a surface sign in code that might point to a deeper design problem |
| KISS | "Keep It Simple": prefer the plainest code that does the job |
| YAGNI | "You Aren't Gonna Need It": don't build features or flexibility before there is a real need |

## Watch

- [Naming Things in Code](https://www.youtube.com/watch?v=-J3wNP6u5YU) · CodeAesthetic · 7 min: single letters, abbreviations, units in names and more, with clear before-and-after examples.
- [Why You Shouldn't Nest Your Code](https://www.youtube.com/watch?v=CFRhGnuXG-4) · CodeAesthetic · 8 min: extraction and inversion (guard clauses) shown visually; a perfect companion to *Guard clauses*.
- [RailsConf 2016 - Get a Whiff of This by Sandi Metz](https://www.youtube.com/watch?v=PJjHfa5yxlU) · Confreaks · 38 min: a friendly tour of code smells and the refactoring that cures each one, refactored live in small steps. The code is Ruby, but the ideas transfer directly.
- [dotGo 2015 - Rob Pike - Simplicity is Complicated](https://www.youtube.com/watch?v=rFejpH_tAHM) · dotconferences · 23 min: why Go was designed to be simple, and why simplicity takes real work to achieve.

## Go deeper

- [Catalog of Refactorings](https://refactoring.com/catalog/) (Martin Fowler): every named refactoring with a short example; look up Extract Function, Introduce Parameter Object and Replace Nested Conditional with Guard Clauses.
- [Code Smell](https://martinfowler.com/bliki/CodeSmell.html) (Martin Fowler): a one-page explanation of what a smell is, and why it's a hint rather than a rule.
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments): the Go team's checklist of naming, error flow and comment conventions that reviewers apply to Go code.
- [The Wrong Abstraction](https://sandimetz.com/blog/2016/1/20/the-wrong-abstraction) (Sandi Metz): the short essay behind "duplication is far cheaper than the wrong abstraction", and how to back out of one.
- [Yagni](https://martinfowler.com/bliki/Yagni.html) (Martin Fowler): the costs of building features before you need them, laid out one by one.
