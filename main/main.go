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
		SymTable:	[][]SymTableEntry{{}}}
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
type SymTableEntry struct {
	entryType	string // should only ever be var, ref, const, type, proc, stdproc
	name		string // name of entry (e.g, x)
	tp		PrimitiveType // primitive type (if applicable)
	ctp		ComplexTypes // for more complicated types; for instance, some entries contain records
	lev		int // scope level
	val		int // the value of (if applicable)
	par 	[]string // list of parameters in a function (if applicable)
}

// Makes a new entry in the symbol table
func NewSymTableEntry(entryType string, name string, tp PrimitiveType, ctp ComplexTypes, lev int, val int, par []string) SymTableEntry{
	return SymTableEntry{
		entryType: 	entryType,
		name: 		name,
		tp:			tp,
		ctp: 		ctp,
		lev:		lev,
		val: 		val,
		par:		par}
}

type PrimitiveType string

const (
	Int 	PrimitiveType 	= "int"
	Bool 	PrimitiveType	= "bool"
	None 	PrimitiveType	= "none"
)

type ComplexTypes struct {
	fields 	[]string	// used for storing the fields in a record
	base	string		// the base type of an array
	lower	int			// lower bound of an array
	length	int			// length of an array
}

// Define a complex type
func NewComplexType(fields []string, base string, lower int, length int, par []string) ComplexTypes{
	return ComplexTypes{
		fields: 	fields,
		base:		base,
		lower:		lower,
		length: 	length}
}

func PrintSymTable(inputData *InputData) {
	fmt.Println(inputData.SymTable)
}

//
// Add new symbol table entry.
//
func NewDecl(inputData *InputData, name string){
	// POTENTIALLY REFACTOR THIS FUNCTION
	topLevel := inputData.SymTable[0]
	lev := len(topLevel) - 1

	for _, entry := range topLevel {
		if entry.name == name {
			PrintError(inputData, "multiple definitions")
			return
		}
	}

	inputData.SymTable[0] = append(inputData.SymTable[0], SymTableEntry{name: name, lev: lev, tp: Int})
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