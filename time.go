package main

import "fmt"

// time in ticks
type Ticks int

const (
	TICKS_PER_MINUTE = 4
	SECONDS_PER_TICK = 60 / TICKS_PER_MINUTE
	Minutes          = Ticks(TICKS_PER_MINUTE)
)

func (t *Ticks) Tick() {
	*t -= Ticks(1)
}

func (t Ticks) String() string {
	return fmt.Sprintf("%2d:%02d", t/TICKS_PER_MINUTE, SECONDS_PER_TICK*(t%TICKS_PER_MINUTE))
}
