package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	wr "github.com/mroth/weightedrand"
	"golang.org/x/image/colornames"
)

const MapSize = 14
const AgentMapSize = MapSize*2 + 1

const TileSize = 40
const SelfTileSize = TileSize / 2

var Margin = 0.99
var DelayInMs = 0

//define constants
var U = 0
var E = 1
var F = 2
var EXIT = 3
var R = 4 //restrictred, means this area inside a loop and no need to visit

//define directions
var North = 0
var South = 1
var East = 2
var West = 3

//define movements
var MoveForward = 0
var MoveBackward = 1
var MoveRight = 2
var MoveLeft = 3

/*var tileMap = [MapSize][MapSize]int{

	{F, F, F, F, EXIT, F, F, F, F, F, F, F, F, F},
	{F, E, F, E, E, E, E, F, E, E, E, E, E, F},
	{F, E, F, E, E, E, E, F, E, E, E, F, E, F},
	{F, E, F, F, E, E, E, F, E, E, E, F, F, F},
	{F, E, E, E, E, E, E, E, E, E, E, E, E, F},
	{F, E, E, E, E, E, E, E, E, E, E, E, E, F},
	{F, E, E, E, E, E, E, E, E, E, F, F, F, F},
	{F, E, E, E, E, E, E, E, E, E, F, E, E, F},
	{F, F, F, E, E, E, E, E, E, E, F, E, E, F},
	{F, E, E, E, F, E, E, E, E, E, E, E, E, F},
	{F, E, E, E, F, E, F, E, E, F, F, F, E, F},
	{F, E, E, E, F, F, F, E, E, E, E, F, E, F},
	{F, E, E, E, E, E, F, E, E, E, E, F, E, F},
	{F, F, F, F, F, F, F, F, F, EXIT, F, F, F, F},
}*/

var tileMap = [MapSize][MapSize]int{

	{F, F, F, F, EXIT, F, F, F, F, F, F, F, F, F},
	{F, E, F, E, E, E, E, E, E, E, E, E, E, F},
	{F, E, F, E, E, E, E, E, E, E, E, F, E, F},
	{F, E, F, F, E, E, E, E, E, E, E, F, F, F},
	{F, E, E, E, E, E, E, E, E, E, E, E, E, F},
	{F, E, E, E, E, E, E, E, E, E, E, E, E, F},
	{F, E, E, E, E, E, E, E, E, E, F, F, F, F},
	{F, E, E, E, E, E, E, E, E, E, F, E, E, F},
	{F, F, F, E, E, E, E, E, E, E, F, E, E, F},
	{F, E, E, E, F, E, E, E, E, E, E, E, E, F},
	{F, E, E, E, F, E, F, E, E, F, F, F, E, F},
	{F, E, E, E, F, F, F, E, E, E, E, F, E, F},
	{F, E, E, E, E, E, F, E, E, E, E, F, E, F},
	{F, F, F, F, F, F, F, F, F, EXIT, F, F, F, F},
}

var tileColour = map[int]color.RGBA{
	E:    colornames.White,
	F:    colornames.Brown,
	U:    colornames.Black,
	EXIT: colornames.Yellow,
	R:    colornames.Gray,
}

type Agent struct {
	currentCol int
	currentRow int

	frontSensor int
	rearSensor  int
	rightSensor int
	leftSensor  int

	color color.RGBA

	currentDirection int
	nextDirection    int

	//self vars
	selfMap              [AgentMapSize][AgentMapSize]int
	selfCurrentCol       int
	selfCurrentRow       int
	selfCurrentDirection int
	selfNextDirection    int

	selfOtherAgentRow int
	selfOtherAgentCol int

	selfPath [][]int
	name     string

	selfFrontSensor int
	selfRearSensor  int
	selfRightSensor int
	selfLeftSensor  int
}

func NewAgent(currentRow int, currentCol int, currentDirection int, color color.RGBA, name string) *Agent {
	return &Agent{
		name:                 name,
		currentCol:           currentCol,
		currentRow:           currentRow,
		currentDirection:     currentDirection,
		color:                color,
		selfCurrentCol:       MapSize,
		selfCurrentRow:       MapSize,
		selfCurrentDirection: North, //initially, agent just assumes he is rotated to north

		selfOtherAgentRow: -1,
		selfOtherAgentCol: -1,

		selfPath: [][]int{{MapSize, MapSize}},
	}
}

//define agents
var agents = []*Agent{
	//NewAgent(12, 7, North, colornames.Orange, "orange"),
	//NewAgent(12, 12, South, colornames.Orange, "orange"),
	//NewAgent(10, 10, South, colornames.Blue, "blue"),
	NewAgent(8, 10, East, colornames.Green, "green"),
	NewAgent(7, 10, West, colornames.Pink, "pink"),
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, MapSize*TileSize, MapSize*TileSize),
		VSync:  true,
	}
	agentMapCfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, MapSize*TileSize, MapSize*TileSize),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	var agentWindows []*pixelgl.Window
	for i := 0; i < len(agents); i++ {
		agentWindow, _ := pixelgl.NewWindow(agentMapCfg)
		agentWindows = append(agentWindows, agentWindow)
	}
	if err != nil {
		panic(err)
	}
	for !win.Closed() {
		win.Clear(colornames.Skyblue)
		drawMap(win)
		drawAgents(win)
		win.Update()

		for i := 0; i < len(agents); i++ {
			agentWindows[i].Clear(colornames.Black)
			drawAgentSelfMap(agentWindows[i], agents[i])
			agentWindows[i].Update()
			drawAgentToSelfMap(agentWindows[i], agents[i])
			agentWindows[i].Update()

		}

		//sense:
		sense()
		//plan:
		plan()
		time.Sleep(time.Duration(DelayInMs) * time.Millisecond)
		//act:
		rotate()
		win.Clear(colornames.Skyblue)
		drawMap(win)
		drawAgents(win)
		win.Update()
		for i := 0; i < len(agents); i++ {
			selfRotate(agents[i])
			agentWindows[i].Clear(colornames.Black)
			drawAgentSelfMap(agentWindows[i], agents[i])
			agentWindows[i].Update()
			drawAgentToSelfMap(agentWindows[i], agents[i])
			agentWindows[i].Update()

		}
		time.Sleep(time.Duration(DelayInMs) * time.Millisecond)
		move()
		for i := 0; i < len(agents); i++ {
			selfMove(agents[i])

		}

	}

}

