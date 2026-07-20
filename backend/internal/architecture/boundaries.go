package architecture

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const modulePath = "github.com/gabrielnakaema/project-chat"

var capabilityRoots = map[string]struct{}{
	"apikey":       {},
	"chat":         {},
	"mcp":          {},
	"notification": {},
	"project":      {},
	"realtime":     {},
	"tasks":        {},
	"user":         {},
}

type Violation struct {
	File    string
	Message string
}

func (v Violation) Error() string {
	return fmt.Sprintf("%s: %s", v.File, v.Message)
}

// CheckImportBoundaries validates the Phase 1 dependency rules using direct
// imports. Transitive violations are still caught at the package that declares
// the forbidden edge.
func CheckImportBoundaries(root string) ([]Violation, error) {
	type packageImports struct {
		files        []string
		capabilities map[string]struct{}
	}

	packages := map[string]*packageImports{}
	var violations []Violation
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "vendor" || entry.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return fmt.Errorf("parse %s: %w", rel, err)
		}

		packageDir := filepath.ToSlash(filepath.Dir(rel))
		imports := packages[packageDir]
		if imports == nil {
			imports = &packageImports{capabilities: map[string]struct{}{}}
			packages[packageDir] = imports
		}
		imports.files = append(imports.files, rel)

		sourceCapability := capabilityForRelativePath(rel)
		for _, spec := range file.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return fmt.Errorf("unquote import in %s: %w", rel, err)
			}
			importedCapability := capabilityForImport(importPath)
			if importedCapability == "" {
				continue
			}
			imports.capabilities[importedCapability] = struct{}{}

			if strings.HasPrefix(rel, "internal/platform/") {
				violations = append(violations, Violation{File: rel, Message: fmt.Sprintf("platform must not import capability %q", importedCapability)})
			}
			if sourceCapability != "" && sourceCapability != importedCapability && !isCompositionRoot(packageDir) {
				violations = append(violations, Violation{File: rel, Message: fmt.Sprintf("capability %q must not import capability %q", sourceCapability, importedCapability)})
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	for packageDir, imports := range packages {
		if len(imports.capabilities) < 2 || isCompositionRoot(packageDir) {
			continue
		}
		capabilities := make([]string, 0, len(imports.capabilities))
		for capability := range imports.capabilities {
			capabilities = append(capabilities, capability)
		}
		sort.Strings(capabilities)
		sort.Strings(imports.files)
		violations = append(violations, Violation{
			File:    imports.files[0],
			Message: fmt.Sprintf("only composition roots may wire multiple capabilities (%s)", strings.Join(capabilities, ", ")),
		})
	}

	sort.Slice(violations, func(i, j int) bool {
		if violations[i].File == violations[j].File {
			return violations[i].Message < violations[j].Message
		}
		return violations[i].File < violations[j].File
	})
	return violations, nil
}

func capabilityForImport(importPath string) string {
	prefix := modulePath + "/internal/"
	if !strings.HasPrefix(importPath, prefix) {
		return ""
	}
	remainder := strings.TrimPrefix(importPath, prefix)
	root := strings.SplitN(remainder, "/", 2)[0]
	if _, ok := capabilityRoots[root]; ok {
		return root
	}
	return ""
}

func capabilityForRelativePath(path string) string {
	const prefix = "internal/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	remainder := strings.TrimPrefix(path, prefix)
	root := strings.SplitN(remainder, "/", 2)[0]
	if _, ok := capabilityRoots[root]; ok {
		return root
	}
	return ""
}

func isCompositionRoot(packageDir string) bool {
	if strings.HasPrefix(packageDir, "cmd/") {
		return true
	}
	if strings.Contains(packageDir, "/tests/") || strings.HasSuffix(packageDir, "/tests") {
		return true
	}
	parts := strings.Split(packageDir, "/")
	return len(parts) == 3 && parts[0] == "internal" && parts[2] == "app"
}
