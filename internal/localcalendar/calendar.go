package localcalendar

import "encoding/json"

func calendarIDsJSON(ids []string) []byte {
	if ids == nil {
		ids = []string{}
	}
	b, _ := json.Marshal(ids)
	return b
}

type Calendar struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Source string `json:"source"`
}
type Event struct {
	ID         string `json:"id"`
	CalendarID string `json:"calendarId"`
	Title      string `json:"title"`
	Location   string `json:"location"`
	Start      int64  `json:"start"`
	End        int64  `json:"end"`
	AllDay     bool   `json:"allDay"`
}
type Result struct {
	Calendars []Calendar `json:"calendars"`
	Events    []Event    `json:"events"`
	Error     string     `json:"error,omitempty"`
}
