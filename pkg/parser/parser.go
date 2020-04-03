package parser

import (
	i "group-11/pkg/inputdata"
	cg "group-11/pkg/codegen"
	s "group-11/pkg/scanner"
	k "group-11/pkg/keywords"
	st "group-11/pkg/symtable"
	"strconv"
)

var FIRSTFACTOR = map[int]int{k.IDENT:1, k.NUMBER:1, k.LPAREN:1, k.NOT:1}
var FOLLOWFACTOR = map[int]int{k.TIMES:1, k.DIV:1, k.MOD:1, k.AND:1, k.OR:1, k.PLUS:1, k.MINUS:1, 
					k.EQ:1, k.NE:1, k.LT:1, k.LE:1, k.GT:1, k.GE:1, k.COMMA:1, k.SEMICOLON:1, k.THEN:1, 
					k.ELSE:1, k.RPAREN:1, k.RBRAK:1, k.DO:1, k.PERIOD:1, k.END:1}
var FIRSTEXPRESSION = map[int]int{k.PLUS:1, k.MINUS:1, k.IDENT:1, k.NUMBER:1, k.LPAREN:1, k.NOT:1}
var FIRSTSTATEMENT = map[int]int{k.IDENT:1, k.IF:1, k.WHILE:1, k.BEGIN:1}
var FOLLOWSTATEMENT = map[int]int{k.SEMICOLON:1, k.END:1, k.ELSE:1}
var FIRSTTYPE = map[int]int{k.IDENT:1, k.RECORD:1, k.ARRAY:1, k.LPAREN:1}
var FOLLOWTYPE = map[int]int{k.SEMICOLON:1}
var FIRSTDECL = map[int]int{k.CONST:1, k.TYPE:1, k.VAR:1, k.PROCEDURE:1}
var FOLLOWDECL = map[int]int{k.BEGIN:1}
var FOLLOWPROCCALL = map[int]int{k.SEMICOLON:1, k.END:1, k.ELSE:1}
var STRONGSYMS = map[int]int{k.CONST:1, k.TYPE:1, k.VAR:1, k.PROCEDURE:1, k.WHILE:1, k.IF:1, k.BEGIN:1, k.EOF:1}

// Helper function to check for that an element is in the first and follow sets.
func exists(a int, dict map[int]int) bool {
	if _, ok := dict[a]; ok {
		return true
	}

	return false
}

// Generates selectors for records and arrays.
func selector(x *st.SymTableEntry, inputData *i.InputData) *st.SymTableEntry {
	for inputData.Sym == k.PERIOD || inputData.Sym == k.LBRAK {
		if inputData.Sym == k.PERIOD {
			s.GetSym(inputData)
			if inputData.Sym == k.IDENT {
				if x.ArrOrRec == "record" {
					ctr := 0
					for _, recfield := range x.Ctp.Fields {
						if recfield.Name == inputData.Val {
							x = cg.GenSelect(x, &recfield, inputData)
							break
						}
						ctr += 1
					}
					if ctr == len(x.Ctp.Fields)-1 {
						s.PrintError(inputData, "not a field")
					}
					s.GetSym(inputData)
				} else {
					s.PrintError(inputData, "not a record")
				}
			} else {
				s.PrintError(inputData,"identifier expected")
			}
		} else { // x[y]
			s.GetSym(inputData)
			y := expression(inputData)
			if x.ArrOrRec == "array" {
				if y .Tp == st.Int {
					if y.EntryType == "const" {

					} else {
						x = cg.GenIndex(x, y, inputData)
					}
				}
			} else {
				s.PrintError(inputData, "not an array")
			}
			if inputData.Sym == k.RBRAK {
				s.GetSym(inputData)
			} else {
				s.PrintError(inputData, "] expected")
			}
		}
	}
	return x
}

