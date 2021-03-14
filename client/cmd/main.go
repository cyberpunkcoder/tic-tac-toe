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
	"reflect"
	"time"
)

// Player of tic-tac-toe
type Player struct {
	Name string
}

// Game of tic-tac-toe
type Game struct {
	Name       string
	Players    map[string]Player
	Board      [][]string
	Turn       int
	MaxPlayers int
}

// State of tic-tac-toe client
type State struct {
	Game  *Game
	Games []Game
}

var (
	serverAddr = "localhost"
	gamePort   = 27960
	pollRate   = time.Second
	server     *rpc.Client
	state      State
	player     Player
)

func main() {
	var err error
	server, err = rpc.DialHTTP("tcp", serverAddr+":"+fmt.Sprint(gamePort))

	if err != nil {
		log.Fatal("Failed to connect to server at ", serverAddr, ":", gamePort, " : ", err)
	}

	Login()
	Render()
	UpdateState()
	Menu()
}

func Login() {
	fmt.Println("--- TIC-TAC-TOE ---")

	// Get player name and register
	for {
		fmt.Print("Enter player name: ")

		reader := bufio.NewReader(os.Stdin)
		name, _ := reader.ReadString('\n')

		fmt.Println()

		// Connect to server and retrieve player struct
		err := server.Call("API.Connect", name, &player)

		if err != nil {
			fmt.Println(err)
			continue
		}

		// Get initial state of client from server
		err = server.Call("API.GetState", player, &state)

		if err != nil {
			log.Fatal(err)
		}

		break
	}
}

func Render() {
	// If not in game, render lobby
	if state.Game == nil {
		fmt.Println("--- GAME LOBBY ---")
		if len(state.Games) == 0 {
			fmt.Println("\tNo games found ...")
		} else {
			for _, g := range state.Games {
				players := 0
				for _, p := range g.Players {
					if p.Name == "" {
						continue
					}
					players++
				}
				fmt.Printf("\tName: %s\tPlayers: %d/%d\n", g.Name, players, len(g.Players))
			}
		}
		prompt := "\n--- COMMANDS ---\n"
		prompt += "\tquit\t\t\t- exit game\n"
		prompt += "\tvsai\t\t\t- play against a computer\n"
		prompt += "\tcreate\t\t\t- make new lobby game\n"
		prompt += "\tjoin <game name>\t- join lobby game\n"

		fmt.Println(prompt)
		return
	}
	// Render game

}

// UpdateState of client with data from server
func UpdateState() {
	ticker := time.NewTicker(pollRate)
	poller := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				oldState := state
				err := server.Call("API.GetState", player, &state)

				if err != nil {
					fmt.Println(err)
					close(poller)
				}
				// State has changed, render changes
				if !reflect.DeepEqual(oldState, state) {
					Render()
				}
			case <-poller:
				ticker.Stop()
				return
			}
		}
	}()
}

func Menu() {
	for {
		reader := bufio.NewReader(os.Stdin)
		cmd, _ := reader.ReadString('\n')

		if cmd == "quit" {
			break
		}
	}
}
