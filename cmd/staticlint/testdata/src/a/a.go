package main

import (
	"os"
	stdos "os"
)

func main() {
	os.Exit(1)    // want `direct call to os.Exit in main function of package main is forbidden`
	stdos.Exit(0) // want `direct call to os.Exit in main function of package main is forbidden`
	helper()
}

func helper() {
	os.Exit(2)
}
