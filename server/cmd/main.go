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

// Player of Tic-Tac-Toe
type Player struct {
	Name string
}

// Client connected to server
type Client struct {
	*Player
	LoggedIn bool
	LastPoll time.Time
}

// Game of Tic-Tac-Toe
type Game struct {
	Name    string
	xPlayer *Player
	oPlayer *Player
	Board   []rune
	Turn    int
}

var (
	gamePort = 27960 // My favorite port 27960 because old quake and wolfenstein :)
	games    []Game
	clients  []*Client
)

// TTT server for tic-tac-toe RPC interface
type TTT int

func main() {
	ttt := new(TTT)
	err := rpc.Register(ttt)

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

// NewPlayer of Tic-Tac-Toe game
func (a *TTT) NewPlayer(name string, player *Player) error {
	// Check if name is empty
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	// Check if name is too long
	if len(name) > 32 {
		return fmt.Errorf("name must be under 32 characters")
	}

	player.Name = name

	if player.isLoggedIn() {
		return fmt.Errorf("player \"%s\" is already logged in", name)
	}

	clients = append(clients, &Client{
		LastPoll: time.Now(),
		LoggedIn: true,
		Player:   player,
	})

	// Log player join to console
	log.Printf("Player %s logged in", name)

	return nil
}

func checkTimeout() {
	ticker := time.NewTicker(time.Second * 2)
	checker := make(chan struct{})

	// Async loop function
	go func() {
		for {
			select {
			case <-ticker.C:
				for _, c := range clients {
					// If client has not polled server in over two seconds
					if c.isLoggedIn() && time.Since(c.LastPoll) > 2*time.Second {
						// Client is set to logged out
						c.LoggedIn = false
						log.Println("Player " + c.Name + " timed out")
					}
				}
			case <-checker:
				ticker.Stop()
				return
			}
		}
	}()
}

func (p *Player) isLoggedIn() bool {
	for _, c := range clients {
		if c.Name == p.Name {
			return c.LoggedIn
		}
	}
	return false
}

func (a *TTT) Poll(name string, player *Player) error {
	for _, c := range clients {
		if c.Name == name {
			c.LastPoll = time.Now()
			return nil
		}
	}
	return nil
}

func (a *TTT) NewGame(player Player, game *Game) error {

	// Check if game already exists
	for _, g := range games {

		// Check if player is already in a game
		if g.xPlayer.Name == player.Name || g.oPlayer.Name == player.Name {
			log.Printf("Player %s tried to create a second game", player.Name)
			return fmt.Errorf("player %s is already hosting game \"%s\"", player.Name, game.Name)
		}

		// Check if game name is already being used
		if g.Name == game.Name {
			log.Printf("Player %s could create game with name \"%s\", already exists", player.Name, game.Name)
			return fmt.Errorf("game with name \"%s\" already exists", game.Name)
		}
	}

	// Log game creation to console
	log.Printf("Player %s created new game \"%s\"", player.Name, game.Name)

	return nil
}

// GetGameState of tic-tac-toe game
func (a *TTT) GetGameState(player *Player, game *Game) error {

	return nil
}
