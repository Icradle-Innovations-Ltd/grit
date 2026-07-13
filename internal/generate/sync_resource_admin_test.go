package generate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeAdminResourceFile creates the full fake project tree and writes a
// resource file at apps/admin/resources/<kebab>.ts, returning the project root.
func makeAdminResourceFile(t *testing.T, kebabPlural, content string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "apps", "admin", "resources")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("makeAdminResourceFile: mkdir: %v", err)
	}
	path := filepath.Join(dir, kebabPlural+".ts")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("makeAdminResourceFile: write: %v", err)
	}
	return root
}

// readAdminResourceFile reads back the resource file after SyncAdminResource runs.
func readAdminResourceFile(t *testing.T, root, kebabPlural string) string {
	t.Helper()
	path := filepath.Join(root, "apps", "admin", "resources", kebabPlural+".ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readAdminResourceFile: %v", err)
	}
	return string(data)
}

// minimalResourceFile builds a realistic resource file with the v3.31.16 markers.
// It pre-populates existingCols and existingFields inside the auto fences so
// the test can check the diff logic.
func minimalResourceFile(existingCols, existingFields string) string {
	return `import { defineResource } from "@/lib/resource";

export const postResource = defineResource({
  name: "Post",
  slug: "posts",
  endpoint: "/api/posts",
  icon: "FileText",
  label: { singular: "Post", plural: "Posts" },
  table: {
    columns: [
      // grit:cols:auto-start
` + existingCols + `      // grit:cols:auto-end
    ],
  },
  form: {
    fields: [
      // grit:fields:auto-start
` + existingFields + `      // grit:fields:auto-end
    ],
  },
});
`
}

// ── SyncAdminResource: no resource file ───────────────────────────────────────

func TestSyncAdminResource_NoFile(t *testing.T) {
	root := t.TempDir()
	// No apps/admin/resources directory at all.
	s := GoStruct{
		Name: "Post",
		Fields: []GoField{
			{Name: "Title", GoType: "string", JSONName: "title"},
		},
	}
	added, warnings, err := SyncAdminResource(root, s)
	if err != nil {
		t.Fatalf("expected no error when file is missing, got: %v", err)
	}
	if added != 0 {
		t.Errorf("expected 0 added, got %d", added)
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings, got: %v", warnings)
	}
}

// ── SyncAdminResource: missing markers (pre-v3.31.16 file) ───────────────────

