package main

import (
	termbox "github.com/nsf/termbox-go"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unicode"
)

var (
	sigterm chan os.Signal     = make(chan os.Signal, 1)
	events  chan termbox.Event = make(chan termbox.Event, 0)
)

func DrawGame(game *GameState) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	termw, termh := termbox.Size()

	width := game.board.width*2 + 2 + 10
	left := (termw - width) / 2

	height := game.board.height + 2
	top := (termh - height) / 2

	for x := 0; x < game.board.width; x += 1 {
		for y := 0; y < game.board.height; y += 1 {
			FprintS(left+2*x, top+y, "· ", termbox.ColorBlue)
		}
	}

	for _, ep := range game.board.entrypoints {
		print(left+ep.Position.x*2, top+ep.Position.y, string(ep.sign))
	}

	for _, navaid := range game.board.navaids {
		print(left+navaid.x*2, top+navaid.y, "*")
	}

	col := left + game.board.width*2 + 2
	row := top

	printPlane := func(plane *Plane, red bool) {
		if plane != nil && plane.IsFlying() {
			if !red {
				printS(left+plane.Position.x*2, top+plane.Position.y, plane.Marker())
			} else {
				FprintS(left+plane.Position.x*2, top+plane.Position.y, plane.Marker(), termbox.ColorRed)
			}
		}
	}

	for _, p := range game.planes {
		if row >= termh {
			row = 0
			col += 10
		}

		// TODO: two column?
		if p.IsVisible() {
			print(col, row, p.Flightplan(), " *")
			row += 1
		} else if p.IsActive() {
			print(col, row, p.Flightplan())
			row += 1

		}

		printPlane(p, false)
	}

	// always show last commanded plane on top
	printPlane(game.ci.last_commanded_plane, false)

	// TODO: dynamic positions
	print(left+0, top+21, game.clock.String())
	if game.end_reason != nil {
		print(left+8, top+21, game.end_reason.message)
		print(left+8, top+22, "(Press Esc to quit / R to restart same game)")

		for _, p := range game.end_reason.planes {
			printPlane(p, true)
		}
	} else {
		print(left+8, top+21, game.ci.StatusLine())
	}
}

func RunGame(setup GameSetup, seed int64) {
	tick_time := time.Duration(SECONDS_PER_TICK) * time.Second
	timer := time.NewTimer(tick_time)
	defer timer.Stop()

	game := NewGame(DEFAULT_SETUP, seed)

	var help_visible bool = false
	var planes_visible bool = false

	for {
		DrawGame(game)
		if help_visible {
			DrawHelp()
		}
		if planes_visible {
			DrawPlanes(game)
		}
		termbox.Flush()

		select {
		case <-timer.C:
			game.Tick()
			timer.Reset(tick_time)

		case ev := <-events:
			switch ev.Type {
			case termbox.EventKey:
				switch ev.Ch {
				case 0:
					switch ev.Key {
					case termbox.KeyEsc, termbox.KeyCtrlC:
						return // end game
					case termbox.KeySpace,
						termbox.KeyEnter,
						termbox.KeyBackspace, termbox.KeyBackspace2:
						if help_visible || planes_visible {
							help_visible = false
							planes_visible = false
						} else {
							game.ci.Clear()
						}
					case termbox.KeyTab:
						if game.setup.show_planes {
							planes_visible = !planes_visible
						}
					}
				case ',':
					game.Tick()

					if game.setup.skip_to_next_tick {
						timer.Reset(tick_time)
					}
				case '?':
					help_visible = !help_visible
				case 'R', 'r':
					if game.end_reason != nil {
						game = NewGame(setup, seed)
					} else {
						game.KeyPressed(unicode.ToUpper(ev.Ch))
					}
				default:
					game.KeyPressed(unicode.ToUpper(ev.Ch))
				}
			case termbox.EventResize:
				// nothing; just redraw
			}

		case <-sigterm:
			return
		}
	}
}

func main() {
	signal.Notify(sigterm, syscall.SIGTERM)

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.HideCursor()

	go func() {
		for {
			events <- termbox.PollEvent()
		}
	}()

	setup := DEFAULT_SETUP
	seed := RandSeed()
	RunGame(setup, seed)
}
