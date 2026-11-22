package routes

import (
	"errors"
	"net/http"

	compAuth "github.com/codearena-platform/codearena-platform/internal/competition/auth"
	"github.com/codearena-platform/codearena-platform/internal/competition/services"
	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils/docs"
	"github.com/gofiber/fiber/v2"
)

var routeTags = []string{"Competition"}
var historyTags = []string{"History"}

type CompetitionHTTPHandler struct {
	matchUsecase      *services.MatchUsecase
	tournamentUsecase *services.TournamentUsecase
	leagueUsecase     *services.LeagueUsecase
	scoreboardService *services.ScoreboardService
	historyUsecase    *services.HistoryUsecase
	analyticsUsecase  *services.AnalyticsUsecase
	hubUsecase        *services.HubUsecase
	botUsecase        *services.BotUsecase
}

func RegisterCompetitionRoutes(app *fiber.App, verifier compAuth.Verifier, mu *services.MatchUsecase, tu *services.TournamentUsecase, lu *services.LeagueUsecase, ss *services.ScoreboardService, hu *services.HistoryUsecase, au *services.AnalyticsUsecase, hbu *services.HubUsecase, bu *services.BotUsecase) {
	h := &CompetitionHTTPHandler{
		matchUsecase:      mu,
		tournamentUsecase: tu,
		leagueUsecase:     lu,
		scoreboardService: ss,
		historyUsecase:    hu,
		analyticsUsecase:  au,
		hubUsecase:        hbu,
		botUsecase:        bu,
	}

	api := app.Group("/api/v1/competition")
	basePath := "/api/v1/competition"

	// Tournaments
	api.Post("/tournaments", compAuth.RequireUser(verifier), h.CreateTournament)
	docs.RegisterApi(docs.ApiWrapper{
		Path:        basePath + "/tournaments",
		Method:      http.MethodPost,
		Name:        "Create Tournament",
		Description: "Create a new tournament",
		RequestBody: &docs.ApiRequestBody{
			Description: "Tournament data",
			Content:     new(entity.Tournament),
		},
		Response: &docs.ApiResponse{
			Description: "Tournament created",
			Content:     new(entity.Tournament),
		},
		Tags: routeTags,
	})

	api.Post("/tournaments/:id/brackets", compAuth.RequireUser(verifier), h.GenerateBrackets)
	docs.RegisterApi(docs.ApiWrapper{
		Path:        basePath + "/tournaments/:id/brackets",
		Method:      http.MethodPost,
		Name:        "Generate Brackets",
		Description: "Generate brackets for a tournament",
		Parameters: []docs.ApiParameter{
			{Name: "id", In: "path", Required: true},
		},
		Tags: routeTags,
	})

	api.Get("/tournaments/:id/brackets", h.GetBrackets)
	docs.RegisterApi(docs.ApiWrapper{
		Path:        basePath + "/tournaments/:id/brackets",
		Method:      http.MethodGet,
		Name:        "Get Brackets",
		Description: "Get brackets for a tournament",
		Parameters: []docs.ApiParameter{
			{Name: "id", In: "path", Required: true},
		},
		Response: &docs.ApiResponse{
			Description: "Brackets list",
			Content:     []*entity.Bracket{},
		},
		Tags: routeTags,
	})

	// Leagues
	api.Post("/leagues", compAuth.RequireUser(verifier), h.CreateLeague)
	docs.RegisterApi(docs.ApiWrapper{
		Path:        basePath + "/leagues",
		Method:      http.MethodPost,
		Name:        "Create League",
		Description: "Create a new league season",
		Response: &docs.ApiResponse{
			Description: "League created",
			Content:     new(entity.League),
		},
		Tags: routeTags,
	})

	api.Get("/leagues/:id/divisions", h.GetDivisions)
	docs.RegisterApi(docs.ApiWrapper{
		Path:        basePath + "/leagues/:id/divisions",
		Method:      http.MethodGet,
		Name:        "Get Divisions",
		Description: "Get divisions in a league",
		Parameters: []docs.ApiParameter{
			{Name: "id", In: "path", Required: true},
		},
		Tags: routeTags,
	})

	api.Get("/divisions/:id/standings", h.GetStandings)
	docs.RegisterApi(docs.ApiWrapper{
		Path:        basePath + "/divisions/:id/standings",
		Method:      http.MethodGet,
		Name:        "Get Standings",
		Description: "Get standings for a division",
		Parameters: []docs.ApiParameter{
			{Name: "id", In: "path", Required: true},
		},
		Response: &docs.ApiResponse{
			Description: "Standings list",
			Content:     []*entity.Standing{},
		},
		Tags: routeTags,
	})

	// Judge Callbacks
	api.Post("/matches/:id/result", compAuth.RequireUser(verifier), h.ReceiveMatchResult)
	docs.RegisterApi(docs.ApiWrapper{
		Path:        basePath + "/matches/:id/result",
		Method:      http.MethodPost,
		Name:        "Receive Match Result",
		Description: "Endpoint for simulation engine to report results",
		Parameters: []docs.ApiParameter{
			{Name: "id", In: "path", Required: true},
		},
		Tags: routeTags,
	})

	// History & Discovery
	api.Get("/history", h.GetGlobalHistory)
	docs.RegisterApi(docs.ApiWrapper{
		Path:        basePath + "/history",
		Method:      http.MethodGet,
		Name:        "Global History",
		Description: "Get lists of all finished matches",
		Parameters: []docs.ApiParameter{
			{Name: "limit", In: "query", Required: false},
		},
		Response: &docs.ApiResponse{
			Description: "Match history",
			Content:     []*entity.Match{},
		},
		Tags: historyTags,
	})

	api.Get("/bots/:id/history", h.GetBotHistory)
	docs.RegisterApi(docs.ApiWrapper{
		Path:        basePath + "/bots/:id/history",
		Method:      http.MethodGet,
		Name:        "Bot History",
		Description: "Get history for a specific bot",
		Parameters: []docs.ApiParameter{
			{Name: "id", In: "path", Required: true},
		},
		Tags: historyTags,
	})

	api.Get("/tournaments/:id/history", h.GetTournamentHistory)
	docs.RegisterApi(docs.ApiWrapper{
		Path:        basePath + "/tournaments/:id/history",
		Method:      http.MethodGet,
		Name:        "Tournament History",
		Description: "Get history for a specific tournament",
		Parameters: []docs.ApiParameter{
			{Name: "id", In: "path", Required: true},
		},
		Tags: historyTags,
	})

	// Analytics
	api.Get("/analytics/global", h.GetGlobalAnalytics)
	api.Get("/analytics/bots/:id", h.GetBotAnalyticsStats)

	// Hubs
	api.Post("/hubs", compAuth.RequireUser(verifier), h.CreateHub)
	// Listing hubs is scoped to the authenticated user instead of a header.
	api.Get("/hubs", compAuth.RequireUser(verifier), h.ListHubs)
	api.Get("/hubs/:id/tournaments", h.GetHubTournaments)
	api.Post("/hubs/:id/members", compAuth.RequireUser(verifier), h.AddHubMember)

	// Bots
	api.Post("/bots", compAuth.RequireUser(verifier), h.SubmitBot)

	// Leaderboard
	api.Get("/leaderboard", h.GetLeaderboard)
}

