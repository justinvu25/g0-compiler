package codegen

import (
	"fmt"
	i "group-11/pkg/inputdata"
	k "group-11/pkg/keywords"
	s "group-11/pkg/scanner"
	st "group-11/pkg/symtable"
	"io/ioutil"
	"log"
	"math"
	"strconv"
)

// Takes the asm string and converts it into a WASM code file with
// the provided fileName.
func WriteWasmFile(fileName string, code string) {
	generatedCode := []byte(code)
	err := ioutil.WriteFile(fileName, generatedCode, 0644)

	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(fileName + " was created.")
	}
}

// Generates the start of programs.
func GenProgStart(inputData *i.InputData) {
	inputData.Asm = append(inputData.Asm, "(module",
		"(import \"P0lib\" \"write\" (func $write (param i32)))",
		"(import \"P0lib\" \"writeln\" (func $writeln))",
		"(import \"P0lib\" \"read\" (func $read (result i32)))")
}

// Specifies the Size of bool typed entries.
func GenBool(entry *st.SymTableEntry) *st.SymTableEntry {
	entry.Size = 1
	return entry
}

// Specifies the Size of int typed entries.
func GenInt(entry *st.SymTableEntry) *st.SymTableEntry {
	entry.Size = 4
	return entry
}

// Generates records, calculating some of the attribute values.
func GenRec(entry *st.SymTableEntry) *st.SymTableEntry {
	s := 0
	for _, f := range entry.Ctp.Fields {
		f.Offset = s
		s = s + f.Size
	}
	entry.Size = s
	return entry
}

// Generates records, calculating its Size.
func GenArray(entry *st.SymTableEntry) *st.SymTableEntry {
	entry.Size = entry.Ctp.Length * entry.Ctp.Size
	return entry
}

// Generates all of the global.
func GenGlobalVars(scope []st.SymTableEntry, start int, inputData *i.InputData) {
	i := start
	for i < len(scope) {
		if scope[i].EntryType == "var" {
			if scope[i].Tp == st.Int || scope[i].Tp == st.Bool {
				inputData.Asm = append(inputData.Asm, "(global $"+scope[i].Name+" (mut i32) i32.const 0)")
			} else if scope[i].EntryType == "array" || scope[i].EntryType == "record" {
				scope[i].Lev = -2
				scope[i].Adr = inputData.Memsize
				inputData.Memsize = inputData.Memsize + scope[i].Size
			} else {
				s.PrintError(inputData, "WASM: type?")
			}
		}
		i += 1
	}
}

// Generates all of the local vars.
func GenLocalVars(scope []st.SymTableEntry, start int, inputData *i.InputData) st.PrimitiveType {
	i := start
	for i < len(scope) {
		if scope[i].EntryType == "var" {
			if scope[i].Tp == st.Int || scope[i].Tp == st.Bool {
				inputData.Asm = append(inputData.Asm, "(local $"+scope[i].Name+" i32)")
			} else if scope[i].EntryType == "array" || scope[i].EntryType == "record" {
				s.PrintError(inputData, "WASM: no local arrays, records")
			} else {
				s.PrintError(inputData, "WASM: type?")
			}
		}
		i += 1
	}

	return st.None
}

// Loads a sym table entry onto the stack.
func loadItem(entry *st.SymTableEntry, inputData *i.InputData) {
	if entry.EntryType == "var" {
		if entry.Lev == 0 {
			inputData.Asm = append(inputData.Asm, "global.get $"+entry.Name)
		} else if entry.Lev == inputData.Curlev {
			inputData.Asm = append(inputData.Asm, "local.get $"+entry.Name)
		} else if entry.Lev == -2 {
			inputData.Asm = append(inputData.Asm, "i32.const"+strconv.Itoa(entry.Adr))
			inputData.Asm = append(inputData.Asm, "i32.load")
		} else if entry.Lev != -1 {
			s.PrintError(inputData, "WASM: var Level")
		}
	} else if entry.EntryType == "ref" {
		if entry.Lev == -1 {
			inputData.Asm = append(inputData.Asm, "i32.load")
		} else if entry.Lev == inputData.Curlev {
			inputData.Asm = append(inputData.Asm, "local.get $"+entry.Name)
			inputData.Asm = append(inputData.Asm, "i32.load")
		} else {
			s.PrintError(inputData, "WASM: ref Level")
		}
	} else if entry.EntryType == "const" {
		inputData.Asm = append(inputData.Asm, "i32.const "+strconv.Itoa(entry.Val))
	}
}

