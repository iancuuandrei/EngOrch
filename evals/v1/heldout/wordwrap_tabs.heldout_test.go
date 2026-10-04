package wordwrap

import (
	"bytes"
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestFabricV1Heldout(t *testing.T) {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("targeted assertion: cannot locate held-out test source")
	}
	root := filepath.Dir(testFile)
	sources, err := filepath.Glob(filepath.Join(root, "*.go"))
	if err != nil {
		t.Fatalf("list package source: %v", err)
	}
	found := false
	for _, sourcePath := range sources {
		if strings.HasSuffix(sourcePath, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), sourcePath, nil, 0)
		if err != nil {
			t.Fatalf("parse package source: %v", err)
		}
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil && fn.Name.Name == "WrapStringWithTabWidth" {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Error("targeted assertion: missing public WrapStringWithTabWidth API")
		return
	}

	// Compile an external consumer only after the API-presence assertion. This
	// keeps the pristine baseline failure an assertion rather than a build error.
	consumerDir, err := os.MkdirTemp(root, "fabric-heldout-tabwidth-")
	if err != nil {
		t.Fatalf("create temporary consumer: %v", err)
	}
	defer os.RemoveAll(consumerDir)
	const consumer = `package main

import (
	"fmt"
	wordwrap "github.com/mitchellh/go-wordwrap"
)

func main() {
	cases := []struct {
		name string
		input string
		limit, tabWidth uint
		want string
	}{
		{"next stop", "a\tb", 5, 4, "a   b"},
		{"wrap after tab", "a\tb", 3, 4, "a\nb"},
		{"consecutive tabs", "\t\tX", 9, 4, "        X"},
		{"unicode code point", "界\tX", 5, 4, "界   X"},
		{"LF resets stop", "a\nb\tc", 5, 4, "a\nb   c"},
	}
	for _, tc := range cases {
		got := wordwrap.WrapStringWithTabWidth(tc.input, tc.limit, tc.tabWidth)
		if got != tc.want {
			panic(fmt.Sprintf("%s: got %q, want %q", tc.name, got, tc.want))
		}
	}
	if got, want := wordwrap.WrapStringWithTabWidth("a\tb", 3, 0), wordwrap.WrapString("a\tb", 3); got != want {
		panic(fmt.Sprintf("zero tab width: got %q, want legacy result %q", got, want))
	}
}
`
	mainPath := filepath.Join(consumerDir, "main.go")
	if err := os.WriteFile(mainPath, []byte(consumer), 0o600); err != nil {
		t.Fatalf("write external consumer: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	goExe := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		goExe += ".exe"
	}
	cmd := exec.CommandContext(ctx, goExe, "run", ".")
	cmd.Dir = consumerDir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Run(); err != nil {
		t.Fatalf("public API behavior check failed: %v (%s)", err, output.String())
	}
}