// Generates factors.
func factor(inputData *i.InputData) *st.SymTableEntry {
	var x = &st.SymTableEntry{}
	if !exists(inputData.Sym, FIRSTFACTOR) {
		s.PrintError(inputData,"expression expected")
		for !(exists(inputData.Sym, FIRSTFACTOR) || exists(inputData.Sym, FOLLOWFACTOR) || exists(inputData.Sym, STRONGSYMS)) {
			s.GetSym(inputData)
		}
	}
	if inputData.Sym == k.IDENT {
		y := st.FindInSymTab(inputData, inputData.Val)
		x = y
		if x.EntryType == "var" || x.EntryType == "ref" {
			x = cg.GenVar(x, inputData)
			s.GetSym(inputData)
		} else if x.EntryType == "const" {
			x = st.Const(x .Tp, x.Val)
			x = cg.GenConst(x)
			s.GetSym(inputData)
		} else {
			s.PrintError(inputData,"expression expected")
		}
		x = selector(x, inputData)
	} else if inputData.Sym == k.NUMBER {
		constVal, err := strconv.Atoi(inputData.Val)
		x = st.Const(st.Int, constVal)
		x = cg.GenConst(x)
		if err != nil {
			s.PrintError(inputData, "error converting number")
		}
		s.GetSym(inputData)
	} else if inputData.Sym == k.LPAREN {
		s.GetSym(inputData)
		x = expression(inputData)
		if inputData.Sym == k.RPAREN {
			s.GetSym(inputData)
		} else {
			s.PrintError(inputData,") expected")
		}
	} else if inputData.Sym == k.NOT {
		s.GetSym(inputData)
		x = factor(inputData)
		if x .Tp != st.Bool {
			s.PrintError(inputData,"not boolean")
		} else if x.EntryType == "const" {
			x.Val = 1 - x.Val
		} else {
			x = cg.GenUnaryOp(k.NOT, x, inputData)
		}
	} else {
		x = st.Const(st.None, 0)
	}
	return x
}

// Generates terms.
func term(inputData *i.InputData) *st.SymTableEntry {
	x := factor(inputData)
	for inputData.Sym == k.TIMES || inputData.Sym == k.DIV || inputData.Sym == k.MOD || inputData.Sym == k.AND {
		op := inputData.Sym
		s.GetSym(inputData)
		if op == k.AND && x.EntryType != "const" {
			x = cg.GenUnaryOp(k.AND, x, inputData)
		}
		y := factor(inputData)
		if (x .Tp == st.Int && y .Tp == st.Int) && exists(op, map[int]int{k.TIMES:1, k.DIV:1, k.MOD:1}) {
			if x.EntryType == "const" && y.EntryType == "const" {
				if op == k.TIMES {
					x.Val = x.Val * y.Val
				} else if op == k.DIV {
					x.Val = x.Val / y.Val
				} else if op == k.MOD {
					x.Val = x.Val % y.Val
				}
			} else {
				x = cg.GenBinaryOp(op, x, y, inputData)
			}
		} else if (x .Tp == st.Bool && y .Tp == st.Bool) && op == k.AND {
			if x.EntryType == "const" {
				if x.Val != st.EmptyInt { // Since Go doesn't provide a good way to check for empty int, I used a massively negative number
					x = y
				}
			} else {
				x = cg.GenBinaryOp(k.AND, x, y, inputData)
			}
		} else {
			s.PrintError(inputData, "bad type")
		}
	}
	return x
}

// Generates simple expressions.
func simpleExpression(inputData *i.InputData) *st.SymTableEntry {
	x := &st.SymTableEntry{}
	if inputData.Sym == k.PLUS {
		s.GetSym(inputData)
		x = term(inputData)
	} else if inputData.Sym == k.MINUS {
		s.GetSym(inputData)
		x = term(inputData)
		if x .Tp != st.Int {
			s.PrintError(inputData,"bad type")
		} else if x.EntryType == "const" {
			x.Val = -1 * x.Val
		} else {
			x = cg.GenUnaryOp(k.MINUS, x, inputData)
		}
	} else {
		x = term(inputData)
	}
	for inputData.Sym == k.PLUS || inputData.Sym == k.MINUS || inputData.Sym == k.OR {
		op := inputData.Sym
		s.GetSym(inputData)
		if op == k.OR && x.EntryType != "const" {
			x = cg.GenUnaryOp(k.OR, x, inputData)
		}
		y := term(inputData)
		if (x .Tp == st.Int && y .Tp == st.Int) && (op == k.PLUS || op == k.MINUS) {
			if x.EntryType == "const" && y.EntryType == "const" {
				if op == k.PLUS {
					x.Val = x.Val + y.Val
				} else if op == k.MINUS {
					x.Val = x.Val - y.Val
				}
			} else {
				x = cg.GenBinaryOp(op, x, y, inputData)
			}
		} else if x .Tp == st.Bool && y .Tp == st.Bool && op == k.OR {
			if x.EntryType == "const" {
				if x.Val != st.EmptyInt {
					x = y
				} else {
					x = cg.GenBinaryOp(k.OR, x, y, inputData)
				}
			}
		} else {
			s.PrintError(inputData, "Bad type")
		}
	}
	return x
}

