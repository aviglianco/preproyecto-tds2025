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

	// Set the language on the parser
	_ = parser.SetLanguage(lang)

	code := []byte("int main(){return 1;}")

	tree := parser.Parse(code, nil)
	defer tree.Close()

	root := tree.RootNode()

	program, err := BuildProgram(root, code)
	if err != nil {
		fmt.Println("build error:", err)
		return
	}

	fmt.Printf("%#v\n", program)
}
