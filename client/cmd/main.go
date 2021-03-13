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
	"strings"
	"time"
)

// State of player tic-tac-toe world
type State struct {
	Player
	Game     Game
	GameList []string
}

// Player of tic-tac-toe
type Player struct {
	Name   string
	Symbol rune
}

// Game of tic-tac-toe
type Game struct {
	Name    string
	Players []Player
	Board   []rune
	Turn    int
}

var (
	serverAddr = "localhost"
	gamePort   = 27960
	pollRate   = 1000

	server *rpc.Client
	state  State
)

func main() {
	var err error
	server, err = rpc.DialHTTP("tcp", serverAddr+":"+fmt.Sprint(gamePort))

	if err != nil {
		log.Fatal("Failed to connect to server at ", serverAddr, ":", gamePort, " : ", err)
	}

	Login()
	Poll()
	Login()
}

func Login() {
	fmt.Println("--- Tic-Tac-Toe Login ---")

	// Get player name and register
	for {
		fmt.Print("Enter player name: ")

		reader := bufio.NewReader(os.Stdin)
		name, _ := reader.ReadString('\n')
		name = strings.ReplaceAll(name, "\n", "")

		err := server.Call("API.Connect", name, &state)

		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("Welcome " + name + "!")
		break
	}
}

// Poll the server to maintain a connection
func Poll() {
	ticker := time.NewTicker(time.Duration(pollRate * int(time.Millisecond)))
	poller := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				err := server.Call("API.Poll", state.Name, &state)
				if err != nil {
					fmt.Println(err)
					close(poller)
				}
			case <-poller:
				ticker.Stop()
				return
			}
		}
	}()
}

/*

func Menu() {
	// Get player name and register
	for {
		fmt.Print("Commands: quit")

		reader := bufio.NewReader(os.Stdin)
		cmd, _ := reader.ReadString('\n')

		if cmd == "quit" {
			break
		}
	}
}

func games() {
	var games []Game
	err := server.Call("API.GetGames", &player, &games)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("--- List of Joinable Games ---")
	for i, g := range games {
		fmt.Printf("%d %s", i, g.Name)
	}
}*/
