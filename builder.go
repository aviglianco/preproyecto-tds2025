package main

import (
	"fmt"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

func BuildProgram(root *sitter.Node, source []byte) (*Program, error) {
	if root == nil || root.Kind() != "source_file" {
		return nil, fmt.Errorf("unexpected root: %v", kindOf(root))
	}

	var retType Type = TypeVoid
	var mainNode *sitter.Node

	for i := uint(0); i < root.ChildCount(); i++ {
		ch := root.Child(i)
		switch ch.Kind() {
		case "int":
			retType = TypeInt
		case "bool":
			retType = TypeBool
		case "void":
			retType = TypeVoid
		case "main":
			mainNode = ch
		}
	}

	if mainNode == nil {
		return nil, fmt.Errorf("main not found")
	}
	m, err := buildMain(mainNode, source)
	if err != nil {
		return nil, err
	}
	return &Program{ReturnType: retType, Main: m}, nil
}

func buildMain(n *sitter.Node, src []byte) (*Main, error) {
	if n == nil || n.Kind() != "main" {
		return nil, fmt.Errorf("expected main, got %v", kindOf(n))
	}
	argsNode := n.ChildByFieldName("args")
	blockNode := n.ChildByFieldName("block")

	args := make([]string, 0)
	if argsNode != nil {
		for i := uint(0); i < argsNode.ChildCount(); i++ {
			ch := argsNode.Child(i)
			if ch.Kind() == "arg" {
				args = append(args, "arg")
			}
		}
	}

	blk, err := buildBlock(blockNode, src)
	if err != nil {
		return nil, err
	}
	return &Main{Args: args, Body: blk}, nil
}

func buildBlock(n *sitter.Node, src []byte) (*Block, error) {
	if n == nil {
		return nil, fmt.Errorf("nil block")
	}
	stmts := make([]Stmt, 0)
	for i := uint(0); i < n.NamedChildCount(); i++ {
		ch := n.NamedChild(i)
		s, ok, err := buildStatement(ch, src)
		if err != nil {
			return nil, err
		}
		if ok {
			stmts = append(stmts, s)
		}
	}
	return &Block{Statements: stmts}, nil
}

func buildStatement(n *sitter.Node, src []byte) (Stmt, bool, error) {
	switch n.Kind() {
	case "declaration_statement":
		tn := n.ChildByFieldName("type")
		id := n.ChildByFieldName("identifier")
		return &Decl{VarType: toType(tn.Kind()), Name: textOf(id, src)}, true, nil
	case "assignment_statement":
		id := n.ChildByFieldName("identifier")
		val := n.ChildByFieldName("value")
		ex, err := buildExpr(val, src)
		if err != nil {
			return nil, false, err
		}
		return &Assign{Name: textOf(id, src), Value: ex}, true, nil
	case "return_statement":
		val := n.ChildByFieldName("value")
		if val == nil {
			return &Return{Value: nil}, true, nil
		}
		ex, err := buildExpr(val, src)
		if err != nil {
			return nil, false, err
		}
		return &Return{Value: ex}, true, nil
	case "skip":
		return &Skip{}, true, nil
	default:
		// expression statement not modeled; ignore
		ex, err := buildExpr(n, src)
		if err != nil {
			return nil, false, nil
		}
		_ = ex
		return nil, false, nil
	}
}

func buildExpr(n *sitter.Node, src []byte) (Expr, error) {
	if n == nil {
		return nil, fmt.Errorf("nil expr node")
	}
	// Unwrap wrappers/parentheses by drilling into the first named child
	if n.Kind() == "_expression" || n.Kind() == "_exp" {
		for i := uint(0); i < n.ChildCount(); i++ {
			ch := n.Child(i)
			if ch.IsNamed() {
				return buildExpr(ch, src)
			}
		}
	}

	switch n.Kind() {
	case "identifier":
		return &Identifier{Name: textOf(n, src)}, nil
	case "num":
		return &IntLiteral{Value: textOf(n, src)}, nil
	case "true":
		return &BoolLiteral{Value: true}, nil
	case "false":
		return &BoolLiteral{Value: false}, nil
	case "int_proc", "int_div", "int_sum", "int_sub":
		l := n.ChildByFieldName("left")
		r := n.ChildByFieldName("right")
		lx, err := buildExpr(l, src)
		if err != nil {
			return nil, err
		}
		rx, err := buildExpr(r, src)
		if err != nil {
			return nil, err
		}
		return &BinaryExpr{Kind: toOpKind(n.Kind()), Left: lx, Right: rx}, nil
	default:
		for i := uint(0); i < n.ChildCount(); i++ {
			ch := n.Child(i)
			if ch.IsNamed() {
				return buildExpr(ch, src)
			}
		}
	}
	return nil, fmt.Errorf("unhandled expr kind: %s", n.Kind())
}

func toType(kind string) Type {
	switch kind {
	case "int":
		return TypeInt
	case "bool":
		return TypeBool
	case "void":
		return TypeVoid
	default:
		return TypeInt
	}
}

func toOpKind(kind string) BinaryOpKind {
	switch kind {
	case "int_proc":
		return OpMul
	case "int_div":
		return OpDiv
	case "int_sum":
		return OpAdd
	case "int_sub":
		return OpSub
	default:
		return OpAdd
	}
}

func textOf(n *sitter.Node, src []byte) string {
	if n == nil {
		return ""
	}
	start := int(n.StartByte())
	end := int(n.EndByte())
	if start < 0 || end > len(src) || start > end {
		return ""
	}
	return string(src[start:end])
}

func kindOf(n *sitter.Node) string {
	if n == nil {
		return "<nil>"
	}
	return n.Kind()
}
