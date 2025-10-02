package main

type Env struct {
	Table Table
	Prev  *Env
}

type Symbol struct {
	Type  TypeKind
	isVar bool
}

type Table map[Identifier]Symbol

func (env Env) Lookup(name Identifier) (Symbol, bool) {
	var currentEnv Env = env
	for {

		// lookup in current frame
		symbol, ok := currentEnv.Table[name]
		if ok {
			return symbol, true
		} else if currentEnv.Prev != nil {
			// lookup in previous frame
			currentEnv = *currentEnv.Prev
		} else {
			break
		}
	}
	return Symbol{0, true}, false
}

func (env Env) Insert(name Identifier, symbol Symbol) {
	env.Table[name] = symbol
}
