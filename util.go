package main

import (
	crand "crypto/rand"
	"math/rand"
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

	val := r.Intn(count)
	for _, e := range a {
		val -= e.weight
		if val < 0 {
			return e
		}
	}
	panic("should not happen")
}
