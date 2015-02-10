package main

import (
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"unicode"
)

var _ = fmt.Println

type MenuEntry struct {
	key    rune
	text   string
	textf  func() string
	action func()
}

type ECloseMenu struct{}

func RunMenu(title string, entries []MenuEntry) {

	close_menu := false
	defer func() {
		if e := recover(); e != nil {
			if _, ok := e.(ECloseMenu); ok {
				close_menu = true
			} else {
				panic(e)
			}
		}
	}()

	active_entry := 0
menu:
	for !close_menu {
		lines := make([]string, 0, len(entries))
		colors := make([]termbox.Attribute, 0, len(entries))

		for nr, e := range entries {
			if e.text != "" {
				lines = append(lines, e.text)
			} else if e.textf != nil {
				lines = append(lines, e.textf())
			} else {
				lines = append(lines, "")
			}

			if nr == active_entry {
				colors = append(colors, termbox.ColorDefault|termbox.AttrReverse)
			} else {
				colors = append(colors, termbox.ColorDefault)
			}
		}

		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		DrawWindow(title, "", lines, colors)
		termbox.Flush()

		select {
		case ev := <-events:
			if ev.Type != termbox.EventKey {
				continue
			}

			if ev.Ch != 0 {
				ch := unicode.ToUpper(ev.Ch)
				for _, e := range entries {
					if e.key == ch {
						e.action()
						active_entry = 0
						continue menu
					}
				}
				continue menu // no action found
			}

			switch ev.Key {
			case termbox.KeyEsc, termbox.KeyCtrlC:
				close_menu = true
			case termbox.KeyEnter:
				entries[active_entry].action()
				active_entry = 0
			case termbox.KeyArrowUp:
				active_entry = (active_entry + len(entries) - 1) % len(entries)
				for entries[active_entry].action == nil {
					active_entry = (active_entry + len(entries) - 1) % len(entries)
				}
			case termbox.KeyArrowDown:
				active_entry = (active_entry + 1) % len(entries)
				for entries[active_entry].action == nil {
					active_entry = (active_entry + 1) % len(entries)
				}
			}

		case <-sigterm:
			return
		default:
			// nothing
		}
	}
}

func CloseMenu() {
	panic(ECloseMenu{})
}

func MenuBoolText(key rune, v *bool, textt string, textf string) MenuEntry {
	return MenuEntry{
		key: key,
		textf: func() string {
			if *v {
				return textt
			} else {
				return textf
			}
		},
		action: func() {
			*v = !*v
		},
	}
}
