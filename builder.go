package main

import (
	"fmt"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

// BuildAST takes a CST node (root of a parsed source file) and returns our AST.
func BuildAST(root *sitter.Node, src []byte) (*Program, error) {
	if root.Kind() != "source_file" {
		return nil, fmt.Errorf("expected root to be source_file, got %s", root.Kind())
	}

	// source_file -> program
	if root.ChildCount() == 0 {
		return nil, fmt.Errorf("empty source file")
	}
	return buildProgram(root.Child(0), src)
}

// ----------------------------------------------------------------------
// Helpers for extracting fields
// ----------------------------------------------------------------------

func text(node *sitter.Node, src []byte) string {
	if node == nil {
		return ""
	}
	return string(src[node.StartByte():node.EndByte()])
}

// ----------------------------------------------------------------------
// Builders
// ----------------------------------------------------------------------

func buildProgram(n *sitter.Node, src []byte) (*Program, error) {
	if n.Kind() != "program" {
		return nil, fmt.Errorf("expected program node, got %s", n.Kind())
	}

	p := &Program{}

	for i := uint(0); i < n.NamedChildCount(); i++ {
		c := n.NamedChild(i)
		if c == nil {
			continue
		}
		switch c.Kind() {
		case "declaration_statement":
			decl, err := buildVarDecl(c, src)
			if err != nil {
				return nil, err
			}
			p.Declarations = append(p.Declarations, decl)
		case "method_declaration_statement":
			m, err := buildMethodDecl(c, src)
			if err != nil {
				return nil, err
			}
			p.Methods = append(p.Methods, m)
		}
	}

	return p, nil
}

func buildVarDecl(n *sitter.Node, src []byte) (*VarDecl, error) {
	typNode := n.ChildByFieldName("type")
	idNode := n.ChildByFieldName("identifier")
	valNode := n.ChildByFieldName("value")

	t, err := buildType(typNode, src)
	if err != nil {
		return nil, err
	}
	name := Identifier(text(idNode, src))
	val, err := buildExpr(valNode, src)
	if err != nil {
		return nil, err
	}
	return &VarDecl{Type: t, Name: name, Value: val}, nil
}

func buildType(n *sitter.Node, src []byte) (*TypeNode, error) {
	if n == nil {
		return nil, fmt.Errorf("nil type node")
	}
	switch n.Kind() {
	case "void":
		return &TypeNode{Kind: TypeVoid}, nil
	case "bool":
		return &TypeNode{Kind: TypeBool}, nil
	case "integer":
		return &TypeNode{Kind: TypeInteger}, nil
	default:
		return nil, fmt.Errorf("unknown type node: %s", n.Kind())
	}
}

func buildMethodDecl(n *sitter.Node, src []byte) (*MethodDecl, error) {
	retNode := n.ChildByFieldName("type")
	idNode := n.ChildByFieldName("identifier")

	t, err := buildType(retNode, src)
	if err != nil {
		return nil, err
	}
	name := Identifier(text(idNode, src))

	// parameters
	var params []*Parameter
	for i := uint(0); i < n.NamedChildCount(); i++ {
		c := n.NamedChild(i)
		if c.Kind() == "parameter" {
			p, err := buildParameter(c, src)
			if err != nil {
				return nil, err
			}
			params = append(params, p)
		}
	}

	// extern or block
	var body *Block
	extern := false
	for i := uint(0); i < n.ChildCount(); i++ {
		c := n.Child(i)
		if c.Kind() == "extern" {
			extern = true
		}
		if c.Kind() == "block" {
			b, err := buildBlock(c, src)
			if err != nil {
				return nil, err
			}
			body = b
		}
	}

	return &MethodDecl{
		Return: t,
		Name:   name,
		Params: params,
		Body:   body,
		Extern: extern,
	}, nil
}

func buildParameter(n *sitter.Node, src []byte) (*Parameter, error) {
	tNode := n.ChildByFieldName("type")
	idNode := n.ChildByFieldName("identifier")

	t, err := buildType(tNode, src)
	if err != nil {
		return nil, err
	}
	return &Parameter{Type: t, Name: Identifier(text(idNode, src))}, nil
}

// ----------------------------------------------------------------------
// Blocks & Statements
// ----------------------------------------------------------------------

func buildBlock(n *sitter.Node, src []byte) (*Block, error) {
	b := &Block{}
	for i := uint(0); i < n.NamedChildCount(); i++ {
		c := n.NamedChild(i)
		switch c.Kind() {
		case "declaration_statement":
			d, err := buildVarDecl(c, src)
			if err != nil {
				return nil, err
			}
			b.Declarations = append(b.Declarations, d)
		case "assignment_statement":
			as, err := buildAssignment(c, src)
			if err != nil {
				return nil, err
			}
			b.Stmts = append(b.Stmts, as)
		case "return_statement":
			rs, err := buildReturnStmt(c, src)
			if err != nil {
				return nil, err
			}
			b.Stmts = append(b.Stmts, rs)
		case "if_statement":
			is, err := buildIfStmt(c, src)
			if err != nil {
				return nil, err
			}
			b.Stmts = append(b.Stmts, is)
		case "while_statement":
			ws, err := buildWhileStmt(c, src)
			if err != nil {
				return nil, err
			}
			b.Stmts = append(b.Stmts, ws)
		case "method_call":
			e, err := buildExpr(c, src)
			if err != nil {
				return nil, err
			}
			b.Stmts = append(b.Stmts, &ExprStmt{Expr: e})
		}
	}
	return b, nil
}

func buildAssignment(n *sitter.Node, src []byte) (*Assignment, error) {
	idNode := n.ChildByFieldName("identifier")
	valNode := n.ChildByFieldName("value")
	val, err := buildExpr(valNode, src)
	if err != nil {
		return nil, err
	}
	return &Assignment{Target: Identifier(text(idNode, src)), Value: val}, nil
}

func buildReturnStmt(n *sitter.Node, src []byte) (*ReturnStmt, error) {
	valNode := n.ChildByFieldName("value")
	if valNode == nil {
		return &ReturnStmt{}, nil
	}
	val, err := buildExpr(valNode, src)
	if err != nil {
		return nil, err
	}
	return &ReturnStmt{Value: val}, nil
}

func buildIfStmt(n *sitter.Node, src []byte) (*IfStmt, error) {
	condNode := n.ChildByFieldName("condition")
	if condNode == nil {
		// fallback: in your grammar it's field-less, just the first child
		condNode = n.NamedChild(0)
	}
	cond, err := buildExpr(condNode, src)
	if err != nil {
		return nil, err
	}

	var thenBlk, elseBlk *Block
	// second block is then, optional third is else
	blocks := []*sitter.Node{}
	for i := uint(0); i < n.NamedChildCount(); i++ {
		if n.NamedChild(i).Kind() == "block" {
			blocks = append(blocks, n.NamedChild(i))
		}
	}
	if len(blocks) > 0 {
		thenBlk, _ = buildBlock(blocks[0], src)
	}
	if len(blocks) > 1 {
		elseBlk, _ = buildBlock(blocks[1], src)
	}

	return &IfStmt{Cond: cond, Then: thenBlk, Else: elseBlk}, nil
}

func buildWhileStmt(n *sitter.Node, src []byte) (*WhileStmt, error) {
	condNode := n.NamedChild(0)
	cond, err := buildExpr(condNode, src)
	if err != nil {
		return nil, err
	}
	bodyNode := n.NamedChild(n.NamedChildCount() - 1)
	body, err := buildBlock(bodyNode, src)
	if err != nil {
		return nil, err
	}
	return &WhileStmt{Cond: cond, Body: body}, nil
}

// ----------------------------------------------------------------------
// Expressions
// ----------------------------------------------------------------------

func buildExpr(n *sitter.Node, src []byte) (Expr, error) {
	if n == nil {
		return nil, fmt.Errorf("nil expression node")
	}
	switch n.Kind() {
	case "num":
		// parse int
		var v int
		fmt.Sscanf(text(n, src), "%d", &v)
		return &IntLiteral{Value: v}, nil
	case "true":
		return &BoolLiteral{Value: true}, nil
	case "false":
		return &BoolLiteral{Value: false}, nil
	case "identifier":
		return &IdentExpr{Name: Identifier(text(n, src))}, nil
	case "method_call":
		return buildCallExpr(n, src)
	case "int_sum", "int_sub", "int_prod", "int_div",
		"rel_eq", "rel_lt", "rel_gt",
		"bool_conjunction", "bool_disjunction":
		return buildBinaryExpr(n, src)
	case "unary_expression": // if you decide to name it so
		return buildUnaryExpr(n, src)
	case "(": // parenthesized
		inner := n.NamedChild(0)
		return &ParenExpr{Inner: mustExpr(inner, src)}, nil
	}
	return nil, fmt.Errorf("unhandled expression node type: %s", n.Kind())
}

func buildCallExpr(n *sitter.Node, src []byte) (Expr, error) {
	idNode := n.Child(0)
	args := []Expr{}
	for i := uint(0); i < n.NamedChildCount(); i++ {
		c := n.NamedChild(i)
		if c.Kind() == "identifier" && i == 0 {
			continue
		}
		e, err := buildExpr(c, src)
		if err != nil {
			return nil, err
		}
		args = append(args, e)
	}
	return &CallExpr{Callee: Identifier(text(idNode, src)), Args: args}, nil
}

func buildBinaryExpr(n *sitter.Node, src []byte) (Expr, error) {
	left := n.NamedChild(0)
	right := n.NamedChild(1)
	l, err := buildExpr(left, src)
	if err != nil {
		return nil, err
	}
	r, err := buildExpr(right, src)
	if err != nil {
		return nil, err
	}
	var op BinOp
	switch n.Kind() {
	case "int_sum":
		op = BinAdd
	case "int_sub":
		op = BinSub
	case "int_prod":
		op = BinMul
	case "int_div":
		op = BinDiv
	case "rel_eq":
		op = BinEq
	case "rel_lt":
		op = BinLT
	case "rel_gt":
		op = BinGT
	case "bool_conjunction":
		op = BinAnd
	case "bool_disjunction":
		op = BinOr
	}
	return &BinaryExpr{Left: l, Op: op, Right: r}, nil
}

func buildUnaryExpr(n *sitter.Node, src []byte) (Expr, error) {
	// depending on how you labelled it; grammar has "-" $._expression and "!" $._expression
	opNode := n.Child(0)
	exprNode := n.Child(1)
	expr, err := buildExpr(exprNode, src)
	if err != nil {
		return nil, err
	}
	var op UnaryOp
	switch text(opNode, src) {
	case "-":
		op = UnaryNeg
	case "!":
		op = UnaryNot
	default:
		return nil, fmt.Errorf("unknown unary op: %s", text(opNode, src))
	}
	return &UnaryExpr{Op: op, Expr: expr}, nil
}

func mustExpr(n *sitter.Node, src []byte) Expr {
	e, _ := buildExpr(n, src)
	return e
}
