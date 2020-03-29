package main

import (
	"fmt"
	"sync"
	"io/ioutil"
	"log"
	"math"
	"strconv"
	"unicode"
	"strings"
)

type InputData struct {
	Input      string // P0 source cmd
	Sym        int    // Symbol that was identified.
	Ch         string // Current character.
	Index      int    // Helps identify symbols
	Val        string // Value
	LineNumber int    // Current line being parsed
	LastLine   int    // Previous line that was parsed
	ErrorLine  int    // Used to help surpress multiple errors
	Pos        int    // Current position of parser in a line
	LastPos    int    // Previous position
	ErrorPos   int    // Used to help surpress multiple errors
	Error      bool   // Set to true when an error is found.
	SymTable   [][]SymTableEntry // Symbol table of items that will be turned into WASM.
	curlev     int    // Current scope level of the code generator.
	memsize    int	  // Size of the required memory allocation.
	asm		   []string // The string that will ultimately become the WASM file. 
}

// constructor for InputData struct
func NewInputData(fileName string) *InputData {
	input := FileHelper(fileName)
	s := InputData{
		Input:      input,
		Sym:        0,
		Ch:         "",
		Index:      0,
		Val:        "",
		LineNumber: 1,
		LastLine:   1,
		ErrorLine:  1,
		Pos:        0,
		LastPos:    0,
		ErrorPos:   0,
		Error:      false,
		SymTable:	[][]SymTableEntry{{}},
		curlev:		0,
		memsize:	0,
		asm:		[]string{}}
	return &s
}

// Read in data from a file.
func FileHelper(fileName string) string {
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	contentsEOFChar := string(contents) + "~"

	return string(contentsEOFChar)
}

func main() {
	var wg sync.WaitGroup
	inputData := NewInputData("p0test.txt") //newInputData("p0code.txt")

	// Add items to the wait group, one for each goroutine.
	wg.Add(3)

	go EatWhiteSpace(inputData, &wg)
	go EatComments(inputData, &wg)
	go ParseInput(inputData, &wg, "a source filename goes here")
	// Wait for the waitgroup counter to reach zero before continuing.
	// The waitgroup counter is decremented each time a thread finishes
	// executing its procedure.
	wg.Wait()
	fmt.Println("done all tasks")
}


/*
	Keywords and consts
*/
const (
	TIMES     = 1
	DIV       = 2
	MOD       = 3
	AND       = 4
	PLUS      = 5
	MINUS     = 6
	OR        = 7
	EQ        = 8
	NE        = 9
	LT        = 10
	GT        = 11
	LE        = 12
	GE        = 13
	PERIOD    = 14
	COMMA     = 15
	COLON     = 16
	RPAREN    = 17
	RBRAK     = 18
	OF        = 19
	THEN      = 20
	DO        = 21
	LPAREN    = 22
	LBRAK     = 23
	NOT       = 24
	BECOMES   = 25
	NUMBER    = 26
	IDENT     = 27
	SEMICOLON = 28
	END       = 29
	ELSE      = 30
	IF        = 31
	WHILE     = 32
	ARRAY     = 33
	RECORD    = 34
	CONST     = 35
	TYPE      = 36
	VAR       = 37
	PROCEDURE = 38
	BEGIN     = 39
	PROGRAM   = 40
	EOF       = 41
)

var Keywords = map[string]int{
	"div":       DIV,
	"mod":       MOD,
	"and":       AND,
	"or":        OR,
	"of":        OF,
	"then":      THEN,
	"do":        DO,
	"not":       NOT,
	"end":       END,
	"else":      ELSE,
	"if":        IF,
	"while":     WHILE,
	"array":     ARRAY,
	"record":    RECORD,
	"const":     CONST,
	"type":      TYPE,
	"var":       VAR,
	"procedure": PROCEDURE,
	"begin":     BEGIN,
	"program":   PROGRAM}

var FIRSTFACTOR 	= [4]int{IDENT, NUMBER, LPAREN, NOT}
var FOLLOWFACTOR 	= [22]int{TIMES, DIV, MOD, AND, OR, PLUS, MINUS, EQ, NE, LT, LE, GT, GE,
	COMMA, SEMICOLON, THEN, ELSE, RPAREN, RBRAK, DO, PERIOD, END}
