package main

import (
    "time"
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

func MakeTicker() (ticks chan struct{}, closer chan struct{}) {
	ticks = make(chan struct{}, 0)
	closer = make(chan struct{}, 0)

	go func() {
        defer close(ticks)
		for {
            time.Sleep(time.Duration(SECONDS_PER_TICK) * time.Second)

            select {
            case <-closer:
                return
            default:
                ticks <- struct{}{}
            }
		}
	}()

    return
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
