package main

import (
	"errors"
	"fmt"
)

type Val struct {
	// TODO: cambiar a pascal case
	boolVal *bool
	intVal  *int
	varType Type
}

type ReferenceTable map[string]Val

func interpretExpression(referenceTable *ReferenceTable, exp Expr) (Val, error) {

	switch ex := exp.(type) {
	case *BoolLiteral:
		return Val{boolVal: &ex.Value}, nil
	case *IntLiteral:
		return Val{intVal: &ex.Value}, nil
	case *Identifier:
		return (*referenceTable)[ex.Name], nil
	case *BinaryExpr:

		// TODO: error handling
		leftVal, _ := interpretExpression(referenceTable, ex.Left)
		rightVal, _ := interpretExpression(referenceTable, ex.Right)

		var val Val

		// for now!
		val.varType = TypeInt

		var res int

		switch ex.Kind {
		case OpSub:
			res = *leftVal.intVal - *rightVal.intVal
		case OpAdd:
			res = *leftVal.intVal + *rightVal.intVal
		case OpDiv:
			res = *leftVal.intVal / *rightVal.intVal
		case OpMul:
			res = *leftVal.intVal * *rightVal.intVal

		}
		val.intVal = &res
		return val, nil
	default:
		fmt.Println(exp)

	}

	res := int(0)
	return Val{intVal: &res, varType: TypeInt}, errors.New("Couldn't evaluate expression, probably new op needs implementation.")
}

func updateReferenceTable(referenceTable *ReferenceTable, statement Stmt) error {

	switch t := statement.(type) {
	case *Decl:
		(*referenceTable)[t.Name] = Val{
			varType: t.VarType,
			intVal:  nil,
			boolVal: nil,
		}
	case *Assign:
		_, ok := (*referenceTable)[t.Name]
		if !ok {
			return fmt.Errorf("Variable %s not set", t.Name)
		}

		// TODO: interpret expressions and assign accordingly
		val := (*referenceTable)[t.Name]
		switch val.varType {
		case TypeInt:
			newVal, err := interpretExpression(referenceTable, t.Value)
			if err != nil {
				val.intVal = newVal.intVal
			} else {
				return fmt.Errorf("Couldn't run statment: %w", err)
			}
		case TypeBool:
			newVal, err := interpretExpression(referenceTable, t.Value)
			if err != nil {
				val.boolVal = newVal.boolVal
			} else {
				return fmt.Errorf("Couldn't run statment: %w", err)
			}
		}

	}

	return nil
}

func interpret(program Program) {

	referenceTable := make(ReferenceTable)

	for _, statement := range program.Main.Body.Statements {
		switch t := statement.(type) {
		case *Return:
			fmt.Print("Return value: ")
			ret, _ := interpretExpression(&referenceTable, t.Value)
			switch ret.varType {
			case TypeBool:
				fmt.Print(*ret.boolVal)
			case TypeInt:
				fmt.Print(*ret.intVal)
			}
		default:
			updateReferenceTable(&referenceTable, statement)

		}

	}
}
