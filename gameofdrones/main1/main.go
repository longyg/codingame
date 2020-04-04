package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
)

type Vec struct {
	x, y int
}

func (vec1 Vec) turns2(vec2 Vec) int {
	return (int)(math.Ceil(norm(vec1.minus(vec2))/100) + 0.1)
}

func (vec1 Vec) turns2Zone(zone *Zone) int {
	return (int)(math.Ceil((norm(vec1.minus(zone.pos))-100)/100) + 0.1)
}

func det(a, b Vec) int {
	return a.x*b.y - a.y*b.x
}
func dot(a, b Vec) int {
	return a.x*b.x + a.y*b.y
}

func norm(vec Vec) float64 {
	return math.Sqrt(math.Pow(float64(vec.x), 2) + math.Pow(float64(vec.y), 2))
}

func (vec1 Vec) minus(vec2 Vec) Vec {
	return Vec{vec1.x - vec2.x, vec1.y - vec2.y}
}

func (vec1 Vec) add(vec2 Vec) Vec {
	return Vec{vec1.x + vec2.x, vec1.y + vec2.y}
}

func (vec1 Vec) divide(f float64) Vec {
	return Vec{int(float64(vec1.x) / f), int(float64(vec1.y) / f)}
}

func (vec1 Vec) multiply(f float64) Vec {
	return Vec{int(float64(vec1.x) * f), int(float64(vec1.y) * f)}
}

func multiply(b float64, a Vec) Vec {
	return Vec{int(float64(a.x) * b), int(float64(a.y) * b)}
}

type Zone struct {
	pos     Vec
	ownerId int
	id      int
}

type ZoneWrapper struct {
	zones []*Zone
	by    func(p, q *Zone) bool
}

func (zw ZoneWrapper) Len() int { // 重写 Len() 方法
	return len(zw.zones)
}
func (zw ZoneWrapper) Swap(i, j int) { // 重写 Swap() 方法
	zw.zones[i], zw.zones[j] = zw.zones[j], zw.zones[i]
}
func (zw ZoneWrapper) Less(i, j int) bool { // 重写 Less() 方法
	return zw.by(zw.zones[i], zw.zones[j])
}

type Drone struct {
	id           int
	playerId     int
	pos          Vec
	prevPos 	 Vec
	speed        Vec
	expectedDest *Zone
	turns2dest   int
}

func (drone *Drone) turns2(vec Vec) int {
	return drone.pos.turns2(vec)
}

func (drone *Drone) turns2Zone(zone *Zone) int {
	return drone.pos.turns2Zone(zone)
}

type DroneWrapper struct {
	drones []*Drone
	by     func(p, q *Drone) bool
}

func (dw DroneWrapper) Len() int { // 重写 Len() 方法
	return len(dw.drones)
}
func (dw DroneWrapper) Swap(i, j int) { // 重写 Swap() 方法
	dw.drones[i], dw.drones[j] = dw.drones[j], dw.drones[i]
}
func (dw DroneWrapper) Less(i, j int) bool { // 重写 Less() 方法
	return dw.by(dw.drones[i], dw.drones[j])
}

type Player struct {
	id         int
	score      int
	drones     []*Drone
	zones      []*Zone
	zoneCenter Vec
}

type Task struct {
	droneId int
	pos     Vec
}

func (task Task) do() {
	fmt.Fprintln(os.Stderr, "Move drone:", task.droneId, "to:", task.pos)
	fmt.Printf("%d %d\n", task.pos.x, task.pos.y)
}

type Objective struct {
	zone         *Zone
	radius       int
	nNeeded      int
	candidates   []int
	value        int
	depends      []*Objective
	done         bool
	nContractors int
}

type ObjectiveWrapper struct {
	objectives []Objective
	by         func(p, q Objective) bool
}

func (dw ObjectiveWrapper) Len() int { // 重写 Len() 方法
	return len(dw.objectives)
}
func (dw ObjectiveWrapper) Swap(i, j int) { // 重写 Swap() 方法
	dw.objectives[i], dw.objectives[j] = dw.objectives[j], dw.objectives[i]
}
func (dw ObjectiveWrapper) Less(i, j int) bool { // 重写 Less() 方法
	return dw.by(dw.objectives[i], dw.objectives[j])
}

