package main

import (
	"fmt"
	"strings"
)

func printAST(program *Program) string {
	if program == nil {
		return "<nil>"
	}
	b := &strings.Builder{}
	printProgram(b, program, 0)
	return b.String()
}

func indentOf(level int) string {
	if level <= 0 {
		return ""
	}
	return strings.Repeat("  ", level)
}

func printProgram(b *strings.Builder, p *Program, level int) {
	fmt.Fprintf(b, "Program(ReturnType=%s)\n", typeString(p.ReturnType))
	if p.Main != nil {
		printMain(b, p.Main, level+1)
	}
}

func printMain(b *strings.Builder, m *Main, level int) {
	indent := indentOf(level)
	fmt.Fprintf(b, "%sMain(Args=%v)\n", indent, m.Args)
	if m.Body != nil {
		printBlock(b, m.Body, level+1)
	}
}

func printBlock(b *strings.Builder, blk *Block, level int) {
	indent := indentOf(level)
	fmt.Fprintf(b, "%sBlock\n", indent)
	for _, s := range blk.Statements {
		printStmt(b, s, level+1)
	}
}

func printStmt(b *strings.Builder, s Stmt, level int) {
	switch st := s.(type) {
	case *Decl:
		fmt.Fprintf(b, "%sDecl(type=%s, name=%s)\n", indentOf(level), typeString(st.VarType), st.Name)
	case *Assign:
		fmt.Fprintf(b, "%sAssign(name=%s)\n", indentOf(level), st.Name)
		printExpr(b, st.Value, level+1)
	case *Return:
		fmt.Fprintf(b, "%sReturn\n", indentOf(level))
		if st.Value != nil {
			printExpr(b, st.Value, level+1)
		}
	case *Skip:
		fmt.Fprintf(b, "%sSkip\n", indentOf(level))
	default:
		fmt.Fprintf(b, "%s<unknown stmt>\n", indentOf(level))
	}
}

func printExpr(b *strings.Builder, e Expr, level int) {
	indent := indentOf(level)
	switch ex := e.(type) {
	case *Identifier:
		fmt.Fprintf(b, "%sIdentifier(%s)\n", indent, ex.Name)
	case *IntLiteral:
		fmt.Fprintf(b, "%sInt(%d)\n", indent, ex.Value)
	case *BoolLiteral:
		fmt.Fprintf(b, "%sBool(%t)\n", indent, ex.Value)
	case *BinaryExpr:
		fmt.Fprintf(b, "%sBinary(%s)\n", indent, opString(ex.Kind))
		printExpr(b, ex.Left, level+1)
		printExpr(b, ex.Right, level+1)
	default:
		fmt.Fprintf(b, "%s<unknown expr>\n", indent)
	}
}

func typeString(t Type) string {
	switch t {
	case TypeInt:
		return "int"
	case TypeBool:
		return "bool"
	case TypeVoid:
		return "void"
	default:
		return "unknown"
	}
}

func opString(k BinaryOpKind) string {
	switch k {
	case OpMul:
		return "*"
	case OpDiv:
		return "/"
	case OpAdd:
		return "+"
	case OpSub:
		return "-"
	default:
		return "?"
	}
}
