package postgres

import (
	"database/sql"
	"encoding/json"

	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
)

type PGTournamentRepository struct {
	db *sql.DB
}

func NewPGTournamentRepository(db *sql.DB) *PGTournamentRepository {
	return &PGTournamentRepository{db: db}
}

func (r *PGTournamentRepository) Save(t *entity.Tournament) error {
	configJSON, _ := json.Marshal(t.Config)
	query := `
		INSERT INTO tournaments (id, name, type, status, description, config, created_at, hub_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			status = EXCLUDED.status,
			config = EXCLUDED.config,
			hub_id = EXCLUDED.hub_id
	`
	_, err := r.db.Exec(query, t.ID, t.Name, t.Type, t.Status, t.Description, configJSON, t.CreatedAt, t.HubID)
	return err
}

func (r *PGTournamentRepository) GetByID(id string) (*entity.Tournament, error) {
	query := `SELECT id, name, type, status, description, config, created_at, hub_id FROM tournaments WHERE id = $1`
	var configRaw []byte
	var t entity.Tournament
	var hubID sql.NullString
	err := r.db.QueryRow(query, id).Scan(&t.ID, &t.Name, &t.Type, &t.Status, &t.Description, &configRaw, &t.CreatedAt, &hubID)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(configRaw, &t.Config)
	if hubID.Valid {
		t.HubID = hubID.String
	}
	return &t, nil
}