// Generates a var using the provided symbol table entry.
func GenVar(entry *st.SymTableEntry, inputData *i.InputData) *st.SymTableEntry {
	y := &st.SymTableEntry{}

	if 0 < entry.Lev && entry.Lev < inputData.Curlev {
		s.PrintError(inputData, "WASM: Level")
	}
	if entry.EntryType == "ref" {
		y = st.Ref(entry.Tp)
		y.Lev = entry.Lev
		y.Name = entry.Name
	} else if entry.EntryType == "var" {
		y = st.Var(entry.Tp)
		y.Lev = entry.Lev
		y.Name = entry.Name
		if entry.Lev == -2 {
			y.Adr = entry.Adr
		}
	}

	if entry.ArrOrRec == "array" || entry.ArrOrRec == "record" {
		y.Ctp = entry.Ctp
		y.ArrOrRec = entry.ArrOrRec
	}

	return y
}

// Constants are simply constants so they do not need any extra work.
func GenConst(entry *st.SymTableEntry) *st.SymTableEntry {
	return entry
}

// Generates code for operations with unary operators.
func GenUnaryOp(op int, entry *st.SymTableEntry, inputData *i.InputData) *st.SymTableEntry {
	loadItem(entry, inputData)
	if op == k.MINUS {
		inputData.Asm = append(inputData.Asm, "i32.const -1")
		inputData.Asm = append(inputData.Asm, "i32.mul")
		entry.EntryType = "var"
		entry.Tp = st.Int
		entry.Lev = -1
	} else if op == k.NOT {
		inputData.Asm = append(inputData.Asm, "i32.eqz")
		entry.Tp = st.Bool
		entry.Lev = -1
	} else if op == k.AND {
		inputData.Asm = append(inputData.Asm, "if (result i32)")
		entry.Tp = st.Bool
		entry.Lev = -1
	} else if op == k.OR {
		inputData.Asm = append(inputData.Asm, "if (result i32)")
		inputData.Asm = append(inputData.Asm, "i32.const 1")
		inputData.Asm = append(inputData.Asm, "else")
		entry.Tp = st.Bool
		entry.Lev = -1
	} else {
		s.PrintError(inputData, "WASM: unary operator?")
	}

	return entry
}

// Generates code for operations with binary operators.
func GenBinaryOp(op int, x *st.SymTableEntry, y *st.SymTableEntry, inputData *i.InputData) *st.SymTableEntry {
	if op == k.PLUS || op == k.MINUS || op == k.TIMES || op == k.DIV || op == k.MOD {
		loadItem(x, inputData)
		loadItem(y, inputData)
		if op == k.PLUS {
			inputData.Asm = append(inputData.Asm, "i32.add")
		} else if op == k.MINUS {
			inputData.Asm = append(inputData.Asm, "i32.sub")
		} else if op == k.TIMES {
			inputData.Asm = append(inputData.Asm, "i32.mul")
		} else if op == k.DIV {
			inputData.Asm = append(inputData.Asm, "i32.div_s")
		} else if op == k.MOD {
			inputData.Asm = append(inputData.Asm, "i32.rem_s")
		} else {
			s.PrintError(inputData, "WASM: binary operator?")
		}
		x = st.Var(st.Int)
		x.Lev = -1
	} else if op == k.AND {
		loadItem(y, inputData)
		inputData.Asm = append(inputData.Asm, "else")
		inputData.Asm = append(inputData.Asm, "i32.const 0")
		inputData.Asm = append(inputData.Asm, "end")
		x = st.Var(st.Bool)
		x.Lev = -1
	} else if op == k.OR {
		loadItem(y, inputData)
		inputData.Asm = append(inputData.Asm, "end")
		x = st.Var(st.Bool)
		x.Lev = -1
	}
	return x
}

// Generates relations between two entries, such as x > 5.
func GenRelation(op int, x *st.SymTableEntry, y *st.SymTableEntry, inputData *i.InputData) *st.SymTableEntry {
	loadItem(x, inputData)
	loadItem(y, inputData)
	if op == k.EQ {
		inputData.Asm = append(inputData.Asm, "i32.eq")
	} else if op == k.NE {
		inputData.Asm = append(inputData.Asm, "i32.ne")
	} else if op == k.LT {
		inputData.Asm = append(inputData.Asm, "i32.lt_s")
	} else if op == k.GT {
		inputData.Asm = append(inputData.Asm, "i32.gt_s")
	} else if op == k.LE {
		inputData.Asm = append(inputData.Asm, "i32.le_s")
	} else if op == k.GE {
		inputData.Asm = append(inputData.Asm, "i32.ge_s")
	}

	x = st.Var(st.Bool)
	x.Lev = -1
	return x
}

