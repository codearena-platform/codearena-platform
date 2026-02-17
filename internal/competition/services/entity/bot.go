package entity

type Bot struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Name      string  `json:"name"`
	AvatarURL string  `json:"avatar_url"`
	Rating    float64 `json:"rating"`
	Matches   int     `json:"matches"`
	Wins      int     `json:"wins"`
	CreatedAt int64   `json:"created_at"`
}

type BotVersion struct {
	ID        string `json:"id"`
	BotID     string `json:"bot_id"`
	Version   string `json:"version"` // e.g., "v1.2", "latest"
	Hash      string `json:"hash"`
	SourceURL string `json:"source_url"` // S3/Cloud Storage link
	CreatedAt int64  `json:"created_at"`
}

type MatchResult struct {
	MatchID     string             `json:"match_id"`
	WinnerID    string             `json:"winner_id"`
	RankingDiff map[string]float64 `json:"ranking_diff"` // bot_id -> points_gained
}