var FIRSTEXPRESSION = [6]int{PLUS, MINUS, IDENT, NUMBER, LPAREN, NOT}
var FIRSTSTATEMENT 	= [4]int{IDENT, IF, WHILE, BEGIN}
var FOLLOWSTATEMENT = [3]int{SEMICOLON, END, ELSE}
var FIRSTTYPE		= [4]int{IDENT, RECORD, ARRAY, LPAREN}
var FOLLOWTYPE		= [1]int{SEMICOLON}
var FIRSTDECL		= [4]int{CONST, TYPE, VAR, PROCEDURE}
var FOLLOWDECL		= [1]int{BEGIN}
var FOLLOWPROCCALL	= [3]int{SEMICOLON, END, ELSE}
var STRONGSYMS		= [8]int{CONST, TYPE, VAR, PROCEDURE, WHILE, IF, BEGIN, EOF}

// Parses characters into tokens.
func ParseInput(inputData *InputData, wg *sync.WaitGroup, srcfile string) {
	defer wg.Done()
	CompileWasm(srcfile, inputData)
	fmt.Println("done parsing")
}

// Removes whitespace.
func EatWhiteSpace(inputData *InputData, wg *sync.WaitGroup) {
	defer wg.Done()
	//
	inputData.Input = strings.Replace(inputData.Input, " do", "!do", -1)
	inputData.Input = strings.Replace(inputData.Input, " end", "!end", -1)
	// Giving -1 to string.Replace removes an unlimited number of whitespaces.
	inputData.Input = strings.Replace(inputData.Input, " ", "", -1)

	fmt.Println("done removing whitespace")
}

// Removes comments
func EatComments(inputData *InputData, wg *sync.WaitGroup) {
	defer wg.Done()
	i := 0
	opening := false
	closing := false
	m := 0
	n := 0

	for string((inputData.Input)[i]) != "~" {
		if string((inputData.Input)[i]) == "{" {
			m = i
			opening = true
		} else if string((inputData.Input)[i]) == "}" && opening == true {
			n = i
			closing = true
		}
		i++
		if opening == true && closing == true {
			opening = false
			closing = false
			inputData.Input = string((inputData.Input)[:m] + "" + (inputData.Input)[n+1:])
			i = m
		}
	}
	// fmt.Println(inputData.input)
	fmt.Println("done removing comments")
}

func Init(inputData *InputData) {
	GetChar(inputData)
	GetSym(inputData)
}

// Moves the index to the next char
func GetChar(inputData *InputData) {
	// This loop will wait at a given index for the whitespace and comment eaters to catch up.
	for string(inputData.Input[inputData.Index]) == " " {
		//pass
	}
	currentChar := string(inputData.Input[inputData.Index])
	if currentChar == "~" {
		inputData.Ch = currentChar
	} else {
		inputData.Ch = currentChar
		inputData.Index += 1
		if currentChar == "\n" {
			inputData.Pos = 0
			inputData.LineNumber += 1
		} else {
			inputData.LastLine = inputData.LineNumber
			inputData.Pos += 1
		}
	}
}

// Identifies keywords in a given sequence of characters.
func IdentKeyword(inputData *InputData) {
	start := inputData.Index - 1
	current := ""
	for unicode.IsLetter([]rune(inputData.Ch)[0]) {
		current = inputData.Input[start:inputData.Index]
		if val, ok := Keywords[current]; ok {
			inputData.Sym = val
			GetChar(inputData)
			break
		} else {
			inputData.Sym = IDENT
		}
		GetChar(inputData)
	}
	inputData.Val = current
	fmt.Print(inputData.Val)
}

// Converts a string of numbers into value
func Number(inputData *InputData) {
	inputData.Sym = NUMBER
	inputData.Val = ""
	for unicode.IsNumber([]rune(inputData.Ch)[0]) {
		inputData.Val += inputData.Ch
		GetChar(inputData)
	}

	res, err3 := strconv.Atoi(inputData.Val)
	if float64(res) >= math.Exp2(31) {
		PrintError(inputData, "number too large")
	}
	if err3 != nil {
		PrintError(inputData, "cannot convert to number")
	}
	fmt.Print("number is: ")
	fmt.Print(inputData.Val)
}

