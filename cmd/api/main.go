package main

import "github.com/nabhag8848/orchex/internal/server"

func main() {
	e := server.NewServer()
	if err := e.Start(":8080"); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}