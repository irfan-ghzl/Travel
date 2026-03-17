package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/irfan-ghzl/pintour/internal/application/auth"
	appbooking "github.com/irfan-ghzl/pintour/internal/application/booking"
	apppayment "github.com/irfan-ghzl/pintour/internal/application/payment"
	appreview "github.com/irfan-ghzl/pintour/internal/application/review"
	apptour "github.com/irfan-ghzl/pintour/internal/application/tour"
	"github.com/irfan-ghzl/pintour/common/config"
	db "github.com/irfan-ghzl/pintour/db/sqlc"
	infrapayment "github.com/irfan-ghzl/pintour/internal/infrastructure/payment"
	"github.com/irfan-ghzl/pintour/internal/infrastructure/oauth"
	"github.com/irfan-ghzl/pintour/internal/infrastructure/persistence"
	"github.com/irfan-ghzl/pintour/common/logger"
	grpchandler "github.com/irfan-ghzl/pintour/common/interface/grpc"
	"github.com/irfan-ghzl/pintour/common/interface/middleware"
	"github.com/irfan-ghzl/pintour/common/token"
	pb "github.com/irfan-ghzl/pintour/pb/pintour/v1"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	conn, err := sql.Open("pgx", cfg.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to database")
	}
	defer conn.Close()

	if err := conn.PingContext(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("cannot ping database")
	}
	log.Info().Msg("connected to database")

	// Infrastructure layer
	store := db.New(conn)
	userRepo := persistence.NewUserRepository(store)
	tourRepo := persistence.NewTourRepository(store)
	bookingRepo := persistence.NewBookingRepository(store)
	paymentRepo := persistence.NewPaymentRepository(store)
	reviewRepo := persistence.NewReviewRepository(store)

	paymentGateway := infrapayment.NewMidtransGateway(cfg.MidtransServerKey, cfg.MidtransIsProduction)
	oauthProvider := oauth.NewGoogleProvider(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL)

	// Token maker
	tokenMaker, err := token.NewJWTMaker(cfg.TokenSymmetricKey)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create token maker")
	}

	// Application layer
	authService := auth.NewService(userRepo, tokenMaker, oauthProvider, cfg)
	tourService := apptour.NewService(tourRepo)
	bookingService := appbooking.NewService(bookingRepo, tourRepo)
	paymentService := apppayment.NewService(paymentRepo, bookingRepo, userRepo, paymentGateway, cfg.MidtransServerKey)
	reviewService := appreview.NewService(reviewRepo, bookingRepo, userRepo)

	// Interface layer — thin gRPC handlers
	server := grpchandler.NewServer(authService, tourService, bookingService, paymentService, reviewService)

	// Start gRPC server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.AuthInterceptor(tokenMaker)),
	)

	pb.RegisterAuthServiceServer(grpcServer, server)
	pb.RegisterTourServiceServer(grpcServer, server)
	pb.RegisterBookingServiceServer(grpcServer, server)
	pb.RegisterPaymentServiceServer(grpcServer, server)
	pb.RegisterReviewServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	grpcListener, err := net.Listen("tcp", cfg.GRPCServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create gRPC listener")
	}

	go func() {
		log.Info().Msgf("starting gRPC server at %s", cfg.GRPCServerAddress)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatal().Err(err).Msg("cannot start gRPC server")
		}
	}()

	// Start HTTP gateway
	gwMux := runtime.NewServeMux()
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := pb.RegisterAuthServiceHandlerFromEndpoint(context.Background(), gwMux, cfg.GRPCServerAddress, dialOpts); err != nil {
		log.Fatal().Err(err).Msg("cannot register auth service handler")
	}
	if err := pb.RegisterTourServiceHandlerFromEndpoint(context.Background(), gwMux, cfg.GRPCServerAddress, dialOpts); err != nil {
		log.Fatal().Err(err).Msg("cannot register tour service handler")
	}
	if err := pb.RegisterBookingServiceHandlerFromEndpoint(context.Background(), gwMux, cfg.GRPCServerAddress, dialOpts); err != nil {
		log.Fatal().Err(err).Msg("cannot register booking service handler")
	}
	if err := pb.RegisterPaymentServiceHandlerFromEndpoint(context.Background(), gwMux, cfg.GRPCServerAddress, dialOpts); err != nil {
		log.Fatal().Err(err).Msg("cannot register payment service handler")
	}
	if err := pb.RegisterReviewServiceHandlerFromEndpoint(context.Background(), gwMux, cfg.GRPCServerAddress, dialOpts); err != nil {
		log.Fatal().Err(err).Msg("cannot register review service handler")
	}

	// Serve Swagger UI
	httpMux := http.NewServeMux()
	httpMux.Handle("/", gwMux)
	httpMux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/swagger/pintour.swagger.json")
	})

	httpServer := &http.Server{
		Addr:    cfg.HTTPServerAddress,
		Handler: logger.HTTPLogger(httpMux),
	}

	go func() {
		log.Info().Msgf("starting HTTP gateway at %s", cfg.HTTPServerAddress)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("cannot start HTTP gateway")
		}
	}()

	// Block until an OS signal is received, then shut down gracefully.
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)
	<-shutdownSignal

	log.Info().Msg("shutting down servers...")

	grpcServer.GracefulStop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("HTTP server shutdown error")
	}

	log.Info().Msg("servers stopped")
}
