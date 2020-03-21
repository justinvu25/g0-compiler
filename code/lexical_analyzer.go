package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

//Get data from the specified file and attach and EOF character (~) onto it.
//Next, call each of the parsing functions to process the input data and
//produce the appropriate tokens.

type inputData struct {
	input      string // P0 source code
	sym        int    // Symbol that was identified.
	ch         string // Current character.
	index      int    // Helps identify symbols
	val        string // Value
	lineNumber int    // Current line being parsed
	lastLine   int    // Previous line that was parsed
	errorLine  int    // Used to help surpress multiple errors
	pos        int    // Current position of parser in a line
	lastPos    int    // Previous position
	errorPos   int    // Used to help surpress multiple errors
	error      bool   // Set to true when an error is found.
}

// constructor for inputData struct
func newInputData(fileName string) *inputData {
	input := fileHelper(fileName)
	s := inputData{
		input:      input,
		sym:        0,
		ch:         "",
		index:      0,
		val:        "",
		lineNumber: 1,
		lastLine:   1,
		errorLine:  1,
		pos:        0,
		lastPos:    0,
		errorPos:   0,
		error:      false}
	return &s
}

func main() {
	var wg sync.WaitGroup
	inputData := newInputData("p0test.txt") //newInputData("p0code.txt")

	// Add items to the wait group, one for each goroutine.
	wg.Add(3)

	go eatWhitespace(inputData, &wg)
	go eatComments(inputData, &wg)
	//go countLines(inputData, &wg)
	go parseInput(inputData, &wg)
	// Wait for the waitgroup counter to reach zero before continuing.
	// The waitgroup counter is decremented each time a thread finishes
	// executing its procedure.
	wg.Wait()
	fmt.Println("Done all tasks")
}

// Read in data from a file.
func fileHelper(fileName string) string {
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	contentsEOFChar := string(contents) + "~"

	return string(contentsEOFChar)
}

// Parses characters into tokens.
func parseInput(inputData *inputData, wg *sync.WaitGroup) {
	defer wg.Done()

	Init(inputData)
	for inputData.ch != "~" {
		GetSym(inputData)
	}

	fmt.Println("done parsing")
}

// Removes whitespace.
func eatWhitespace(inputData *inputData, wg *sync.WaitGroup) {
	defer wg.Done()
	//
	inputData.input = strings.Replace(inputData.input, " do", "!do", -1)
	inputData.input = strings.Replace(inputData.input, " end", "!end", -1)
	// Giving -1 to string.Replace removes an unlimited number of whitespaces.
	inputData.input = strings.Replace(inputData.input, " ", "", -1)

	fmt.Println("done removing whitespace")
}

// counts the number of lines in input string
// func countLines(input *string, wg *sync.WaitGroup) int {
// 	defer wg.Done()
// 	lineCount := strings.Count(*input, "\n")
// 	return lineCount
// }

func eatComments(inputData *inputData, wg *sync.WaitGroup) {
	defer wg.Done()
	i := 0
	opening := false
	closing := false
	m := 0
	n := 0

	for string((inputData.input)[i]) != "~" {
		if string((inputData.input)[i]) == "{" {
			m = i
			opening = true
		} else if string((inputData.input)[i]) == "}" && opening == true {
			n = i
			closing = true
		}
		i++
		if opening == true && closing == true {
			opening = false
			closing = false
			inputData.input = string((inputData.input)[:m] + "" + (inputData.input)[n+1:])
			i = m
		}
	}
	// fmt.Println(inputData.input)
	fmt.Println("done removing comments")
}

