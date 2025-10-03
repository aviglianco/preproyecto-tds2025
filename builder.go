package main

import (
	"fmt"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

type Builder struct {
	symbolTable Env
	src         []byte
}

// BuildAST takes a CST node (root of a parsed source file) and returns our AST.
func BuildAST(root *sitter.Node, src []byte) (*Program, error) {
	if root.Kind() != "source_file" {
		return nil, fmt.Errorf("expected root to be source_file, got %s", root.Kind())
	}

	// source_file -> program
	if root.ChildCount() == 0 {
		return nil, fmt.Errorf("empty source file")
	}

	symbolTable := Env{Table: make(map[Identifier]Symbol)}

	builder := Builder{
		src:         src,
		symbolTable: symbolTable,
	}

	return builder.buildProgram(root.Child(0))
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

func (builder Builder) buildProgram(n *sitter.Node) (*Program, error) {
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
			decl, err := builder.buildVarDecl(c)
			if err != nil {
				return nil, err
			}
			p.Declarations = append(p.Declarations, decl)
		case "method_declaration_statement":
			m, err := builder.buildMethodDecl(c)
			if err != nil {
				return nil, err
			}
			p.Methods = append(p.Methods, m)
		}
	}

	return p, nil
}

func (builder Builder) buildVarDecl(n *sitter.Node) (*VarDecl, error) {
	typNode := n.ChildByFieldName("type")
	idNode := n.ChildByFieldName("identifier")
	valNode := n.ChildByFieldName("value")

	t, err := builder.buildType(typNode)
	if err != nil {
		return nil, err
	}
	name := Identifier(text(idNode, builder.src))
	val, err := builder.buildExpr(valNode)

	_, ok := builder.symbolTable.Table[name]
	if ok {
		return nil, fmt.Errorf("cannot double declare :%s", name)
	} else {
		builder.symbolTable.Insert(name, Symbol{Type: t, isVar: true})
	}

	if err != nil {
		return nil, err
	}
	return &VarDecl{NodeBase: NodeBase{Line: nodeLine(n)}, Type: t, Name: name, Value: val}, nil
}