//real life'da yanımızda bir engel mi var yoksa agent mı var bunu tespit edebiliriz
// ama buradaki simülasyonda bu bilgiyi diğer agentların pozisyonlarını kontrol edereek elde edebiliriz.
func sense() {
	for i := 0; i < len(agents); i++ {

		if agents[i].currentDirection == North {
			agents[i].frontSensor = tileMap[agents[i].currentRow-1][agents[i].currentCol]
			agents[i].rearSensor = tileMap[agents[i].currentRow+1][agents[i].currentCol]
			agents[i].rightSensor = tileMap[agents[i].currentRow][agents[i].currentCol+1]
			agents[i].leftSensor = tileMap[agents[i].currentRow][agents[i].currentCol-1]

			for j := 0; j < len(agents); j++ {
				if agents[i].currentRow-1 == agents[j].currentRow && agents[i].currentCol == agents[j].currentCol {
					switch agents[i].selfCurrentDirection {
					case North:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow - 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case East:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol + 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case West:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol - 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case South:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow + 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					}
				}

				if agents[i].currentCol+1 == agents[j].currentCol && agents[i].currentRow == agents[j].currentRow {
					switch agents[i].selfCurrentDirection {
					case North:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol + 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case East:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow + 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case West:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow - 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case South:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol - 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					}
				}
				if agents[i].currentCol-1 == agents[j].currentCol && agents[i].currentRow == agents[j].currentRow {
					switch agents[i].selfCurrentDirection {
					case North:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol - 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case East:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow - 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case West:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow + 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case South:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol + 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					}
				}

				if agents[i].currentRow+1 == agents[j].currentRow && agents[i].currentCol == agents[j].currentCol {
					switch agents[i].selfCurrentDirection {
					case North:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow + 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case East:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol - 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case West:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol + 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case South:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow - 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					}
				}

			}

		} else if agents[i].currentDirection == East {
			agents[i].frontSensor = tileMap[agents[i].currentRow][agents[i].currentCol+1]
			agents[i].rearSensor = tileMap[agents[i].currentRow][agents[i].currentCol-1]
			agents[i].rightSensor = tileMap[agents[i].currentRow+1][agents[i].currentCol]
			agents[i].leftSensor = tileMap[agents[i].currentRow-1][agents[i].currentCol]
			for j := 0; j < len(agents); j++ {
				if agents[i].currentRow-1 == agents[j].currentRow && agents[i].currentCol == agents[j].currentCol {
					switch agents[i].selfCurrentDirection {
					case North:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol - 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case East:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow - 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case West:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow + 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case South:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol + 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					}
				}

				if agents[i].currentCol+1 == agents[j].currentCol && agents[i].currentRow == agents[j].currentRow {
					switch agents[i].selfCurrentDirection {
					case North:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow - 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case East:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol + 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case West:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol - 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case South:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow + 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					}
				}
				if agents[i].currentCol-1 == agents[j].currentCol && agents[i].currentRow == agents[j].currentRow {
					switch agents[i].selfCurrentDirection {
					case North:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow + 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case East:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol - 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case West:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol + 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case South:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow - 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					}
				}

				if agents[i].currentRow+1 == agents[j].currentRow && agents[i].currentCol == agents[j].currentCol {
					switch agents[i].selfCurrentDirection {
					case North:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol + 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case East:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow + 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case West:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow - 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case South:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol - 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					}
				}
			}
		} else if agents[i].currentDirection == South {
			agents[i].frontSensor = tileMap[agents[i].currentRow+1][agents[i].currentCol]
			agents[i].rearSensor = tileMap[agents[i].currentRow-1][agents[i].currentCol]
			agents[i].rightSensor = tileMap[agents[i].currentRow][agents[i].currentCol-1]
			agents[i].leftSensor = tileMap[agents[i].currentRow][agents[i].currentCol+1]

			for j := 0; j < len(agents); j++ {
				if agents[i].currentRow-1 == agents[j].currentRow && agents[i].currentCol == agents[j].currentCol {
					switch agents[i].selfCurrentDirection {
					case North:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow + 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case East:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol - 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case West:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol + 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case South:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow - 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					}
				}

				if agents[i].currentCol+1 == agents[j].currentCol && agents[i].currentRow == agents[j].currentRow {
					switch agents[i].selfCurrentDirection {
					case North:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol - 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case East:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow - 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case West:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow + 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case South:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol + 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					}
				}
				if agents[i].currentCol-1 == agents[j].currentCol && agents[i].currentRow == agents[j].currentRow {
					switch agents[i].selfCurrentDirection {
					case North:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol + 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case East:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow + 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case West:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow - 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case South:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol - 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					}
				}

				if agents[i].currentRow+1 == agents[j].currentRow && agents[i].currentCol == agents[j].currentCol {
					switch agents[i].selfCurrentDirection {
					case North:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow - 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case East:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol + 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case West:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol - 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case South:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow + 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					}
				}
			}

		} else {
			agents[i].frontSensor = tileMap[agents[i].currentRow][agents[i].currentCol-1]
			agents[i].rearSensor = tileMap[agents[i].currentRow][agents[i].currentCol+1]
			agents[i].rightSensor = tileMap[agents[i].currentRow-1][agents[i].currentCol]
			agents[i].leftSensor = tileMap[agents[i].currentRow+1][agents[i].currentCol]

			for j := 0; j < len(agents); j++ {
				if agents[i].currentRow-1 == agents[j].currentRow && agents[i].currentCol == agents[j].currentCol {
					switch agents[i].selfCurrentDirection {
					case North:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol + 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case East:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow + 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case West:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow - 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case South:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol - 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					}
				}

				if agents[i].currentCol+1 == agents[j].currentCol && agents[i].currentRow == agents[j].currentRow {
					switch agents[i].selfCurrentDirection {
					case North:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow + 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case East:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol - 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case West:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol + 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case South:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow - 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					}
				}
				if agents[i].currentCol-1 == agents[j].currentCol && agents[i].currentRow == agents[j].currentRow {
					switch agents[i].selfCurrentDirection {
					case North:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow - 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case East:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol + 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case West:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol - 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case South:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow + 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					}
				}

				if agents[i].currentRow+1 == agents[j].currentRow && agents[i].currentCol == agents[j].currentCol {
					switch agents[i].selfCurrentDirection {
					case North:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol - 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					case East:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow - 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case West:
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow + 1
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol

					case South:
						agents[i].selfOtherAgentCol = agents[i].selfCurrentCol + 1
						agents[i].selfOtherAgentRow = agents[i].selfCurrentRow

					}
				}
			}
		}

		if agents[i].selfCurrentDirection == North {
			agents[i].selfFrontSensor = agents[i].selfMap[agents[i].selfCurrentRow-1][agents[i].selfCurrentCol]
			agents[i].selfRearSensor = agents[i].selfMap[agents[i].selfCurrentRow+1][agents[i].selfCurrentCol]
			agents[i].selfRightSensor = agents[i].selfMap[agents[i].selfCurrentRow][agents[i].selfCurrentCol+1]
			agents[i].selfLeftSensor = agents[i].selfMap[agents[i].selfCurrentRow][agents[i].selfCurrentCol-1]
		} else if agents[i].selfCurrentDirection == East {
			agents[i].selfFrontSensor = agents[i].selfMap[agents[i].selfCurrentRow][agents[i].selfCurrentCol+1]
			agents[i].selfRearSensor = agents[i].selfMap[agents[i].selfCurrentRow][agents[i].selfCurrentCol-1]
			agents[i].selfRightSensor = agents[i].selfMap[agents[i].selfCurrentRow+1][agents[i].selfCurrentCol]
			agents[i].selfLeftSensor = agents[i].selfMap[agents[i].selfCurrentRow-1][agents[i].selfCurrentCol]

		} else if agents[i].selfCurrentDirection == West {
			agents[i].selfFrontSensor = agents[i].selfMap[agents[i].selfCurrentRow][agents[i].selfCurrentCol-1]
			agents[i].selfRearSensor = agents[i].selfMap[agents[i].selfCurrentRow][agents[i].selfCurrentCol+1]
			agents[i].selfRightSensor = agents[i].selfMap[agents[i].selfCurrentRow-1][agents[i].selfCurrentCol]
			agents[i].selfLeftSensor = agents[i].selfMap[agents[i].selfCurrentRow+1][agents[i].selfCurrentCol]
		} else {
			agents[i].selfFrontSensor = agents[i].selfMap[agents[i].selfCurrentRow+1][agents[i].selfCurrentCol]
			agents[i].selfRearSensor = agents[i].selfMap[agents[i].selfCurrentRow-1][agents[i].selfCurrentCol]
			agents[i].selfRightSensor = agents[i].selfMap[agents[i].selfCurrentRow][agents[i].selfCurrentCol-1]
			agents[i].selfLeftSensor = agents[i].selfMap[agents[i].selfCurrentRow][agents[i].selfCurrentCol+1]
		}

	}
	var agentsPair []*Agent
	for i := 0; i < len(agents); i++ {
		if agents[i].selfOtherAgentRow != -1 && agents[i].selfOtherAgentCol != -1 {
			agentsPair = append(agentsPair, agents[i])
		}
	}
	//TODO: BU satırdan dolayı 2 den fazla agent varsa exchange yapmıyor
	if len(agentsPair) == 2 {
		exchangeMapInfo(agentsPair)

	}
	updateSelfMap()

}

