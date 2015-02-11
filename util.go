package main

import (
	crand "crypto/rand"
	"math/rand"
	"strings"
)

func Abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func Max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func RandSeed() int64 {
	rbuf := make([]byte, 4)
	_, err := crand.Read(rbuf)
	if err != nil {
		panic(err)
	}
	var seed int64 = int64(rbuf[0])<<24 + int64(rbuf[1])<<16 + int64(rbuf[2])<<8 + int64(rbuf[3])
	return seed
}

// random value from [a, b]
func RandRange(r *rand.Rand, a, b int) int {
	return r.Intn(b-a+1) + a
}

func ChoosePlaneType(r *rand.Rand, a []*PlaneType) *PlaneType {
	count := 0
	for _, e := range a {
		count += e.weight
	}

	val := r.Intn(count)
	for _, e := range a {
		val -= e.weight
		if val < 0 {
			return e
		}
	}
	panic("should not happen")
}

func ChooseRoute(r *rand.Rand, a []Route) Route {
	count := 0
	for _, e := range a {
		count += e.weight
	}

	if count == 0 {
		panic("no route to choose from")
	}

	val := r.Intn(count)
	for _, e := range a {
		val -= e.weight
		if val < 0 {
			return e
		}
	}
	panic("should not happen")
}

// split and deindent lines (1st line as reference)
func SplitLines(s string) []string {
	lines := strings.Split(s, "\n")
	for lines[0] == "" {
		lines = lines[1:]
	}

	l0 := lines[0]

	var deindent int
	for deindent = 0; deindent < len(l0) && l0[deindent] == ' '; deindent += 1 {
	}

	for n, _ := range lines {
		if len(lines[n]) >= deindent {
			lines[n] = lines[n][deindent:]
		} else {
			lines[n] = ""
		}
	}
	return lines
}

func FirstRune(s string) rune {
	for _, r := range s {
		return r
	}
	return 0
}

const PAD_SPACE = "                                                              "

func Pad(width int, left string, right string) string {
	pad := Max(0, width-len(left)-len(right))
	return left + PAD_SPACE[0:pad] + right
}
