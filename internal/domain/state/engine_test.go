package state

import "testing"

func TestGetPRNGReturnsSameSequenceForSameSeedAndComponent(t *testing.T) {
	engineA := NewEngine(42)
	engineB := NewEngine(42)

	prngA := engineA.GetPRNG("terrain")
	prngB := engineB.GetPRNG("terrain")

	for i := 0; i < 8; i++ {
		got := prngA.Uint64()
		want := prngB.Uint64()
		if got != want {
			t.Fatalf("draw %d mismatch: got %d want %d", i, got, want)
		}
	}
}

func TestSettlementScopedRNGIsolation(t *testing.T) {
	engine := NewEngine(42)

	rngA1 := engine.GetPRNG("settlement:Alpha")
	rngB := engine.GetPRNG("settlement:Beta")
	rngA2 := engine.GetPRNG("settlement:Alpha")

	for i := 0; i < 10; i++ {
		_ = rngB.Uint64()
	}

	for i := 0; i < 8; i++ {
		got := rngA1.Uint64()
		want := rngA2.Uint64()
		if got != want {
			t.Fatalf("settlement Alpha RNG changed after Beta usage: draw %d got %d want %d", i, got, want)
		}
	}
}

func TestGetPRNGKeepsComponentStreamsIndependent(t *testing.T) {
	engine := NewEngine(42)

	terrainA := engine.GetPRNG("terrain")
	weather := engine.GetPRNG("weather")
	terrainB := engine.GetPRNG("terrain")

	for i := 0; i < 4; i++ {
		_ = weather.Uint64()
	}

	for i := 0; i < 8; i++ {
		got := terrainA.Uint64()
		want := terrainB.Uint64()
		if got != want {
			t.Fatalf("terrain draw %d changed after weather usage: got %d want %d", i, got, want)
		}
	}
}
