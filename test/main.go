package main

import (
	"fmt"
	cg "group-11/pkg/codegen"
	i "group-11/pkg/inputdata"
	stf "group-11/pkg/symtablefuncs"
)

func main() {
	testCodeGen()
}

func testCodeGen() {
	id := i.NewInputData("p0test.txt")
	fmt.Println(id.Input)
	stf.NewDecl(id, "test", "var")
	stf.NewDecl(id, "test2", "var")
	stf.NewDecl(id, "test3", "var")
	stf.NewDecl(id, "test4", "var")


	cg.WriteWasmFile("test.wasm", id)
	fmt.Println("all tests done")
}