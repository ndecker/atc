package main

type GameSetup struct {
	duration         Ticks
	last_plane_start Ticks

	num_planes int

	skip_to_next_tick bool // if true "," will skip to the beginning of the next tick
	delayed_commands  bool

	have_jet       bool
	have_prop      bool
	have_heli      bool
	have_blackbird bool

	show_pending_planes bool
}

func DefaultSetup() *GameSetup {
	return &GameSetup{
		duration:         25 * Minutes,
		last_plane_start: 15 * Minutes,
		num_planes:       30,

		skip_to_next_tick: true,
		delayed_commands:  true,

		have_jet:       true,
		have_prop:      true,
		have_heli:      true,
		have_blackbird: true,

		show_pending_planes: false,
	}
}

type EndReason struct {
	message string
	planes  []*Plane
}

type GameState struct {
	setup *GameSetup
	board *Board

	seed int64

	clock      Ticks
	end_reason *EndReason

	ci CommandInterpreter

	planes             []*Plane
	reusable_callsigns []rune
}

func (g *GameState) Tick() {
	if g.end_reason == nil {
		g.end_reason = g.doTick()
	}
}

func (g *GameState) doTick() *EndReason {
	g.clock.Tick()
	if g.clock == 0 {
		return &EndReason{message: "Time is up"}
	}

	// TODO: update once before first tick
	remaining := 0
	for _, p := range g.planes {
		er := p.Tick(g)
		if er != nil {
			return er
		}

		if p.callsign == 0 && (p.state == StateIncoming || p.state == StateWaiting) {
			if len(g.reusable_callsigns) == 0 {
				return &EndReason{message: "Too many active planes"}
			}
			p.callsign = g.reusable_callsigns[0]
			g.reusable_callsigns = g.reusable_callsigns[1:]
		}

		if !p.IsDone() {
			remaining += 1
		} else if p.callsign != 0 && g.setup.num_planes > 26 {
			// plane is done; reuse callsign
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
				return &EndReason{
					message: "Conflict",
					planes:  []*Plane{p1, p2},
				}
			}
		}
	}

	if remaining == 0 {
		return &EndReason{message: "Success"}
	}

	// apply delayed commands
	g.ci.Tick(g)
	return nil
}

func (g *GameState) KeyPressed(k rune) {
	if g.end_reason != nil {
		return
	}
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

func NewGame(setup *GameSetup, board *Board, seed int64) *GameState {

	slowest_plane_ticks := 1
	for _, pt := range PlaneTypes(setup) {
		slowest_plane_ticks = Max(slowest_plane_ticks, int(pt.ticks_per_move))
	}

	// allow the slowest plane to cross the board
	setup.last_plane_start = Ticks(Max(
		int(setup.last_plane_start),
		board.width*slowest_plane_ticks))

	planes := MakePlanes(setup, board, seed)

	var game = &GameState{
		seed:  seed,
		setup: setup,
		board: board,

		clock:  setup.duration,
		planes: planes,
	}
	return game
}
