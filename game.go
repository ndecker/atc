package main

import (
	"fmt"
)

type GameSetup struct {
	duration         Ticks
	last_plane_start Ticks

	num_planes int

	skip_to_next_tick bool // if true "," will skip to the beginning of the next tick
	delayed_commands  bool

	have_jet  bool
	have_prop bool
	have_heli bool

	show_planes_at_start bool

	commands [][]Command
}

var DEFAULT_SETUP = GameSetup{
	duration:         60 * Minutes,
	last_plane_start: 15 * Minutes,
	num_planes:       50,

	skip_to_next_tick: true,
	delayed_commands:  true,

	have_jet:  true,
	have_prop: true,
	have_heli: true,

	show_planes_at_start: true,
}

type GameState struct {
	setup GameSetup
	board *Board

	seed int64

	clock      Ticks
	end_reason string

	ci CommandInterpreter

	planes             []*Plane
	reusable_callsigns []rune
}

func (g *GameState) Tick() {
	g.clock.Tick()

	if g.clock == 0 {
		g.end_reason = "Time is up"
	}

	// TODO: update once before first tick
	remaining := 0
	for _, p := range g.planes {
		p.Tick(g)

		if p.callsign == 0 && (p.state == StateIncoming || p.state == StateWaiting) {
			if len(g.reusable_callsigns) == 0 {
				g.end_reason = "Too many active planes"
			}
			p.callsign = g.reusable_callsigns[0]
			g.reusable_callsigns = g.reusable_callsigns[1:]
		}

		if p.state != StateLanded && p.state != StateDeparted {
			remaining += 1
		} else if p.callsign != 0 {
			// plane is done; reusable callsign
			g.reusable_callsigns = append(g.reusable_callsigns, p.callsign)
			p.callsign = 0
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

	// apply delayed commands
	g.ci.Tick(g)
}

func (g *GameState) KeyPressed(k rune) {
	g.ci.KeyPressed(g, k)
}

func (g *GameState) FindPlane(callsign rune) *Plane {
	var plane *Plane
	for _, p := range g.planes {
		if p.callsign == callsign {
			plane = p
			break
		}
	}
	return plane
}

func (g *GameState) String() string {
	res := ""
	for _, p := range g.planes {
		res += p.Flightplan() + "\n"
	}
	return res
}

func NewGame(setup GameSetup, seed int64) *GameState {
	board := ParseBoard(DEFAULT_BOARD, DEFAULT_ROUTES)

	var game = &GameState{
		seed:  seed,
		setup: setup,
		board: board,

		clock:  setup.duration,
		planes: MakePlanes(setup, board, seed),
	}
	game.ci.setup = setup

	return game
}