func plan() {

	for i := 0; i < len(agents); i++ {
		//şimdilik movement'ı random olarak seç
		// eğer sadece etrafları empty ise hareket et.
		//TODO:tek satıra alınabilir?
		frontProbability := uint(4)
		rightProbability := uint(3)
		leftProbability := uint(2)
		rearProbability := uint(1)
		//self check
		if agents[i].selfFrontSensor == R {
			frontProbability = 0
		}
		if agents[i].selfRightSensor == R {
			rightProbability = 0
		}
		if agents[i].selfLeftSensor == R {
			leftProbability = 0
		}
		if agents[i].selfRearSensor == R {
			rearProbability = 0
		}
		//real check
		//TODO:EXITS WILL BE REMOVED
		if agents[i].frontSensor == F ||agents[i].frontSensor == EXIT  {
			frontProbability = 0
		}
		if agents[i].rightSensor == F || agents[i].rightSensor == EXIT {
			rightProbability = 0
		}
		if agents[i].leftSensor == F || agents[i].leftSensor == EXIT {
			leftProbability = 0
		}
		if agents[i].rearSensor == F || agents[i].rearSensor == EXIT {
			rearProbability = 0
		}
		rand.Seed(time.Now().UTC().UnixNano())
		c := wr.NewChooser(
			wr.Choice{Item: "front", Weight: frontProbability},
			wr.Choice{Item: "right", Weight: rightProbability},
			wr.Choice{Item: "left", Weight: leftProbability},
			wr.Choice{Item: "rear", Weight: rearProbability},
		)
		result := c.Pick().(string)

		if result == "front" {
			if agents[i].currentDirection == North {
				agents[i].nextDirection = North

			} else if agents[i].currentDirection == East {

				agents[i].nextDirection = East

			} else if agents[i].currentDirection == South {

				agents[i].nextDirection = South

			} else {
				agents[i].nextDirection = West
			}
			selfUpdateNextRotation(agents[i], "no_rotation_change")

		} else if result == "right" {

			if agents[i].currentDirection == North {
				agents[i].nextDirection = East

			} else if agents[i].currentDirection == East {
				agents[i].nextDirection = South

			} else if agents[i].currentDirection == South {
				agents[i].nextDirection = West

			} else {
				agents[i].nextDirection = North
			}
			selfUpdateNextRotation(agents[i], "rotate_right")

		} else if result == "left" {

			if agents[i].currentDirection == North {
				agents[i].nextDirection = West

			} else if agents[i].currentDirection == East {
				agents[i].nextDirection = North

			} else if agents[i].currentDirection == South {
				agents[i].nextDirection = East
			} else {
				agents[i].nextDirection = South
			}
			selfUpdateNextRotation(agents[i], "rotate_left")

		} else if result == "rear" {

			if agents[i].currentDirection == North {
				agents[i].nextDirection = South

			} else if agents[i].currentDirection == East {
				agents[i].nextDirection = West

			} else if agents[i].currentDirection == South {
				agents[i].nextDirection = North

			} else {
				agents[i].nextDirection = East
			}
			selfUpdateNextRotation(agents[i], "rotate_backward")

		}
	}
}

//bu fonk. ve nextDirection olmasa da olur,sadece rotating'i draw etmek için varlar.
func rotate() {

	for i := 0; i < len(agents); i++ {
		agents[i].currentDirection = agents[i].nextDirection
	}
}

func move() {
	for i := 0; i < len(agents); i++ {
		switch agents[i].currentDirection {
		case North:
			agents[i].currentRow -= 1
		case East:
			agents[i].currentCol += 1
		case South:
			agents[i].currentRow += 1
		case West:
			agents[i].currentCol -= 1

		}
	}
}

func drawMap(win *pixelgl.Window) {
	for row := 0; row < MapSize; row++ {
		for col := 0; col < MapSize; col++ {
			imd := imdraw.New(nil)
			imd.Color = tileColour[tileMap[row][col]]
			imd.Push(pixel.V(float64(col*TileSize), float64((MapSize-row-1)*TileSize)))
			imd.Push(pixel.V(float64(col*TileSize+TileSize), float64((MapSize-row-1)*TileSize+TileSize)))
			imd.Rectangle(0)
			imd.Draw(win)

		}
	}
}

func drawAgents(win *pixelgl.Window) {
	for i := 0; i < len(agents); i++ {
		agentCurrentCol, agentCurrentRow, agentCurrentDirection := float64(agents[i].currentCol), float64(agents[i].currentRow), agents[i].currentDirection
		imd := imdraw.New(nil)

		switch agentCurrentDirection {
		case North:
			imd.Color = agents[i].color
			imd.Push(pixel.V(agentCurrentCol*TileSize, (MapSize-agentCurrentRow-1)*TileSize))
			imd.Color = agents[i].color
			imd.Push(pixel.V(agentCurrentCol*TileSize+TileSize, (MapSize-agentCurrentRow-1)*TileSize))
			imd.Color = agents[i].color
			imd.Push(pixel.V((agentCurrentCol*TileSize)+(TileSize/2), (MapSize-agentCurrentRow-1)*TileSize+TileSize))
		case East:
			imd.Color = agents[i].color
			imd.Push(pixel.V(agentCurrentCol*TileSize, (MapSize-agentCurrentRow-1)*TileSize+TileSize))
			imd.Color = agents[i].color
			imd.Push(pixel.V(agentCurrentCol*TileSize, (MapSize-agentCurrentRow-1)*TileSize))
			imd.Color = agents[i].color
			imd.Push(pixel.V((agentCurrentCol*TileSize)+(TileSize), (MapSize-agentCurrentRow-1)*TileSize+(TileSize/2)))
		case South:
			imd.Color = agents[i].color
			imd.Push(pixel.V(agentCurrentCol*TileSize+TileSize, (MapSize-agentCurrentRow-1)*TileSize+TileSize))
			imd.Color = agents[i].color
			imd.Push(pixel.V(agentCurrentCol*TileSize, (MapSize-agentCurrentRow-1)*TileSize+TileSize))
			imd.Color = agents[i].color
			imd.Push(pixel.V((agentCurrentCol*TileSize)+(TileSize/2), (MapSize-agentCurrentRow-1)*TileSize))
		case West:
			imd.Color = agents[i].color
			imd.Push(pixel.V(agentCurrentCol*TileSize+TileSize, (MapSize-agentCurrentRow-1)*TileSize))
			imd.Color = agents[i].color
			imd.Push(pixel.V(agentCurrentCol*TileSize+TileSize, (MapSize-agentCurrentRow-1)*TileSize+TileSize))
			imd.Color = agents[i].color
			imd.Push(pixel.V(agentCurrentCol*TileSize, (MapSize-agentCurrentRow-1)*TileSize+(TileSize/2)))
		}

		imd.Polygon(0)
		imd.Draw(win)

	}
}

func selfUpdateNextRotation(agent *Agent, rotation string) {
	if agent.selfCurrentDirection == North {
		switch rotation {
		case "no_rotation_change":
			agent.selfNextDirection = North
		case "rotate_right":
			agent.selfNextDirection = East
		case "rotate_left":
			agent.selfNextDirection = West
		case "rotate_backward":
			agent.selfNextDirection = South

		}
	} else if agent.selfCurrentDirection == East {
		switch rotation {
		case "no_rotation_change":
			agent.selfNextDirection = East
		case "rotate_right":
			agent.selfNextDirection = South
		case "rotate_left":
			agent.selfNextDirection = North
		case "rotate_backward":
			agent.selfNextDirection = West

		}
	} else if agent.selfCurrentDirection == West {
		switch rotation {
		case "no_rotation_change":
			agent.selfNextDirection = West
		case "rotate_right":
			agent.selfNextDirection = North
		case "rotate_left":
			agent.selfNextDirection = South
		case "rotate_backward":
			agent.selfNextDirection = East

		}
	} else {
		switch rotation {
		case "no_rotation_change":
			agent.selfNextDirection = South
		case "rotate_right":
			agent.selfNextDirection = West
		case "rotate_left":
			agent.selfNextDirection = East
		case "rotate_backward":
			agent.selfNextDirection = North

		}
	}
}
func selfRotate(agent *Agent) {
	agent.selfCurrentDirection = agent.selfNextDirection

}

func selfMove(agent *Agent) {

	switch agent.selfCurrentDirection {
	case North:
		agent.selfCurrentRow -= 1
	case East:
		agent.selfCurrentCol += 1
	case South:
		agent.selfCurrentRow += 1
	case West:
		agent.selfCurrentCol -= 1

	}

	loopCheck(agent)
	agent.selfPath = append(agent.selfPath, []int{agent.selfCurrentRow, agent.selfCurrentCol})

}

