package shared

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestDomainImportsNoOtherLayer enforces ADR-0011: the domain layer imports no other layer.
func TestDomainImportsNoOtherLayer(t *testing.T) {
	forbidden := []string{
		"/internal/application",
		"/internal/infrastructure",
		"/internal/presentation",
	}
	root := filepath.Join("..") // internal/domain
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			for _, layer := range forbidden {
				if strings.Contains(p, layer) {
					t.Errorf("%s imports %s", path, p)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
