package main

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
	"github.ibm.com/soub4i/kdebug-mcp/internal/srv"
)

func main() {

	s := srv.NewSHandler()
	s.RegisterHandlers()

	if err := server.ServeStdio(s.Server); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
