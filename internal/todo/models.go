package todo

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type Task struct {
	ID                 string  `json:"id"`
	ParentID           *string `json:"parentId"`
	RootID             string  `json:"rootId"`
	Title              string  `json:"title"`
	Description        *string `json:"description"`
	Date               *string `json:"date"`
	OriginalDate       *string `json:"originalDate"`
	Status             string  `json:"status"`
	Color              *string `json:"color"`
	Order              int     `json:"order"`
	CompletedAt        *int64  `json:"completedAt"`
	Postpone           *string `json:"postpone"`
	SeriesID           *string `json:"seriesId"`
	RuleID             *string `json:"ruleId"`
	OccurrenceDate     *string `json:"occurrenceDate"`
	Exception          bool    `json:"exception"`
	RemindAt           *int64  `json:"remindAt"`
	ReminderTime       *string `json:"reminderTime"`
	ReminderZone       *string `json:"reminderZone"`
	ReminderGeneration int     `json:"reminderGeneration"`
	Version            int64   `json:"version,string"`
	CreatedAt          int64   `json:"createdAt"`
	UpdatedAt          int64   `json:"updatedAt"`
}
type Query struct {
	View  string `json:"view"`
	Date  string `json:"date"`
	End   string `json:"end"`
	Limit int    `json:"limit"`
}
type Snapshot struct {
	Tasks     []Task         `json:"tasks"`
	Counts    map[string]int `json:"counts"`
	Revision  string         `json:"revision"`
	Today     string         `json:"today"`
	Truncated bool           `json:"truncated"`
}
type Command struct {
	RepeatRule      *Rule   `json:"repeatRule,omitempty"`
	Following       bool    `json:"following,omitempty"`
	CommandID       string  `json:"commandId"`
	ID              string  `json:"id"`
	RootID          string  `json:"rootId"`
	ParentID        *string `json:"parentId"`
	ExpectedVersion string  `json:"expectedVersion"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Date            *string `json:"date"`
	Color           string  `json:"color"`
	Postpone        string  `json:"postpone"`
	RemindAt        *int64  `json:"remindAt"`
	ReminderTime    *string `json:"reminderTime"`
	Completed       bool    `json:"completed"`
	BeforeID        string  `json:"beforeId"`
	AfterID         string  `json:"afterId"`
}
type Rule struct {
	ID               string  `json:"id"`
	SeriesID         string  `json:"seriesId"`
	From             string  `json:"from"`
	Until            *string `json:"until"`
	Anchor           string  `json:"anchor"`
	Frequency        string  `json:"frequency"`
	Interval         int     `json:"interval"`
	Weekdays         []int   `json:"weekdays"`
	MonthDay         int     `json:"monthDay"`
	Template         []Task  `json:"template"`
	GeneratedThrough string  `json:"generatedThrough"`
}
type RepeatRequest struct {
	Command   Command `json:"command"`
	Rule      Rule    `json:"rule"`
	Following bool    `json:"following"`
}
type Focus struct {
	ID         string  `json:"id"`
	TaskID     *string `json:"taskId"`
	Title      string  `json:"title"`
	Duration   int     `json:"duration"`
	Remaining  int     `json:"remaining"`
	Deadline   int64   `json:"deadline"`
	State      string  `json:"state"`
	StartedAt  int64   `json:"startedAt"`
	FinishedAt *int64  `json:"finishedAt"`
}
type Countdown struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	AllDay     bool   `json:"allDay"`
	AnchorDate string `json:"anchorDate"`
	AnchorTime string `json:"anchorTime"`
	Target     int64  `json:"target"`
	Zone       string `json:"zone"`
	Frequency  string `json:"frequency"`
	Enabled    bool   `json:"enabled"`
	Offset     int    `json:"offset"`
	Generation int    `json:"generation"`
}
type Notice struct {
	Silent     bool   `json:"silent"`
	ID         string `json:"id"`
	SourceType string `json:"sourceType"`
	SourceID   string `json:"sourceId"`
	State      string `json:"state"`
	FireAt     int64  `json:"fireAt"`
	Title      string `json:"title"`
}
type Holiday struct {
	Date    string `json:"date"`
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Source  string `json:"source"`
	Version string `json:"version"`
}
type WindowState struct {
	Key       string  `json:"key"`
	X         int     `json:"x"`
	Y         int     `json:"y"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	Mode      string  `json:"mode"`
	Opacity   float64 `json:"opacity"`
	Collapsed bool    `json:"collapsed"`
}
type Backup struct {
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	CreatedAt int64  `json:"createdAt"`
}
type Series struct {
	ID          string  `json:"id"`
	StoppedFrom *string `json:"stoppedFrom"`
	Version     int64   `json:"version"`
}
type Job struct {
	ID         string `json:"id"`
	SourceType string `json:"sourceType"`
	SourceID   string `json:"sourceId"`
	Generation int    `json:"generation"`
	CycleKey   string `json:"cycleKey"`
	FireAt     int64  `json:"fireAt"`
	State      string `json:"state"`
	Title      string `json:"title"`
}
type Transfer struct {
	SourceStart *string           `json:"sourceStart,omitempty"`
	Series      []Series          `json:"series"`
	Jobs        []Job             `json:"notificationJobs"`
	Version     int               `json:"formatVersion"`
	AppVersion  string            `json:"appVersion"`
	ExportedAt  int64             `json:"exportedAt"`
	Tasks       []Task            `json:"tasks"`
	Rules       []Rule            `json:"rules"`
	Focus       []Focus           `json:"focus"`
	Countdowns  []Countdown       `json:"countdowns"`
	Settings    map[string]string `json:"settings"`
}
type Preview struct {
	Token     string   `json:"token"`
	Tasks     []Task   `json:"tasks"`
	Roots     int      `json:"roots"`
	Nodes     int      `json:"nodes"`
	Warnings  []string `json:"warnings"`
	Replacing bool     `json:"replacing"`
}
type pendingImport struct {
	Data     Transfer
	Expires  time.Time
	Revision string
	Replace  bool
}

func ID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func ptr[T any](v T) *T { return &v }
func val[T any](v *T) (z T) {
	if v != nil {
		return *v
	}
	return
}
func fail(code, msg string) error { return fmt.Errorf("%s: %s", code, msg) }
func today(t time.Time) string    { return t.Format("2006-01-02") }
