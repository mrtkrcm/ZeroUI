package main

import (
	"os"

	"github.com/mrtkrcm/ZeroUI/cmd"
)

func main() {
	exitCode := cmd.Run()
	os.Exit(exitCode)
}
