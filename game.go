package main

import (
	"fmt"
)

type GameSetup struct {
	duration         Ticks
	last_plane_start Ticks

	num_planes int

	skip_to_next_tick bool // if true "," will skip to the beginning of the next tick
	have_jet          bool
	have_prop         bool
	have_heli         bool
}

var DEFAULT_SETUP = GameSetup{
	duration:         25 * Minutes,
	last_plane_start: 15 * Minutes,
	num_planes:       26,

	skip_to_next_tick: true,
	have_jet:          true,
	have_prop:         true,
	have_heli:         true,
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
	board := ParseBoard(DEFAULT_BOARD, DEFAULT_ROUTES)

	var game = &GameState{
		seed:  seed,
		setup: setup,
		board: board,

		clock:  setup.duration,
		planes: MakePlanes(setup, board, seed),
	}
	game.ci.game = game

	return game
}
