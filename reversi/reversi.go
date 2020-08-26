package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// Initializing constants
const lineSep string = "\n----------------------------------------------------------------------------------------------------------------------"
const blue int = 1
const red int = -1
const tie int = 0
const maxChips int = 64
const playouts int = 500

// Used by MCT heuristic function
var corners map[int]bool = map[int]bool{0: true, 7: true, 56: true, 63: true}
var badPositions map[int]bool = map[int]bool{1: true, 8: true, 6: true, 15: true, 55: true, 62: true, 57: true, 48: true}
var worstPositions map[int]bool = map[int]bool{9: true, 14: true, 54: true, 49: true}

// Game struct
type Reversi struct {
	board             []int
	turn              int
	End               bool
	playerColor       int
	computerColor     int
	playOutsPerSecond []float64
	mctTime           []float64
}

// Initialize and return a new game instance
func NewGame() *Reversi {

	// Create new game
	game := new(Reversi)

	// Initialize empty board
	game.board = make([]int, maxChips)

	// Set initial four chips
	game.board[27] = red
	game.board[36] = red
	game.board[28] = blue
	game.board[35] = blue

	// User input vars
	var turn string
	var color string

	// Set player color
	fmt.Print("Please select a color of (r)ed or (b)lue chips: ")
	_, _ = fmt.Scan(&color)

	// Trim and lowercase input
	color = strings.ToLower(color)
	color = strings.TrimSpace(color)

	// Assign colors to both sides
	if color == "b" || color == "blue" {
		game.playerColor = blue
		game.computerColor = red
	} else {
		game.playerColor = red
		game.computerColor = blue
	}

	// Set player turn
	fmt.Print("Enter '1' to play first, or enter '2' to play second: ")
	_, _ = fmt.Scan(&turn)
	if turn == "1" {
		game.turn = game.playerColor
	} else {
		game.turn = game.computerColor
	}

	return game
}

// Reset the current game instance
func (r *Reversi) reset() {
	*r = *NewGame()
}

// Get the display string for a chip
func (r *Reversi) getDisplayChar(ind, code int) string {
	// If the position is empty (coded 0)
	if code == 0 {
		// If it's the player's turn, display color-code their valid next positions
		if r.turn == r.playerColor {
			positions := r.getValidPositions()
			if isInValidPositions(strconv.Itoa(ind), positions) {
				return "\033[92m" + strconv.Itoa(ind) + "\033[0m"
			}

		}
		return strconv.Itoa(ind)
	} else if code == -1 {
		// Circle icon unicode is \u2B24
		// Color-code red
		return "\033[91m" + "\u2B24 " + "\033[0m"
	} else if code == 1 {
		// Color-code blue
		return "\033[94m" + "\u2B24 " + "\033[0m"
	} else {
		panic("Unknown display code given")
	}
}

// Get the score of a given color
func (r *Reversi) getScore(color int) int {
	score := 0
	for _, elm := range r.board {
		// If the chip is of the given color, increment score
		if elm == color {
			score += 1
		}
	}
	return score
}

// Get the score of blue
func (r *Reversi) getBlueScore() int {
	return r.getScore(blue)
}

// Get the score of red
func (r *Reversi) getRedScore() int {
	return r.getScore(red)
}

// Display the game board
func (r *Reversi) Display() {

	// Display board
	for i, elm := range r.board {
		if i%8 == 0 {
			if i != 0 {
				fmt.Println(lineSep)
			}
			fmt.Printf("%v\t", r.getDisplayChar(i, elm))
		} else {
			fmt.Printf("|\t%v\t", r.getDisplayChar(i, elm))
		}
	}

	// Display score
	fmt.Printf("\n\nBlue score:\t\033[94m%v\033[0m", r.getBlueScore())
	fmt.Printf("\nRed score:\t\033[91m%v\033[0m\n\n", r.getRedScore())
}

// Determine who won based on given scores
func determineWinner(blueScore, redScore int) int {
	if blueScore > redScore {
		// If blue has won
		return blue
	} else if blueScore < redScore {
		// If red has won
		return red
	} else {
		// If tie
		return tie
	}
}

// If someone has won.
// -1: red has won
// 0: it's a tie
// 1: blue has won
// 2: nobody has won yet
func (r *Reversi) checkWin(forceWin bool) int {

	// Get scores for each color
	blueScore := r.getBlueScore()
	redScore := r.getRedScore()

	// If there are no empty positions or the caller of this function knows that neither players can make a move, then they can opt to end it early.
	if blueScore+redScore == maxChips || forceWin {
		return determineWinner(blueScore, redScore)
	}

	// If the game is still on-going
	return 2
}

