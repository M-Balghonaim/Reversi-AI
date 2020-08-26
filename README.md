### About
This project implements Reversi using Golang. You interact with the game using a terminal. There are two versions of the game: `reversi` and `reversiSimulation`. The `reversi` program is player vs. AI. The `reversiSimulation` program plays an AI (red player) that uses heuristics against another AI (blue player) that does not use heuristics.

The AIs are search-based and use Monte Cristro Tree Search. 

### How to run the code:
#### Using Docker:
1. In the root of the project, run:
    1. Build the image: `docker build . -t reversi-games`
        2. Run regular reversi (player vs. computer):  `docker run --rm -ti reversi-games ./reversi`
        3. Run simulation reversi (computer vs. computer): `docker run --rm -ti reversi-games ./reversiSimulation`
      
#### Run locally:

1. Copy the packages `reversi` and `reversiSimulation` to your $GOPATH/src (if your GOPATH env. variable is set):
2. Run either packages:
    1. Player vs. Computer: `go run reversi`
    2. Computer (red and uses heuristics) vs. Computer (blue and no heuristics): `go run reversiSimulation`


### Please note:

* The language used is Go (v1.14)
* The maximum amount of time a computer can take to pick its next move is 10 seconds
* There are two version of the program: 
    1. reversi: player vs. computer (heuristics)
    2. reversiSimulation: red computer (heuristics) vs. blue computer (no heuristics)
* Some windows terminal fonts lack CJK characters. If the chip (⬤) does not display correctly, change your terminal font to `SimSun-ExtB`:
    1. Open cmd
    2. Right-click cmd terminal icon top-left of the window
    3. Click 'properties' -> 'font' and set the font
    
