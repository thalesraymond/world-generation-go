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

func TestFormatEventWithTargetSettlement(t *testing.T) {
	e := Event{
		Year:             100,
		Category:         "Raid",
		Description:      "Alpha raided Beta",
		SettlementName:   "Alpha",
		TargetSettlement: "Beta",
	}

	got := FormatEvent(e)
	want := "[100] (Raid) Alpha → Beta: Alpha raided Beta"

	if got != want {
		t.Errorf("FormatEvent() = %q, want %q", got, want)
	}
}

func TestFormatEventWithoutTargetSettlementBackwardCompat(t *testing.T) {
	e := Event{
		Year:        100,
		Category:    "Economy",
		Description: "Alpha prospers",
	}

	got := FormatEvent(e)
	want := "[100] (Economy) Alpha prospers"

	if got != want {
		t.Errorf("FormatEvent() = %q, want %q", got, want)
	}
}

func TestEventJSONRoundTripWithTargetSettlement(t *testing.T) {
	e := Event{
		Year:             105,
		Category:         "Conquest",
		Description:      "Alpha conquered Beta",
		SettlementName:   "Alpha",
		TargetSettlement: "Beta",
	}

	data, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	if !strings.Contains(string(data), "targetSettlement") {
		t.Errorf("json.Marshal() output %q missing targetSettlement", string(data))
	}

	var got Event
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(got, e) {
		t.Errorf("round-trip mismatch = %+v, want %+v", got, e)
	}
}

func TestEventJSONRoundTripWithoutTargetSettlement(t *testing.T) {
	e := Event{
		Year:        110,
		Category:    "Economy",
		Description: "Alpha prospers",
	}

	data, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	if strings.Contains(string(data), "targetSettlement") {
		t.Errorf("json.Marshal() output %q unexpectedly contains targetSettlement", string(data))
	}

	var got Event
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(got, e) {
		t.Errorf("round-trip mismatch = %+v, want %+v", got, e)
	}
}

func TestEventJSONRoundTripWithIDAndArtifactID(t *testing.T) {
	e := Event{
		Year:        42,
		Category:    "War",
		Description: "A great battle.",
		ID:          "event-42-0",
		ArtifactID:  "artifact-settlement-0",
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

func TestEventJSONRoundTripOmitsIDAndArtifactID(t *testing.T) {
	e := Event{
		Year:        100,
		Category:    "Politics",
		Description: "A figure rose.",
	}

	data, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	for _, field := range []string{"\"id\"", "\"artifactID\""} {
		if strings.Contains(string(data), field) {
			t.Errorf("json.Marshal() output %q unexpectedly contains %s", string(data), field)
		}
	}
}

func TestEventBackwardCompatWithIDFields(t *testing.T) {
	oldJSON := `{"year":100,"category":"War","description":"The siege began."}`

	var got Event
	if err := json.Unmarshal([]byte(oldJSON), &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if got.ID != "" {
		t.Errorf("ID = %q, want empty", got.ID)
	}
	if got.ArtifactID != "" {
		t.Errorf("ArtifactID = %q, want empty", got.ArtifactID)
	}
}
