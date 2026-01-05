package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	pb "starter-kit-grpc-golang/api/gen/v1"
	"starter-kit-grpc-golang/config"
	"starter-kit-grpc-golang/internal/grpc_handler"
	"starter-kit-grpc-golang/internal/interceptor"
	"starter-kit-grpc-golang/internal/repository"
	"starter-kit-grpc-golang/internal/service"
	"starter-kit-grpc-golang/pkg/logger"
	"starter-kit-grpc-golang/pkg/swagger"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/proto"
)

// HttpResponseModifier checks for a specific metadata key to change the HTTP Status Code
func HttpResponseModifier(ctx context.Context, w http.ResponseWriter, p proto.Message) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return nil
	}
	if vals := md.HeaderMD.Get("x-http-code"); len(vals) > 0 {
		code, err := strconv.Atoi(vals[0])
		if err == nil {
			delete(md.HeaderMD, "x-http-code")
			w.Header().Del("Grpc-Metadata-X-Http-Code")
			w.WriteHeader(code)
		}
	}
	return nil
}

func main() {
	// 1. Load Config & Logger
	cfg := config.LoadConfig()
	logger.InitLogger(cfg.Env)
	logger.Log.Info("Starting server...", "env", cfg.Env)

	// 2. Connect DB
	config.ConnectDB(cfg)

	// 3. Dependency Injection
	userRepo := repository.NewUserRepository(config.DB)
	tokenRepo := repository.NewTokenRepository(config.DB)

	tokenService := service.NewTokenService(tokenRepo, cfg)
	emailService := service.NewEmailService(cfg)
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, tokenRepo, tokenService, emailService, cfg)

	authHandler := grpc_handler.NewAuthHandler(authService)
	userHandler := grpc_handler.NewUserHandler(userService)
	healthHandler := grpc_handler.NewHealthHandler()

	// 4. Setup gRPC Server
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryInterceptor(),
			interceptor.LoggerInterceptor(),
			// interceptor.RateLimitInterceptor(), // --> Uncomment for using RateLimiter
			interceptor.AuthInterceptor(cfg),
		),
	)

	pb.RegisterAuthServiceServer(grpcServer, authHandler)
	pb.RegisterUserServiceServer(grpcServer, userHandler)
	pb.RegisterHealthServiceServer(grpcServer, healthHandler)

	if cfg.Env == "development" {
		reflection.Register(grpcServer)
	}

	// 5. Start Servers
	errChan := make(chan error, 1)

	// --- gRPC Server ---
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

	// --- HTTP Gateway & Swagger ---
	go func() {
		time.Sleep(time.Second)

		ctx := context.Background()
		
		// Create the gRPC-Gateway Mux
		gwmux := runtime.NewServeMux(
			runtime.WithForwardResponseOption(HttpResponseModifier),
		)
		
		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}
		grpcEndpoint := "localhost:" + cfg.GRPCPort

		// Register Services to Gateway
		if err := pb.RegisterAuthServiceHandlerFromEndpoint(ctx, gwmux, grpcEndpoint, opts); err != nil {
			errChan <- err; return
		}
		if err := pb.RegisterUserServiceHandlerFromEndpoint(ctx, gwmux, grpcEndpoint, opts); err != nil {
			errChan <- err; return
		}
		if err := pb.RegisterHealthServiceHandlerFromEndpoint(ctx, gwmux, grpcEndpoint, opts); err != nil {
			errChan <- err; return
		}

		// Create a Root Mux to handle both Swagger and Gateway
		mux := http.NewServeMux()
		
		// Mount Gateway (API)
		mux.Handle("/", gwmux)

		// Mount Swagger
		mux.HandleFunc("/swagger.json", swagger.ServeJSON)
		mux.HandleFunc("/swagger-ui", swagger.ServeUI)

		logger.Log.Info("HTTP Gateway & Swagger listening", "port", cfg.GatewayPort)
		if err := http.ListenAndServe(":"+cfg.GatewayPort, mux); err != nil {
			errChan <- fmt.Errorf("gateway server error: %v", err)
		}
	}()

	// 6. Shutdown
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