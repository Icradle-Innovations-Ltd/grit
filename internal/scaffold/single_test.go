package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteSingleMainGo(t *testing.T) {
	root := t.TempDir()
	opts := Options{ProjectName: "test-single"}

	// create mock cmd/server
	if err := os.MkdirAll(filepath.Join(root, "cmd", "server"), 0755); err != nil {
		t.Fatalf("mkdirall: %v", err)
	}
	// mock file that should be deleted
	cmdServerMain := filepath.Join(root, "cmd", "server", "main.go")
	if err := writeFile(cmdServerMain, "package main"); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
	
	// Create frontend/dist to test placeholder
	if err := os.MkdirAll(filepath.Join(root, "frontend", "dist"), 0755); err != nil {
		t.Fatalf("mkdirall: %v", err)
	}

	if err := writeSingleMainGo(root, opts); err != nil {
		t.Fatalf("writeSingleMainGo: %v", err)
	}

	// Verify main.go is at root
	mainPath := filepath.Join(root, "main.go")
	if _, err := os.Stat(mainPath); err != nil {
		t.Errorf("main.go was not created at root: %v", err)
	}
	
	// Verify cmd/server/main.go is deleted
	if _, err := os.Stat(cmdServerMain); !os.IsNotExist(err) {
		t.Errorf("cmd/server/main.go was not deleted")
	}

	// Verify frontend/dist/index.html is created
	indexPath := filepath.Join(root, "frontend", "dist", "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		t.Errorf("frontend/dist/index.html was not created: %v", err)
	}
}

func TestWriteSingleFrontendFiles(t *testing.T) {
	root := t.TempDir()
	opts := Options{ProjectName: "test-single"}
	
	if err := os.MkdirAll(filepath.Join(root, "frontend", "src", "routes", "blog"), 0755); err != nil {
		t.Fatalf("mkdirall: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "frontend", "src", "components"), 0755); err != nil {
		t.Fatalf("mkdirall: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "frontend", "src", "lib"), 0755); err != nil {
		t.Fatalf("mkdirall: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "frontend", "src", "hooks"), 0755); err != nil {
		t.Fatalf("mkdirall: %v", err)
	}

	if err := writeSingleFrontendFiles(root, opts); err != nil {
		t.Fatalf("writeSingleFrontendFiles: %v", err)
	}

	expected := []string{
		filepath.Join("frontend", "package.json"),
		filepath.Join("frontend", "vite.config.ts"),
		filepath.Join("frontend", "tailwind.config.ts"),
		filepath.Join("frontend", "postcss.config.cjs"),
		filepath.Join("frontend", "tsconfig.json"),
		filepath.Join("frontend", "index.html"),
		filepath.Join("frontend", "src", "main.tsx"),
		filepath.Join("frontend", "src", "vite-env.d.ts"),
		filepath.Join("frontend", "src", "globals.css"),
		filepath.Join("frontend", "src", "routes", "__root.tsx"),
		filepath.Join("frontend", "src", "components", "navbar.tsx"),
		filepath.Join("frontend", "src", "components", "footer.tsx"),
		filepath.Join("frontend", "src", "lib", "api.ts"),
	}

	for _, f := range expected {
		if _, err := os.Stat(filepath.Join(root, f)); err != nil {
			t.Errorf("expected file %s was not created", f)
		}
	}
}

func TestWriteSingleRootFiles(t *testing.T) {
	root := t.TempDir()
	opts := Options{ProjectName: "test-single"}

	if err := writeSingleRootFiles(root, opts); err != nil {
		t.Fatalf("writeSingleRootFiles: %v", err)
	}

	expected := []string{
		"Makefile",
		".gitignore",
	}

	for _, f := range expected {
		if _, err := os.Stat(filepath.Join(root, f)); err != nil {
			t.Errorf("expected file %s was not created", f)
		}
	}
	
	makefileContent, err := os.ReadFile(filepath.Join(root, "Makefile"))
	if err != nil {
		t.Fatalf("readFile: %v", err)
	}
	if !strings.Contains(string(makefileContent), "test-single") {
		t.Errorf("Makefile doesn't contain project name")
	}
}
