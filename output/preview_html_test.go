package output

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qa-script/rules"
)

func TestWriteHTMLPreview_BasicLayoutAndPriority(t *testing.T) {
	grouped := rules.TitleGroupedLocations{
		"A": []string{"loc1", "loc2", "loc3", "loc4"},
	}
	// MaxRows=2 => 2 columns; loc2 is priority
	data := NewOutputData([]string{"A"}, grouped, []string{"loc2"}, 0, 2)

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "preview.html")
	if err := WriteHTMLPreview(outPath, data); err != nil {
		t.Fatalf("WriteHTMLPreview failed: %v", err)
	}

	contentBytes, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read preview: %v", err)
	}
	content := string(contentBytes)

	// Header merged across 2 columns
	if !strings.Contains(content, `<th colspan="2">A</th>`) {
		t.Fatalf("expected merged header colspan=2, content=%q", content)
	}
	// Priority cell class present
	if !strings.Contains(content, `class="priority">loc2</td>`) {
		t.Fatalf("expected priority cell class, content=%q", content)
	}
}