func loopCheck(agent *Agent) {
	code := `
import numpy as np

visited = %s

offset = []
distinctedOffset = []
for i in visited:
    coordsWithSameRow = []
    coordsWithSameCol = []
    for j in visited:
        if i[0] == j[0]:
            coordsWithSameRow.append(j)
        elif i[1] == j[1]:
            coordsWithSameCol.append(j)
    for item in coordsWithSameCol:
        rang = range(min(i[0], item[0])+1, max(i[0], item[0]))
        for n in rang:
            offset.append([n, i[1]])

    for item in coordsWithSameRow:
        rang = range(min(i[1], item[1])+1, max(i[1], item[1]))
        for n in rang:
            offset.append([i[0], n])

distinctedOffset = [list(x) for x in set(tuple(x) for x in offset)]
intersection = [x for x in visited if x in distinctedOffset]
result = [x for x in distinctedOffset if x not in intersection]

for el in reversed(result):
    els_with_same_col = []
    els_with_same_row = []
    for point in visited:
        if point[1] == el[1]:
            els_with_same_col.append(point)
        if point[0] == el[0]:
            els_with_same_row.append(point)
    if len(list(filter(lambda p: p[0] < el[0], els_with_same_col))) < 1 or len(list(filter(lambda p: p[0] > el[0], els_with_same_col))) < 1:
        result = []
    if len(list(filter(lambda p: p[1] < el[1], els_with_same_row))) < 1 or len(list(filter(lambda p: p[1] > el[1], els_with_same_row))) < 1:
        result = []

print(result)
`

	for i := len(agent.selfPath) - 1; i >= 0; i-- {
		if agent.selfPath[i][0] == agent.selfCurrentRow && agent.selfPath[i][1] == agent.selfCurrentCol {
			println("burda")
			arrList := agent.selfPath[i:len(agent.selfPath)]
			//firstEl := arrList[0]
			//arrList = arrList[1:]

		Loop:
			//if len(arrList) > 2 {
			for x := 1; x < len(arrList); x++ {
				for y := len(arrList) - 1; y >= 1; y-- {
					if x != y {

						if arrList[x][0] == arrList[y][0] && arrList[x][1] == arrList[y][1] {
							firstPart := arrList[0:x]
							secondPart := arrList[y:len(arrList)]
							arrList = append(firstPart, secondPart...)
							goto Loop

						}
					}

				}
			}
			//}

			list := "["
			for _, val := range arrList {
				list += fmt.Sprintf("[%d,%d],", val[0], val[1])
			}
			list += "]"
			cmd := exec.Command("python", "-c", fmt.Sprintf(code, list))

			out, _ := cmd.Output()
			str := strings.TrimSpace(string(out))
			str = strings.Replace(str, " ", "", -1)
			str = str[1 : len(str)-1]
			if len(str) != 0 {
				pointsInLoop := pythonListToSlice(str)
				for _, el := range pointsInLoop {
					//if agent.selfMap[el[0]][el[1]] != R {
					agent.selfMap[el[0]][el[1]] = R

					//}
				}
				//pythonListToSlice("[1,2],[3,4],[5,6],")

			}

			return

		}
	}

}
func pythonListToSlice(out string) [][]int {
	var slice [][]int
	println(out)
	re := regexp.MustCompile(`\[([^\[\]]*)\]`)
	subMatchAll := re.FindAllString(out, -1)
	for _, element := range subMatchAll {
		element = strings.Trim(element, "[")
		element = strings.Trim(element, "]")
		coordsInStr := strings.Split(element, ",")
		rowInt, _ := strconv.Atoi(coordsInStr[0])
		colInt, _ := strconv.Atoi(coordsInStr[1])
		slice = append(slice, []int{rowInt, colInt})
		fmt.Println(element)
	}
	return slice

}

//agentin kendi haritasındaki directionuna bağlı olarak,
//gerçek haritada sense ettiği veriyi kendi haritasına doğru şekilde işlemek
func updateSelfMap() {
	for i := 0; i < len(agents); i++ {

		upperRow := agents[i].selfMap[agents[i].selfCurrentRow-1][agents[i].selfCurrentCol]
		lowerRow := agents[i].selfMap[agents[i].selfCurrentRow+1][agents[i].selfCurrentCol]
		rightCol := agents[i].selfMap[agents[i].selfCurrentRow][agents[i].selfCurrentCol+1]
		leftCol := agents[i].selfMap[agents[i].selfCurrentRow][agents[i].selfCurrentCol-1]

		if agents[i].selfCurrentDirection == North {
			//normal haritada frontSensorde gördüğünü, kendi haritasında, önüne işaretle
			agents[i].selfMap[agents[i].selfCurrentRow-1][agents[i].selfCurrentCol] = agents[i].frontSensor
			//normal haritada rearSensorde gördüğünü, kendi haritasında, arkana işaretle
			agents[i].selfMap[agents[i].selfCurrentRow+1][agents[i].selfCurrentCol] = agents[i].rearSensor
			agents[i].selfMap[agents[i].selfCurrentRow][agents[i].selfCurrentCol+1] = agents[i].rightSensor
			agents[i].selfMap[agents[i].selfCurrentRow][agents[i].selfCurrentCol-1] = agents[i].leftSensor
		} else if agents[i].selfCurrentDirection == East {
			agents[i].selfMap[agents[i].selfCurrentRow+1][agents[i].selfCurrentCol] = agents[i].rightSensor
			agents[i].selfMap[agents[i].selfCurrentRow-1][agents[i].selfCurrentCol] = agents[i].leftSensor
			agents[i].selfMap[agents[i].selfCurrentRow][agents[i].selfCurrentCol-1] = agents[i].rearSensor
			agents[i].selfMap[agents[i].selfCurrentRow][agents[i].selfCurrentCol+1] = agents[i].frontSensor
		} else if agents[i].selfCurrentDirection == West {
			agents[i].selfMap[agents[i].selfCurrentRow+1][agents[i].selfCurrentCol] = agents[i].leftSensor
			agents[i].selfMap[agents[i].selfCurrentRow-1][agents[i].selfCurrentCol] = agents[i].rightSensor
			agents[i].selfMap[agents[i].selfCurrentRow][agents[i].selfCurrentCol-1] = agents[i].frontSensor
			agents[i].selfMap[agents[i].selfCurrentRow][agents[i].selfCurrentCol+1] = agents[i].rearSensor
		} else {
			agents[i].selfMap[agents[i].selfCurrentRow-1][agents[i].selfCurrentCol] = agents[i].rearSensor
			agents[i].selfMap[agents[i].selfCurrentRow+1][agents[i].selfCurrentCol] = agents[i].frontSensor
			agents[i].selfMap[agents[i].selfCurrentRow][agents[i].selfCurrentCol+1] = agents[i].leftSensor
			agents[i].selfMap[agents[i].selfCurrentRow][agents[i].selfCurrentCol-1] = agents[i].rightSensor
		}

		if upperRow == R {
			agents[i].selfMap[agents[i].selfCurrentRow-1][agents[i].selfCurrentCol] = R
		}
		if lowerRow == R {
			agents[i].selfMap[agents[i].selfCurrentRow+1][agents[i].selfCurrentCol] = R
		}
		if rightCol == R {
			agents[i].selfMap[agents[i].selfCurrentRow][agents[i].selfCurrentCol+1] = R
		}
		if leftCol == R {
			agents[i].selfMap[agents[i].selfCurrentRow][agents[i].selfCurrentCol-1] = R
		}
	}

}

