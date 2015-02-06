package main

// TODO: fuel: jet 15m prop 21m

type PlaneType struct {
	mark           rune
	name           string
	ticks_per_move Ticks
	ticks_pending  Ticks
	ticks_rolling  Ticks

	initial_fuel Ticks

	immediate_turn bool
	can_hoover     bool
}

var (
	PLANE_TYPE_JET = PlaneType{
		mark:           'J',
		name:           "Jet",
		ticks_per_move: 1,
		ticks_pending:  4,
		ticks_rolling:  2,

		initial_fuel: 15 * Minutes,

		immediate_turn: false,
		can_hoover:     false,
	}

	PLANE_TYPE_PROP = PlaneType{
		mark:           'P',
		name:           "Prop",
		ticks_per_move: 2,
		ticks_pending:  4,
		ticks_rolling:  4,

		initial_fuel: 21 * Minutes,

		immediate_turn: false,
		can_hoover:     false,
	}

	PLANE_TYPE_HELI = PlaneType{
		mark:           'H',
		name:           "Heli",
		ticks_per_move: 2,
		ticks_pending:  4,
		ticks_rolling:  0,

		initial_fuel: 15 * Minutes,

		immediate_turn: true,
		can_hoover:     true,
	}

	PLANE_TYPES = []PlaneType{
		PLANE_TYPE_JET,
		PLANE_TYPE_PROP,
		PLANE_TYPE_HELI,
	}
)
