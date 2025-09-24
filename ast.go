package main

import "strconv"

// Node is the common interface implemented by all AST nodes.
type Node interface {
	// optionally add Pos/Span methods here later
	NodeType() string
}

// ===== Program / Top-level =====

type Program struct {
	// e.g. "program { ... }"
	Declarations []*VarDecl    // top-level variable declarations
	Methods      []*MethodDecl // top-level method (function) declarations
}

func (p *Program) NodeType() string { return "Program" }

// ===== Types =====

type TypeKind int

const (
	TypeVoid TypeKind = iota
	TypeBool
	TypeInteger
)

func (t TypeKind) String() string {
	switch t {
	case TypeVoid:
		return "void"
	case TypeBool:
		return "bool"
	case TypeInteger:
		return "integer"
	default:
		return "unknown"
	}
}

// A simple wrapper node for a type if you want a Node for types.
type TypeNode struct {
	Kind TypeKind
}

func (t *TypeNode) NodeType() string { return "Type" }

// ===== Identifiers =====

type Identifier string

func (id Identifier) NodeType() string { return "Identifier" }
func (id Identifier) String() string   { return string(id) }

// ===== Declarations & Parameters =====

// VarDecl corresponds to `declaration_statement` in your grammar:
//
//	<type> <identifier> = <expression> ;
type VarDecl struct {
	Type  *TypeNode
	Name  Identifier
	Value Expr
}

func (d *VarDecl) NodeType() string { return "VarDecl" }

// Parameter corresponds to `parameter` (type + identifier)
type Parameter struct {
	Type *TypeNode
	Name Identifier
}

func (p *Parameter) NodeType() string { return "Parameter" }

// ===== Method Declarations =====
//
// method_declaration_statement:
//   <type_or_void> <identifier> "(" commaSeparatedOptional(parameter) ")" ( block | "extern" ";" )

type MethodDecl struct {
	Return *TypeNode // pointer so we can represent void (TypeVoid) or nil if desired
	Name   Identifier
	Params []*Parameter
	Body   *Block // nil if extern or if you want to represent "extern" via Extern=true
	Extern bool
}

func (m *MethodDecl) NodeType() string { return "MethodDecl" }

// ===== Statements =====

type Stmt interface {
	Node
	isStmt()
}

type Block struct {
	Declarations []*VarDecl // declarations local to the block (corresponds to repeat(field("declaration", ...)))
	Stmts        []Stmt
}

func (b *Block) NodeType() string { return "Block" }
func (b *Block) isStmt()          {}

type Assignment struct {
	Target Identifier // field("identifier", $.identifier)
	Value  Expr       // field("value", $._expression)
}

func (a *Assignment) NodeType() string { return "Assignment" }
func (a *Assignment) isStmt()          {}

type ExprStmt struct {
	Expr Expr // used for method_call followed by ';' or any expression statement
}

func (e *ExprStmt) NodeType() string { return "ExprStmt" }
func (e *ExprStmt) isStmt()          {}

// ReturnStmt corresponds to `return` optional expression + ';'
type ReturnStmt struct {
	Value Expr // nil if no value
}

func (r *ReturnStmt) NodeType() string { return "ReturnStmt" }
func (r *ReturnStmt) isStmt()          {}

type IfStmt struct {
	Cond Expr
	Then *Block
	Else *Block // nil if absent
}

func (i *IfStmt) NodeType() string { return "IfStmt" }
func (i *IfStmt) isStmt()          {}

type WhileStmt struct {
	Cond Expr
	Body *Block
}

func (w *WhileStmt) NodeType() string { return "WhileStmt" }
func (w *WhileStmt) isStmt()          {}

// VarDecl can be stored in Block.Declarations and top-level Program.Declarations.
// If you want a single AST node type for declaration statements (rather than a dedicated VarDecl),
// the above structure already models it directly.

// ===== Expressions =====

type Expr interface {
	Node
	isExpr()
}

type IntLiteral struct {
	Value int
}

func (n *IntLiteral) NodeType() string { return "IntLiteral" }
func (n *IntLiteral) isExpr()          {}

