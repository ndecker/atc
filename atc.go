package main

import (
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"os"
	"os/signal"
	"syscall"
	"unicode"
)

var (
    sigterm chan os.Signal = make(chan os.Signal, 1)
	events chan termbox.Event = make(chan termbox.Event, 0)
)

func DrawGame(game *GameState) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	for x := 0; x < game.setup.width; x += 1 {
		for y := 0; y < game.setup.height; y += 1 {
			termbox.SetCell(2*x, y, '·',
				termbox.ColorBlue, termbox.ColorDefault)
			termbox.SetCell(2*x+1, y, ' ',
				termbox.ColorBlue, termbox.ColorDefault)
		}
	}

	for _, ep := range game.board.entrypoints {
		print(ep.Position.x*2, ep.Position.y, string(ep.sign))
	}

	for _, b := range game.board.beacons {
		print(b.Position.x*2, b.Position.y, "*")
	}

	row := 0
	for _, p := range game.planes {
		// TODO: two column?
		if p.IsVisible() {
			print(52, row, p.Flightplan(), " *")
			row += 1
		} else if p.IsActive() {
			print(52, row, p.Flightplan())
			row += 1

		}

		if p.IsFlying() {
			print(p.Position.x*2, p.Position.y, p.Marker())
		}
	}

    if game.last_commanded_plane != nil {
        // always show last commanded plane on top
        p := game.last_commanded_plane
        if p.IsFlying() {
			print(p.Position.x*2, p.Position.y, p.Marker())
        }
    }

	// TODO: dynamic positions
	print(0, 21, game.clock.String())
	if game.end_reason != "" {
		print(8, 21, game.end_reason)
	} else {
		print(8, 21, game.ci.StatusLine())
	}

	termbox.Flush()
}

func print(x, y int, args ...interface{}) {
	s := fmt.Sprint(args...)
	for _, r := range s {
		termbox.SetCell(x, y, r,
			termbox.ColorDefault, termbox.ColorDefault)
		x += 1
	}
}


func GameLoop(game *GameState) {
    ticks, close_ticks := MakeTicker()
    defer close(close_ticks)


	for {
		DrawGame(game)

		select {
		case <-ticks:
			game.Tick()

		case ev := <-events:
			switch ev.Type {
			case termbox.EventKey:
				switch ev.Ch {
				case 0:
					switch ev.Key {
					case termbox.KeyEsc, termbox.KeyCtrlC:
						return // end game
					case termbox.KeySpace:
						game.ci.KeyPressed(' ')
					case termbox.KeyBackspace, termbox.KeyBackspace2:
						game.ci.Clear()
					}
				case ',':
					game.Tick()
				default:
					game.ci.KeyPressed(unicode.ToUpper(ev.Ch))
				}
			case termbox.EventResize:
				// nothing; just redraw
			}

		case <-sigterm:
			return
		}

		if game.end_reason != "" {
			DrawGame(game)
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
        GameLoop(game)

		ev := <-events
        switch ev.Type {
        case termbox.EventKey:
            switch ev.Ch {
            case 0:
                switch ev.Key {
                case termbox.KeyEsc, termbox.KeyCtrlC:
                    return // end game
                }
            }
        }
    }
}