// Generates selectors for records.
func GenSelect(entry *st.SymTableEntry, field *st.SymTableEntry, inputData *i.InputData) *st.SymTableEntry {
	if entry.EntryType == "var" {
		entry.Adr += field.Offset
	} else if entry.EntryType == "ref" {
		if entry.Lev > 0 {
			inputData.Asm = append(inputData.Asm, "local.get $"+entry.Name)
		}
		inputData.Asm = append(inputData.Asm, "i32.const "+strconv.Itoa(field.Offset))
		inputData.Asm = append(inputData.Asm, "i32.add")
		entry.Lev = -1
	}
	entry.Tp = field.Tp
	return entry
}

// Generates indexes for arrays.
func GenIndex(x *st.SymTableEntry, y *st.SymTableEntry, inputData *i.InputData) *st.SymTableEntry {
	if x.EntryType == "var" {
		if y.EntryType == "const" {
			x.Adr += (y.Val - x.Ctp.Lower) * x.Ctp.Size
			x.Tp = x.Ctp.Base
		} else {
			loadItem(y, inputData)
			if x.Ctp.Lower != 0 {
				inputData.Asm = append(inputData.Asm, "i32.const "+strconv.Itoa(x.Ctp.Lower))
				inputData.Asm = append(inputData.Asm, "i32.sub")
			}
			inputData.Asm = append(inputData.Asm, "i32.const "+strconv.Itoa(x.Ctp.Size))
			inputData.Asm = append(inputData.Asm, "i32.mul")
			inputData.Asm = append(inputData.Asm, "i32.const "+strconv.Itoa(x.Adr))
			inputData.Asm = append(inputData.Asm, "i32.add")
			x.EntryType = "ref"
			x.Tp = x.Ctp.Base
		}
	} else {
		if x.Lev == inputData.Curlev {
			loadItem(x, inputData)
			x.Lev = -1
		}
		if x.EntryType == "const" {
			inputData.Asm = append(inputData.Asm, "i32.const "+strconv.Itoa((y.Val-x.Ctp.Lower)*x.Ctp.Size))
			inputData.Asm = append(inputData.Asm, "i32.add")
		} else {
			loadItem(y, inputData)
			inputData.Asm = append(inputData.Asm, "i32.const "+strconv.Itoa(x.Ctp.Lower))
			inputData.Asm = append(inputData.Asm, "i32.sub")
			inputData.Asm = append(inputData.Asm, "i32.const "+strconv.Itoa(x.Ctp.Size))
			inputData.Asm = append(inputData.Asm, "i32.mul")
			inputData.Asm = append(inputData.Asm, "i32.add")
		}
	}

	return x
}

// Generates assignment to variables.
func GenAssign(x *st.SymTableEntry, y *st.SymTableEntry, inputData *i.InputData) {
	if x.EntryType == "var" {
		if x.Lev == -2 {
			inputData.Asm = append(inputData.Asm, "i32.const " + strconv.Itoa(x.Adr))
		}
		loadItem(y, inputData)
		if x.Lev == 0 {
			inputData.Asm = append(inputData.Asm, "global.set $" + x.Name)
		} else if x.Lev == inputData.Curlev {
			inputData.Asm = append(inputData.Asm, "local.set $ " + x.Name)
		} else if x.Lev == -2 {
			inputData.Asm = append(inputData.Asm, "i32.store")
		} else {
			s.PrintError(inputData, "WASM: Level")
		}
	} else if x.EntryType == "ref" {
		if x.Lev == inputData.Curlev {
			inputData.Asm = append(inputData.Asm, "local.get $" + x.Name)
		}
		loadItem(y, inputData)
		inputData.Asm = append(inputData.Asm, "i32.store")
	}
}

// Generates the entry to the program.
func GenProgEntry(ident string, inputData *i.InputData) {
	inputData.Asm = append(inputData.Asm, "(func $program")
}

