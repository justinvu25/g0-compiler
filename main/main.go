package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

type InputData struct {
	Input      string            // P0 source cmd
	Sym        int               // Symbol that was identified.
	Ch         string            // Current character.
	Index      int               // Helps identify symbols
	Val        string            // Value
	LineNumber int               // Current line being parsed
	LastLine   int               // Previous line that was parsed
	ErrorLine  int               // Used to help surpress multiple errors
	Pos        int               // Current position of parser in a line
	LastPos    int               // Previous position
	ErrorPos   int               // Used to help surpress multiple errors
	Error      bool              // Set to true when an error is found.
	SymTable   [][]SymTableEntry // Symbol table of items that will be turned into WASM.
	Curlev     int               // Current scope level of the code generator.
	Memsize    int               // Size of the required memory allocation.
	Asm        []string          // The string that will ultimately become the WASM file.
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
		SymTable:   [][]SymTableEntry{{}},
		Curlev:     0,
		Memsize:    0,
		Asm:        []string{}}
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
	runtime.GOMAXPROCS(runtime.NumCPU())
	var wg sync.WaitGroup
	inputData := NewInputData("p0test.txt")
	// Add items to the wait group, one for each goroutine.
	wg.Add(3)
	start := time.Now()
	go EatWhiteSpace(inputData, &wg)
	go EatComments(inputData, &wg)
	go ParseInput(inputData, &wg)
	// Wait for the waitgroup counter to reach zero before continuing.
	// The waitgroup counter is decremented each time a thread finishes
	// executing its procedure.
	wg.Wait()
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Print("done all tasks in ")
	fmt.Println(elapsed)
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

var FIRSTFACTOR = map[int]int{IDENT:1, NUMBER:1, LPAREN:1, NOT:1}
var FOLLOWFACTOR = map[int]int{TIMES:1, DIV:1, MOD:1, AND:1, OR:1, PLUS:1, MINUS:1, EQ:1, NE:1, LT:1, LE:1, GT:1, GE:1,
	COMMA:1, SEMICOLON:1, THEN:1, ELSE:1, RPAREN:1, RBRAK:1, DO:1, PERIOD:1, END:1}
var FIRSTEXPRESSION = map[int]int{PLUS:1, MINUS:1, IDENT:1, NUMBER:1, LPAREN:1, NOT:1}
var FIRSTSTATEMENT = map[int]int{IDENT:1, IF:1, WHILE:1, BEGIN:1}
var FOLLOWSTATEMENT = map[int]int{SEMICOLON:1, END:1, ELSE:1}
var FIRSTTYPE = map[int]int{IDENT:1, RECORD:1, ARRAY:1, LPAREN:1}
var FOLLOWTYPE = map[int]int{SEMICOLON:1}
var FIRSTDECL = map[int]int{CONST:1, TYPE:1, VAR:1, PROCEDURE:1}
var FOLLOWDECL = map[int]int{BEGIN:1}
var FOLLOWPROCCALL = map[int]int{SEMICOLON:1, END:1, ELSE:1}
var STRONGSYMS = map[int]int{CONST:1, TYPE:1, VAR:1, PROCEDURE:1, WHILE:1, IF:1, BEGIN:1, EOF:1}

// Parses characters into tokens.
func ParseInput(inputData *InputData, wg *sync.WaitGroup) {
	defer wg.Done()
	CompileWasm(inputData)
	fmt.Println("done parsing")
}

// Removes whitespace.
func EatWhiteSpace(inputData *InputData, wg *sync.WaitGroup) {
	defer wg.Done()
	// Giving -1 to string.Replace removes an unlimited number of whitespaces.
	inputData.Input = strings.Replace(inputData.Input, " do", "!do", -1)
	inputData.Input = strings.Replace(inputData.Input, " end", "!end", -1)
	inputData.Input = strings.Replace(inputData.Input, " div", "!div", -1)
	inputData.Input = strings.Replace(inputData.Input, " mod", "!mod", -1)
	inputData.Input = strings.Replace(inputData.Input, " ", "", -1)
	inputData.Input = strings.Replace(inputData.Input, "\t", "", -1)
	inputData.Input = strings.Replace(inputData.Input, "\r", "", -1)

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
		inputData.LastPos = inputData.Pos
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
}

