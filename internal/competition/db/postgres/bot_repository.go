package postgres

import (
	"database/sql"
	"fmt"

	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
	_ "github.com/lib/pq"
)

type PGBotRepository struct {
	db *sql.DB
}

func NewPGBotRepository(db *sql.DB) *PGBotRepository {
	return &PGBotRepository{db: db}
}

func (r *PGBotRepository) Save(bot *entity.Bot) error {
	query := `
		INSERT INTO bots (id, user_id, name, avatar_url, rating, matches, wins, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			rating = EXCLUDED.rating,
			matches = EXCLUDED.matches,
			wins = EXCLUDED.wins
	`
	_, err := r.db.Exec(query,
		bot.ID, bot.UserID, bot.Name, bot.AvatarURL,
		bot.Rating, bot.Matches, bot.Wins, bot.CreatedAt,
	)
	return err
}

func (r *PGBotRepository) GetByID(id string) (*entity.Bot, error) {
	query := `SELECT id, user_id, name, avatar_url, rating, matches, wins, created_at FROM bots WHERE id = $1`
	row := r.db.QueryRow(query, id)

	var bot entity.Bot
	err := row.Scan(
		&bot.ID, &bot.UserID, &bot.Name, &bot.AvatarURL,
		&bot.Rating, &bot.Matches, &bot.Wins, &bot.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("bot not found")
		}
		return nil, err
	}
	return &bot, nil
}

func (r *PGBotRepository) GetByUserID(userID string) ([]*entity.Bot, error) {
	query := `SELECT id, user_id, name, avatar_url, rating, matches, wins, created_at FROM bots WHERE user_id = $1`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bots []*entity.Bot
	for rows.Next() {
		var bot entity.Bot
		err := rows.Scan(
			&bot.ID, &bot.UserID, &bot.Name, &bot.AvatarURL,
			&bot.Rating, &bot.Matches, &bot.Wins, &bot.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		bots = append(bots, &bot)
	}
	return bots, nil
}

func (r *PGBotRepository) GetTop(limit int) ([]*entity.Bot, error) {
	query := `
		SELECT id, user_id, name, avatar_url, rating, matches, wins, created_at 
		FROM bots 
		ORDER BY rating DESC 
		LIMIT $1
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bots []*entity.Bot
	for rows.Next() {
		var bot entity.Bot
		err := rows.Scan(
			&bot.ID, &bot.UserID, &bot.Name, &bot.AvatarURL,
			&bot.Rating, &bot.Matches, &bot.Wins, &bot.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		bots = append(bots, &bot)
	}
	return bots, nil
}

func (r *PGBotRepository) UpdateRating(botID string, newRating float64) error {
	query := `UPDATE bots SET rating = $1 WHERE id = $2`
	_, err := r.db.Exec(query, newRating, botID)
	return err
}
