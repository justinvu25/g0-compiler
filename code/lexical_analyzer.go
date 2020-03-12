package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"
)

//
// Get data from the specified file and attach and EOF character (~) onto it.
// Next, call each of the parsing functions to process the input data and
// produce the appropriate tokens.
//
func main() {
	var wg sync.WaitGroup
	fileToParse := readData("p0code.txt")
	fileToParse += "~"

	// Add two items to the wait group, one for each goroutine.
	wg.Add(3)
	go eatWhitespace(&fileToParse, &wg)
	go countLines(&fileToParse, &wg)
	go readChars(&fileToParse, &wg)
	fmt.Println(fileToParse)

	// Wait for the waitgroup counter to reach zero before continuing.
	// The waitgroup counter is decremented each time a thread finishes
	// executing its procedure.
	wg.Wait()
}

//
// Read in data from a file.
//
func readData(name string) string {
	content, err := ioutil.ReadFile(name)
	if err != nil {
		log.Fatal(err)
	}

	return string(content)
}

//
// Parses characters into tokens.
//
func readChars(input *string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Currently, this is used to get an idea of how the P0 code
	// is parsed.
	test := ""
	i := 0
	for string((*input)[i]) != "~" {
		test += string((*input)[i])
		i += 1
	}
	fmt.Println(test)
}

//
// Removes whitespace.
//
func eatWhitespace(input *string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Giving -1 to string.Replace removes an unlimited number of whitespaces.
	*input = strings.Replace(*input, " ", "", -1)
	fmt.Println("done removing whitespace")
}

// counts the number of lines in input string
func countLines(input *string, wg *sync.WaitGroup) int {
	defer wg.Done()
	lineCount := strings.Count(*input, "\n")
	return lineCount
}

//
// Removes comments.
//
func removeComments(input *string) {
	//pass
}

//
// Keeps track of the newlines in the program.
//
func readNewlines(input *string) {
	//pass
}
