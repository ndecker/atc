package main

type PlaneType struct {
	mark   rune
	name   string
	weight int

	ticks_per_move Ticks
	moves_per_tick int
	ticks_pending  Ticks
	ticks_rolling  Ticks

	entry_min_height int
	entry_max_height int
	exit_height      int

	initial_fuel Ticks

	immediate_turn  bool
	can_hoover      bool
	can_enter_nofly bool

	entry_exit_routes bool
	airport_loop      bool
	airport_entry     bool
	airport_exit      bool
}

var (
	PLANE_TYPE_JET = PlaneType{
		mark:   'J',
		name:   "Jet",
		weight: 12,

		ticks_per_move: 1,
		moves_per_tick: 1,
		ticks_pending:  4,
		ticks_rolling:  2,

		entry_min_height: 6,
		entry_max_height: 9,
		exit_height:      5,

		initial_fuel: 15 * Minutes,

		immediate_turn: false,
		can_hoover:     false,

		entry_exit_routes: true,
		airport_loop:      false,
		airport_entry:     true,
		airport_exit:      true,
	}

	PLANE_TYPE_PROP = PlaneType{
		mark:   'P',
		name:   "Prop",
		weight: 8,

		ticks_per_move: 2,
		moves_per_tick: 1,
		ticks_pending:  4,
		ticks_rolling:  4,

		entry_min_height: 6,
		entry_max_height: 9,
		exit_height:      5,

		initial_fuel: 21 * Minutes,

		immediate_turn: false,
		can_hoover:     false,

		entry_exit_routes: true,
		airport_loop:      true,
		airport_entry:     true,
		airport_exit:      true,
	}

	PLANE_TYPE_HELI = PlaneType{
		mark:   'H',
		name:   "Heli",
		weight: 2,

		ticks_per_move: 2,
		moves_per_tick: 1,
		ticks_pending:  4,
		ticks_rolling:  0,

		entry_min_height: 2,
		entry_max_height: 4,
		exit_height:      5,

		initial_fuel: 15 * Minutes,

		immediate_turn: true,
		can_hoover:     true,

		entry_exit_routes: false,
		airport_loop:      false,
		airport_entry:     true,
		airport_exit:      true,
	}

	PLANE_TYPE_BLACKBIRD = PlaneType{
		mark:   'B',
		name:   "Blackbird",
		weight: 1,

		ticks_per_move: 1,
		moves_per_tick: 2,
		ticks_pending:  2,
		ticks_rolling:  0,

		entry_min_height: 10,
		entry_max_height: 10,
		exit_height:      10,

		initial_fuel: 15 * Minutes,

		immediate_turn:  false,
		can_hoover:      false,
		can_enter_nofly: true,

		entry_exit_routes: true,
		airport_loop:      false,
		airport_entry:     false,
		airport_exit:      false,
	}
)

func PlaneTypes(setup *GameSetup) []*PlaneType {
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
	if setup.have_blackbird {
		plane_types = append(plane_types, &PLANE_TYPE_BLACKBIRD)
	}
	return plane_types
}
