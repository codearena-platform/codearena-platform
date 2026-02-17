package entity

type MatchStatus string

const (
	MatchReady     MatchStatus = "READY"
	MatchRunning   MatchStatus = "RUNNING"
	MatchFinished  MatchStatus = "FINISHED"
	MatchCompleted MatchStatus = "COMPLETED"
	MatchFailed    MatchStatus = "FAILED"
)

type Match struct {
	ID              string
	TournamentID    string // Optional: link to a tournament
	DivisionID      string // Optional: link to a league division
	ChallengeID     string
	Status          MatchStatus
	Participants    []string
	WinnerID        string
	StartedAt       int64
	EndedAt         int64
	FinishedAt      int64
	EngineURL       string         // The URL of the simulation engine that ran the match
	ReplayAvailable bool           // Whether a detailed replay is available
	Metadata        map[string]any // For telemetry (energy, damage, etc.)
}

type Participant struct {
	ID     string
	UserID string
	Name   string
	Score  int
}
