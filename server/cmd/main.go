/*
	Tic-Tac-Toe, a multiplayer Tic-Tac-Toe server
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
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	// I tried so so hard to get gRPC to work, I couldn't in time :(
	// It would have been so much cleaner, more robust and secure
	"net/rpc"
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

// Client connected to server
type Client struct {
	*Player
	LoggedIn bool
	LastPoll time.Time
}

var (
	pollRate    = 1000  // Poll rate between client and server in miliseconds
	pollTimeout = 5000  // Timeout for poll between client and server in miliseconds
	gamePort    = 27960 // My favorite port 27960 because old quake and wolfenstein :)
	games       []Game
	clients     []*Client
)

// API server for tic-tac-toe RPC interface
type API int

func main() {
	api := new(API)
	err := rpc.Register(api)

	if err != nil {
		log.Fatal("Failed to register rpc: ", err)
	}

	rpc.HandleHTTP()

	portString := fmt.Sprint(gamePort)
	lis, err := net.Listen("tcp", ":"+portString)

	if err != nil {
		log.Fatal("Failed to listen on port "+portString+": ", err)
	}

	checkTimeout()

	log.Println("Server started on port " + portString)
	err = http.Serve(lis, nil)

	if err != nil {
		log.Fatal("Failed to serve: ", err)
	}

}

// Connect to API
func (a *API) Connect(name string, player *Player) error {

	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if len(name) > 32 {
		return fmt.Errorf("name must be under 32 characters")
	}

	for _, c := range clients {
		if c.Name == name {
			if c.LoggedIn {
				return fmt.Errorf("player %s is already logged in", name)
			}
			// Client has returned from previous login
			c.LoggedIn = true
			c.LastPoll = time.Now()

			log.Printf("Client %s logged in", name)

			return nil
		}
	}

	// Register new client
	player.Name = name
	clients = append(clients, &Client{
		Player:   player,
		LastPoll: time.Now(),
		LoggedIn: true,
	})

	log.Printf("Client %s registered", name)

	return nil
}

func checkTimeout() {
	ticker := time.NewTicker(time.Duration(pollRate * int(time.Millisecond)))
	checker := make(chan struct{})

	// Async loop function
	go func() {
		for {
			select {
			case <-ticker.C:
				for _, c := range clients {
					if c.LoggedIn {
						// Check if client has timed out
						fmt.Println(time.Since(c.LastPoll))
						fmt.Println(time.Duration(pollTimeout * int(time.Millisecond)))
						if time.Since(c.LastPoll) > time.Duration(pollTimeout*int(time.Millisecond)) {
							// Logout client
							c.LoggedIn = false
							log.Println("Client " + c.Name + " timed out")
						}
					}
				}
			case <-checker:
				ticker.Stop()
				return
			}
		}
	}()
}

func (a *API) Poll(name string, state *State) error {
	for _, c := range clients {
		if c.Name == name {
			c.LastPoll = time.Now()
			return nil
		}
	}
	return nil
}

func (a *API) NewGame(player Client, game *Game) error {

	// Check if game already exists
	for _, g := range games {
		// Check if game name is already being used
		if g.Name == game.Name {
			log.Printf("Client %s could create game with name \"%s\", already exists", player.Name, game.Name)
			return fmt.Errorf("game with name \"%s\" already exists", game.Name)
		}
	}

	// Log game creation to console
	log.Printf("Client %s created new game \"%s\"", player.Name, game.Name)

	return nil
}

// GetGameState of tic-tac-toe game
func (a *API) GetGameState(player *Client, game *Game) error {

	return nil
}
