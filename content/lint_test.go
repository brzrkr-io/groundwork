package content_test

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pjanderson/gopherlings/content"
	"github.com/pjanderson/gopherlings/internal/course"
	"github.com/pjanderson/gopherlings/internal/workspace"
)

var lintPacks = flag.String("packs", "", "comma-separated pack names for TestLint (default: all)")

// goplsSettings mirrors a learner's editor with the stricter checks turned
// on: staticcheck plus the opt-in analyzers. Every warning, hint and
// suggestion an editor would show must be the lesson, not noise.
var goplsSettings = map[string]any{
	"staticcheck": true,
	"analyses":    map[string]bool{"nilness": true, "unusedparams": true, "unusedwrite": true},
}

// starterDiagnostics lists, per pack, the diagnostic sources a starter may
// show because they point straight at the bug the learner is there to fix.
// Solutions may show nothing at all.
var starterDiagnostics = map[string][]string{
	"variables_2":           {"compiler", "nonewvars"},
	"constants":             {"compiler"},
	"constants_2":           {"compiler"},
	"string_formatting_2":   {"printf"},
	"type_conversion_2":     {"compiler"},
	"if_else_2":             {"compiler"},
	"maps_2":                {"SA5000", "nilness"},
	"structs_2":             {"unusedwrite"},
	"pointers_2":            {"SA4009", "SA4006", "unusedwrite"},
	"pointer_receivers":     {"SA4005", "unusedwrite"},
	"pointer_receivers_2":   {"SA4005"},
	"errors_2":              {"ST1005"},
	"panic_recover_2":       {"SA4006", "SA4017", "SA5000", "nilness"},
	"reading_lines_2":       {"scannererr"},
	"json_encoding_2":       {"SA9005", "structtag"},
	"json_decoding_2":       {"structtag", "SA5008", "unmarshal", "SA1014"},
	"packages":              {"compiler", "go list"},
	"packages_2":            {"compiler"},
	"exported_names":        {"compiler", "unusedfunc"},
	"exported_names_2":      {"compiler"},
	"internal_packages":     {"compiler"},
	"internal_packages_2":   {"compiler", "go list"},
	"embed_2":               {"SA9009"},
	"benchmarks":            {"stringsbuilder"},
	"sdk_interfaces":        {"any"},
	"standard_streams_2":    {"scannererr"},
	"walking_directories_2": {"SA4009"},
	"label_selectors_2":     {"nilness"},
	"reconcile_2":           {"SA9003"},
}

// TestLint opens every starter and solution in gopls, as an editor would.
// It needs gopls on PATH and takes a minute or two, so -short skips it.
//
//	go test ./content -run TestLint -args -packs=variables,variables_2
func TestLint(t *testing.T) {
	if testing.Short() {
		t.Skip("slow: opens every pack in gopls")
	}
	if _, err := exec.LookPath("gopls"); err != nil {
		t.Skip("gopls not on PATH")
	}
	c, err := course.Load(content.Files)
	if err != nil {
		t.Fatal(err)
	}
	var exercises []*course.Exercise
	for _, e := range c.Exercises {
		if *lintPacks == "" || slices.Contains(strings.Split(*lintPacks, ","), e.Name) {
			exercises = append(exercises, e)
		}
	}
	if len(exercises) == 0 {
		t.Fatalf("no pack matches -packs=%s", *lintPacks)
	}
	for _, kind := range []string{"starter", "solution"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			for _, e := range exercises {
				pack, err := c.Pack(e)
				if err != nil {
					t.Fatal(err)
				}
				if err := workspace.Materialize(pack, filepath.Join(root, filepath.FromSlash(e.Dir())), kind == "solution"); err != nil {
					t.Fatal(err)
				}
			}
			diags, err := goplsDiagnostics(root)
			if err != nil {
				t.Fatal(err)
			}
			for _, e := range exercises {
				allowed := starterDiagnostics[e.Name]
				if kind == "solution" {
					allowed = nil
				}
				for _, d := range diags {
					if strings.HasPrefix(d.file, e.Dir()+"/") && !slices.Contains(allowed, d.source) {
						t.Errorf("%s:%d: [%s] %s", d.file, d.line, d.source, d.message)
					}
				}
			}
		})
	}
}

type diagnostic struct {
	file, source, message string
	line                  int
}

