package main

import (
	"fmt"

	parserlang "compilador/bindings/go"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

func main() {
	parser := sitter.NewParser()
	defer parser.Close()

	// Wrap the unsafe.Pointer from parserlang.Language()
	rawLang := parserlang.Language()
	lang := sitter.NewLanguage(rawLang)

	// Set the language on the parser
	e := parser.SetLanguage(lang)

	code := []byte("bool main(){int a; a = 2 + 2;return 1+2;}")

	if e != nil {
		panic(fmt.Errorf("couldn't configure parser: %w", e))
	}

	fmt.Println(parser.Language())

	tree := parser.Parse(code, nil)
	defer tree.Close()

	root := tree.RootNode()

	fmt.Println(root.ToSexp())

	program, err := BuildProgram(root, code)
	if err != nil {
		fmt.Println("build error:", err)
		return
	}

	// Pretty-print the resulting AST
	fmt.Println(PrettyProgram(program))

	interpret(*program)
}
