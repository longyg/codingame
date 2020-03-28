package main-3

import (
	"fmt"
	"math"
	"os"
)

var nPlayers, myID, nDrones, nZones int
var allZones = make(map[int]*Zone)
var allDrones = make(map[string]*Drone)
var allPlayers = make(map[int]*Player)

// RADIUS of zone
var RADIUS = 100

var objectivies []*Objective

// Pos represents a point in space
type Pos struct {
	x int
	y int
}

// Zone represent a zone
type Zone struct {
	id        int
	prevOwner int
	ownerID   int
	center    Pos
	turns     map[int]int
	inDrones  map[string]*Drone
}

// Drone represent a drone
type Drone struct {
	id        string
	ownerID   int
	pos       Pos
	prevPos   Pos
	objective *Objective
}

// Player represent a player
type Player struct {
	id     int
	drones map[string]*Drone
}

// Objective represent possible target
type Objective struct {
	zone   *Zone
	turns  int
	drones []*Drone
}

func buildObjectivies() {
	objectivies = make([]*Objective, 0)
	for _, zone := range allZones {
		for i := 1; i <= zone.getMaxTurns(); i++ {
			drones := zone.findDronesWithTurns(i)
			if nil != drones && len(drones) > 0 {
				objectivies = append(objectivies, &Objective{zone, i, drones})
			}
		}
	}
}

func sortObjectivies() {
	nObjective := len(objectivies)
	for i := 0; i < nObjective; i++ {
		for j := i + 1; j < nObjective; j++ {
			if objectivies[i].turns > objectivies[j].turns {
				tmp := objectivies[j]
				objectivies[j] = objectivies[i]
				objectivies[i] = tmp
			}
		}
	}
}

func (zone *Zone) getMaxTurns() int {
	maxTurns := 0
	for _, drone := range allDrones {
		turns := drone.turns2Zone(zone)
		if turns > maxTurns {
			maxTurns = turns
		}
	}
	return maxTurns
}

func (zone *Zone) findDronesWithTurns(turns int) []*Drone {
	var drones []*Drone
	for _, drone := range allDrones {
		if drone.turns2Zone(zone) == turns {
			drones = append(drones, drone)
		}
	}
	return drones
}

// Distance from p1 to p2
func (p1 Pos) distance(p2 Pos) float64 {
	return math.Sqrt(math.Pow(float64(p2.x-p1.x), 2) + math.Pow(float64(p2.y-p1.y), 2))
}

func moveToTarget(target Pos) {
	fmt.Printf("%d %d\n", target.x, target.y)
}

func buildVoronoiDiagram() map[*Drone]*Zone {
	voronoi := make(map[*Drone]*Zone)
	for _, drone := range allDrones {
		delete(voronoi, drone)
		for _, zone := range allZones {
			tmp := drone.turns2Zone(zone)
			_, exist := voronoi[drone]
			if !exist || tmp < drone.turns2Zone(voronoi[drone]) {
				voronoi[drone] = zone
			}
		}
	}
	return voronoi
}

func getOwnedVoronoi(voronoi map[*Drone]*Zone, drone *Drone) *Zone {
	var z *Zone
	for d, zone := range voronoi {
		if d.pos.x == drone.pos.x && d.pos.y == drone.pos.y {
			z = zone
		}
	}
	return z
}

func getDronesOfPlayer(playerID int) map[string]*Drone {
	for _, player := range allPlayers {
		if player.id == playerID {
			return player.drones
		}
	}
	return nil
}

func setZoneOwner(zoneID, playerID int) {
	for id, zone := range allZones {
		if id == zoneID && zone.id == zoneID {
			zone.prevOwner = zone.ownerID
			zone.ownerID = playerID
		}
	}
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

func (p1 Pos) turns2Zone(zone *Zone) int {
	return (int)(math.Ceil((norm(p1.minus(zone.center))-100)/100) + 0.1)
}

func (drone *Drone) turns2(p Pos) int {
	return drone.pos.turns2(p)
}

func (drone *Drone) turns2Zone(zone *Zone) int {
	return drone.pos.turns2Zone(zone)
}

func (player *Player) turns2(zone *Zone) int {
	turns := 0
	for _, drone := range player.drones {
		turns += drone.turns2Zone(zone)
	}
	return turns
}

func (drone *Drone) isInside(zone *Zone) bool {
	dist := math.Pow(float64(zone.center.x-drone.pos.x), 2) + math.Pow(float64(zone.center.y-drone.pos.y), 2)
	if dist > math.Pow(float64(RADIUS), 2) {
		return false
	}
	return true
}

func setTurnsForZones() {
	for _, zone := range allZones {
		turns := make(map[int]int)
		for _, player := range allPlayers {
			turns[player.id] = player.turns2(zone)
		}
		zone.turns = turns
	}
}

func getSortedZonesByTurns() []*Zone {
	var tmpZones []*Zone
	for _, zone := range allZones {
		tmpZones = append(tmpZones, zone)
	}

	for i := 0; i < len(tmpZones); i++ {
		for j := i + 1; j < len(tmpZones); j++ {
			tmp1 := tmpZones[i].turns[myID]
			tmp2 := tmpZones[j].turns[myID]
			if tmp2 < tmp1 {
				tmp := tmpZones[j]
				tmpZones[j] = tmpZones[i]
				tmpZones[i] = tmp
			}
		}
	}
	return tmpZones
}

func (player *Player) getSortedDrones(zone *Zone) []*Drone {
	n := len(player.drones)
	var drones []*Drone
	for _, drone := range player.drones {
		drones = append(drones, drone)
	}

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			tmp1 := drones[i].pos.turns2Zone(zone)
			tmp2 := drones[j].pos.turns2Zone(zone)
			if tmp2 < tmp1 {
				tmp := drones[j]
				drones[j] = drones[i]
				drones[i] = tmp
			}
		}
	}
	return drones
}

