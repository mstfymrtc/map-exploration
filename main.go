package main

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"image/color"
	"math/rand"
	"time"
)

const MapWidth = 14
const MapHeight = 14

const TileSize = 40

var Margin = 0.99
var DelayInMs = 0

//define constants
var U = 0
var E = 1
var F = 2
var EXIT = 3

//define directions
var NORTH = 0
var SOUTH = 1
var EAST = 2
var WEST = 3

var tileMap = [MapHeight][MapWidth]int{

	{F, F, F, F, EXIT, F, F, F, F, F, F, F, F, F},
	{F, E, F, E, E, E, E, F, E, E, E, E, E, F},
	{F, E, F, E, E, E, E, F, E, E, E, F, E, F},
	{F, E, F, F, E, E, E, F, E, E, E, F, F, F},
	{F, E, E, E, E, F, F, F, F, E, E, E, E, F},
	{F, E, E, E, E, F, E, E, E, E, E, E, E, F},
	{F, E, E, E, E, F, E, E, E, E, F, F, F, F},
	{F, E, E, E, E, F, F, E, E, E, F, E, E, F},
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
}

type Agent struct {
	currentCol     int
	currentRow     int
	selfCurrentCol int
	selfCurrentRow int

	selfMap [MapHeight * 2][MapWidth * 2]int

	color color.RGBA
	//agentın up down right ve left inde ne var (E,F,A olabilir)
	frontSensor int
	rearSensor  int
	rightSensor int
	leftSensor  int

	name string

	directionInRealMap int
}

func NewAgent(currentRow int, currentCol int, directionInRealMap int, color color.RGBA, name string) *Agent {
	return &Agent{currentCol: currentCol, currentRow: currentRow, directionInRealMap: directionInRealMap, color: color, selfCurrentCol: MapWidth, selfCurrentRow: MapHeight, name: name}
}

//define agents
var agents = []*Agent{
	//NewAgent(4, 3, EAST, colornames.Blue, "mavi"),
	NewAgent(2, 1, NORTH, colornames.Orange, "turuncu"),
	//NewAgent(3, 2, colornames.Pink),
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, MapWidth*TileSize, MapHeight*TileSize),
		VSync:  true,
	}
	agentMapCfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, MapWidth*TileSize, MapHeight*TileSize),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	var agentWindows []*pixelgl.Window
	for i := 0; i < len(agents); i++ {
		win, err := pixelgl.NewWindow(agentMapCfg)
		print(err)
		agentWindows = append(agentWindows, win)
		print(agentWindows)

	}
	if err != nil {
		panic(err)
	}
	for !win.Closed() {
		win.Clear(colornames.Skyblue)
		for i := 0; i < len(agentWindows); i++ {
			agentWindows[i].Clear(colornames.Skyblue)
		}

		drawMap(win)
		drawAgents(win)
		sense()
		plan()
		for i := 0; i < len(agents); i++ {
			drawAgentMap(agentWindows[i], agents[i])
		}
		win.Update()
		for i := 0; i < len(agents); i++ {
			agentWindows[i].Update()
		}
		time.Sleep(time.Duration(DelayInMs) * time.Millisecond)

	}

}