// goplsDiagnostics opens every .go file under root in one gopls session and
// returns everything it publishes once it has gone quiet.
func goplsDiagnostics(root string) ([]diagnostic, error) {
	cmd := exec.Command("gopls")
	cmd.Env = append(os.Environ(), "GOWORK=off")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	defer cmd.Process.Kill()

	lsp := &lspConn{w: stdin, published: map[string][]lspDiagnostic{}, activity: make(chan struct{}, 1)}
	go lsp.read(bufio.NewReader(stdout))

	uri := "file://" + root
	lsp.call(1, "initialize", map[string]any{
		"processId":             os.Getpid(),
		"rootUri":               uri,
		"workspaceFolders":      []map[string]string{{"uri": uri, "name": "packs"}},
		"initializationOptions": goplsSettings,
		"capabilities": map[string]any{
			"workspace": map[string]any{"configuration": true, "workspaceFolders": true},
		},
	})
	lsp.waitQuiet(time.Second)
	lsp.call(0, "initialized", map[string]any{})

	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && entry.Name() == "testdata" {
			return fs.SkipDir
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		text, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		lsp.call(0, "textDocument/didOpen", map[string]any{"textDocument": map[string]any{
			"uri": "file://" + path, "languageId": "go", "version": 1, "text": string(text),
		}})
		return nil
	})
	if err != nil {
		return nil, err
	}
	lsp.waitQuiet(15 * time.Second)

	lsp.mu.Lock()
	defer lsp.mu.Unlock()
	var out []diagnostic
	for fileURI, ds := range lsp.published {
		rel := strings.TrimPrefix(strings.TrimPrefix(fileURI, uri), "/")
		for _, d := range ds {
			out = append(out, diagnostic{file: rel, line: d.Range.Start.Line + 1, source: d.Source, message: d.Message})
		}
	}
	slices.SortFunc(out, func(a, b diagnostic) int {
		if c := strings.Compare(a.file, b.file); c != 0 {
			return c
		}
		return a.line - b.line
	})
	return out, nil
}

type lspDiagnostic struct {
	Range struct {
		Start struct{ Line int }
	}
	Source  string
	Message string
}

type lspConn struct {
	w         io.Writer
	wmu       sync.Mutex
	mu        sync.Mutex
	published map[string][]lspDiagnostic
	activity  chan struct{}
}

func (c *lspConn) send(msg map[string]any) {
	msg["jsonrpc"] = "2.0"
	body, _ := json.Marshal(msg)
	c.wmu.Lock()
	defer c.wmu.Unlock()
	fmt.Fprintf(c.w, "Content-Length: %d\r\n\r\n%s", len(body), body)
}

// call sends a request (id > 0) or a notification (id 0).
func (c *lspConn) call(id int, method string, params any) {
	msg := map[string]any{"method": method, "params": params}
	if id > 0 {
		msg["id"] = id
	}
	c.send(msg)
}

func (c *lspConn) read(r *bufio.Reader) {
	for {
		length := 0
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimSpace(line)
			if line == "" {
				break
			}
			if v, ok := strings.CutPrefix(strings.ToLower(line), "content-length:"); ok {
				length, _ = strconv.Atoi(strings.TrimSpace(v))
			}
		}
		body := make([]byte, length)
		if _, err := io.ReadFull(r, body); err != nil {
			return
		}
		var msg struct {
			ID     json.RawMessage
			Method string
			Params json.RawMessage
		}
		if json.Unmarshal(body, &msg) != nil {
			continue
		}
		switch {
		case msg.Method == "textDocument/publishDiagnostics":
			var p struct {
				URI         string
				Diagnostics []lspDiagnostic
			}
			if json.Unmarshal(msg.Params, &p) == nil {
				c.mu.Lock()
				c.published[p.URI] = p.Diagnostics
				c.mu.Unlock()
			}
		case msg.Method != "" && msg.ID != nil: // a request from gopls
			var result any
			if msg.Method == "workspace/configuration" {
				var p struct{ Items []json.RawMessage }
				json.Unmarshal(msg.Params, &p)
				items := make([]any, len(p.Items))
				for i := range items {
					items[i] = goplsSettings
				}
				result = items
			}
			c.send(map[string]any{"id": msg.ID, "result": result})
		}
		select {
		case c.activity <- struct{}{}:
		default:
		}
	}
}

// waitQuiet returns once gopls has sent nothing for the given duration.
func (c *lspConn) waitQuiet(quiet time.Duration) {
	for {
		select {
		case <-c.activity:
		case <-time.After(quiet):
			return
		}
	}
}
