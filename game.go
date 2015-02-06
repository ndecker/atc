package main

import (
	"fmt"
	"math/rand"
	"sort"
)

var _ = fmt.Println

var PLANE_NAMES = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

type GameSetup struct {
	duration         Ticks
	last_plane_start Ticks

	num_planes int
}

var DEFAULT_SETUP = GameSetup{
	duration:         25 * Minutes,
	last_plane_start: 15 * Minutes,
	num_planes:       26,
}

type GameState struct {
	setup  GameSetup
	board  *Board

    seed   int64
	random *rand.Rand

	clock      Ticks
	end_reason string

	ci CommandInterpreter

	planes      []*Plane
	plane_names []rune
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
		if p.state != StateAway {
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

func NewGame(setup GameSetup, seed int64) *GameState {
	var gs = &GameState{
		setup:  setup,
		board:  ParseBoard(DEFAULT_BOARD, DEFAULT_ROUTES),
		clock:  setup.duration,
		planes: make([]*Plane, 0),

        seed: seed,
		random: rand.New(rand.NewSource(seed)),
	}
	gs.ci.game = gs

	ps := gs.random.Perm(gs.setup.num_planes)
	for _, p := range ps {
		gs.plane_names = append(gs.plane_names, PLANE_NAMES[p])
	}

	for n := 0; n < gs.setup.num_planes; n += 1 {
		plane := NewPlane(gs,
			gs.board.routes[gs.random.Intn(len(gs.board.routes))],
			gs.random.Intn(9-6+1)+6,
		)

		gs.planes = append(gs.planes, plane)
	}

	sort.Sort(ByTime(gs.planes))
	return gs
}

