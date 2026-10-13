# Practice: reproduce a bug report

When a bug report comes in, experienced engineers don't jump straight to the
fix. They first write a test that fails because of the bug. Then they fix the
code and watch the test pass. The test stays in the codebase forever, so the
same bug can never quietly return. This is called a *regression test*.

## Learn

### Recap

The workflow is red, green, keep:

1. **Red.** Write a test case for the exact input from the bug report. Run
   it and watch it fail. If it passes, you haven't reproduced the bug yet.
2. **Green.** Fix the code until the test passes.
3. **Keep.** Leave the case in the table. It now guards against the bug
   coming back.

Use a table-driven test with subtests, as in *Table-driven tests* and
*Subtests and helpers*. Cases that should fail need a `wantErr bool` field.
A good fix also mustn't go too far, so add cases showing that bad input is
still rejected.

As before, `TestGrader` re-runs your test against broken versions of the code.
Here one mutant is the original bug itself, so your test must catch it.
While the real code still has a bug, `TestGrader` reports that your tests
must pass on the real code first. That is expected between the red and green
steps.

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

Add cases checking that `"Production"` returns `"production"` and
`"STAGING"` returns `"staging"`, both with no error. Also keep a plain case,
`"production"` returns `"production"`. Check the returned name, not just the
error. Run the tests: the bug report cases fail. That's the red step.

### 2. Fix bug report #1

In `exercise.go`, make `realNormalizeEnvironment` accept the name in any
case and return it in lower case. `strings.ToLower` does the conversion;
you'll need to add `"strings"` to the imports. Your test from task 1 now
passes.

### 3. Reproduce bug report #2

Add cases checking that `"staging "` returns `"staging"` and
`"  Development "` returns `"development"`. Watch them fail.

### 4. Fix bug report #2

Make `realNormalizeEnvironment` ignore spaces around the name.
`strings.TrimSpace` removes them.

### 5. Don't over-fix

Add cases checking that `"prod"` and `""` (an empty name) both return an
error. They catch a "fix" that accepts anything, or one that quietly turns an
empty name into `"development"`.

The exercise is done when `TestNormalizeEnvironment` passes and
`TestGrader` reports every mutant caught.

## Learn more

- [Go wiki: Table-driven tests](https://go.dev/wiki/TableDrivenTests)
- [strings.ToLower](https://pkg.go.dev/strings#ToLower) and [strings.TrimSpace](https://pkg.go.dev/strings#TrimSpace)
