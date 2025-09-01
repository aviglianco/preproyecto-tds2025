package main

type Node interface{ isNode() }
type Stmt interface {
	Node
	isStmt()
}
type Expr interface {
	Node
	isExpr()
}

type Program struct {
	ReturnType Type
	Main       *Main
}

func (*Program) isNode() {}

type Main struct {
	Args []string
	Body *Block
}

func (*Main) isNode() {}

type Block struct {
	Statements []Stmt
}

func (*Block) isNode() {}

type Type int

const (
	TypeInt Type = iota
	TypeBool
	TypeVoid
)

type Decl struct {
	VarType Type
	Name    string
}

func (*Decl) isNode() {}
func (*Decl) isStmt() {}

type Assign struct {
	Name  string
	Value Expr
}

func (*Assign) isNode() {}
func (*Assign) isStmt() {}

type Return struct {
	Value Expr // may be nil for void return
}

func (*Return) isNode() {}
func (*Return) isStmt() {}

type Skip struct{}

func (*Skip) isNode() {}
func (*Skip) isStmt() {}

type Identifier struct{ Name string }

func (*Identifier) isNode() {}
func (*Identifier) isExpr() {}

type IntLiteral struct{ Value string }

func (*IntLiteral) isNode() {}
func (*IntLiteral) isExpr() {}

type BoolLiteral struct{ Value bool }

func (*BoolLiteral) isNode() {}
func (*BoolLiteral) isExpr() {}

type BinaryOpKind int

const (
	OpMul BinaryOpKind = iota
	OpDiv
	OpAdd
	OpSub
)

type BinaryExpr struct {
	Kind  BinaryOpKind
	Left  Expr
	Right Expr
}

func (*BinaryExpr) isNode() {}
func (*BinaryExpr) isExpr() {}
