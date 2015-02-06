package main

type PlaneType struct {
    mark rune
    ticks_per_move Ticks
    ticks_pending Ticks
    ticks_rolling Ticks
    immediate_turn bool
}

var (
    PLANE_TYPE_JET = PlaneType {
        mark: 'J',
        ticks_per_move: 1,
        ticks_pending: 4,
        ticks_rolling: 2,
        immediate_turn: false,
    }

    PLANE_TYPE_PROP = PlaneType {
        mark: 'P',
        ticks_per_move: 2,
        ticks_pending: 4,
        ticks_rolling: 4,
        immediate_turn: false,
    }

    PLANE_TYPE_HELI = PlaneType {
        mark: 'H',
        ticks_per_move: 2,
        ticks_pending: 4,
        ticks_rolling: 0,
        immediate_turn: true,
        // TODO: stand still
    }

    PLANE_TYPES = []PlaneType{PLANE_TYPE_HELI}
)
