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