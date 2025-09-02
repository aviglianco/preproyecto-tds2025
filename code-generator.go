package main

import (
	"fmt"
	"strings"
)

type VarInfo struct {
	Type   Type
	Offset int
}

type CodeGenErrorKind int

const (
	ErrDuplicateDecl CodeGenErrorKind = iota
	ErrUseBeforeDeclare
	ErrTypeMismatch
	ErrUnknownStmt
	ErrUnknownExpr
	ErrNoRegisters
)

type CodeGenError struct {
	Line int
	Kind CodeGenErrorKind
	Msg  string
	Name string
	Have Type
	Want Type
}

func FormatCodeGenError(e CodeGenError) string {
	base := ""

	switch e.Kind {
	case ErrDuplicateDecl:
		base = fmt.Sprintf("duplicate declaration of '%s'", e.Name)
	case ErrUseBeforeDeclare:
		base = fmt.Sprintf("use of undeclared variable '%s'", e.Name)
	case ErrTypeMismatch:
		base = fmt.Sprintf("type mismatch assigning to '%s': assigned %s, declared %s", e.Name, typeString(e.Have), typeString(e.Want))
	case ErrUnknownStmt:
		base = "unknown statement kind"
	case ErrUnknownExpr:
		base = "unknown expression kind"
	case ErrNoRegisters:
		base = "register allocation failed: no free registers"
	default:
		base = e.Msg
	}

	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, base)
	}

	return base
}

func (e CodeGenError) Error() string {
	return FormatCodeGenError(e)
}

type CodeGenDiagnostics struct {
	Errors []CodeGenError
}

func (d *CodeGenDiagnostics) add(err CodeGenError) { d.Errors = append(d.Errors, err) }
func (d *CodeGenDiagnostics) HasErrors() bool      { return len(d.Errors) > 0 }

func generateAssembly(p *Program) string {
	asm, diags := generateAssemblyWithDiagnostics(p)

	if !diags.HasErrors() {
		return asm
	}

	b := &strings.Builder{}
	b.WriteString("; ---- codegen diagnostics ----\n")

	for _, e := range diags.Errors {
		fmt.Fprintf(b, "; %s\n", e.Error())
	}
	return b.String()
}

func generateAssemblyWithDiagnostics(p *Program) (string, CodeGenDiagnostics) {
	var diags CodeGenDiagnostics

	if p == nil || p.Main == nil || p.Main.Body == nil {
		return "; <empty program>\n", diags
	}

	// Pass 1: build symtab, compute frame size, validate types/usages; no emission
	pass1 := newCodeGen(nil, &diags, true)
	pass1.generateBlock(p.Main.Body)
	if diags.HasErrors() {
		return "", diags
	}
	frame := pass1.nextOffset

	// Pass 2: actual emission using the same semantics (offsets recomputed deterministically)
	b := &strings.Builder{}
	pass2 := newCodeGen(b, &diags, false)

	pass2.emit(".text")
	pass2.emit(".global main")
	pass2.emit("main:")
	pass2.emit("\t; prologue")
	pass2.emit("\tPUSH BP")
	pass2.emit("\tMOV BP, SP")
	if frame > 0 {
		pass2.emit(fmt.Sprintf("\tSUB SP, %d", frame))
	}

	pass2.generateBlock(p.Main.Body)
	if diags.HasErrors() {
		return "", diags
	}

	return b.String(), diags
}

type codeGen struct {
	b           *strings.Builder
	symtab      map[string]VarInfo
	nextOffset  int
	slotSize    int
	diagnostics *CodeGenDiagnostics
	abort       bool
	dryRun      bool
	regs        []string
	free        []string
}

func newCodeGen(b *strings.Builder, diags *CodeGenDiagnostics, dry bool) *codeGen {
	g := &codeGen{
		b:           b,
		symtab:      map[string]VarInfo{},
		nextOffset:  0,
		slotSize:    8,
		diagnostics: diags,
		abort:       false,
		dryRun:      dry,
		regs:        []string{"R0", "R1", "R2", "R3"},
	}
	g.free = append(g.free, g.regs...)
	return g
}

func (g *codeGen) emit(line string) {
	if g.dryRun || g.abort || g.b == nil {
		return
	}
	fmt.Fprintln(g.b, line)
}

func (g *codeGen) addErr(err CodeGenError) {
	g.diagnostics.add(err)
	g.abort = true
}

func (g *codeGen) errDuplicateDecl(line int, name string) {
	g.addErr(CodeGenError{Line: line, Kind: ErrDuplicateDecl, Name: name})
}

func (g *codeGen) errUseBefore(line int, name string) {
	g.addErr(CodeGenError{Line: line, Kind: ErrUseBeforeDeclare, Name: name})
}

func (g *codeGen) errTypeMismatch(line int, name string, have, want Type) {
	g.addErr(CodeGenError{Line: line, Kind: ErrTypeMismatch, Name: name, Have: have, Want: want})
}

func (g *codeGen) errUnknownStmt() { g.addErr(CodeGenError{Kind: ErrUnknownStmt}) }
func (g *codeGen) errUnknownExpr() { g.addErr(CodeGenError{Kind: ErrUnknownExpr}) }
func (g *codeGen) errNoRegisters() { g.addErr(CodeGenError{Kind: ErrNoRegisters}) }