// Check if the given position is valid within given direction
func (r *Reversi) checkDirection(pos int, dir int, isWithinLimit func(int, int) bool) bool {

	// Initialize some helper variables
	foundOpposite := false
	currPos := pos

	// Loop while still within board limits
	for isWithinLimit(currPos, dir) == true {
		currPos = currPos + dir

		// If chip is not empty and is opposite color
		if r.board[currPos] == r.turn*-1 {
			foundOpposite = true
		}

		// If chip is not empty and is current color
		if r.board[currPos] == r.turn {
			if foundOpposite {
				return true
			}
			return false
		}

		// If it's empty
		if r.board[currPos] == 0 {
			return false
		}
	}

	return false
}

// Check if the a chip can be placed in given position
func (r *Reversi) isValidPosition(pos int) bool {

	// If position is empty
	if r.board[pos] == 0 {

		// Return true if any direction is valid, as in chips of the opposite color are sandwiched
		// between the current empty space and another chip of the current color.

		// check up
		if r.checkDirection(pos, -8, func(curr int, dir int) bool { return curr+dir >= 0 }) ||
			// check left
			r.checkDirection(pos, -1, func(curr int, dir int) bool { return curr%8 != 0 }) ||
			// check below
			r.checkDirection(pos, 8, func(curr int, dir int) bool { return curr+dir < 64 }) ||
			// check right
			r.checkDirection(pos, 1, func(curr int, dir int) bool { return (curr+1)%8 != 0 }) ||
			// check right
			r.checkDirection(pos, 1, func(curr int, dir int) bool { return (curr+1)%8 != 0 }) ||
			// check up-left
			r.checkDirection(pos, -9, func(curr int, dir int) bool { return (curr+dir >= 0) && (curr%8 != 0) }) ||
			// check below-left
			r.checkDirection(pos, 7, func(curr int, dir int) bool { return (curr+dir < 64) && (curr%8 != 0) }) ||
			// check below-right
			r.checkDirection(pos, 9, func(curr int, dir int) bool { return (curr+dir < 64) && ((curr+1)%8 != 0) }) ||
			// check up-right
			r.checkDirection(pos, -7, func(curr int, dir int) bool { return (curr+dir >= 0) && ((curr+1)%8 != 0) }) {
			return true
		}

	}

	return false
}

// Get a deep copy of the current game
func (r *Reversi) deepCopy() *Reversi {
	cpy := new(Reversi)

	cpy.board = make([]int, maxChips)
	copy(cpy.board, r.board)
	cpy.computerColor = r.computerColor
	cpy.playerColor = r.playerColor
	cpy.turn = r.turn
	cpy.End = r.End

	return cpy
}

// Check if the given position is valid within given direction
func (r *Reversi) flipDirection(pos int, dir int, isWithinLimit func(int, int) bool) bool {

	// Initialize some helper variables
	foundOpposite := false
	performFlips := false
	currPos := pos

	// Loop while still within physical limits of board
	for isWithinLimit(currPos, dir) == true {
		currPos = currPos + dir

		// If chip is not empty and is opposite color
		if r.board[currPos] == r.turn*-1 {
			foundOpposite = true
		}

		// If chip is not empty and is current color
		if r.board[currPos] == r.turn {
			if foundOpposite {
				performFlips = true
				break
			}
			return false
		}

		if r.board[currPos] == 0 {
			return false
		}
	}

	// If some enemy chips have been "captured"
	if performFlips == true {

		// Set back to original position
		currPos = pos

		for isWithinLimit(currPos, dir) == true {
			currPos = currPos + dir

			// If chip is not empty and is current color
			if r.board[currPos] == r.turn*-1 {
				r.board[currPos] = r.turn
				continue
			}

			// If chip is not empty and is opposite color
			if r.board[currPos] == r.turn || r.board[currPos] == 0 {
				break
			}

		}
	}
	return false
}

