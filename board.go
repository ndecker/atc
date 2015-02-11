package main

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	DEFAULT_BOARD *Board = ParseBoard("ATC Standard", `
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
    `, `
    # Format: weight: entry-exit-direction
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
    `, []Difficulty{
		Difficulty{"normal", 50 * Minutes, 26},
	})

	CROSSWAYS_BOARD *Board = ParseBoard("Crossways", `
        .....4.........6.........8.....
        ...............................
        ...............................
        ...............................
        ...............................
        ...............................
        0.............................1
        ...............................
        ...............................
        ...............................
        ....+=.........*...............
        ...............................
        ...............................
        ...............................
        2.............................3
        ...............................
        ...............................
        ...............................
        ...............................
        ...............................
        .....5.........7.........9.....
    `, `
        4: 0-1-E 3-2-W
        2: 4-5-S 5-4-N
        2: 6-7-S 7-6-N
        2: 8-9-S 9-8-N
        2: 4-9-SE 9-4-NW
        2: 8-5-SW 5-8-NE

        2: 0-=-E  3-=-W
        2: =-0-W  =-3-W
        1: 4-=-SE 5-=-NE 6-=-S  7-=-N  8-=-SW  9-=-NW
        1: =-4-W  =-5-W  =-6-W  =-7-W  =-8-W   =-9-W
    `, []Difficulty{
		Difficulty{"extreme", 90 * Minutes, 100},
	})

	NOFLY_BOARD *Board = ParseBoard("NoFly Zone", `
        ............5....6.............
        ...............................
        ...............................
        ...............................
        ...............................
        ...............................
        .......xxxxxxxxxxxxxxxxx.......
        .......xxxxxxxxxxxxxxxxx.......
        1......xxxxxxxxxxxxxxxxx......3
        .......xxxxxxxxxxxxxxxxx.......
        .......xxxxxxxxxxxxxxxxx.......
        2......xxxxxxxxxxxxxxxxx......4
        .......xxxxxxxxxxxxxxxxx.......
        .......xxxxxxxxxxxxxxxxx.......
        ...............................
        ...............................
        ...............................
        ...............................
        ...............................
        ...............................
        ............7..8...............
    `, `
        1: 1-3-E 4-2-W
        1: 5-7-S 8-6-N
    `, []Difficulty{})

	BOARDS []*Board = []*Board{
		DEFAULT_BOARD, CROSSWAYS_BOARD, NOFLY_BOARD,
	}
)

type Board struct {
	name string

	width  int
	height int

	entrypoints map[rune]*EntryPoint
	navaids     []Position
	routes      []Route
	nofly       []Position

	difficulties []Difficulty
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

func (b *Board) GetNavaid(p Position) *Position {
	for _, navaid := range b.navaids {
		if navaid == p {
			return &navaid
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

type Route struct {
	entry rune
	exit  rune
	Direction
	weight int
}

func (r Route) String() string {
	return fmt.Sprintf("%s-%s", string(r.entry), string(r.exit))
}

func ParseBoard(name string, s string, rs string, df []Difficulty) *Board {
	b := &Board{
		name:         name,
		entrypoints:  make(map[rune]*EntryPoint),
		navaids:      make([]Position, 0),
		routes:       make([]Route, 0),
		nofly:        make([]Position, 0),
		difficulties: df,
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
				b.navaids = append(b.navaids, pos)
			case 'x':
				b.nofly = append(b.nofly, pos)
			case '.':
			default:
				panic("unknown spec: " + string(ch))
			}
		}
	}

	// TODO: better parsing
	for _, l := range strings.Split(rs, "\n") {
		l = strings.Trim(l, " \r\n")
		if l == "" || l[0] == '#' {
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
