package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PackAdminResourceTable re-runs the column packing heuristics against an
// existing generated TS resource file.
func PackAdminResourceTable(root string, resourceName string) error {
	def, err := DefinitionFromModel(resourceName)
	if err != nil {
		return fmt.Errorf("loading model definition: %w", err)
	}

	plural := Pluralize(toSnakeCase(def.Name))
	kebab := strings.ReplaceAll(plural, "_", "-")
	path := filepath.Join(root, "apps", "admin", "resources", kebab+".ts")

	data, statErr := os.ReadFile(path)
	if os.IsNotExist(statErr) {
		return fmt.Errorf("admin resource file not found: %s", path)
	}
	if statErr != nil {
		return fmt.Errorf("reading %s: %w", path, statErr)
	}

	content := string(data)
	packs, usesStackedCell := detectColumnPacks(def.Fields)
	if !usesStackedCell {
		return nil // nothing to pack
	}

	updated := content
	appliedPacks := make(map[string]bool)

	// Keep track of which absorbed fields to remove
	absorbedFields := make(map[string]bool)
	for key, pack := range packs {
		if key != pack.primary {
			absorbedFields[key] = true
		}
	}

	// We'll iterate lines to safely rewrite the array elements
	lines := strings.Split(updated, "\n")
	var newLines []string
	changed := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// 1. Identify and skip absorbed fields (e.g. email, currency)
		skipAbsorbed := false
		for absorbedKey := range absorbedFields {
			if strings.Contains(trimmed, fmt.Sprintf(`key: "%s"`, absorbedKey)) {
				skipAbsorbed = true
				changed = true
				break
			}
		}
		if skipAbsorbed {
			continue
		}

		// 2. Identify the primary field and replace it with the pack line
		replacedPrimary := false
		for primaryKey, pack := range packs {
			if primaryKey == pack.primary && !appliedPacks[primaryKey] {
				if strings.Contains(trimmed, fmt.Sprintf(`key: "%s"`, primaryKey)) {
					// Grab the indent from the original line to match formatting
					indent := line[:strings.Index(line, trimmed)]
					newLines = append(newLines, indent+pack.line)
					appliedPacks[primaryKey] = true
					replacedPrimary = true
					changed = true
					break
				}
			}
		}

		if !replacedPrimary {
			newLines = append(newLines, line)
		}
	}

	if !changed {
		return nil
	}

	updated = strings.Join(newLines, "\n")

	// Inject StackedCell import if missing
	if !strings.Contains(updated, "StackedCell") {
		importLine := `import { StackedCell } from "@/components/tables/stacked-cell"`
		if !strings.Contains(updated, importLine) {
			// Find the last import
			lastImportIdx := strings.LastIndex(updated, "import ")
			if lastImportIdx != -1 {
				eol := strings.Index(updated[lastImportIdx:], "\n")
				if eol != -1 {
					insertPos := lastImportIdx + eol + 1
					updated = updated[:insertPos] + importLine + "\n" + updated[insertPos:]
				}
			}
		}
	}

	if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	fmt.Printf("  ✓ Packed columns for %s in %s\n", def.Name, kebab+".ts")
	return nil
}
