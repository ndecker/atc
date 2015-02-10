package main

import (
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"os"
	"os/signal"
	"strconv"
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
	bottom := top + height

	for x := 0; x < game.board.width; x += 1 {
		for y := 0; y < game.board.height; y += 1 {
			printC(left+2*x, top+y, termbox.ColorBlue, "· ")
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

	printPlane := func(plane *Plane, color termbox.Attribute) {
		if plane != nil && plane.IsFlying() {
			printC(left+plane.Position.x*2, top+plane.Position.y,
				color, plane.Marker())
		}
	}

	for _, p := range game.planes {
		if row >= bottom {
			row = top
			col += 10
		}

		if p.IsVisible() {
			print(col, row, p.Flightplan(), " *")
			row += 1
		} else if p.IsActive() {
			print(col, row, p.Flightplan())
			row += 1
		}

		printPlane(p, termbox.ColorDefault)
	}

	// always show last commanded plane on top
	printPlane(game.ci.last_commanded_plane, termbox.ColorDefault)

	x := left
	y := game.board.height + 2

	x = print(x, y, game.clock.String(), "  ")
	if game.end_reason != nil {
		x0 := print(x, y+0, "-- ", game.end_reason.message, " --")

		for _, p := range game.end_reason.planes {
			printPlane(p, termbox.ColorRed)
			x0 = print(x0, y, " ", p.Marker())
		}
		print(x, y+1, "(Press Esc to quit / R to restart same game)")
	} else {
		print(x, y, game.ci.StatusLine())
	}
}

func RunGame(setup *GameSetup, board *Board, seed int64) {
	tick_time := time.Duration(SECONDS_PER_TICK) * time.Second
	timer := time.NewTimer(tick_time)
	defer timer.Stop()

	game := NewGame(setup, board, seed)

	var help_visible bool = false
	var help_screen int = 0
	var planes_visible bool = false

	for {
		DrawGame(game)
		if help_visible {
			DrawHelp(help_screen)
		}
		if planes_visible {
			planes_visible = DrawPlanes(game)
		}
		termbox.Flush()

		select {
		case <-timer.C:
			game.Tick()
			timer.Reset(tick_time)

		case ev := <-events:
			switch ev.Type {
			case termbox.EventKey:
				switch {
				case ev.Ch == 0 && ev.Key == termbox.KeyCtrlC:
					// always handle Ctrl-C
					return

				case help_visible:
					DialogKeys(ev, &help_visible, &help_screen)
				case planes_visible:
					DialogKeys(ev, &planes_visible, nil)
				default:
					switch ev.Ch {
					case 0:
						switch ev.Key {
						case termbox.KeyEsc, termbox.KeyCtrlC:
							return // end game
						case termbox.KeySpace,
							termbox.KeyEnter:
							game.ci.Clear()
						case termbox.KeyBackspace, termbox.KeyBackspace2:
							game.ci.Clear()
						case termbox.KeyTab:
							planes_visible = true
						}
					case ',':
						game.Tick()

						if game.setup.skip_to_next_tick {
							timer.Reset(tick_time)
						}
					case '?':
						help_visible = true
					case 'R', 'r':
						if game.end_reason != nil {
							game = NewGame(setup, board, seed)
						} else {
							game.KeyPressed(unicode.ToUpper(ev.Ch))
						}
					default:
						game.KeyPressed(unicode.ToUpper(ev.Ch))
					}
				}

			case termbox.EventResize:
				// nothing; just redraw
			}

		case <-sigterm:
			return
		}
	}
}

func usage() {
	fmt.Println("usage: atc [time [planes]]")
}

func main() {
	single_game := false
	setup := DefaultSetup()
	board := DEFAULT_BOARD

	switch len(os.Args) {
	case 2:
		time, err := strconv.Atoi(os.Args[1])
		if err != nil {
			usage()
			return
		}
		setup.duration = Ticks(time) * Minutes
		single_game = true
	case 3:
		time, err := strconv.Atoi(os.Args[1])
		if err != nil {
			usage()
			return
		}
		planes, err := strconv.Atoi(os.Args[2])
		if err != nil {
			usage()
			return
		}
		setup.duration = Ticks(time) * Minutes
		setup.num_planes = planes
		single_game = true
	}

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

	if single_game {
		seed := RandSeed()
		RunGame(setup, board, seed)
		return
	}

	RunMenu("ATC - Air Traffic Control", []MenuEntry{
		MenuEntry{key: 'S', text: "Start", action: func() {
			seed := RandSeed()
			RunGame(setup, board, seed)
		}},
		MenuEntry{key: 'B', textf: func() string {
			return fmt.Sprintf("Choose Board [%s]", board.name)
		}, action: func() {
			RunMenu("Choose Board", []MenuEntry{
				MenuEntry{key: 'S', text: "ATC Standard", action: func() {
					board = DEFAULT_BOARD
					CloseMenu()
				}},
				MenuEntry{key: 'C', text: "Crossways", action: func() {
					board = CROSSWAYS_BOARD
					CloseMenu()
				}},
			})
		}},

		MenuEntry{},
		MenuEntry{key: 'O', text: "Options", action: func() {
			RunMenu("Options", []MenuEntry{
				MenuEntry{text: "Return", action: CloseMenu},
				MenuEntry{},

				MenuBoolText('J', &setup.have_jet, "Jet enabled", "Jet disabled"),
				MenuBoolText('P', &setup.have_prop, "Prop enabled", "Prop disabled"),
				MenuBoolText('H', &setup.have_heli, "Heli enabled", "Prop disabled"),
				MenuBoolText('B', &setup.have_blackbird, "Blackbird enabled", "Blackbird disabled"),
			})
		}},

		MenuEntry{},
		MenuEntry{key: 'Q', text: "Quit", action: CloseMenu},
	})

}
