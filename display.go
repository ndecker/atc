package main

import (
	"fmt"
	termbox "github.com/nsf/termbox-go"
)

func DisplayWindow(width, height int) (int, int) {
	termw, termh := termbox.Size()

	left := Max((termw-width-4)/2, 0)
	right := left + width + 1
	top := Max((termh-height-2)/2, 0)
	bottom := top + height + 1

	for x := left; x <= right; x += 1 {
		printS(x, top, "-")
		printS(x, bottom, "-")
	}

	for y := top + 1; y <= bottom-1; y += 1 {
		printS(left, y, "|")
		printS(right, y, "|")

		for x := left + 1; x <= right-1; x += 1 {
			printS(x, y, " ")
		}
	}

	return left + 2, top + 1
}

func ShowPlanes(game *GameState) {
	x, y := DisplayWindow(32, len(game.planes)+1)

	x += 1
	print(x, y, "Planes:")

	x += 1
	y += 1
	for row, p := range game.planes {
		print(x, y+row, p)
	}
	termbox.Flush()
	WaitForContinue()

}

func ShowHelp() {
	lines, maxlen := SplitLines(COMMAND_HELP)
	x, y := DisplayWindow(maxlen+4, len(lines)+1)

	print(x, y, "Help")
	y += 1

	for _, l := range lines {
		print(x+1, y, l)
		y += 1
	}
	termbox.Flush()
	WaitForContinue()
}

func WaitForContinue() bool {
	// Enter, Space, Esc, Ctrl-C
	for {
		ev := <-events
		switch ev.Type {
		case termbox.EventKey:
			switch ev.Ch {
			case 0:
				switch ev.Key {
				case termbox.KeyEsc,
					termbox.KeyCtrlC:
					return false
				case termbox.KeySpace,
					termbox.KeyEnter:
					return true
				}
			}
		}
	}
}

func printS(x, y int, s string) {
	for _, r := range s {
		termbox.SetCell(x, y, r,
			termbox.ColorDefault, termbox.ColorDefault)
		x += 1
	}
}

func print(x, y int, args ...interface{}) {
	s := fmt.Sprint(args...)
	for _, r := range s {
		termbox.SetCell(x, y, r,
			termbox.ColorDefault, termbox.ColorDefault)
		x += 1
	}
}
