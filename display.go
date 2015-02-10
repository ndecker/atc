package main

import (
	"fmt"
	termbox "github.com/nsf/termbox-go"
)

func DisplayWindow(width, height int) (int, int) {
	termw, termh := termbox.Size()

	left := Max((termw-width-4)/2, 0)
	right := left + width + 2
	top := Max((termh-height-2)/2, 0)
	bottom := top + height + 2

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

func DrawPlanes(game *GameState) {
	COL_WIDTH := 26

	num_planes := len(game.planes)
	var cols int
	switch {
	case num_planes < 20:
		cols = 1
	case num_planes < 60:
		cols = 2
	case num_planes < 100:
		cols = 3
	case num_planes < 150:
		cols = 4
	default:
		cols = 5
	}

	rows := num_planes / cols
	if num_planes%cols != 0 {
		rows += 1
	}

	x, y := DisplayWindow(cols*COL_WIDTH, rows)

	print(x, y, "Planes:")

	x += 1
	y += 1
	for pnr, p := range game.planes {
		row, col := pnr%rows, pnr/rows
		if p.IsActive() {
			printS(x+(col*COL_WIDTH), y+row, p.String())
		} else if p.IsDone() {
			FprintS(x+(col*COL_WIDTH), y+row, p.String(), termbox.ColorGreen)
		} else {
			FprintS(x+(col*COL_WIDTH), y+row, p.String(), termbox.ColorBlue)
		}
	}
	termbox.Flush()
}

func DrawHelp() {
	lines, maxlen := SplitLines(COMMAND_HELP)
	x, y := DisplayWindow(maxlen+4, len(lines)+1)

	print(x, y, "Help")
	y += 1

	for _, l := range lines {
		print(x+1, y, l)
		y += 1
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

func FprintS(x, y int, s string, fg termbox.Attribute) {
	for _, r := range s {
		termbox.SetCell(x, y, r,
			fg, termbox.ColorDefault)
		x += 1
	}
}
