package main

type Difficulty struct {
	name       string
	duration   Ticks
	num_planes int
}

type GameRules struct {
	last_plane_start Ticks

	skip_to_next_tick bool // if true "," will skip to the beginning of the next tick
	delayed_commands  bool

	have_jet       bool
	have_prop      bool
	have_heli      bool
	have_blackbird bool

	show_pending_planes bool
}

func ATCDefaultRules() *GameRules {
	return &GameRules{
		last_plane_start: 15 * Minutes,

		skip_to_next_tick: false,
		delayed_commands:  false,

		have_jet:       true,
		have_prop:      true,
		have_heli:      false,
		have_blackbird: false,

		show_pending_planes: false,
	}
}

func DefaultRules() *GameRules {
	return &GameRules{
		last_plane_start: 15 * Minutes,

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
	rules *GameRules
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
		} else if p.callsign != 0 && len(g.planes) > 26 {
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

func NewGame(rules *GameRules, board *Board, diff *Difficulty, seed int64) *GameState {

	slowest_plane_ticks := 1
	for _, pt := range PlaneTypes(rules) {
		slowest_plane_ticks = Max(slowest_plane_ticks, int(pt.ticks_per_move))
	}

	// allow the slowest plane to cross the board
	rules.last_plane_start = Ticks(Max(
		int(rules.last_plane_start),
		board.width*slowest_plane_ticks))

	planes := MakePlanes(rules, board, diff, seed)

	var game = &GameState{
		seed:  seed,
		rules: rules,
		board: board,

		clock:  diff.duration,
		planes: planes,
	}
	return game
}
