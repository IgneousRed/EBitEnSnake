package main

import (
	"fmt"
	"math"

	eb "github.com/IgneousRed/EBitEn"
	m "github.com/IgneousRed/gomisc"
	ebt "github.com/hajimehoshi/ebiten/v2"
)

var windowSize = m.Vec2I(800, 600)
var tileCount = m.Vec2I(16, 12)
var tileSize = windowSize.Float32().Div(tileCount.Float32())

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

type game struct {
	rng       m.PCG32
	tileTaken [][]bool
	dir       direction
	head      m.Vec[int]
	body      []m.Vec[int]
	food      m.Vec[int]
	moves     m.Queue[direction]
	tick      int
}

func (g *game) moveHead(dir direction) m.Vec[int] {
	ang := float32(dir) * math.Pi / 2.
	return g.head.Add(m.Vec2F(m.Cos(ang), m.Sin(ang)).RoundI()).Wrap(tileCount)
}
func (g *game) randomPoint() m.Vec[int] {
	return m.Vec2I(g.rng.Range(tileCount[0]), g.rng.Range(tileCount[1]))
}
func (g *game) taken(point m.Vec[int]) *bool {
	return &g.tileTaken[point[0]][point[1]]
}
func (g *game) placeFood() {
	for {
		p := g.randomPoint()
		if !*g.taken(p) {
			*g.taken(p) = true
			g.food = p
			return
		}
	}
}
func gameNew() game {
	g := game{}
	g.rng = m.PCG32Init()
	g.tileTaken = m.Make2[bool](tileCount[0], tileCount[1])
	g.dir = direction(g.rng.Range(4))
	g.head = g.randomPoint()
	g.body = append(g.body, g.moveHead(g.dir.opposite()))
	g.placeFood()
	return g
}
func (g *game) Update() {
	// Direction
	dirs := []bool{
		eb.KeysDown(ebt.KeyArrowRight, ebt.KeyD),
		eb.KeysDown(ebt.KeyArrowUp, ebt.KeyW),
		eb.KeysDown(ebt.KeyArrowLeft, ebt.KeyA),
		eb.KeysDown(ebt.KeyArrowDown, ebt.KeyS),
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
		return
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
		return
	}
	if newHead.Equals(g.food) {
		g.body = append(g.body, g.head)
		g.head = newHead
		g.placeFood()
		return
	}
	fmt.Println("Scored", len(g.body)-1)
	*g = gameNew()
}
func (g *game) drawTile(p m.Vec[int], col eb.Color) {
	eb.DrawRectangleF(tileSize.Mul(p.Float32()), tileSize, col)
}
func (g *game) Draw() {
	g.drawTile(g.head, eb.Red)
	for _, b := range g.body {
		g.drawTile(b, eb.Green)
	}
	g.drawTile(g.food, eb.Blue)
}
func main() {
	g := gameNew()
	eb.InitGame("Snake", windowSize, &g)
}
