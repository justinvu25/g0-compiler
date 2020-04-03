package symtable

import (
	"fmt"
	i "group-11/pkg/inputdata"
	s "group-11/pkg/scanner"
)

// Struct for data related to symbol table entries.
type SymTableEntry struct {
	EntryType string          // should only ever be var, ref, const, type, proc, stdproc, array, record
	Name      string          // Name of entry (e.g, x)
	Tp        PrimitiveType   // primitive type (if applicable)
	Ctp       ComplexType     // for more complicated types; for instance, some entries contain records
	Lev       int             // scope Level
	Val       int             // the Value of (if applicable)
	Par       []SymTableEntry // list of Parameters in a function (if applicable)
	Size      int             // Memory required to represent the type (for Bool and Int)
	Adr       int             // Address in memory
	Offset    int             // Offset for a given element in a record or array
	ArrOrRec  string		  // If applicable, is it an array or record
}

// Enum for the three allowed P0 primitive types.
type PrimitiveType string

const (
	Int      PrimitiveType = "int"
	Bool     PrimitiveType = "bool"
	None     PrimitiveType = "none"
	Nil      PrimitiveType = ""
	EmptyInt int           = -9999999999 //Go doesn't have null ints, so for simplicity I just put a big negative number.
)

// Represents an array or record.
type ComplexType struct {
	Fields    []SymTableEntry // used for storing the Fields in a record
	Base      PrimitiveType   // the Base type of an array
	Lower     int             // Lower bound of an array
	Length    int             // Length of an array
	Size      int             // Size of the type allowed in an array
}

// Generates var symbol table entries.
func Var(Tp PrimitiveType) *SymTableEntry {
	e := &SymTableEntry{}
	e.EntryType = "var"
	e.Tp = Tp
	return e
}

// Generates ref symbol table entries.
func Ref(Tp PrimitiveType) *SymTableEntry {
	e := &SymTableEntry{}
	e.EntryType = "ref"
	e.Tp = Tp
	return e
}

// Generates const symbol table entries.
func Const(Tp PrimitiveType, Val int) *SymTableEntry {
	e := &SymTableEntry{}
	e.EntryType = "const"
	e.Tp = Tp
	e.Val = Val
	return e
}

// Generates type symbol table entries.
func Type(Tp PrimitiveType) *SymTableEntry {
	e := &SymTableEntry{}
	e.EntryType = "type"
	e.Tp = Tp
	e.Val = EmptyInt
	return e
}

// Generates proc symbol table entries.
func Proc(Par []SymTableEntry) *SymTableEntry {
	e := &SymTableEntry{}
	e.EntryType = "proc"
	e.Tp = None
	e.Par = Par
	return e
}

// Generates stdproc symbol table entries.
func StdProc(Par []SymTableEntry) *SymTableEntry {
	e := &SymTableEntry{}
	e.EntryType = "stdproc"
	e.Tp = None
	e.Par = Par
	return e
}

// Generates record symbol table entries.
func Record(Fields []SymTableEntry) *SymTableEntry {
	e := &SymTableEntry{}
	e.ArrOrRec = "record"
	e.Ctp = ComplexType{}
	e.Ctp.Fields = Fields
	return e
}

// Generates array symbol table entries.
func Array(Base PrimitiveType, Lower int, Length int) *SymTableEntry {
	e := &SymTableEntry{}
	e.ArrOrRec = "array"
	e.Ctp = ComplexType{}
	e.Ctp.Base = Base
	e.Ctp.Lower = Lower
	e.Ctp.Length = Length
	return e
}

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

	inputData.SymTable[0] = append(inputData.SymTable[0], SymTableEntry{EntryType: EntryType, Name: Name, Lev: Lev, Tp: Int})
}

// Finds a symbol table entry with a given Name.
func FindInSymTab(inputData *i.InputData, Name string) SymTableEntry {
	for _, Level := range inputData.SymTable {
		for _, entry := range Level {
			if entry.Name == Name {
				return entry
			}
		}
	}
	s.PrintError(inputData, "undefined identifier " + Name)
	return SymTableEntry{}
}

// Each list of lists of entries is a scope level; a new scope can be added
// by appending a new list of entries.
func OpenScope(inputData *i.InputData){
	inputData.SymTable = append([][]SymTableEntry{{}}, inputData.SymTable...)
}

// Simply returns the top-level scope.
func TopScope(inputData *i.InputData) []SymTableEntry {
	return inputData.SymTable[0]
}

// Close the top level scope by removing the 0th list of lists.
func CloseScope(inputData *i.InputData){
	inputData.SymTable = inputData.SymTable[1:]
}