func GetSym(inputData *InputData) {
	// If any of these are detected, we can just move on.
	for inputData.Ch == "\n" || inputData.Ch == "\t" || inputData.Ch == "\r" || inputData.Ch == string([]byte{13})|| inputData.Ch == " " || inputData.Ch == "!" {
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
		if inputData.Ch == "=" {
			GetChar(inputData)
			inputData.Sym = BECOMES
		} else {
			inputData.Sym = COLON
		}
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
}

// Prints out an error and the line and pos it was found on
func PrintError(inputData *InputData, errorMsg string) {
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
	entryType string          // should only ever be var, ref, const, type, proc, stdproc, array, record
	name      string          // name of entry (e.g, x)
	tp        PrimitiveType   // primitive type (if applicable)
	ctp       ComplexType     // for more complicated types; for instance, some entries contain records
	lev       int             // scope level
	val       int             // the value of (if applicable)
	par       []SymTableEntry // list of parameters in a function (if applicable)
	size      int             // Memory required to represent the type (for Bool and Int)
	adr       int             // Address in memory
	offset    int             // Offset for a given element in a record or array
	arrOrRec  string		  // If applicable, is it an array or record
}

// Enum for the three allowed P0 primitive types.
type PrimitiveType string

const (
	Int      PrimitiveType = "int"
	Bool     PrimitiveType = "bool"
	None     PrimitiveType = "none"
	Nil      PrimitiveType = ""
	EmptyInt int           = -9999999999 //Go doesn't have null ints, so for simplicity I just put a big negative number.
)

// Represents an array or record.
type ComplexType struct {
	fields    []SymTableEntry // used for storing the fields in a record
	base      PrimitiveType   // the base type of an array
	lower     int             // lower bound of an array
	length    int             // length of an array
	size      int             // size of the type allowed in an array
}

// Generates var symbol table entries.
func Var(tp PrimitiveType) *SymTableEntry {
	e := &SymTableEntry{}
	e.entryType = "var"
	e.tp = tp
	return e
}

// Generates ref symbol table entries.
func Ref(tp PrimitiveType) *SymTableEntry {
	e := &SymTableEntry{}
	e.entryType = "ref"
	e.tp = tp
	return e
}

// Generates const symbol table entries.
func Const(tp PrimitiveType, val int) *SymTableEntry {
	e := &SymTableEntry{}
	e.entryType = "const"
	e.tp = tp
	e.val = val
	return e
}

// Generates type symbol table entries.
func Type(tp PrimitiveType) *SymTableEntry {
	e := &SymTableEntry{}
	e.entryType = "type"
	e.tp = tp
	e.val = EmptyInt
	return e
}

// Generates proc symbol table entries.
func Proc(par []SymTableEntry) *SymTableEntry {
	e := &SymTableEntry{}
	e.entryType = "proc"
	e.tp = None
	e.par = par
	return e
}

// Generates stdproc symbol table entries.
func StdProc(par []SymTableEntry) *SymTableEntry {
	e := &SymTableEntry{}
	e.entryType = "stdproc"
	e.tp = None
	e.par = par
	return e
}

// Generates record symbol table entries.
func Record(fields []SymTableEntry) *SymTableEntry {
	e := &SymTableEntry{}
	e.arrOrRec = "record"
	e.ctp = ComplexType{}
	e.ctp.fields = fields
	return e
}

// Generates array symbol table entries.
func Array(base PrimitiveType, lower int, length int) *SymTableEntry {
	e := &SymTableEntry{}
	e.arrOrRec = "array"
	e.ctp = ComplexType{}
	e.ctp.base = base
	e.ctp.lower = lower
	e.ctp.length = length
	return e
}

//Simply prints out the symbol table.
func PrintSymTable(inputData *InputData) {
	fmt.Println(inputData.SymTable)
}

// Add new symbol table entry.
func NewDecl(Name string, entry *SymTableEntry, inputData *InputData) {
	topLevel := inputData.SymTable[0]
	Lev := len(inputData.SymTable) - 1

	for _, entry := range topLevel {
		if entry.name == Name {
			PrintError(inputData, "multiple definitions")
			return
		}
	}

	entry.name = Name
	entry.lev = Lev
	inputData.SymTable[0] = append(inputData.SymTable[0], *entry)
}

// Checks the symbol table for an entry with the provided name.
func FindInSymTab(inputData *InputData, name string) *SymTableEntry {
	for _, level := range inputData.SymTable {
		for _, entry := range level {
			if entry.name == name {
				return &entry
			}
		}
	}
	PrintError(inputData, "undefined identifier " + name)
	return &SymTableEntry{}
}

// Opens a new scope in the symbol table by appended an empty symbol table to the front.
func OpenScope(inputData *InputData) {
	inputData.SymTable = append([][]SymTableEntry{{}}, inputData.SymTable...)
}

// Returns the top scope of the symbol table.
func TopScope(inputData *InputData) []SymTableEntry {
	return inputData.SymTable[0]
}

// Closes the top scope.
func CloseScope(inputData *InputData) {
	inputData.SymTable = inputData.SymTable[1:]
}

/*
	Grammar functions
*/
// Helper function to check for that an element is in the first and follow sets.
func exists(a int, dict map[int]int) bool {
	if _, ok := dict[a]; ok {
		return true
	}

	return false
}

// Generates selectors for records and arrays.
func selector(x *SymTableEntry, inputData *InputData) *SymTableEntry {
	for inputData.Sym == PERIOD || inputData.Sym == LBRAK {
		if inputData.Sym == PERIOD {
			GetSym(inputData)
			if inputData.Sym == IDENT {
				if x.arrOrRec == "record" {
					ctr := 0
					for _, recfield := range x.ctp.fields {
						if recfield.name == inputData.Val {
							x = GenSelect(x, &recfield, inputData)
							break
						}
						ctr += 1
					}
					if ctr == len(x.ctp.fields)-1 {
						PrintError(inputData, "not a field")
					}
					GetSym(inputData)
				} else {
					PrintError(inputData, "not a record")
				}
			} else {
				PrintError(inputData,"identifier expected")
			}
		} else { // x[y]
			GetSym(inputData)
			y := expression(inputData)
			if x.arrOrRec == "array" {
				if y.tp == Int {
					if y.entryType == "const" {

					} else {
						x = GenIndex(x, y, inputData)
					}
				}
			} else {
				PrintError(inputData, "not an array")
			}
			if inputData.Sym == RBRAK {
				GetSym(inputData)
			} else {
				PrintError(inputData, "] expected")
			}
		}
	}
	return x
}

// Generates factors.
func factor(inputData *InputData) *SymTableEntry {
	var x = &SymTableEntry{}
	if !exists(inputData.Sym, FIRSTFACTOR) {
		PrintError(inputData,"expression expected")
		for !(exists(inputData.Sym, FIRSTFACTOR) || exists(inputData.Sym, FOLLOWFACTOR) || exists(inputData.Sym, STRONGSYMS)) {
			GetSym(inputData)
		}
	}
	if inputData.Sym == IDENT {
		y := FindInSymTab(inputData, inputData.Val)
		x = y
		if x.entryType == "var" || x.entryType == "ref" {
			x = GenVar(x, inputData)
			GetSym(inputData)
		} else if x.entryType == "const" {
			x = Const(x.tp, x.val)
			x = GenConst(x)
			GetSym(inputData)
		} else {
			PrintError(inputData,"expression expected")
		}
		x = selector(x, inputData)
	} else if inputData.Sym == NUMBER {
		constVal, err := strconv.Atoi(inputData.Val)
		x = Const(Int, constVal)
		x = GenConst(x)
		if err != nil {
			PrintError(inputData, "error converting number")
		}
		GetSym(inputData)
	} else if inputData.Sym == LPAREN {
		GetSym(inputData)
		x = expression(inputData)
		if inputData.Sym == RPAREN {
			GetSym(inputData)
		} else {
			PrintError(inputData,") expected")
		}
	} else if inputData.Sym == NOT {
		GetSym(inputData)
		x = factor(inputData)
		if x.tp != Bool {
			PrintError(inputData,"not boolean")
		} else if x.entryType == "const" {
			x.val = 1 - x.val
		} else {
			x = GenUnaryOp(NOT, x, inputData)
		}
	} else {
		x = Const(None, 0)
	}
	return x
}

// Generates terms.
func term(inputData *InputData) *SymTableEntry {
	x := factor(inputData)
	for inputData.Sym == TIMES || inputData.Sym == DIV || inputData.Sym == MOD || inputData.Sym == AND {
		op := inputData.Sym
		GetSym(inputData)
		if op == AND && x.entryType != "const" {
			x = GenUnaryOp(AND, x, inputData)
		}
		y := factor(inputData)
		if (x.tp == Int && y.tp == Int) && exists(op, map[int]int{TIMES:1, DIV:1, MOD:1}) {
			if x.entryType == "const" && y.entryType == "const" {
				if op == TIMES {
					x.val = x.val * y.val
				} else if op == DIV {
					x.val = x.val / y.val
				} else if op == MOD {
					x.val = x.val % y.val
				}
			} else {
				x = GenBinaryOp(op, x, y, inputData)
			}
		} else if (x.tp == Bool && y.tp == Bool) && op == AND {
			if x.entryType == "const" {
				if x.val != EmptyInt { // Since Go doesn't provide a good way to check for empty int, I used a massively negative number
					x = y
				}
			} else {
				x = GenBinaryOp(AND, x, y, inputData)
			}
		} else {
			PrintError(inputData, "bad type")
		}
	}
	return x
}

// Generates simple expressions.
func simpleExpression(inputData *InputData) *SymTableEntry {
	x := &SymTableEntry{}
	if inputData.Sym == PLUS {
		GetSym(inputData)
		x = term(inputData)
	} else if inputData.Sym == MINUS {
		GetSym(inputData)
		x = term(inputData)
		if x.tp != Int {
			PrintError(inputData,"bad type")
		} else if x.entryType == "const" {
			x.val = -1 * x.val
		} else {
			x = GenUnaryOp(MINUS, x, inputData)
		}
	} else {
		x = term(inputData)
	}
	for inputData.Sym == PLUS || inputData.Sym == MINUS || inputData.Sym == OR {
		op := inputData.Sym
		GetSym(inputData)
		if op == OR && x.entryType != "const" {
			x = GenUnaryOp(OR, x, inputData)
		}
		y := term(inputData)
		if (x.tp == Int && y.tp == Int) && (op == PLUS || op == MINUS) {
			if x.entryType == "const" && y.entryType == "const" {
				if op == PLUS {
					x.val = x.val + y.val
				} else if op == MINUS {
					x.val = x.val - y.val
				}
			} else {
				x = GenBinaryOp(op, x, y, inputData)
			}
		} else if x.tp == Bool && y.tp == Bool && op == OR {
			if x.entryType == "const" {
				if x.val != EmptyInt {
					x = y
				} else {
					x = GenBinaryOp(OR, x, y, inputData)
				}
			}
		} else {
			PrintError(inputData, "Bad type")
		}
	}
	return x
}

// Generates whole expressions.
func expression(inputData *InputData) *SymTableEntry {
	x := simpleExpression(inputData)
	for inputData.Sym == EQ || inputData.Sym == NE || inputData.Sym == LT || inputData.Sym == LE || inputData.Sym == GT || inputData.Sym == GE {
		op := inputData.Sym
		GetSym(inputData)
		y := simpleExpression(inputData)

		if x.tp == y.tp {
			if x.entryType == "const" && y.entryType == "const" {
				if op == EQ {
					if x.val == y.val {
						x.val = 1
					} else {
						x.val = 0
					}
				} else if op == NE {
					if x.val != y.val {
						x.val = 1
					} else {
						x.val = 0
					}
				} else if op == LT {
					if x.val < y.val {
						x.val = 1
					} else {
						x.val = 0
					}
				} else if op == LE {
					if x.val <= y.val {
						x.val = 1
					} else {
						x.val = 0
					}
				} else if op == GT {
					if x.val > y.val {
						x.val = 1
					} else {
						x.val = 0
					}
				} else if op == GE {
					if x.val >= y.val {
						x.val = 1
					} else {
						x.val = 0
					}
				}
				x.tp = Bool
			} else {
				x = GenRelation(op, x, y, inputData)
			}
		} else {
			PrintError(inputData, "bad type")
		}
	}

	return x
}

// Generates compound statements.
func compoundStatement(inputData *InputData) *SymTableEntry {
	if inputData.Sym == BEGIN {
		GetSym(inputData)
	} else {
		PrintError(inputData, "'begin' expected")
	}
	x := statement(inputData)
	for inputData.Sym == SEMICOLON || exists(inputData.Sym, FIRSTSTATEMENT) {
		if inputData.Sym == SEMICOLON {
			GetSym(inputData)
		} else {
			PrintError(inputData, "; missing")
		}
		y := statement(inputData)
		GenSeq(x, y, inputData)
	}
	if inputData.Sym == END {
		GetSym(inputData)
	} else {
		PrintError(inputData, "'end' expected")
	}
	return x
}

// Generates statements.
func statement(inputData *InputData) *SymTableEntry {
	x := &SymTableEntry{}
	y := &SymTableEntry{}
	if !exists(inputData.Sym, FIRSTSTATEMENT) {
		PrintError(inputData, "statement expected")
		GetSym(inputData)
		for !(exists(inputData.Sym, FIRSTFACTOR) || exists(inputData.Sym, FOLLOWSTATEMENT) || exists(inputData.Sym, STRONGSYMS)){
			GetSym(inputData)
		}
	}
	if inputData.Sym == IDENT {
		x = FindInSymTab(inputData, inputData.Val)
		GetSym(inputData)
		if x.entryType == "var" || x.entryType == "ref" {
			x = GenVar(x, inputData)
			x = selector(x, inputData)
			if inputData.Sym == BECOMES {
				GetSym(inputData)
				y = expression(inputData)
				if x.tp == Bool || x.tp == Int || y.tp == Bool || y.tp == Int {
					GenAssign(x, y, inputData)
				} else {
					PrintError(inputData, "incompatible assignment")
				}
			} else if inputData.Sym == EQ {
				PrintError(inputData,":= expected")
				GetSym(inputData)
				y = expression(inputData)
				GenSeq(y, y, inputData) // THIS IS ONLY CALLED BECAUSE GO DOESN'T LIKE UNUSED DECLARATIONS
			} else {
				PrintError(inputData, ":= expected")
			}
		} else if x.entryType == "proc" || x.entryType == "stdproc" {
			fp := x.par
			var ap []*SymTableEntry
			i := 0
			if inputData.Sym == LPAREN {
				GetSym(inputData)
				if exists(inputData.Sym, FIRSTEXPRESSION) {
					y = expression(inputData)
					if i < len(fp) {
						if (y.entryType == "var" || fp[i].entryType == "var") && (fp[i].entryType == y.entryType) {
							if x.entryType == "proc" {
								ap = append(ap, GenActualPara(y, &fp[i], inputData))
							}
						} else if x.name != "read" {
							PrintError(inputData, "illegal parameter mode")
						}
					} else {
						PrintError(inputData, "extra parameter")
					}
					i++
					for inputData.Sym == COMMA {
						GetSym(inputData)
						y = expression(inputData)
						if i < len(fp) {
							if (y.entryType == "var" || fp[i].entryType == "var") && (fp[i].entryType == y.entryType) {
								if x.entryType == "proc" {
									ap = append(ap, GenActualPara(y, &fp[i], inputData))
								}
							} else {
								PrintError(inputData, "illegal parameter mode")
							}
						} else {
							PrintError(inputData, "extra parameter")
						}
						i++
					}
				}
				if inputData.Sym == RPAREN {
					GetSym(inputData)
				} else {
					PrintError(inputData, ") expected")
				}
			}
			if i < len(fp) {
				PrintError(inputData, "too few parameters")
			} else if x.entryType == "stdproc" {
				if x.name == "read" {
					GenRead(y, inputData)
				} else if x.name == "write" {
					GenWrite(y, inputData)
				} else if x.name == "writeln" {
					GenWriteln(inputData)
				}
			} else {
				x = GenCall(x, inputData)
			}
		} else {
			PrintError(inputData, "variable or procedure expected")
		}
	} else if inputData.Sym == BEGIN {
		x = compoundStatement(inputData)
	} else if inputData.Sym == IF {
		GetSym(inputData)
		x := expression(inputData)
		if x.tp == Bool {
			x = GenThen(x, inputData)
		} else {
			PrintError(inputData, "boolean expected")
		}
		if inputData.Sym == THEN {
			GetSym(inputData)
		} else {
			PrintError(inputData,"'then' expected")
		}
		y = statement(inputData)
		if inputData.Sym == ELSE {
			if x.tp == Bool {
				y = GenElse(x, y, inputData)
			}
			GetSym(inputData)
			z:= statement(inputData)
			if x.tp == Bool {
				x = GenIfElse(x, y, z, inputData)
			}
		} else {
			if x.tp == Bool {
				x = GenIfThen(x, y, inputData)
			}
		}
	} else if inputData.Sym == WHILE {
		GetSym(inputData)
		GenWhile(inputData)
		x = expression(inputData)
		if x.tp == Bool {
			x = GenDo(x, inputData)
		} else {
			PrintError(inputData, "boolean expected")
		}
		if inputData.Sym == DO {
			GetSym(inputData)
		} else {
			PrintError(inputData, "'do' expected")
		}
		y = statement(inputData)
		if x.tp == Bool {
			GenWhileDo(x, y, inputData)
		}
	} else {
		x = nil
	}
	return x
}

// Generates the type of an identifier.
func typ(inputData *InputData) *SymTableEntry {
	x := &SymTableEntry{}
	if !exists(inputData.Sym, FIRSTTYPE) {
		PrintError(inputData, "type expected")
		for !(exists(inputData.Sym, FIRSTFACTOR) || exists(inputData.Sym, FOLLOWFACTOR) || exists(inputData.Sym, STRONGSYMS)) {
			GetSym(inputData)
		}
	}
	if inputData.Sym == IDENT {
		ident := inputData.Val
		x = FindInSymTab(inputData, ident)
		GetSym(inputData)
		if x.entryType == "type" {
			x = Type(x.tp)
		} else {
			PrintError(inputData, "not a type")
			x = Type(None)
		}
	} else if inputData.Sym == ARRAY {
		GetSym(inputData)
		if inputData.Sym == LBRAK {
			GetSym(inputData)
		} else {
			PrintError(inputData, "'[' expected")
		}
		x = expression(inputData)
		if inputData.Sym == PERIOD {
			GetSym(inputData)
		} else {
			PrintError(inputData,"'.' expected")
		}
		if inputData.Sym == PERIOD {
			GetSym(inputData)
		} else {
			PrintError(inputData,"'.' expected")
		}
		y := expression(inputData)
		if inputData.Sym == RBRAK {
			GetSym(inputData)
		} else {
			PrintError(inputData,"']' expected")
		}
		if inputData.Sym == OF {
			GetSym(inputData)
		} else {
			PrintError(inputData,"of expected")
		}
		z := typ(inputData).tp
		if x.entryType != "const" || x.val < 0 {
			PrintError(inputData,"bad lower bound")
			x = Type(None)
		} else if y.entryType != "const" || y.val < 0 {
			PrintError(inputData,"bad upper bound")
			x = Type(None)
		} else {
			arr := Array(z, x.val, y.val - x.val + 1)
			x = GenArray(arr)
			x.tp = z
		}
	} else if inputData.Sym == RECORD {
		GetSym(inputData)
		OpenScope(inputData)
		typedIds("var", inputData)
		for {
			if inputData.Sym == SEMICOLON {
				GetSym(inputData)
				typedIds("var", inputData)
			} else {
				break
			}
		}
		if inputData.Sym == END {
			GetSym(inputData)
		} else {
			PrintError(inputData,"'end' expected")
		}
		r := TopScope(inputData)
		CloseScope(inputData)
		rec := Record(r)
		x = GenRec(rec)
		x.tp = Nil
	}

	return x
}

// Helps generate typed identifiers.
func typedIds(entryType string, inputData *InputData) {
	tid := []string{}
	if inputData.Sym == IDENT {
		tid = append([]string{inputData.Val}, tid...)
		GetSym(inputData)
	} else {
		PrintError(inputData,"identifier expected")
	}

	for inputData.Sym == COMMA {
		GetSym(inputData)
		if inputData.Sym == IDENT {
			tid = append(tid, inputData.Val)
			GetSym(inputData)
		} else {
			PrintError(inputData, "identifier expected")
		}
	}

	if inputData.Sym == COLON {
		GetSym(inputData)
		tp := typ(inputData)
		if tp != (&SymTableEntry{}) {
			for i := 0; i < len(tid); i++ {
				if entryType == "var" {
					if tp.arrOrRec == "array" || tp.arrOrRec == "record" {
						tp.entryType =  entryType
						NewDecl(tid[i], tp, inputData)
					} else {
						v := Var(tp.tp)
						NewDecl(tid[i], v, inputData)
					}
				} else if entryType == "ref" {
					if tp.arrOrRec == "array" || tp.arrOrRec == "record" {
						tp.entryType =  entryType
						NewDecl(tid[i], tp, inputData)
					} else {
						r := Ref(tp.tp)
						NewDecl(tid[i], r, inputData)
					}
				}
			}
		}
	} else {
		PrintError(inputData,": expected")
	}
}

// Generates various declarations.
func declaration(allocVarLevel string, inputData *InputData) {
	if !(exists(inputData.Sym, FIRSTDECL) || exists(inputData.Sym, FOLLOWDECL)) {
		PrintError(inputData, "'begin' or declaration expected")
		for !(exists(inputData.Sym, FIRSTDECL) || exists(inputData.Sym, FOLLOWDECL) || exists(inputData.Sym, STRONGSYMS)) {
			GetSym(inputData)
		}
	}
	for inputData.Sym == CONST {
		GetSym(inputData)
		if inputData.Sym == IDENT {
			ident := inputData.Val
			GetSym(inputData)
			if inputData.Sym == EQ {
				GetSym(inputData)
			} else {
				PrintError(inputData,"= expected")
			}
			x := expression(inputData)
			if x.entryType == "const" {
				c := Const(x.tp, x.val)
				NewDecl(ident, c, inputData)
			} else {
				PrintError(inputData,"expression not constant")
			}
		} else {
			PrintError(inputData,"constant name expected")
		}
		if inputData.Sym == SEMICOLON {
			GetSym(inputData)
		} else {
			PrintError(inputData, "; expected")
		}
	}
	for inputData.Sym == TYPE {
		GetSym(inputData)
		if inputData.Sym == IDENT {
			ident := inputData.Val
			GetSym(inputData)
			if inputData.Sym == EQ {
				GetSym(inputData)
			} else {
				PrintError(inputData,"= expected")
			}
			x := typ(inputData)
			NewDecl(ident, x, inputData)
			if inputData.Sym == SEMICOLON {
				GetSym(inputData)
			} else {
				PrintError(inputData, "; expected")
			}
		} else {
			PrintError(inputData,"type name expected")
		}
	}
	start := len(TopScope(inputData))
	for inputData.Sym == VAR {
		GetSym(inputData)
		typedIds("var", inputData)
		if inputData.Sym == SEMICOLON {
			GetSym(inputData)
		} else {
			PrintError(inputData,"; expected")
		}
	}
	if allocVarLevel == "global" {
		GenGlobalVars(TopScope(inputData), start, inputData)
	} else {
		GenLocalVars(TopScope(inputData), start, inputData)
	}
	for inputData.Sym == PROCEDURE {
		GetSym(inputData)
		if inputData.Sym == IDENT {
			GetSym(inputData)
		} else {
			PrintError(inputData,"procedure named expected")
		}
		ident := inputData.Val
		NewDecl(ident, Proc([]SymTableEntry{}), inputData)
		sc := TopScope(inputData)
		OpenScope(inputData)
		if inputData.Sym == LPAREN {
			GetSym(inputData)
			if inputData.Sym == VAR || inputData.Sym == IDENT {
				if inputData.Sym == VAR {
					GetSym(inputData)
					typedIds("ref", inputData)
				} else {
					typedIds("var", inputData)
				}
				for inputData.Sym == SEMICOLON {
					GetSym(inputData)
					if inputData.Sym == VAR {
						GetSym(inputData)
						typedIds("ref", inputData)
					} else {
						typedIds("var", inputData)
					}
				}
			} else {
				PrintError(inputData,"formal parameters expected")
			}
			fp := TopScope(inputData)
			sc[len(sc) - 1].par = fp
			if inputData.Sym == RPAREN {
				GetSym(inputData)
			} else {
				PrintError(inputData,") expected")
			}
		} else {
			fp := []SymTableEntry{}
			GenProcStart(ident, fp, inputData)
		}
		if inputData.Sym == SEMICOLON {
			GetSym(inputData)
		} else {
			PrintError(inputData,"; expected")
		}
		declaration("local", inputData)
		GenProcEntry(inputData)
		x := compoundStatement(inputData)
		GenProcExit(x, inputData)
		if inputData.Sym == SEMICOLON {
			GetSym(inputData)
		} else {
			PrintError(inputData,"; expected")
		}
	}
}

// Parses the "program" part of the grammar.
func Program(inputData *InputData) string {
	NewDecl("boolean", GenBool(Type(Bool)), inputData)
	NewDecl("integer", GenInt(Type(Int)), inputData)
	NewDecl("true", Const(Bool, 1), inputData)
	NewDecl("false", Const(Bool, 0), inputData)
	NewDecl("read", StdProc([]SymTableEntry{*Ref(Int)}), inputData)
	NewDecl("write", StdProc([]SymTableEntry{*Var(Int)}), inputData)
	NewDecl("writeln", StdProc([]SymTableEntry{}), inputData)
	GenProgStart(inputData)
	if inputData.Sym == PROGRAM {
		GetSym(inputData)
	} else {
		PrintError(inputData, "'program' expected")
	}
	ident := inputData.Val
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
	declaration("global", inputData)
	GenProgEntry(ident, inputData)
	x := compoundStatement(inputData)
	return GenProgExit(x, inputData)
}

// Compiles the code into WASM.
func CompileWasm(inputData *InputData) {
	Init(inputData)
	p := Program(inputData)
	WriteWasmFile("result.wasm", p)
}

/*
	WASM code generator.
*/
// Takes the asm string and converts it into a WASM code file with
// the provided filename.
func WriteWasmFile(filename string, code string) {
	generatedCode := []byte(code)
	err := ioutil.WriteFile(filename, generatedCode, 0644)

	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(filename + " was created.")
	}
}