// Generates whole expressions.
// Generates whole expressions.
func expression(inputData *i.InputData) *st.SymTableEntry {
	x := simpleExpression(inputData)
	for inputData.Sym == k.EQ || inputData.Sym == k.NE || inputData.Sym == k.LT || inputData.Sym == k.LE || inputData.Sym == k.GT || inputData.Sym == k.GE {
		op := inputData.Sym
		s.GetSym(inputData)
		y := simpleExpression(inputData)

		if x.Tp == y.Tp {
			if x.EntryType == "const" && y.EntryType == "const" {
				if op == k.EQ {
					if x.Val == y.Val {
						x.Val = 1
					} else {
						x.Val = 0
					}
				} else if op == k.NE {
					if x.Val != y.Val {
						x.Val = 1
					} else {
						x.Val = 0
					}
				} else if op == k.LT {
					if x.Val < y.Val {
						x.Val = 1
					} else {
						x.Val = 0
					}
				} else if op == k.LE {
					if x.Val <= y.Val {
						x.Val = 1
					} else {
						x.Val = 0
					}
				} else if op == k.GT {
					if x.Val > y.Val {
						x.Val = 1
					} else {
						x.Val = 0
					}
				} else if op == k.GE {
					if x.Val >= y.Val {
						x.Val = 1
					} else {
						x.Val = 0
					}
				}
				x.Tp = st.Bool
			} else {
				x = cg.GenRelation(op, x, y, inputData)
			}
		} else {
			s.PrintError(inputData, "bad type")
		}
	}

	return x
}

// Generates compound statements.
func compoundStatement(inputData *i.InputData) *st.SymTableEntry {
	if inputData.Sym == k.BEGIN {
		s.GetSym(inputData)
	} else {
		s.PrintError(inputData, "'begin' expected")
	}
	x := statement(inputData)
	for inputData.Sym == k.SEMICOLON || exists(inputData.Sym, FIRSTSTATEMENT) {
		if inputData.Sym == k.SEMICOLON {
			s.GetSym(inputData)
		} else {
			s.PrintError(inputData, "; missing")
		}
		y := statement(inputData)
		cg.GenSeq(x, y, inputData)
	}
	if inputData.Sym == k.END {
		s.GetSym(inputData)
	} else {
		s.PrintError(inputData, "'end' expected")
	}
	return x
}