func drawAgentSelfMap(win *pixelgl.Window, agent *Agent) {

	for row := 0; row < MapSize*2+1; row++ {
		for col := 0; col < MapSize*2+1; col++ {
			imd := imdraw.New(nil)
			imd.Color = tileColour[agent.selfMap[row][col]]
			imd.Push(pixel.V(float64(col*SelfTileSize), float64((MapSize*2-row-1)*SelfTileSize)))
			imd.Push(pixel.V(float64(col*SelfTileSize+SelfTileSize), float64((MapSize*2-row-1)*SelfTileSize+SelfTileSize)))
			imd.Rectangle(0)
			imd.Draw(win)

		}
	}

}
func drawAgentToSelfMap(win *pixelgl.Window, agent *Agent) {
	agentCurrentCol, agentCurrentRow, agentCurrentDirection := float64(agent.selfCurrentCol), float64(agent.selfCurrentRow), agent.selfCurrentDirection

	switch agentCurrentDirection {
	case North:
		imd := imdraw.New(nil)

		imd.Color = agent.color
		imd.Push(pixel.V(agentCurrentCol*SelfTileSize, (MapSize*2-agentCurrentRow-1)*SelfTileSize))
		imd.Color = agent.color
		imd.Push(pixel.V(agentCurrentCol*SelfTileSize+SelfTileSize, (MapSize*2-agentCurrentRow-1)*SelfTileSize))
		imd.Color = agent.color
		imd.Push(pixel.V((agentCurrentCol*SelfTileSize)+(SelfTileSize/2), (MapSize*2-agentCurrentRow-1)*SelfTileSize+SelfTileSize))
		imd.Polygon(0)
		imd.Draw(win)

	case East:
		imd := imdraw.New(nil)

		imd.Color = agent.color
		imd.Push(pixel.V(agentCurrentCol*SelfTileSize, (MapSize*2-agentCurrentRow-1)*SelfTileSize+SelfTileSize))
		imd.Color = agent.color
		imd.Push(pixel.V(agentCurrentCol*SelfTileSize, (MapSize*2-agentCurrentRow-1)*SelfTileSize))
		imd.Color = agent.color
		imd.Push(pixel.V((agentCurrentCol*SelfTileSize)+(SelfTileSize), (MapSize*2-agentCurrentRow-1)*SelfTileSize+(SelfTileSize/2)))
		imd.Polygon(0)
		imd.Draw(win)

	case South:
		imd := imdraw.New(nil)

		imd.Color = agent.color
		imd.Push(pixel.V(agentCurrentCol*SelfTileSize+SelfTileSize, (MapSize*2-agentCurrentRow-1)*SelfTileSize+SelfTileSize))
		imd.Color = agent.color
		imd.Push(pixel.V(agentCurrentCol*SelfTileSize, (MapSize*2-agentCurrentRow-1)*SelfTileSize+SelfTileSize))
		imd.Color = agent.color
		imd.Push(pixel.V((agentCurrentCol*SelfTileSize)+(SelfTileSize/2), (MapSize*2-agentCurrentRow-1)*SelfTileSize))
		imd.Polygon(0)
		imd.Draw(win)

	case West:
		imd := imdraw.New(nil)

		imd.Color = agent.color
		imd.Push(pixel.V(agentCurrentCol*SelfTileSize+SelfTileSize, (MapSize*2-agentCurrentRow-1)*SelfTileSize))
		imd.Color = agent.color
		imd.Push(pixel.V(agentCurrentCol*SelfTileSize+SelfTileSize, (MapSize*2-agentCurrentRow-1)*SelfTileSize+SelfTileSize))
		imd.Color = agent.color
		imd.Push(pixel.V(agentCurrentCol*SelfTileSize, (MapSize*2-agentCurrentRow-1)*SelfTileSize+(SelfTileSize/2)))
		imd.Polygon(0)
		imd.Draw(win)

	}
}
func rotate90DegreeClockwise(arr [AgentMapSize][AgentMapSize]int, selfCurrentRow int, selfCurrentCol int) ([AgentMapSize][AgentMapSize]int, int, int) {

	newArr := [AgentMapSize][AgentMapSize]int{}

	for row := 0; row < AgentMapSize; row++ {
		for col := 0; col < AgentMapSize; col++ {
			newArr[col][AgentMapSize-1-row] = arr[row][col]
		}
	}

	return newArr, selfCurrentCol, (AgentMapSize - 1) - selfCurrentRow
}

