package progress

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writeProgress(t *testing.T, body string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, Dir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(root), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestLoadRefusesFoldersFromANewerGroundwork(t *testing.T) {
	_, err := Load(writeProgress(t, `{"version": 99, "done": {}}`))
	if !errors.Is(err, ErrNewer) {
		t.Fatalf("Load(version 99) error = %v, want ErrNewer", err)
	}
}

func TestLoadUpgradesTheLastFormatKeepingProgress(t *testing.T) {
	s, err := Load(writeProgress(t, `{"version": 3, "current": "maps", "done": {"hello_go": "2026-01-01T00:00:00Z"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, done := s.Done["hello_go"]; !done || s.Current != "maps" || s.Version != version {
		t.Fatalf("upgraded state = %+v, want hello_go done, current maps, version %d", s, version)
	}
	if day := s.Activity[DateKey(s.Done["hello_go"])]; day.Passed != 1 {
		t.Errorf("activity rebuilt from older progress = %+v, want hello_go's pass on its day", s.Activity)
	}
	if s.Folders == nil || s.Revealed == nil || s.Starters == nil || s.Read == nil || s.Reviews == nil || s.Activity == nil {
		t.Fatal("upgraded state has nil maps; writes to them would panic")
	}
}
