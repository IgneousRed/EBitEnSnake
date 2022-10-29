package main

import (
	"fmt"
	"math"

	eb "github.com/IgneousRed/EBitEngine"
	m "github.com/IgneousRed/gomisc"
	ebit "github.com/hajimehoshi/ebiten/v2"
)

var WindowSize = m.Vec2I(800, 600)

type direction int

const (
	dRight direction = iota
	dUp
	dLeft
	dDown
	dCount int = iota
)

func (d direction) opposite() direction {
	return (d + 2) % direction(dCount)
}

type Game struct {
	rng       m.PCG32
	tileCount m.Vec[int]
	tileSize  m.Vec[float32]
	tileTaken [][]bool
	dir       direction
	head      m.Vec[int]
	body      []m.Vec[int]
	food      m.Vec[int]
	moves     m.Queue[direction]
	tick      int
}

func (g *Game) moveHead(dir direction) m.Vec[int] {
	ang := float32(dir) * math.Pi / 2.
	return g.head.Add(m.Vec2F(m.Cos(ang), m.Sin(ang)).RoundI()).Wrap(g.tileCount)
}
func (g *Game) RandomPoint() m.Vec[int] {
	return m.Vec2I(g.rng.Range(g.tileCount[0]), g.rng.Range(g.tileCount[1]))
}
func (g *Game) taken(point m.Vec[int]) *bool {
	return &g.tileTaken[point[0]][point[1]]
}
func (g *Game) PlaceFood() {
	for {
		p := g.RandomPoint()
		if !*g.taken(p) {
			*g.taken(p) = true
			g.food = p
			return
		}
	}
}
func GameNew() Game {
	g := Game{}
	g.rng = m.PCG32Init()
	g.tileCount = m.Vec2I(16, 12)
	g.tileSize = WindowSize.Float32().Div(g.tileCount.Float32())
	g.tileTaken = m.Make2[bool](g.tileCount[0], g.tileCount[1])
	g.dir = direction(g.rng.Range(4))
	g.head = g.RandomPoint()
	g.body = append(g.body, g.moveHead(g.dir.opposite()))
	g.PlaceFood()
	return g
}
func (g *Game) Update() error {
	// Direction
	eb.KeysUpdate()
	dirs := []bool{
		eb.KeysDown(ebit.KeyArrowRight, ebit.KeyD),
		eb.KeysDown(ebit.KeyArrowUp, ebit.KeyW),
		eb.KeysDown(ebit.KeyArrowLeft, ebit.KeyA),
		eb.KeysDown(ebit.KeyArrowDown, ebit.KeyS),
	}
	if m.CountTrue(dirs) == 1 {
		newDir := direction(m.FirstTrueIndex(dirs))
		l, err := g.moves.Last()
		last := m.Ternary(err == nil, l, g.dir)
		if newDir != last && newDir != last.opposite() {
			g.moves.Push(newDir)
		}
	}

	// Step
	g.tick++
	if g.tick < 10 {
		return nil
	}
	g.tick = 0
	if d, err := g.moves.Pop(); err == nil {
		g.dir = d
	}
	newHead := g.moveHead(g.dir)
	if !*g.taken(newHead) {
		*g.taken(newHead) = true
		*g.taken(g.body[0]) = false
		lastIndex := len(g.body) - 1
		copy(g.body[0:lastIndex], g.body[1:len(g.body)])
		g.body[lastIndex] = g.head
		g.head = newHead
		return nil
	}
	if newHead.Equals(g.food) {
		g.body = append(g.body, g.head)
		g.head = newHead
		g.PlaceFood()
		return nil
	}
	fmt.Println("Scored", len(g.body)-1)
	*g = GameNew()
	return nil
}
func (g *Game) drawTile(scr *ebit.Image, p m.Vec[int], col eb.Color) {
	eb.DrawRectangleF(scr, g.tileSize.Mul(p.Float32()), g.tileSize, col)
}
func (g *Game) Draw(scr *ebit.Image) {
	g.drawTile(scr, g.head, eb.Red)
	for _, b := range g.body {
		g.drawTile(scr, b, eb.Green)
	}
	g.drawTile(scr, g.food, eb.Blue)
}
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return WindowSize[0], WindowSize[1]
}
func main() {
	g := GameNew()
	eb.InitGame("Snake", WindowSize, &g)
}
