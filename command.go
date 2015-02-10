package main

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	COMMANDS_WITHOUT_ARG = "SMPHK%="
	COMMANDS_WITH_ARG    = "LRA"

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
        Tab              show planes (if enabled)
    `
)

type Command struct {
	valid   bool
	delayed int

	callsign rune
	command  rune
	arg      int
}

func (c *Command) Apply(p *Plane) string {
	if !c.valid {
		return "--- Say Again? ---"
	}

	if p == nil {
		return "---------"
	}

	if c.command == 'S' && p.IsActive() {
		return p.StateMessage()
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
	// case 'S': handled above
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
	setup GameSetup

	buf   string
	last  string
	reply string

	delayed_commands []*Command

	last_commanded_plane *Plane
}

func (ci *CommandInterpreter) KeyPressed(g *GameState, key rune) {
	ci.buf += string(key)

	cmd := ci.parse_command(ci.buf)

	if cmd == nil {
		// incomplete
		return
	}

	ci.last = ci.buf
	ci.buf = ""

	if cmd.delayed > 0 {
		ci.delayed_commands = append(ci.delayed_commands, cmd)
	} else {
		plane := g.FindPlane(cmd.callsign)
		ci.reply = cmd.Apply(plane)
		ci.last_commanded_plane = plane
	}
}

func (ci *CommandInterpreter) Tick(g *GameState) {
	ci.last_commanded_plane = nil

	// TODO: old commands stay in delayed_commands with delayed < 0
	for _, cmd := range ci.delayed_commands {
		cmd.delayed -= 1
		if cmd.delayed == 0 {
			plane := g.FindPlane(cmd.callsign)
			_ = cmd.Apply(plane) // apply silently
		}
	}
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

func (ci *CommandInterpreter) parse_command(s string) *Command {
	var cmd Command
	var state int

	for _, char := range s {
		switch {
		case state == 0 && char == '.':
			cmd.delayed += 1
		case state == 0 && char >= 'A' && char <= 'Z':
			state = 1
			cmd.callsign = char
		case state == 1 && strings.ContainsRune(COMMANDS_WITHOUT_ARG, char):
			cmd.command = char
			cmd.valid = true
			return &cmd
		case state == 1 && strings.ContainsRune(COMMANDS_WITH_ARG, char):
			cmd.command = char
			state = 2
		case state == 2 && char >= '0' && char <= '9':
			arg, _ := strconv.Atoi(string(char))
			cmd.arg = arg
			cmd.valid = true
			return &cmd
		default:
			// valid == false
			return &cmd
		}
	}
	return nil
}
