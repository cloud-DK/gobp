package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectName(t *testing.T) {
	cases := []struct {
		module string
		outDir string
		want   string
	}{
		{"github.com/user/myapp", "/any/dir", "myapp"},
		{"myapp", "/any/dir", "myapp"},
		{"", "/some/dir", "dir"},
		{"github.com/user/", "/any/dir", "dir"}, // trailing slash → falls back to outDir base
	}
	for _, tc := range cases {
		got := ProjectName(tc.module, tc.outDir)
		if got != tc.want {
			t.Errorf("ProjectName(%q, %q) = %q, want %q", tc.module, tc.outDir, got, tc.want)
		}
	}
}

func TestGenerate_server(t *testing.T) {
	dir := t.TempDir()
	moduleName := "github.com/test/myserver"
	if err := WriteShared(moduleName, dir); err != nil {
		t.Fatalf("WriteShared: %v", err)
	}
	cfg := Config{
		ModuleName: moduleName,
		Category:   "server",
		Option:     "stdlib",
		OutputDir:  dir,
	}
	if err := Generate(cfg); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	assertFileExists(t, dir, "cmd/main.go")
	assertFileExists(t, dir, "internal/server/server.go")
	assertFileExists(t, dir, ".gitignore")
}

func TestGenerate_database_with_variant(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		ModuleName: "github.com/test/myapp",
		Category:   "database",
		Option:     "gorm",
		Variant:    "postgres",
		OutputDir:  dir,
	}
	if err := Generate(cfg); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	assertFileExists(t, dir, "internal/database/db.go")
	assertFileContains(t, dir, "internal/database/db.go", "postgres")
}

func TestGenerate_database_variant_mysql(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		ModuleName: "github.com/test/myapp",
		Category:   "database",
		Option:     "gorm",
		Variant:    "mysql",
		OutputDir:  dir,
	}
	if err := Generate(cfg); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	assertFileContains(t, dir, "internal/database/db.go", "mysql")
}

func TestGenerate_cli(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		ModuleName: "github.com/test/mytool",
		Category:   "cli",
		Option:     "cobra",
		OutputDir:  dir,
	}
	if err := Generate(cfg); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	assertFileExists(t, dir, "main.go")
	assertFileExists(t, dir, "cmd/root.go")
}

func TestGenerate_skips_meta_json(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		ModuleName: "github.com/test/myserver",
		Category:   "server",
		Option:     "chi",
		OutputDir:  dir,
	}
	if err := Generate(cfg); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	metaPath := filepath.Join(dir, "meta.json")
	if _, err := os.Stat(metaPath); err == nil {
		t.Error("meta.json should not be written to the output directory")
	}
}

func TestGenerate_module_name_in_output(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		ModuleName: "github.com/acme/webapp",
		Category:   "server",
		Option:     "gin",
		OutputDir:  dir,
	}
	if err := Generate(cfg); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	assertFileContains(t, dir, "cmd/main.go", "github.com/acme/webapp")
}

// helpers

func assertFileExists(t *testing.T, base, rel string) {
	t.Helper()
	path := filepath.Join(base, filepath.FromSlash(rel))
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file %s to exist: %v", rel, err)
	}
}

func assertFileContains(t *testing.T, base, rel, substr string) {
	t.Helper()
	path := filepath.Join(base, filepath.FromSlash(rel))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	if !strings.Contains(string(data), substr) {
		t.Errorf("file %s does not contain %q", rel, substr)
	}
}