func TestSyncAdminResource_MissingMarkers(t *testing.T) {
	// A resource file without the grit:cols:auto-* / grit:fields:auto-* markers.
	content := `import { defineResource } from "@/lib/resource";

export const postResource = defineResource({
  name: "Post",
  table: {
    columns: [
      { key: "title", label: "Title" },
    ],
  },
  form: {
    fields: [
      { key: "title", label: "Title", type: "text" },
    ],
  },
});
`
	root := makeAdminResourceFile(t, "posts", content)
	s := GoStruct{
		Name: "Post",
		Fields: []GoField{
			{Name: "Title", GoType: "string", JSONName: "title"},
			{Name: "Body", GoType: "string", JSONName: "body"},
		},
	}
	added, warnings, err := SyncAdminResource(root, s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if added != 0 {
		t.Errorf("expected 0 added (no markers), got %d", added)
	}
	if len(warnings) == 0 {
		t.Error("expected a warning about missing markers, got none")
	}
	if !strings.Contains(warnings[0], "auto-add skipped") {
		t.Errorf("warning should mention 'auto-add skipped', got: %q", warnings[0])
	}
	// File must be unchanged.
	got := readAdminResourceFile(t, root, "posts")
	if got != content {
		t.Errorf("file should be unchanged when markers are missing:\ngot:\n%s", got)
	}
}

// ── SyncAdminResource: new fields added ──────────────────────────────────────

func TestSyncAdminResource_AddsNewFields(t *testing.T) {
	// Resource file has only "title" inside the fences; "slug" and "published" are new.
	existing := `      { key: "title", label: "Title", sortable: true, searchable: true },
`
	root := makeAdminResourceFile(t, "posts", minimalResourceFile(existing, existing))
	s := GoStruct{
		Name: "Post",
		Fields: []GoField{
			{Name: "Title", GoType: "string", JSONName: "title"},
			{Name: "Slug", GoType: "string", JSONName: "slug"},
			{Name: "Published", GoType: "bool", JSONName: "published"},
		},
	}
	added, warnings, err := SyncAdminResource(root, s)
	if err != nil {
		t.Fatalf("SyncAdminResource: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
	if added != 2 {
		t.Errorf("expected 2 fields added (slug + published), got %d", added)
	}

	got := readAdminResourceFile(t, root, "posts")

	// Both new keys must appear in the file.
	if !strings.Contains(got, `key: "slug"`) {
		t.Error("slug column not injected into file")
	}
	if !strings.Contains(got, `key: "published"`) {
		t.Error("published column not injected into file")
	}

	// The markers must still be present.
	if !strings.Contains(got, "// grit:cols:auto-end") {
		t.Error("grit:cols:auto-end marker was removed")
	}
	if !strings.Contains(got, "// grit:fields:auto-end") {
		t.Error("grit:fields:auto-end marker was removed")
	}

	// The original "title" entry must be preserved.
	if !strings.Contains(got, `key: "title"`) {
		t.Error("pre-existing title key was removed")
	}
}

// ── SyncAdminResource: idempotent (nothing new to add) ────────────────────────

func TestSyncAdminResource_Idempotent(t *testing.T) {
	existing := `      { key: "title", label: "Title", sortable: true, searchable: true },
      { key: "slug", label: "Slug", sortable: true, searchable: true },
`
	content := minimalResourceFile(existing, existing)
	root := makeAdminResourceFile(t, "posts", content)
	s := GoStruct{
		Name: "Post",
		Fields: []GoField{
			{Name: "Title", GoType: "string", JSONName: "title"},
			{Name: "Slug", GoType: "string", JSONName: "slug"},
		},
	}
	added, warnings, err := SyncAdminResource(root, s)
	if err != nil {
		t.Fatalf("SyncAdminResource: %v", err)
	}
	if added != 0 {
		t.Errorf("expected 0 added (all fields already exist), got %d", added)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
	// File must be byte-for-byte identical.
	got := readAdminResourceFile(t, root, "posts")
	if got != content {
		t.Errorf("file changed even though all fields were already present:\ngot:\n%s", got)
	}
}

// ── SyncAdminResource: auto fields skipped ───────────────────────────────────

func TestSyncAdminResource_AutoFieldsSkipped(t *testing.T) {
	root := makeAdminResourceFile(t, "posts", minimalResourceFile("", ""))
	s := GoStruct{
		Name: "Post",
		Fields: []GoField{
			// These are auto fields — should never be injected.
			{Name: "ID", GoType: "uint", JSONName: "id"},
			{Name: "CreatedAt", GoType: "time.Time", JSONName: "created_at"},
			{Name: "UpdatedAt", GoType: "time.Time", JSONName: "updated_at"},
			{Name: "DeletedAt", GoType: "gorm.DeletedAt", JSONName: "deleted_at"},
		},
	}
	added, _, err := SyncAdminResource(root, s)
	if err != nil {
		t.Fatalf("SyncAdminResource: %v", err)
	}
	if added != 0 {
		t.Errorf("expected 0 added (all auto fields), got %d", added)
	}
	got := readAdminResourceFile(t, root, "posts")
	for _, key := range []string{`key: "id"`, `key: "created_at"`, `key: "updated_at"`, `key: "deleted_at"`} {
		if strings.Contains(got, key) {
			t.Errorf("auto field %q should not be injected", key)
		}
	}
}

// ── SyncAdminResource: field already referenced outside fence ─────────────────

func TestSyncAdminResource_FieldOutsideFenceNotDuplicated(t *testing.T) {
	// "status" is defined outside the auto fence (a customised entry).
	// SyncAdminResource must detect the key: "status" string anywhere in the
	// file and skip re-injecting it.
	content := `import { defineResource } from "@/lib/resource";

export const orderResource = defineResource({
  name: "Order",
  table: {
    columns: [
      { key: "status", label: "Order Status", format: "badge" },
      // grit:cols:auto-start
      { key: "total_amount", label: "Total amount", format: "currency", currencyPrefix: "$", sortable: true },
      // grit:cols:auto-end
    ],
  },
  form: {
    fields: [
      // grit:fields:auto-start
      // grit:fields:auto-end
    ],
  },
});
`
	root := makeAdminResourceFile(t, "orders", content)
	s := GoStruct{
		Name: "Order",
		Fields: []GoField{
			{Name: "Status", GoType: "string", JSONName: "status"},
			{Name: "TotalAmount", GoType: "float64", JSONName: "total_amount"},
		},
	}
	added, _, err := SyncAdminResource(root, s)
	if err != nil {
		t.Fatalf("SyncAdminResource: %v", err)
	}
	if added != 0 {
		t.Errorf("expected 0 added (both fields already referenced), got %d", added)
	}
	got := readAdminResourceFile(t, root, "orders")
	// Only one occurrence of key: "status" — no duplicates.
	if strings.Count(got, `key: "status"`) != 1 {
		t.Errorf("key: \"status\" duplicated in file:\n%s", got)
	}
}

// ── SyncAdminResource: correct column attributes by Go type ──────────────────

func TestSyncAdminResource_ColumnAttributes(t *testing.T) {
	tests := []struct {
		name           string
		field          GoField
		wantColAttrs   []string // substrings expected in the injected column line
		wantFormAttrs  []string // substrings expected in the injected form line
	}{
		{
			name:  "bool field gets format:boolean + type:toggle",
			field: GoField{Name: "Published", GoType: "bool", JSONName: "published"},
			wantColAttrs:  []string{`key: "published"`, `format: "boolean"`},
			wantFormAttrs: []string{`key: "published"`, `type: "toggle"`},
		},
		{
			name:  "float price field gets format:currency + type:number + numberKind:float",
			field: GoField{Name: "Price", GoType: "float64", JSONName: "price"},
			wantColAttrs:  []string{`key: "price"`, `format: "currency"`},
			wantFormAttrs: []string{`key: "price"`, `type: "number"`, `numberKind: "float"`},
		},
		{
			name:  "string email field gets format:email + type:text",
			field: GoField{Name: "Email", GoType: "string", JSONName: "email"},
			wantColAttrs:  []string{`key: "email"`, `format: "email"`},
			wantFormAttrs: []string{`key: "email"`, `type: "text"`},
		},
		{
			name:  "*time.Time gets format:date + type:datetime",
			field: GoField{Name: "PublishedAt", GoType: "*time.Time", JSONName: "published_at"},
			wantColAttrs:  []string{`key: "published_at"`, `format: "relative"`},
			wantFormAttrs: []string{`key: "published_at"`, `type: "datetime"`},
		},
		{
			name:  "int field is sortable + type:number + numberKind:int",
			field: GoField{Name: "ViewCount", GoType: "int", JSONName: "view_count"},
			wantColAttrs:  []string{`key: "view_count"`, `sortable: true`},
			wantFormAttrs: []string{`key: "view_count"`, `type: "number"`, `numberKind: "int"`},
		},
		{
			name:  "description string uses type:textarea",
			field: GoField{Name: "Description", GoType: "string", JSONName: "description"},
			wantColAttrs:  []string{`key: "description"`},
			wantFormAttrs: []string{`key: "description"`, `type: "textarea"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plural := Pluralize(toSnakeCase(tt.field.Name))
			kebab := strings.ReplaceAll(plural, "_", "-")

			root := makeAdminResourceFile(t, kebab, minimalResourceFile("", ""))
			s := GoStruct{
				Name:   tt.field.Name,
				Fields: []GoField{tt.field},
			}
			added, _, err := SyncAdminResource(root, s)
			if err != nil {
				t.Fatalf("SyncAdminResource: %v", err)
			}
			if added != 1 {
				t.Fatalf("expected 1 field added, got %d", added)
			}
			got := readAdminResourceFile(t, root, kebab)
			for _, attr := range tt.wantColAttrs {
				if !strings.Contains(got, attr) {
					t.Errorf("column line missing %q in:\n%s", attr, got)
				}
			}
			for _, attr := range tt.wantFormAttrs {
				if !strings.Contains(got, attr) {
					t.Errorf("form field line missing %q in:\n%s", attr, got)
				}
			}
		})
	}
}

// ── SyncAdminResource: compound model name ────────────────────────────────────

func TestSyncAdminResource_CompoundModelName(t *testing.T) {
	// BlogPost → blog-posts (kebab plural)
	root := makeAdminResourceFile(t, "blog-posts", minimalResourceFile("", ""))
	s := GoStruct{
		Name: "BlogPost",
		Fields: []GoField{
			{Name: "Title", GoType: "string", JSONName: "title"},
		},
	}
	added, _, err := SyncAdminResource(root, s)
	if err != nil {
		t.Fatalf("SyncAdminResource: %v", err)
	}
	if added != 1 {
		t.Errorf("expected 1 field added, got %d", added)
	}
	got := readAdminResourceFile(t, root, "blog-posts")
	if !strings.Contains(got, `key: "title"`) {
		t.Error("title not injected into blog-posts.ts")
	}
}

// ── SyncAdminResource: Version field is always skipped ───────────────────────

func TestSyncAdminResource_VersionFieldSkipped(t *testing.T) {
	root := makeAdminResourceFile(t, "posts", minimalResourceFile("", ""))
	s := GoStruct{
		Name: "Post",
		Fields: []GoField{
			{Name: "Title", GoType: "string", JSONName: "title"},
			{Name: "Version", GoType: "int", JSONName: "version"},
		},
	}
	added, _, err := SyncAdminResource(root, s)
	if err != nil {
		t.Fatalf("SyncAdminResource: %v", err)
	}
	if added != 1 {
		t.Errorf("expected 1 added (title only; Version skipped), got %d", added)
	}
	got := readAdminResourceFile(t, root, "posts")
	if strings.Contains(got, `key: "version"`) {
		t.Error("Version field should be skipped but was injected")
	}
}

// ── insertBeforeMarker ────────────────────────────────────────────────────────

func TestInsertBeforeMarker(t *testing.T) {
	t.Run("injects line before marker preserving indent", func(t *testing.T) {
		content := "      // grit:cols:auto-end\n"
		got := insertBeforeMarker(content, "// grit:cols:auto-end", `{ key: "name" },`)
		if !strings.Contains(got, "      { key: \"name\" },\n      // grit:cols:auto-end") {
			t.Errorf("unexpected output:\n%s", got)
		}
		// Marker must still be present.
		if !strings.Contains(got, "// grit:cols:auto-end") {
			t.Error("marker was removed")
		}
	})

	t.Run("missing marker returns content unchanged", func(t *testing.T) {
		content := "no marker here\n"
		got := insertBeforeMarker(content, "// grit:cols:auto-end", "injected")
		if got != content {
			t.Errorf("content changed when marker was absent:\n%s", got)
		}
	})

	t.Run("multiple calls accumulate insertions in order", func(t *testing.T) {
		content := "      // grit:cols:auto-end\n"
		content = insertBeforeMarker(content, "// grit:cols:auto-end", `{ key: "a" },`)
		content = insertBeforeMarker(content, "// grit:cols:auto-end", `{ key: "b" },`)
		if !strings.Contains(content, `key: "a"`) || !strings.Contains(content, `key: "b"`) {
			t.Errorf("both insertions not found:\n%s", content)
		}
		// Marker must appear exactly once.
		if strings.Count(content, "// grit:cols:auto-end") != 1 {
			t.Errorf("marker duplicated:\n%s", content)
		}
	})
}

// ── humanLabel ────────────────────────────────────────────────────────────────

func TestHumanLabel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Title", "Title"},
		{"FirstName", "First name"},
		{"PhoneNumber", "Phone number"},
		{"email", "Email"},
		{"UserId", "User id"},
		{"TotalAmount", "Total amount"},
		{"PublishedAt", "Published at"},
	}
	for _, tt := range tests {
		got := humanLabel(tt.input)
		if got != tt.want {
			t.Errorf("humanLabel(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ── guessFormat ───────────────────────────────────────────────────────────────

func TestGuessFormat(t *testing.T) {
	tests := []struct {
		field GoField
		want  string
	}{
		{GoField{Name: "Published", GoType: "bool"}, "boolean"},
		{GoField{Name: "Price", GoType: "float64"}, "currency"},
		{GoField{Name: "Amount", GoType: "float64"}, "currency"},
		{GoField{Name: "Weight", GoType: "float64"}, ""},
		{GoField{Name: "CreatedAt", GoType: "*time.Time"}, "relative"},
		{GoField{Name: "PublishedAt", GoType: "*time.Time"}, "relative"},
		{GoField{Name: "StartDate", GoType: "*time.Time"}, "date"},
		{GoField{Name: "Avatar", GoType: "string"}, "image"},
		{GoField{Name: "Email", GoType: "string"}, "email"},
		{GoField{Name: "Title", GoType: "string"}, ""},
	}
	for _, tt := range tests {
		got := guessFormat(tt.field)
		if got != tt.want {
			t.Errorf("guessFormat({Name:%q, GoType:%q}) = %q, want %q",
				tt.field.Name, tt.field.GoType, got, tt.want)
		}
	}
}

// ── guessFormFieldType ────────────────────────────────────────────────────────

func TestGuessFormFieldType(t *testing.T) {
	tests := []struct {
		field GoField
		want  string
	}{
		{GoField{Name: "Published", GoType: "bool"}, "toggle"},
		{GoField{Name: "Price", GoType: "float64"}, "number"},
		{GoField{Name: "Count", GoType: "int"}, "number"},
		{GoField{Name: "Qty", GoType: "uint"}, "number"},
		{GoField{Name: "PublishedAt", GoType: "*time.Time"}, "datetime"},
		{GoField{Name: "StartDate", GoType: "time.Time"}, "date"},
		{GoField{Name: "Description", GoType: "string"}, "textarea"},
		{GoField{Name: "Notes", GoType: "string"}, "textarea"},
		{GoField{Name: "Title", GoType: "string"}, "text"},
		{GoField{Name: "Email", GoType: "string"}, "text"},
	}
	for _, tt := range tests {
		got := guessFormFieldType(tt.field)
		if got != tt.want {
			t.Errorf("guessFormFieldType({Name:%q, GoType:%q}) = %q, want %q",
				tt.field.Name, tt.field.GoType, got, tt.want)
		}
	}
}

// ── isSortable / isSearchable ─────────────────────────────────────────────────

func TestIsSortableAndSearchable(t *testing.T) {
	sortable := []GoField{
		{Name: "Title", GoType: "string"},
		{Name: "Count", GoType: "int"},
		{Name: "Price", GoType: "float64"},
		{Name: "CreatedAt", GoType: "*time.Time"},
	}
	for _, f := range sortable {
		if !isSortable(f) {
			t.Errorf("isSortable(%q %q) = false, want true", f.Name, f.GoType)
		}
	}

	notSortable := []GoField{
		{Name: "Tags", GoType: "[]string"},
		{Name: "Active", GoType: "bool"},
	}
	for _, f := range notSortable {
		if isSortable(f) {
			t.Errorf("isSortable(%q %q) = true, want false", f.Name, f.GoType)
		}
	}

	if !isSearchable(GoField{Name: "Title", GoType: "string"}) {
		t.Error("isSearchable: string should be searchable")
	}
	if isSearchable(GoField{Name: "Published", GoType: "bool"}) {
		t.Error("isSearchable: bool should not be searchable")
	}
}