func (r *PGTournamentRepository) ListByType(tType entity.TournamentType) ([]*entity.Tournament, error) {
	query := `SELECT id, name, type, status, description, config, created_at, hub_id FROM tournaments WHERE type = $1`
	rows, err := r.db.Query(query, tType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.Tournament
	for rows.Next() {
		var t entity.Tournament
		var configRaw []byte
		var hubID sql.NullString
		err := rows.Scan(&t.ID, &t.Name, &t.Type, &t.Status, &t.Description, &configRaw, &t.CreatedAt, &hubID)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(configRaw, &t.Config)
		if hubID.Valid {
			t.HubID = hubID.String
		}
		list = append(list, &t)
	}
	return list, nil
}

func (r *PGTournamentRepository) ListByHub(hubID string) ([]*entity.Tournament, error) {
	query := `SELECT id, name, type, status, description, config, created_at, hub_id FROM tournaments WHERE hub_id = $1`
	rows, err := r.db.Query(query, hubID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.Tournament
	for rows.Next() {
		var t entity.Tournament
		var configRaw []byte
		err := rows.Scan(&t.ID, &t.Name, &t.Type, &t.Status, &t.Description, &configRaw, &t.CreatedAt, &t.HubID)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(configRaw, &t.Config)
		list = append(list, &t)
	}
	return list, nil
}

func (r *PGTournamentRepository) SaveBracket(b *entity.Bracket) error {
	query := `
		INSERT INTO brackets (id, tournament_id, level, match_id, position)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			match_id = EXCLUDED.match_id
	`
	_, err := r.db.Exec(query, b.ID, b.TournamentID, b.Level, b.MatchID, b.Position)
	return err
}

func (r *PGTournamentRepository) GetBrackets(tournamentID string) ([]*entity.Bracket, error) {
	query := `SELECT id, tournament_id, level, match_id, position FROM brackets WHERE tournament_id = $1`
	rows, err := r.db.Query(query, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.Bracket
	for rows.Next() {
		var b entity.Bracket
		err := rows.Scan(&b.ID, &b.TournamentID, &b.Level, &b.MatchID, &b.Position)
		if err != nil {
			return nil, err
		}
		list = append(list, &b)
	}
	return list, nil
}

type PGLeagueRepository struct {
	db *sql.DB
}

func NewPGLeagueRepository(db *sql.DB) *PGLeagueRepository {
	return &PGLeagueRepository{db: db}
}

func (r *PGLeagueRepository) Save(l *entity.League) error {
	query := `INSERT INTO leagues (id, tournament_id, season) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING`
	_, err := r.db.Exec(query, l.ID, l.TournamentID, l.Season)
	return err
}

func (r *PGLeagueRepository) GetByTournament(tournamentID string) (*entity.League, error) {
	query := `SELECT id, tournament_id, season FROM leagues WHERE tournament_id = $1`
	var l entity.League
	err := r.db.QueryRow(query, tournamentID).Scan(&l.ID, &l.TournamentID, &l.Season)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *PGLeagueRepository) SaveDivision(d *entity.Division) error {
	query := `
		INSERT INTO divisions (id, league_id, name, capacity, promotion_slots, relegation_slots)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			capacity = EXCLUDED.capacity
	`
	_, err := r.db.Exec(query, d.ID, d.LeagueID, d.Name, d.Capacity, d.PromotionSlots, d.RelegationSlots)
	return err
}

func (r *PGLeagueRepository) GetDivisions(leagueID string) ([]*entity.Division, error) {
	query := `SELECT id, league_id, name, capacity, promotion_slots, relegation_slots FROM divisions WHERE league_id = $1`
	rows, err := r.db.Query(query, leagueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.Division
	for rows.Next() {
		var d entity.Division
		err := rows.Scan(&d.ID, &d.LeagueID, &d.Name, &d.Capacity, &d.PromotionSlots, &d.RelegationSlots)
		if err != nil {
			return nil, err
		}
		list = append(list, &d)
	}
	return list, nil
}

func (r *PGLeagueRepository) SaveStanding(s *entity.Standing) error {
	query := `
		INSERT INTO standings (division_id, bot_id, points, wins, draws, losses, matches, dano_causado)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (division_id, bot_id) DO UPDATE SET
			points = EXCLUDED.points,
			wins = EXCLUDED.wins,
			draws = EXCLUDED.draws,
			losses = EXCLUDED.losses,
			matches = EXCLUDED.matches,
			dano_causado = EXCLUDED.dano_causado
	`
	_, err := r.db.Exec(query, s.DivisionID, s.BotID, s.Points, s.Wins, s.Draws, s.Losses, s.Matches, s.DanoCausado)
	return err
}

func (r *PGLeagueRepository) GetStandings(divisionID string) ([]*entity.Standing, error) {
	query := `SELECT division_id, bot_id, points, wins, draws, losses, matches, dano_causado FROM standings WHERE division_id = $1 ORDER BY points DESC, wins DESC`
	rows, err := r.db.Query(query, divisionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.Standing
	for rows.Next() {
		var s entity.Standing
		err := rows.Scan(&s.DivisionID, &s.BotID, &s.Points, &s.Wins, &s.Draws, &s.Losses, &s.Matches, &s.DanoCausado)
		if err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, nil
}

type PGHubRepository struct {
	db *sql.DB
}

func NewPGHubRepository(db *sql.DB) *PGHubRepository {
	return &PGHubRepository{db: db}
}

func (r *PGHubRepository) Save(h *entity.Hub) error {
	query := `
		INSERT INTO hubs (id, name, type, visibility, owner_id, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			visibility = EXCLUDED.visibility,
			description = EXCLUDED.description
	`
	_, err := r.db.Exec(query, h.ID, h.Name, h.Type, h.Visibility, h.OwnerID, h.Description, h.CreatedAt)
	return err
}

func (r *PGHubRepository) GetByID(id string) (*entity.Hub, error) {
	query := `SELECT id, name, type, visibility, owner_id, description, created_at FROM hubs WHERE id = $1`
	var h entity.Hub
	err := r.db.QueryRow(query, id).Scan(&h.ID, &h.Name, &h.Type, &h.Visibility, &h.OwnerID, &h.Description, &h.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *PGHubRepository) ListPublic() ([]*entity.Hub, error) {
	query := `SELECT id, name, type, visibility, owner_id, description, created_at FROM hubs WHERE visibility = 'PUBLIC'`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.Hub
	for rows.Next() {
		var h entity.Hub
		err := rows.Scan(&h.ID, &h.Name, &h.Type, &h.Visibility, &h.OwnerID, &h.Description, &h.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, &h)
	}
	return list, nil
}

func (r *PGHubRepository) ListByUser(userID string) ([]*entity.Hub, error) {
	query := `
		SELECT h.id, h.name, h.type, h.visibility, h.owner_id, h.description, h.created_at 
		FROM hubs h
		JOIN hub_memberships m ON h.id = m.hub_id
		WHERE m.user_id = $1
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.Hub
	for rows.Next() {
		var h entity.Hub
		err := rows.Scan(&h.ID, &h.Name, &h.Type, &h.Visibility, &h.OwnerID, &h.Description, &h.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, &h)
	}
	return list, nil
}

func (r *PGHubRepository) SaveMembership(m *entity.HubMembership) error {
	query := `
		INSERT INTO hub_memberships (hub_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (hub_id, user_id) DO UPDATE SET
			role = EXCLUDED.role
	`
	_, err := r.db.Exec(query, m.HubID, m.UserID, m.Role, m.JoinedAt)
	return err
}

func (r *PGHubRepository) GetMemberships(hubID string) ([]*entity.HubMembership, error) {
	query := `SELECT hub_id, user_id, role, joined_at FROM hub_memberships WHERE hub_id = $1`
	rows, err := r.db.Query(query, hubID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.HubMembership
	for rows.Next() {
		var m entity.HubMembership
		err := rows.Scan(&m.HubID, &m.UserID, &m.Role, &m.JoinedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, nil
}
