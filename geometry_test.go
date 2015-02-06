package main

import "testing"
import "fmt"

var _ = fmt.Println

func TestDirection(t *testing.T) {
	test := func(d Direction, c int, expected Direction) {
		res := d.Right(c)
		if res != expected {
			t.Error(d, c, expected, "!=", res, int(res))
		}

	}
	test(DIR_N, 1, DIR_NO)
	test(DIR_N, 8, DIR_N)
	test(DIR_N, -1, DIR_NW)
	test(DIR_N, -4, DIR_S)
	test(DIR_N, 4, DIR_S)
}

func TestPosition(t *testing.T) {
	// 8 times to the right goes back to start
	p := Position{0, 0}
	for _, d := range DIRECTIONS {
		p = p.Move(d, 1)
	}
	if p != (Position{0, 0}) {
		t.Error(p)
	}
}

func TestPosDirection(t *testing.T) {
	p := Position{5, 6}

	for _, d := range DIRECTIONS {
		p2 := p.Move(d, 4)
		d2, _, ok := p.Direction(p2)
		if !ok {
			t.Error(ok)
		}
		if d != d2 {
			t.Error(d, d2)
		}
	}
}
