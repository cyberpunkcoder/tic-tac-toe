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
	"strings"
	"time"

	// I tried so so hard to get gRPC to work, I couldn't in time :(
	// It would have been so much cleaner, more robust and secure
	"net/rpc"
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

// State of player in tic-tac-toe
type State struct {
	Game  *Game
	Games []Game
}

// Client connected to server
type Client struct {
	*Player
	LoggedIn bool
	LastPoll time.Time
}

var (
	pollTimeout = time.Duration(2 * time.Second) // Timeout for poll between client and server
	gamePort    = 27960                          // My favorite port 27960 because old quake and wolfenstein :)
	games       []Game
	clients     []*Client
)

// API server for tic-tac-toe RPC interface
type API int

func main() {
	api := new(API)
	err := rpc.Register(api)

	if err != nil {
		log.Fatal("Failed to register rpc:", err)
	}

	rpc.HandleHTTP()

	portString := fmt.Sprint(gamePort)
	lis, err := net.Listen("tcp", ":"+portString)

	if err != nil {
		log.Fatal("Failed to listen on port "+portString+":", err)
	}

	CheckTimeout()

	log.Println("Server started on port " + portString)
	err = http.Serve(lis, nil)

	if err != nil {
		log.Fatal("Failed to serve http server:", err)
	}

}

// CheckTimeout of each client
func CheckTimeout() {
	ticker := time.NewTicker(pollTimeout)
	checker := make(chan struct{})

	// Asynchronous check for clients that have timed out
	go func() {
		for {
			select {
			case <-ticker.C:
				for _, c := range clients {
					if c.LoggedIn {
						// Check if client has timed out
						if time.Since(c.LastPoll) > pollTimeout {
							// Logout client
							c.LoggedIn = false
							log.Println("Player " + c.Name + " timed out")
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

// Register client with server
func (a *API) Register(name string, player *Player) error {
	// Remove any unwanted characters
	player.Name = strings.ReplaceAll(name, "\n", "")

	// Ensure name meets criteria
	if player.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if len(player.Name) > 32 {
		return fmt.Errorf("name must be under 32 characters")
	}

	for _, c := range clients {
		if c.Name == player.Name {
			if c.LoggedIn {
				return fmt.Errorf("player \"%s\" is already logged in", player.Name)
			}
			// Client has returned from previous login
			c.LoggedIn = true
			c.LastPoll = time.Now()

			log.Printf("Player \"%s\" logged in", player.Name)

			return nil
		}
	}

	// Register new client
	clients = append(clients, &Client{
		Player:   player,
		LastPoll: time.Now(),
		LoggedIn: true,
	})
	log.Printf("Player \"%s\" registered", player.Name)

	return nil
}

// GetState of player in server
func (a *API) GetState(player Player, state *State) error {
	for _, c := range clients {
		if c.Name == player.Name {
			c.LastPoll = time.Now()
			state.Games = games

			// Check if player is in a game
			for _, g := range state.Games {
				for _, p := range g.Players {
					if p.Name == player.Name {
						state.Game = &g
						return nil
					}
				}
			}
			state.Game = nil
			return nil
		}
	}
	return fmt.Errorf("player \"%s\" is already logged in", player.Name)
}

// NewGame of classic two player tic-tac-toe
func (a *API) NewGame(player Player, state *State) error {
	// Check if player is already in game
	for _, g := range games {
		for _, p := range g.Players {
			if p.Name == player.Name {
				log.Printf("Player \"%s\" failed to create game, already in game \"%s\"", p.Name, g.Name)
				return fmt.Errorf("player \"%s\" is already in game \"%s\"", p.Name, g.Name)
			}
		}
	}
	// Set player symbol to X since they will host the game
	player.Symbol = "X"

	// Create a new game struct
	game := Game{
		Name:       player.Name + "'s game",
		Players:    []Player{player},
		Board:      [][]string{{" ", " ", " "}, {" ", " ", " "}, {" ", " ", " "}},
		MaxPlayers: 2,
	}
	// Check if game already exists
	for _, g := range games {
		// Check if game name is already being used
		if g.Name == game.Name {
			log.Printf("Player \"%s\" failed to create game \"%s\", already exists", player.Name, game.Name)
			return fmt.Errorf("game \"%s\" already exists", game.Name)
		}
	}
	// Append new game to list of games
	games = append(games, game)

	// Log game creation to console
	log.Printf("Player \"%s\" created game \"%s\"", player.Name, game.Name)

	return nil
}
