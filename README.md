# Tic-Tac-Toe
> Tic Tac Toe backend and frontend!
> This software allows multiplayer Tic-Tac-Toe games and is extra lightweight (No non-go-vanilla packages!).

## Video Of Game
https://youtu.be/X8L_hQv7Qnk

## Quick Start
```git clone https://github.com/cyberpunkcoder/tic-tac-toe.git && cd cd tic-tac-toe```
1) Start the TTT backend with ```go run backend/cmd/main.go```
2) Open another shell in the same directory
3) Start a TTT frontend with ```go run backend/cmd/main.go```
4) Lonely? Start another TTT frontend and play multiplayer Tic-Tac-Toe!

## Lobby
```
--- TIC-TAC-TOE ---
Enter user name: sally

--- GAME LOBBY ---

 Game found: bob's game
 Game found: sue's game
 Game found: sarah's game

--- COMMANDS ---
 exit                   - exit program
 create                 - make new lobby game
 join <player name>     - join lobby game (example: join bob)
 ```
 ## Gameplay
 ```
 --- TIC-TAC-TOE GAME ---
Player: bob     Symbol: X
Player: sally   Symbol: O

    1   2   3
   ╔═══╦═══╦═══╗
a ║   ║ O ║   ║
   ╠═══╬═══╬═══╣
b ║ X ║ X ║ O ║
   ╠═══╬═══╬═══╣
c ║ X ║   ║   ║
   ╚═══╩═══╩═══╝

--- COMMANDS ---
 exit                   - exit program
 quit                   - quit game
 <x y coordinates>      - mark spot on board (example: a 1)

YOUR TURN
```
