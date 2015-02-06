package main

import (
	"fmt"
)

type PlaneState int

const (
	StatePending  = 0
	StateIncoming = iota
	StateWaiting  = iota
	StateRolling  = iota
	StateFlying   = iota
	StateAproach  = iota

	StateLanded   = iota
	StateDeparted = iota
)

const SAFE_DISTANCE = 3
const FUEL_INDICATOR = 10 * Minutes

type Plane struct {
	callsign rune
	typ      *PlaneType

	entry *EntryPoint
	exit  *EntryPoint

	start      Ticks
	state      PlaneState
	wait_ticks Ticks

	fuel_left Ticks

	Position
	is_hoovering bool

	Direction
	want_turn int

	height         int
	want_height    int
	last_height    int
	initial_height int

	hold_at_navaid      bool
	is_holding          bool
	direction_at_navaid rune
}

func (p *Plane) Tick(game *GameState) {

	if p.IsConsumingFuel() {
		p.fuel_left -= 1
		if p.fuel_left == 0 {
			game.end_reason = fmt.Sprintf("fuel exhausted %s", p.Marker())
			return
		}
	}

	if p.wait_ticks > 0 {
		p.wait_ticks -= 1
		return
	}

	p.last_height = p.height

	switch p.state {
	case StatePending:
		// Wait until visible
		if (game.clock - p.start) < p.typ.ticks_pending {
			switch p.entry.class {
			case TypeRoute:
				p.state = StateIncoming
				p.wait_ticks = p.typ.ticks_pending
				p.height = p.initial_height

				var dir Direction
				var valid bool
				switch p.exit.class {
				case TypeRoute:
					dir, valid = p.entry.Position.Direction(p.exit.Position)
					if !valid {
						panic("invalid direction")
					}
				case TypeAirport:
					// find valid beacon
					for _, b := range game.board.beacons {
						dir, valid = p.entry.Position.Direction(b.Position)
						if valid {
							break
						}
					}
				}
				p.Direction = dir

			case TypeAirport:
				p.state = StateWaiting

				p.height = 0
				p.Direction = p.entry.Direction
			}

			p.want_height = p.height
			p.Position = p.entry.Position
		}
	case StateIncoming: // after wait ticks
		p.state = StateFlying

	case StateWaiting: // wait for DoHeight
	case StateRolling: // after wait ticks
		p.UpdatePosition(game)
		p.ApplyWants()

		p.state = StateFlying
	case StateFlying, StateAproach:
		beacon := game.board.GetBeacon(p.Position)
		if beacon != nil {
			if p.hold_at_navaid {
				p.is_holding = true
			}

			if ep, ok := game.board.entrypoints[p.direction_at_navaid]; ok {
				// always use the direction of the airport.
				p.Direction = ep.Direction
			}
		}

		if p.is_holding {
			p.Direction = p.Direction.Left(1)
		}

		p.UpdatePosition(game)
		p.ApplyWants()

		p.wait_ticks = p.typ.ticks_per_move - 1
	case StateDeparted, StateLanded:
	default:
		panic("unhandled case")
	}

}

func (p *Plane) Collides(p2 *Plane) bool {
	height_match := false

	if p.height == p2.height {
		// same height
		height_match = true
	} else if p.height == p2.last_height && p.last_height == p2.height {
		// crossover
		height_match = true
	}

	if !height_match {
		return false
	}

	distance := p.Position.Distance(p2.Position)
	return distance < SAFE_DISTANCE
}

func (p *Plane) ApplyWants() {
	if p.want_turn > 0 {
		p.Direction = p.Direction.Right(1)
		p.want_turn -= 1
	}
	if p.want_turn < 0 {
		p.Direction = p.Direction.Left(1)
		p.want_turn += 1
	}

	if p.want_height > p.height {
		p.height += 1
	}
	if p.want_height < p.height {
		p.height -= 1
	}
}

func (p *Plane) UpdatePosition(game *GameState) {
	var next_pos Position
	if !p.is_hoovering {
		next_pos = p.Position.Move(p.Direction, 1)
	} else {
		next_pos = p.Position
	}

	if !game.board.Contains(next_pos) {
		// left the playing field
		if p.Position != p.exit.Position {
			game.end_reason = fmt.Sprintf("Boundary Error -- %c%d", p.callsign, p.height)
			return
		}
		if p.height != 5 {
			game.end_reason = fmt.Sprintf("Boundary Error -- %c%d", p.callsign, p.height)
			return
		}
		p.state = StateDeparted
		return
	}

	if p.state == StateAproach {
		ap := game.board.GetEntryPoint(next_pos)
		if ap != nil {
			if ap == p.exit && p.height == 0 {
				p.state = StateLanded
				return
			}

			// call off landing
			if !p.is_hoovering {
				// if hoovering over the airport do not reset to flying
				p.state = StateFlying
				p.height = 1
			}
		}
	}
	p.Position = next_pos
}