// Generates the start of programs.
func GenProgStart(inputData *InputData) {
	inputData.Asm = append(inputData.Asm, "(module",
		"(import \"P0lib\" \"write\" (func $write (param i32)))",
		"(import \"P0lib\" \"writeln\" (func $writeln))",
		"(import \"P0lib\" \"read\" (func $read (result i32)))")
}

// Specifies the size of bool typed entries.
func GenBool(entry *SymTableEntry) *SymTableEntry {
	entry.size = 1
	return entry
}

// Specifies the size of int typed entries.
func GenInt(entry *SymTableEntry) *SymTableEntry {
	entry.size = 4
	return entry
}

// Generates records, calculating some of the attribute values.
func GenRec(entry *SymTableEntry) *SymTableEntry {
	s := 0
	for _, f := range entry.ctp.fields {
		f.offset = s
		s = s + f.size
	}
	entry.size = s
	return entry
}

// Generates records, calculating its size.
func GenArray(entry *SymTableEntry) *SymTableEntry {
	entry.size = entry.ctp.length * entry.ctp.size
	return entry
}

// Generates all of the global.
func GenGlobalVars(scope []SymTableEntry, start int, inputData *InputData) {
	i := start
	for i < len(scope) {
		if scope[i].entryType == "var" {
			if scope[i].tp == Int || scope[i].tp == Bool {
				inputData.Asm = append(inputData.Asm, "(global $"+scope[i].name+" (mut i32) i32.const 0)")
			} else if scope[i].entryType == "array" || scope[i].entryType == "record" {
				scope[i].lev = -2
				scope[i].adr = inputData.Memsize
				inputData.Memsize = inputData.Memsize + scope[i].size
			} else {
				PrintError(inputData, "WASM: type?")
			}
		}
		i += 1
	}
}

