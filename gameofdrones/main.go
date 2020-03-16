package main

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

var nTarget int

var isTargetInitialized bool

var targets []*Target

var centerOfTargets Pos

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
	id      string
	ownerID int
	pos     Pos
	prevPos *Pos
	task    *Task
}

// Player represent a player
type Player struct {
	id     int
	drones map[string]*Drone
}

// Target represent my target zones
type Target struct {
	zone         *Zone
	completed    bool
	taskAssigned bool
	task         *Task
}

// Task represent a task of a drone
type Task struct {
	target *Target
	done   bool
	drones map[string]*Drone
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
			tmp := drone.pos.distance(zone.center)
			_, exist := voronoi[drone]
			if !exist || tmp < drone.pos.distance(voronoi[drone].center) {
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

func getCenterOfTargets() Pos {
	x, y := 0, 0
	for _, target := range targets {
		x += target.zone.center.x
		y += target.zone.center.y
	}
	return Pos{x / 3, y / 3}
}

func (drone *Drone) moveToCenterOfTargets() {
	fmt.Printf("%d %d\n", centerOfTargets.x, centerOfTargets.y)
}

func (drone *Drone) isTaskDone() bool {
	if drone.isInside(drone.task.target.zone) {
		fmt.Fprintln(os.Stderr, "Task Done:", drone.id, ":", drone.task.target.zone.id)
		drone.task.done = true
		drone.task = nil
		return true
	}
	return false
}

func (drone *Drone) doTask() {
	if nil != drone.task {
		drone.moveToZone(drone.task.target.zone)
	}
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

func getMaxDrones(zone *Zone) int {
	max := 0
	for _, player := range allPlayers {
		n := 0
		for _, drone := range zone.inDrones {
			if drone.ownerID == player.id {
				n++
			}
		}
		if n > max {
			max = n
		}
	}
	return max
}

func getNBestDrones(zone *Zone, n int) []*Drone {
	myself := allPlayers[myID]
	sortedDrones := myself.getSortedDrones(zone)
	var drones []*Drone
	for _, drone := range sortedDrones {
		if drone.task == nil {
			drones = append(drones, drone)
		}
	}
	if len(drones) < n {
		drones = make([]*Drone, 0)
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

func (target *Target) checkTask() {
	if target.taskAssigned && target.task != nil {
		done := true
		for _, drone := range target.task.drones {
			_, ok := target.zone.inDrones[drone.id]
			if !ok {
				done = false
			}
		}
		target.task.done = done
	}
}

func main() {
	// P: number of players in the game (2 to 4 players)
	// ID: ID of your player (0, 1, 2, or 3)
	// D: number of drones in each team (3 to 11)
	// Z: number of zones on the map (4 to 8)
	fmt.Scan(&nPlayers, &myID, &nDrones, &nZones)

	nTarget = nZones
	fmt.Fprintln(os.Stderr, "nPlayers =", nPlayers, ", myID =", myID, ", nDrones =", nDrones, ", nZones =", nZones, ", nTarget = ", nTarget)

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
					drone = &Drone{id: droneID, ownerID: i, pos: Pos{DX, DY}, prevPos: nil}
					player.drones[droneID] = drone
				} else {
					drone.prevPos = &drone.pos
					drone.pos = Pos{DX, DY}
				}
				delete(allDrones, droneID)
				allDrones[droneID] = drone

				drone.checkAndSetZone()
			}
			allPlayers[i] = player
		}

		if !isTargetInitialized {
			setTurnsForZones()
			sortedZones := getSortedZonesByTurns()
			targets = make([]*Target, nTarget)
			for i := 0; i < nTarget; i++ {
				zone := sortedZones[i]
				fmt.Fprintln(os.Stderr, "SortedZone", i, ":", zone.id, ":", zone.center)
				targets[i] = &Target{zone: zone, completed: false, nDrones: 0}
			}

			centerOfTargets = getCenterOfTargets()
			isTargetInitialized = true
			fmt.Fprintln(os.Stderr, "Targets:", len(targets), ", center: ", centerOfTargets)
		}

		printAllZones()

		// voronoi := buildVoronoiDiagram()

		// myself := allPlayers[myID]

		for _, target := range targets {

			if target.zone.ownerID == myID {
				target.completed = true
			}
			target.checkTask()
			if target.taskAssigned && !target.task.done {
				continue
			}

			if !target.completed {
				// n := myself.getNumberOfDronesInZone(target.zone)
				n := getMaxDrones(target.zone)
				fmt.Fprintln(os.Stderr, "Max drones in target:", target.zone.id, ":", target.zone.center, ":", n)
				// if need more than 3 drones to control a zone, abandon the target
				if n < 5 {
					needed := 1
					if target.zone.ownerID != -1 {
						needed = n + 1
					}
					drones := getNBestDrones(target.zone, needed)
					for _, drone := range drones {
						fmt.Fprintln(os.Stderr, "Set task:", drone.id, ":", target.zone.id)
						drone.task = &Task{target: target, done: false}
					}
				}
			} else {

			}
		}

		for _, drone := range getDronesOfPlayer(myID) {
			if drone.task != nil {
				drone.doTask()
			} else {
				// zone := getOwnedVoronoi(voronoi, drone)
				// drone.moveToZone(zone)
			}
		}
	}
}