type void struct{}

var none void = void{}

type Set struct {
}

var mapSize = Vec{4000, 1800}
var nPlayers, myID, nDrones, nZones int
var players []*Player
var zones []*Zone

const PI float64 = 3.14

func flipCoin() bool {
	i := rand.Intn(1)
	if i == 1 {
		return true
	}
	return false
}

func intersect(c1 Vec, r1_turns float64, c2 Vec, r2_turns float64, inter map[Vec]void) {
	c2toc1 := c1.minus(c2)
	r1 := (r1_turns)*100 - 1
	r2 := (r2_turns)*100 - 1
	dist := norm(c2toc1)
	if dist == 0 {
		return
	}
	if dist > r1+r2 {
		return
	}
	if dist+r2 < r1 {
		return
	}
	if dist+r1 < r2 {
		return
	}
	frac := ((r2*r2-r1*r1)/dist/dist + 1) / 2
	h := math.Sqrt(r2*r2 - frac*dist*frac*dist)
	m := c2.add(c2toc1.multiply(frac))
	inter1 := m.add(multiply(h, Vec{c2toc1.y, -c2toc1.x}).divide(dist))
	inter2 := m.add(multiply(h, Vec{-c2toc1.y, c2toc1.x}).divide(dist))
	if inter1.x > 0 && inter1.y > 0 && inter1.x < mapSize.x && inter1.y < mapSize.y {
		inter[inter1] = none
	}
	if inter2.x > 0 && inter2.y > 0 && inter2.x < mapSize.x && inter2.y < mapSize.y {
		inter[inter2] = none
	}
	fmt.Fprintln(os.Stderr, "inter:", inter)
}

type DecisionContext struct {
	taskPerDrone []Task
	interSet     []map[Vec]void
	contracts    [][]*Objective
}

func getDroneById(droneId int, drones []*Drone) *Drone {
	for _, drone := range drones {
		if drone.id == droneId {
			return drone
		}
	}
	return nil
}

func getMyDroneById(id int) *Drone {
	for _, drone := range myself().drones {
		if nil != drone && drone.id == id {
			return drone
		}
	}
	return nil
}

