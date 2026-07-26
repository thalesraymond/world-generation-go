package narrative

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	validCases := []struct {
		name     string
		input    string
		expected *Grammar
	}{
		{
			name:  "simple rule",
			input: `event ::= "something happened"`,
			expected: &Grammar{
				Rules: map[string]Rule{
					"event": {
						Name: "event",
						Alternatives: []Alternative{
							{Terminal{Text: "something happened"}},
						},
					},
				},
			},
		},
		{
			name:  "multiple alternatives",
			input: `event ::= "option A" | "option B"`,
			expected: &Grammar{
				Rules: map[string]Rule{
					"event": {
						Name: "event",
						Alternatives: []Alternative{
							{Terminal{Text: "option A"}},
							{Terminal{Text: "option B"}},
						},
					},
				},
			},
		},
		{
			name: "non-terminal references",
			input: `event ::= <subject> " did " <action>
subject ::= "the king" | "a warrior"
action ::= "fight" | "flee"`,
			expected: &Grammar{
				Rules: map[string]Rule{
					"event": {
						Name: "event",
						Alternatives: []Alternative{
							{NonTerminal{Name: "subject"}, Terminal{Text: " did "}, NonTerminal{Name: "action"}},
						},
					},
					"subject": {
						Name: "subject",
						Alternatives: []Alternative{
							{Terminal{Text: "the king"}},
							{Terminal{Text: "a warrior"}},
						},
					},
					"action": {
						Name: "action",
						Alternatives: []Alternative{
							{Terminal{Text: "fight"}},
							{Terminal{Text: "flee"}},
						},
					},
				},
			},
		},
		{
			name:  "variables",
			input: `event ::= $name " did " $action " in " $location`,
			expected: &Grammar{
				Rules: map[string]Rule{
					"event": {
						Name: "event",
						Alternatives: []Alternative{
							{Variable{Name: "name"}, Terminal{Text: " did "}, Variable{Name: "action"}, Terminal{Text: " in "}, Variable{Name: "location"}},
						},
					},
				},
			},
		},
		{
			name: "multi-line rules",
			input: `story ::= "Once upon a time"
        | "In a land far away"`,
			expected: &Grammar{
				Rules: map[string]Rule{
					"story": {
						Name: "story",
						Alternatives: []Alternative{
							{Terminal{Text: "Once upon a time"}},
							{Terminal{Text: "In a land far away"}},
						},
					},
				},
			},
		},
		{
			name: "mixed symbols",
			input: `event ::= <char_intro> " approached the " $landmark
char_intro ::= $name " the warrior"`,
			expected: &Grammar{
				Rules: map[string]Rule{
					"event": {
						Name: "event",
						Alternatives: []Alternative{
							{NonTerminal{Name: "char_intro"}, Terminal{Text: " approached the "}, Variable{Name: "landmark"}},
						},
					},
					"char_intro": {
						Name: "char_intro",
						Alternatives: []Alternative{
							{Variable{Name: "name"}, Terminal{Text: " the warrior"}},
						},
					},
				},
			},
		},
		{
			name: "comments",
			input: `# This is a comment
event ::= "something" # inline comment?`,
			expected: &Grammar{
				Rules: map[string]Rule{
					"event": {
						Name: "event",
						Alternatives: []Alternative{
							{Terminal{Text: "something"}},
						},
					},
				},
			},
		},
		{
			name:  "escape sequences in strings",
			input: `event ::= "He said: \"Hello\""`,
			expected: &Grammar{
				Rules: map[string]Rule{
					"event": {
						Name: "event",
						Alternatives: []Alternative{
							{Terminal{Text: `He said: "Hello"`}},
						},
					},
				},
			},
		},
	}

	for _, tc := range validCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("parsed grammar does not match expected\ngot:      %#v\nexpected: %#v", got, tc.expected)
			}
		})
	}
}

func TestParse_Invalid(t *testing.T) {
	invalidCases := []struct {
		name  string
		input string
	}{
		{
			name:  "missing definition separator",
			input: `event "something"`,
		},
		{
			name:  "empty rule name",
			input: ` ::= "something"`,
		},
		{
			name: "duplicate rule names",
			input: `event ::= "first"
event ::= "second"`,
		},
		{
			name:  "empty grammar with only comments",
			input: `# just comments`,
		},
		{
			name:  "rule with no alternatives",
			input: `event ::=`,
		},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Parse(tc.input); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestParse_Determinism(t *testing.T) {
	input := `event ::= <subject> " did " <action>
subject ::= "the king" | "a warrior"
action ::= "fight" | "flee"`

	first, err := Parse(input)
	if err != nil {
		t.Fatalf("first parse failed: %v", err)
	}

	second, err := Parse(input)
	if err != nil {
		t.Fatalf("second parse failed: %v", err)
	}

	if !reflect.DeepEqual(first, second) {
		t.Errorf("parsing the same grammar twice produced different results\nfirst:  %#v\nsecond: %#v", first, second)
	}
}
