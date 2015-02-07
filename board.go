package main

import (
	"fmt"
	"strconv"
	"strings"
)

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

	// Format: weight: entry-exit-direction
	DEFAULT_ROUTES = `
        6: 0-9-E  9-0-W
        6: 1-8-S  8-1-N
        6: 2-7-SE 7-2-NW
        6: 3-6-S  6-3-N
        6: 4-5-SE 5-4-NW

        1: 0-=-E  1-=-S  2-=-SE 3-=-S  4-=-SE 5-=-NW 6-=-N  7-=-NW 8-=-N  9-=-W
        1: 0-%-E  1-%-S  2-%-SE 3-%-S  4-%-SE 5-%-NW 6-%-N  7-%-NW 8-%-N  9-%-W
        1: =-0-W  =-1-W  =-2-W  =-3-W  =-4-W  =-5-W  =-6-W  =-7-W  =-8-W  =-9-W
        1: %-0-NW %-1-NW %-2-NW %-3-NW %-4-NW %-5-NW %-6-NW %-7-NW %-8-NW %-9-NW
        2: =-=-W  %-%-NW =-%-W %-=-NW
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

type EntryPoint struct {
	sign rune
	Position
	Direction
	is_airport bool
}

type Beacon struct {
	Position
}

type Route struct {
	entry rune
	exit  rune
	Direction
	weight int
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
					sign:       rune(ch),
					Position:   pos,
					is_airport: false,
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
					sign:       rune(ch),
					Position:   pos,
					Direction:  dir,
					is_airport: true,
				}
			case '+':
				// direction marker for Airport
			case '*':
				b.beacons = append(b.beacons, Beacon{pos})
			case '.':
			default:
				panic("unknown spec: " + string(ch))
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

			r_parts := strings.SplitN(r, "-", 3)

			route := Route{
				entry:     rune(r_parts[0][0]),
				exit:      rune(r_parts[1][0]),
				Direction: ParseDirection(r_parts[2]),
				weight:    weight,
			}
			_, ok_entry := b.entrypoints[route.entry]
			_, ok_exit := b.entrypoints[route.exit]
			if !ok_entry || !ok_exit {
				panic("unknown entrypoint: " + string(route.entry) + " or " + string(route.exit))
			}
			b.routes = append(b.routes, route)

		}

	}

	return b
}
