package main

import (
	crand "crypto/rand"
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
