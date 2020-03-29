package main

import "fmt"

type Vec struct {
	x, y int
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
}
