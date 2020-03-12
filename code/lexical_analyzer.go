package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"
)

//Get data from the specified file and attach and EOF character (~) onto it.
//Next, call each of the parsing functions to process the input data and
//produce the appropriate tokens.

type inputData struct {
	input      string
	lineNumber int
	error      bool
}

// constructor for inputData struct
func newInputData(fileName string) *inputData {
	input := fileHelper(fileName)
	s := inputData{input: input, lineNumber: 0, error: false}
	return &s
}

func main() {
	var wg sync.WaitGroup
	inputData := newInputData("p0code.txt")

	// Add two items to the wait group, one for each goroutine.
	wg.Add(2)
	go eatWhitespace(inputData, &wg)
	//go countLines(inputData, &wg)
	go readChars(inputData, &wg)

	// Wait for the waitgroup counter to reach zero before continuing.
	// The waitgroup counter is decremented each time a thread finishes
	// executing its procedure.
	wg.Wait()
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
func readChars(inputData *inputData, wg *sync.WaitGroup) {
	defer wg.Done()

	// Currently, this is used to get an idea of how the P0 code is parsed.
	test := ""
	i := 0
	for string(inputData.input[i]) != "~" {
		test += string(inputData.input[i])
		i++
	}
	fmt.Println(test)
}

// Removes whitespace.
func eatWhitespace(inputData *inputData, wg *sync.WaitGroup) {
	defer wg.Done()

	// Giving -1 to string.Replace removes an unlimited number of whitespaces.
	inputData.input = strings.Replace(inputData.input, " ", "", -1)
	fmt.Println("done removing whitespace")
}

// counts the number of lines in input string
//func countLines(input *string, wg *sync.WaitGroup) int {
//	defer wg.Done()
//	lineCount := strings.Count(*input, "\n")
//	return lineCount
//}

// Removes comments.
func removeComments(input *string) {
	//pass
}

// Keeps track of the newlines in the program.
func readNewlines(input *string) {
	//pass
}
