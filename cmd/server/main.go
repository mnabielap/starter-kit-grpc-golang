package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "starter-kit-grpc-golang/api/gen/v1"
	"starter-kit-grpc-golang/config"
	"starter-kit-grpc-golang/internal/grpc_handler"
	"starter-kit-grpc-golang/internal/interceptor"
	"starter-kit-grpc-golang/internal/repository"
	"starter-kit-grpc-golang/internal/service"
	"starter-kit-grpc-golang/pkg/logger"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. Load Configuration & Logger
	cfg := config.LoadConfig()
	logger.InitLogger(cfg.Env)
	logger.Log.Info("Starting server...", "env", cfg.Env)

	// 2. Connect to Database
	config.ConnectDB(cfg)

	// 3. Dependency Injection (Repository -> Service -> Handler)
	
	// Repositories
	userRepo := repository.NewUserRepository(config.DB)
	tokenRepo := repository.NewTokenRepository(config.DB)

	// Services
	tokenService := service.NewTokenService(tokenRepo, cfg)
	emailService := service.NewEmailService(cfg)
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, tokenRepo, tokenService, emailService, cfg)

	// Handlers (Controllers)
	authHandler := grpc_handler.NewAuthHandler(authService)
	userHandler := grpc_handler.NewUserHandler(userService)
	healthHandler := grpc_handler.NewHealthHandler()

	// 4. Setup gRPC Server with Interceptors
	// Chain order: Recovery -> Logger -> RateLimit -> Auth -> Handler
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryInterceptor(),
			interceptor.LoggerInterceptor(),
			interceptor.RateLimitInterceptor(),
			interceptor.AuthInterceptor(cfg),
		),
	)

	// Register Proto Services
	pb.RegisterAuthServiceServer(grpcServer, authHandler)
	pb.RegisterUserServiceServer(grpcServer, userHandler)
	pb.RegisterHealthServiceServer(grpcServer, healthHandler)

	// Enable Reflection (Useful for testing with Postman/grpcurl)
	if cfg.Env == "development" {
		reflection.Register(grpcServer)
	}

	// 5. Start Servers (gRPC & HTTP Gateway)
	
	// Channel to listen for errors
	errChan := make(chan error, 1)

	// --- Run gRPC Server ---
	go func() {
		listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			errChan <- fmt.Errorf("failed to listen gRPC: %v", err)
			return
		}
		logger.Log.Info("gRPC Server listening", "port", cfg.GRPCPort)
		if err := grpcServer.Serve(listener); err != nil {
			errChan <- fmt.Errorf("gRPC server error: %v", err)
		}
	}()

	// --- Run HTTP Gateway Server ---
	go func() {
		// Wait a moment for gRPC to start
		time.Sleep(time.Second)

		ctx := context.Background()
		mux := runtime.NewServeMux()
		
		// Dial options to connect to local gRPC server
		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}
		grpcEndpoint := "localhost:" + cfg.GRPCPort

		// Register Handlers for Gateway
		err := pb.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
		if err != nil {
			errChan <- err
			return
		}
		err = pb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
		if err != nil {
			errChan <- err
			return
		}
		err = pb.RegisterHealthServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
		if err != nil {
			errChan <- err
			return
		}

		logger.Log.Info("HTTP Gateway listening", "port", cfg.GatewayPort)
		if err := http.ListenAndServe(":"+cfg.GatewayPort, mux); err != nil {
			errChan <- fmt.Errorf("gateway server error: %v", err)
		}
	}()

	// 6. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Log.Info("Shutting down servers...")
		grpcServer.GracefulStop()
	case err := <-errChan:
		logger.Log.Error("Server failed", "error", err)
	}
}