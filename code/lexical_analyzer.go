package main

import (
	"fmt"
	"io/ioutil"
	"log"
)

//
// Get data from the specified file and attach and EOF character (~) onto it.
// Next, call each of the parsing functions to process the input data and
// produce the appropriate tokens.
//
func main() {
	fileToParse := readData("p0code.txt")
	fileToParse += "~"
	fmt.Println(fileToParse)
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
func readChars(input *string) {
	//pass
	//maybe increment a pointer i and access it like filecontents[i], building
	//a lexeme the whole time
}

//
// Removes whitespace.
//
func eatWhitespace(input *string) {
	//pass
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
