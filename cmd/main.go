package main

import "go2net/server"

func main() {
	server := server.NewServer(":3003")
	server.Start()
}