func (builder Builder) buildType(n *sitter.Node) (TypeKind, error) {
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

func (builder Builder) buildMethodDecl(n *sitter.Node) (*MethodDecl, error) {
	retNode := n.ChildByFieldName("type")
	idNode := n.ChildByFieldName("identifier")

	t, err := builder.buildType(retNode)
	if err != nil {
		return nil, err
	}
	name := Identifier(text(idNode, builder.src))

	// Type checking within the same frame
	_, ok := builder.symbolTable.Table[name]
	if ok {
		return nil, fmt.Errorf("cannot redefine:%s", name)
	} else { // insert function into symbol table
		builder.symbolTable.Insert(name, Symbol{Type: t, isVar: false})
	}

	// parameters
	var params []*Parameter
	for i := uint(0); i < n.NamedChildCount(); i++ {
		c := n.NamedChild(i)
		if c.Kind() == "parameter" {
			p, err := builder.buildParameter(c)
			if err != nil {
				return nil, err
			}
			params = append(params, p)
		}
	}

	if len(params) > 0 {
		paramNames := make(map[Identifier]struct{})
		for _, p := range params {
			if _, clash := paramNames[p.Name]; clash {
				return nil, fmt.Errorf("duplicate parameter name: %s", p.Name)
			}
			paramNames[p.Name] = struct{}{}
		}

		prevEnv := builder.symbolTable
		funcEnv := Env{Prev: &prevEnv, Table: make(Table)}
		for _, p := range params {
			funcEnv.Insert(p.Name, Symbol{Type: p.Type, isVar: true})
		}
		builder.symbolTable = funcEnv
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
			b, err := builder.buildBlock(c)
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

func (builder Builder) buildParameter(n *sitter.Node) (*Parameter, error) {
	tNode := n.ChildByFieldName("type")
	idNode := n.ChildByFieldName("identifier")

	t, err := builder.buildType(tNode)
	if err != nil {
		return nil, err
	}
	return &Parameter{NodeBase: NodeBase{Line: nodeLine(n)}, Type: t, Name: Identifier(text(idNode, builder.src))}, nil
}

// ----------------------------------------------------------------------
// Blocks & Statements
// ----------------------------------------------------------------------

func (builder Builder) buildBlock(n *sitter.Node) (*Block, error) {
	b := &Block{NodeBase: NodeBase{Line: nodeLine(n)}}
	prevEnv := builder.symbolTable
	builder.symbolTable = Env{Prev: &prevEnv, Table: make(Table)}

	for i := uint(0); i < n.NamedChildCount(); i++ {
		c := n.NamedChild(i)
		switch c.Kind() {
		case "declaration_statement":
			d, err := builder.buildVarDecl(c)
			if err != nil {
				return nil, err
			}
			b.Declarations = append(b.Declarations, d)
		case "assignment_statement":
			as, err := builder.buildAssignment(c)
			if err != nil {
				return nil, err
			}
			b.Stmts = append(b.Stmts, as)
		case "return_statement":
			rs, err := builder.buildReturnStmt(c)
			if err != nil {
				return nil, err
			}
			b.Stmts = append(b.Stmts, rs)
		case "if_statement":
			is, err := builder.buildIfStmt(c)
			if err != nil {
				return nil, err
			}
			b.Stmts = append(b.Stmts, is)
		case "while_statement":
			ws, err := builder.buildWhileStmt(c)
			if err != nil {
				return nil, err
			}
			b.Stmts = append(b.Stmts, ws)
		case "method_call":
			e, err := builder.buildExpr(c)
			if err != nil {
				return nil, err
			}
			b.Stmts = append(b.Stmts, &ExprStmt{Expr: e})
		}
	}

	return b, nil
}

func (builder Builder) buildAssignment(n *sitter.Node) (*Assignment, error) {
	idNode := n.ChildByFieldName("identifier")
	valNode := n.ChildByFieldName("value")
	val, err := builder.buildExpr(valNode)
	if err != nil {
		return nil, err
	}
	return &Assignment{NodeBase: NodeBase{Line: nodeLine(n)}, Target: Identifier(text(idNode, builder.src)), Value: val}, nil
}

func (builder Builder) buildReturnStmt(n *sitter.Node) (*ReturnStmt, error) {
	valNode := n.ChildByFieldName("value")
	if valNode == nil {
		return &ReturnStmt{NodeBase: NodeBase{Line: nodeLine(n)}}, nil
	}
	val, err := builder.buildExpr(valNode)
	if err != nil {
		return nil, err
	}
	return &ReturnStmt{NodeBase: NodeBase{Line: nodeLine(n)}, Value: val}, nil
}

func (builder Builder) buildIfStmt(n *sitter.Node) (*IfStmt, error) {
	condNode := n.ChildByFieldName("condition")
	if condNode == nil {
		// fallback: in your grammar it's field-less, just the first child
		condNode = n.NamedChild(0)
	}
	cond, err := builder.buildExpr(condNode)
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
		thenBlk, _ = builder.buildBlock(blocks[0])
	}
	if len(blocks) > 1 {
		elseBlk, _ = builder.buildBlock(blocks[1])
	}

	return &IfStmt{NodeBase: NodeBase{Line: nodeLine(n)}, Cond: cond, Then: thenBlk, Else: elseBlk}, nil
}

func (builder Builder) buildWhileStmt(n *sitter.Node) (*WhileStmt, error) {
	condNode := n.NamedChild(0)
	cond, err := builder.buildExpr(condNode)
	if err != nil {
		return nil, err
	}
	bodyNode := n.NamedChild(n.NamedChildCount() - 1)
	body, err := builder.buildBlock(bodyNode)
	if err != nil {
		return nil, err
	}
	return &WhileStmt{NodeBase: NodeBase{Line: nodeLine(n)}, Cond: cond, Body: body}, nil
}

// ----------------------------------------------------------------------
// Expressions
// ----------------------------------------------------------------------

func (builder Builder) buildExpr(n *sitter.Node) (Expr, error) {
	if n == nil {
		return nil, fmt.Errorf("nil expression node")
	}
	switch n.Kind() {
	case "num":
		// parse int
		var v int
		fmt.Sscanf(text(n, builder.src), "%d", &v)
		return &IntLiteral{NodeBase: NodeBase{Line: nodeLine(n)}, Value: v, Type: TypeInteger}, nil
	case "true":
		return &BoolLiteral{NodeBase: NodeBase{Line: nodeLine(n)}, Value: true, Type: TypeBool}, nil
	case "false":
		return &BoolLiteral{NodeBase: NodeBase{Line: nodeLine(n)}, Value: false, Type: TypeBool}, nil
	case "identifier":
		name := Identifier(text(n, builder.src))
		symbol, ok := builder.symbolTable.Lookup(name)
		if !ok {
			return nil, fmt.Errorf("could not resolve type of %s", name)
		}
		return &IdentExpr{NodeBase: NodeBase{Line: nodeLine(n)}, Name: name, Type: symbol.Type}, nil
	case "method_call":
		return builder.buildCallExpr(n)
	case "int_sum", "int_sub", "int_prod", "int_div",
		"rel_eq", "rel_lt", "rel_gt",
		"bool_conjunction", "bool_disjunction":
		return builder.buildBinaryExpr(n)
	case "unary_expression": // if you decide to name it so
		return builder.buildUnaryExpr(n)
	case "(": // parenthesized
		inner := n.NamedChild(0)
		return &ParenExpr{NodeBase: NodeBase{Line: nodeLine(n)}, Inner: builder.mustExpr(inner)}, nil
	}
	return nil, fmt.Errorf("unhandled expression node type: %s", n.Kind())
}

func (builder Builder) buildCallExpr(n *sitter.Node) (Expr, error) {
	idNode := n.Child(0)
	args := []Expr{}
	for i := uint(0); i < n.NamedChildCount(); i++ {
		c := n.NamedChild(i)
		if c.Kind() == "identifier" && i == 0 {
			continue
		}
		e, err := builder.buildExpr(c)
		if err != nil {
			return nil, err
		}
		args = append(args, e)
	}
	return &CallExpr{NodeBase: NodeBase{Line: nodeLine(n)}, Callee: Identifier(text(idNode, builder.src)), Args: args}, nil
}

func (builder Builder) buildBinaryExpr(n *sitter.Node) (Expr, error) {
	left := n.NamedChild(0)
	right := n.NamedChild(1)
	l, err := builder.buildExpr(left)
	if err != nil {
		return nil, err
	}
	r, err := builder.buildExpr(right)
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

func (builder Builder) buildUnaryExpr(n *sitter.Node) (Expr, error) {
	// depending on how you labelled it; grammar has "-" $._expression and "!" $._expression
	opNode := n.Child(0)
	exprNode := n.Child(1)
	expr, err := builder.buildExpr(exprNode)
	if err != nil {
		return nil, err
	}
	var op UnaryOp
	var t TypeKind
	switch text(opNode, builder.src) {
	case "-":
		op = UnaryNeg
		t = TypeInteger
	case "!":
		op = UnaryNot
		t = TypeBool
	default:
		return nil, fmt.Errorf("unknown unary op: %s", text(opNode, builder.src))
	}
	return &UnaryExpr{NodeBase: NodeBase{Line: nodeLine(n)}, Op: op, Expr: expr, Type: t}, nil
}

func (builder Builder) mustExpr(n *sitter.Node) Expr {
	e, _ := builder.buildExpr(n)
	return e
}
