package main

func main() {

	// Initialize a new game
	game := NewGame()

	// Play turns
	for !game.End {
		game.PlayTurn()
	}
}
