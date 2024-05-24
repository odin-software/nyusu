package main

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand/v2"
	"strconv"
)

func GetNewHash() string {
	r := strconv.FormatFloat(rand.Float64(), 'f', -1, 64)
	h := sha256.New()
	h.Write([]byte(r))

	return hex.EncodeToString(h.Sum((nil)))
}