func (g *codeGen) allocVar(name string, t Type, line int) VarInfo {
	// Detect duplicate declarations
	if _, exists := g.symtab[name]; exists {
		g.errDuplicateDecl(line, name)
	}

	if g.abort {
		return VarInfo{Type: t, Offset: 0}
	}

	g.nextOffset += g.slotSize
	info := VarInfo{Type: t, Offset: g.nextOffset}
	g.symtab[name] = info
	return info
}

func (g *codeGen) allocReg() (string, bool) {
	if len(g.free) == 0 {
		g.errNoRegisters()
		return "", false
	}

	r := g.free[len(g.free)-1]
	g.free = g.free[:len(g.free)-1]
	return r, true
}

func (g *codeGen) freeReg(r string) {
	if r == "" {
		return
	}
	g.free = append(g.free, r)
}

func (g *codeGen) generateBlock(blk *Block) {
	if blk == nil || g.abort {
		return
	}

	for _, s := range blk.Statements {
		if g.abort {
			return
		}
		g.generateStmt(s)
	}
}

func (g *codeGen) generateStmt(s Stmt) {
	if g.abort {
		return
	}
	switch st := s.(type) {
	case *Decl:
		_ = g.allocVar(st.Name, st.VarType, st.Line)
	case *Assign:
		info, ok := g.symtab[st.Name]
		if !ok {
			g.errUseBefore(st.Line, st.Name)
			return
		}
		valType, ok := g.typeOfExpr(st.Value)
		if !ok {
			return
		}
		if valType != info.Type {
			g.errTypeMismatch(st.Line, st.Name, valType, info.Type)
			return
		}
		reg, ok := g.evalExprToReg(st.Value)
		if !ok {
			return
		}
		g.emit(fmt.Sprintf("\tMOV [BP-%d], %s", info.Offset, reg))
		g.freeReg(reg)
	case *Return:
		if st.Value != nil {
			reg, ok := g.evalExprToReg(st.Value)
			if !ok {
				return
			}

			// Use R0 as return register
			if reg != "R0" {
				g.emit(fmt.Sprintf("\tMOV R0, %s", reg))
			}
			
			g.freeReg(reg)
		}
		g.emit("\t; epilogue")
		g.emit("\tMOV SP, BP")
		g.emit("\tPOP BP")
		g.emit("\tRET")
	case *Skip:
		g.emit("\t; skip")
	default:
		g.errUnknownStmt()
	}
}

func (g *codeGen) evalExprToReg(e Expr) (string, bool) {
	if g.abort {
		return "", false
	}

	switch ex := e.(type) {
	case *Identifier:
		info, ok := g.symtab[ex.Name]
		if !ok {
			g.errUseBefore(ex.Line, ex.Name)
			return "", false
		}

		r, ok := g.allocReg()
		if !ok {
			return "", false
		}

		g.emit(fmt.Sprintf("\tMOV %s, [BP-%d] ; %s", r, info.Offset, ex.Name))
		return r, true
	case *IntLiteral:
		r, ok := g.allocReg()
		if !ok {
			return "", false
		}

		g.emit(fmt.Sprintf("\tMOV %s, %d", r, ex.Value))
		return r, true
	case *BoolLiteral:
		r, ok := g.allocReg()
		if !ok {
			return "", false
		}

		val := 0
		if ex.Value {
			val = 1
		}

		g.emit(fmt.Sprintf("\tMOV %s, %d", r, val))
		return r, true
	case *BinaryExpr:
		leftReg, ok := g.evalExprToReg(ex.Left)
		if !ok {
			return "", false
		}

		rightReg, ok := g.evalExprToReg(ex.Right)
		if !ok {
			g.freeReg(leftReg)
			return "", false
		}

		mn := opMnemonic(ex.Kind)
		if mn == "NOP" {
			g.errUnknownExpr()
			g.freeReg(leftReg)
			g.freeReg(rightReg)
			return "", false
		}

		g.emit(fmt.Sprintf("\t%s %s, %s", mn, leftReg, rightReg))
		g.freeReg(rightReg)
		return leftReg, true
	default:
		g.errUnknownExpr()
		return "", false
	}
}

func (g *codeGen) typeOfExpr(e Expr) (Type, bool) {
	switch ex := e.(type) {
	case *Identifier:
		if info, ok := g.symtab[ex.Name]; ok {
			return info.Type, true
		}
		g.errUseBefore(ex.Line, ex.Name)
		return TypeInt, false
	case *IntLiteral:
		return TypeInt, true
	case *BoolLiteral:
		return TypeBool, true
	case *BinaryExpr:
		// Propagate unknowns if either side cannot be resolved
		_, lok := g.typeOfExpr(ex.Left)
		_, rok := g.typeOfExpr(ex.Right)
		return TypeInt, lok && rok
	default:
		g.errUnknownExpr()
		return TypeInt, false
	}
}

func opMnemonic(k BinaryOpKind) string {
	switch k {
	case OpMul:
		return "MUL"
	case OpDiv:
		return "DIV"
	case OpAdd:
		return "ADD"
	case OpSub:
		return "SUB"
	default:
		return "NOP"
	}
}
