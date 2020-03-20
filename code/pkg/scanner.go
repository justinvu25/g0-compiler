package scanner

import (
	"fmt"
)


//
// Moves the index to the next char
//
func GetChar(inputData *inputData) {
	// This loop will wait at a given index for the whitespace and comment eaters to catch up.
	for string(inputData.input[inputData.index]) == " " {
		//pass
	}

	currentChar := string(inputData.input[inputData.index])
	if currentChar == '~' {
		inputData.ch = currentChar
	} else {
		inputData.ch = currentChar
		inputData.index += 1 
		if currentChar == '\n' {
			inputData.pos = 0
			inputData.lineNumber += 1
		} else {
			inputData.lastLine = inputData.lineNumber
			inputData.pos += 1
		}
	}
}

//
//
//
func GetSym(inputData *inputData) {

}

//
//
//
func PrintError(inputData *inputData, errorMsg string) {
	if inputData.lastLine > inputData.errorLine || inputData.lastPos > inputData.errorPos {
		fmt.Println("Error: " + errorMsg)
	}
	inputData.errorLine = inputData.lastLine
	inputData.errorPos = inputData.lastPos
	inputData.error = true
}

//
//
//
func IdentKeyword(inputData *inputData) {
	start := inputData.index - 1
}