func GetSym(inputData *InputData) {
	// []byte{13} is an invisible character that gets picked up from the input file.
	// ! is the delimiter to help identify the "do" from "while-do" statements.
	for inputData.Ch == "\n" || inputData.Ch == " " || inputData.Ch == "!" {
		GetChar(inputData)
	}
	if unicode.IsLetter([]rune(inputData.Ch)[0]) {
		IdentKeyword(inputData)
	} else if unicode.IsNumber([]rune(inputData.Ch)[0]) {
		Number(inputData)
	} else if inputData.Ch == "*" {
		GetChar(inputData)
		inputData.Sym = TIMES
	} else if inputData.Ch == "+" {
		GetChar(inputData)
		inputData.Sym = PLUS
	} else if inputData.Ch == "-" {
		GetChar(inputData)
		inputData.Sym = MINUS
	} else if inputData.Ch == "=" {
		GetChar(inputData)
		inputData.Sym = EQ
	} else if inputData.Ch == "<" {
		GetChar(inputData)
		if inputData.Ch == "=" {
			GetChar(inputData)
			inputData.Sym = LE
		} else if inputData.Ch == ">" {
			GetChar(inputData)
			inputData.Sym = NE
		} else {
			inputData.Sym = LT
		}
	} else if inputData.Ch == ">" {
		GetChar(inputData)
		if inputData.Ch == "=" {
			GetChar(inputData)
			inputData.Sym = GE
		} else {
			inputData.Sym = GT
		}
	} else if inputData.Ch == ";" {
		GetChar(inputData)
		inputData.Sym = SEMICOLON
	} else if inputData.Ch == "," {
		GetChar(inputData)
		inputData.Sym = COMMA
	} else if inputData.Ch == ":" {
		GetChar(inputData)
		inputData.Sym = COLON
	} else if inputData.Ch == "." {
		GetChar(inputData)
		inputData.Sym = PERIOD
	} else if inputData.Ch == "(" {
		GetChar(inputData)
		inputData.Sym = LPAREN
	} else if inputData.Ch == ")" {
		GetChar(inputData)
		inputData.Sym = RPAREN
	} else if inputData.Ch == "[" {
		GetChar(inputData)
		inputData.Sym = LBRAK
	} else if inputData.Ch == "]" {
		GetChar(inputData)
		inputData.Sym = RBRAK
	} else if inputData.Ch == "~" {
		inputData.Sym = EOF
	} else {
		PrintError(inputData, "illegal character")
		GetChar(inputData)
		inputData.Sym = 0
	}
	fmt.Print("\t")
	fmt.Println(inputData.Sym)
}

// Prints out an error and the line and pos it was found on
func PrintError(inputData *InputData, errorMsg string) {
	// if inputData.LastLine > inputData.ErrorLine || inputData.LastPos > inputData.ErrorPos {
	// 	fmt.Println("Error: line " + strconv.Itoa(inputData.LastLine) + ", pos " + strconv.Itoa(inputData.LastPos) + " " + errorMsg)
	// }
	fmt.Println("Error: line " + strconv.Itoa(inputData.LastLine) + ", pos " + strconv.Itoa(inputData.Pos) + " " + errorMsg)
	inputData.ErrorLine = inputData.LastLine
	inputData.ErrorPos = inputData.LastPos
	inputData.Error = true
}


/*
	Symbol table
*/
// Struct for data related to symbol table entries.
type SymTableEntry struct {
	entryType	string // should only ever be var, ref, const, type, proc, stdproc
	name		string // name of entry (e.g, x)
	tp			PrimitiveType // primitive type (if applicable)
	ctp			ComplexTypes // for more complicated types; for instance, some entries contain records
	lev			int // scope level
	val			int // the value of (if applicable)
	par 		[]string // list of parameters in a function (if applicable)
	size		int	// Memory required to represent the type (for Bool and Int)
	adr			int	// Address in memory
	offset		int // Offset for a given element in a record or array
}

