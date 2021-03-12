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

	// I tried so so hard to get gRPC to work, I couldn't in time :(
	"net/rpc"
)

var (
	// My favorite port 27960 because old quake and wolfenstein :)
	gamePort = 27960
	players  []Player
)

// TTT server RPC struct
type TTT int

// Player of Tic-Tac-Toe game
type Player struct {
	Name string
}

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
	// Check if a player with same name exists
	for _, p := range players {
		if name == p.Name {
			log.Printf("Player %s could not join, name in use", name)
			return fmt.Errorf("name \"%s\" is already in use", name)
		}
	}

	// Create new player and add to players
	newPlayer := Player{name}
	players = append(players, newPlayer)

	// Log player join to console
	log.Printf("Player %s joined", name)

	return nil
}
