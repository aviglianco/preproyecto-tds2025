package main

import (
	"fmt"

	sitter "github.com/tree-sitter/go-tree-sitter"
	parserlang "github.com/tree-sitter/tree-sitter-parser/bindings/go"
)

func main() {
	parser := sitter.NewParser()
	defer parser.Close()

	// Wrap the unsafe.Pointer from parserlang.Language()
	rawLang := parserlang.Language()

	lang := sitter.NewLanguage(rawLang)

	fmt.Println(lang.Version())

	// Set the language on the parser
	e := parser.SetLanguage(lang)

	if e != nil {
		fmt.Println(e)
	}

	fmt.Println(lang)

	code := []byte("void main(){return 1}")

	tree := parser.Parse(code, nil)
	defer tree.Close()

	root := tree.RootNode()
	fmt.Print(root)

}
