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
			termbox.SetCell(left+2*x, top+y, '·',
				termbox.ColorBlue, termbox.ColorDefault)
			termbox.SetCell(left+2*x+1, top+y, ' ',
				termbox.ColorBlue, termbox.ColorDefault)
		}
	}

	for _, ep := range game.board.entrypoints {
		print(left+ep.Position.x*2, top+ep.Position.y, string(ep.sign))
	}

	for _, b := range game.board.beacons {
		print(left+b.Position.x*2, top+b.Position.y, "*")
	}

	col := left + game.board.width*2 + 2
	row := top

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

		if p.IsFlying() {
			print(left+p.Position.x*2, top+p.Position.y, p.Marker())
		}
	}

	if game.ci.last_commanded_plane != nil {
		// always show last commanded plane on top
		p := game.ci.last_commanded_plane
		if p.IsFlying() {
			print(left+p.Position.x*2, top+p.Position.y, p.Marker())
		}
	}

	// TODO: dynamic positions
	print(left+0, top+21, game.clock.String())
	if game.end_reason != "" {
		print(left+8, top+21, game.end_reason)
	} else {
		print(left+8, top+21, game.ci.StatusLine())
	}
}

func GameLoop(game *GameState) {
	tick_time := time.Duration(SECONDS_PER_TICK) * time.Second
	timer := time.NewTimer(tick_time)
	defer timer.Stop()

	var help_visible bool = false

	for {
		DrawGame(game)
		if help_visible {
			DrawHelp()
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
						if help_visible {
							help_visible = false
						} else {
							game.ci.Clear()
						}
					}
				case ',':
					game.Tick()

					if game.setup.skip_to_next_tick {
						timer.Reset(tick_time)
					}
				case '?':
					help_visible = !help_visible
				default:
					game.KeyPressed(unicode.ToUpper(ev.Ch))
				}
			case termbox.EventResize:
				// nothing; just redraw
			}

		case <-sigterm:
			return
		}

		if game.end_reason != "" {
			DrawGame(game)
			termbox.Flush()
			<-events
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

	for {
		seed := RandSeed()
		game := NewGame(DEFAULT_SETUP, seed)

		if game.setup.show_planes_at_start {
			ShowPlanes(game)
		}

		GameLoop(game)
		if !WaitForContinue() {
			return
		}
	}
}