func (dc DecisionContext) addObjective(obj *Objective) bool {
	if obj.done {
		return true
	}
	var tmpTasks []Task
	tmpInterSet := make([]map[Vec]void, nDrones)
	for i, _ := range tmpInterSet {
		tmpInterSet[i] = make(map[Vec]void)
	}
	savedContracts := make([][]*Objective, nDrones)
	for _, di := range obj.candidates {
		d := getMyDroneById(di)
		if len(dc.contracts[di]) != 0 {
			interOthers := dc.interSet[di]
			toDelete := make([]Vec, 0)
			for i, _ := range interOthers {
				if i.turns2Zone(obj.zone) > obj.radius {
					// delete(interOthers, i)
					toDelete = append(toDelete, i)
				}
			}
			for _, vec := range toDelete {
				delete(interOthers, vec)
			}
			if len(interOthers) == 0 { goto MergeFailed }
			{
				tmpInterSet[di] = interOthers
				interNew := make(map[Vec]void)
				intersect(d.pos, 1, obj.zone.pos, float64(obj.radius+1), interNew)
				for _, c := range dc.contracts[di] {
					intersect(c.zone.pos, float64(c.radius+1), obj.zone.pos, float64(obj.radius+1), interNew)
				}
				toDelete = make([]Vec, 0)
				for i, _ := range interNew {
					if norm(i.minus(d.pos)) > 100 {
						// delete(interNew, i)
						toDelete = append(toDelete, i)
					}
				}
				for _, vec := range toDelete {
					delete(interNew, vec)
				}
				toDelete = make([]Vec, 0)
				for _, c := range dc.contracts[di] {
					for i, _ := range interNew {
						if i.turns2Zone(c.zone) > c.radius {
							// delete(interNew, i)
							toDelete = append(toDelete, i)
						}
					}
				}
				for _, vec := range toDelete {
					delete(interNew, vec)
				}
				for i, v := range interNew {
					tmpInterSet[di][i] = v
				}

				m := Vec{0, 0}
				for i, _ := range tmpInterSet[di] {
					m = m.add(i)
				}
				m = m.divide(float64(len(tmpInterSet[di])))

				allOk := true
				if d.turns2(m) > 1 {
					fmt.Fprintln(os.Stderr, "Merging contracts failed when adding (", obj.zone.id, "dist", obj.radius, ") (point too far)")
					for i, _ := range tmpInterSet[di] {
						fmt.Fprintln(os.Stderr, i.x, i.y, ":", d.turns2(i))
					}
					allOk = false
				}
				for _, c := range dc.contracts[di] {
					if m.turns2Zone(c.zone) > c.radius {
						fmt.Fprintln(os.Stderr, "Merging contracts failed when adding (", obj.zone.id, "dist", obj.radius, ") (zone", c.zone.id, ", dist", c.radius, ")")
						for i, _ := range tmpInterSet[di] {
							fmt.Fprintln(os.Stderr, i.x, i.y, ":", i.turns2Zone(c.zone))
						}
						allOk = false
					}
				}
				if allOk {
					tmpTasks = append(tmpTasks, Task{di, m})
					continue
				}
			}
		MergeFailed:
			if len(tmpTasks) >= obj.nNeeded {
				continue
			}
			available := true
			for _, c := range dc.contracts[di] {
				if c.nContractors == c.nNeeded {
					available = false
					break
				}
			}
			if !available {
				continue
			}
			for _, c := range dc.contracts[di] {
				c.nContractors--
			}
			savedContracts[di] = dc.contracts[di]
			dc.contracts[di] = make([]*Objective, 0)
			tmpInterSet[di] = make(map[Vec]void)
		}

		if len(dc.contracts[di]) == 0 {
			intersect(d.pos, 1, obj.zone.pos, float64(obj.radius+1), tmpInterSet[di])
			if len(tmpInterSet[di]) == 0 {
				var i float64
				for i = 0; i < 6; i++ {
					x := int(100 * math.Cos(2*PI*i/6))
					y := int(100 * math.Sin(2*PI*i/6))
					v := Vec{x, y}
					vec := d.pos.add(v)
					tmpInterSet[di][vec] = none
				}
			}
			tmpTasks = append(tmpTasks, Task{di, obj.zone.pos})
		}
	}
	if len(tmpTasks) >= obj.nNeeded {
		for _, t := range tmpTasks {
			dc.taskPerDrone[t.droneId] = t
			dc.interSet[t.droneId] = tmpInterSet[t.droneId]
			dc.contracts[t.droneId] = append(dc.contracts[t.droneId], obj)
			fmt.Fprintln(os.Stderr, "Assigned drone", t.droneId, "to zone: id =", obj.zone.id, ", dist =", obj.radius, " [ next pos =", t.pos, "]")
		}
		obj.nContractors = len(tmpTasks)
		return true
	}
	for di := 0; di < nDrones; di++ {
		for _, c := range savedContracts[di] {
			dc.contracts[di] = append(dc.contracts[di], c)
			c.nContractors++
		}
	}
	return false
}

func getAnotherPlayer() *Player {
	for i, player := range players {
		if i != myID {
			return player
		}
	}
	return nil
}

func getPlayerById(id int) *Player {
	for i, player := range players {
		if i == id {
			return player
		}
	}
	return nil
}

func myself() *Player {
	return getPlayerById(myID)
}

var isInitialized bool

type Target struct {
	drone *Drone
	zone *Zone
	turns2 int
}

var targets []Target

func (drone *Drone) getClosestZone() *Zone {
	min := zones[0]
	turns := drone.turns2Zone(zones[0])
	fmt.Fprint(os.Stderr, "drone: ", drone.id, " ", drone.pos, " : ")
	for _, zone := range zones {
		tmp := drone.turns2Zone(zone)
		if tmp < turns {
			min = zone
			turns = tmp
		}
		fmt.Fprint(os.Stderr, tmp, " ")
	}
	fmt.Fprintln(os.Stderr, "min turns: ", turns)
	return min
}

func (drone *Drone) moveToZone(zone *Zone) {
	fmt.Fprintln(os.Stderr, "drone [", drone.id, "] move to zone [", zone.id, "]:", zone.pos)
	fmt.Printf("%d %d\n", zone.pos.x, zone.pos.y)
}

