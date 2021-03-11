package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

var (
	gamePort = ":27960"
)

func main() {
	startServer()
}

func startServer() {
	log.Println("Server started on port " + gamePort)

	// My favorite port 27960 because old quake and wolfenstein
	listenPort, _ := net.Listen("tcp", gamePort)
	connection, _ := listenPort.Accept()

	// Control + C to exit
	for {
		message, _ := bufio.NewReader(connection).ReadString('\n')
		log.Println("Received: ", string(message))
		fmt.Fprintf(connection, message+"\n")
	}
}
