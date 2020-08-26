package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
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

// To track statistics across simulated games
var redWins int = 0
var blueWins int = 0
var ties int = 0

// Game struct
type Reversi struct {
	board             []int
	turn              int
	End               bool
	computerTwoColor  int // using heuristics
	computerOneColor  int // not using heuristics
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

	game.computerTwoColor = red
	game.computerOneColor = blue

	game.turn = game.computerTwoColor

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

	// Get each score
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
	cpy.computerOneColor = r.computerOneColor
	cpy.computerTwoColor = r.computerTwoColor
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
func doPlayOut(r *Reversi, useHeuristics bool) int {
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
		return doPlayOut(r, useHeuristics)
	}
	var pos int

	// Get the best move depending on whether or not the called wants us to use heuristics
	if useHeuristics == true {
		pos = getHeuristicPos(positions)
	} else {
		pos = getRandPos(positions)
	}
	r.setChip(pos)
	r.switchTurns()
	return doPlayOut(r, useHeuristics)
}

// Return the best move using MCT
// The list of best, good, bad, and worst positions is defined at the top of the file
func (r *Reversi) getBestMove(useHeuristics bool) int {

	// Create hash map to store move scores
	scores := make(map[int]int)
	positions := r.getValidPositions()

	// If there are no valid positions
	if positions == nil {
		return -1
	}

	startTime := time.Now()
	numPlayOuts := 0
	timeLimitExceeded := false

	// MCT
	for _, pos := range positions {
		scores[pos] = 0

		if timeLimitExceeded {
			break
		}

		for i := 1; i <= playouts; i++ {

			// Increment number of playouts
			numPlayOuts += 1

			// If more than 10 seconds have elapsed since we started all playouts, end early
			if time.Since(startTime).Seconds() > 10 {
				fmt.Print("\nMax amount of time exceeded. Making decision...\n")
				timeLimitExceeded = true
				break
			}

			cpy := r.deepCopy()
			cpy.setChip(pos)
			cpy.switchTurns()
			result := doPlayOut(cpy, useHeuristics)

			// Add weighted scores based on result
			if result == r.turn {
				scores[pos] += 2
			} else if result == r.turn*-1 {
				scores[pos] -= 10
			} else {
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

func (r *Reversi) switchTurns() {
	r.turn = r.turn * -1
}

// Play blue computer's turn, which does not use heuristics
func (r *Reversi) playBlueTurn() {
	fmt.Print("Computer 1 (BLUE) thinking without heuristics....")

	// Get the best move without heuristics (random playouts)
	pos := r.getBestMove(false)

	// If the red computer has no moves to make, pass the turn
	if pos == -1 {
		fmt.Print("Skipping turn.")
		r.switchTurns()
		return
	}

	r.setChip(pos)

	r.switchTurns()
}

// Play red computer's turn, which does uses heuristics
func (r *Reversi) playRedTurn() {
	fmt.Print("Computer 2 (RED) thinking with heuristics....")

	// Get the best move with heuristics
	pos := r.getBestMove(true)

	// If the red computer has no moves to make, pass the turn
	if pos == -1 {
		fmt.Print("Skipping turn.")
		r.switchTurns()
		return
	}

	r.setChip(pos)

	r.switchTurns()
	return
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

	// If there is a winner, tie, or both computers have no remaining moves
	if winResult < 2 || (currPositions == nil && nextPositions == nil) {
		winString := ""
		// If winResult is blue
		if winResult == blue {
			fmt.Print("\033[94mBlue\033[0m has won.\n\n")
			blueWins += 1
			winString = "Blue has won.\n"
			// If winResult is red
		} else if winResult == red {
			fmt.Print("\033[91mRed\033[0m has won.\n\n")
			winString = "Red has won.\n"
			redWins += 1
			// If the game is a tie
		} else if winResult == tie {
			fmt.Print("\033[93mIt's a tie!\033[0m\n\n")
			winString = "It's a tie.\n"
			ties += 1
		}

		fmt.Printf("\nThe average number of playouts per second is: %v\n", r.getAvgPlayOutsPerSecond())
		fmt.Printf("\nThe average MCT turn: %v\n", r.getAvgMctTime())

		fmt.Printf("\nBlue wins: %v", blueWins)
		fmt.Printf("\nRed wins: %v", redWins)
		fmt.Printf("\nTie wins: %v", ties)

		// Source: https://golang.org/pkg/os/#example_OpenFile_append
		// If the file doesn't exist, create it, or append to the file
		f, err := os.OpenFile("results2.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		// Write to file
		if _, err := f.Write([]byte(winString)); err != nil {
			log.Fatal(err)
		}

		// Close file
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}

		// Reset the game
		r.reset()

		return
	}

	if r.turn == blue {
		r.playBlueTurn()
	} else {
		r.playRedTurn()
	}

	fmt.Print("\n\n\n")
}
