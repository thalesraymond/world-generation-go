package exporter

import "testing"

func TestFrontmatterEmpty(t *testing.T) {
	result := frontmatter(nil)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestFrontmatterSingleField(t *testing.T) {
	result := frontmatter([]field{{"key", "value"}})
	expected := "---\nkey: value\n---\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFrontmatterMultipleFieldsDeterministicOrder(t *testing.T) {
	result := frontmatter([]field{
		{"alpha", "first"},
		{"beta", "second"},
		{"gamma", "third"},
	})
	expected := "---\nalpha: first\nbeta: second\ngamma: third\n---\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestQuoteIfNeededSpecialChars(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"foo:bar", `"foo:bar"`},
		{"foo#bar", `"foo#bar"`},
		{"foo[bar]", `"foo[bar]"`},
	}

	for _, tc := range cases {
		actual := quoteIfNeeded(tc.input)
		if actual != tc.expected {
			t.Errorf("quoteIfNeeded(%q) = %q, want %q", tc.input, actual, tc.expected)
		}
	}
}

func TestQuoteIfNeededSimpleString(t *testing.T) {
	result := quoteIfNeeded("simple")
	if result != "simple" {
		t.Errorf("expected simple string unchanged, got %q", result)
	}
}

func TestQuoteIfNeededSpecialYAMLValues(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"true", `"true"`},
		{"false", `"false"`},
		{"null", `"null"`},
	}

	for _, tc := range cases {
		actual := quoteIfNeeded(tc.input)
		if actual != tc.expected {
			t.Errorf("quoteIfNeeded(%q) = %q, want %q", tc.input, actual, tc.expected)
		}
	}
}

func TestQuoteIfNeededEmptyString(t *testing.T) {
	result := quoteIfNeeded("")
	if result != `""` {
		t.Errorf("expected quoted empty string, got %q", result)
	}
}