// Generates all of the local vars.
func GenLocalVars(scope []SymTableEntry, start int, inputData *InputData) PrimitiveType {
	i := start
	for i < len(scope) {
		if scope[i].entryType == "var" {
			if scope[i].tp == Int || scope[i].tp == Bool {
				inputData.Asm = append(inputData.Asm, "(local $"+scope[i].name+" i32)")
			} else if scope[i].entryType == "array" || scope[i].entryType == "record" {
				PrintError(inputData, "WASM: no local arrays, records")
			} else {
				PrintError(inputData, "WASM: type?")
			}
		}
		i += 1
	}

	return None
}

// Loads a sym table entry onto the stack.
func loadItem(entry *SymTableEntry, inputData *InputData) {
	if entry.entryType == "var" {
		if entry.lev == 0 {
			inputData.Asm = append(inputData.Asm, "global.get $"+entry.name)
		} else if entry.lev == inputData.Curlev {
			inputData.Asm = append(inputData.Asm, "local.get $"+entry.name)
		} else if entry.lev == -2 {
			inputData.Asm = append(inputData.Asm, "i32.const"+strconv.Itoa(entry.adr))
			inputData.Asm = append(inputData.Asm, "i32.load")
		} else if entry.lev != -1 {
			PrintError(inputData, "WASM: var level")
		}
	} else if entry.entryType == "ref" {
		if entry.lev == -1 {
			inputData.Asm = append(inputData.Asm, "i32.load")
		} else if entry.lev == inputData.Curlev {
			inputData.Asm = append(inputData.Asm, "local.get $"+entry.name)
			inputData.Asm = append(inputData.Asm, "i32.load")
		} else {
			PrintError(inputData, "WASM: ref level")
		}
	} else if entry.entryType == "const" {
		inputData.Asm = append(inputData.Asm, "i32.const "+strconv.Itoa(entry.val))
	}
}

