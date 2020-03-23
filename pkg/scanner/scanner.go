package scanner

import (
	"fmt"
	i "group-11/pkg/inputdata"
	k "group-11/pkg/keywords"
	"math"
	"strconv"
	"unicode"
)

func Init(inputData *i.InputData) {
	GetChar(inputData)
	GetSym(inputData)
}

// Moves the index to the next char
func GetChar(inputData *i.InputData) {
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
func IdentKeyword(inputData *i.InputData) {
	start := inputData.Index - 1
	current := ""
	for unicode.IsLetter([]rune(inputData.Ch)[0]) {
		current = inputData.Input[start:inputData.Index]
		if val, ok := k.Keywords[current]; ok {
			inputData.Sym = val
			GetChar(inputData)
			break
		} else {
			inputData.Sym = k.IDENT
		}
		GetChar(inputData)
	}
	inputData.Val = current
	fmt.Print(inputData.Val)
}

// Converts a string of numbers into value
func Number(inputData *i.InputData) {
	inputData.Sym = k.NUMBER
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

func GetSym(inputData *i.InputData) {
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
		inputData.Sym = k.TIMES
	} else if inputData.Ch == "+" {
		GetChar(inputData)
		inputData.Sym = k.PLUS
	} else if inputData.Ch == "-" {
		GetChar(inputData)
		inputData.Sym = k.MINUS
	} else if inputData.Ch == "=" {
		GetChar(inputData)
		inputData.Sym = k.EQ
	} else if inputData.Ch == "<" {
		GetChar(inputData)
		if inputData.Ch == "=" {
			GetChar(inputData)
			inputData.Sym = k.LE
		} else if inputData.Ch == ">" {
			GetChar(inputData)
			inputData.Sym = k.NE
		} else {
			inputData.Sym = k.LT
		}
	} else if inputData.Ch == ">" {
		GetChar(inputData)
		if inputData.Ch == "=" {
			GetChar(inputData)
			inputData.Sym = k.GE
		} else {
			inputData.Sym = k.GT
		}
	} else if inputData.Ch == ";" {
		GetChar(inputData)
		inputData.Sym = k.SEMICOLON
	} else if inputData.Ch == "," {
		GetChar(inputData)
		inputData.Sym = k.COMMA
	} else if inputData.Ch == ":" {
		GetChar(inputData)
		inputData.Sym = k.COLON
	} else if inputData.Ch == "." {
		GetChar(inputData)
		inputData.Sym = k.PERIOD
	} else if inputData.Ch == "(" {
		GetChar(inputData)
		inputData.Sym = k.LPAREN
	} else if inputData.Ch == ")" {
		GetChar(inputData)
		inputData.Sym = k.RPAREN
	} else if inputData.Ch == "[" {
		GetChar(inputData)
		inputData.Sym = k.LBRAK
	} else if inputData.Ch == "]" {
		GetChar(inputData)
		inputData.Sym = k.RBRAK
	} else if inputData.Ch == "~" {
		inputData.Sym = k.EOF
	} else {
		PrintError(inputData, "illegal character")
		GetChar(inputData)
		inputData.Sym = 0
	}
	fmt.Print("\t")
	fmt.Println(inputData.Sym)
}

// Prints out an error and the line and pos it was found on
func PrintError(inputData *i.InputData, errorMsg string) {
	if inputData.LastLine > inputData.ErrorLine || inputData.LastPos > inputData.ErrorPos {
		fmt.Println("Error: line " + strconv.Itoa(inputData.LastLine) + ", pos " + strconv.Itoa(inputData.LastPos) + errorMsg)
	}
	inputData.ErrorLine = inputData.LastLine
	inputData.ErrorPos = inputData.LastPos
	inputData.Error = true
}
