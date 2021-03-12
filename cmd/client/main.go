/*
	Tic-Tac-Toe, a multiplayer Tic-Tac-Toe client
    Copyright (C) 2021	cyberpunkcoder
	github.com/cyberpunkcoder
	cyberpunkcoder@gmail.com
    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.
    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.
    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"bufio"
	"fmt"
	"log"
	"net/rpc"
	"os"
)

var (
	serverAddr = "localhost"
	gamePort   = 27960
	player     Player
)

// Player of tic-tac-toe game
type Player struct {
	Name string
}

func main() {
	portString := fmt.Sprint(gamePort)
	server, err := rpc.DialHTTP("tcp", serverAddr+":"+portString)

	if err != nil {
		log.Fatal("Failed to connect to server at "+serverAddr+":"+portString+": ", err)
	}

	fmt.Println("Tic-Tac-Toe")

	// Get player name and register
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter player name: ")
		name, _ := reader.ReadString('\n')

		err = server.Call("TTT.NewPlayer", name, &player)

		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}
