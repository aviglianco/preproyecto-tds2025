package main

import (
	"fmt"

	parserlang "compilador/bindings/go"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

const testCode = `bool main(){
		int a;
		a = 5 + 2;
		bool b;
		b = true;
		int c;
		c = a + 17;
		return c;
	;}`

func main() {
	parser := sitter.NewParser()
	defer parser.Close()

	// Wrap the unsafe.Pointer from parserlang.Language()
	rawLang := parserlang.Language()
	lang := sitter.NewLanguage(rawLang)

	// Set the language on the parser
	e := parser.SetLanguage(lang)
	if e != nil {
		panic(fmt.Errorf("couldn't configure parser: %w", e))
	}

	code := []byte(testCode)

	// Parse the code
	tree := parser.Parse(code, nil)
	defer tree.Close()

	// Get the root node
	root := tree.RootNode()

	// Pretty-print the syntax tree
	fmt.Println(root.ToSexp())

	// Build AST
	program, err := buildProgram(root, code)
	if err != nil {
		fmt.Println("build error:", err)
		return
	}

	// Pretty-print the resulting AST
	fmt.Println(printAST(program))

	// Generate assembly
	asm := generateAssembly(program)
	fmt.Println(asm)
}