// Makes a new entry for the symbol table.
func NewSymTableEntry(entryType string, name string, tp PrimitiveType, ctp ComplexTypes, lev int, val int, par []string) SymTableEntry{
	return SymTableEntry{
		entryType: 	entryType,
		name: 		name,
		tp:			tp,
		ctp: 		ctp,
		lev:		lev,
		val: 		val,
		par:		par,
		size:		0,
		adr:		0,
		offset:		0}
}

// Enum for the three allowed P0 primitive types.
type PrimitiveType string
const (
	Int 	PrimitiveType 	= "int"
	Bool 	PrimitiveType	= "bool"
	None 	PrimitiveType	= "none"
)

// Represents an array or record.
type ComplexTypes struct {
	entryType	string		// whether it's an array or record
	fields 		[]SymTableEntry	// used for storing the fields in a record
	base		PrimitiveType// the base type of an array
	lower		int			// lower bound of an array
	length		int			// length of an array
	size		int			// size of the type allowed in an array
}

// Define a complex type.
func NewComplexType(entryType string, fields []SymTableEntry, base PrimitiveType, lower int, length int, par []string) ComplexTypes{
	return ComplexTypes{
		entryType:	entryType,
		fields: 	fields,
		base:		base,
		lower:		lower,
		length: 	length,
		size:		0}
}

func PrintSymTable(inputData *InputData) {
	fmt.Println(inputData.SymTable)
}

//
// Add new symbol table entry.
//
func NewDecl(inputData *InputData, name string, entryType string){
	topLevel := inputData.SymTable[0]
	lev := len(topLevel) - 1

	for _, entry := range topLevel {
		if entry.name == name {
			PrintError(inputData, "multiple definitions")
			return
		}
	}

	inputData.SymTable[0] = append(inputData.SymTable[0], SymTableEntry{entryType: entryType, name: name, lev: lev, tp: Int})
}

func FindInSymTab(inputData *InputData, name string) SymTableEntry{
	for _, level := range inputData.SymTable {
		for _, entry := range level {
			if entry.name == name {
				return entry
			}
		}
	}
	PrintError(inputData, "undefined identifier " + name)
	return SymTableEntry{}
}

func OpenScope(inputData *InputData){
	inputData.SymTable = append([][]SymTableEntry{{}}, inputData.SymTable...)
}

func TopScope(inputData *InputData) []SymTableEntry {
	return inputData.SymTable[0]
}

func CloseScope(inputData *InputData){
	inputData.SymTable = inputData.SymTable[1:]
}


/*
	Grammar functions
*/

//
// Compiles the code into WASM.
//
func CompileWasm(srcfile string, inputData *InputData) {
	Init(inputData)
	Program(inputData)
}

//
// Helper function to check for that an element is in the 
// first and follow sets.
//
func ElementInSet(item int, set []int) bool {
	for _, i := range set {
		if i == item {
			return true
		}
	}
	return false
}

//
// Parses the "program" part of the grammar.
//
func Program(inputData *InputData) {
	fmt.Println(inputData.Sym)
	if inputData.Sym == PROGRAM {
		GetSym(inputData)
	} else {
		fmt.Println("test")
		PrintError(inputData, "'program' expected")
	}
	if inputData.Sym == IDENT {
		GetSym(inputData)
	} else {
		PrintError(inputData, "program name expected")
	}
	if inputData.Sym == SEMICOLON {
		GetSym(inputData)
	} else {
		PrintError(inputData, "; expected")
	}
}


/*
	WASM code generator.
*/

//
// Takes the asm string and converts it into a WASM code file with
// the provided filename.
//
func WriteWasmFile(filename string, inputData *InputData) {
	generatedCode := []byte(GenProgExit(inputData))
	err := ioutil.WriteFile(filename, generatedCode, 0644)

	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(filename + " was created.")
	}
}

func GenProgStart(inputData *InputData) {
	inputData.asm = append(inputData.asm, "(module")
}

func GenBool(entry *SymTableEntry) *SymTableEntry {
	entry.size = 1
	return entry
}

func GenInt(entry *SymTableEntry) *SymTableEntry {
	entry.size = 4
	return entry
}

func GenRec(entry *SymTableEntry) *SymTableEntry{
	s := 0
	for _, f := range entry.ctp.fields {
		f.offset = s
		s = s + f.size
	}
	entry.size = s
	return entry
}

