package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"image/color"
	"math/rand"
	"time"
)

const MapSizeX = 12
const MapSizeY = 12

const TileSize = 40

var Margin = 0.99
var DelayInMs = 0

//define constants
var U = 0
var E = 1
var F = 2
var OOI = 3 //OUT OF INDEX (HARİTA DIŞI, -1)

var tileMap = [MapSizeX][MapSizeY]int{{E, F, E, F, F, F, E, F, F, E, F, F},
	{F, E, F, F, E, E, F, F, E, E, F, F},
	{E, E, F, E, F, E, F, F, F, F, F, F},
	{F, F, E, E, E, E, E, E, E, E, E, E},
	{F, E, F, E, E, E, F, E, E, F, F, E},
	{E, E, F, E, E, E, F, F, F, F, E, E},
	{F, E, F, E, E, E, F, F, E, E, F, F},
	{F, E, F, E, E, E, F, E, E, F, F, E},
	{F, F, E, E, E, F, E, E, E, E, E, E},
	{E, F, E, F, F, F, E, E, F, E, F, F},
	{E, E, F, F, E, E, E, F, F, E, E, E},
	{F, E, F, F, E, F, F, F, E, E, F, F}}

var tileColour = map[int]color.RGBA{
	E:   colornames.White,
	F:   colornames.Brown,
	U:   colornames.Black,
	OOI: colornames.Gray,
}

type Agent struct {
	posX     int
	posY     int
	selfPosX int
	selfPosY int
	selfMap  [MapSizeX * 2][MapSizeY * 2]int
	color    color.RGBA
	//agentın up down right ve left inde ne var (E,F,A olabilir)
	up    int
	down  int
	right int
	left  int
}

func NewAgent(posX int, posY int, color color.RGBA) *Agent {
	return &Agent{posX: posX, posY: posY, color: color, selfPosX: MapSizeX, selfPosY: MapSizeY}
}

//define agents
var agents = []*Agent{
	NewAgent(4, 3, colornames.Blue),
	NewAgent(5, 5, colornames.Orange),
	//NewAgent(3, 2, colornames.Pink),
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, MapSizeX*TileSize, MapSizeY*TileSize),
		VSync:  true,
	}
	agentMapCfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, MapSizeX*TileSize, MapSizeY*TileSize),
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
		sense()
		plan()
		drawMap(win)
		drawAgents(win)
		for i := 0; i < len(agents); i++ {
			drawAgentMap(agentWindows[i], agents[i])
		}
		win.Update()
		for i := 0; i < len(agents); i++ {
			agentWindows[i].Update()
		}

	}

}