// Generates the exit to the program.
func GenProgExit(x *st.SymTableEntry, inputData *i.InputData) string {
	closingString := ")\n(memory " + strconv.Itoa(inputData.Memsize/int(math.Exp2(16))+1) + ")\n(start $program)\n)"
	inputData.Asm = append(inputData.Asm, closingString)
	outputCode := ""
	for _, asm := range inputData.Asm {
		outputCode += "\n" + asm
	}

	return outputCode
}

// Generates function signatures.
func GenProcStart(ident string, listOfParams []st.SymTableEntry, inputData *i.InputData) {
	if inputData.Curlev > 0 {
		s.PrintError(inputData, "WASM: no nested procedures")
	}
	inputData.Curlev += 1
	params := ""

	for _, param := range listOfParams {
		params += "(param $" + param.Name + " i32)"
	}

	inputData.Asm = append(inputData.Asm, "(func $"+ident+params)
}

// Dummy function for generating procedure entries.
func GenProcEntry(inputData *i.InputData) {
	//pass
}

// Generates procedure exits, which is simply a closing parenthesis.
func GenProcExit(x *st.SymTableEntry, inputData *i.InputData) {
	inputData.Curlev -= 1
	inputData.Asm = append(inputData.Asm, ")")
}

// Generates the actual parameters using the provided formal parameters.
func GenActualPara(ap *st.SymTableEntry, fp *st.SymTableEntry, inputData *i.InputData) *st.SymTableEntry {
	if fp.EntryType == "ref" {
		if ap.Lev == -2 {
			inputData.Asm = append(inputData.Asm, "i32.const "+strconv.Itoa(ap.Adr))
		}
	} else if ap.EntryType == "var" || ap.EntryType == "ref" || ap.EntryType == "const" {
		loadItem(ap, inputData)
	} else {
		s.PrintError(inputData, "unsupported parameter type")
	}

	return ap
}

// Generates function calls.
func GenCall(entry *st.SymTableEntry, inputData *i.InputData) *st.SymTableEntry {
	inputData.Asm = append(inputData.Asm, "call $" + entry.Name)
	return entry
}

// Generates call to the WASM stdproc read().
func GenRead(x *st.SymTableEntry, inputData *i.InputData) {
	inputData.Asm = append(inputData.Asm, "call $read")
	y := st.Var(st.Int)
	y.Lev = -1
}

// Generates call to the WASM stdproc write().
func GenWrite(x *st.SymTableEntry, inputData *i.InputData) {
	loadItem(x, inputData)
	inputData.Asm = append(inputData.Asm, "call $write")
}

// Generates call to the WASM stdproc writeln().
func GenWriteln(inputData *i.InputData) {
	inputData.Asm = append(inputData.Asm, "call $writeln")
}

// Dummy function for generating sequences.
func GenSeq(x *st.SymTableEntry, y *st.SymTableEntry, inputData *i.InputData) {
	//pass
}

// Generates then.
func GenThen(x *st.SymTableEntry, inputData *i.InputData) *st.SymTableEntry {
	loadItem(x, inputData)
	inputData.Asm = append(inputData.Asm, "if")
	return x
}

// Generates if/then.
func GenIfThen(x *st.SymTableEntry, y *st.SymTableEntry, inputData *i.InputData) *st.SymTableEntry {
	inputData.Asm = append(inputData.Asm, "end")
	return x
}

// Generates else.
func GenElse(x *st.SymTableEntry, y *st.SymTableEntry, inputData *i.InputData) *st.SymTableEntry {
	inputData.Asm = append(inputData.Asm, "else")
	return y
}

// Generates if/else
func GenIfElse(x *st.SymTableEntry, y *st.SymTableEntry, z *st.SymTableEntry, inputData *i.InputData) *st.SymTableEntry {
	inputData.Asm = append(inputData.Asm, "end")
	return x
}

// Generates while.
func GenWhile(inputData *i.InputData) {
	inputData.Asm = append(inputData.Asm, "loop")
}

// Generates do.
func GenDo(x *st.SymTableEntry, inputData *i.InputData) *st.SymTableEntry {
	loadItem(x, inputData)
	inputData.Asm = append(inputData.Asm, "if")
	return x
}

// Generates while/do.
func GenWhileDo(x *st.SymTableEntry, y *st.SymTableEntry, inputData *i.InputData) {
	inputData.Asm = append(inputData.Asm, "br 1")
	inputData.Asm = append(inputData.Asm, "end")
	inputData.Asm = append(inputData.Asm, "end")
}