/*

	EVENTUALLY MOVE THIS TO A SEPARATE MODULE

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

//
//
//
func Init(inputData *inputData) {
	GetChar(inputData)
	GetSym(inputData)
}

//
// Moves the index to the next char
//
func GetChar(inputData *inputData) {
	// This loop will wait at a given index for the whitespace and comment eaters to catch up.
	for string(inputData.input[inputData.index]) == " " {
		//pass
	}
	currentChar := string(inputData.input[inputData.index])
	if currentChar == "~" {
		inputData.ch = currentChar
	} else {
		inputData.ch = currentChar
		inputData.index += 1
		if currentChar == "\n" {
			inputData.pos = 0
			inputData.lineNumber += 1
		} else {
			inputData.lastLine = inputData.lineNumber
			inputData.pos += 1
		}
	}
}

//
// Identifies keywords in a given sequence of characters.
//
func IdentKeyword(inputData *inputData) {
	start := inputData.index - 1
	current := ""
	for unicode.IsLetter([]rune(inputData.ch)[0]) {
		current = inputData.input[start:inputData.index]
		if val, ok := Keywords[current]; ok {
			inputData.sym = val
			GetChar(inputData)
			break
		} else {
			inputData.sym = IDENT
		}
		GetChar(inputData)
	}
	inputData.val = current
	fmt.Print(inputData.val)
}

//
// Converts a string of numbers into value
//
func Number(inputData *inputData) {
	inputData.sym = NUMBER
	inputData.val = ""
	for unicode.IsNumber([]rune(inputData.ch)[0]) {
		inputData.val += inputData.ch
		GetChar(inputData)
	}

	res, err3 := strconv.Atoi(inputData.val)
	if float64(res) >= math.Exp2(31) {
		PrintError(inputData, "number too large")
	}
	if err3 != nil {
		PrintError(inputData, "cannot convert to number")
	}
	fmt.Print("number is: ")
	fmt.Print(inputData.val)
}

//
//
//
func GetSym(inputData *inputData) {
	// []byte{13} is an invisible character that gets picked up from the input file.
	// ! is the delimiter to help identify the "do" from "while-do" statements.
	for inputData.ch == "\n" || inputData.ch == " " || inputData.ch == "!" {
		GetChar(inputData)
	}
	if unicode.IsLetter([]rune(inputData.ch)[0]) {
		IdentKeyword(inputData)
	} else if unicode.IsNumber([]rune(inputData.ch)[0]) {
		Number(inputData)
	} else if inputData.ch == "*" {
		GetChar(inputData)
		inputData.sym = TIMES
	} else if inputData.ch == "+" {
		GetChar(inputData)
		inputData.sym = PLUS
	} else if inputData.ch == "-" {
		GetChar(inputData)
		inputData.sym = MINUS
	} else if inputData.ch == "=" {
		GetChar(inputData)
		inputData.sym = EQ
	} else if inputData.ch == "<" {
		GetChar(inputData)
		if inputData.ch == "=" {
			GetChar(inputData)
			inputData.sym = LE
		} else if inputData.ch == ">" {
			GetChar(inputData)
			inputData.sym = NE
		} else {
			inputData.sym = LT
		}
	} else if inputData.ch == ">" {
		GetChar(inputData)
		if inputData.ch == "=" {
			GetChar(inputData)
			inputData.sym = GE
		} else {
			inputData.sym = GT
		}
	} else if inputData.ch == ";" {
		GetChar(inputData)
		inputData.sym = SEMICOLON
	} else if inputData.ch == "," {
		GetChar(inputData)
		inputData.sym = COMMA
	} else if inputData.ch == ":" {
		GetChar(inputData)
		inputData.sym = COLON
	} else if inputData.ch == "." {
		GetChar(inputData)
		inputData.sym = PERIOD
	} else if inputData.ch == "(" {
		GetChar(inputData)
		inputData.sym = LPAREN
	} else if inputData.ch == ")" {
		GetChar(inputData)
		inputData.sym = RPAREN
	} else if inputData.ch == "[" {
		GetChar(inputData)
		inputData.sym = LBRAK
	} else if inputData.ch == "]" {
		GetChar(inputData)
		inputData.sym = RBRAK
	} else if inputData.ch == "~" {
		inputData.sym = EOF
	} else {
		PrintError(inputData, "illegal character")
		GetChar(inputData)
		inputData.sym = 0
	}
	fmt.Print("\t")
	fmt.Println(inputData.sym)
}

//
// Prints out an error and the line and pos it was found on
//
func PrintError(inputData *inputData, errorMsg string) {
	if inputData.lastLine > inputData.errorLine || inputData.lastPos > inputData.errorPos {
		fmt.Println("Error: line " + strconv.Itoa(inputData.lastLine) + ", pos " + strconv.Itoa(inputData.lastPos) + errorMsg)
	}
	inputData.errorLine = inputData.lastLine
	inputData.errorPos = inputData.lastPos
	inputData.error = true
}
