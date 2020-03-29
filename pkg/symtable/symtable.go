package symtable

// Struct for data related to symbol table entries.
type SymTableEntry struct {
	EntryType	string // should only ever be var, ref, const, type, proc, stdproc
	Name		string // Name of entry (e.g, x)
	Tp			PrimitiveType // primitive type (if applicable)
	Ctp			ComplexTypes // for more complicated types; for instance, some entries contain records
	Lev			int // scope Level
	Val			int // the Value of (if applicable)
	Par 		[]string // list of Parameters in a function (if applicable)
	Size		int	// Memory required to represent the type (for Bool and Int)
	Adr			int	// Address in memory
	Offset		int // Offset for a given element in a record or array
}

// Makes a new entry for the symbol table.
func NewSymTableEntry(EntryType string, Name string, Tp PrimitiveType, Ctp ComplexTypes, Lev int, Val int, Par []string) SymTableEntry{
	return SymTableEntry{
		EntryType: 	EntryType,
		Name: 		Name,
		Tp:			Tp,
		Ctp: 		Ctp,
		Lev:		Lev,
		Val: 		Val,
		Par:		Par,
		Size:		0,
		Adr:		0,
		Offset:		0}
}

// Enum for the three allowed P0 primitive types.
type PrimitiveType string
const (
	Int 	PrimitiveType 	= "int"
	Bool 	PrimitiveType	= "bool"
	None 	PrimitiveType	= "none"
)

// Represents an array or record.
type ComplexTypes struct {
	EntryType	string		// whether it's an array or record
	Fields 		[]SymTableEntry	// used for storing the Fields in a record
	Base		PrimitiveType// the Base type of an array
	Lower		int			// Lower bound of an array
	Length		int			// Length of an array
	Size		int			// Size of the type allowed in an array
}

// Define a complex type.
func NewComplexType(EntryType string, Fields []SymTableEntry, Base PrimitiveType, Lower int, Length int, Par []string) ComplexTypes{
	return ComplexTypes{
		EntryType:	EntryType,
		Fields: 	Fields,
		Base:		Base,
		Lower:		Lower,
		Length: 	Length,
		Size:		0}
}