func (p *Plane) DoTurn(c int) bool {
	if c < -4 || c > 4 {
		return false
	}

	if p.typ.immediate_turn {
		p.Direction = p.Direction.Right(c)
		p.want_turn = 0
	} else {
		p.want_turn = c
	}
	p.is_holding = false
	p.hold_at_navaid = false
	p.direction_at_navaid = 0
	return true
}

func (p *Plane) DoHeight(h int) bool {
	if h > 5 || h < 0 {
		return false
	}

	if h == 0 {
		// aproach
		if p.state != StateFlying {
			return false
		}
		p.state = StateAproach
		p.want_height = h
		return true
	}

	p.want_height = h

	if p.state == StateWaiting {
		p.state = StateRolling
		p.wait_ticks = p.typ.ticks_rolling
	}
	return true
}

func (p *Plane) DoHold() bool {
	p.hold_at_navaid = true
	p.direction_at_navaid = 0
	return true
}

func (p *Plane) DoKeep() bool {
	if !p.typ.can_hoover {
		return false
	}
	p.is_hoovering = !p.is_hoovering
	return true
}

func (p *Plane) TurnAtNavaid(navaid rune) bool {
	p.direction_at_navaid = navaid
	p.hold_at_navaid = false
	return true
}

func (p Plane) AcceptsCommands() bool {
	return p.state == StateWaiting || p.state == StateRolling || p.state == StateFlying
}
func (p Plane) IsVisible() bool {
	return p.state == StateIncoming || p.state == StateWaiting
}
func (p Plane) IsActive() bool {
	return p.state == StateIncoming ||
		p.state == StateWaiting ||
		p.state == StateRolling ||
		p.state == StateFlying ||
		p.state == StateAproach
}
func (p Plane) IsFlying() bool {
	return p.state == StateFlying || p.state == StateAproach
}
func (p Plane) IsConsumingFuel() bool {
	return p.state == StateWaiting ||
		p.state == StateRolling ||
		p.state == StateFlying ||
		p.state == StateAproach
}

func (p Plane) String() string {
	return fmt.Sprintf("%c: %s %s %d", p.callsign, p.Position, p.Direction, p.height)
}

func (p Plane) Flightplan() string {
	return fmt.Sprintf("%c%d%c %c-%c",
		p.callsign, p.initial_height, p.typ.mark, p.entry.sign, p.exit.sign)
}

func (p Plane) Marker() string {
	return fmt.Sprintf("%c%d", p.callsign, p.height)
}

func (p Plane) State() string {
	res := fmt.Sprintf("%c%d%c %c-%c %s",
		p.callsign, p.height, p.typ.mark, p.entry.sign, p.exit.sign, p.Direction)

	if p.is_hoovering {
		res += " +H+ "
	}

	if p.fuel_left >= FUEL_INDICATOR {
		res += " + "
	}
	// height not shown on approach
	show_height := p.want_height != p.height && p.state != StateAproach
	show_dir := p.want_turn != 0

	if show_height && show_dir {
		res += fmt.Sprintf(" [%s %d]",
			p.Direction.Right(p.want_turn),
			p.want_height)
	} else if show_height {
		res += fmt.Sprintf(" [%d]", p.want_height)
	} else if show_dir {
		res += fmt.Sprintf(" [%s]",
			p.Direction.Right(p.want_turn))
	}

	switch {
	case p.state == StateWaiting:
		res += " -- Awaiting Takeoff --"
	case p.state == StateRolling:
		res += " -- Rolling! --"
	case p.is_holding:
		res += " -- Holding --"
	case p.state == StateAproach:
		res += " -- Final Approach --"
	case p.direction_at_navaid != 0:
		res += " -- Cleared --"
	case p.state == StateLanded:
		res += " -- Landed --"
	case p.state == StateDeparted:
		res += " -- Departed Area --"
	}

	return res
}

type ByTime []*Plane

func (a ByTime) Len() int           { return len(a) }
func (a ByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTime) Less(i, j int) bool { return a[i].start > a[j].start }
