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
	"unicode"
)

// User registered with server
type User struct {
	Name string
}

// Player of tic-tac-toe game
type Player struct {
	User
	Symbol string
}

// Mark action in tic-tac-toe
type Mark struct {
	User User
	X, Y int
}

// Lobby of users wanting to play
type Lobby struct {
	Users []User
}

// Game of tic-tac-toe
type Game struct {
	Players    []Player
	MaxPlayers int
	Turn       int
	Winner     *Player
	Board      [][]string
}

var (
	user  User
	game  Game
	lobby Lobby

	server *rpc.Client

	gamePort   = 27960
	serverAddr = "localhost"
	pollRate   = time.Second

	alpha = "abcdefghijklmnopqrstuvwxyz"
)

func main() {
	var err error
	server, err = rpc.DialHTTP("tcp", serverAddr+":"+fmt.Sprint(gamePort))

	if err != nil {
		fmt.Println("Failed to connect to server at ", serverAddr, ":", gamePort, " :", err)
		os.Exit(1)
	}

	Register()
	MaintainState()
	Input()
}

// Connect to server
func Register() {
	fmt.Println("--- TIC-TAC-TOE ---")

	// Get user name and register
	for {
		fmt.Print("Enter user name: ")

		reader := bufio.NewReader(os.Stdin)
		name, _ := reader.ReadString('\n')

		fmt.Println()

		// Register with server retrieve user struct
		err := server.Call("TTT.Register", name, &user)

		if err != nil {
			fmt.Println("Failed to register:", err)
			continue
		}

		// Render initial game state
		UpdateState()
		Render()

		break
	}
}

// Update state of client with data
func UpdateState() {
	var err error
	var newGame Game
	var newLobby Lobby

	err = server.Call("TTT.GetGame", user, &newGame)

	if err != nil {
		fmt.Println("Failed to get game:", err)
		os.Exit(1)
	}

	err = server.Call("TTT.GetLobby", user, &newLobby)

	if err != nil {
		fmt.Println("Failed to get lobby:", err)
		os.Exit(1)
	}

	if !reflect.DeepEqual(game, newGame) || !reflect.DeepEqual(lobby, newLobby) {
		game = newGame
		lobby = newLobby
		Render()
	}
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
				UpdateState()
			case <-poller:
				ticker.Stop()
				return
			}
		}
	}()
}

// Render interface
func Render() {

	// Check if player is in game
	if len(game.Players) != 0 {

		header := "--- TIC-TAC-TOE GAME ---\n"
		for _, p := range game.Players {
			header += "Player: " + p.Name + "\tSymbol: " + p.Symbol + "\n"
		}

		prompt := "--- COMMANDS ---\n"
		prompt += " exit\t\t\t- exit program\n"
		prompt += " quit\t\t\t- quit game\n"

		fmt.Println(header)

		if game.Turn < 0 {
			fmt.Println("Waiting for player to join ...")
			fmt.Println("\n" + prompt)
			return
		}

		// Could have made a dynamically rendered board but this was faster haha...
		board := "    1   2   3\n"
		board += "  ╔═══╦═══╦═══╗\n"
		board += "a ║ " + game.Board[0][0] + " ║ " + game.Board[0][1] + " ║ " + game.Board[0][2] + " ║\n"
		board += "  ╠═══╬═══╬═══╣\n"
		board += "b ║ " + game.Board[1][0] + " ║ " + game.Board[1][1] + " ║ " + game.Board[1][2] + " ║\n"
		board += "  ╠═══╬═══╬═══╣\n"
		board += "c ║ " + game.Board[2][0] + " ║ " + game.Board[2][1] + " ║ " + game.Board[2][2] + " ║\n"
		board += "  ╚═══╩═══╩═══╝\n"

		fmt.Println(board)

		if game.Winner != nil {
			if game.Winner.Name == user.Name {
				if len(game.Players) == 1 {
					fmt.Print("Other player left ... ")
				}
				fmt.Println("YOU WON!")
			} else {
				fmt.Println(game.Winner.Name, "won, you lost ...")
			}
			fmt.Println("\n" + prompt)
			return
		}

		if game.Turn == len(game.Board)*len(game.Board[0]) {
			fmt.Println("IT IS A DRAW, NOBODY WINS")
			fmt.Println("\n" + prompt)
			return
		}

		turn := game.Players[game.Turn%len(game.Players)]

		if turn.Name == user.Name {
			prompt += " <x y coordinates>\t- mark spot on board (example: a 1)\n\n"
			prompt += "YOUR TURN"
		} else {
			prompt += "\n" + turn.Name + "'s turn ...\n"
		}
		fmt.Println(prompt)

		return
	}

	// Not in game, render lobby
	fmt.Print("--- GAME LOBBY ---\n\n")
	if len(lobby.Users) == 0 {
		fmt.Println(" No games found ...")
	} else {
		for _, l := range lobby.Users {
			fmt.Println(" Game found: " + l.Name + "'s game")
		}
	}
	prompt := "\n--- COMMANDS ---\n"
	prompt += " exit\t\t\t- exit program\n"
	prompt += " create\t\t\t- make new lobby game\n"
	prompt += " join <player name>\t- join lobby game (example: join bob)\n"

	fmt.Println(prompt)
}

// Input handler function
func Input() {
	var err error
	for {
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')

		fmt.Println()

		input = strings.ReplaceAll(input, "\n", "")
		cmd := strings.SplitN(input, " ", 2)

		switch cmd[0] {
		case "exit":
			return
		case "quit":
			err = server.Call("TTT.QuitGame", user, &Game{})

			if err != nil {
				fmt.Println("Failed to quit game:", err)
			}
		case "create":
			err = server.Call("TTT.NewGame", user, &Game{})

			if err != nil {
				fmt.Println("Failed to create game:", err)
			}
		case "join":
			if len(cmd) != 2 {
				fmt.Println("Invalid cmd, use the format: join <player name>")
				continue
			}
			friend := User{strings.ReplaceAll(cmd[1], "\n", "")}
			err = server.Call("TTT.JoinGame", []User{user, friend}, &Game{})

			if err != nil {
				fmt.Println("Failed to join game:", err)
				continue
			}
		case "":
		default:
			// Handle coordinate input
			if len(cmd) == 2 && len(cmd[0]) == 1 && len(cmd[1]) == 1 {
				xRune, yRune := rune(cmd[0][0]), rune(cmd[1][0])
				if unicode.IsLetter(xRune) && unicode.IsNumber(yRune) {

					for x, s := range alpha {
						if s == xRune {
							y := int(yRune-'0') - 1
							err = server.Call("TTT.NewMark", Mark{user, x, y}, &Game{})

							if err != nil {
								fmt.Println("Failed to create mark:", err)
							}
							break
						}
					}
				} else {
					fmt.Println("Coordinates " + cmd[0] + "," + cmd[1] + " incorrectly formatted")
				}
			} else {
				fmt.Println("Command", cmd[0], "not recognized")
			}
		}
	}
}