type BoolLiteral struct {
	Value bool
}

func (n *BoolLiteral) NodeType() string { return "BoolLiteral" }
func (n *BoolLiteral) isExpr()          {}

type IdentExpr struct {
	Name Identifier
}

func (n *IdentExpr) NodeType() string { return "IdentExpr" }
func (n *IdentExpr) isExpr()          {}

// Unary operator kinds (for '-' and '!')
type UnaryOp int

const (
	UnaryNeg UnaryOp = iota // "-"
	UnaryNot                // "!"
)

func (op UnaryOp) String() string {
	switch op {
	case UnaryNeg:
		return "-"
	case UnaryNot:
		return "!"
	default:
		return "unknown_unary"
	}
}

type UnaryExpr struct {
	Op   UnaryOp
	Expr Expr
}

func (n *UnaryExpr) NodeType() string { return "UnaryExpr" }
func (n *UnaryExpr) isExpr()          {}

// Binary operators (covers int ops, relational ops, boolean ops)
type BinOp int

const (
	// arithmetic
	BinAdd BinOp = iota
	BinSub
	BinMul
	BinDiv

	// relational
	BinEq
	BinLT
	BinGT

	// boolean
	BinAnd
	BinOr
)

func (op BinOp) String() string {
	switch op {
	case BinAdd:
		return "+"
	case BinSub:
		return "-"
	case BinMul:
		return "*"
	case BinDiv:
		return "/"
	case BinEq:
		return "=="
	case BinLT:
		return "<"
	case BinGT:
		return ">"
	case BinAnd:
		return "&&"
	case BinOr:
		return "||"
	default:
		return "unknown_binop"
	}
}

type BinaryExpr struct {
	Left  Expr
	Op    BinOp
	Right Expr
}

func (n *BinaryExpr) NodeType() string { return "BinaryExpr" }
func (n *BinaryExpr) isExpr()          {}

// CallExpr / Method call: identifier "(" args... ")"
type CallExpr struct {
	Callee Identifier
	Args   []Expr
}

func (n *CallExpr) NodeType() string { return "CallExpr" }
func (n *CallExpr) isExpr()          {}

// Parenthesized expression (explicit in grammar as "(" _expression ")")
type ParenExpr struct {
	Inner Expr
}

func (n *ParenExpr) NodeType() string { return "ParenExpr" }
func (n *ParenExpr) isExpr()          {}

// ===== Helpers (optional) =====

// Convenience constructors (not required but often handy)
func NewIntLit(v int) *IntLiteral     { return &IntLiteral{Value: v} }
func NewBoolLit(v bool) *BoolLiteral  { return &BoolLiteral{Value: v} }
func NewIdent(name string) *IdentExpr { return &IdentExpr{Name: Identifier(name)} }

func (p *Program) String() string {
	s := "program {\n"
	for _, d := range p.Declarations {
		s += "  var " + d.Type.Kind.String() + " " + string(d.Name) + " = <expr>\n"
	}
	for _, m := range p.Methods {
		ret := "void"
		if m.Return != nil {
			ret = m.Return.Kind.String()
		}
		params := ""
		for i, pr := range m.Params {
			if i > 0 {
				params += ", "
			}
			params += pr.Type.Kind.String() + " " + string(pr.Name)
		}
		body := "{ ... }"
		if m.Extern {
			body = "extern;"
		}
		s += "  " + ret + " " + string(m.Name) + "(" + params + ") " + body + "\n"
	}
	s += "}\n"
	return s
}

// simple debug helpers
func (i *IntLiteral) GoString() string  { return strconv.Itoa(i.Value) }
func (b *BoolLiteral) GoString() string { return strconv.FormatBool(b.Value) }
func (id *IdentExpr) GoString() string  { return string(id.Name) }
func (p *ParenExpr) GoString() string   { return "(" + p.Inner.NodeType() + ")" }
func (c *CallExpr) GoString() string    { return string(c.Callee) + "(...)" }
func (b *BinaryExpr) GoString() string {
	return "(" + b.Left.NodeType() + " " + b.Op.String() + " " + b.Right.NodeType() + ")"
}
