package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"harness.local/engorch/internal/config"
	"harness.local/engorch/internal/control"
	"harness.local/engorch/internal/repository"
	"harness.local/engorch/internal/ri"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestActualScipTypeScriptCLI exercises a real scip-typescript 0.4.0 index
// through the admitted controller lifecycle. It requires an explicit Node
// binary, the published entry script and a built Rust RI binary.
func TestActualScipTypeScriptCLI(t *testing.T) {
	node := os.Getenv("ENGORCH_SCIP_TYPESCRIPT_NODE")
	entry := os.Getenv("ENGORCH_SCIP_TYPESCRIPT_ENTRY")
	rust := os.Getenv("ENGORCH_RI_BINARY")
	if node == "" || entry == "" || rust == "" {
		t.Skip("requires scip-typescript node, entry and Rust RI")
	}
	// Normalize executable paths so forward-slash Windows values satisfy the
	// normalized absolute-path checks in import planning.
	node = filepath.Clean(node)
	entry = filepath.Clean(entry)
	rust = filepath.Clean(rust)
	root := t.TempDir()
	library := "/* \U0001F600 */ export function Greeting() { return \"hi\"; }\nexport const abcdef = Greeting();\n"
	fixtures := map[string]string{
		"package.json":  "{\"name\":\"tsfixture\",\"version\":\"1.0.0\",\"private\":true}\n",
		"tsconfig.json": "{\"compilerOptions\":{\"target\":\"ES2020\",\"module\":\"commonjs\",\"strict\":true,\"skipLibCheck\":true,\"noLib\":true},\"include\":[\"library.ts\"]}\n",
		"library.ts":    library,
	}
	for name, content := range fixtures {
		if strings.Contains(content, "\r") {
			t.Fatal("fixture must use LF line endings")
		}
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "package.json", "tsconfig.json", "library.ts"}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "fixture"}} {
		if out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatal(err, string(out))
		}
	}
	run := func(args ...string) []byte {
		t.Helper()
		var out bytes.Buffer
		if err := Execute(context.Background(), args, root, &out); err != nil {
			t.Fatal(args, err)
		}
		return out.Bytes()
	}
	write := func(name string, value any) {
		t.Helper()
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	decode := func(data []byte, value any) {
		t.Helper()
		if err := json.Unmarshal(data, value); err != nil {
			t.Fatal(err)
		}
	}
	run("init")
	var s control.Snapshot
	decode(run("plan", "Index TypeScript module"), &s)
	run("approve", s.RunID, s.PlanID, "operator")
	decode(run("run", s.RunID), &s)
	index := filepath.Join(root, "index.scip")
	write("check.json", config.Check{Name: "scip-typescript", Argv: []string{node, entry, "index", "--output", index, "--no-progress-bar"}, TimeoutSeconds: 120})
	var preview producerPreview
	decode(run("ri", "prepare-producer", s.RunID, "check.json", "index.scip"), &preview)
	write("producer.json", preview)
	decode(run("ri", "produce", s.RunID, "producer.json", preview.IntentID, "operator"), &s)
	if s.RIProducer == nil || s.RIProducer.Outcome != "CONFIRMED" {
		t.Fatal("CLI producer not confirmed")
	}
	binary, err := os.ReadFile(rust)
	if err != nil {
		t.Fatal(err)
	}
	source, err := ri.FromRepository(s.Creation.Repository)
	if err != nil {
		t.Fatal(err)
	}
	digestSource := func(name string) repository.SourceDigest {
		t.Helper()
		digest, err := repository.DigestSource(context.Background(), s.Creation.Repository, name)
		if err != nil {
			t.Fatal(err)
		}
		return digest
	}
	digest := digestSource("library.ts")
	packageDigest := digestSource("package.json")
	tsconfigDigest := digestSource("tsconfig.json")
	entryBytes, err := os.ReadFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	// The admitted project-root URI must use Node's url.pathToFileURL
	// formatting, including Windows drive-letter handling.
	pathURL, err := exec.Command(node, "-e", "process.stdout.write(require('url').pathToFileURL(process.argv[1]).href)", s.Workspace.Request.Path).Output()
	if err != nil {
		t.Fatal(err)
	}
	projectRoot := string(pathURL)
	if !strings.HasPrefix(projectRoot, "file:") {
		t.Fatal("Node file URL missing", projectRoot)
	}
	policy := fmt.Sprintf("%x", sha256.Sum256([]byte("engorch.scip-typescript.0.4.0.positions.v2:utf16-code-units+omit-invalid-synthetic-file-enclosing")))
	plan := ri.ImportPlan{Version: 1, Executable: rust, ExecutableSHA256: fmt.Sprintf("%x", sha256.Sum256(binary)), Repository: s.Creation.Repository, OutputPath: filepath.Join(root, "snapshot.staging"), Request: ri.ImportRequest{Source: source, Producer: "scip-typescript", ProjectRoot: projectRoot, Policy: "scip_typescript040", Sources: map[string]string{"library.ts": filepath.Join(s.Workspace.Request.Path, "library.ts")}, Manifest: ri.Manifest{Format: 1, Source: source, Producers: []ri.Producer{{ID: "scip-typescript", Name: "scip-typescript", Version: "0.4.0", Inputs: []ri.Input{{Name: "source:library.ts", SHA256: digest.SHA256}, {Name: "config:package.json", SHA256: packageDigest.SHA256}, {Name: "config:tsconfig.json", SHA256: tsconfigDigest.SHA256}, {Name: "entry:scip-typescript", SHA256: fmt.Sprintf("%x", sha256.Sum256(entryBytes))}, {Name: "scip:position-policy", SHA256: policy}}}}}}}
	write("proposal.json", plan)
	decode(run("ri", "bind-import", s.RunID, "proposal.json"), &plan)
	if plan.ProducerIntentID == "" {
		t.Fatal("producer provenance missing")
	}
	write("bound.json", plan)
	var approval struct {
		IntentID string `json:"intent_id"`
	}
	decode(run("ri", "prepare-import", s.RunID, "bound.json"), &approval)
	decode(run("ri", "import", s.RunID, "bound.json", approval.IntentID, "operator"), &s)
	if s.RIImport == nil || s.RIImport.Outcome != "CONFIRMED" {
		t.Fatal("CLI real index import not confirmed")
	}
	if err := os.Mkdir(filepath.Join(root, "store"), 0700); err != nil {
		t.Fatal(err)
	}
	decode(run("ri", "prepare-publish", s.RunID, "store"), &approval)
	decode(run("ri", "publish", s.RunID, "store", approval.IntentID, "operator"), &s)
	if s.RIPublish == nil || s.RIPublish.Outcome != "CONFIRMED" {
		t.Fatal("real index publication failed")
	}
	artifact := s.RIPublish.Observation.Artifact
	var selected struct {
		Snapshot         ri.SnapshotRef `json:"snapshot"`
		Executable       string         `json:"executable"`
		ExecutableSHA256 string         `json:"executable_sha256"`
	}
	decode(run("ri", "runtime-binding", s.RunID), &selected)
	if selected.Snapshot != *artifact || selected.Executable != rust || selected.ExecutableSHA256 != plan.ExecutableSHA256 {
		t.Fatal("CLI runtime binding differs from admitted publication")
	}
	contents, err := os.ReadFile(filepath.Join(s.Workspace.Request.Path, "library.ts"))
	if err != nil {
		t.Fatal(err)
	}
	offset := strings.Index(string(contents), "Greeting")
	if offset < 0 {
		t.Fatal("fixture spelling absent")
	}
	var located ri.OccurrencePage
	decode(run("ri", "locate", rust, plan.ExecutableSHA256, artifact.Path, artifact.ID, "library.ts", strconv.Itoa(offset), "scip-typescript", "16"), &located)
	if len(located.Occurrences) != 1 || located.Occurrences[0].Symbol == nil {
		t.Fatal("published real symbol not located")
	}
	if located.AbsenceProven {
		t.Fatal("locate claimed absence")
	}
	var definitions ri.OccurrencePage
	decode(run("ri", "definition", rust, plan.ExecutableSHA256, artifact.Path, artifact.ID, *located.Occurrences[0].Symbol, "scip-typescript", "16"), &definitions)
	if len(definitions.Occurrences) != 1 || definitions.Occurrences[0].Spelling != "Greeting" || definitions.Occurrences[0].SourceSHA256 != digest.SHA256 || definitions.AbsenceProven {
		t.Fatal("published definition provenance mismatch")
	}
	var references ri.OccurrencePage
	decode(run("ri", "references", rust, plan.ExecutableSHA256, artifact.Path, artifact.ID, *located.Occurrences[0].Symbol, "scip-typescript", "16"), &references)
	if len(references.Occurrences) < 1 || references.AbsenceProven {
		t.Fatal("published references missing or absence claimed")
	}
	requireSource := func(records []ri.Occurrence, what string) {
		t.Helper()
		for _, record := range records {
			if record.SourceSHA256 != digest.SHA256 {
				t.Fatal(what, "source hash mismatch")
			}
		}
	}
	requireSource(references.Occurrences, "reference")
	write("semantic-query.json", map[string]any{"vocabulary": "references", "symbol": *located.Occurrences[0].Symbol, "producer": "scip-typescript", "limit": 16})
	var semantic ri.SemanticSnapshotResult
	decode(run("ri", "semantic", rust, plan.ExecutableSHA256, artifact.Path, artifact.ID, "semantic-query.json"), &semantic)
	if semantic.Coverage != "PARTIAL" || semantic.AbsenceProven || len(semantic.Occurrences) < 1 {
		t.Fatal("semantic reference coverage mismatch")
	}
	requireSource(semantic.Occurrences, "semantic")
}
