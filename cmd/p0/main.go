package main

import (
	"fmt"
	i "group-11/pkg/inputdata"
	l "group-11/pkg/lexical_analayzer"
	"sync"

)

func main() {
	var wg sync.WaitGroup
	inputData := i.NewInputData("../../config/p0test.txt") //newInputData("p0code.txt")

	// Add items to the wait group, one for each goroutine.
	wg.Add(3)

	go l.EatWhiteSpace(inputData, &wg)
	go l.EatComments(inputData, &wg)
	go l.ParseInput(inputData, &wg)
	// Wait for the waitgroup counter to reach zero before continuing.
	// The waitgroup counter is decremented each time a thread finishes
	// executing its procedure.
	wg.Wait()
	fmt.Println("Done all tasks")
}