func GenArray(entry *SymTableEntry) *SymTableEntry {
	entry.size = entry.ctp.length * entry.ctp.size
	return entry
}

func GenGlobalVars(scope []SymTableEntry, start int, inputData *InputData) {
	i := start
	for i < len(scope) {
		if scope[i].entryType == "var" {
			if scope[i].tp == Int || scope[i].tp == Bool {
				inputData.asm = append(inputData.asm, "(global $" + scope[i].name + " (mut i32) i32.const 0)")
			} else if scope[i].ctp.entryType == "array" || scope[i].ctp.entryType == "record" {
				scope[i].lev = -2
				scope[i].adr = inputData.memsize
				inputData.memsize = inputData.memsize + scope[i].size
			} else {
				PrintError(inputData, "WASM: type?")
			}
		}
	}
}

func GenLocalVars(scope []SymTableEntry, start int, inputData *InputData) PrimitiveType {
	i := start
	for i < len(scope) {
		if scope[i].entryType == "var" {
			if scope[i].tp == Int || scope[i].tp == Bool {
				inputData.asm = append(inputData.asm, "(local $" + scope[i].name + " i32)")
			} else if scope[i].ctp.entryType == "array" || scope[i].ctp.entryType == "record" {
				PrintError(inputData, "WASM: no local arrays, records")
			} else {
				PrintError(inputData, "WASM: type?")
			}
		}
	}

	return None
}

func loadItem(entry *SymTableEntry, inputData *InputData) {
	if entry.entryType == "var" {
		if entry.lev == 0 {
			inputData.asm = append(inputData.asm, "global.get $" + entry.name)
		} else if entry.lev == inputData.curlev {
			inputData.asm = append(inputData.asm, "local.get $" + entry.name)
		} else if entry.lev == -2 {
			inputData.asm = append(inputData.asm, "i32.const" + strconv.Itoa(entry.adr))
			inputData.asm = append(inputData.asm, "i32.load")
		} else if entry.lev != -1 {
			PrintError(inputData, "WASM: var level")
		} 
	} else if entry.entryType == "ref" {
		if entry.lev == -1 {
			inputData.asm = append(inputData.asm, "i32.load")
		} else if entry.lev == inputData.curlev {
			inputData.asm = append(inputData.asm, "local.get $" + entry.name)
			inputData.asm = append(inputData.asm, "i32.load")
		} else {
			PrintError(inputData, "WASM: ref level")
		}
	} else if entry.entryType == "const" {
		inputData.asm = append(inputData.asm, "i32.const " + strconv.Itoa(entry.val))
	}
}

func GenVar(entry *SymTableEntry, inputData *InputData) SymTableEntry {
	y := SymTableEntry{}

	if 0 < entry.lev && entry.lev < inputData.curlev {
		PrintError(inputData, "WASM: level")
	}
	if entry.entryType == "ref" {
		y = NewSymTableEntry("ref", entry.name, entry.tp, NewComplexType("", []SymTableEntry{}, None, int(0), int(0), []string{}), entry.lev, 0, nil)
	} else if entry.entryType == "var" {
		y = NewSymTableEntry("var", entry.name, entry.tp, NewComplexType("", []SymTableEntry{}, None, int(0), int(0), []string{}), entry.lev, 0, nil)
		if entry.lev == -2 {
			y.adr = entry.adr
		}
	}

	return y
}

func GenConst(entry *SymTableEntry) *SymTableEntry {
	return entry
}

func GenUnaryOp(op int, entry *SymTableEntry, inputData *InputData) *SymTableEntry{
	loadItem(entry, inputData)
	if op == MINUS {
		inputData.asm = append(inputData.asm, "i32.const -1")
		inputData.asm = append(inputData.asm, "i32.mul")
		entry.entryType = "var"
		entry.tp = Int
		entry.lev = -1
	} else if op == NOT {
		inputData.asm = append(inputData.asm, "i32.eqz")
		entry.tp = Bool
		entry.lev = -1
	} else if op == AND {
		inputData.asm = append(inputData.asm, "if (result i32)")
		entry.tp = Bool
		entry.lev = -1
	} else if op == OR {
		inputData.asm = append(inputData.asm, "if (result i32)")
		inputData.asm = append(inputData.asm, "i32.const 1")
		inputData.asm = append(inputData.asm, "else")
		entry.tp = Bool
		entry.lev = -1
	} else {
		PrintError(inputData, "WASM: unary operator?")
	}

	return entry
}