// Generates a var using the provided symbol table entry.
func GenVar(entry *SymTableEntry, inputData *InputData) *SymTableEntry {
	y := &SymTableEntry{}

	if 0 < entry.lev && entry.lev < inputData.Curlev {
		PrintError(inputData, "WASM: level")
	}
	if entry.entryType == "ref" {
		y = Ref(entry.tp)
		y.lev = entry.lev
		y.name = entry.name
	} else if entry.entryType == "var" {
		y = Var(entry.tp)
		y.lev = entry.lev
		y.name = entry.name
		if entry.lev == -2 {
			y.adr = entry.adr
		}
	}

	if entry.arrOrRec == "array" || entry.arrOrRec == "record" {
		y.ctp = entry.ctp
		y.arrOrRec = entry.arrOrRec
	}

	return y
}

// Constants are simply constants so they do not need any extra work.
func GenConst(entry *SymTableEntry) *SymTableEntry {
	return entry
}

// Generates code for operations with unary operators.
func GenUnaryOp(op int, entry *SymTableEntry, inputData *InputData) *SymTableEntry {
	loadItem(entry, inputData)
	if op == MINUS {
		inputData.Asm = append(inputData.Asm, "i32.const -1")
		inputData.Asm = append(inputData.Asm, "i32.mul")
		entry.entryType = "var"
		entry.tp = Int
		entry.lev = -1
	} else if op == NOT {
		inputData.Asm = append(inputData.Asm, "i32.eqz")
		entry.tp = Bool
		entry.lev = -1
	} else if op == AND {
		inputData.Asm = append(inputData.Asm, "if (result i32)")
		entry.tp = Bool
		entry.lev = -1
	} else if op == OR {
		inputData.Asm = append(inputData.Asm, "if (result i32)")
		inputData.Asm = append(inputData.Asm, "i32.const 1")
		inputData.Asm = append(inputData.Asm, "else")
		entry.tp = Bool
		entry.lev = -1
	} else {
		PrintError(inputData, "WASM: unary operator?")
	}

	return entry
}

