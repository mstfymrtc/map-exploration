package main

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"image/color"
	"time"
)

const MapWidth = 14
const MapHeight = 14

const TileSize = 40

var Margin = 0.99
var DelayInMs = 250

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

//define movements
var MoveForward = 0
var MoveBackward = 1
var MoveRight = 2
var MoveLeft = 3
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

	//buna en başta north diyeceğiz. agent en başta kendi haritasında initial olarak north'a bakıyorum diyecek.
	//yönü nereye bakarsa baksın (örn sola dönük olsa bile) farketmez. north'a bakıyor diyeceğiz.
	//sola bakıyor olduğunda tek değişen haritayı sola bakar şekilde render ederiz.
	//yani ne yöne baktığı önemli değil, sonuçta ben north'a yönelmiş vaziyetteyim diye varsayıyor.
	directionInSelfMap int

	currentDirection int
	nextDirection    int
}

func NewAgent(currentRow int, currentCol int, currentDirection int, color color.RGBA, name string) *Agent {
	return &Agent{currentCol: currentCol, currentRow: currentRow, currentDirection: currentDirection, color: color, selfCurrentCol: MapWidth, selfCurrentRow: MapHeight, name: name}
}

//define agents
var agents = []*Agent{
	NewAgent(2, 3, SOUTH, colornames.Blue, "mavi"),
	NewAgent(1, 1, NORTH, colornames.Orange, "turuncu"),
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, MapWidth*TileSize, MapHeight*TileSize),
		VSync:  true,
	}
	//agentMapCfg := pixelgl.WindowConfig{
	//	Title:  "Pixel Rocks!",
	//	Bounds: pixel.R(0, 0, MapWidth*TileSize, MapHeight*TileSize),
	//	VSync:  true,
	//}
	win, err := pixelgl.NewWindow(cfg)
	//var agentWindows []*pixelgl.Window
	//for i := 0; i < len(agents); i++ {
	//	win, err := pixelgl.NewWindow(agentMapCfg)
	//	print(err)
	//	agentWindows = append(agentWindows, win)
	//	print(agentWindows)
	//
	//}
	if err != nil {
		panic(err)
	}
	for !win.Closed() {
		win.Clear(colornames.Skyblue)
		//for i := 0; i < len(agentWindows); i++ {
		//	agentWindows[i].Clear(colornames.Skyblue)
		//}

		drawMap(win)
		drawAgents(win)
		win.Update()
		sense()
		plan()

		rotate()

		win.Clear(colornames.Skyblue)
		drawMap(win)
		drawAgents(win)

		//for i := 0; i < len(agents); i++ {
		//	drawAgentMap(agentWindows[i], agents[i])
		//}
		win.Update()
		time.Sleep(time.Duration(100) * time.Millisecond)

		//for i := 0; i < len(agents); i++ {
		//	agentWindows[i].Update()
		//}

		move()

		win.Clear(colornames.Skyblue)
		drawMap(win)
		drawAgents(win)

		//for i := 0; i < len(agents); i++ {
		//	drawAgentMap(agentWindows[i], agents[i])
		//}
		win.Update()

		//for i := 0; i < len(agents); i++ {
		//	agentWindows[i].Update()
		//}
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

		if agents[i].currentDirection == NORTH {
			agents[i].frontSensor = tileMap[agents[i].currentRow-1][agents[i].currentCol]
			agents[i].rearSensor = tileMap[agents[i].currentRow+1][agents[i].currentCol]
			agents[i].rightSensor = tileMap[agents[i].currentRow][agents[i].currentCol+1]
			agents[i].leftSensor = tileMap[agents[i].currentRow][agents[i].currentCol-1]
		} else if agents[i].currentDirection == EAST {
			agents[i].frontSensor = tileMap[agents[i].currentRow][agents[i].currentCol+1]
			agents[i].rearSensor = tileMap[agents[i].currentRow][agents[i].currentCol-1]
			agents[i].rightSensor = tileMap[agents[i].currentRow+1][agents[i].currentCol]
			agents[i].leftSensor = tileMap[agents[i].currentRow-1][agents[i].currentCol]
		} else if agents[i].currentDirection == SOUTH {
			agents[i].frontSensor = tileMap[agents[i].currentRow+1][agents[i].currentCol]
			agents[i].rearSensor = tileMap[agents[i].currentRow-1][agents[i].currentCol]
			agents[i].rightSensor = tileMap[agents[i].currentRow][agents[i].currentCol-1]
			agents[i].leftSensor = tileMap[agents[i].currentRow][agents[i].currentCol+1]
		} else {
			agents[i].frontSensor = tileMap[agents[i].currentRow][agents[i].currentCol-1]
			agents[i].rearSensor = tileMap[agents[i].currentRow][agents[i].currentCol+1]
			agents[i].rightSensor = tileMap[agents[i].currentRow-1][agents[i].currentCol]
			agents[i].leftSensor = tileMap[agents[i].currentRow+1][agents[i].currentCol]
		}

		fmt.Println(agents[i].name, agents[i].frontSensor, agents[i].rightSensor, agents[i].rearSensor, agents[i].leftSensor)

		//BURAYI COMMENTLEDİM!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		//updateSelfMap(agents[i])

		//up,down,right veya left'de bir agent var mı? var ise, exchange edecekler
		//for j := 0; j < len(agents); j++ {
		//	if (agents[i].currentRow-1 == agents[j].currentRow && agents[i].currentCol == agents[j].currentCol) ||
		//		(agents[i].currentRow+1 == agents[j].currentRow && agents[i].currentCol == agents[j].currentCol) ||
		//		(agents[i].currentCol+1 == agents[j].currentCol && agents[i].currentRow == agents[j].currentRow) ||
		//		(agents[i].currentCol-1 == agents[j].currentCol && agents[i].currentRow == agents[j].currentRow) {
		//		//var,exchange
		//		//burada bir karşılaşmada iki kez exchange yazıyor.
		//		// bu bir bug değil, aksine hem a agentından b'ye
		//		//hem de b agentından a'ya aktarım yapmamız gerekiyor zaten.
		//
		//		//println("exchange")
		//
		//	}
		//
		//}

	}
}