// TODO: will be refactored
func exchangeMapInfo(agentsPair []*Agent) () {

	firstAgent := agentsPair[0]
	secondAgent := agentsPair[1]
	//1st pos

	if firstAgent.selfOtherAgentRow == firstAgent.selfCurrentRow-1 &&
		firstAgent.selfOtherAgentCol == firstAgent.selfCurrentCol &&
		secondAgent.selfOtherAgentRow == secondAgent.selfCurrentRow-1 &&
		secondAgent.selfOtherAgentCol == secondAgent.selfCurrentCol {

		secondAgentRotatedSelfMap, secondAgentRotatedSelfCurrentRow, secondAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(rotate90DegreeClockwise(secondAgent.selfMap, secondAgent.selfCurrentRow, secondAgent.selfCurrentCol))
		firstAgentTempSelfMap := firstAgent.selfMap

		rowOffset := firstAgent.selfOtherAgentRow - secondAgentRotatedSelfCurrentRow
		colOffset := firstAgent.selfOtherAgentCol - secondAgentRotatedSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if secondAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					firstAgentTempSelfMap[row+rowOffset][col+colOffset] = secondAgentRotatedSelfMap[row][col]
				}

			}
		}

		firstAgentRotatedSelfMap, firstAgentRotatedSelfCurrentRow, firstAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(rotate90DegreeClockwise(firstAgent.selfMap, firstAgent.selfCurrentRow, firstAgent.selfCurrentCol))
		secondAgentTempSelfMap := secondAgent.selfMap

		rowOffset = secondAgent.selfOtherAgentRow - firstAgentRotatedSelfCurrentRow
		colOffset = secondAgent.selfOtherAgentCol - firstAgentRotatedSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if firstAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					secondAgentTempSelfMap[row+rowOffset][col+colOffset] = firstAgentRotatedSelfMap[row][col]
				}

			}
		}
		firstAgent.selfMap = firstAgentTempSelfMap
		secondAgent.selfMap = secondAgentTempSelfMap
	}
	//2nd
	if firstAgent.selfOtherAgentRow == firstAgent.selfCurrentRow &&
		firstAgent.selfOtherAgentCol == firstAgent.selfCurrentCol+1 &&
		secondAgent.selfCurrentRow == secondAgent.selfCurrentRow-1 &&
		secondAgent.selfCurrentCol == secondAgent.selfCurrentCol {

		secondAgentRotatedSelfMap, secondAgentRotatedSelfCurrentRow, secondAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(rotate90DegreeClockwise(rotate90DegreeClockwise(secondAgent.selfMap, secondAgent.selfCurrentRow, secondAgent.selfCurrentCol)))
		firstAgentTempSelfMap := firstAgent.selfMap

		rowOffset := firstAgent.selfOtherAgentRow - secondAgentRotatedSelfCurrentRow
		colOffset := firstAgent.selfOtherAgentCol - secondAgentRotatedSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if secondAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					firstAgentTempSelfMap[row+rowOffset][col+colOffset] = secondAgentRotatedSelfMap[row][col]
				}

			}
		}

		firstAgentRotatedSelfMap, firstAgentRotatedSelfCurrentRow, firstAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(firstAgent.selfMap, firstAgent.selfCurrentRow, firstAgent.selfCurrentCol)
		secondAgentTempSelfMap := secondAgent.selfMap

		rowOffset = secondAgent.selfOtherAgentRow - firstAgentRotatedSelfCurrentRow
		colOffset = secondAgent.selfOtherAgentCol - firstAgentRotatedSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if firstAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					secondAgentTempSelfMap[row+rowOffset][col+colOffset] = firstAgentRotatedSelfMap[row][col]
				}

			}
		}
		firstAgent.selfMap = firstAgentTempSelfMap
		secondAgent.selfMap = secondAgentTempSelfMap

	}
	//3rd
	if firstAgent.selfOtherAgentRow == firstAgent.selfCurrentRow &&
		firstAgent.selfOtherAgentCol == firstAgent.selfCurrentCol-1 &&
		secondAgent.selfOtherAgentRow == secondAgent.selfCurrentRow-1 &&
		secondAgent.selfOtherAgentCol == secondAgent.selfCurrentCol {

		secondAgentRotatedSelfMap, secondAgentRotatedSelfCurrentRow, secondAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(secondAgent.selfMap, secondAgent.selfCurrentRow, secondAgent.selfCurrentCol)
		firstAgentTempSelfMap := firstAgent.selfMap

		rowOffset := firstAgent.selfOtherAgentRow - secondAgentRotatedSelfCurrentRow
		colOffset := firstAgent.selfOtherAgentCol - secondAgentRotatedSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if secondAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					firstAgentTempSelfMap[row+rowOffset][col+colOffset] = secondAgentRotatedSelfMap[row][col]
				}

			}
		}

		firstAgentRotatedSelfMap, firstAgentRotatedSelfCurrentRow, firstAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(rotate90DegreeClockwise(rotate90DegreeClockwise(firstAgent.selfMap, firstAgent.selfCurrentRow, firstAgent.selfCurrentCol)))
		secondAgentTempSelfMap := secondAgent.selfMap

		rowOffset = secondAgent.selfOtherAgentRow - firstAgentRotatedSelfCurrentRow
		colOffset = secondAgent.selfOtherAgentCol - firstAgentRotatedSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if firstAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					secondAgentTempSelfMap[row+rowOffset][col+colOffset] = firstAgentRotatedSelfMap[row][col]
				}

			}
		}
		firstAgent.selfMap = firstAgentTempSelfMap
		secondAgent.selfMap = secondAgentTempSelfMap

	}
	//4rd
	if firstAgent.selfOtherAgentRow == firstAgent.selfCurrentRow+1 &&
		firstAgent.selfOtherAgentCol == secondAgent.selfCurrentCol &&
		secondAgent.selfOtherAgentRow == secondAgent.selfCurrentRow-1 &&
		secondAgent.selfOtherAgentCol == secondAgent.selfCurrentCol {

		secondAgentSelfMap := secondAgent.selfMap
		secondAgentSelfCurrentRow := secondAgent.selfCurrentRow
		secondAgentSelfCurrentCol := secondAgent.selfCurrentCol

		firstAgentTempSelfMap := firstAgent.selfMap

		rowOffset := firstAgent.selfOtherAgentRow - secondAgentSelfCurrentRow
		colOffset := firstAgent.selfOtherAgentCol - secondAgentSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if secondAgentSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					firstAgentTempSelfMap[row+rowOffset][col+colOffset] = secondAgentSelfMap[row][col]
				}

			}
		}

		firstAgentSelfMap := firstAgent.selfMap
		firstAgentSelfCurrentRow := firstAgent.selfCurrentRow
		firstAgentSelfCurrentCol := firstAgent.selfCurrentCol

		secondAgentTempSelfMap := secondAgent.selfMap

		rowOffset = secondAgent.selfOtherAgentRow - firstAgentSelfCurrentRow
		colOffset = secondAgent.selfOtherAgentCol - firstAgentSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if firstAgentSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					secondAgentTempSelfMap[row+rowOffset][col+colOffset] = firstAgentSelfMap[row][col]
				}

			}
		}

		firstAgent.selfMap = firstAgentTempSelfMap
		secondAgent.selfMap = secondAgentTempSelfMap
	}
	//5th
	if firstAgent.selfOtherAgentRow == firstAgent.selfCurrentRow-1 &&
		firstAgent.selfOtherAgentCol == secondAgent.selfCurrentCol &&
		secondAgent.selfOtherAgentRow == secondAgent.selfCurrentRow &&
		secondAgent.selfOtherAgentCol == secondAgent.selfCurrentCol+1 {

		secondAgentRotatedSelfMap, secondAgentRotatedSelfCurrentRow, secondAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(secondAgent.selfMap, secondAgent.selfCurrentRow, secondAgent.selfCurrentCol)
		firstAgentTempSelfMap := firstAgent.selfMap

		rowOffset := firstAgent.selfOtherAgentRow - secondAgentRotatedSelfCurrentRow
		colOffset := firstAgent.selfOtherAgentCol - secondAgentRotatedSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if secondAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					firstAgentTempSelfMap[row+rowOffset][col+colOffset] = secondAgentRotatedSelfMap[row][col]
				}

			}
		}

		firstAgentRotatedSelfMap, firstAgentRotatedSelfCurrentRow, firstAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(rotate90DegreeClockwise(rotate90DegreeClockwise(firstAgent.selfMap, firstAgent.selfCurrentRow, firstAgent.selfCurrentCol)))
		secondAgentTempSelfMap := secondAgent.selfMap

		rowOffset = secondAgent.selfOtherAgentRow - firstAgentRotatedSelfCurrentRow
		colOffset = secondAgent.selfOtherAgentCol - firstAgentRotatedSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if firstAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					secondAgentTempSelfMap[row+rowOffset][col+colOffset] = firstAgentRotatedSelfMap[row][col]
				}

			}
		}
		firstAgent.selfMap = firstAgentTempSelfMap
		secondAgent.selfMap = secondAgentTempSelfMap

	}
	//6th
	if firstAgent.selfOtherAgentRow == firstAgent.selfCurrentRow &&
		firstAgent.selfOtherAgentCol == firstAgent.selfCurrentCol+1 &&
		secondAgent.selfOtherAgentRow == secondAgent.selfCurrentRow &&
		secondAgent.selfOtherAgentCol == secondAgent.selfCurrentCol+1 {

		secondAgentRotatedSelfMap, secondAgentRotatedSelfCurrentRow, secondAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(rotate90DegreeClockwise(secondAgent.selfMap, secondAgent.selfCurrentRow, secondAgent.selfCurrentCol))
		firstAgentTempSelfMap := firstAgent.selfMap

		rowOffset := firstAgent.selfOtherAgentRow - secondAgentRotatedSelfCurrentRow
		colOffset := firstAgent.selfOtherAgentCol - secondAgentRotatedSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if secondAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					firstAgentTempSelfMap[row+rowOffset][col+colOffset] = secondAgentRotatedSelfMap[row][col]
				}

			}
		}

		firstAgentRotatedSelfMap, firstAgentRotatedSelfCurrentRow, firstAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(rotate90DegreeClockwise(firstAgent.selfMap, firstAgent.selfCurrentRow, firstAgent.selfCurrentCol))
		secondAgentTempSelfMap := secondAgent.selfMap

		rowOffset = secondAgent.selfOtherAgentRow - firstAgentRotatedSelfCurrentRow
		colOffset = secondAgent.selfOtherAgentCol - firstAgentRotatedSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if firstAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					secondAgentTempSelfMap[row+rowOffset][col+colOffset] = firstAgentRotatedSelfMap[row][col]
				}

			}
		}
		firstAgent.selfMap = firstAgentTempSelfMap
		secondAgent.selfMap = secondAgentTempSelfMap

	}
	//7th
	if firstAgent.selfOtherAgentRow == firstAgent.selfCurrentRow &&
		firstAgent.selfOtherAgentCol == firstAgent.selfCurrentCol-1 &&
		secondAgent.selfOtherAgentRow == secondAgent.selfCurrentRow &&
		secondAgent.selfOtherAgentCol == secondAgent.selfCurrentCol+1 {

		secondAgentSelfMap := secondAgent.selfMap
		secondAgentSelfCurrentRow := secondAgent.selfCurrentRow
		secondAgentSelfCurrentCol := secondAgent.selfCurrentCol

		firstAgentTempSelfMap := firstAgent.selfMap

		rowOffset := firstAgent.selfOtherAgentRow - secondAgentSelfCurrentRow
		colOffset := firstAgent.selfOtherAgentCol - secondAgentSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if secondAgentSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					firstAgentTempSelfMap[row+rowOffset][col+colOffset] = secondAgentSelfMap[row][col]
				}

			}
		}

		firstAgentSelfMap := firstAgent.selfMap
		firstAgentSelfCurrentRow := firstAgent.selfCurrentRow
		firstAgentSelfCurrentCol := firstAgent.selfCurrentCol

		secondAgentTempSelfMap := secondAgent.selfMap

		rowOffset = secondAgent.selfOtherAgentRow - firstAgentSelfCurrentRow
		colOffset = secondAgent.selfOtherAgentCol - firstAgentSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if firstAgentSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					secondAgentTempSelfMap[row+rowOffset][col+colOffset] = firstAgentSelfMap[row][col]
				}

			}
		}

		firstAgent.selfMap = firstAgentTempSelfMap
		secondAgent.selfMap = secondAgentTempSelfMap

		//8th
		if firstAgent.selfOtherAgentRow == firstAgent.selfCurrentRow+1 &&
			firstAgent.selfOtherAgentCol == firstAgent.selfCurrentCol &&
			secondAgent.selfOtherAgentRow == secondAgent.selfCurrentRow &&
			secondAgent.selfOtherAgentCol == secondAgent.selfCurrentCol+1 {

			secondAgentRotatedSelfMap, secondAgentRotatedSelfCurrentRow, secondAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(rotate90DegreeClockwise(rotate90DegreeClockwise(secondAgent.selfMap, secondAgent.selfCurrentRow, secondAgent.selfCurrentCol)))
			firstAgentTempSelfMap := firstAgent.selfMap

			rowOffset := firstAgent.selfOtherAgentRow - secondAgentRotatedSelfCurrentRow
			colOffset := firstAgent.selfOtherAgentCol - secondAgentRotatedSelfCurrentCol

			for row := 0; row < AgentMapSize; row++ {
				for col := 0; col < AgentMapSize; col++ {
					if secondAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
						firstAgentTempSelfMap[row+rowOffset][col+colOffset] = secondAgentRotatedSelfMap[row][col]
					}

				}
			}

			firstAgentRotatedSelfMap, firstAgentRotatedSelfCurrentRow, firstAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(firstAgent.selfMap, firstAgent.selfCurrentRow, firstAgent.selfCurrentCol)
			secondAgentTempSelfMap := secondAgent.selfMap

			rowOffset = secondAgent.selfOtherAgentRow - firstAgentRotatedSelfCurrentRow
			colOffset = secondAgent.selfOtherAgentCol - firstAgentRotatedSelfCurrentCol

			for row := 0; row < AgentMapSize; row++ {
				for col := 0; col < AgentMapSize; col++ {
					if firstAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
						secondAgentTempSelfMap[row+rowOffset][col+colOffset] = firstAgentRotatedSelfMap[row][col]
					}

				}
			}
			firstAgent.selfMap = firstAgentTempSelfMap
			secondAgent.selfMap = secondAgentTempSelfMap

		}
		//9th
		if firstAgent.selfOtherAgentRow == firstAgent.selfCurrentRow-1 &&
			firstAgent.selfOtherAgentCol == firstAgent.selfCurrentCol &&
			secondAgent.selfOtherAgentRow == secondAgent.selfCurrentRow &&
			secondAgent.selfOtherAgentCol == secondAgent.selfCurrentCol-1 {

			secondAgentRotatedSelfMap, secondAgentRotatedSelfCurrentRow, secondAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(rotate90DegreeClockwise(rotate90DegreeClockwise(secondAgent.selfMap, secondAgent.selfCurrentRow, secondAgent.selfCurrentCol)))
			firstAgentTempSelfMap := firstAgent.selfMap

			rowOffset := firstAgent.selfOtherAgentRow - secondAgentRotatedSelfCurrentRow
			colOffset := firstAgent.selfOtherAgentCol - secondAgentRotatedSelfCurrentCol

			for row := 0; row < AgentMapSize; row++ {
				for col := 0; col < AgentMapSize; col++ {
					if secondAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
						firstAgentTempSelfMap[row+rowOffset][col+colOffset] = secondAgentRotatedSelfMap[row][col]
					}

				}
			}

			firstAgentRotatedSelfMap, firstAgentRotatedSelfCurrentRow, firstAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(firstAgent.selfMap, firstAgent.selfCurrentRow, firstAgent.selfCurrentCol)
			secondAgentTempSelfMap := secondAgent.selfMap

			rowOffset = secondAgent.selfOtherAgentRow - firstAgentRotatedSelfCurrentRow
			colOffset = secondAgent.selfOtherAgentCol - firstAgentRotatedSelfCurrentCol

			for row := 0; row < AgentMapSize; row++ {
				for col := 0; col < AgentMapSize; col++ {
					if firstAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
						secondAgentTempSelfMap[row+rowOffset][col+colOffset] = firstAgentRotatedSelfMap[row][col]
					}

				}
			}
			firstAgent.selfMap = firstAgentTempSelfMap
			secondAgent.selfMap = secondAgentTempSelfMap
		}

	}
	//10th
	if firstAgent.selfOtherAgentRow == firstAgent.selfCurrentRow &&
		firstAgent.selfOtherAgentCol == firstAgent.selfCurrentCol+1 &&
		secondAgent.selfOtherAgentRow == secondAgent.selfCurrentRow &&
		secondAgent.selfOtherAgentCol == secondAgent.selfCurrentCol-1 {

		secondAgentSelfMap := secondAgent.selfMap
		secondAgentSelfCurrentRow := secondAgent.selfCurrentRow
		secondAgentSelfCurrentCol := secondAgent.selfCurrentCol

		firstAgentTempSelfMap := firstAgent.selfMap

		rowOffset := firstAgent.selfOtherAgentRow - secondAgentSelfCurrentRow
		colOffset := firstAgent.selfOtherAgentCol - secondAgentSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if secondAgentSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					firstAgentTempSelfMap[row+rowOffset][col+colOffset] = secondAgentSelfMap[row][col]
				}

			}
		}

		firstAgentSelfMap := firstAgent.selfMap
		firstAgentSelfCurrentRow := firstAgent.selfCurrentRow
		firstAgentSelfCurrentCol := firstAgent.selfCurrentCol

		secondAgentTempSelfMap := secondAgent.selfMap

		rowOffset = secondAgent.selfOtherAgentRow - firstAgentSelfCurrentRow
		colOffset = secondAgent.selfOtherAgentCol - firstAgentSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if firstAgentSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					secondAgentTempSelfMap[row+rowOffset][col+colOffset] = firstAgentSelfMap[row][col]
				}

			}
		}

		firstAgent.selfMap = firstAgentTempSelfMap
		secondAgent.selfMap = secondAgentTempSelfMap
	}
	//11th
	if firstAgent.selfOtherAgentRow == firstAgent.selfCurrentRow &&
		firstAgent.selfOtherAgentCol == firstAgent.selfCurrentCol-1 &&
		secondAgent.selfOtherAgentRow == secondAgent.selfCurrentRow &&
		secondAgent.selfOtherAgentCol == secondAgent.selfCurrentCol-1 {

		secondAgentRotatedSelfMap, secondAgentRotatedSelfCurrentRow, secondAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(rotate90DegreeClockwise(secondAgent.selfMap, secondAgent.selfCurrentRow, secondAgent.selfCurrentCol))
		firstAgentTempSelfMap := firstAgent.selfMap

		rowOffset := firstAgent.selfOtherAgentRow - secondAgentRotatedSelfCurrentRow
		colOffset := firstAgent.selfOtherAgentCol - secondAgentRotatedSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if secondAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					firstAgentTempSelfMap[row+rowOffset][col+colOffset] = secondAgentRotatedSelfMap[row][col]
				}

			}
		}

		firstAgentRotatedSelfMap, firstAgentRotatedSelfCurrentRow, firstAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(rotate90DegreeClockwise(firstAgent.selfMap, firstAgent.selfCurrentRow, firstAgent.selfCurrentCol))
		secondAgentTempSelfMap := secondAgent.selfMap

		rowOffset = secondAgent.selfOtherAgentRow - firstAgentRotatedSelfCurrentRow
		colOffset = secondAgent.selfOtherAgentCol - firstAgentRotatedSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if firstAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					secondAgentTempSelfMap[row+rowOffset][col+colOffset] = firstAgentRotatedSelfMap[row][col]
				}

			}
		}
		firstAgent.selfMap = firstAgentTempSelfMap
		secondAgent.selfMap = secondAgentTempSelfMap
	}
	//12th
	if firstAgent.selfOtherAgentRow == firstAgent.selfCurrentRow+1 &&
		firstAgent.selfOtherAgentCol == firstAgent.selfCurrentCol &&
		secondAgent.selfOtherAgentRow == secondAgent.selfCurrentRow &&
		secondAgent.selfOtherAgentCol == secondAgent.selfCurrentCol-1 {

		secondAgentRotatedSelfMap, secondAgentRotatedSelfCurrentRow, secondAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(secondAgent.selfMap, secondAgent.selfCurrentRow, secondAgent.selfCurrentCol)
		firstAgentTempSelfMap := firstAgent.selfMap

		rowOffset := firstAgent.selfOtherAgentRow - secondAgentRotatedSelfCurrentRow
		colOffset := firstAgent.selfOtherAgentCol - secondAgentRotatedSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if secondAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					firstAgentTempSelfMap[row+rowOffset][col+colOffset] = secondAgentRotatedSelfMap[row][col]
				}

			}
		}

		firstAgentRotatedSelfMap, firstAgentRotatedSelfCurrentRow, firstAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(rotate90DegreeClockwise(rotate90DegreeClockwise(firstAgent.selfMap, firstAgent.selfCurrentRow, firstAgent.selfCurrentCol)))
		secondAgentTempSelfMap := secondAgent.selfMap

		rowOffset = secondAgent.selfOtherAgentRow - firstAgentRotatedSelfCurrentRow
		colOffset = secondAgent.selfOtherAgentCol - firstAgentRotatedSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if firstAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					secondAgentTempSelfMap[row+rowOffset][col+colOffset] = firstAgentRotatedSelfMap[row][col]
				}

			}
		}
		firstAgent.selfMap = firstAgentTempSelfMap
		secondAgent.selfMap = secondAgentTempSelfMap

	}
	//13th
	if firstAgent.selfOtherAgentRow == firstAgent.selfCurrentRow-1 &&
		firstAgent.selfOtherAgentCol == firstAgent.selfCurrentCol &&
		secondAgent.selfOtherAgentRow == secondAgent.selfCurrentRow+1 &&
		secondAgent.selfOtherAgentCol == secondAgent.selfCurrentCol {

		secondAgentSelfMap := secondAgent.selfMap
		secondAgentSelfCurrentRow := secondAgent.selfCurrentRow
		secondAgentSelfCurrentCol := secondAgent.selfCurrentCol

		firstAgentTempSelfMap := firstAgent.selfMap

		rowOffset := firstAgent.selfOtherAgentRow - secondAgentSelfCurrentRow
		colOffset := firstAgent.selfOtherAgentCol - secondAgentSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if secondAgentSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					firstAgentTempSelfMap[row+rowOffset][col+colOffset] = secondAgentSelfMap[row][col]
				}

			}
		}

		firstAgentSelfMap := firstAgent.selfMap
		firstAgentSelfCurrentRow := firstAgent.selfCurrentRow
		firstAgentSelfCurrentCol := firstAgent.selfCurrentCol

		secondAgentTempSelfMap := secondAgent.selfMap

		rowOffset = secondAgent.selfOtherAgentRow - firstAgentSelfCurrentRow
		colOffset = secondAgent.selfOtherAgentCol - firstAgentSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if firstAgentSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					secondAgentTempSelfMap[row+rowOffset][col+colOffset] = firstAgentSelfMap[row][col]
				}

			}
		}

		firstAgent.selfMap = firstAgentTempSelfMap
		secondAgent.selfMap = secondAgentTempSelfMap

	}
	//14th
	if firstAgent.selfOtherAgentRow == firstAgent.selfCurrentRow &&
		firstAgent.selfOtherAgentCol == firstAgent.selfCurrentCol+1 &&
		secondAgent.selfOtherAgentRow == secondAgent.selfCurrentRow+1 &&
		secondAgent.selfOtherAgentCol == secondAgent.selfCurrentCol {

		secondAgentRotatedSelfMap, secondAgentRotatedSelfCurrentRow, secondAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(secondAgent.selfMap, secondAgent.selfCurrentRow, secondAgent.selfCurrentCol)
		firstAgentTempSelfMap := firstAgent.selfMap

		rowOffset := firstAgent.selfOtherAgentRow - secondAgentRotatedSelfCurrentRow
		colOffset := firstAgent.selfOtherAgentCol - secondAgentRotatedSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if secondAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					firstAgentTempSelfMap[row+rowOffset][col+colOffset] = secondAgentRotatedSelfMap[row][col]
				}

			}
		}

		firstAgentRotatedSelfMap, firstAgentRotatedSelfCurrentRow, firstAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(rotate90DegreeClockwise(rotate90DegreeClockwise(firstAgent.selfMap, firstAgent.selfCurrentRow, firstAgent.selfCurrentCol)))
		secondAgentTempSelfMap := secondAgent.selfMap

		rowOffset = secondAgent.selfOtherAgentRow - firstAgentRotatedSelfCurrentRow
		colOffset = secondAgent.selfOtherAgentCol - firstAgentRotatedSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if firstAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					secondAgentTempSelfMap[row+rowOffset][col+colOffset] = firstAgentRotatedSelfMap[row][col]
				}

			}
		}
		firstAgent.selfMap = firstAgentTempSelfMap
		secondAgent.selfMap = secondAgentTempSelfMap

	}
	//haritayı çeviriyoruz real mapte ama agent mapinde çeviriyor muyuz?
	//15th
	if firstAgent.selfOtherAgentRow == firstAgent.selfCurrentRow &&
		firstAgent.selfOtherAgentCol == firstAgent.selfCurrentCol-1 &&
		secondAgent.selfOtherAgentRow == secondAgent.selfCurrentRow+1 &&
		secondAgent.selfOtherAgentCol == secondAgent.selfCurrentCol {

		secondAgentRotatedSelfMap, secondAgentRotatedSelfCurrentRow, secondAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(rotate90DegreeClockwise(rotate90DegreeClockwise(secondAgent.selfMap, secondAgent.selfCurrentRow, secondAgent.selfCurrentCol)))
		firstAgentTempSelfMap := firstAgent.selfMap

		rowOffset := firstAgent.selfOtherAgentRow - secondAgentRotatedSelfCurrentRow
		colOffset := firstAgent.selfOtherAgentCol - secondAgentRotatedSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if secondAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					firstAgentTempSelfMap[row+rowOffset][col+colOffset] = secondAgentRotatedSelfMap[row][col]
				}

			}
		}

		firstAgentRotatedSelfMap, firstAgentRotatedSelfCurrentRow, firstAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(firstAgent.selfMap, firstAgent.selfCurrentRow, firstAgent.selfCurrentCol)
		secondAgentTempSelfMap := secondAgent.selfMap

		rowOffset = secondAgent.selfOtherAgentRow - firstAgentRotatedSelfCurrentRow
		colOffset = secondAgent.selfOtherAgentCol - firstAgentRotatedSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if firstAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					secondAgentTempSelfMap[row+rowOffset][col+colOffset] = firstAgentRotatedSelfMap[row][col]
				}

			}
		}

		firstAgent.selfMap = firstAgentTempSelfMap
		secondAgent.selfMap = secondAgentTempSelfMap

	}
	//16th
	if firstAgent.selfOtherAgentRow == firstAgent.selfCurrentRow+1 &&
		firstAgent.selfOtherAgentCol == firstAgent.selfCurrentCol &&
		secondAgent.selfOtherAgentRow == secondAgent.selfCurrentRow+1 &&
		secondAgent.selfOtherAgentCol == secondAgent.selfCurrentCol {

		secondAgentRotatedSelfMap, secondAgentRotatedSelfCurrentRow, secondAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(rotate90DegreeClockwise(secondAgent.selfMap, secondAgent.selfCurrentRow, secondAgent.selfCurrentCol))
		firstAgentTempSelfMap := firstAgent.selfMap

		rowOffset := firstAgent.selfOtherAgentRow - secondAgentRotatedSelfCurrentRow
		colOffset := firstAgent.selfOtherAgentCol - secondAgentRotatedSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if secondAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					firstAgentTempSelfMap[row+rowOffset][col+colOffset] = secondAgentRotatedSelfMap[row][col]
				}

			}
		}

		firstAgentRotatedSelfMap, firstAgentRotatedSelfCurrentRow, firstAgentRotatedSelfCurrentCol := rotate90DegreeClockwise(rotate90DegreeClockwise(firstAgent.selfMap, firstAgent.selfCurrentRow, firstAgent.selfCurrentCol))
		secondAgentTempSelfMap := secondAgent.selfMap

		rowOffset = secondAgent.selfOtherAgentRow - firstAgentRotatedSelfCurrentRow
		colOffset = secondAgent.selfOtherAgentCol - firstAgentRotatedSelfCurrentCol

		for row := 0; row < AgentMapSize; row++ {
			for col := 0; col < AgentMapSize; col++ {
				if firstAgentRotatedSelfMap[row][col] != U && row+rowOffset < AgentMapSize && row+rowOffset > -1 && col+colOffset < AgentMapSize && col+colOffset > -1 {
					secondAgentTempSelfMap[row+rowOffset][col+colOffset] = firstAgentRotatedSelfMap[row][col]
				}

			}
		}

		firstAgent.selfMap = firstAgentTempSelfMap
		secondAgent.selfMap = secondAgentTempSelfMap

	}
	firstAgent.selfOtherAgentRow = -1
	firstAgent.selfOtherAgentCol = -1
	secondAgent.selfOtherAgentRow = -1
	secondAgent.selfOtherAgentCol = -1
}
func main() {
	pixelgl.Run(run)
}
