package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/cloud-dk/gobp/templates"
)

type Config struct {
	ModuleName string
	Category   string
	Option     string
	Variant    string // dialect or sub-option, empty if none
	OutputDir  string
}

type templateData struct {
	ModuleName  string
	ProjectName string
	Category    string
	Framework   string
	Dialect     string
}

type cmdFile struct {
	Cmd []string `json:"cmd"`
}

// WriteShared writes the _shared template directory into outDir. Call this
// once before calling Generate for the first non-UI selection.
func WriteShared(moduleName, outDir string) error {
	data := templateData{
		ModuleName:  moduleName,
		ProjectName: ProjectName(moduleName, outDir),
	}
	return writeDir("_shared", outDir, data)
}

func Generate(cfg Config) error {
	optionPath := cfg.Category + "/" + cfg.Option

	raw, err := templates.TemplateFS.ReadFile(optionPath + "/cmd.json")
	if err == nil {
		return runCmd(raw, cfg)
	}

	data := templateData{
		ModuleName:  cfg.ModuleName,
		ProjectName: ProjectName(cfg.ModuleName, cfg.OutputDir),
		Category:    cfg.Category,
		Framework:   cfg.Option,
		Dialect:     cfg.Variant,
	}

	return writeDir(optionPath, cfg.OutputDir, data)
}

func GoModInit(moduleName, dir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "mod", "init", moduleName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	return cmd.Run()
}

func writeDir(srcDir, outBase string, data templateData) error {
	return fs.WalkDir(templates.TemplateFS, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		base := filepath.Base(path)
		if base == "meta.json" {
			return nil
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("rel path %s: %w", path, err)
		}
		rel = filepath.ToSlash(rel)

		outName := strings.TrimSuffix(rel, ".tmpl")
		// embed.FS silently drops files whose name starts with '.', so template
		// authors name the file "gitignore" and we rename it on the way out.
		if outName == "gitignore" {
			outName = ".gitignore"
		}

		outPath := filepath.Join(outBase, filepath.FromSlash(outName))

		// Protect hand-edited Go source files from silent overwrite.
		// Check before MkdirAll so a conflict never leaves an empty directory on disk.
		if strings.HasSuffix(outName, ".go") {
			if _, statErr := os.Stat(outPath); statErr == nil {
				return fmt.Errorf("%s already exists; run gobp in an empty directory or remove the file first", outName)
			}
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}

		raw, err := templates.TemplateFS.ReadFile(path)
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".tmpl") {
			return os.WriteFile(outPath, raw, 0644)
		}

		tmpl, err := template.New(path).Parse(string(raw))
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("execute %s: %w", path, err)
		}
		out := buf.Bytes()
		if strings.HasSuffix(outName, ".go") {
			formatted, fmtErr := format.Source(out)
			if fmtErr != nil {
				fmt.Fprintf(os.Stderr, "warning: could not format %s: %v\n", outPath, fmtErr)
			} else {
				out = formatted
			}
		}
		return os.WriteFile(outPath, out, 0644)
	})
}

func runCmd(raw []byte, cfg Config) error {
	var cf cmdFile
	if err := json.Unmarshal(raw, &cf); err != nil {
		return fmt.Errorf("parse cmd.json: %w", err)
	}
	if len(cf.Cmd) == 0 {
		return fmt.Errorf("cmd.json has an empty cmd array")
	}
	name := ProjectName(cfg.ModuleName, cfg.OutputDir)
	args := make([]string, len(cf.Cmd))
	for i, a := range cf.Cmd {
		args[i] = strings.ReplaceAll(a, "${projectName}", name)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = cfg.OutputDir
	return cmd.Run()
}

func ProjectName(moduleName, outputDir string) string {
	if moduleName != "" {
		parts := strings.Split(moduleName, "/")
		if last := parts[len(parts)-1]; last != "" {
			return last
		}
	}
	return filepath.Base(outputDir)
}