// Generates statements.
func statement(inputData *i.InputData) *st.SymTableEntry {
	x := &st.SymTableEntry{}
	y := &st.SymTableEntry{}
	if !exists(inputData.Sym, FIRSTSTATEMENT) {
		s.PrintError(inputData, "statement expected")
		s.GetSym(inputData)
		for !(exists(inputData.Sym, FIRSTFACTOR) || exists(inputData.Sym, FOLLOWSTATEMENT) || exists(inputData.Sym, STRONGSYMS)){
			s.GetSym(inputData)
		}
	}
	if inputData.Sym == k.IDENT {
		x = st.FindInSymTab(inputData, inputData.Val)
		s.GetSym(inputData)
		if x.EntryType == "var" || x.EntryType == "ref" {
			x = cg.GenVar(x, inputData)
			x = selector(x, inputData)
			if inputData.Sym == k.BECOMES {
				s.GetSym(inputData)
				y = expression(inputData)
				if x .Tp == st.Bool || x .Tp == st.Int || y .Tp == st.Bool || y .Tp == st.Int {
					cg.GenAssign(x, y, inputData)
				} else {
					s.PrintError(inputData, "incompatible assignment")
				}
			} else if inputData.Sym == k.EQ {
				s.PrintError(inputData,":= expected")
				s.GetSym(inputData)
				y = expression(inputData)
				cg.GenSeq(y, y, inputData) // THIS IS ONLY CALk.LED BECAUSE GO k.DOESN'T LIKE UNUSED DECLARATIONS
			} else {
				s.PrintError(inputData, ":= expected")
			}
		} else if x.EntryType == "proc" || x.EntryType == "stdproc" {
			fp := x.Par
			var ap []*st.SymTableEntry
			i := 0
			if inputData.Sym == k.LPAREN {
				s.GetSym(inputData)
				if exists(inputData.Sym, FIRSTEXPRESSION) {
					y = expression(inputData)
					if i < len(fp) {
						if (y.EntryType == "var" || fp[i].EntryType == "var") && (fp[i].EntryType == y.EntryType) {
							if x.EntryType == "proc" {
								ap = append(ap, cg.GenActualPara(y, &fp[i], inputData))
							}
						} else if x.Name != "read" {
							s.PrintError(inputData, "illegal parameter mode")
						}
					} else {
						s.PrintError(inputData, "extra parameter")
					}
					i++
					for inputData.Sym == k.COMMA {
						s.GetSym(inputData)
						y = expression(inputData)
						if i < len(fp) {
							if (y.EntryType == "var" || fp[i].EntryType == "var") && (fp[i].EntryType == y.EntryType) {
								if x.EntryType == "proc" {
									ap = append(ap, cg.GenActualPara(y, &fp[i], inputData))
								}
							} else {
								s.PrintError(inputData, "illegal parameter mode")
							}
						} else {
							s.PrintError(inputData, "extra parameter")
						}
						i++
					}
				}
				if inputData.Sym == k.RPAREN {
					s.GetSym(inputData)
				} else {
					s.PrintError(inputData, ") expected")
				}
			}
			if i < len(fp) {
				s.PrintError(inputData, "too few parameters")
			} else if x.EntryType == "stdproc" {
				if x.Name == "read" {
					cg.GenRead(y, inputData)
				} else if x.Name == "write" {
					cg.GenWrite(y, inputData)
				} else if x.Name == "writeln" {
					cg.GenWriteln(inputData)
				}
			} else {
				x = cg.GenCall(x, inputData)
			}
		} else {
			s.PrintError(inputData, "variable or procedure expected")
		}
	} else if inputData.Sym == k.BEGIN {
		x = compoundStatement(inputData)
	} else if inputData.Sym == k.IF {
		s.GetSym(inputData)
		x := expression(inputData)
		if x .Tp == st.Bool {
			x = cg.GenThen(x, inputData)
		} else {
			s.PrintError(inputData, "boolean expected")
		}
		if inputData.Sym == k.THEN {
			s.GetSym(inputData)
		} else {
			s.PrintError(inputData,"'then' expected")
		}
		y = statement(inputData)
		if inputData.Sym == k.ELSE {
			if x .Tp == st.Bool {
				y = cg.GenElse(x, y, inputData)
			}
			s.GetSym(inputData)
			z:= statement(inputData)
			if x .Tp == st.Bool {
				x = cg.GenIfElse(x, y, z, inputData)
			}
		} else {
			if x .Tp == st.Bool {
				x = cg.GenIfThen(x, y, inputData)
			}
		}
	} else if inputData.Sym == k.WHILE {
		s.GetSym(inputData)
		cg.GenWhile(inputData)
		x = expression(inputData)
		if x .Tp == st.Bool {
			x = cg.GenDo(x, inputData)
		} else {
			s.PrintError(inputData, "boolean expected")
		}
		if inputData.Sym == k.DO {
			s.GetSym(inputData)
		} else {
			s.PrintError(inputData, "'do' expected")
		}
		y = statement(inputData)
		if x .Tp == st.Bool {
			cg.GenWhileDo(x, y, inputData)
		}
	} else {
		x = nil
	}
	return x
}

