package state

import (
	"encoding/binary"
	"hash/fnv"
	randv2 "math/rand/v2"
)

// Engine deterministically derives isolated PRNG streams from a master seed.
type Engine struct {
	masterSeed uint64
}

// NewEngine records the master seed used for all component-specific PRNGs.
func NewEngine(masterSeed uint64) *Engine {
	return &Engine{masterSeed: masterSeed}
}

// GetPRNG returns a reproducible PRNG for the provided component identifier.
func (e *Engine) GetPRNG(componentID string) *randv2.Rand {
	seed1 := deriveSeed(e.masterSeed, componentID, "stream")
	seed2 := deriveSeed(e.masterSeed, componentID, "sequence")

	return randv2.New(randv2.NewPCG(seed1, seed2))
}

func deriveSeed(masterSeed uint64, componentID, lane string) uint64 {
	hasher := fnv.New64a()
	var seedBytes [8]byte

	binary.LittleEndian.PutUint64(seedBytes[:], masterSeed)
	_, _ = hasher.Write(seedBytes[:])
	_, _ = hasher.Write([]byte{0})
	_, _ = hasher.Write([]byte(componentID))
	_, _ = hasher.Write([]byte{0})
	_, _ = hasher.Write([]byte(lane))

	return hasher.Sum64()
}