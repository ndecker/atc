package main

// TODO: fuel: jet 15m prop 21m

type PlaneType struct {
	mark   rune
	name   string
	weight int

	ticks_per_move Ticks
	ticks_pending  Ticks
	ticks_rolling  Ticks

	initial_fuel Ticks

	immediate_turn bool
	can_hoover     bool

	entry_exit_routes bool
	airport_loop      bool
}

var (
	PLANE_TYPE_JET = PlaneType{
		mark:   'J',
		name:   "Jet",
		weight: 6,

		ticks_per_move: 1,
		ticks_pending:  4,
		ticks_rolling:  2,

		initial_fuel: 15 * Minutes,

		immediate_turn: false,
		can_hoover:     false,

		entry_exit_routes: true,
		airport_loop:      false,
	}

	PLANE_TYPE_PROP = PlaneType{
		mark:   'P',
		name:   "Prop",
		weight: 4,

		ticks_per_move: 2,
		ticks_pending:  4,
		ticks_rolling:  4,

		initial_fuel: 21 * Minutes,

		immediate_turn: false,
		can_hoover:     false,

		entry_exit_routes: true,
		airport_loop:      true,
	}

	PLANE_TYPE_HELI = PlaneType{
		mark:   'H',
		name:   "Heli",
		weight: 1,

		ticks_per_move: 2,
		ticks_pending:  4,
		ticks_rolling:  0,

		initial_fuel: 15 * Minutes,

		immediate_turn: true,
		can_hoover:     true,

		entry_exit_routes: false,
		airport_loop:      false,
	}
)

func PlaneTypes(setup GameSetup) []*PlaneType {
	plane_types := make([]*PlaneType, 0, 3)
	if setup.have_jet {
		plane_types = append(plane_types, &PLANE_TYPE_JET)
	}
	if setup.have_prop {
		plane_types = append(plane_types, &PLANE_TYPE_PROP)
	}
	if setup.have_heli {
		plane_types = append(plane_types, &PLANE_TYPE_HELI)
	}
	return plane_types
}
