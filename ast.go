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
	Line int
}

func (*Main) isNode() {}

type Block struct {
	Statements []Stmt
	Line       int
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
	Line    int
}

func (*Decl) isNode() {}
func (*Decl) isStmt() {}

type Assign struct {
	Name  string
	Value Expr
	Line  int
}

func (*Assign) isNode() {}
func (*Assign) isStmt() {}

type Return struct {
	Value Expr // may be nil for void return
	Line  int
}

func (*Return) isNode() {}
func (*Return) isStmt() {}

type Skip struct{}

func (*Skip) isNode() {}
func (*Skip) isStmt() {}

type Identifier struct {
	Name string
	Line int
}

func (*Identifier) isNode() {}
func (*Identifier) isExpr() {}

type IntLiteral struct {
	Value int
	Line  int
}

func (*IntLiteral) isNode() {}
func (*IntLiteral) isExpr() {}

type BoolLiteral struct {
	Value bool
	Line  int
}

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
	Line  int
}

func (*BinaryExpr) isNode() {}
func (*BinaryExpr) isExpr() {}
