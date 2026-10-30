package localcalendar

import (
	"encoding/json"
	"testing"
)

func TestEmptyCalendarIDsAreArrayNotNull(t *testing.T) {
	if string(calendarIDsJSON(nil)) != "[]" {
		t.Fatal("native Objective-C needs an array")
	}
	var out []string
	if err := json.Unmarshal(calendarIDsJSON([]string{"a\"b", "c"}), &out); err != nil || len(out) != 2 || out[0] != "a\"b" {
		t.Fatal(out, err)
	}
}
