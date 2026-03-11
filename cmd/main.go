package main

import "go2net/server"

func main() {
	server := server.NewServer(":3000")
	server.Start()
}
