package main

import (
	"fmt"
	"math/rand"
	"sort"
)

var _ = fmt.Println

type GameSetup struct {
	duration         Ticks
	last_plane_start Ticks

	num_planes int

	skip_to_next_tick bool // if true "," will skip to the beginning of the next tick
}

var DEFAULT_SETUP = GameSetup{
	duration:         25 * Minutes,
	last_plane_start: 15 * Minutes,
	num_planes:       26,

	skip_to_next_tick: true,
}

type GameState struct {
	setup GameSetup
	board *Board

	seed int64

	clock      Ticks
	end_reason string

	ci CommandInterpreter

	planes               []*Plane
	last_commanded_plane *Plane
}

func (g *GameState) Tick() {
	g.clock.Tick()

	g.last_commanded_plane = nil
	if g.clock == 0 {
		g.end_reason = "Time is up"
	}

	// TODO: update once before first tick
	remaining := 0
	for _, p := range g.planes {
		p.Tick(g)
		if p.state != StateLanded && p.state != StateDeparted {
			remaining += 1
		}
	}

	for _, p1 := range g.planes {
		for _, p2 := range g.planes {
			if p1 == p2 {
				continue
			}
			if !p1.IsFlying() || !p2.IsFlying() {
				continue
			}

			if p1.Collides(p2) {
				g.end_reason = fmt.Sprintf("-- Conflict -- %s %s", p1.Marker(), p2.Marker())
			}
		}
	}

	if remaining == 0 {
		g.end_reason = "Won"
	}
}

func (g *GameState) String() string {
	res := ""
	for _, p := range g.planes {
		res += p.Flightplan() + "\n"
	}
	return res
}

func NewGame(setup GameSetup, seed int64) *GameState {
	r := rand.New(rand.NewSource(seed))

	var game = &GameState{
		seed:  seed,
		setup: setup,
		board: ParseBoard(DEFAULT_BOARD, DEFAULT_ROUTES),

		clock:  setup.duration,
		planes: make([]*Plane, 0, setup.num_planes),
	}
	game.ci.game = game

	callsigns := r.Perm(game.setup.num_planes)
	for _, callsign := range callsigns {
		var plane *Plane
		tries := 0

	retry_plane:
		for {
			if tries > 100 {
				panic("cannot find valid plane")
			}

			typ := &PLANE_TYPES[r.Intn(len(PLANE_TYPES))]

			// try until valid plan found
			route := game.board.routes[r.Intn(len(game.board.routes))]
			height := r.Intn(9-6+1) + 6

			// entries are present. checked in board.go
			entry := game.board.entrypoints[route.entry]
			exit := game.board.entrypoints[route.exit]

			start := Ticks(r.Intn(int(game.setup.duration-game.setup.last_plane_start))) + game.setup.last_plane_start

			plane = &Plane{
				callsign: rune(callsign + 'A'),
				typ:      typ,

				entry:     entry,
				exit:      exit,
				Direction: route.Direction,

				start:     start,
				fuel_left: typ.initial_fuel,

				initial_height: height,

				is_holding:   false,
				is_hoovering: typ.can_hoover && entry.class == TypeAirport,

				hold_at_navaid: exit.class == TypeAirport,
			}

			// no two planes from the same origin share the same altitude<
			for _, other_plane := range game.planes {
				if other_plane.entry == plane.entry &&
					other_plane.entry.class == TypeRoute &&
					other_plane.initial_height == plane.initial_height {
					// retry another plane
					continue retry_plane
				}
			}

			break
		}

		game.planes = append(game.planes, plane)
	}

	sort.Sort(ByTime(game.planes))
	return game
}