// Generates the type of an identifier.
func typ(inputData *i.InputData) *st.SymTableEntry {
	x := &st.SymTableEntry{}
	if !exists(inputData.Sym, FIRSTTYPE) {
		s.PrintError(inputData, "type expected")
		for !(exists(inputData.Sym, FIRSTFACTOR) || exists(inputData.Sym, FOLLOWFACTOR) || exists(inputData.Sym, STRONGSYMS)) {
			s.GetSym(inputData)
		}
	}
	if inputData.Sym == k.IDENT {
		ident := inputData.Val
		x = st.FindInSymTab(inputData, ident)
		s.GetSym(inputData)
		if x.EntryType == "type" {
			x = st.Type(x .Tp)
		} else {
			s.PrintError(inputData, "not a type")
			x = st.Type(st.None)
		}
	} else if inputData.Sym == k.ARRAY {
		s.GetSym(inputData)
		if inputData.Sym == k.LBRAK {
			s.GetSym(inputData)
		} else {
			s.PrintError(inputData, "'[' expected")
		}
		x = expression(inputData)
		if inputData.Sym == k.PERIOD {
			s.GetSym(inputData)
		} else {
			s.PrintError(inputData,"'.' expected")
		}
		if inputData.Sym == k.PERIOD {
			s.GetSym(inputData)
		} else {
			s.PrintError(inputData,"'.' expected")
		}
		y := expression(inputData)
		if inputData.Sym == k.RBRAK {
			s.GetSym(inputData)
		} else {
			s.PrintError(inputData,"']' expected")
		}
		if inputData.Sym == k.OF {
			s.GetSym(inputData)
		} else {
			s.PrintError(inputData,"of expected")
		}
		z := typ(inputData) .Tp
		if x.EntryType != "const" || x.Val < 0 {
			s.PrintError(inputData,"bad lower bound")
			x = st.Type(st.None)
		} else if y.EntryType != "const" || y.Val < 0 {
			s.PrintError(inputData,"bad upper bound")
			x = st.Type(st.None)
		} else {
			arr := st.Array(z, x.Val, y.Val - x.Val + 1)
			x = cg.GenArray(arr)
			x .Tp = z
		}
	} else if inputData.Sym == k.RECORD {
		s.GetSym(inputData)
		st.OpenScope(inputData)
		typedIds("var", inputData)
		for {
			if inputData.Sym == k.SEMICOLON {
				s.GetSym(inputData)
				typedIds("var", inputData)
			} else {
				break
			}
		}
		if inputData.Sym == k.END {
			s.GetSym(inputData)
		} else {
			s.PrintError(inputData,"'end' expected")
		}
		r := st.TopScope(inputData)
		st.CloseScope(inputData)
		rec := st.Record(r)
		x = cg.GenRec(rec)
		x .Tp = st.Nil
	}

	return x
}

// Helps generate typed identifiers.
func typedIds(entryType string, inputData *i.InputData) {
	tid := []string{}
	if inputData.Sym == k.IDENT {
		tid = append([]string{inputData.Val}, tid...)
		s.GetSym(inputData)
	} else {
		s.PrintError(inputData,"identifier expected")
	}

	for inputData.Sym == k.COMMA {
		s.GetSym(inputData)
		if inputData.Sym == k.IDENT {
			tid = append(tid, inputData.Val)
			s.GetSym(inputData)
		} else {
			s.PrintError(inputData, "identifier expected")
		}
	}

	if inputData.Sym == k.COLON {
		s.GetSym(inputData)
		tp := typ(inputData)
		if tp != (&st.SymTableEntry{}) {
			for i := 0; i < len(tid); i++ {
				if entryType == "var" {
					if tp.ArrOrRec == "array" || tp.ArrOrRec == "record" {
						tp.EntryType =  entryType
						st.NewDecl(tid[i], tp, inputData)
					} else {
						v := st.Var(tp .Tp)
						st.NewDecl(tid[i], v, inputData)
					}
				} else if entryType == "ref" {
					if tp.ArrOrRec == "array" || tp.ArrOrRec == "record" {
						tp.EntryType =  entryType
						st.NewDecl(tid[i], tp, inputData)
					} else {
						r := st.Ref(tp .Tp)
						st.NewDecl(tid[i], r, inputData)
					}
				}
			}
		}
	} else {
		s.PrintError(inputData,": expected")
	}
}

