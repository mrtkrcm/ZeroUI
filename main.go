package main

// TODO: Add graceful shutdown handling for long-running operations (Week 3)
// TODO: Implement signal handling (SIGINT, SIGTERM) for clean exit
// TODO: Add context cancellation for all operations

import (
	"github.com/mrtkrcm/ZeroUI/cmd"
)

func main() {
	cmd.Execute()
}