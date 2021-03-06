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

// User registered with server
type User struct {
	Name string
}

// Player of tic-tac-toe
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

// Session connected to server
type Session struct {
	*User
	LoggedIn bool
	LastPoll time.Time
}

var (
	sessionTimeout = time.Duration(2 * time.Second)
	gamePort       = 27960 // My favorite port 27960 because old quake and wolfenstein :)
	games          []*Game
	sessions       []*Session
)

// TTT server for tic-tac-toe RPC interface
type TTT int

func main() {
	fmt.Println("--- TIC-TAC-TOE BACKEND ---")

	ttt := new(TTT)
	err := rpc.Register(ttt)

	if err != nil {
		log.Fatal("Failed to register rpc:", err)
	}

	rpc.HandleHTTP()

	portString := fmt.Sprint(gamePort)
	lis, err := net.Listen("tcp", ":"+portString)

	if err != nil {
		log.Fatal("Failed to listen on port "+portString+":", err)
	}

	AuditSessions()

	log.Println("Server started on port " + portString)
	err = http.Serve(lis, nil)

	if err != nil {
		log.Fatal("Failed to serve http server:", err)
	}
}

// AuditSessions for timeouts
func AuditSessions() {

	ticker := time.NewTicker(sessionTimeout)
	checker := make(chan struct{})

	// Asynchronous check for sessions that have timed out
	go func() {
		for {
			select {
			case <-ticker.C:
				for _, s := range sessions {
					if s.LoggedIn {
						// Check if session has timed out
						if time.Since(s.LastPoll) > sessionTimeout {
							// Logout session
							s.LoggedIn = false

							log.Println("User " + s.Name + " timed out")

							// Kick user from any games
							s.User.kick()
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

func (u *User) auth() error {
	for _, s := range sessions {
		if s.Name == u.Name {
			if s.LoggedIn {
				s.LastPoll = time.Now()
				return nil
			}
			return fmt.Errorf("user \"%s\" is not logged in", u.Name)
		}
	}
	return fmt.Errorf("user \"%s\" is not registered", u.Name)
}

func (u *User) kick() {
	for _, g := range games {
		for i := 0; i < len(g.Players); i++ {

			// Check if player is in game
			if g.Players[i].Name == u.Name {

				// Remove player from player list
				g.Players = append(g.Players[:i], g.Players[i+1:]...)

				// If only one player remaining, set to winner
				if g.Turn >= 0 && len(g.Players) == 1 && g.Winner == nil {
					g.Winner = &g.Players[0]
				}
				return
			}
		}
	}
}

func (u *User) game() *Game {
	for _, g := range games {
		for _, p := range g.Players {
			if p.Name == u.Name {
				return g
			}
		}
	}
	return nil
}

// Register client with server
func (a *TTT) Register(name string, user *User) error {

	// Remove any unwanted characters
	user.Name = strings.ReplaceAll(name, "\n", "")

	// Ensure name meets criteria
	if user.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if len(user.Name) > 32 {
		return fmt.Errorf("name must be under 32 characters")
	}

	for _, s := range sessions {
		if s.Name == user.Name {
			if s.LoggedIn {
				return fmt.Errorf("user \"%s\" is already logged in", user.Name)
			}
			// Session has returned from previous login
			s.LoggedIn = true
			s.LastPoll = time.Now()

			log.Printf("User \"%s\" logged in", user.Name)

			return nil
		}
	}

	// Register new client
	sessions = append(sessions, &Session{
		User:     user,
		LastPoll: time.Now(),
		LoggedIn: true,
	})
	log.Printf("User \"%s\" registered", user.Name)

	return nil
}

// GetGame user is in
func (a *TTT) GetGame(user User, game *Game) error {

	// Ensure user is registered and logged in
	if err := user.auth(); err != nil {
		return err
	}

	// Get game state
	for _, g := range games {
		// Check if user is in game
		for _, p := range g.Players {
			if p.Name == user.Name {

				// Deep copy of tic-tac-toe
				game.Turn = g.Turn
				game.Players = g.Players
				game.MaxPlayers = g.MaxPlayers
				game.Winner = g.Winner
				game.Board = g.Board

				return nil
			}
		}
	}
	return nil
}

// GetLobby of online games for user
func (a *TTT) GetLobby(user User, lobby *Lobby) error {

	// Ensure user is registered and logged in
	if err := user.auth(); err != nil {
		return err
	}

	for _, g := range games {

		// If game hasn't started and isn't full
		if len(g.Players) > 0 && len(g.Players) < g.MaxPlayers && g.Turn < 0 {

			// Make sure player is not already in game
			for _, p := range g.Players {
				if p.Name == user.Name {
					continue
				}
			}
			lobby.Users = append(lobby.Users, g.Players[0].User)
		}
	}
	return nil
}

// NewGame of classic two user tic-tac-toe
func (a *TTT) NewGame(user User, game *Game) error {

	// Ensure user is registered and logged in
	if err := user.auth(); err != nil {
		return err
	}

	// Check if user is already in game
	if user.game() != nil {
		log.Printf("User \"%s\" failed to create game, already in a game", user.Name)
		return fmt.Errorf("you are already in a game")
	}

	// New two player game
	game.Turn = -1
	game.MaxPlayers = 2
	game.Players = []Player{{user, "X"}}
	game.Board = [][]string{{" ", " ", " "}, {" ", " ", " "}, {" ", " ", " "}}

	// Append new game to list of games
	games = append(games, game)

	// Log game creation to console
	log.Printf("User \"%s\" created a game", user.Name)

	return nil
}

// JoinGame of tic-tac-toe on server
func (a *TTT) JoinGame(args []User, game *Game) error {

	// Ensure that proper arguments are given
	if len(args) != 2 {
		return fmt.Errorf("invalid args for JoinGame([]User{you, friend}, *Game)")
	}

	// Ensure user is registered and logged in
	if err := args[0].auth(); err != nil {
		return err
	}

	for _, g := range games {
		for _, p := range g.Players {

			// Check if user is already in game
			if p.Name == args[0].Name {
				log.Printf("User \"%s\" failed to join a game, already in a game", args[0].Name)
				return fmt.Errorf("you are already in a game")
			}

			// Check if game is the one we are looking for
			if p.Name == args[1].Name {
				if len(g.Players) == g.MaxPlayers {
					log.Printf("User \"%s\" failed to join a game with user \"%s\", game full", args[0].Name, args[1].Name)
					return fmt.Errorf("game with user \"%s\" is full, %d/%d players", args[1].Name, len(g.Players), g.MaxPlayers)
				}
				game = g
			}
		}
	}
	if len(game.Players) == 0 {
		log.Printf("User \"%s\" failed to join a game with user \"%s\", not found", args[0].Name, args[1].Name)
		return fmt.Errorf("game with user \"%s\" not found", args[1].Name)
	}
	// Append user to players of game
	game.Players = append(game.Players, Player{args[0], "O"})

	// If there are enough players, start game
	if len(game.Players) == game.MaxPlayers {
		game.Turn = 0
	}
	log.Printf("User \"%s\" joined game with user \"%s\"", args[0].Name, args[1].Name)

	return nil
}

// NewMark on tic-tac-toe board
func (a *TTT) NewMark(mark Mark, unused *Game) error {

	// Ensure user is registered and logged in
	if err := mark.User.auth(); err != nil {
		return err
	}

	// Get game user is in
	for _, g := range games {
		for _, p := range g.Players {
			if p.Name == mark.User.Name {

				// Make sure game has started
				if g.Turn < 0 {
					return fmt.Errorf("game has not started")
				}

				// Make sure game is not already won
				if g.Winner != nil {
					return fmt.Errorf("game is already won")
				}

				// Make sure it's your turn
				if mark.User.Name != g.Players[g.Turn%len(g.Players)].Name {
					return fmt.Errorf("it's not your turn")
				}

				// Make sure mark is within board
				if mark.X >= 0 && mark.Y >= 0 && mark.X < len(g.Board) && mark.Y < len(g.Board[0]) {

					// Make sure spot is empty
					spot := g.Board[mark.X][mark.Y]
					if spot == " " {
						g.Board[mark.X][mark.Y] = p.Symbol
						g.Turn++

						log.Printf("user \"%s\" marked \"%s\" at %d,%d", mark.User.Name, p.Symbol, mark.X, mark.Y)

						// Check if too few moves to win
						if g.Turn < len(g.Board)*len(g.Players)-1 {
							return nil
						}

						// Check horizontal win
						for x := 0; x < len(g.Board); x++ {
							for y, symbol := range g.Board[x] {
								if symbol != p.Symbol {
									break
								}
								if y == len(g.Board[x])-1 {
									g.Winner = &p
									return nil
								}
							}
						}

						// Check vertical win
						for y := 0; y < len(g.Board[0]); y++ {
							for x := 0; x < len(g.Board); x++ {
								if g.Board[x][y] != p.Symbol {
									break
								}
								if x == len(g.Board)-1 {
									g.Winner = &p
									return nil
								}
							}
						}

						// Check diagonal win
						for x, y := 0, 0; x < len(g.Board); x, y = x+1, y+1 {
							if g.Board[x][y] != p.Symbol {
								break
							}
							if x == len(g.Board)-1 {
								g.Winner = &p
								return nil
							}
						}

						// Check counter diagonal win
						for x, y := 0, len(g.Board)-1; x < len(g.Board) && y >= 0; x, y = x+1, y-1 {
							if g.Board[x][y] != p.Symbol {
								break
							}
							if x == len(g.Board)-1 {
								g.Winner = &p
								return nil
							}
						}
						// Player did not win, return
						return nil
					}
					return fmt.Errorf("mark \"%s\" already exists at %d,%d", spot, mark.X, mark.Y)
				}
				return fmt.Errorf("mark is out of bounds %d, %d", mark.X, mark.Y)
			}
		}
	}
	return fmt.Errorf("user \"%s\" is not in a game", mark.User.Name)
}

// QuitGame of tic-tac-toe
func (a *TTT) QuitGame(user User, unused *Game) error {

	// Ensure user is registered and logged in
	if err := user.auth(); err != nil {
		return err
	}

	// Check if user is not in game
	if user.game() == nil {
		return fmt.Errorf("user \"%s\" is not in game", user.Name)
	}

	user.kick()
	log.Printf("user \"%s\" quit their game", user.Name)

	return nil
}
