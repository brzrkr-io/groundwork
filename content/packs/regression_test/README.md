# Practice: reproduce a bug report

When a bug report comes in, experienced engineers don't jump straight to the
fix. They first write a test that fails because of the bug, then fix the code
and watch the test pass. The test stays forever, so the bug can't quietly
return. This is a *regression test*.

**New here:** nothing; practice.

## Learn

### Red, green, keep

1. **Red.** Add a test case for the exact input from the bug report. Run it
   and watch it fail. If it passes, you haven't reproduced the bug yet.
2. **Green.** Fix the code until the test passes.
3. **Keep.** Leave the case in the table. It now guards against the bug
   coming back.

### Recap: table tests with error cases

Use a table-driven test with subtests, as in *Table-driven tests* and
*Subtests and helpers*. Cases that should fail get a `wantErr bool` field:

```go
tests := []struct {
	name    string
	input   string
	want    int
	wantErr bool
}{
	{name: "plain", input: "8080", want: 8080},
	{name: "not a number", input: "http", wantErr: true},
}
```

> **Why is it written like this?** Field names in rows (`name: ...`) are your
> choice, but they let you leave out fields you don't need: a missing field
> gets its zero value, so `wantErr` is `false` unless you set it. Calling the
> fields `want` and `wantErr` is convention.

A good fix also mustn't go too far, so add cases showing that bad input is
still rejected.

### Grading

`TestGrader` re-runs your test against broken versions (*mutants*) of the
code. Here two mutants are the original bugs, so your test must catch them.
While the real code still has a bug, `TestGrader` says your tests must pass
on the real code first. That's expected between the red and green steps.

## Your task

The deploy command calls `NormalizeEnvironment` on the environment name the
user types. It accepts `development`, `staging` and `production`, and returns
the name in lower case:

```go
NormalizeEnvironment("production")  // => "production", nil
NormalizeEnvironment("prod")        // => "", error
```

Two bug reports have arrived.

> **Bug report #1:** Deploys to "Production" with a capital P are rejected
> with `unknown environment "Production"`. The case of the name shouldn't
> matter.

> **Bug report #2:** Pasting the environment name from our wiki adds a space
> at the end, and `"staging "` is rejected. Spaces around the name should be
> ignored.

Write all your cases in `TestNormalizeEnvironment` in `exercise_test.go`.

### 1. Reproduce bug report #1

**Goal:** write a table test with cases showing that capital letters are
accepted and lowered.

| Input | Want |
|---|---|
| `"production"` | `"production"`, no error |
| `"Production"` | `"production"`, no error |
| `"STAGING"` | `"staging"`, no error |

Check the returned name, not just the error. Run the tests: the bug report
cases fail. That's the red step.

**Think it through:**
- Which fields does each row need: the input, the wanted name, and what else?
- What should the loop check first, the error or the name?

### 2. Fix bug report #1

**Goal:** make `realNormalizeEnvironment` in `exercise.go` accept the name in
any case and return it in lower case.

`strings.ToLower` does the conversion; add `"strings"` to the imports. Your
test from task 1 now passes.

**Think it through:**
- The `switch` compares one value. Which value should it compare, and which
  should the function return?

### 3. Reproduce bug report #2

**Goal:** add rows showing that spaces around the name are ignored, and
watch them fail.

| Input | Want |
|---|---|
| `"staging "` | `"staging"`, no error |
| `"  Development "` | `"development"`, no error |

**Think it through:**
- Does anything in the loop need to change, or only the table?

### 4. Fix bug report #2

**Goal:** make `realNormalizeEnvironment` ignore spaces around the name.

`strings.TrimSpace` removes them.

**Think it through:**
- You already clean the name in one place. Can the new step go there too?
- Should the spaces go before or after lower-casing? Does it matter?

### 5. Don't over-fix

**Goal:** add rows checking that `"prod"` and `""` (an empty name) both
return an error.

They catch a "fix" that accepts anything, or one that quietly turns an empty
name into `"development"`.

**Think it through:**
- Which field marks a row that must fail, and which field can you leave out?

The exercise is done when `TestNormalizeEnvironment` passes and `TestGrader`
reports every mutant caught.

## Learn more

- [Go wiki: Table-driven tests](https://go.dev/wiki/TableDrivenTests)
- [strings.ToLower](https://pkg.go.dev/strings#ToLower) and [strings.TrimSpace](https://pkg.go.dev/strings#TrimSpace)