// Generates code for operations with binary operators.
func GenBinaryOp(op int, x *SymTableEntry, y *SymTableEntry, inputData *InputData) *SymTableEntry {
	if op == PLUS || op == MINUS || op == TIMES || op == DIV || op == MOD {
		loadItem(x, inputData)
		loadItem(y, inputData)
		if op == PLUS {
			inputData.Asm = append(inputData.Asm, "i32.add")
		} else if op == MINUS {
			inputData.Asm = append(inputData.Asm, "i32.sub")
		} else if op == TIMES {
			inputData.Asm = append(inputData.Asm, "i32.mul")
		} else if op == DIV {
			inputData.Asm = append(inputData.Asm, "i32.div_s")
		} else if op == MOD {
			inputData.Asm = append(inputData.Asm, "i32.rem_s")
		} else {
			PrintError(inputData, "WASM: binary operator?")
		}
		x = Var(Int)
		x.lev = -1
	} else if op == AND {
		loadItem(y, inputData)
		inputData.Asm = append(inputData.Asm, "else")
		inputData.Asm = append(inputData.Asm, "i32.const 0")
		inputData.Asm = append(inputData.Asm, "end")
		x = Var(Bool)
		x.lev = -1
	} else if op == OR {
		loadItem(y, inputData)
		inputData.Asm = append(inputData.Asm, "end")
		x = Var(Bool)
		x.lev = -1
	}
	return x
}

