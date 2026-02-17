package postgres

import (
	"database/sql"

	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
)

type PGScoreRepository struct {
	db *sql.DB
}

func NewPGScoreRepository(db *sql.DB) *PGScoreRepository {
	return &PGScoreRepository{db: db}
}

func (r *PGScoreRepository) UpdateScore(matchID string, participantID string, points int) error {
	// For simplicity, we might just log this or have a match_scores table.
	// Let's assume a simple tournament_scores table for the challenge.
	// But the interface says UpdateScore(matchID...).
	// Let's implement a basic accumulation if we had a challenge context.
	// For this MVP, we might just store it in a 'scores' table linked to user/bot.

	query := `
		INSERT INTO scores (match_id, participant_id, points)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(query, matchID, participantID, points)
	return err
}

func (r *PGScoreRepository) GetLeaderboard(challengeID string) ([]*entity.Participant, error) {
	// Aggregate scores from the scores table
	// This assumes valid SQL schema provided in connection.go or migration
	query := `
		SELECT participant_id, SUM(points) as total_score
		FROM scores
		GROUP BY participant_id
		ORDER BY total_score DESC
		LIMIT 100
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*entity.Participant
	for rows.Next() {
		var p entity.Participant
		err := rows.Scan(&p.ID, &p.Score)
		if err != nil {
			return nil, err
		}
		result = append(result, &p)
	}
	return result, nil
}
