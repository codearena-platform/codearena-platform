package routes

import (
	"context"
	"log"
	"net"

	pb "github.com/codearena-platform/codearena-core/pkg/api/v1"
	"github.com/codearena-platform/codearena-platform/internal/competition/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MatchGRPCServer struct {
	pb.UnimplementedMatchServiceServer
	service *services.MatchUsecase
	// Client to trigger simulation
	simClient pb.SimulationServiceClient
}

func NewMatchGRPCServer(u *services.MatchUsecase, simAddr string) *MatchGRPCServer {
	conn, err := grpc.Dial(simAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("WARNING: Could not connect to Simulation Engine at %s: %v", simAddr, err)
	}

	return &MatchGRPCServer{
		service:   u,
		simClient: pb.NewSimulationServiceClient(conn),
	}
}

func (s *MatchGRPCServer) CreateMatch(ctx context.Context, config *pb.ArenaConfig) (*pb.MatchResponse, error) {
	log.Printf("Competition: Creating match for arena: %s", config.Name)

	match, err := s.service.CreateMatch(config.Id, []string{})
	if err != nil {
		return nil, err
	}

	// Trigger Execution through the Judge (via StartMatch)
	if err := s.service.StartMatch(ctx, match.ID); err != nil {
		log.Printf("ERROR: Failed to start match execution: %v", err)
	}

	return &pb.MatchResponse{
		MatchId: match.ID,
		Status:  pb.MatchStatus_RUNNING,
	}, nil
}

func (s *MatchGRPCServer) RegisterBot(ctx context.Context, req *pb.RegisterBotRequest) (*pb.RegisterBotResponse, error) {
	log.Printf("Competition: Registering bot: %s for user: %s", req.Name, req.UserId)

	bot, version, err := s.service.RegisterBot(req.UserId, req.Name, req.Version, req.SourceCode, req.Language)
	if err != nil {
		return &pb.RegisterBotResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.RegisterBotResponse{
		BotId:     bot.ID,
		VersionId: version.ID,
		Success:   true,
		Message:   "Bot registered successfully",
	}, nil
}

func StartMatchServer(addr string, u *services.MatchUsecase) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	srv := &MatchGRPCServer{service: u}
	pb.RegisterMatchServiceServer(s, srv)

	log.Printf("Match gRPC Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
