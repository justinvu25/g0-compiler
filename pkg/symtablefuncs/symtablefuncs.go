package symtablefuncs

import (
	"fmt"
	i "group-11/pkg/inputdata"
	s "group-11/pkg/scanner"
	"group-11/pkg/symtable"
)

// Prints the symbol table to the command line.
func PrintSymTable(inputData *i.InputData) {
	fmt.Println(inputData.SymTable)
}

// Add new symbol table entry.
func NewDecl(inputData *i.InputData, Name string, EntryType string){
	topLevel := inputData.SymTable[0]
	Lev := len(topLevel) - 1

	for _, entry := range topLevel {
		if entry.Name == Name {
			s.PrintError(inputData, "multiple definitions")
			return
		}
	}

	inputData.SymTable[0] = append(inputData.SymTable[0], symtable.SymTableEntry{EntryType: EntryType, Name: Name, Lev: Lev, Tp: symtable.Int})
}

// Finds a symbol table entry with a given Name.
func FindInSymTab(inputData *i.InputData, Name string) symtable.SymTableEntry {
	for _, Level := range inputData.SymTable {
		for _, entry := range Level {
			if entry.Name == Name {
				return entry
			}
		}
	}
	s.PrintError(inputData, "undefined identifier " + Name)
	return symtable.SymTableEntry{}
}

// Each list of lists of entries is a scope level; a new scope can be added
// by appending a new list of entries.
func OpenScope(inputData *i.InputData){
	inputData.SymTable = append([][]symtable.SymTableEntry{{}}, inputData.SymTable...)
}

// Simply returns the top-level scope.
func TopScope(inputData *i.InputData) []symtable.SymTableEntry {
	return inputData.SymTable[0]
}

// Close the top level scope by removing the 0th list of lists.
func CloseScope(inputData *i.InputData){
	inputData.SymTable = inputData.SymTable[1:]
}