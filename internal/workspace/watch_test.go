package workspace

import "testing"

func TestRelevantIgnoresEditorTempFilesAndScratchRepos(t *testing.T) {
	for path, want := range map[string]bool{
		"/c/exercises/01_shell/001_pipes/count.sh":        true,
		"/c/exercises/03_sql/020_joins/report.sql":        true,
		"/c/exercises/01_basics/003_vars/exercise.go":     true,
		"/c/exercises/x/testdata/app.log":                 true,
		"/c/exercises/x/.exercise.go.swp":                 false,
		"/c/exercises/x/exercise.go~":                     false,
		"/c/exercises/x/4913":                             false,
		"/c/exercises/02_git/030_branches/sandbox/a.txt":  false,
		"/c/exercises/02_git/030_branches/sandbox/.git/x": false,
	} {
		if got := relevant(path); got != want {
			t.Errorf("relevant(%s) = %v, want %v", path, got, want)
		}
	}
}
