package main

import "instant_messaging/src/server"

func main() {
	newServer := server.NewServer("127.0.0.1", "8080")
	newServer.Start()
}