func main() {
	// P: number of players in the game (2 to 4 players)
	// ID: ID of your player (0, 1, 2, or 3)
	// D: number of drones in each team (3 to 11)
	// Z: number of zones on the map (4 to 8)
	fmt.Scan(&nPlayers, &myID, &nDrones, &nZones)
	rand.Seed(int64(myID) * 27)

	fmt.Fprintln(os.Stderr, "nPlayers =", nPlayers, ", myID =", myID, ", nDrones =", nDrones, ", nZones =", nZones)

	players = make([]*Player, nPlayers)
	for i := 0; i < nPlayers; i++ {
		player := Player{}
		player.drones = make([]*Drone, nDrones)
		player.zones = make([]*Zone, nZones)
		player.score = 0
		player.zoneCenter = Vec{0, 0}
		players[i] = &player
	}
	zones = make([]*Zone, nZones)

	for i := 0; i < nZones; i++ {
		// X: corresponds to the position of the center of a zone. A zone is a circle with a radius of 100 units.
		var X, Y int
		fmt.Scan(&X, &Y)
		zones[i] = &Zone{id: i, pos: Vec{X, Y}, ownerId: -1}
	}

	for {
		for _, player := range players {
			if nil != player {
				player.zones = make([]*Zone, 0)
			}
		}

		for i := 0; i < nZones; i++ {
			// TID: ID of the team controlling the zone (0, 1, 2, or 3) or -1 if it is not controlled.
			// The zones are given in the same order as in the initialization.
			var TID int
			fmt.Scan(&TID)
			zone := zones[i]
			zone.ownerId = TID
			if zone.ownerId != -1 {
				player := getPlayerById(zone.ownerId)
				if nil != player {
					player.score++
					player.zones = append(player.zones, zone)
				}
			}
		}

		for i := 0; i < nPlayers; i++ {
			player := players[i]
			player.id = i

			for j := 0; j < nDrones; j++ {
				// DX: The first D lines contain the coordinates of drones of a player with the ID 0,
				// the following D lines those of the drones of player 1, and thus it continues until the last player.
				var DX, DY int
				fmt.Scan(&DX, &DY)

				drone := player.drones[j]
				if nil == drone {
					drone = &Drone{pos: Vec{-1, -1}}
					player.drones[j] = drone
				}
				drone.id = j
				drone.playerId = player.id
				prevPos := drone.pos
				drone.prevPos = prevPos
				drone.pos = Vec{DX, DY}
				if prevPos.x < 0 {
					continue
				}

				drone.speed = drone.pos.minus(prevPos)
				drone.expectedDest = nil
				for _, zone := range zones {
					a := norm(drone.speed)
					b := math.Abs(float64(det(zone.pos.minus(drone.pos), drone.speed.divide(norm(drone.speed)))))
					c := dot(zone.pos.minus(drone.pos), drone.speed)
					if a > 70 && b < 100 && c > 0 {
						dist := drone.turns2Zone(zone)
						if drone.expectedDest != nil && drone.turns2dest < dist {
							continue
						}
						if drone.expectedDest != nil {
							// changed!
						}
						drone.expectedDest = zone
						drone.turns2dest = dist
					}
				}
			}
			if len(player.zones) == 0 {
				player.zoneCenter = Vec{mapSize.x / 2, mapSize.y / 2}
			} else {
				player.zoneCenter = Vec{0, 0}
				for _, zone := range player.zones {
					player.zoneCenter = player.zoneCenter.add(zone.pos)
				}
				player.zoneCenter = player.zoneCenter.divide(float64(len(player.zones)))
			}
			fmt.Fprintln(os.Stderr, " Player", i, ": id =", player.id, ", score =", player.score, ", center =", player.zoneCenter, ", zones =", len(player.zones))
		}

		if !isInitialized {
			for _, drone := range myself().drones {
				zone := drone.getClosestZone()
				target := Target {drone: drone, zone: zone}
				fmt.Fprintln(os.Stderr, "Drone ", drone.id, ": pos", drone.pos, " close to zone", zone.id, ": pos", zone.pos, ": turns", drone.turns2Zone(zone))
				targets = append(targets, target)
			}
			isInitialized = true
		}

		for _, target := range targets {
			target.drone.moveToZone(target.zone)
		}
	}
}