//real life'da yanımızda bir engel mi var yoksa agent mı var bunu tespit edebiliriz
// ama buradaki simülasyonda bu bilgiyi diğer agentların pozisyonlarını kontrol edereek elde edebiliriz.
func sense() {
	time.Sleep(time.Duration(DelayInMs) * time.Millisecond)
	for i := 0; i < len(agents); i++ {
		//eğer agentlar harita kenarlarında ise up,down,right veya left'ı OOI olarak işaretle
		//yani harita dışına çıkartmama kontrolü
		//onun dışında up,down,right ve left'te ne varsa onu agentlara yaz.

		//real life'de sensör ile detection yapınca bunu görebiliyorlar, fakat
		//burada koordinat ile kontrol yapmak zorundayız.
		if agents[i].posY == MapSizeY-1 {
			agents[i].up = OOI
		} else {
			agents[i].up = tileMap[agents[i].posX][agents[i].posY+1]
		}

		if agents[i].posY == 0 {
			agents[i].down = OOI
		} else {
			agents[i].down = tileMap[agents[i].posX][agents[i].posY-1]
		}

		if agents[i].posX == MapSizeX-1 {
			agents[i].right = OOI
		} else {
			agents[i].right = tileMap[agents[i].posX+1][agents[i].posY]
		}

		if agents[i].posX == 0 {
			agents[i].left = OOI
		} else {
			agents[i].left = tileMap[agents[i].posX-1][agents[i].posY]
		}
		updateSelfMap(agents[i])
		//up,down,right veya left'de bir agent var mı? var ise, exchange edecekler
		for j := 0; j < len(agents); j++ {
			if (agents[i].posY-1 == agents[j].posY && agents[i].posX == agents[j].posX) ||
				(agents[i].posY+1 == agents[j].posY && agents[i].posX == agents[j].posX) ||
				(agents[i].posX+1 == agents[j].posX && agents[i].posY == agents[j].posY) ||
				(agents[i].posX-1 == agents[j].posX && agents[i].posY == agents[j].posY) {
				//var,exchange
				//burada bir karşılaşmada iki kez exchange yazıyor.
				// bu bir bug değil, aksine hem a agentından b'ye
				//hem de b agentından a'ya aktarım yapmamız gerekiyor zaten.

				println("exchange")

			}

		}

	}
}
func updateSelfPos(agent *Agent, movement string) {
	if movement == "up" {
		agent.selfPosY += 1
	} else if movement == "down" {
		agent.selfPosY -= 1

	} else if movement == "right" {
		agent.selfPosX += 1

	} else {
		agent.selfPosX -= 1

	}

}
func updateSelfMap(agent *Agent) {
	agent.selfMap[agent.selfPosX][agent.selfPosY+1] = agent.up
	agent.selfMap[agent.selfPosX][agent.selfPosY-1] = agent.down
	agent.selfMap[agent.selfPosX+1][agent.selfPosY] = agent.right
	agent.selfMap[agent.selfPosX-1][agent.selfPosY] = agent.left

}

func plan() {
	for i := 0; i < len(agents); i++ {
		//şimdilik movement'ı random olarak seç
		rand.Seed(time.Now().UnixNano())
		num := rand.Intn(3-0+1) + 0
		// eğer sadece etrafları empty ise hareket et.
		//TODO:tek satıra alınabilir?
		if num == 0 {
			if agents[i].up == E {
				agents[i].posY += 1 //up
				updateSelfPos(agents[i], "up")

			}
		} else if num == 1 { //right
			if agents[i].right == E {
				agents[i].posX += 1
				updateSelfPos(agents[i], "right")

			}
		} else if num == 2 { //left
			if agents[i].left == E {
				agents[i].posX -= 1
				updateSelfPos(agents[i], "left")

			}
		} else { //down
			if agents[i].down == E {
				agents[i].posY -= 1
				updateSelfPos(agents[i], "down")

			}
		}
	}

}

func drawMap(win *pixelgl.Window) {
	for row := 0; row < MapSizeY; row++ {
		for col := 0; col < MapSizeX; col++ {
			imd := imdraw.New(nil)
			imd.Color = tileColour[tileMap[row][col]]
			imd.Push(pixel.V(float64(col*TileSize), float64(row*TileSize)))
			imd.Push(pixel.V(float64(col*TileSize+TileSize), float64(row*TileSize+TileSize)))
			imd.Rectangle(0)
			imd.Draw(win)

		}
	}
}
func drawAgentMap(win *pixelgl.Window, agent *Agent) {
	for row := 0; row < MapSizeY*2; row++ {
		for col := 0; col < MapSizeX*2; col++ {

			imd := imdraw.New(nil)
			imd.Color = tileColour[agent.selfMap[row][col]]
			imd.Push(pixel.V(float64(col*TileSize/2), float64(row*TileSize/2)))
			imd.Push(pixel.V(float64(col*TileSize+TileSize/2), float64((row*TileSize+TileSize)/2)))
			imd.Rectangle(0)
			imd.Draw(win)

		}
	}
}

func drawAgents(win *pixelgl.Window) {
	for i := 0; i < len(agents); i++ {

		imd := imdraw.New(nil)
		imd.Color = agents[i].color
		imd.Push(pixel.V(float64(agents[i].posY*TileSize), float64(agents[i].posX*TileSize)))
		imd.Push(pixel.V(float64(agents[i].posY*TileSize+TileSize), float64(agents[i].posX*TileSize+TileSize)))
		imd.Rectangle(0)
		imd.Draw(win)
	}
}
func main() {
	pixelgl.Run(run)
}