func (drone *Drone) moveToZone(zone *Zone) {
	fmt.Fprintln(os.Stderr, "drone [", drone.id, "] move to zone [", zone.id, "]:", zone.center)
	fmt.Printf("%d %d\n", zone.center.x, zone.center.y)
}

func (drone *Drone) moveToPos(pos Pos) {
	fmt.Printf("%d %d\n", pos.x, pos.y)
}

func (zone *Zone) addDrone(drone *Drone) {
	if nil == zone.inDrones {
		zone.inDrones = make(map[string]*Drone)
	}
	_, ok := zone.inDrones[drone.id]
	if !ok {
		fmt.Fprintln(os.Stderr, "drone [", drone.id, "] fly into zone [", zone.id, "]")
		zone.inDrones[drone.id] = drone
	}
}

func (drone *Drone) checkAndSetZone() {
	for _, zone := range allZones {
		if drone.isInside(zone) {
			zone.addDrone(drone)
		} else {
			_, ok := zone.inDrones[drone.id]
			if ok {
				fmt.Fprintln(os.Stderr, "drone [", drone.id, "] left zone [", zone.id, "]")
				delete(zone.inDrones, drone.id)
			}
		}
	}
}

// get number of drones which are already stayed in a zone
func (player *Player) getNumberOfDronesInZone(zone *Zone) int {
	n := 0
	for _, drone := range zone.inDrones {
		if drone.ownerID == player.id {
			n++
		}
	}
	return n
}

func (zone *Zone) getMaxDrones() (int, int) {
	max := 0
	pid := -1
	for _, player := range allPlayers {
		n := 0
		for _, drone := range zone.inDrones {
			if drone.ownerID == player.id {
				n++
			}
		}
		if n > max {
			max = n
			pid = player.id
		}
	}
	return pid, max
}

func (zone *Zone) hasDroneOfOthers() bool {
	for _, drone := range zone.inDrones {
		if drone.ownerID != myID {
			return true
		}
	}
	return false
}

func (zone *Zone) getMyDrones() []*Drone {
	var drones []*Drone
	for _, drone := range zone.inDrones {
		if drone.ownerID == myID {
			drones = append(drones, drone)
		}
	}
	return drones
}

func printAllZones() {
	fmt.Fprintln(os.Stderr, "All zones: ")
	for _, zone := range allZones {
		fmt.Fprintln(os.Stderr, " Zone: ", zone.id, ",", zone.ownerID, ",", len(zone.inDrones), zone.turns)
		for _, drone := range zone.inDrones {
			fmt.Fprintln(os.Stderr, "     Drone: ", drone.id, ",", drone.ownerID, ",", drone.pos)
		}
	}
}

func main() {
	// P: number of players in the game (2 to 4 players)
	// ID: ID of your player (0, 1, 2, or 3)
	// D: number of drones in each team (3 to 11)
	// Z: number of zones on the map (4 to 8)
	fmt.Scan(&nPlayers, &myID, &nDrones, &nZones)

	fmt.Fprintln(os.Stderr, "nPlayers =", nPlayers, ", myID =", myID, ", nDrones =", nDrones, ", nZones =", nZones)

	for i := 0; i < nZones; i++ {
		// X: corresponds to the position of the center of a zone. A zone is a circle with a radius of 100 units.
		var X, Y int
		fmt.Scan(&X, &Y)
		allZones[i] = &Zone{id: i, prevOwner: -1, ownerID: -1, center: Pos{X, Y}}
	}

	for {
		controllers := make([]int, nZones)
		for i := 0; i < nZones; i++ {
			// TID: ID of the team controlling the zone (0, 1, 2, or 3) or -1 if it is not controlled.
			// The zones are given in the same order as in the initialization.
			var TID int
			fmt.Scan(&TID)

			setZoneOwner(i, TID)
			controllers[i] = TID
		}
		fmt.Fprintln(os.Stderr, "Controllers:", controllers)

		for i := 0; i < nPlayers; i++ {
			player, ok := allPlayers[i]
			if !ok {
				player = &Player{id: i, drones: make(map[string]*Drone)}
			}

			for j := 0; j < nDrones; j++ {
				// DX: The first D lines contain the coordinates of drones of a player with the ID 0,
				// the following D lines those of the drones of player 1, and thus it continues until the last player.
				var DX, DY int
				fmt.Scan(&DX, &DY)

				droneID := fmt.Sprintf("%d-%d", i, j)
				drone, ok := player.drones[droneID]
				if !ok {
					drone = &Drone{id: droneID, ownerID: i, pos: Pos{DX, DY}, prevPos: Pos{DX, DY}}
					player.drones[droneID] = drone
				} else {
					drone.prevPos = drone.pos
					drone.pos = Pos{DX, DY}
				}
				delete(allDrones, droneID)
				allDrones[droneID] = drone

				drone.checkAndSetZone()
			}
			allPlayers[i] = player
		}

		buildObjectivies()
		sortObjectivies()

		for _, objective := range objectivies {
			for _, drone := range objective.drones {
				if drone.ownerID == myID && drone.objective == nil && !drone.isInside(objective.zone) {
					drone.objective = objective
				} else if drone.isInside(objective.zone) {
					drone.objective = nil
				}
			}
		}
		for _, drone := range getDronesOfPlayer(myID) {
			if drone.objective != nil {
				drone.moveToZone(drone.objective.zone)
			}
		}
	}
}
