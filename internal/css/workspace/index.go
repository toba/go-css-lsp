// Package workspace provides workspace-wide CSS variable
// indexing for cross-file features.
package workspace

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/toba/go-css-lsp/internal/css/parser"
	"github.com/toba/go-css-lsp/internal/css/scanner"
)

const (
	varFunctionName      = "var"
	customPropertyPrefix = "--"
)

// skipDirs contains directory names to skip during workspace
// scanning.
var skipDirs = map[string]bool{
	"node_modules":     true,
	".git":             true,
	"dist":             true,
	"vendor":           true,
	".next":            true,
	"bower_components": true,
}

// VariableDefinition represents a CSS custom property
// definition.
type VariableDefinition struct {
	Name     string
	URI      string
	StartPos int
	EndPos   int
}

// Index maintains a workspace-wide index of CSS custom
// properties.
type Index struct {
	mu          sync.RWMutex
	definitions map[string][]VariableDefinition // name -> defs
	fileVars    map[string][]string             // uri -> var names
}

// NewIndex creates a new workspace index.
func NewIndex() *Index {
	return &Index{
		definitions: make(map[string][]VariableDefinition),
		fileVars:    make(map[string][]string),
	}
}

// ScanWorkspace scans all CSS files in the root directory and
// indexes their custom properties.
func (idx *Index) ScanWorkspace(rootPath string) error {
	return filepath.WalkDir(rootPath, func(
		path string,
		d fs.DirEntry,
		err error,
	) error {
		if err != nil {
			return err
		}

		// Skip common non-source directories
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".css") {
			return nil
		}

		src, err := os.ReadFile(path) //nolint:gosec
		if err != nil {
			return err
		}

		uri := "file://" + path
		idx.IndexFile(uri, src)
		return nil
	})
}

// IndexFile indexes a single file's custom properties.
func (idx *Index) IndexFile(uri string, src []byte) {
	ss, _ := parser.Parse(src)
	if ss == nil {
		return
	}
	idx.IndexFileWithStylesheet(uri, ss)
}

// IndexFileWithStylesheet indexes a file's custom properties
// using a pre-parsed stylesheet, avoiding a redundant parse.
func (idx *Index) IndexFileWithStylesheet(
	uri string,
	ss *parser.Stylesheet,
) {
	if ss == nil {
		return
	}

	var defs []VariableDefinition
	var names []string

	parser.Walk(ss, func(n parser.Node) bool {
		decl, ok := n.(*parser.Declaration)
		if !ok {
			return true
		}

		name := decl.Property.Value
		if !strings.HasPrefix(name, customPropertyPrefix) {
			return true
		}

		defs = append(defs, VariableDefinition{
			Name:     name,
			URI:      uri,
			StartPos: decl.Property.Offset,
			EndPos:   decl.Property.End,
		})
		names = append(names, name)

		return true
	})

	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Remove old definitions for this file
	idx.removeFileVarsLocked(uri)

	// Add new definitions
	for _, def := range defs {
		idx.definitions[def.Name] = append(
			idx.definitions[def.Name], def,
		)
	}
	idx.fileVars[uri] = names
}

// RemoveFile removes a file's entries from the index.
func (idx *Index) RemoveFile(uri string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.removeFileVarsLocked(uri)
}

func (idx *Index) removeFileVarsLocked(uri string) {
	names, ok := idx.fileVars[uri]
	if !ok {
		return
	}

	for _, name := range names {
		defs := idx.definitions[name]
		filtered := make([]VariableDefinition, 0, len(defs))
		for _, d := range defs {
			if d.URI != uri {
				filtered = append(filtered, d)
			}
		}
		if len(filtered) == 0 {
			delete(idx.definitions, name)
		} else {
			idx.definitions[name] = filtered
		}
	}
	delete(idx.fileVars, uri)
}

// LookupDefinitions returns all definitions for a custom
// property name across the workspace.
func (idx *Index) LookupDefinitions(
	name string,
) []VariableDefinition {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	defs := idx.definitions[name]
	result := make([]VariableDefinition, len(defs))
	copy(result, defs)
	return result
}

// AllVariableNames returns all known custom property names.
func (idx *Index) AllVariableNames() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	names := make([]string, 0, len(idx.definitions))
	for name := range idx.definitions {
		names = append(names, name)
	}
	return names
}

// FindReferences returns all locations where a custom
// property is used (var() calls) across indexed files.
func (idx *Index) FindReferences(
	name string,
	files map[string][]byte,
	parsedFiles map[string]*parser.Stylesheet,
) []VariableDefinition {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	// Start with definitions
	var refs []VariableDefinition
	refs = append(refs, idx.definitions[name]...)

	// Search all known files for var() usages
	for uri, src := range files {
		ss := parsedFiles[uri]
		if ss == nil {
			parsed, _ := parser.Parse(src)
			ss = parsed
		}
		if ss == nil {
			continue
		}

		parser.Walk(ss, func(n parser.Node) bool {
			decl, ok := n.(*parser.Declaration)
			if !ok {
				return true
			}
			if decl.Value == nil {
				return true
			}

			tokens := decl.Value.Tokens
			for i, tok := range tokens {
				if tok.Kind != scanner.Function {
					continue
				}
				if strings.ToLower(tok.Value) != varFunctionName {
					continue
				}
				for j := i + 1; j < len(tokens); j++ {
					if tokens[j].Kind == scanner.Whitespace {
						continue
					}
					if tokens[j].Kind == scanner.Ident &&
						tokens[j].Value == name {
						refs = append(refs,
							VariableDefinition{
								Name:     name,
								URI:      uri,
								StartPos: tokens[j].Offset,
								EndPos:   tokens[j].End,
							},
						)
					}
					break
				}
			}
			return true
		})
	}

	return refs
}
