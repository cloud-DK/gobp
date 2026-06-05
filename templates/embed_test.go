package templates

import (
	"testing"
)

func TestGetCategories_returnsAllDirs(t *testing.T) {
	cats, err := GetCategories()
	if err != nil {
		t.Fatalf("GetCategories: %v", err)
	}
	if len(cats) == 0 {
		t.Error("expected at least one category")
	}
	// _shared is a raw filesystem directory and IS returned; callers are
	// responsible for filtering it (e.g. tui.filterCategories).
	found := false
	for _, c := range cats {
		if c == "_shared" {
			found = true
		}
	}
	if !found {
		t.Error("expected _shared to be present in raw GetCategories output")
	}
}

func TestGetCategories_containsExpected(t *testing.T) {
	cats, err := GetCategories()
	if err != nil {
		t.Fatalf("GetCategories: %v", err)
	}
	want := map[string]bool{"server": false, "cli": false, "database": false, "ui": false}
	for _, c := range cats {
		want[c] = true
	}
	for name, found := range want {
		if !found {
			t.Errorf("expected category %q to be present", name)
		}
	}
}

func TestGetMeta_knownOption(t *testing.T) {
	m, err := GetMeta("server", "stdlib")
	if err != nil {
		t.Fatalf("GetMeta: %v", err)
	}
	if m.Name == "" {
		t.Error("expected non-empty Name")
	}
	if m.Description == "" {
		t.Error("expected non-empty Description")
	}
}

func TestGetMeta_missingFile_returnsDefault(t *testing.T) {
	// A path that does not exist should return a default, not an error.
	m, err := GetMeta("server", "nonexistent_option_xyz")
	if err != nil {
		t.Fatalf("expected no error for missing meta.json, got: %v", err)
	}
	if m == nil {
		t.Fatal("expected non-nil Meta for missing meta.json")
	}
	if m.Name != "nonexistent_option_xyz" {
		t.Errorf("expected fallback Name %q, got %q", "nonexistent_option_xyz", m.Name)
	}
}

func TestGetMeta_isCmdSet(t *testing.T) {
	for _, opt := range []string{"react", "svelte", "vue"} {
		m, err := GetMeta("ui", opt)
		if err != nil {
			t.Fatalf("GetMeta(ui, %s): %v", opt, err)
		}
		if !m.IsCmd {
			t.Errorf("ui/%s: expected IsCmd=true", opt)
		}
	}
}

func TestGetMeta_isCmdFalseForGoOptions(t *testing.T) {
	for _, tc := range []struct{ cat, opt string }{
		{"server", "stdlib"},
		{"server", "gin"},
		{"cli", "cobra"},
	} {
		m, err := GetMeta(tc.cat, tc.opt)
		if err != nil {
			t.Fatalf("GetMeta(%s, %s): %v", tc.cat, tc.opt, err)
		}
		if m.IsCmd {
			t.Errorf("%s/%s: expected IsCmd=false", tc.cat, tc.opt)
		}
	}
}

func TestHasVariants_databaseHasDialects(t *testing.T) {
	for _, opt := range []string{"gorm", "sqlx", "bun", "xorm"} {
		has, err := HasVariants("database", opt)
		if err != nil {
			t.Fatalf("HasVariants(database, %s): %v", opt, err)
		}
		if !has {
			t.Errorf("database/%s: expected HasVariants=true", opt)
		}
	}
}

func TestHasVariants_serverHasNoDialects(t *testing.T) {
	for _, opt := range []string{"stdlib", "gin", "chi", "echo", "fiber"} {
		has, err := HasVariants("server", opt)
		if err != nil {
			t.Fatalf("HasVariants(server, %s): %v", opt, err)
		}
		if has {
			t.Errorf("server/%s: expected HasVariants=false", opt)
		}
	}
}

func TestGetVariants_gorm(t *testing.T) {
	dialects, err := GetVariants("database", "gorm")
	if err != nil {
		t.Fatalf("GetVariants: %v", err)
	}
	if len(dialects) == 0 {
		t.Fatal("expected at least one dialect for gorm")
	}
	names := map[string]bool{}
	for _, d := range dialects {
		names[d.Name] = true
		if d.Description == "" {
			t.Errorf("dialect %q has empty Description", d.Name)
		}
	}
	for _, want := range []string{"postgres", "mysql", "sqlite"} {
		if !names[want] {
			t.Errorf("expected dialect %q in gorm variants", want)
		}
	}
}
