package main

import (
	"fmt"
	"math"
	"os"
)

var zones []Zone
var drones []Drone
var players []Player
var sortedDronesOfZone = make(map[Zone][]Drone)
var objectivies []Objective
var sortedObjectivies []Objective

// Pos represents a point in space
type Pos struct {
	x int
	y int
}

// Zone represent a zone
type Zone struct {
	id      int
	ownerID int
	center  Pos
	turns   map[int]int
}

// Drone represent a drone
type Drone struct {
	id      string
	ownerID int
	pos     Pos
}

// Player represent a player
type Player struct {
	id     int
	drones []Drone
	zones  []Zone
}

// Objective represent zone objective
type Objective struct {
	zone       Zone
	drones     []Drone
	needed     int
	controlled bool
}

// Distance from p1 to p2
func (p1 Pos) distance(p2 Pos) float64 {
	return math.Sqrt(math.Pow(float64(p2.x-p1.x), 2) + math.Pow(float64(p2.y-p1.y), 2))
}

func moveToTarget(target Pos) {
	fmt.Printf("%d %d\n", target.x, target.y)
}

func buildVoronoiDiagram() map[Drone]Zone {
	voronoi := make(map[Drone]Zone)
	for _, drone := range drones {
		delete(voronoi, drone)
		for _, zone := range zones {
			tmp := drone.pos.distance(zone.center)
			_, exist := voronoi[drone]
			if !exist || tmp < drone.pos.distance(voronoi[drone].center) {
				voronoi[drone] = zone
			}
		}
	}
	return voronoi
}

func getOwnedVoronoi(voronoi map[Drone]Zone, drone Drone) Zone {
	var z Zone
	for d, zone := range voronoi {
		if d.pos.x == drone.pos.x && d.pos.y == drone.pos.y {
			z = zone
		}
	}
	return z
}

func getDronesOfPlayer(playerID int) []Drone {
	for _, player := range players {
		if player.id == playerID {
			return player.drones
		}
	}
	return nil
}

func setZoneOwner(zoneID, playerID int) {
	for _, zone := range zones {
		if zone.id == zoneID {
			zone.ownerID = playerID
		}
	}
}

func sortDronesPerZone() {
	for _, zone := range zones {
		delete(sortedDronesOfZone, zone)
		nDrones := len(drones)
		zDrones := make([]Drone, nDrones)
		for i := 0; i < nDrones; i++ {
			minDistance := drones[i].pos.distance(zone.center)
			minDrone := drones[i]
			for j := i + 1; j < nDrones; j++ {
				d := drones[j].pos.distance(zone.center)
				if d < minDistance {
					minDrone = drones[j]
					minDistance = d
				}
			}
			zDrones[i] = minDrone
		}
		sortedDronesOfZone[zone] = zDrones
	}
}

func buildObjectivies() {
	objectivies = make([]Objective, len(zones))
	i := 0
	for zone, drones := range sortedDronesOfZone {
		objective := Objective{zone: zone, drones: drones, needed: 1, controlled: false}
		objectivies[i] = objective
		i++
	}
}

func sortObjectivies() {
	sortedObjectivies = make([]Objective, len(zones))
}

func (p1 Pos) minus(p2 Pos) Pos {
	return Pos{p1.x - p2.x, p1.y - p2.y}
}

func norm(p Pos) float64 {
	return math.Sqrt(math.Pow(float64(p.x), 2) + math.Pow(float64(p.y), 2))
}

func (p1 Pos) turns2(p2 Pos) int {
	return (int)(math.Ceil(norm(p1.minus(p2))/100) + 0.1)
}

func (p1 Pos) turns2Zone(zone Zone) int {
	return (int)(math.Ceil((norm(p1.minus(zone.center))-100)/100) + 0.1)
}

func (drone Drone) turns2(p Pos) int {
	return drone.pos.turns2(p)
}

func (drone Drone) turns2Zone(zone Zone) int {
	return drone.pos.turns2Zone(zone)
}

func getNBestObjectivies() {

}

func setTurnsForZones() {
	for _, zone := range zones {
		turns := make(map[int]int)
		for _, player := range players {
			turns[player.id] = player.turns2(zone)
		}
	}
}

func (player Player) turns2(zone Zone) int {
	turns := 0
	for _, drone := range player.drones {
		turns += drone.turns2Zone(zone)
	}
	return turns
}

func sortZonesByTurns() []Zone {
	nZones := len(zones)
	sorted := make([]Zone, nZones)
	for i := 0; i < nZones; i++ {
		minTurns := zones[i].turns[myID]
		minZone := zones[i]
		for j := i + 1; j < nZones; j++ {
			tmp := zones[j].turns[myID]
			if tmp < minTurns {
				minZone = zones[j]
				minTurns = tmp
			}
		}
		sorted[i] = minZone
	}
	return sorted
}

var myID int

func main() {
	// P: number of players in the game (2 to 4 players)
	// ID: ID of your player (0, 1, 2, or 3)
	// D: number of drones in each team (3 to 11)
	// Z: number of zones on the map (4 to 8)
	var P, ID, D, Z int
	fmt.Scan(&P, &ID, &D, &Z)

	fmt.Fprintln(os.Stderr, "P =", P, ", ID =", ID, ", D =", D, ", Z =", Z)
	myID = ID

	zones = make([]Zone, Z)
	for i := 0; i < Z; i++ {
		// X: corresponds to the position of the center of a zone. A zone is a circle with a radius of 100 units.
		var X, Y int
		fmt.Scan(&X, &Y)
		zones[i] = Zone{id: i, ownerID: -1, center: Pos{X, Y}}
	}
	for {
		var controllers []int
		for i := 0; i < Z; i++ {
			// TID: ID of the team controlling the zone (0, 1, 2, or 3) or -1 if it is not controlled.
			// The zones are given in the same order as in the initialization.
			var TID int
			fmt.Scan(&TID)

			setZoneOwner(i, TID)
			controllers = append(controllers, TID)
		}
		fmt.Fprintln(os.Stderr, "Controllers: ", controllers)
		fmt.Fprintln(os.Stderr, "Zones: ", zones)

		drones = make([]Drone, P*D)
		players = make([]Player, P)
		k := 0
		for i := 0; i < P; i++ {
			player := Player{id: i}
			pDrones := make([]Drone, D)
			for j := 0; j < D; j++ {
				// DX: The first D lines contain the coordinates of drones of a player with the ID 0,
				// the following D lines those of the drones of player 1, and thus it continues until the last player.
				var DX, DY int
				fmt.Scan(&DX, &DY)

				drone := Drone{fmt.Sprintf("%d-%d", i, j), i, Pos{DX, DY}}
				drones[k] = drone
				pDrones[j] = drone
				k++
			}
			player.drones = pDrones
			players[i] = player
		}

		setTurnsForZones()

		sortedZones := sortZonesByTurns()

		for i := 0; i < D; i++ {

			// fmt.Fprintln(os.Stderr, "Debug messages...")

			// output a destination point to be reached by one of your drones.
			// The first line corresponds to the first of your drones that you were provided as input, the next to the second, etc.

			drone := myDrones[i]
			zone := getOwnedVoronoi(voronoi, drone)
			fmt.Fprintln(os.Stderr, "drone: ", drone, "zone: ", zone)
			moveToTarget(zone.center)
		}
	}
}