func GenBinaryOp(op int, x *SymTableEntry, y *SymTableEntry, inputData *InputData) *SymTableEntry{
	if op == PLUS || op == MINUS || op == TIMES || op == DIV || op == MOD {
		loadItem(x, inputData)
		loadItem(y, inputData)
		if op == PLUS {
			inputData.asm = append(inputData.asm, "i32.add")
		} else if op == MINUS {
			inputData.asm = append(inputData.asm, "i32.sub")
		} else if op == TIMES {
			inputData.asm = append(inputData.asm, "i32.mul")
		} else if op == DIV {
			inputData.asm = append(inputData.asm, "i32.div_s")
		} else if op == MOD {
			inputData.asm = append(inputData.asm, "i32.rem_s")
		} else {
			PrintError(inputData, "WASM: binary operator?")
		}
		x.tp = Int
	} else if op == AND {
		loadItem(y, inputData)
		inputData.asm = append(inputData.asm, "else")
		inputData.asm = append(inputData.asm, "i32.const 0")
		inputData.asm = append(inputData.asm, "end")
		x.tp = Bool
	} else if op == OR {
		loadItem(y, inputData)
		inputData.asm = append(inputData.asm, "end")
		x.tp = Bool
	}
	x.lev = -1
	return x
}

func GenRelation(op int, x *SymTableEntry, y *SymTableEntry, inputData *InputData) *SymTableEntry {
	loadItem(x, inputData)
	loadItem(y, inputData)
	if op == EQ {
		inputData.asm = append(inputData.asm, "i32.eq")
	} else if op == NE {
		inputData.asm = append(inputData.asm, "i32.ne")
	} else if op == LT {
		inputData.asm = append(inputData.asm, "i32.lt_s")
	} else if op == GT {
		inputData.asm = append(inputData.asm, "i32.gt_s")
	} else if op == LE {
		inputData.asm = append(inputData.asm, "i32.le_s")
	} else if op == GE {
		inputData.asm = append(inputData.asm, "i32.ge_s")
	}

	x.tp = Bool
	x.lev = -1
	return x
}

func GenSelect(entry *SymTableEntry, field *SymTableEntry, inputData *InputData) *SymTableEntry{
	if entry.entryType == "var"{
		entry.adr += field.offset
	} else if entry.entryType == "ref" {
		if entry.lev > 0 {
			inputData.asm = append(inputData.asm, "local.get $" + entry.name)
		}
		inputData.asm = append(inputData.asm, "i32.const " + strconv.Itoa(field.offset))
		inputData.asm = append(inputData.asm, "i32.add")
		entry.lev = -1
	}
	entry.tp = field.tp
	return entry
}

func GenIndex(x *SymTableEntry, y *SymTableEntry, inputData *InputData) {
	if x.entryType == "var" {
		if y.entryType == "const" {
			x.adr += (y.val - x.ctp.lower) * x.ctp.size
			x.tp = x.ctp.base
		} else {
			loadItem(y, inputData)
			if x.ctp.lower != 0 {
				inputData.asm = append(inputData.asm, "i32.const " + strconv.Itoa(x.ctp.lower))
				inputData.asm = append(inputData.asm, "i32.sub")
			}
			inputData.asm = append(inputData.asm, "i32.const " + strconv.Itoa(x.ctp.size))
			inputData.asm = append(inputData.asm, "i32.mul")
			inputData.asm = append(inputData.asm, "i32.const " + strconv.Itoa(x.adr))
			inputData.asm = append(inputData.asm, "i32.add")
			x.entryType = "ref"
			x.tp = x.ctp.base
		}
	} else {
		if x.lev == inputData.curlev {
			loadItem(x, inputData)
			x.lev = -1
		}
		if x.entryType == "const" {
			inputData.asm = append(inputData.asm, "i32.const " + strconv.Itoa((y.val - x.ctp.lower) * x.ctp.size))
			inputData.asm = append(inputData.asm, "i32.add")
		} else {
			loadItem(y, inputData)
			inputData.asm = append(inputData.asm, "i32.const " + strconv.Itoa(x.ctp.lower))
			inputData.asm = append(inputData.asm, "i32.sub")
			inputData.asm = append(inputData.asm, "i32.const " + strconv.Itoa(x.ctp.size))
			inputData.asm = append(inputData.asm, "i32.mul")
			inputData.asm = append(inputData.asm, "i32.add")
		}
	}
}

