package exporter

import "testing"

func TestNameTrackerSanitize(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"basic name", "Village", "Village"},
		{"less than char", "Vil<lage", "Village"},
		{"pipe and colon", "A|B:C", "ABC"},
		{"empty string", "", "unnamed"},
		{"whitespace only", "   ", "unnamed"},
		{"multiple spaces", "Alpha   Base", "Alpha Base"},
		{"trimmed output", "  Center  ", "Center"},
	}

	tracker := newNameTracker()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tracker.sanitize(tc.input)
			if actual != tc.expected {
				t.Errorf("sanitize(%q) = %q, want %q", tc.input, actual, tc.expected)
			}
		})
	}
}

func TestNameTrackerSameLowercaseNoSuffix(t *testing.T) {
	tracker := newNameTracker()

	tracker.sanitize("Haven")
	tracker.sanitize("haven")

	if tracker.used["haven"] != 2 {
		t.Errorf("expected second lowercase duplicate to increment count, got %d", tracker.used["haven"])
	}
}

func TestNameTrackerFlagsDuplicate(t *testing.T) {
	tracker := newNameTracker()

	tracker.sanitize("Ruin")

	if tracker.used["ruin"] != 1 {
		t.Errorf("expected first call to set count to 1, got %d", tracker.used["ruin"])
	}

	tracker.sanitize("Ruin")

	if tracker.used["ruin"] != 2 {
		t.Errorf("expected second call to set count to 2, got %d", tracker.used["ruin"])
	}
}