// Set the given position to the current color chip
func (r *Reversi) setChip(pos int) {

	// If position is empty
	if r.board[pos] == 0 {

		r.board[pos] = r.turn

		// Return true if any direction is valid, as in chips of the opposite color are sandwiched
		// between the current empty space and another chip of the current color.

		// flip up
		r.flipDirection(pos, -8, func(curr int, dir int) bool { return curr+dir >= 0 })

		// flip left
		r.flipDirection(pos, -1, func(curr int, dir int) bool { return curr%8 != 0 })

		// flip below
		r.flipDirection(pos, 8, func(curr int, dir int) bool { return curr+dir < 64 })

		// flip right
		r.flipDirection(pos, 1, func(curr int, dir int) bool { return (curr+1)%8 != 0 })

		// flip right
		r.flipDirection(pos, 1, func(curr int, dir int) bool { return (curr+1)%8 != 0 })

		// flip up-left
		r.flipDirection(pos, -9, func(curr int, dir int) bool { return (curr+dir >= 0) && (curr%8 != 0) })

		// flip below-left
		r.flipDirection(pos, 7, func(curr int, dir int) bool { return (curr+dir < 64) && (curr%8 != 0) })

		// flip below-right
		r.flipDirection(pos, 9, func(curr int, dir int) bool { return (curr+dir < 64) && ((curr+1)%8 != 0) })

		// flip up-right
		r.flipDirection(pos, -7, func(curr int, dir int) bool { return (curr+dir >= 0) && ((curr+1)%8 != 0) })

	}
}

