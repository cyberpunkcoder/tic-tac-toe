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
	"net/rpc"
	"os"
	"reflect"
	"strings"
	"time"
)

// Player of tic-tac-toe
type Player struct {
	Name   string
	Symbol string
}

// Game of tic-tac-toe
type Game struct {
	Name       string
	Players    []Player
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
		fmt.Println("Failed to connect to server at ", serverAddr, ":", gamePort, " :", err)
		os.Exit(1)
	}

	Register()
	Render()
	MaintainState()
	Input()
}

// Connect to server
func Register() {
	fmt.Println("--- TIC-TAC-TOE ---")

	// Get player name and register
	for {
		fmt.Print("Enter player name: ")

		reader := bufio.NewReader(os.Stdin)
		name, _ := reader.ReadString('\n')

		fmt.Println()

		// Register with server retrieve player struct
		err := server.Call("API.Register", name, &player)

		if err != nil {
			fmt.Println("Failed to register:", err)
			continue
		}

		// Get initial state of client from server
		err = server.Call("API.GetState", player, &state)

		if err != nil {
			fmt.Println("Failed to get state:", err)
			os.Exit(1)
		}

		break
	}
}

// Render client state
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
				fmt.Printf(" Name: %s\tPlayers: %d/%d\n", g.Name, players, len(g.Players))
			}
		}
		prompt := "\n--- COMMANDS ---\n"
		prompt += " exit\t\t\t- exit program\n"
		prompt += " create\t\t\t- make new lobby game\n"
		prompt += " join <game name>\t- join lobby game\n"

		fmt.Println(prompt)
		return
	}
	// Render game header
	fmt.Println("--- TIC-TAC-TOE GAME ---")
	fmt.Println("Name:", state.Game.Name)
	for _, p := range state.Game.Players {
		fmt.Printf("Player: %s\tSymbol: %s\n", p.Name, p.Symbol)
	}

	// Render board (I could have made a dynamically rendered board but this was faster haha)
	board := "\n    1   2   3\n"
	board += "  ╔═══╦═══╦═══╗\n"
	board += "A ║ " + state.Game.Board[0][0] + " ║ " + state.Game.Board[0][1] + " ║ " + state.Game.Board[0][2] + " ║\n"
	board += "  ╠═══╬═══╬═══╣\n"
	board += "B ║ " + state.Game.Board[1][0] + " ║ " + state.Game.Board[1][1] + " ║ " + state.Game.Board[1][2] + " ║\n"
	board += "  ╠═══╬═══╬═══╣\n"
	board += "C ║ " + state.Game.Board[2][0] + " ║ " + state.Game.Board[2][1] + " ║ " + state.Game.Board[2][2] + " ║\n"
	board += "  ╚═══╩═══╩═══╝\n"

	fmt.Println(board)

	prompt := "--- COMMANDS ---\n"
	prompt += " exit\t\t\t- exit program\n"
	prompt += " quit\t\t\t- quit game\n"

	// Check who's turn it is
	turn := state.Game.Players[state.Game.Turn%len(state.Game.Players)]

	if player.Name == turn.Name {
		prompt += " <coordinates>\t\t- mark a spot on the board (example: A1)\n\n"
		prompt += "YOUR TURN"
	} else {
		prompt += "\n" + turn.Name + "'s turn ...\n"
	}
	fmt.Println(prompt)
}

// MaintainState of client with data from server
func MaintainState() {
	ticker := time.NewTicker(pollRate)
	poller := make(chan struct{})

	// Asynchronous state update from server
	go func() {
		for {
			select {
			case <-ticker.C:
				oldState := state
				err := server.Call("API.GetState", player, &state)

				if err != nil {
					fmt.Println("Failed to get state:", err)
					os.Exit(1)
				}
				// Check for any changes in state
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

// Input handler function
func Input() {
	for {
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')

		fmt.Println()

		input = strings.ReplaceAll(input, "\n", "")
		command := strings.SplitN(input, " ", 2)

		switch command[0] {
		case "exit":
			return
		case "create":
			err := server.Call("API.NewGame", player, &state)

			if err != nil {
				fmt.Println("Failed to create game:", err)
			}
		case "join":
			if len(command) != 2 {
				fmt.Println("Invalid command, use the format: join <game name>")
				continue
			}
			err := server.Call("API.JoinGame", command[1], &state)

			if err != nil {
				fmt.Println("Failed to join game:", err)
				continue
			}
		case "":
		default:
			fmt.Println("Command", command[0], "not recognized")
		}
	}
}