// Generates relations between two entries, such as x > 5.
func GenRelation(op int, x *SymTableEntry, y *SymTableEntry, inputData *InputData) *SymTableEntry {
	loadItem(x, inputData)
	loadItem(y, inputData)
	if op == EQ {
		inputData.Asm = append(inputData.Asm, "i32.eq")
	} else if op == NE {
		inputData.Asm = append(inputData.Asm, "i32.ne")
	} else if op == LT {
		inputData.Asm = append(inputData.Asm, "i32.lt_s")
	} else if op == GT {
		inputData.Asm = append(inputData.Asm, "i32.gt_s")
	} else if op == LE {
		inputData.Asm = append(inputData.Asm, "i32.le_s")
	} else if op == GE {
		inputData.Asm = append(inputData.Asm, "i32.ge_s")
	}

	x = Var(Bool)
	x.lev = -1
	return x
}

// Generates selectors for records.
func GenSelect(entry *SymTableEntry, field *SymTableEntry, inputData *InputData) *SymTableEntry {
	if entry.entryType == "var" {
		entry.adr += field.offset
	} else if entry.entryType == "ref" {
		if entry.lev > 0 {
			inputData.Asm = append(inputData.Asm, "local.get $"+entry.name)
		}
		inputData.Asm = append(inputData.Asm, "i32.const "+strconv.Itoa(field.offset))
		inputData.Asm = append(inputData.Asm, "i32.add")
		entry.lev = -1
	}
	entry.tp = field.tp
	return entry
}

// Generates indexes for arrays.
func GenIndex(x *SymTableEntry, y *SymTableEntry, inputData *InputData) *SymTableEntry {
	if x.entryType == "var" {
		if y.entryType == "const" {
			x.adr += (y.val - x.ctp.lower) * x.ctp.size
			x.tp = x.ctp.base
		} else {
			loadItem(y, inputData)
			if x.ctp.lower != 0 {
				inputData.Asm = append(inputData.Asm, "i32.const "+strconv.Itoa(x.ctp.lower))
				inputData.Asm = append(inputData.Asm, "i32.sub")
			}
			inputData.Asm = append(inputData.Asm, "i32.const "+strconv.Itoa(x.ctp.size))
			inputData.Asm = append(inputData.Asm, "i32.mul")
			inputData.Asm = append(inputData.Asm, "i32.const "+strconv.Itoa(x.adr))
			inputData.Asm = append(inputData.Asm, "i32.add")
			x.entryType = "ref"
			x.tp = x.ctp.base
		}
	} else {
		if x.lev == inputData.Curlev {
			loadItem(x, inputData)
			x.lev = -1
		}
		if x.entryType == "const" {
			inputData.Asm = append(inputData.Asm, "i32.const "+strconv.Itoa((y.val-x.ctp.lower)*x.ctp.size))
			inputData.Asm = append(inputData.Asm, "i32.add")
		} else {
			loadItem(y, inputData)
			inputData.Asm = append(inputData.Asm, "i32.const "+strconv.Itoa(x.ctp.lower))
			inputData.Asm = append(inputData.Asm, "i32.sub")
			inputData.Asm = append(inputData.Asm, "i32.const "+strconv.Itoa(x.ctp.size))
			inputData.Asm = append(inputData.Asm, "i32.mul")
			inputData.Asm = append(inputData.Asm, "i32.add")
		}
	}

	return x
}