func GenAssign(x *SymTableEntry, y *SymTableEntry, inputData *InputData) {
	if x.entryType == "var" {
		if x.lev == -2 {
			inputData.asm = append(inputData.asm, "i32.const " + strconv.Itoa(x.adr))
		}
		loadItem(y, inputData)
		if x.lev == 0 {
			inputData.asm = append(inputData.asm, "global.set $ " + x.name)
		} else if x.lev == inputData.curlev {
			inputData.asm = append(inputData.asm, "local.set $ " + x.name)
		} else if x.lev == -2 {
			inputData.asm = append(inputData.asm, "i32.store")
		} else {
			PrintError(inputData, "WASM: level")
		}
	} else if x.entryType == "ref" {
		if x.lev == inputData.curlev {
			inputData.asm = append(inputData.asm, "local.get $" + x.name)
		}
		loadItem(y, inputData)
		inputData.asm = append(inputData.asm, "i32.store")
	}
}

func GenProgEntry(inputData *InputData) {
	inputData.asm = append(inputData.asm, "(func $program")
}

func GenProgExit(inputData *InputData) string{
	closingString := ")\n(memory " + strconv.Itoa(inputData.memsize / int(math.Exp2(16)) + 1) + ")\n(start $program)\n"
	inputData.asm = append(inputData.asm, closingString)
	outputCode := ""
	for _, asm := range inputData.asm {
		outputCode += "\n" + asm
	}

	return outputCode
}

func GenProcStart(ident string, listOfParams []SymTableEntry, inputData *InputData) {
	if inputData.curlev > 0 {
		PrintError(inputData, "WASM: no nested procedures")
	}
	inputData.curlev += 1
	params := ""

	for _, param := range listOfParams {
		params += "(param $" + param.name + " i32)"
	}

	inputData.asm = append(inputData.asm, "(func $" + ident + params)
}

func GenProcEntry(inputData *InputData) {
	//pass
}

func GenProcExit(inputData *InputData) {
	inputData.curlev -= 1
	inputData.asm = append(inputData.asm, ")")
}

func GenActualPara(ap *SymTableEntry, fp *SymTableEntry, inputData *InputData) {
	if ap.entryType == "ref" {
		if ap.lev == -2 {
			inputData.asm = append(inputData.asm, "i32.const " + strconv.Itoa(ap.adr))
		}
	} else if ap.entryType == "var" || ap.entryType == "ref" || ap.entryType == "const" {
		loadItem(ap, inputData)
	} else {
		PrintError(inputData, "unsupported parameter type")
	}
}

func GenCall(entry *SymTableEntry, inputData *InputData) {
	inputData.asm = append(inputData.asm, "call $" + entry.name)
}

func GenSeq(inputData *InputData) {
	//pass
}

func GenThen(x *SymTableEntry, inputData *InputData) *SymTableEntry {
	loadItem(x, inputData)
	inputData.asm = append(inputData.asm, "if")
	return x
}

func GenIfThen(inputData *InputData) {
	inputData.asm = append(inputData.asm, "end")
}

func GenElse(inputData *InputData) {
	inputData.asm = append(inputData.asm, "else")
}

func GenIfElse(inputData *InputData) {
	inputData.asm = append(inputData.asm, "end")
}

func GenWhile(inputData *InputData) {
	inputData.asm = append(inputData.asm, "loop")
}

func GenDo(x *SymTableEntry, inputData *InputData) *SymTableEntry {
	loadItem(x, inputData)
	inputData.asm = append(inputData.asm, "if")
	return x
}

func GenWhileDo(inputData *InputData) {
	inputData.asm = append(inputData.asm, "br 1")
	inputData.asm = append(inputData.asm, "end")
	inputData.asm = append(inputData.asm, "end")
}