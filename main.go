package main

import (
	"fmt"
	"os"
	"path/filepath"

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
	if e != nil {
		panic(fmt.Errorf("couldn't configure parser: %w", e))
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: compilador <input.ctds>")
		os.Exit(1)
	}

	inputArg := os.Args[1]

	if filepath.Ext(inputArg) != ".ctds" {
		fmt.Fprintln(os.Stderr, "error: input file must have .ctds extension")
		os.Exit(1)
	}

	var code []byte
	var err error
	code, err = os.ReadFile(inputArg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
		os.Exit(1)
	}

	// Parse the code
	tree := parser.Parse(code, nil)
	defer tree.Close()

	// Get the root node
	root := tree.RootNode()

	if root.HasError() {
		fmt.Fprintf(os.Stderr, "could not parse file %s: syntax error\n", inputArg)

		os.Exit(1)
	}

	ast, err := BuildAST(root, code)

	if err != nil {
		fmt.Printf("Coudldn't buil AST: %s", err.Error())
	}
	fmt.Println(ast)

	// Pretty-print the syntax tree and write to .sint file
	output := []byte(root.ToSexp())
	base := inputArg[:len(inputArg)-len(filepath.Ext(inputArg))]
	outputPath := base + ".sint"
	if err := os.WriteFile(outputPath, output, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Output written to:", outputPath)
}
