# G0 Compiler
## Compiling:
Ensure that your GOPATH is where Go is installed on your machine
```bash
$ cd main
$ go build main.go
$ ./main
```

## Packages: 
### Code Generator
- Generates WASM code
### Keywords
- Identifies all keywords in language
### Lexical Analyser 
- Reads in the data to be analysed from a file.
- Parses the data into tokens
- Removes whitespace
- Removes comments
### Scanner
- Identifies keywords 
- Iterates to the next character
### Symtable
A struct outlining the definition of a symbol table entry. 

| Field         | Description           | Type  |
| ------------- |:-------------:| -----:|
| EntryType     | options: `var`, `ref`, `proc`, `const`, `type`, `proc`, `stdproc`, `array`, `record` | string |
| Tp     | Options: `Int`, `Bool`, `None` |   PrimitiveType |
| Ctp | Should be used if entry is an array or record |    ComplexType |
| Lev | Integer determining scope level |    int |
| Val | The value of the entry |       int |
| Par | The list of parameters in a function |   []SymTableEntry |
| Size | Size of memory allocation |    int |
| Adr | The address in memory      |    int |
| Offset | The offset for a given element in an array/record      |    int |
| ArrOrRec | indication if entry is an array or record      |    string |

### Symtablefuncs
- Given a name, can find a symbol table entry
- Gets top level scope 
- Closes top level scope
- Adds a new scope 
