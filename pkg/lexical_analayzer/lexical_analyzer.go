package lexical_analayzer

import (
	"fmt"
	i "group-11/pkg/inputdata"
	s "group-11/pkg/scanner"
	"io/ioutil"
	"log"
	"strings"
	"sync"
)

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
func ParseInput(inputData *i.InputData, wg *sync.WaitGroup) {
	defer wg.Done()

	s.Init(inputData)
	for inputData.Ch != "~" {
		s.GetSym(inputData)
	}

	fmt.Println("done parsing")
}

// Removes whitespace.
func EatWhiteSpace(inputData *i.InputData, wg *sync.WaitGroup) {
	defer wg.Done()
	//
	inputData.Input = strings.Replace(inputData.Input, " do", "!do", -1)
	inputData.Input = strings.Replace(inputData.Input, " end", "!end", -1)
	// Giving -1 to string.Replace removes an unlimited number of whitespaces.
	inputData.Input = strings.Replace(inputData.Input, " ", "", -1)

	fmt.Println("done removing whitespace")
}

func EatComments(inputData *i.InputData, wg *sync.WaitGroup) {
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
