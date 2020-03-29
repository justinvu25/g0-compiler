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
// the provided filename.
func WriteWasmFile(filename string, inputData *i.InputData) {
	generatedCode := []byte(GenProgExit(inputData))
	err := ioutil.WriteFile(filename, generatedCode, 0644)

	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(filename + " was created.")
	}
}

// Generates the start of the WASM code.
func GenProgStart(inputData *i.InputData) {
	inputData.Asm = append(inputData.Asm, "(module")
}

// Specifies the size of the bool type.
func GenBool(entry *st.SymTableEntry) *st.SymTableEntry {
	entry.Size = 1
	return entry
}

// Specifies the size of the Int type
func GenInt(entry *st.SymTableEntry) *st.SymTableEntry {
	entry.Size = 4
	return entry
}

// Generates data for a record.
func GenRec(entry *st.SymTableEntry) *st.SymTableEntry{
	s := 0
	for _, f := range entry.Ctp.Fields {
		f.Offset = s
		s = s + f.Size
	}
	entry.Size = s
	return entry
}

// Generates data for an array
func GenArray(entry *st.SymTableEntry) *st.SymTableEntry {
	entry.Size = entry.Ctp.Length * entry.Ctp.Size
	return entry
}

// Generates WASM code for global vars.
func GenGlobalVars(scope []st.SymTableEntry, start int, inputData *i.InputData) {
	i := start
	for i < len(scope) {
		if scope[i].EntryType == "var" {
			if scope[i].Tp == st.Int || scope[i].Tp == st.Bool {
				inputData.Asm = append(inputData.Asm, "(global $" + scope[i].Name + " (mut i32) i32.const 0)")
			} else if scope[i].Ctp.EntryType == "array" || scope[i].Ctp.EntryType == "record" {
				scope[i].Lev = -2
				scope[i].Adr = inputData.Memsize
				inputData.Memsize = inputData.Memsize + scope[i].Size
			} else {
				s.PrintError(inputData, "WASM: type?")
			}
		}
	}
}

// Generates WASM code for local vars.
func GenLocalVars(scope []st.SymTableEntry, start int, inputData *i.InputData) st.PrimitiveType {
	i := start
	for i < len(scope) {
		if scope[i].EntryType == "var" {
			if scope[i].Tp == st.Int || scope[i].Tp == st.Bool {
				inputData.Asm = append(inputData.Asm, "(local $" + scope[i].Name + " i32)")
			} else if scope[i].Ctp.EntryType == "array" || scope[i].Ctp.EntryType == "record" {
				s.PrintError(inputData, "WASM: no local arrays, records")
			} else {
				s.PrintError(inputData, "WASM: type?")
			}
		}
	}

	return st.None
}

// Loads items onto WASM stack.
func loadItem(entry *st.SymTableEntry, inputData *i.InputData) {
	if entry.EntryType == "var" {
		if entry.Lev == 0 {
			inputData.Asm = append(inputData.Asm, "global.get $" + entry.Name)
		} else if entry.Lev == inputData.Curlev {
			inputData.Asm = append(inputData.Asm, "local.get $" + entry.Name)
		} else if entry.Lev == -2 {
			inputData.Asm = append(inputData.Asm, "i32.const" + strconv.Itoa(entry.Adr))
			inputData.Asm = append(inputData.Asm, "i32.load")
		} else if entry.Lev != -1 {
			s.PrintError(inputData, "WASM: var level")
		} 
	} else if entry.EntryType == "ref" {
		if entry.Lev == -1 {
			inputData.Asm = append(inputData.Asm, "i32.load")
		} else if entry.Lev == inputData.Curlev {
			inputData.Asm = append(inputData.Asm, "local.get $" + entry.Name)
			inputData.Asm = append(inputData.Asm, "i32.load")
		} else {
			s.PrintError(inputData, "WASM: ref level")
		}
	} else if entry.EntryType == "const" {
		inputData.Asm = append(inputData.Asm, "i32.const " + strconv.Itoa(entry.Val))
	}
}