// Generates assignment to variables.
func GenAssign(x *SymTableEntry, y *SymTableEntry, inputData *InputData) {
	if x.entryType == "var" {
		if x.lev == -2 {
			inputData.Asm = append(inputData.Asm, "i32.const " + strconv.Itoa(x.adr))
		}
		loadItem(y, inputData)
		if x.lev == 0 {
			inputData.Asm = append(inputData.Asm, "global.set $" + x.name)
		} else if x.lev == inputData.Curlev {
			inputData.Asm = append(inputData.Asm, "local.set $ " + x.name)
		} else if x.lev == -2 {
			inputData.Asm = append(inputData.Asm, "i32.store")
		} else {
			PrintError(inputData, "WASM: level")
		}
	} else if x.entryType == "ref" {
		if x.lev == inputData.Curlev {
			inputData.Asm = append(inputData.Asm, "local.get $" + x.name)
		}
		loadItem(y, inputData)
		inputData.Asm = append(inputData.Asm, "i32.store")
	}
}

// Generates the entry to the program.
func GenProgEntry(ident string, inputData *InputData) {
	inputData.Asm = append(inputData.Asm, "(func $program")
}

// Generates the exit to the program.
func GenProgExit(x *SymTableEntry, inputData *InputData) string {
	closingString := ")\n(memory " + strconv.Itoa(inputData.Memsize/int(math.Exp2(16))+1) + ")\n(start $program)\n)"
	inputData.Asm = append(inputData.Asm, closingString)
	outputCode := ""
	for _, asm := range inputData.Asm {
		outputCode += "\n" + asm
	}

	return outputCode
}

// Generates function signatures.
func GenProcStart(ident string, listOfParams []SymTableEntry, inputData *InputData) {
	if inputData.Curlev > 0 {
		PrintError(inputData, "WASM: no nested procedures")
	}
	inputData.Curlev += 1
	params := ""

	for _, param := range listOfParams {
		params += "(param $" + param.name + " i32)"
	}

	inputData.Asm = append(inputData.Asm, "(func $"+ident+params)
}

// Dummy function for generating procedure entries.
func GenProcEntry(inputData *InputData) {
	//pass
}

// Generates procedure exits, which is simply a closing parenthesis.
func GenProcExit(x *SymTableEntry, inputData *InputData) {
	inputData.Curlev -= 1
	inputData.Asm = append(inputData.Asm, ")")
}

// Generates the actual parameters using the provided formal parameters.
func GenActualPara(ap *SymTableEntry, fp *SymTableEntry, inputData *InputData) *SymTableEntry {
	if fp.entryType == "ref" {
		if ap.lev == -2 {
			inputData.Asm = append(inputData.Asm, "i32.const "+strconv.Itoa(ap.adr))
		}
	} else if ap.entryType == "var" || ap.entryType == "ref" || ap.entryType == "const" {
		loadItem(ap, inputData)
	} else {
		PrintError(inputData, "unsupported parameter type")
	}

	return ap
}

// Generates function calls.
func GenCall(entry *SymTableEntry, inputData *InputData) *SymTableEntry {
	inputData.Asm = append(inputData.Asm, "call $" + entry.name)
	return entry
}

// Generates call to the WASM stdproc read().
func GenRead(x *SymTableEntry, inputData *InputData) {
	inputData.Asm = append(inputData.Asm, "call $read")
	y := Var(Int)
	y.lev = -1
}

// Generates call to the WASM stdproc write().
func GenWrite(x *SymTableEntry, inputData *InputData) {
	loadItem(x, inputData)
	inputData.Asm = append(inputData.Asm, "call $write")
}

// Generates call to the WASM stdproc writeln().
func GenWriteln(inputData *InputData) {
	inputData.Asm = append(inputData.Asm, "call $writeln")
}

// Dummy function for generating sequences.
func GenSeq(x *SymTableEntry, y *SymTableEntry, inputData *InputData) {
	//pass
}

// Generates then.
func GenThen(x *SymTableEntry, inputData *InputData) *SymTableEntry {
	loadItem(x, inputData)
	inputData.Asm = append(inputData.Asm, "if")
	return x
}

// Generates if/then.
func GenIfThen(x *SymTableEntry, y *SymTableEntry, inputData *InputData) *SymTableEntry {
	inputData.Asm = append(inputData.Asm, "end")
	return x
}

// Generates else.
func GenElse(x *SymTableEntry, y *SymTableEntry, inputData *InputData) *SymTableEntry {
	inputData.Asm = append(inputData.Asm, "else")
	return y
}

// Generates if/else
func GenIfElse(x *SymTableEntry, y *SymTableEntry, z *SymTableEntry, inputData *InputData) *SymTableEntry {
	inputData.Asm = append(inputData.Asm, "end")
	return x
}

// Generates while.
func GenWhile(inputData *InputData) {
	inputData.Asm = append(inputData.Asm, "loop")
}

// Generates do.
func GenDo(x *SymTableEntry, inputData *InputData) *SymTableEntry {
	loadItem(x, inputData)
	inputData.Asm = append(inputData.Asm, "if")
	return x
}

// Generates while/do.
func GenWhileDo(x *SymTableEntry, y *SymTableEntry, inputData *InputData) {
	inputData.Asm = append(inputData.Asm, "br 1")
	inputData.Asm = append(inputData.Asm, "end")
	inputData.Asm = append(inputData.Asm, "end")
}
