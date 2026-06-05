package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

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

func Generate(cfg Config) error {
	optionPath := cfg.Category + "/" + cfg.Option

	raw, err := templates.TemplateFS.ReadFile(optionPath + "/cmd.json")
	if err == nil {
		return runCmd(raw, cfg)
	}

	data := templateData{
		ModuleName:  cfg.ModuleName,
		ProjectName: projectName(cfg.ModuleName, cfg.OutputDir),
		Category:    cfg.Category,
		Framework:   cfg.Option,
		Dialect:     cfg.Variant,
	}

	if err := writeDir("_shared", cfg.OutputDir, data); err != nil {
		return fmt.Errorf("shared templates: %w", err)
	}
	if err := writeDir(optionPath, cfg.OutputDir, data); err != nil {
		return fmt.Errorf("option templates: %w", err)
	}
	return goModInit(cfg.ModuleName, cfg.OutputDir)
}

func goModInit(moduleName, dir string) error {
	cmd := exec.Command("go", "mod", "init", moduleName)
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

		rel, _ := filepath.Rel(srcDir, path)
		rel = filepath.ToSlash(rel)

		outName := strings.TrimSuffix(rel, ".tmpl")
		if outName == "gitignore" {
			outName = ".gitignore"
		}

		outPath := filepath.Join(outBase, filepath.FromSlash(outName))
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
			if formatted, err := format.Source(out); err == nil {
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
	name := projectName(cfg.ModuleName, cfg.OutputDir)
	args := make([]string, len(cf.Cmd))
	for i, a := range cf.Cmd {
		args[i] = strings.ReplaceAll(a, "${projectName}", name)
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = cfg.OutputDir
	return cmd.Run()
}

func projectName(moduleName, outputDir string) string {
	if moduleName != "" {
		parts := strings.Split(moduleName, "/")
		if last := parts[len(parts)-1]; last != "" {
			return last
		}
	}
	return filepath.Base(outputDir)
}
