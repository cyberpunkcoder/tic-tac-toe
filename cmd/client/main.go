package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

var (
	serverAddr = "localhost"
	gamePort   = ":27960"
)

func main() {
	connection, _ := net.Dial("tcp", serverAddr+gamePort)

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Send test message: ")

		text, _ := reader.ReadString('\n')
		fmt.Fprintf(connection, text+"\n")

		message, _ := bufio.NewReader(connection).ReadString('\n')
		fmt.Print("Server replies: " + message)
	}
}
