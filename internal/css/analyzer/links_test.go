package analyzer

import (
	"testing"

	"github.com/toba/go-css-lsp/internal/css/parser"
)

func TestFindDocumentLinks_Import(t *testing.T) {
	src := []byte(`@import "styles/reset.css";`)
	ss, _ := parser.Parse(src)
	links := FindDocumentLinks(ss, src)

	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}

	if links[0].Target != "styles/reset.css" {
		t.Errorf(
			"expected styles/reset.css, got %s",
			links[0].Target,
		)
	}
}

func TestFindDocumentLinks_NoLinks(t *testing.T) {
	src := []byte(`.foo { color: red; }`)
	ss, _ := parser.Parse(src)
	links := FindDocumentLinks(ss, src)

	if len(links) != 0 {
		t.Fatalf("expected 0 links, got %d", len(links))
	}
}

func TestFindDocumentLinks_Nil(t *testing.T) {
	links := FindDocumentLinks(nil, nil)
	if links != nil {
		t.Fatal("expected nil for nil stylesheet")
	}
}