// authorizeTournament checks hub membership for tournaments that belong to a
// hub. Global tournaments (no hub) remain readable/writable by any verified
// user because the data model does not record an owner for them yet.
func (h *CompetitionHTTPHandler) authorizeTournament(tournamentID string, identity *compAuth.Identity) error {
	tournament, err := h.tournamentUsecase.GetByID(tournamentID)
	if err != nil {
		return err
	}
	if tournament.HubID == "" {
		return nil
	}
	member, err := h.hubUsecase.IsMember(tournament.HubID, identity.ID)
	if err != nil {
		return err
	}
	if !member {
		return errors.New("caller is not a member of the tournament's hub")
	}
	return nil
}

func (h *CompetitionHTTPHandler) GetLeaderboard(c *fiber.Ctx) error {
	topBots, err := h.scoreboardService.GetGlobalHallOfFame()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(topBots)
}

func (h *CompetitionHTTPHandler) SubmitBot(c *fiber.Ctx) error {
	var req struct {
		Name       string `json:"name"`
		Chassis    string `json:"chassis"`
		SourceCode string `json:"source_code"` // Contains the raw script logic submitted by user
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	identity, err := compAuth.IdentityFromFiber(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	ctx := c.Context()

	bot, err := h.botUsecase.SubmitBot(ctx, identity.ID, req.Name, req.Chassis, req.SourceCode)
	if err != nil {
		// Differentiate plagiarism (business error) vs server error
		return c.Status(406).JSON(fiber.Map{"error": err.Error(), "type": "PlagiarismDetected"})
	}

	return c.Status(201).JSON(bot)
}

func (h *CompetitionHTTPHandler) CreateTournament(c *fiber.Ctx) error {
	var req struct {
		Name string                `json:"name"`
		Type entity.TournamentType `json:"type"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	t, err := h.tournamentUsecase.CreateTournament(req.Name, req.Type)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(t)
}

func (h *CompetitionHTTPHandler) GenerateBrackets(c *fiber.Ctx) error {
	tournamentID := c.Params("id")
	identity, err := compAuth.IdentityFromFiber(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.authorizeTournament(tournamentID, identity); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}
	var req struct {
		Participants []string `json:"participants"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	err = h.tournamentUsecase.GenerateBrackets(tournamentID, req.Participants)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(204)
}

func (h *CompetitionHTTPHandler) GetBrackets(c *fiber.Ctx) error {
	tournamentID := c.Params("id")
	brackets, err := h.scoreboardService.GetTournamentBracket(tournamentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(brackets)
}

func (h *CompetitionHTTPHandler) CreateLeague(c *fiber.Ctx) error {
	identity, err := compAuth.IdentityFromFiber(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	var req struct {
		TournamentID string `json:"tournament_id"`
		Season       string `json:"season"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.authorizeTournament(req.TournamentID, identity); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	l, err := h.leagueUsecase.CreateLeague(req.TournamentID, req.Season)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Setup default divisions
	h.leagueUsecase.SetupDivisions(l.ID)

	return c.Status(201).JSON(l)
}

func (h *CompetitionHTTPHandler) GetDivisions(c *fiber.Ctx) error {
	// Not implemented in usecase yet but planned
	return c.SendStatus(501)
}

func (h *CompetitionHTTPHandler) GetStandings(c *fiber.Ctx) error {
	divisionID := c.Params("id")
	standings, err := h.scoreboardService.GetDivisionStandings(divisionID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(standings)
}

func (h *CompetitionHTTPHandler) ReceiveMatchResult(c *fiber.Ctx) error {
	identity, err := compAuth.IdentityFromFiber(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	if err := compAuth.AuthorizeMatchResult(identity); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}
	matchID := c.Params("id")
	var req struct {
		WinnerID string         `json:"winner_id"`
		Metrics  map[string]any `json:"metrics"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Update match status and ELO
	if err := h.matchUsecase.CompleteMatch(matchID, req.WinnerID, req.Metrics); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(200)
}

func (h *CompetitionHTTPHandler) GetGlobalHistory(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	history, err := h.historyUsecase.GetGlobalHistory(limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(history)
}

func (h *CompetitionHTTPHandler) GetBotHistory(c *fiber.Ctx) error {
	botID := c.Params("id")
	history, err := h.historyUsecase.GetBotHistory(botID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(history)
}

func (h *CompetitionHTTPHandler) GetTournamentHistory(c *fiber.Ctx) error {
	tournamentID := c.Params("id")
	history, err := h.historyUsecase.GetTournamentHistory(tournamentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(history)
}

func (h *CompetitionHTTPHandler) GetGlobalAnalytics(c *fiber.Ctx) error {
	stats, err := h.analyticsUsecase.GetGlobalStats()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(stats)
}

func (h *CompetitionHTTPHandler) GetBotAnalyticsStats(c *fiber.Ctx) error {
	botID := c.Params("id")
	stats, err := h.analyticsUsecase.GetBotAnalytics(botID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(stats)
}

func (h *CompetitionHTTPHandler) CreateHub(c *fiber.Ctx) error {
	var req struct {
		Name        string               `json:"name"`
		Type        entity.HubType       `json:"type"`
		Visibility  entity.HubVisibility `json:"visibility"`
		Description string               `json:"description"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	identity, err := compAuth.IdentityFromFiber(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	hub, err := h.hubUsecase.CreateHub(req.Name, identity.ID, req.Description, req.Type, req.Visibility)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(hub)
}

func (h *CompetitionHTTPHandler) ListHubs(c *fiber.Ctx) error {
	identity, err := compAuth.IdentityFromFiber(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	hubs, err := h.hubUsecase.ListUserHubs(identity.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(hubs)
}

func (h *CompetitionHTTPHandler) GetHubTournaments(c *fiber.Ctx) error {
	hubID := c.Params("id")
	tournaments, err := h.hubUsecase.GetHubTournaments(hubID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(tournaments)
}

func (h *CompetitionHTTPHandler) AddHubMember(c *fiber.Ctx) error {
	identity, err := compAuth.IdentityFromFiber(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	hubID := c.Params("id")
	hub, err := h.hubUsecase.GetHub(hubID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "hub not found"})
	}
	if hub.OwnerID != identity.ID {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "only the hub owner can add members"})
	}
	var req struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.hubUsecase.AddMember(hubID, req.UserID, req.Role); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(204)
}
