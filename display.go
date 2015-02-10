package main

import (
	"fmt"
	termbox "github.com/nsf/termbox-go"
)

const (
	BORDER_H  = 2
	BORDER_V  = 1
	PAD_SPACE = "                                                              "
)

func DrawWindow(title string, footer string, lines []string, colors []termbox.Attribute) {
	termw, termh := termbox.Size()
	contw, conth := termw-2*BORDER_H, termh-2*BORDER_V

	num_lines := len(lines)
	max_len := 0
	for _, line := range lines {
		max_len = Max(max_len, len(line))
	}

	cols := 1
	rows := num_lines
	for (cols+1)*max_len+cols < contw && rows > conth {
		cols++
		rows = (num_lines + cols - 1) / cols
	}

	contw = cols*max_len + (cols - 1)
	conth = rows

	if len(title)+2 > contw || len(footer)+2 > contw {
		contw = Max(len(title), len(footer))
	}

	left := Max((termw-contw-2*BORDER_H)/2, 0)
	right := left + contw + BORDER_H + 1
	top := Max((termh-conth-2*BORDER_V)/2, 0)
	bottom := top + conth + BORDER_V

	// draw border
	for x := left; x <= right; x += 1 {
		print(x, top, "-")
		print(x, bottom, "-")
	}

	for y := top + 1; y <= bottom-1; y += 1 {
		print(left, y, "|")
		print(right, y, "|")

		// clear content area
		for x := left + 1; x <= right-1; x += 1 {
			print(x, y, " ")
		}
	}

	if title != "" {
		print(left+contw/2-len(title)/2, top, " ", title, " ")
	}
	if footer != "" {
		print(left+contw/2-len(footer)/2, bottom, " ", footer, " ")
	}

	left += BORDER_H
	top += BORDER_V

	for pos, line := range lines {
		color := termbox.ColorDefault
		if pos < len(colors) {
			color = colors[pos]
		}
		printC(
			left+(pos/rows)*(max_len+1),
			top+(pos%rows),
			color, line, PAD_SPACE[0:max_len-len(line)])
	}
}

func DrawPlanes(game *GameState) bool {
	lines := make([]string, 0, len(game.planes))
	colors := make([]termbox.Attribute, 0, len(game.planes))

	for _, p := range game.planes {
		if !game.rules.show_pending_planes && p.state == StatePending {
			continue
		}

		lines = append(lines, p.String())
		switch {
		case p.IsActive():
			colors = append(colors, termbox.ColorDefault)
		case p.IsDone():
			colors = append(colors, termbox.ColorGreen)
		default:
			colors = append(colors, termbox.ColorBlue)
		}
	}

	if len(lines) == 0 {
		return false
	}
	DrawWindow("Planes", "", lines, colors)
	return true
}

func DrawHelp(screen uint) {
	screen = screen % uint(len(HELP))
	lines := SplitLines(HELP[screen])
	DrawWindow(
		fmt.Sprintf("Help (page %d of %d)", screen+1, len(HELP)),
		"<- / -> / Space", lines, nil)
}

func DialogKeys(ev termbox.Event, visible *bool, screen *uint) {
	switch ev.Ch {
	case 0:
		switch ev.Key {
		case termbox.KeyEsc,
			termbox.KeyBackspace, termbox.KeyBackspace2,
			termbox.KeyTab:
			*visible = false
		case termbox.KeySpace,
			termbox.KeyEnter:
			if screen != nil {
				*screen++
			}
		case termbox.KeyArrowRight:
			if screen != nil {
				*screen++
			}
		case termbox.KeyArrowLeft:
			if screen != nil {
				*screen--
			}
		}
	case 'x', 'X', 'q', 'Q':
		*visible = false
	}
}

func print(x, y int, strings ...string) int {
	return printC(x, y, termbox.ColorDefault, strings...)
}

func printC(x, y int, fg termbox.Attribute, strings ...string) int {
	for _, s := range strings {
		for _, r := range s {
			termbox.SetCell(x, y, r,
				fg, termbox.ColorDefault)
			x += 1
		}
	}
	return x
}