// Creates a variable.
func GenVar(entry *st.SymTableEntry, inputData *i.InputData) st.SymTableEntry {
	y := st.SymTableEntry{}

	if 0 < entry.Lev && entry.Lev < inputData.Curlev {
		s.PrintError(inputData, "WASM: level")
	}
	if entry.EntryType == "ref" {
		y = st.NewSymTableEntry("ref", entry.Name, entry.Tp, st.NewComplexType("", []st.SymTableEntry{}, st.None, int(0), int(0), []string{}), entry.Lev, 0, nil)
	} else if entry.EntryType == "var" {
		y = st.NewSymTableEntry("var", entry.Name, entry.Tp, st.NewComplexType("", []st.SymTableEntry{}, st.None, int(0), int(0), []string{}), entry.Lev, 0, nil)
		if entry.Lev == -2 {
			y.Adr = entry.Adr
		}
	}

	return y
}

// Creates a const.
func GenConst(entry *st.SymTableEntry) *st.SymTableEntry {
	return entry
}

// Generates WASM code for unary operators.
func GenUnaryOp(op int, entry *st.SymTableEntry, inputData *i.InputData) *st.SymTableEntry{
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

// Generates WASM code for binary operators.
func GenBinaryOp(op int, x *st.SymTableEntry, y *st.SymTableEntry, inputData *i.InputData) *st.SymTableEntry{
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
		x.Tp = st.Int
	} else if op == k.AND {
		loadItem(y, inputData)
		inputData.Asm = append(inputData.Asm, "else")
		inputData.Asm = append(inputData.Asm, "i32.const 0")
		inputData.Asm = append(inputData.Asm, "end")
		x.Tp = st.Bool
	} else if op == k.OR {
		loadItem(y, inputData)
		inputData.Asm = append(inputData.Asm, "end")
		x.Tp = st.Bool
	}
	x.Lev = -1
	return x
}

// Generates WASM code for relations, such as x > y.
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

	x.Tp = st.Bool
	x.Lev = -1
	return x
}

// Generates WASM code for select.
func GenSelect(entry *st.SymTableEntry, field *st.SymTableEntry, inputData *i.InputData) *st.SymTableEntry{
	if entry.EntryType == "var"{
		entry.Adr += field.Offset
	} else if entry.EntryType == "ref" {
		if entry.Lev > 0 {
			inputData.Asm = append(inputData.Asm, "local.get $" + entry.Name)
		}
		inputData.Asm = append(inputData.Asm, "i32.const " + strconv.Itoa(field.Offset))
		inputData.Asm = append(inputData.Asm, "i32.add")
		entry.Lev = -1
	}
	entry.Tp = field.Tp
	return entry
}

// Generates WASM code for getting items out of indexed elements.
func GenIndex(x *st.SymTableEntry, y *st.SymTableEntry, inputData *i.InputData) {
	if x.EntryType == "var" {
		if y.EntryType == "const" {
			x.Adr += (y.Val - x.Ctp.Lower) * x.Ctp.Size
			x.Tp = x.Ctp.Base
		} else {
			loadItem(y, inputData)
			if x.Ctp.Lower != 0 {
				inputData.Asm = append(inputData.Asm, "i32.const " + strconv.Itoa(x.Ctp.Lower))
				inputData.Asm = append(inputData.Asm, "i32.sub")
			}
			inputData.Asm = append(inputData.Asm, "i32.const " + strconv.Itoa(x.Ctp.Size))
			inputData.Asm = append(inputData.Asm, "i32.mul")
			inputData.Asm = append(inputData.Asm, "i32.const " + strconv.Itoa(x.Adr))
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
			inputData.Asm = append(inputData.Asm, "i32.const " + strconv.Itoa((y.Val - x.Ctp.Lower) * x.Ctp.Size))
			inputData.Asm = append(inputData.Asm, "i32.add")
		} else {
			loadItem(y, inputData)
			inputData.Asm = append(inputData.Asm, "i32.const " + strconv.Itoa(x.Ctp.Lower))
			inputData.Asm = append(inputData.Asm, "i32.sub")
			inputData.Asm = append(inputData.Asm, "i32.const " + strconv.Itoa(x.Ctp.Size))
			inputData.Asm = append(inputData.Asm, "i32.mul")
			inputData.Asm = append(inputData.Asm, "i32.add")
		}
	}
}

