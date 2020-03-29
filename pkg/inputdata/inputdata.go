package InputData

import (
	st "group-11/pkg/symtable"
	"io/ioutil"
	"log"
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
	SymTable   [][]st.SymTableEntry // Symbol table of items that will be turned into WASM.
	Curlev     int    // Current scope level of the code generator.
	Memsize    int	  // Size of the required memory allocation.
	Asm		   []string // The string that will ultimately become the WASM file.
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
		SymTable:	[][]st.SymTableEntry{{}},
		Curlev:		0,
		Memsize:	0,
		Asm:		[]string{}}
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
