package main

import (
	"fmt"
	"log"
	"net"
	"os"

	compAuth "github.com/codearena-platform/codearena-platform/internal/competition/auth"
	"github.com/codearena-platform/codearena-platform/internal/competition/db/mongo"
	"github.com/codearena-platform/codearena-platform/internal/competition/db/postgres"
	"github.com/codearena-platform/codearena-platform/internal/competition/routes"
	"github.com/codearena-platform/codearena-platform/internal/competition/services"
	"github.com/codearena-platform/codearena-platform/internal/competition/services/repository"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"

	pb "github.com/codearena-platform/codearena-core/pkg/api/v1"
)

func main() {
	fmt.Println("CodeArena Competition Manager starting...")

	// Configuration (Envs)
	pgDSN := os.Getenv("POSTGRES_DSN")
	if pgDSN == "" {
		pgDSN = "postgres://codearena:codearena_password@postgres:5432/codearena_db?sslmode=disable"
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://codearena:codearena_password@mongodb:27017"
	}

	// In production, losing the database must not silently switch to in-memory
	// repositories: that would make results and rankings disappear unnoticed.
	allowMemoryFallback := os.Getenv("ALLOW_IN_MEMORY_FALLBACK") == "true"

	// Initialize Postgres
	db, err := postgres.InitPostgres(pgDSN)
	if err != nil {
		if !allowMemoryFallback {
			log.Fatalf("failed to connect to Postgres: %v", err)
		}
		log.Printf("WARNING: falling back to in-memory repositories: %v", err)
	} else {
		log.Println("Connected to PostgreSQL")
	}

	// Initialize Mongo
	mongoClient, err := mongo.InitMongo(mongoURI)
	if err != nil {
		if !allowMemoryFallback {
			log.Fatalf("failed to connect to Mongo: %v", err)
		}
		log.Printf("WARNING: match replays will not be persisted: %v", err)
	} else {
		log.Println("Connected to MongoDB")
	}

	// Repositories
	var matchRepo repository.MatchRepository
	var botRepo repository.BotRepository
	var versionRepo repository.BotVersionRepository
	var scoreRepo repository.ScoreRepository
	var tournamentRepo repository.TournamentRepository
	var leagueRepo repository.LeagueRepository
	var hubRepo repository.HubRepository

	if db != nil && mongoClient != nil {
		botRepo = postgres.NewPGBotRepository(db)
		versionRepo = repository.NewMemoryBotVersionRepository() // TODO: Implement PG variant
		scoreRepo = postgres.NewPGScoreRepository(db)
		matchRepo = mongo.NewMongoMatchRepository(mongoClient, "codearena")
		tournamentRepo = postgres.NewPGTournamentRepository(db)
		leagueRepo = postgres.NewPGLeagueRepository(db)
		hubRepo = postgres.NewPGHubRepository(db)
		log.Println("Using Persistent Repositories (Postgres + Mongo)")
	} else {
		log.Println("Using In-Memory Repositories (Fallback)")
		tournamentRepo = repository.NewMemoryTournamentRepository()
		leagueRepo = repository.NewMemoryLeagueRepository()
		matchRepo = repository.NewMemoryMatchRepository()
		botRepo = repository.NewMemoryBotRepository()
		versionRepo = repository.NewMemoryBotVersionRepository()
		scoreRepo = repository.NewMemoryScoreRepository()
		hubRepo = repository.NewMemoryHubRepository()
		log.Println("Memory Repositories Initialized")
	}

	// Initialize Usecases & Services
	simAddr := os.Getenv("SIMULATION_URL")
	if simAddr == "" {
		simAddr = "simulation:50051"
	}
	judgeService := services.NewJudgeService(matchRepo, botRepo, simAddr)
	plagiarismSvc := services.NewPlagiarismService()

	botUsecase := services.NewBotUsecase(botRepo, versionRepo, plagiarismSvc)
	matchUsecase := services.NewMatchUsecase(matchRepo, scoreRepo, botRepo, leagueRepo, judgeService)
	tournamentUsecase := services.NewTournamentUsecase(tournamentRepo, matchRepo)
	leagueUsecase := services.NewLeagueUsecase(leagueRepo, botRepo)
	scoreboardService := services.NewScoreboardService(tournamentRepo, leagueRepo, botRepo)
	historyUsecase := services.NewHistoryUsecase(matchRepo, botRepo)
	analyticsUsecase := services.NewAnalyticsUsecase(matchRepo, botRepo)
	hubUsecase := services.NewHubUsecase(hubRepo, tournamentRepo)
	// metricsService := services.NewMetricsService(matchRepo) // Planned but initialized inside scoreboard if needed

	// Tokens are verified through IAM; never trust caller-supplied identity fields.
	iamURL := os.Getenv("IAM_URL")
	verifier := compAuth.NewIAMVerifier(iamURL)

	// Run gRPC Server
	go StartMatchServer(":50052", simAddr, matchUsecase, verifier)

	// Run HTTP Server
	StartHTTPServer(":7001", verifier, matchUsecase, tournamentUsecase, leagueUsecase, scoreboardService, historyUsecase, analyticsUsecase, hubUsecase, botUsecase)
}

func StartHTTPServer(addr string, verifier compAuth.Verifier, mu *services.MatchUsecase, tu *services.TournamentUsecase, lu *services.LeagueUsecase, ss *services.ScoreboardService, hu *services.HistoryUsecase, au *services.AnalyticsUsecase, hbu *services.HubUsecase, bu *services.BotUsecase) {
	app := fiber.New(fiber.Config{
		AppName: "CodeArena Competition Manager v1.0.0",
	})

	// Health check
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "UP",
			"version": "1.0.0",
		})
	})

	routes.RegisterCompetitionRoutes(app, verifier, mu, tu, lu, ss, hu, au, hbu, bu)

	log.Printf("Competition HTTP Server listening at %s", addr)
	log.Fatal(app.Listen(addr))
}

func StartMatchServer(addr string, simAddr string, u *services.MatchUsecase, verifier compAuth.Verifier) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(compAuth.UnaryInterceptor(verifier)))
	srv := routes.NewMatchGRPCServer(u, simAddr)
	pb.RegisterMatchServiceServer(s, srv)

	log.Printf("Match gRPC Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