// Generates WASM code for assignment.
func GenAssign(x *st.SymTableEntry, y *st.SymTableEntry, inputData *i.InputData) {
	if x.EntryType == "var" {
		if x.Lev == -2 {
			inputData.Asm = append(inputData.Asm, "i32.const " + strconv.Itoa(x.Adr))
		}
		loadItem(y, inputData)
		if x.Lev == 0 {
			inputData.Asm = append(inputData.Asm, "global.set $ " + x.Name)
		} else if x.Lev == inputData.Curlev {
			inputData.Asm = append(inputData.Asm, "local.set $ " + x.Name)
		} else if x.Lev == -2 {
			inputData.Asm = append(inputData.Asm, "i32.store")
		} else {
			s.PrintError(inputData, "WASM: level")
		}
	} else if x.EntryType == "ref" {
		if x.Lev == inputData.Curlev {
			inputData.Asm = append(inputData.Asm, "local.get $" + x.Name)
		}
		loadItem(y, inputData)
		inputData.Asm = append(inputData.Asm, "i32.store")
	}
}

// Generates WASM code for entering a program.
func GenProgEntry(inputData *i.InputData) {
	inputData.Asm = append(inputData.Asm, "(func $program")
}

func GenProgExit(inputData *i.InputData) string{
	closingString := ")\n(memory " + strconv.Itoa(inputData.Memsize / int(math.Exp2(16)) + 1) + ")\n(start $program)\n"
	inputData.Asm = append(inputData.Asm, closingString)
	outputCode := ""
	for _, asm := range inputData.Asm {
		outputCode += "\n" + asm
	}

	return outputCode
}

// Generates WASM code for the signature of a procedure.
func GenProcStart(ident string, listOfParams []st.SymTableEntry, inputData *i.InputData) {
	if inputData.Curlev > 0 {
		s.PrintError(inputData, "WASM: no nested procedures")
	}
	inputData.Curlev += 1
	params := ""

	for _, param := range listOfParams {
		params += "(param $" + param.Name + " i32)"
	}

	inputData.Asm = append(inputData.Asm, "(func $" + ident + params)
}

func GenProcEntry(inputData *i.InputData) {
	//pass
}

// Generates WASM code for closing procedure code.
func GenProcExit(inputData *i.InputData) {
	inputData.Curlev -= 1
	inputData.Asm = append(inputData.Asm, ")")
}

// Generates WASM code for parameter values.
func GenActualPara(ap *st.SymTableEntry, fp *st.SymTableEntry, inputData *i.InputData) {
	if ap.EntryType == "ref" {
		if ap.Lev == -2 {
			inputData.Asm = append(inputData.Asm, "i32.const " + strconv.Itoa(ap.Adr))
		}
	} else if ap.EntryType == "var" || ap.EntryType == "ref" || ap.EntryType == "const" {
		loadItem(ap, inputData)
	} else {
		s.PrintError(inputData, "unsupported parameter type")
	}
}

// Generates WASM code for function calls.
func GenCall(entry *st.SymTableEntry, inputData *i.InputData) {
	inputData.Asm = append(inputData.Asm, "call $" + entry.Name)
}

func GenSeq(inputData *i.InputData) {
	//pass
}

// Generates WASM code for Then statements.
func GenThen(x *st.SymTableEntry, inputData *i.InputData) *st.SymTableEntry {
	loadItem(x, inputData)
	inputData.Asm = append(inputData.Asm, "if")
	return x
}

// Generates WASM code for closing IfThen.
func GenIfThen(inputData *i.InputData) {
	inputData.Asm = append(inputData.Asm, "end")
}

// Generates WASM code for Else statements.
func GenElse(inputData *i.InputData) {
	inputData.Asm = append(inputData.Asm, "else")
}

// Generates WASM code for Else statements.
func GenIfElse(inputData *i.InputData) {
	inputData.Asm = append(inputData.Asm, "end")
}
// Generates WASM code for the start of while loops.
func GenWhile(inputData *i.InputData) {
	inputData.Asm = append(inputData.Asm, "loop")
}

// Generates WASM code for Do statements.
func GenDo(x *st.SymTableEntry, inputData *i.InputData) *st.SymTableEntry {
	loadItem(x, inputData)
	inputData.Asm = append(inputData.Asm, "if")
	return x
}

// Generates WASM code for closing while loops.
func GenWhileDo(inputData *i.InputData) {
	inputData.Asm = append(inputData.Asm, "br 1")
	inputData.Asm = append(inputData.Asm, "end")
	inputData.Asm = append(inputData.Asm, "end")
}