//real life'da yanımızda bir engel mi var yoksa agent mı var bunu tespit edebiliriz
// ama buradaki simülasyonda bu bilgiyi diğer agentların pozisyonlarını kontrol edereek elde edebiliriz.
func sense() {
	for i := 0; i < len(agents); i++ {
		//eğer agentlar harita kenarlarında ise up,down,right veya left'ı OOI olarak işaretle
		//yani harita dışına çıkartmama kontrolü
		//onun dışında up,down,right ve left'te ne varsa onu agentlara yaz.

		//real life'de sensör ile detection yapınca bunu görebiliyorlar, fakat
		//burada koordinat ile kontrol yapmak zorundayız.
		agents[i].frontSensor = tileMap[agents[i].currentRow-1][agents[i].currentCol]
		agents[i].rearSensor = tileMap[agents[i].currentRow+1][agents[i].currentCol]
		agents[i].rightSensor = tileMap[agents[i].currentRow][agents[i].currentCol+1]
		agents[i].leftSensor = tileMap[agents[i].currentRow][agents[i].currentCol-1]
		fmt.Println(agents[i].name, agents[i].frontSensor, agents[i].rightSensor, agents[i].rearSensor, agents[i].leftSensor)
		updateSelfMap(agents[i])
		//up,down,right veya left'de bir agent var mı? var ise, exchange edecekler
		for j := 0; j < len(agents); j++ {
			if (agents[i].currentRow-1 == agents[j].currentRow && agents[i].currentCol == agents[j].currentCol) ||
				(agents[i].currentRow+1 == agents[j].currentRow && agents[i].currentCol == agents[j].currentCol) ||
				(agents[i].currentCol+1 == agents[j].currentCol && agents[i].currentRow == agents[j].currentRow) ||
				(agents[i].currentCol-1 == agents[j].currentCol && agents[i].currentRow == agents[j].currentRow) {
				//var,exchange
				//burada bir karşılaşmada iki kez exchange yazıyor.
				// bu bir bug değil, aksine hem a agentından b'ye
				//hem de b agentından a'ya aktarım yapmamız gerekiyor zaten.

				//println("exchange")

			}

		}

	}
}
func updateSelfPos(agent *Agent, movement string) {
	if movement == "forward" {
		agent.selfCurrentRow -= 1
	} else if movement == "backward" {
		agent.selfCurrentRow += 1

	} else if movement == "right" {
		agent.selfCurrentCol += 1

	} else {
		agent.selfCurrentCol -= 1

	}

}
func updateSelfMap(agent *Agent) {
	//normal haritada frontSensorde gördüğünü, kendi haritasında, önüne işaretle
	agent.selfMap[agent.selfCurrentRow-1][agent.selfCurrentCol] = agent.frontSensor
	//normal haritada rearSensorde gördüğünü, kendi haritasında, arkana işaretle
	agent.selfMap[agent.selfCurrentRow+1][agent.selfCurrentCol] = agent.rearSensor
	agent.selfMap[agent.selfCurrentRow][agent.selfCurrentCol+1] = agent.rightSensor
	agent.selfMap[agent.selfCurrentRow][agent.selfCurrentCol-1] = agent.leftSensor

}

func plan() {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	for i := 0; i < len(agents); i++ {
		//şimdilik movement'ı random olarak seç
		num := r.Intn(4)
		// eğer sadece etrafları empty ise hareket et.
		//TODO:tek satıra alınabilir?
		if num == 0 {
			if agents[i].frontSensor == E {
				agents[i].currentRow -= 1 //up, agentı bir adım öne ilerlet
				updateSelfPos(agents[i], "forward")

			}
		} else if num == 1 { //down
			if agents[i].rearSensor == E {
				agents[i].currentRow += 1
				updateSelfPos(agents[i], "backward")

			}
		} else if num == 2 { //right
			if agents[i].rightSensor == E {
				agents[i].currentCol += 1
				updateSelfPos(agents[i], "right")

			}
		} else { //left

			if agents[i].leftSensor == E {
				agents[i].currentCol -= 1
				updateSelfPos(agents[i], "left")
			}
		}
	}

}

func drawMap(win *pixelgl.Window) {
	for row := 0; row < MapHeight; row++ {
		for col := 0; col < MapWidth; col++ {
			imd := imdraw.New(nil)
			imd.Color = tileColour[tileMap[row][col]]
			imd.Push(pixel.V(float64(col*TileSize), float64((MapHeight-row-1)*TileSize)))
			imd.Push(pixel.V(float64(col*TileSize+TileSize), float64((MapHeight-row-1)*TileSize+TileSize)))
			imd.Rectangle(0)
			imd.Draw(win)

		}
	}
}
func drawAgentMap(win *pixelgl.Window, agent *Agent) {
	for row := 0; row < MapHeight*2; row++ {
		for col := 0; col < MapWidth*2; col++ {
			imd := imdraw.New(nil)
			imd.Color = tileColour[agent.selfMap[row][col]]
			imd.Push(pixel.V(float64(col*(TileSize/2)), float64((MapHeight*2-row-1)*(TileSize/2))))
			imd.Push(pixel.V(float64(col*TileSize+(TileSize/2)), float64(((MapHeight*2-row-1)*TileSize+TileSize)/2)))
			imd.Rectangle(0)
			imd.Draw(win)

		}
	}
}

func drawAgents(win *pixelgl.Window) {
	for i := 0; i < len(agents); i++ {
		agentCurrentCol, agentCurrentRow := float64(agents[i].currentCol), float64(agents[i].currentRow)
		imd := imdraw.New(nil)
		imd.Color = agents[i].color
		imd.Push(pixel.V(agentCurrentCol*TileSize, (MapHeight-agentCurrentRow-1)*TileSize))
		imd.Color = agents[i].color
		imd.Push(pixel.V(agentCurrentCol*TileSize+TileSize, (MapHeight-agentCurrentRow-1)*TileSize))
		imd.Color = agents[i].color
		imd.Push(pixel.V((agentCurrentCol*TileSize)+(TileSize/2), (MapHeight-agentCurrentRow-1)*TileSize+TileSize))
		imd.Polygon(0)
		imd.Draw(win)

	}
}

func main() {
	pixelgl.Run(run)
}
