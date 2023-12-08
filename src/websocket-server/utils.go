package main

import (
	"crypto/rand"
	"encoding/binary"
)

func genRandNum(min, max int) (int, error) {
	var num int
	err := binary.Read(rand.Reader, binary.LittleEndian, &num)
	return int(num*(max-min) + min), err
}
