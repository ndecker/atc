package main

import (
	termbox "github.com/nsf/termbox-go"
	"unicode"
)

const (
	MENU_ESCAPE = -1
)

func RunMenu(title string, entries []string, active int) int {
	colors := make([]termbox.Attribute, len(entries))
	if active < 0 || active >= len(entries) {
		panic("invalid active value")
	}

	shortcuts := make(map[rune]int, len(entries))
	for nr, e := range entries {
		sc := unicode.ToUpper(FirstRune(e))
		_, ok := shortcuts[sc]
		if !ok {
			shortcuts[sc] = nr
		}
	}

	for {

		for nr, _ := range entries {
			colors[nr] = termbox.ColorDefault
		}
		colors[active] = colors[active] | termbox.AttrReverse

		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		DrawWindow(title, "", entries, colors)
		termbox.Flush()

		ev := <-events

		if ev.Type != termbox.EventKey {
			continue
		}

		if ev.Ch != 0 {
			pos, ok := shortcuts[unicode.ToUpper(ev.Ch)]
			if ok {
				return pos
			}
			continue
		}

		switch ev.Key {
		case termbox.KeyEsc, termbox.KeyCtrlC:
			return MENU_ESCAPE
		case termbox.KeyEnter, termbox.KeySpace:
			return active
		case termbox.KeyArrowUp:
			active = (active + len(entries) - 1) % len(entries)
			for entries[active] == "" {
				active = (active + len(entries) - 1) % len(entries)
			}
		case termbox.KeyArrowDown:
			active = (active + 1) % len(entries)
			for entries[active] == "" {
				active = (active + 1) % len(entries)
			}
		}
	}
}
