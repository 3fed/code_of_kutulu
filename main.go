package main

import (
	"bufio"
	"fmt"
	"os"
)

type grid [][]cell
type cell int

const (
	inputWall  = "#"
	inputSpawn = "w"
	inputEmpty = "."
)

const (
	cellWall  = iota
	cellSpawn = iota
	cellEmpty = iota
)

type coord struct {
	x, y int
}

type minionState int

const (
	stateSpawning  minionState = 0
	stateWandering minionState = 1
)

type explorer struct {
	id    int
	coord coord
}

type wanderer struct {
	id         int
	coord      coord
	state      minionState
	target     int
	recallTime int
}

type spawningMinion struct {
	id        int
	coord     coord
	state     minionState
	target    int
	spawnTime int
}

type loggable interface {
	String() string
}

func (e explorer) String() string {
	return fmt.Sprintf("explorer %d %d %d", e.id, e.coord.x, e.coord.y)
}

func (w wanderer) String() string {
	return fmt.Sprintf("wanderer %d %d %d %d %d %d", w.id, w.coord.x, w.coord.y, w.state, w.target, w.recallTime)
}

func (s spawningMinion) String() string {
	return fmt.Sprintf("spawningMinion %d %d %d %d %d %d", s.id, s.coord.x, s.coord.y, s.state, s.target, s.spawnTime)
}

const (
	entityTypeExplorer = "EXPLORER"
	entityTypeWanderer = "WANDERER"
)

func buildGridOfWalls(width int, height int) grid {
	res := make(grid, height)
	for i := 0; i < height; i++ {
		res[i] = make([]cell, width)
	}
	return res
}

func printGrid(g grid) {
	res := ""
	for _, line := range g {
		for _, cell := range line {
			res += cellToString(cell)
		}
		res += "\n"
	}
	log(res)
}

func log(mes string) {
	fmt.Fprintln(os.Stderr, mes)
}

func cellToString(c cell) string {
	switch c {
	case cellWall:
		return "#"
	case cellSpawn:
		return "w"
	case cellEmpty:
		return "."
	default:
		panic("unrecognized cell " + string(c))
	}
}

func parseCell(c string) cell {
	switch c {
	case inputWall:
		return cellWall
	case inputSpawn:
		return cellSpawn
	case inputEmpty:
		return cellEmpty
	default:
		panic("unrecognized string " + c)
	}
}

func parseGrid(scanner *bufio.Scanner, width int, height int) grid {
	res := buildGridOfWalls(width, height)
	for i := 0; i < height; i++ {
		scanner.Scan()
		line := scanner.Text()
		for j, c := range line {
			res[i][j] = parseCell(string(c))
		}
	}
	return res
}

func send(command string) {
	fmt.Println(command)
}

func sendMove(x, y int) {
	send(fmt.Sprintf("MOVE %d %d", x, y))
}

func sendWait() {
	send("WAIT")
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func dist(from coord, to coord) int {
	return abs(to.x-from.x) + abs(to.y-from.y)
}

func getClosestWanderer(from coord, wanderers []wanderer) wanderer {
	if len(wanderers) == 0 {
		panic("cannot find closest wanderer if there is no wanderer")
	}
	bestIndex := -1
	bestDistance := -1
	for i, w := range wanderers {
		d := dist(w.coord, from)
		if bestDistance == -1 || d < bestDistance {
			bestIndex = i
			bestDistance = d
		}
	}
	return wanderers[bestIndex]
}

func getEmptyCells(g grid) []coord {
	res := make([]coord, 0)
	for i, line := range g {
		for j, cell := range line {
			if cell == cellEmpty {
				res = append(res, coord{j, i})
			}
		}
	}
	return res
}

func getCloseEmptyCells(g grid, from coord) []coord {
	res := make([]coord, 0)
	for i, line := range g {
		for j, cell := range line {
			if cell == cellEmpty && dist(from, coord{j, i}) <= 4 {
				res = append(res, coord{j, i})
			}
		}
	}
	return res
}

func getFarestCoord(from coord, candidates []coord) coord {
	if len(candidates) == 0 {
		panic("no candidates for farest coord")
	}
	bestIndex := -1
	bestDistance := -1
	for i, c := range candidates {
		d := dist(from, c)
		if bestDistance == -1 || d > bestDistance {
			bestIndex = i
			bestDistance = d
		}
	}
	return candidates[bestIndex]
}

func getAwayFromClosestWanderer(g grid, me explorer, wanderers []wanderer) coord {
	closestWanderer := getClosestWanderer(me.coord, wanderers)
	empties := getCloseEmptyCells(g, me.coord)
	return getFarestCoord(closestWanderer.coord, empties)
}

func main() {

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1000000), 1000000)

	var width int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &width)

	var height int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &height)

	currentGrid := parseGrid(scanner, width, height)
	printGrid(currentGrid)

	var sanityLossLonely, sanityLossGroup, wandererSpawnTime, wandererLifeTime int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &sanityLossLonely, &sanityLossGroup, &wandererSpawnTime, &wandererLifeTime)

	for {
		var entityCount int
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &entityCount)

		explorers := make([]explorer, 0)
		wanderers := make([]wanderer, 0)
		spawningMinions := make([]spawningMinion, 0)

		for i := 0; i < entityCount; i++ {
			var entityType string
			var id, x, y, param0, param1, param2 int
			scanner.Scan()
			fmt.Sscan(scanner.Text(), &entityType, &id, &x, &y, &param0, &param1, &param2)

			switch entityType {
			case entityTypeExplorer:
				explorers = append(explorers, explorer{id, coord{x, y}})
			case entityTypeWanderer:
				state := minionState(param1)
				switch state {
				case stateSpawning:
					spawningMinions = append(spawningMinions, spawningMinion{id, coord{x, y}, stateSpawning, param2, param0})
				case stateWandering:
					wanderers = append(wanderers, wanderer{id, coord{x, y}, stateWandering, param2, param0})
				default:
					panic("unrecognized state " + string(state))
				}
			default:
				panic("unrecognized entityType " + string(entityType))
			}
		}

		for _, e := range explorers {
			log(e.String())
		}

		for _, w := range wanderers {
			log(w.String())
		}

		for _, s := range spawningMinions {
			log(s.String())
		}

		myExplorer := explorers[0]

		log("Me :")
		log(myExplorer.String())

		if len(wanderers) > 0 {
			away := getAwayFromClosestWanderer(currentGrid, myExplorer, wanderers)
			sendMove(away.x, away.y)
		} else {
			sendWait()
		}
	}
}
