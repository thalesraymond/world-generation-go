package simulation

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestFormatEventWithFigureID(t *testing.T) {
	e := Event{
		Year:        100,
		Category:    "Politics",
		Description: "Description",
		FigureID:    "fig-1",
	}

	got := FormatEvent(e)
	want := "[100] (Politics) fig-1: Description"

	if got != want {
		t.Errorf("FormatEvent() = %q, want %q", got, want)
	}
}

func TestFormatEventWithoutFigureFields(t *testing.T) {
	e := Event{
		Year:        100,
		Category:    "War",
		Description: "The siege began.",
	}

	got := FormatEvent(e)
	want := "[100] (War) The siege began."

	if got != want {
		t.Errorf("FormatEvent() = %q, want %q", got, want)
	}
}

func TestEventJSONRoundTripWithFigureFields(t *testing.T) {
	e := Event{
		Year:           105,
		Category:       "Politics",
		Description:    "A figure rose to power.",
		FigureID:       "fig-2",
		RelatedFigures: []string{"fig-3", "fig-4"},
		SettlementName: "Aldcrest",
	}

	data, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var got Event
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(got, e) {
		t.Errorf("round-trip mismatch = %+v, want %+v", got, e)
	}
}

func TestEventJSONRoundTripWithoutFigureFields(t *testing.T) {
	e := Event{
		Year:        110,
		Category:    "Disaster",
		Description: "A flood swept the valley.",
	}

	data, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	for _, field := range []string{"figureID", "relatedFigures", "settlementName"} {
		if strings.Contains(string(data), field) {
			t.Errorf("json.Marshal() output %q unexpectedly contains field %q", string(data), field)
		}
	}

	var got Event
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(got, e) {
		t.Errorf("round-trip mismatch = %+v, want %+v", got, e)
	}
}

func TestEventBackwardCompat(t *testing.T) {
	oldJSON := `{"year":100,"category":"War","description":"The siege began."}`

	var got Event
	if err := json.Unmarshal([]byte(oldJSON), &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	want := Event{
		Year:        100,
		Category:    "War",
		Description: "The siege began.",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unmarshaled event = %+v, want %+v", got, want)
	}
}
