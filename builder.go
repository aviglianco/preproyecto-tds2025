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

// nodeLine returns the 1-based start line for a CST node.
func nodeLine(n *sitter.Node) int {
	if n == nil {
		return 0
	}
	return int(n.Range().StartPoint.Row) + 1
}

// ----------------------------------------------------------------------
// Builders
// ----------------------------------------------------------------------

func buildProgram(n *sitter.Node, src []byte) (*Program, error) {
	if n.Kind() != "program" {
		return nil, fmt.Errorf("expected program node, got %s", n.Kind())
	}

	p := &Program{NodeBase: NodeBase{Line: nodeLine(n)}}

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

	t, err := buildType(typNode)
	if err != nil {
		return nil, err
	}
	name := Identifier(text(idNode, src))
	val, err := buildExpr(valNode, src)
	if err != nil {
		return nil, err
	}
	return &VarDecl{NodeBase: NodeBase{Line: nodeLine(n)}, Type: t, Name: name, Value: val}, nil
}

func buildType(n *sitter.Node) (TypeKind, error) {
	if n == nil {
		return 0, fmt.Errorf("nil type node")
	}
	switch n.Kind() {
	case "void":
		return TypeVoid, nil
	case "bool":
		return TypeBool, nil
	case "integer":
		return TypeInteger, nil
	default:
		return 0, fmt.Errorf("unknown type node: %s", n.Kind())
	}
}

func buildMethodDecl(n *sitter.Node, src []byte) (*MethodDecl, error) {
	retNode := n.ChildByFieldName("type")
	idNode := n.ChildByFieldName("identifier")

	t, err := buildType(retNode)
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
		NodeBase: NodeBase{Line: nodeLine(n)},
		Return:   t,
		Name:     name,
		Params:   params,
		Body:     body,
		Extern:   extern,
	}, nil
}

func buildParameter(n *sitter.Node, src []byte) (*Parameter, error) {
	tNode := n.ChildByFieldName("type")
	idNode := n.ChildByFieldName("identifier")

	t, err := buildType(tNode)
	if err != nil {
		return nil, err
	}
	return &Parameter{NodeBase: NodeBase{Line: nodeLine(n)}, Type: t, Name: Identifier(text(idNode, src))}, nil
}

// ----------------------------------------------------------------------
// Blocks & Statements
// ----------------------------------------------------------------------

func buildBlock(n *sitter.Node, src []byte) (*Block, error) {
	b := &Block{NodeBase: NodeBase{Line: nodeLine(n)}}
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
	return &Assignment{NodeBase: NodeBase{Line: nodeLine(n)}, Target: Identifier(text(idNode, src)), Value: val}, nil
}

func buildReturnStmt(n *sitter.Node, src []byte) (*ReturnStmt, error) {
	valNode := n.ChildByFieldName("value")
	if valNode == nil {
		return &ReturnStmt{NodeBase: NodeBase{Line: nodeLine(n)}}, nil
	}
	val, err := buildExpr(valNode, src)
	if err != nil {
		return nil, err
	}
	return &ReturnStmt{NodeBase: NodeBase{Line: nodeLine(n)}, Value: val}, nil
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

	return &IfStmt{NodeBase: NodeBase{Line: nodeLine(n)}, Cond: cond, Then: thenBlk, Else: elseBlk}, nil
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
	return &WhileStmt{NodeBase: NodeBase{Line: nodeLine(n)}, Cond: cond, Body: body}, nil
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
		return &IntLiteral{NodeBase: NodeBase{Line: nodeLine(n)}, Value: v, Type: TypeInteger}, nil
	case "true":
		return &BoolLiteral{NodeBase: NodeBase{Line: nodeLine(n)}, Value: true, Type: TypeBool}, nil
	case "false":
		return &BoolLiteral{NodeBase: NodeBase{Line: nodeLine(n)}, Value: false, Type: TypeBool}, nil
	case "identifier":
		return &IdentExpr{NodeBase: NodeBase{Line: nodeLine(n)}, Name: Identifier(text(n, src))}, nil
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
		return &ParenExpr{NodeBase: NodeBase{Line: nodeLine(n)}, Inner: mustExpr(inner, src)}, nil
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
	return &CallExpr{NodeBase: NodeBase{Line: nodeLine(n)}, Callee: Identifier(text(idNode, src)), Args: args}, nil
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
	var t TypeKind

	switch n.Kind() {
	case "int_sum":
		op = BinAdd
		t = TypeInteger
	case "int_sub":
		op = BinSub
		t = TypeInteger
	case "int_prod":
		op = BinMul
		t = TypeInteger
	case "int_div":
		op = BinDiv
		t = TypeInteger
	case "rel_eq":
		op = BinEq
		t = TypeBool
	case "rel_lt":
		op = BinLT
		t = TypeBool
	case "rel_gt":
		op = BinGT
		t = TypeBool
	case "bool_conjunction":
		op = BinAnd
		t = TypeBool
	case "bool_disjunction":
		op = BinOr
		t = TypeBool
	}
	return &BinaryExpr{NodeBase: NodeBase{Line: nodeLine(n)}, Left: l, Op: op, Right: r, Type: t}, nil
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
	var t TypeKind
	switch text(opNode, src) {
	case "-":
		op = UnaryNeg
		t = TypeInteger
	case "!":
		op = UnaryNot
		t = TypeBool
	default:
		return nil, fmt.Errorf("unknown unary op: %s", text(opNode, src))
	}
	return &UnaryExpr{NodeBase: NodeBase{Line: nodeLine(n)}, Op: op, Expr: expr, Type: t}, nil
}

func mustExpr(n *sitter.Node, src []byte) Expr {
	e, _ := buildExpr(n, src)
	return e
}
