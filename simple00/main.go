package main

import (
	"bufio"
	"os"
)

// main is the starting point of the program
func main() {
	Startup()

	// Hit enter to terminate the program gracefully
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	Shutdown()
}