// Generates various declarations.
func declaration(allocVarLevel string, inputData *i.InputData) {
	if !(exists(inputData.Sym, FIRSTDECL) || exists(inputData.Sym, FOLLOWDECL)) {
		s.PrintError(inputData, "'begin' or declaration expected")
		for !(exists(inputData.Sym, FIRSTDECL) || exists(inputData.Sym, FOLLOWDECL) || exists(inputData.Sym, STRONGSYMS)) {
			s.GetSym(inputData)
		}
	}
	for inputData.Sym == k.CONST {
		s.GetSym(inputData)
		if inputData.Sym == k.IDENT {
			ident := inputData.Val
			s.GetSym(inputData)
			if inputData.Sym == k.EQ {
				s.GetSym(inputData)
			} else {
				s.PrintError(inputData,"= expected")
			}
			x := expression(inputData)
			if x.EntryType == "const" {
				c := st.Const(x .Tp, x.Val)
				st.NewDecl(ident, c, inputData)
			} else {
				s.PrintError(inputData,"expression not constant")
			}
		} else {
			s.PrintError(inputData,"constant name expected")
		}
		if inputData.Sym == k.SEMICOLON {
			s.GetSym(inputData)
		} else {
			s.PrintError(inputData, "; expected")
		}
	}
	for inputData.Sym == k.TYPE {
		s.GetSym(inputData)
		if inputData.Sym == k.IDENT {
			ident := inputData.Val
			s.GetSym(inputData)
			if inputData.Sym == k.EQ {
				s.GetSym(inputData)
			} else {
				s.PrintError(inputData,"= expected")
			}
			x := typ(inputData)
			st.NewDecl(ident, x, inputData)
			if inputData.Sym == k.SEMICOLON {
				s.GetSym(inputData)
			} else {
				s.PrintError(inputData, "; expected")
			}
		} else {
			s.PrintError(inputData,"type name expected")
		}
	}
	start := len(st.TopScope(inputData))
	for inputData.Sym == k.VAR {
		s.GetSym(inputData)
		typedIds("var", inputData)
		if inputData.Sym == k.SEMICOLON {
			s.GetSym(inputData)
		} else {
			s.PrintError(inputData,"; expected")
		}
	}
	if allocVarLevel == "global" {
		cg.GenGlobalVars(st.TopScope(inputData), start, inputData)
	} else {
		cg.GenLocalVars(st.TopScope(inputData), start, inputData)
	}
	for inputData.Sym == k.PROCEDURE {
		s.GetSym(inputData)
		if inputData.Sym == k.IDENT {
			s.GetSym(inputData)
		} else {
			s.PrintError(inputData,"procedure named expected")
		}
		ident := inputData.Val
		st.NewDecl(ident, st.Proc([]st.SymTableEntry{}), inputData)
		sc := st.TopScope(inputData)
		st.OpenScope(inputData)
		if inputData.Sym == k.LPAREN {
			s.GetSym(inputData)
			if inputData.Sym == k.VAR || inputData.Sym == k.IDENT {
				if inputData.Sym == k.VAR {
					s.GetSym(inputData)
					typedIds("ref", inputData)
				} else {
					typedIds("var", inputData)
				}
				for inputData.Sym == k.SEMICOLON {
					s.GetSym(inputData)
					if inputData.Sym == k.VAR {
						s.GetSym(inputData)
						typedIds("ref", inputData)
					} else {
						typedIds("var", inputData)
					}
				}
			} else {
				s.PrintError(inputData,"formal parameters expected")
			}
			fp := st.TopScope(inputData)
			sc[len(sc) - 1].Par = fp
			if inputData.Sym == k.RPAREN {
				s.GetSym(inputData)
			} else {
				s.PrintError(inputData,") expected")
			}
		} else {
			fp := []st.SymTableEntry{}
			cg.GenProcStart(ident, fp, inputData)
		}
		if inputData.Sym == k.SEMICOLON {
			s.GetSym(inputData)
		} else {
			s.PrintError(inputData,"; expected")
		}
		declaration("local", inputData)
		cg.GenProcEntry(inputData)
		x := compoundStatement(inputData)
		cg.GenProcExit(x, inputData)
		if inputData.Sym == k.SEMICOLON {
			s.GetSym(inputData)
		} else {
			s.PrintError(inputData,"; expected")
		}
	}
}

// Parses the "program" part of the grammar.
func Program(inputData *i.InputData) string {
	st.NewDecl("boolean", cg.GenBool(st.Type(st.Bool)), inputData)
	st.NewDecl("integer", cg.GenInt(st.Type(st.Int)), inputData)
	st.NewDecl("true", st.Const(st.Bool, 1), inputData)
	st.NewDecl("false", st.Const(st.Bool, 0), inputData)
	st.NewDecl("read", st.StdProc([]st.SymTableEntry{*st.Ref(st.Int)}), inputData)
	st.NewDecl("write", st.StdProc([]st.SymTableEntry{*st.Var(st.Int)}), inputData)
	st.NewDecl("writeln", st.StdProc([]st.SymTableEntry{}), inputData)
	cg.GenProgStart(inputData)
	if inputData.Sym == k.PROGRAM {
		s.GetSym(inputData)
	} else {
		s.PrintError(inputData, "'program' expected")
	}
	ident := inputData.Val
	if inputData.Sym == k.IDENT {
		s.GetSym(inputData)
	} else {
		s.PrintError(inputData, "program name expected")
	}
	if inputData.Sym == k.SEMICOLON {
		s.GetSym(inputData)
	} else {
		s.PrintError(inputData, "; expected")
	}
	declaration("global", inputData)
	cg.GenProgEntry(ident, inputData)
	x := compoundStatement(inputData)
	return cg.GenProgExit(x, inputData)
}

// Compiles the code into WASM.
func CompileWasm(inputData *i.InputData) {
	s.Init(inputData)
	p := Program(inputData)
	cg.WriteWasmFile("result.wasm", p)
}