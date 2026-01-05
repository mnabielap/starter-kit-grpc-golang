package grpc_handler

import (
	"context"

	pb "starter-kit-grpc-golang/api/gen/v1"
)

type HealthHandler struct {
	pb.UnimplementedHealthServiceServer
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Status:  "healthy",
		Message: "Server is running",
	}, nil
}