package main

import "github.com/sab1tm/go2net/cmd/storage"

func main() {
	server := server.NewServer(":3000")
	server.Start()
}
