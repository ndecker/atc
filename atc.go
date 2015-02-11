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
	events chan termbox.Event = make(chan termbox.Event, 0)
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

	for _, nf := range game.board.nofly {
		printC(left+nf.x*2, top+nf.y, termbox.ColorBlue, "XX")
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

func RunGame(rules *GameRules, board *Board, diff *Difficulty, seed int64) {
	tick_time := time.Duration(SECONDS_PER_TICK) * time.Second
	timer := time.NewTimer(tick_time)
	defer timer.Stop()

	game := NewGame(rules, board, diff, seed)

	var help_visible bool = false
	var help_screen uint = 0
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
				case help_visible:
					DialogKeys(ev, &help_visible, &help_screen)
				case planes_visible:
					DialogKeys(ev, &planes_visible, nil)
				default:
					switch ev.Ch {
					case 0:
						switch ev.Key {
						case termbox.KeyEsc:
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

						if game.rules.skip_to_next_tick {
							timer.Reset(tick_time)
						}
					case '?':
						help_visible = true
					case 'R', 'r':
						if game.end_reason != nil {
							game = NewGame(rules, board, diff, seed)
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
		}
	}
}

func MainMenu() {
	rules := &ATC_ORIGINAL_RULES
	board := DEFAULT_BOARD
	diff := DIFFICULTIES[0]

	active := 0
	for {
		menu := []string{
			"Start Game",
			"",
			Pad(30, "Board", "["+board.name+"]"),
			Pad(30, "Rules", "["+rules.name+"]"),
			Pad(30, "Difficulty", "["+diff.name+"]"),
			"",
			"Options",
			"",
			"Quit",
		}

		res := RunMenu("ATC - Air Traffic Control", menu, active)
		switch res {
		case MENU_ESCAPE, 8:
			return
		case 0:
			seed := RandSeed()
			RunGame(rules, board, diff, seed)
		case 2:
			BoardMenu(&board)
		case 3:
			RulesMenu(&rules)
		case 4:
			DifficultyMenu(&diff)
		case 6:
			OptionsMenu(&rules)
		}
		active = res
	}
}

func BoardMenu(board **Board) {
	menu := make([]string, len(BOARDS))
	active := 0
	for nr, b := range BOARDS {
		menu[nr] = b.name
		if b == *board {
			active = nr
		}
	}

	for {
		res := RunMenu("Choose Board", menu, active)
		switch {
		case res == MENU_ESCAPE:
			return
		case res >= 0:
			*board = BOARDS[res]
			return
		}
	}
}

func RulesMenu(rules **GameRules) {
	menu := make([]string, len(RULES))
	active := 0
	for nr, r := range RULES {
		menu[nr] = r.name
		if r == *rules {
			active = nr
		}
	}

	for {
		res := RunMenu("Select Rules", menu, active)
		switch {
		case res == MENU_ESCAPE:
			return
		case res >= 0:
			*rules = RULES[res]
			return
		}
	}
}

func DifficultyMenu(diff **Difficulty) {
	menu := make([]string, len(DIFFICULTIES))
	active := 0
	for nr, d := range DIFFICULTIES {
		menu[nr] = d.name
		if d == *diff {
			active = nr
		}
	}

	for {
		res := RunMenu("Select difficulty", menu, active)
		switch {
		case res == MENU_ESCAPE:
			return
		case res >= 0:
			*diff = DIFFICULTIES[res]
			return
		}
	}
}

func OptionsMenu(rules **GameRules) {
	WIDTH := 25
	active := 0

	mark := func(x bool) string {
		if x {
			return "[X]"
		} else {
			return "[ ]"
		}
	}

	r := **rules

	for {
		menu := []string{
			"Main Menu",
			"",
			Pad(WIDTH, "Jet", mark(r.have_jet)),
			Pad(WIDTH, "Prop", mark(r.have_prop)),
			Pad(WIDTH, "Helicopter", mark(r.have_heli)),
			Pad(WIDTH, "Blackbird", mark(r.have_blackbird)),
			"",
			Pad(WIDTH, "Show pending planes", mark(r.show_pending_planes)),
			Pad(WIDTH, ". delays commands", mark(r.delayed_commands)),
			Pad(WIDTH, ", skips to next tick", mark(r.skip_to_next_tick)),
		}
		res := RunMenu("Choose options", menu, active)
		switch res {
		case MENU_ESCAPE, 0:
			r.name = "Custom"
			*rules = &r
			return
		case 2:
			r.have_jet = !r.have_jet
		case 3:
			r.have_prop = !r.have_prop
		case 4:
			r.have_heli = !r.have_heli
		case 5:
			r.have_blackbird = !r.have_blackbird
		case 7:
			r.show_pending_planes = !r.show_pending_planes
		case 8:
			r.delayed_commands = !r.delayed_commands
		case 9:
			r.skip_to_next_tick = !r.skip_to_next_tick
		}
		active = res
	}
}

func main() {
	var err error

	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.HideCursor()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	go func() {
		<-sigterm
		termbox.Close()
		os.Exit(1)
	}()

	go func() {
		for {
            ev := termbox.PollEvent()
            if ev.Ch == 0 && ev.Key == termbox.KeyCtrlC {
                // always terminate on Ctrl+C
                termbox.Close()
                os.Exit(1)
            }
            if ev.Type == termbox.EventError {
                termbox.Close()
                fmt.Println(ev)
                os.Exit(1)
            }
			events <- ev
		}
	}()

	usage := func() {
		termbox.Close()
		fmt.Println("usage: atc [time [planes]]")
		os.Exit(1)
	}

	num_planes := 26
	switch len(os.Args) {
	case 3:
		num_planes, err = strconv.Atoi(os.Args[2])
		if err != nil {
			usage()
		}
		fallthrough
	case 2:
		time, err := strconv.Atoi(os.Args[1])
		if err != nil {
			usage()
		}

		time = Max(time, 16) // minimum 16 minutes
		diff := &Difficulty{
			duration:   Ticks(time) * Minutes,
			num_planes: num_planes,
		}
		seed := RandSeed()
		RunGame(&ATC_ORIGINAL_RULES, DEFAULT_BOARD, diff, seed)
	case 1:
		MainMenu()
	default:
		usage()
	}
}
