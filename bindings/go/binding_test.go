package tree_sitter_parser_test

import (
	"testing"

	tree_sitter_parser "compilador/bindings/go"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

func TestCanLoadGrammar(t *testing.T) {
	language := tree_sitter.NewLanguage(tree_sitter_parser.Language())
	if language == nil {
		t.Errorf("Error loading Parser grammar")
	}
}
