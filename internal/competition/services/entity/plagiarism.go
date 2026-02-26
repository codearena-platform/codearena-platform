package entity

import "time"

// BotSignature represents the structural fingerprint of a Bot's code.
// It is used by the Plagiarism Engine to detect copy-paste or slightly modified bots.
type BotSignature struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	BotID     string    `json:"bot_id" bson:"bot_id"`
	HashTree  []uint32  `json:"hash_tree" bson:"hash_tree"` // Ordered rolling hashes (Winnowing)
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

// PlagiarismReport details the findings when a Bot is compared against the database.
type PlagiarismReport struct {
	IsPlagiarized  bool    `json:"is_plagiarized"`
	Similarity     float64 `json:"similarity"`      // 0.0 to 1.0 (Percentage of overlap)
	MatchedBotID   string  `json:"matched_bot_id"`  // ID of the bot it was copied from (if any)
	MatchedLines   int     `json:"matched_lines"`
}
