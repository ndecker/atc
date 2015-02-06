package main

import (
	"fmt"
	"strconv"
	"strings"
)

var _ = fmt.Printf

const (
	DEFAULT_BOARD = `
        .....1....2.........3....
        .........................
        .........................
        .........................
        ..............+..........
        4..............%.........
        .........................
        .........................
        .........................
        .........................
        0....*...+=.........*...9
        .........................
        .........................
        .........................
        ........................7
        .........................
        .........................
        .........................
        .........................
        .....8........5.....6....
    `

	// Entry Fix  Initial Heading  Exit Fix
	// 0          E                9
	// 1          S                8
	// 2          SE               7
	// 3          S                6
	// 4          SE               5
	// 5          NW               4
	// 6          N                3
	// 7          NW               2
	// 8          N                1
	// 9          W                0

	// TODO: =-= %-% nur Prop
	DEFAULT_ROUTES = `
        4: 0-9 9-0
        4: 4-5 5-4
        4: 1-8 8-1
        4: 2-7 7-2
        4: 3-6 6-3
        1: 0-= 1-= 2-= 3-= 4-= 5-= 6-= 7-= 8-= 9-=
        1: 0-% 1-% 2-% 3-% 4-% 5-% 6-% 7-% 8-% 9-%
        1: =-0 =-1 =-2 =-3 =-4 =-5 =-6 =-7 =-8 =-9
        1: %-0 %-1 %-2 %-3 %-4 %-5 %-6 %-7 %-8 %-9
        1: =-= %-%
    `
)

type Board struct {
	width  int
	height int

	entrypoints map[rune]*EntryPoint
	beacons     []Beacon // TODO: rename to navaid
	routes      []Route
}

func (b Board) Contains(p Position) bool {
	if p.x < 0 || p.y < 0 {
		return false
	}
	if p.x > b.width-1 || p.y > b.height-1 {
		return false
	}
	return true
}

func (b Board) String() string {
	return fmt.Sprintf(`
        w/h: %d/%d
        entrypoints: %v
        beacons:     %v
        routes:      %v`,
		b.width, b.height, b.entrypoints, b.beacons, b.routes)

}

func (b *Board) GetBeacon(p Position) *Beacon {
	for _, beacon := range b.beacons {
		if beacon.Position == p {
			return &beacon
		}
	}
	return nil
}

func (b *Board) GetEntryPoint(p Position) *EntryPoint {
	for _, ep := range b.entrypoints {
		if ep.Position == p {
			return ep
		}
	}
	return nil
}

const (
	TypeRoute   = iota
	TypeAirport = iota
)

type EntryPoint struct {
	class int

	sign rune
	Position
	Direction
}

type Beacon struct {
	Position
	airports map[rune]Direction
}

type Route struct {
	entry rune
	exit  rune
}

func (r Route) String() string {
	return fmt.Sprintf("%s-%s", string(r.entry), string(r.exit))
}

func ParseBoard(s string, rs string) *Board {
	b := &Board{
		entrypoints: make(map[rune]*EntryPoint),
		beacons:     make([]Beacon, 0),
		routes:      make([]Route, 0),
	}

	lines := make([]string, 0, 40)

	for _, l := range strings.Split(s, "\n") {
		l = strings.Trim(l, " \r\n")
		if len(l) > 0 {
			lines = append(lines, l)

			if b.width == 0 {
				b.width = len(l)
			} else if b.width != len(l) {
				fmt.Println(l, len(l), b.width)
				panic("inconsistent width")
			}
		}
	}

	b.height = len(lines)
	if b.height <= 0 {
		panic("board has no height")
	}

	for x := 0; x < b.width; x += 1 {
		for y := 0; y < b.height; y += 1 {
			pos := Position{x: x, y: y}

			ch := lines[y][x]
			switch ch {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				b.entrypoints[rune(ch)] = &EntryPoint{
					class:    TypeRoute,
					sign:     rune(ch),
					Position: pos,
				}
			case '%', '=':
				// find direction for airport
				var dir Direction
				for _, d := range DIRECTIONS {
					pos2 := pos.Move(d, 1)
					if lines[pos2.y][pos2.x] == '+' {
						dir = d
					}
				}
				b.entrypoints[rune(ch)] = &EntryPoint{
					class:     TypeAirport,
					sign:      rune(ch),
					Position:  pos,
					Direction: dir,
				}
			case '+':
				// direction marker for Airport
			case '*':
				b.beacons = append(b.beacons, Beacon{
					pos,
					make(map[rune]Direction),
				})
			case '.':
			default:
				panic("unknown spec: " + string(ch))
			}
		}
	}

	// find airports for beacons
	for _, beacon := range b.beacons {
		for sign, ep := range b.entrypoints {
			if ep.class == TypeAirport {
				dir, valid := beacon.Position.Direction(ep.Position)
				if valid {
					beacon.airports[sign] = dir
				}
			}
		}
	}

	// TODO: better parsing
	for _, l := range strings.Split(rs, "\n") {
		l = strings.Trim(l, " \r\n")
		if l == "" {
			continue
		}

		parts := strings.SplitN(l, ":", 2)
		weight, _ := strconv.Atoi(parts[0])
		routes := strings.Split(parts[1], " ")
		for _, r := range routes {
			if r == "" {
				continue
			}

			route := Route{
				entry: rune(r[0]),
				exit:  rune(r[2]),
			}
			_, ok_entry := b.entrypoints[route.entry]
			_, ok_exit := b.entrypoints[route.exit]
			if !ok_entry || !ok_exit {
				panic("unknown entrypoint: " + string(route.entry) + " or " + string(route.exit))
			}
			for n := 0; n < weight; n += 1 {
				b.routes = append(b.routes, route)
			}

		}

	}

	return b
}
