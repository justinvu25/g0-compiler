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
		Error:      false}
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
	go ParseInput(inputData, &wg)
	// Wait for the waitgroup counter to reach zero before continuing.
	// The waitgroup counter is decremented each time a thread finishes
	// executing its procedure.
	wg.Wait()
	fmt.Println("Done all tasks")
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


// Parses characters into tokens.
func ParseInput(inputData *InputData, wg *sync.WaitGroup) {
	defer wg.Done()

	Init(inputData)
	for inputData.Ch != "~" {
		GetSym(inputData)
	}

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
	if inputData.LastLine > inputData.ErrorLine || inputData.LastPos > inputData.ErrorPos {
		fmt.Println("Error: line " + strconv.Itoa(inputData.LastLine) + ", pos " + strconv.Itoa(inputData.LastPos) + errorMsg)
	}
	inputData.ErrorLine = inputData.LastLine
	inputData.ErrorPos = inputData.LastPos
	inputData.Error = true
}