func plan() {
	for i := 0; i < len(agents); i++ {
		//şimdilik movement'ı random olarak seç
		// eğer sadece etrafları empty ise hareket et.
		//TODO:tek satıra alınabilir?
		if agents[i].frontSensor == E {
			if agents[i].currentDirection == NORTH {
				agents[i].nextDirection = NORTH

			} else if agents[i].currentDirection == EAST {

				agents[i].nextDirection = EAST

			} else if agents[i].currentDirection == SOUTH {

				agents[i].nextDirection = SOUTH

			} else {
				agents[i].nextDirection = WEST
			}

			//updateSelfPos(agents[i], "forward")

		} else if agents[i].rightSensor == E {

			if agents[i].currentDirection == NORTH {
				agents[i].nextDirection = EAST

			} else if agents[i].currentDirection == EAST {
				agents[i].nextDirection = SOUTH

			} else if agents[i].currentDirection == SOUTH {
				agents[i].nextDirection = WEST

			} else {
				agents[i].nextDirection = NORTH
			}

		} else if agents[i].leftSensor == E {

			if agents[i].currentDirection == NORTH {
				agents[i].nextDirection = WEST

			} else if agents[i].currentDirection == EAST {
				agents[i].nextDirection = NORTH

			} else if agents[i].currentDirection == SOUTH {
				agents[i].nextDirection = EAST
			} else {
				agents[i].nextDirection = SOUTH
			}

		} else if agents[i].rearSensor == E {

			if agents[i].currentDirection == NORTH {
				agents[i].nextDirection = SOUTH

			} else if agents[i].currentDirection == EAST {
				agents[i].nextDirection = WEST

			} else if agents[i].currentDirection == SOUTH {
				agents[i].nextDirection = NORTH

			} else {
				agents[i].nextDirection = EAST
			}

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
		case NORTH:
			agents[i].currentRow -= 1
		case EAST:
			agents[i].currentCol += 1
		case SOUTH:
			agents[i].currentRow += 1
		case WEST:
			agents[i].currentCol -= 1

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

func drawAgents(win *pixelgl.Window) {
	for i := 0; i < len(agents); i++ {
		agentCurrentCol, agentCurrentRow, agentcurrentDirection := float64(agents[i].currentCol), float64(agents[i].currentRow), agents[i].currentDirection
		imd := imdraw.New(nil)

		switch agentcurrentDirection {
		case NORTH:
			imd.Color = agents[i].color
			imd.Push(pixel.V(agentCurrentCol*TileSize, (MapHeight-agentCurrentRow-1)*TileSize))
			imd.Color = agents[i].color
			imd.Push(pixel.V(agentCurrentCol*TileSize+TileSize, (MapHeight-agentCurrentRow-1)*TileSize))
			imd.Color = agents[i].color
			imd.Push(pixel.V((agentCurrentCol*TileSize)+(TileSize/2), (MapHeight-agentCurrentRow-1)*TileSize+TileSize))
		case EAST:
			imd.Color = agents[i].color
			imd.Push(pixel.V(agentCurrentCol*TileSize, (MapHeight-agentCurrentRow-1)*TileSize+TileSize))
			imd.Color = agents[i].color
			imd.Push(pixel.V(agentCurrentCol*TileSize, (MapHeight-agentCurrentRow-1)*TileSize))
			imd.Color = agents[i].color
			imd.Push(pixel.V((agentCurrentCol*TileSize)+(TileSize), (MapHeight-agentCurrentRow-1)*TileSize+(TileSize/2)))
		case SOUTH:
			imd.Color = agents[i].color
			imd.Push(pixel.V(agentCurrentCol*TileSize+TileSize, (MapHeight-agentCurrentRow-1)*TileSize+TileSize))
			imd.Color = agents[i].color
			imd.Push(pixel.V(agentCurrentCol*TileSize, (MapHeight-agentCurrentRow-1)*TileSize+TileSize))
			imd.Color = agents[i].color
			imd.Push(pixel.V((agentCurrentCol*TileSize)+(TileSize/2), (MapHeight-agentCurrentRow-1)*TileSize))
		case WEST:
			imd.Color = agents[i].color
			imd.Push(pixel.V(agentCurrentCol*TileSize+TileSize, (MapHeight-agentCurrentRow-1)*TileSize))
			imd.Color = agents[i].color
			imd.Push(pixel.V(agentCurrentCol*TileSize+TileSize, (MapHeight-agentCurrentRow-1)*TileSize+TileSize))
			imd.Color = agents[i].color
			imd.Push(pixel.V(agentCurrentCol*TileSize, (MapHeight-agentCurrentRow-1)*TileSize+(TileSize/2)))
		}

		imd.Polygon(0)
		imd.Draw(win)

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

//agentin kendi haritasındaki directionuna bağlı olarak,
//gerçek haritada sense ettiği veriyi kendi haritasına doğru şekilde işlemek
func updateSelfMap(agent *Agent) {
	if agent.directionInSelfMap == NORTH {
		//normal haritada frontSensorde gördüğünü, kendi haritasında, önüne işaretle
		agent.selfMap[agent.selfCurrentRow-1][agent.selfCurrentCol] = agent.frontSensor
		//normal haritada rearSensorde gördüğünü, kendi haritasında, arkana işaretle
		agent.selfMap[agent.selfCurrentRow+1][agent.selfCurrentCol] = agent.rearSensor
		agent.selfMap[agent.selfCurrentRow][agent.selfCurrentCol+1] = agent.rightSensor
		agent.selfMap[agent.selfCurrentRow][agent.selfCurrentCol-1] = agent.leftSensor
	} else if agent.directionInSelfMap == EAST {
		agent.selfMap[agent.selfCurrentRow-1][agent.selfCurrentCol] = agent.rightSensor
		agent.selfMap[agent.selfCurrentRow+1][agent.selfCurrentCol] = agent.leftSensor
		agent.selfMap[agent.selfCurrentRow][agent.selfCurrentCol+1] = agent.rearSensor
		agent.selfMap[agent.selfCurrentRow][agent.selfCurrentCol-1] = agent.frontSensor
	} else if agent.directionInSelfMap == WEST {
		agent.selfMap[agent.selfCurrentRow-1][agent.selfCurrentCol] = agent.leftSensor
		agent.selfMap[agent.selfCurrentRow+1][agent.selfCurrentCol] = agent.rightSensor
		agent.selfMap[agent.selfCurrentRow][agent.selfCurrentCol+1] = agent.frontSensor
		agent.selfMap[agent.selfCurrentRow][agent.selfCurrentCol-1] = agent.rearSensor
	} else {
		agent.selfMap[agent.selfCurrentRow-1][agent.selfCurrentCol] = agent.rearSensor
		agent.selfMap[agent.selfCurrentRow+1][agent.selfCurrentCol] = agent.frontSensor
		agent.selfMap[agent.selfCurrentRow][agent.selfCurrentCol+1] = agent.leftSensor
		agent.selfMap[agent.selfCurrentRow][agent.selfCurrentCol-1] = agent.rightSensor
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

func main() {
	pixelgl.Run(run)
}
