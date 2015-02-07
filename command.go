package main

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	VALID_COMMANDS    = "LRASMPHK%="
	COMMANDS_WITH_ARG = "LRA"

	COMMAND_HELP = `
        <aircraft>A0     aproach airport
        <aircraft>A<1-5> assign altitude
        <aircraft>M      maintain current altitude
        <aircraft>L<0-4> turn left
        <aircraft>R<0-4> turn right
        <aircraft>P      proceed on current heading
        <aircraft>H      hold at navaid
        <aircraft>K      keep current position
        <aircraft><airport>
                         turn towards airport at navaid

        <aircraft>S      status of aircraft

        Esc              quit game
        .                advance time
        ?                show help
    `
)

type Command struct {
	callsign rune
	command  rune
	arg      int
}

func (c *Command) Apply(p *Plane) string {
	if c.command == 'S' && p.IsActive() {
		return p.State()
	}

	if !p.AcceptsCommands() {
		return "---------"
	}

	var res bool

	switch c.command {
	case 'L': // turn left 0-4
		res = p.DoTurn(-c.arg)
	case 'R': // turn right 0-4
		res = p.DoTurn(c.arg)
	case 'A': // change altitude 0-5 (0: aproach)
		res = p.DoHeight(c.arg)
	case 'S':
		return p.State()
	case 'M': // maintain current altitude
		res = p.DoHeight(p.height)
	case 'P': // proceed current heading
		res = p.DoTurn(0)
	case 'H': // hold at navaid
		res = p.DoHold()
	case 'K': // keep position
		res = p.DoKeep()
	case '%', '=':
		res = p.TurnAtNavaid(c.command)
	default:
		panic("should not happen")
	}

	if res {
		return "Roger"
	} else {
		return "Unable"
	}
}

type CommandInterpreter struct {
	game *GameState

	buf   string
	last  string
	reply string
}

func (ci *CommandInterpreter) KeyPressed(key rune) {
	ci.buf += string(key)
	ci.try_command()
}

func (ci *CommandInterpreter) Clear() {
	ci.buf = ""
	ci.last = ""
	ci.reply = ""
}

func (ci CommandInterpreter) StatusLine() string {
	if len(ci.buf) > 0 {
		return ci.buf
	} else {
		return fmt.Sprintf("%s %s", ci.last, ci.reply)
	}
}

func (ci *CommandInterpreter) try_command() {
	valid, complete, cmd := parse_command(ci.buf)

	if !valid {
		ci.reply = "--- Say Again? ---"
		ci.last = ci.buf
		ci.buf = ""
		return
	}

	if !complete {
		return
	}

	ci.last = ci.buf
	ci.buf = ""

	var plane *Plane
	for _, p := range ci.game.planes {
		if p.callsign == cmd.callsign {
			plane = p
			break
		}
	}

	if plane == nil {
		ci.reply = "-----------"
	} else {
		ci.reply = cmd.Apply(plane)
		ci.game.last_commanded_plane = plane
	}

}

func parse_command(s string) (valid bool, complete bool, cmd Command) {
	for pos, char := range s {
		switch pos {
		case 0:
			if char < 'A' || char > 'Z' {
				return
			}
			cmd.callsign = char
		case 1:
			if !strings.ContainsRune(VALID_COMMANDS, char) {
				return
			}
			cmd.command = char
			complete = !strings.ContainsRune(COMMANDS_WITH_ARG, char)
		case 2:
			arg, err := strconv.Atoi(string(char))
			if err != nil || arg < 0 || arg > 9 || complete {
				// command without args should not go here
				return
			}
			cmd.arg = arg
			complete = true
		default:
			panic("should not happen")
		}
	}

	valid = true
	return
}
