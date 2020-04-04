package main

import (
	"fmt"
	"math"
)

type Vec struct {
	x, y int
}

type Zone struct {
	pos     Vec
}

func norm(vec Vec) float64 {
	return math.Sqrt(math.Pow(float64(vec.x), 2) + math.Pow(float64(vec.y), 2))
}

func (vec1 Vec) minus(vec2 Vec) Vec {
	return Vec{vec1.x - vec2.x, vec1.y - vec2.y}
}


func (vec1 Vec) turns2(vec2 Vec) int {
	return (int)(math.Ceil(norm(vec1.minus(vec2))/100) + 0.1)
}

func (vec1 Vec) turns2Zone(zone *Zone) int {
	return (int)(math.Ceil((norm(vec1.minus(zone.pos))-100)/100) + 0.1)
}

type Player struct {
	zoneCenter Vec
}

var players []*Player

func getAnotherPlayer() *Player {
	for i, player := range players {
		if i != myID {
			return player
		}
	}
	return nil
}

var myID int

func main() {
	myID = 0
	players = make([]*Player, 2)
	for i := 0; i < 2; i++ {
		player := Player{}
		player.zoneCenter = Vec{10, 20}
		players[i] = &player
	}

	var defaultV *Vec
	defaultV = &getAnotherPlayer().zoneCenter

	cur := defaultV

	fmt.Println(defaultV)

	for i := 0; i < 2; i++ {
		player := players[i]
		player.zoneCenter = Vec {2, 1}
	}
	fmt.Println(defaultV)
	fmt.Println(cur)

	vec := Vec{1033, 29}
	vec2 := Vec{3222, 101}
	zone := Zone{pos: vec2}
	fmt.Println(vec.turns2Zone(&zone))
	fmt.Println(vec.turns2(vec2))

	var i float64 = 0
	fmt.Println(i)

	fmt.Println()
	fmt.Println("=======================")
	fmt.Println()

	v := Vec{3799, 1254}
	z := Zone{pos: Vec{3043, 1183}}
	fmt.Println(v.turns2Zone(&z))
}