// Get a random integer within range of given values
func getRandInt(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

// Return best position based on heuristics
// The list of best, good, bad, and worst positions is defined at the top of the file
func getHeuristicPos(positions []int) int {

	// Build a "best" list
	var bestList []int
	for _, pos := range positions {
		if _, ok := corners[pos]; ok {
			bestList = append(bestList, pos)
		}
	}

	// Return a random position from the best list if not empty
	if bestList != nil {
		return getRandPos(bestList)
	}

	// Build a "good" list
	var goodList []int
	for _, pos := range positions {
		if _, ok := badPositions[pos]; !ok {
			if _, ok := worstPositions[pos]; !ok {
				goodList = append(goodList, pos)
			}
		}
	}

	// Return a random position from the good list if not empty
	if goodList != nil {
		return getRandPos(goodList)
	}

	// Build a "bad" list
	var badList []int
	for _, pos := range positions {
		if _, ok := badPositions[pos]; ok {
			badList = append(badList, pos)
		}
	}

	// Return a random position from the bad list if not empty
	if badList != nil {
		return getRandPos(badList)
	}

	// If all of the above failed, return a random position
	return getRandPos(positions)
}

// Get a random position from a given list
func getRandPos(positions []int) int {
	rndNum := getRandInt(0, len(positions))
	return positions[rndNum]
}

// Perform a playout for the given instance of reversi
func doPlayOut(r *Reversi) int {
	winResult := r.checkWin(false)
	// If someone has won or it's a tie, return the result
	if winResult < 2 {
		return winResult
	}
	positions := r.getValidPositions()
	// If there are no valid positions, pass the turn to the other player
	if positions == nil {
		r.switchTurns()
		positions = r.getValidPositions()
		// If the other player also does not have any valid positions, end the game
		if positions == nil {
			return r.checkWin(true)
		}
		return doPlayOut(r)
	}
	// Get the best move based on heuristics
	pos := getHeuristicPos(positions)
	r.setChip(pos)
	r.switchTurns()
	return doPlayOut(r)
}

// Return the best move using MCT
func (r *Reversi) getBestMove() int {

	// Create hash map to store move scores
	scores := make(map[int]int)
	positions := r.getValidPositions()

	// If there are no valid positions
	if positions == nil {
		return -1
	}

	numPlayOuts := 0
	startTime := time.Now()
	timeLimitExceeded := false

	// MCT
	for _, pos := range positions {
		scores[pos] = 0

		if timeLimitExceeded {
			break
		}

		// For each playout
		for i := 1; i <= playouts; i++ {

			// Increment number of playouts
			numPlayOuts += 1

			// If more than 10 seconds have elapsed since we started all playouts, end early
			if time.Since(startTime).Seconds() > 10 {
				fmt.Print("\nMax amount of time exceeded. Making decision...\n")
				timeLimitExceeded = true
				break
			}

			// Make a deep copy to perform playouts on
			cpy := r.deepCopy()
			cpy.setChip(pos)
			cpy.switchTurns()
			result := doPlayOut(cpy)

			// Add weighted scores based on result
			// If the current user has won
			if result == r.turn {
				scores[pos] += 2
				// If the opponent has won
			} else if result == r.turn*-1 {
				scores[pos] -= 10
			} else {
				// If it's a tie
				scores[pos] += 1
			}
		}
	}

	// Keep track of the average number of playouts per second
	elapsedSeconds := time.Since(startTime).Seconds()
	r.playOutsPerSecond = append(r.playOutsPerSecond, float64(numPlayOuts)/elapsedSeconds)
	r.mctTime = append(r.mctTime, elapsedSeconds)

	maxScore := -int(^uint(0) >> 1)
	bestPos := -1
	// Get the best next move
	for pos, score := range scores {
		if score >= maxScore {
			maxScore = score
			bestPos = pos
		}
	}

	return bestPos
}

// Return slice of valid positions for current turn
func (r *Reversi) getValidPositions() []int {
	var positions []int
	for pos := range r.board {
		if r.isValidPosition(pos) {
			positions = append(positions, pos)
		}
	}
	return positions
}

// Return whether the given position is a member of the given valid positions list
func isInValidPositions(p string, validPositions []int) bool {

	// Convert to str
	pos, err := strconv.Atoi(p)

	if err != nil {
		fmt.Print("Failed to convert input to string.")
		return false
	}

	for _, b := range validPositions {
		if pos == b {
			return true
		}
	}

	return false
}

func (r *Reversi) switchTurns() {
	r.turn = r.turn * -1
}

func (r *Reversi) playPlayerTurn() {

	// Get the valid positions for the player
	positons := r.getValidPositions()

	// If the player has no valid positions, pass the turn
	if positons == nil {
		fmt.Print("Skipping turn.")
		r.switchTurns()
		return
	}

	// Get next player position
	var nextPos string
	fmt.Print("\nPlease enter your next position: ")
	_, _ = fmt.Scan(&nextPos)

	nextPos = strings.TrimSpace(nextPos)

	// If the entered position is invalid
	for !isInValidPositions(nextPos, positons) {
		fmt.Print("\nInvalid position entered. Please enter your next position: ")
		_, _ = fmt.Scan(&nextPos)

		nextPos = strings.TrimSpace(nextPos)
	}

	// Convert to int
	p, _ := strconv.Atoi(nextPos)

	r.setChip(p)
	r.switchTurns()
}

func (r *Reversi) playComputerTurn() {

	fmt.Print("Computer 1 thinking....")

	// Get the best move for the computer using heuristics
	pos := r.getBestMove()

	// If the computer has no moves to make, pass the turn
	if pos == -1 {
		fmt.Print("Skipping turn.")
		r.switchTurns()
		return
	}

	r.setChip(pos)
	r.switchTurns()
}

// Decide who's blue and play their turn
func (r *Reversi) playBlueTurn() {
	// If it's the player who's blue, play the player's turn
	if r.turn == r.playerColor {
		r.playPlayerTurn()
	} else {
		// Else play the computer turn if it's blue
		r.playComputerTurn()
	}
}

// Decide who's red and play their turn
func (r *Reversi) playRedTurn() {
	// If it's the player who's red, play the player's turn
	if r.turn == r.playerColor {
		r.playPlayerTurn()
	} else {
		// Else play the computer turn if it's red
		r.playComputerTurn()
	}
}

func getListAvg(list []float64) float64 {
	sum := float64(0)

	for _, avg := range list {
		sum += avg
	}

	return sum / float64(len(list))
}

// Get the avg number of playouts
func (r *Reversi) getAvgPlayOutsPerSecond() float64 {
	return getListAvg(r.playOutsPerSecond)
}

// Get the avg time of MCT per turn
func (r *Reversi) getAvgMctTime() float64 {
	return getListAvg(r.mctTime)
}

// Drives main game loop
func (r *Reversi) PlayTurn() {
	winResult := r.checkWin(false)

	r.Display()

	currPositions := r.getValidPositions()
	r.switchTurns()
	nextPositions := r.getValidPositions()
	r.switchTurns()

	// If there is a winner, tie, or both players have no remaining moves
	if winResult < 2 || (currPositions == nil && nextPositions == nil) {
		// If winResult is blue
		if winResult == blue {
			fmt.Print("\033[94mBlue\033[0m has won.\n\n")
			// If winResult is red
		} else if winResult == red {
			fmt.Print("\033[91mRed\033[0m has won.\n\n")
			// If the game is a tie
		} else if winResult == tie {
			fmt.Print("\033[93mIt's a tie!\033[0m\n\n")
		}

		fmt.Printf("\nThe average number of playouts per second is: %v\n", r.getAvgPlayOutsPerSecond())
		fmt.Printf("\nThe average MCT turn: %v\n", r.getAvgMctTime())

		// Prompt restart
		var input string
		fmt.Print("Enter 'p' to play again, anything else to quit: ")
		_, _ = fmt.Scan(&input)

		input = strings.ToLower(input)
		input = strings.TrimSpace(input)

		if input == "p" {
			r.reset()
		} else {
			r.End = true
			return
		}
	}

	if r.turn == blue {
		r.playBlueTurn()
	} else {
		r.playRedTurn()
	}

	fmt.Print("\n\n\n")
}
